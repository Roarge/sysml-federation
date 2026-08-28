// graphql.js: the two apps' only way to the router (SR-40).
//
// query() sends a GraphQL document by fetch POST to /graphql on the page's
// own origin and surfaces GraphQL errors as thrown Errors. subscribe()
// opens a subscription over server-sent events on a fetch POST, the path
// the router documents, reads the stream frame by frame and reconnects
// with backoff until unsubscribed (AD-0014). Frames are "event: next" with
// a "data:" line, ":heartbeat" comments, and "event: complete".
//
// A subscription is alive only while frames keep arriving. A connection
// that dies without a close leaves the read pending forever, and the app
// then shows stale data and never reconnects. The router's heartbeat is
// what makes liveness testable, so subscribe() holds a watchdog over the
// request and tears it down when too long passes in silence.

const ENDPOINT = "/graphql";
const BACKOFF_START_MS = 1000;
const BACKOFF_CAP_MS = 8000;

// The router writes ":heartbeat" every five seconds, measured against
// router 0.343.1 over the demo's own configuration. Four intervals is the
// liveness budget: it survives three lost heartbeats before it acts, and a
// browser that throttles timers in a hidden tab can only fire the watchdog
// late, which delays detection rather than causing a false one.
const HEARTBEAT_MS = 5000;
const STALL_MS = 4 * HEARTBEAT_MS;

const messageOf = (errors) => errors.map((e) => e.message).join("; ");
const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

// statusOf names a response by its status line. statusText is empty over
// HTTP/2, so it is left off rather than trailing a space.
const statusOf = (response) =>
  response.statusText === "" ? `${response.status}` : `${response.status} ${response.statusText}`;

// readJSON answers the parsed body, or null where there is nothing a caller
// can read: an empty body, one that is not JSON, or one that is not the
// object a GraphQL answer always is.
async function readJSON(response) {
  try {
    const body = await response.json();
    return typeof body === "object" && body !== null ? body : null;
  } catch {
    return null;
  }
}

// refusalIn answers the sentence a GraphQL answer carries, or null where it
// carries none. A refusal is a sentence the adapter wrote, and that sentence
// is what the person needs rather than the number in front of it.
const refusalIn = (body) =>
  body !== null && Array.isArray(body.errors) && body.errors.length > 0 ? messageOf(body.errors) : null;

// failureOf names why an answer cannot be used. The body is read first and
// the status is the fallback for an answer that carries no error of its own.
// Both entry points below judge a refusal this way, so the same failure
// reaches a person in the same words whichever one met it.
async function failureOf(response) {
  return refusalIn(await readJSON(response)) ?? `the router answered ${statusOf(response)}`;
}

export async function query(document, variables = {}) {
  const response = await fetch(ENDPOINT, {
    method: "POST",
    headers: { "Content-Type": "application/json", Accept: "application/json" },
    body: JSON.stringify({ query: document, variables }),
  });
  // The body is read before the status is judged.
  const body = await readJSON(response);
  const refusal = refusalIn(body);
  if (refusal !== null) {
    throw new Error(refusal);
  }
  if (body === null) {
    throw new Error(`the router answered ${statusOf(response)} and no GraphQL body`);
  }
  if (!response.ok) {
    throw new Error(`the router answered ${statusOf(response)}`);
  }
  return body.data;
}

// parseFrame turns one SSE frame, the text between two blank lines, into
// its event name and data. Comment lines are ignored, which is how the
// router's ":heartbeat" is dropped.
export function parseFrame(frame) {
  let event = "message";
  const data = [];
  for (const line of frame.split("\n")) {
    if (line.startsWith(":")) continue;
    if (line.startsWith("event:")) event = line.slice(6).trim();
    else if (line.startsWith("data:")) data.push(line.slice(5).replace(/^ /, ""));
  }
  return { event, data: data.join("\n") };
}

async function readStream(body, onFrame) {
  const reader = body.pipeThrough(new TextDecoderStream()).getReader();
  let buffer = "";
  for (;;) {
    const { value, done } = await reader.read();
    if (done) return;
    buffer += value.replace(/\r\n/g, "\n");
    let cut;
    while ((cut = buffer.indexOf("\n\n")) >= 0) {
      const frame = buffer.slice(0, cut);
      buffer = buffer.slice(cut + 2);
      if (frame.length > 0) onFrame(parseFrame(frame));
    }
  }
}

export function subscribe(document, onEvent, onError = () => {}) {
  let active = true;
  let controller = null;
  let delay = BACKOFF_START_MS;
  let watchdog = null;

  const stopWatching = () => {
    if (watchdog !== null) {
      clearTimeout(watchdog);
      watchdog = null;
    }
  };

  // watch is the liveness contract. Every frame the router writes starts it
  // again, a heartbeat as much as an event, because a heartbeat is the
  // signal that the connection is alive. On expiry the request is aborted,
  // which fails the pending read and hands the loop below to its backoff.
  const watch = (request) => {
    stopWatching();
    watchdog = setTimeout(() => {
      const quiet = `no frame for ${STALL_MS / 1000} s, reconnecting in ${delay / 1000} s`;
      request.abort(new Error(quiet));
    }, STALL_MS);
  };

  // report hands an error to the app. The loop below is the only thing
  // keeping the app in touch with the router, and a handler that throws must
  // not take it down: the app would then sit on stale data with nothing left
  // to reconnect it. There is nowhere further to report a reporter that
  // failed, so the throw stops here and the subscription carries on.
  const report = (err) => {
    try {
      onError(err);
    } catch {
      // deliberately dropped, see above
    }
  };

  const onFrame = ({ event, data }) => {
    if (event !== "next" || data === "") return;
    const payload = JSON.parse(data);
    const refusal = refusalIn(payload);
    if (refusal !== null) {
      report(new Error(refusal));
      return;
    }
    delay = BACKOFF_START_MS;
    onEvent(payload.data);
  };

  (async () => {
    while (active) {
      const request = new AbortController();
      controller = request;
      try {
        // Armed before the fetch, so a request that never answers is as
        // dead as a stream that stops mid-flight.
        watch(request);
        const response = await fetch(ENDPOINT, {
          method: "POST",
          headers: { "Content-Type": "application/json", Accept: "text/event-stream" },
          body: JSON.stringify({ query: document }),
          signal: request.signal,
        });
        if (!response.ok || response.body === null) {
          throw new Error(await failureOf(response));
        }
        await readStream(response.body, (frame) => {
          watch(request);
          onFrame(frame);
        });
        if (active) report(new Error(`the stream ended, reconnecting in ${delay / 1000} s`));
      } catch (err) {
        if (active) report(err);
      } finally {
        // However this attempt ended, the response is finished with. A
        // refused status and a handler that threw both leave the body
        // unread, and an unread body holds its connection until the
        // collector reaches it while the loop opens the next one. The
        // abort closes it now, and it retires a watchdog that would
        // otherwise fire on an attempt that is already over.
        stopWatching();
        request.abort();
      }
      if (!active) return;
      await sleep(delay);
      delay = Math.min(delay * 2, BACKOFF_CAP_MS);
    }
  })();

  return () => {
    active = false;
    stopWatching();
    if (controller !== null) controller.abort();
  };
}

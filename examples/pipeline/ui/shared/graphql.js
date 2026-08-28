// graphql.js: the two apps' only way to the router (SR-40).
//
// query() sends a GraphQL document by fetch POST to /graphql on the page's
// own origin and surfaces GraphQL errors as thrown Errors. subscribe()
// opens a subscription over server-sent events on a fetch POST, the path
// the router documents, reads the stream frame by frame and reconnects
// with backoff until unsubscribed (AD-0014). Frames are "event: next" with
// a "data:" line, ":heartbeat" comments, and "event: complete".

const ENDPOINT = "/graphql";
const BACKOFF_START_MS = 1000;
const BACKOFF_CAP_MS = 8000;

const messageOf = (errors) => errors.map((e) => e.message).join("; ");
const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

export async function query(document, variables = {}) {
  const response = await fetch(ENDPOINT, {
    method: "POST",
    headers: { "Content-Type": "application/json", Accept: "application/json" },
    body: JSON.stringify({ query: document, variables }),
  });
  if (!response.ok) {
    throw new Error(`the router answered ${response.status} ${response.statusText}`);
  }
  const body = await response.json();
  if (Array.isArray(body.errors) && body.errors.length > 0) {
    throw new Error(messageOf(body.errors));
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

  const onFrame = ({ event, data }) => {
    if (event !== "next" || data === "") return;
    const payload = JSON.parse(data);
    if (Array.isArray(payload.errors) && payload.errors.length > 0) {
      onError(new Error(messageOf(payload.errors)));
      return;
    }
    delay = BACKOFF_START_MS;
    onEvent(payload.data);
  };

  (async () => {
    while (active) {
      controller = new AbortController();
      try {
        const response = await fetch(ENDPOINT, {
          method: "POST",
          headers: { "Content-Type": "application/json", Accept: "text/event-stream" },
          body: JSON.stringify({ query: document }),
          signal: controller.signal,
        });
        if (!response.ok || response.body === null) {
          throw new Error(`the router answered ${response.status} ${response.statusText}`);
        }
        await readStream(response.body, onFrame);
        if (active) onError(new Error(`the stream ended; reconnecting in ${delay / 1000} s`));
      } catch (err) {
        if (active) onError(err);
      }
      if (!active) return;
      await sleep(delay);
      delay = Math.min(delay * 2, BACKOFF_CAP_MS);
    }
  })();

  return () => {
    active = false;
    if (controller !== null) controller.abort();
  };
}

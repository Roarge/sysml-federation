# The check session

The demo is one container, and this directory is a project beside it that
checks the demo from outside: a Checkly project of thirty checks and monitors,
a compose profile that opens a tunnel to a running instance, an OpenTelemetry
collector the router exports to, and a trace viewer. The session is optional.
Nothing here runs unless it is asked for. A session needs a Checkly account
and its two credentials, and without them `docker run` runs the demo exactly
as before: the variable that hands the router a configuration file is unset,
the router's environment is what the supervisor sets, and nothing under
`checkly/` is built, run or read. The decision is
[AD-0031](../docs/decisions/AD-0031-an-optional-check-session.md), and every
check here is a verification case of its own in
[the model of the demo](../model/README.md).

Release 0.2.0 is the first published image that carries the router's tracing
opt-in. With `DEMO_IMAGE` unset, the compose file starts
`ghcr.io/roarge/sysml-federation`. `docker compose up` pulls it only when the
host holds no image under that name, so a host that pulled the demo before
release 0.2.0 still starts the older image, which ignores the opt-in and
exports no traces, until one `docker pull ghcr.io/roarge/sysml-federation`
replaces it. A local build can still run in place of the published image.
`make image` from the repository root builds `sysml-federation:dev`, and
`DEMO_IMAGE=sysml-federation:dev` in the environment or in `checkly/.env` makes
the compose file start that build instead.

## Running a session

The session reads its settings from the environment and from `checkly/.env`
when the file exists, which `.gitignore` excludes. Two are required and the
rest are optional.

| Variable | What it does |
|---|---|
| `CHECKLY_API_KEY`, `CHECKLY_ACCOUNT_ID` | the account the project is tested, deployed and destroyed in, both required for a session |
| `SESSION_WITHOUT_ACCOUNT` | `1` starts the stack, the tunnel and the subscription probe with no account at all, whatever the file carries, and deploys nothing |
| `DEMO_IMAGE` | the demo's image, `ghcr.io/roarge/sysml-federation` when unset, or `sysml-federation:dev` for the local build from `make image` |
| `LOG_LEVEL` | the demo's log level, passed through to its container |
| `TUNNEL_TOKEN` | the named tunnel's token. When set, the named tunnel runs in place of the quick one, and `DEMO_HOSTNAME` must be set with it |
| `DEMO_HOSTNAME` | the named tunnel's public hostname, `demo.sysml-federation.org` for the owner's. It also constructs the three hostname monitors |
| `CHECKLY_ALERT_EMAIL`, `CHECKLY_WEBHOOK_URL` | an alert channel each, constructed only when set |
| `OTEL_COLLECTOR_CONFIG` | `collector-local.yaml` by default, or `collector-checkly.yaml` for traces beside the check results |
| `CHECKLY_OTEL_API_KEY` | the account's tracing key, read by `collector-checkly.yaml` |
| `OTEL_INGEST_TOKEN` | any secret string, needed by `collector-checkly.yaml`, which guards its inbound port with it |
| `CHECKLY_DASHBOARD_SLUG`, `CHECKLY_STATUS_SLUG` | the dashboard's and the status page's subdomains, derived from the account id when unset |
| `CHECKLY_INCIDENTS`, `CHECKLY_MAINTENANCE` | `1` constructs the status page's automation rules and the weekly maintenance window, each a paid feature |
| `CHECKLY_PRIVATE_LOCATION_SLUG` | constructs a private location and moves every group onto it, a paid feature |
| `CHECKLY_PL_API_KEY` | the private location's own key, read by the compose file's `checkly-agent` service under the `checkly-private` profile, which the session script never starts and which has not been run |

One command starts everything.

```
bash checkly/scripts/session.sh up
```

On the host the script loads `checkly/.env`, refuses to start without the two
credentials unless `SESSION_WITHOUT_ACCOUNT=1`, chooses the named tunnel when
`TUNNEL_TOKEN` is set and the quick one otherwise, warns once when no alert
channel is configured, and hands over to `docker compose` with the profiles
that choice calls for. Five containers come up: the demo, one tunnel, the
collector, the viewer and the runner. The runner is a Node image with the
project mounted from this directory and its dependencies in a named volume, so
the dependencies are not written into the host's tree. It installs an HTTP
client and a certificate store on every start, about fifteen seconds and a
network dependency, because the pinned image carries neither and the probes
and the pings are `curl`.

Inside the runner the same script walks the session's twelve steps and logs
each under its number, `[1/12] installDependencies` through
`[12/12] destroyOnStop`. It prints the demo's public address once it has one,
whether the viewer answers through the tunnel, the verdict of the subscription
probe as `SSE_STREAMS=1` or `SSE_STREAMS=0`, the account the session runs as
and its plan, the name of the recorded test session, the dashboard's and the
status page's addresses after the deploy, and `heartbeat sent` every five
minutes until stopped. No credential is printed.

Stopping is `Ctrl-C` on the `up` command. Compose forwards the stop, and the
runner destroys the deployed project, removes the account variable it
published and exits with the test session's status, inside a grace period of
90 seconds. Then

```
docker compose -f checkly/compose.yml --profile checkly --profile quick down -v
```

removes the containers, the network and the runner's volume, with
`--profile named` in place of `quick` after a named-tunnel session.

Without an account, a contributor runs the checks against a local demo. From
`checkly/`, after `npm ci` and `npx playwright install chromium`, with the
demo answering on port 8080:

```
DEMO_URL=http://localhost:8080 SSE_STREAMS=1 npx playwright test --project chromium
```

runs the twelve browser specs one file at a time, which is how the shipped
configuration runs them everywhere, and `SSE_STREAMS=0` shows the seven that
skip without a stream. The three multistep specs are outside the shipped
configuration's `testMatch`, and a file argument on the command line does not
widen it, so they run under a copy of `playwright.config.ts` whose `testMatch`
reads `'*.multistep.spec.ts'`, passed with `--config`. Every mutating spec
puts the demo back as it found it, and the caption reads `capacity 1200,
bottleneck parse` afterwards. `npx tsc --noEmit` type-checks the project.
`npx checkly test` needs the account before it parses anything, and the
record below carries one load of the project through the CLI's own parser
without one.

## What runs

The thirty checks and monitors are listed in `__checks__/checks.json`, the
manifest the story checks are constructed from and the agreement test reads,
and each has a case in the model whose attributes equal its entry.

| id | kind | what it asserts | frequency | deployed |
|---|---|---|---|---|
| `router-version` | ApiCheck | `{ model { version } }` answers 200 with a version above zero | hourly | yes |
| `router-join` | ApiCheck | the join query for `PIPE-R1` answers the shipped text, a verdict and document number 1 | hourly | yes |
| `router-capacity-bottleneck` | ApiCheck | `PIPE-P1` answers a capacity and a bottleneck | hourly | yes |
| `router-introspection` | ApiCheck | the schema names `Verdict`, `Document`, `Node` and `Model` | hourly | yes |
| `router-playground` | ApiCheck | `/playground` answers 200 with HTML | hourly | yes |
| `router-root-redirect` | ApiCheck | `/` answers 302 to `/viewer/` | hourly | yes |
| `router-health-not-proxied` | ApiCheck | `/health/ready` answers 404 | hourly | yes |
| `refusals` | MultiStepCheck | a bound limit and a negative value are refused, and the version stands | every 6 hours | yes |
| `monitor-viewer` | UrlMonitor | `/viewer/` answers 200 | every 10 minutes | yes |
| `monitor-document` | UrlMonitor | `/document/` answers 200 | every 10 minutes | yes |
| `monitor-root` | UrlMonitor | `/` answers 200 after its redirect | every 10 minutes | yes |
| `session-heartbeat` | HeartbeatMonitor | a ping arrives every 10 minutes, with 5 minutes of grace | on the ping | yes |
| `demo-certificate` | SslMonitor | the hostname's certificate is valid, alerting 14 days before expiry | hourly | with a hostname |
| `demo-dns` | DnsMonitor | an A query for the hostname answers | every 10 minutes | with a hostname |
| `demo-tcp` | TcpMonitor | port 443 of the hostname accepts a connection | every 10 minutes | with a hostname |
| `us01-launch` | BrowserCheck | both apps open and request nothing outside the demo | daily | yes |
| `us02-read-the-model` | BrowserCheck | the text pane, the sketch with `parse` as the bottleneck, `PIPE-R1` failing at 1200 against 1500 | daily | yes |
| `us03-nothing-moves` | BrowserCheck | `ingest` at 3000 moves nothing | daily | test only |
| `us04-bottleneck-moves` | BrowserCheck | `parse` at 1700 moves the cut, `indexA` at 900 passes the requirement | every 6 hours | yes |
| `us05-tighten-limit` | BrowserCheck | the limit at 1000 passes, at 2500 fails, and the text follows | daily | test only |
| `us06-read-document` | BrowserCheck | the numbering, `PIPE-R1.4`'s row and `PIPE-R2` inconclusive | every 6 hours | yes |
| `us07-reorder-and-nest` | BrowserCheck | two moves renumber the document and leave the model alone | daily | yes |
| `us08-shape-the-document` | BrowserCheck | a heading, a paragraph, an exclusion and its restore | daily | yes |
| `us09-edit-from-document` | BrowserCheck | two document edits reach the viewer within two seconds | daily | yes |
| `us10-viewer-to-document` | BrowserCheck | a viewer edit reaches the document within two seconds | daily | yes |
| `us11-query-the-graph` | BrowserCheck | the join query, the schema and the playground | daily | yes |
| `us12-reset` | BrowserCheck | a reset returns both apps to the shipped state | daily | yes |
| `arithmetic-walk` | MultiStepCheck | the four rows of the capacity table hold through the router | every 6 hours | yes |
| `document-operations` | MultiStepCheck | the six document operations renumber the document and leave the model alone | daily | test only |
| `story-suite` | PlaywrightSuite | the twelve browser specs as one suite | daily | yes |

A check that is test only runs in the recorded test session and is never
deployed: two of the stories mutate and teach nothing new about the shipped
state as monitors, and the document walk mutates the document at length. The
three hostname monitors are constructed only when `DEMO_HOSTNAME` is set,
since a quick tunnel's hostname is not known until the session runs.

Every check belongs to one of five groups, and a group carries the locations,
the concurrency and the retry policy of every check in it. The router group
reads and never writes, so it runs from three locations, retries a failure
twice in the same region, and alerts on the second failed run. The viewer,
document and cross-app groups edit the demo's one in-memory state, and a
retried or a parallel run would edit it twice, so each runs from one location,
with no retries and no parallel locations, and alerts on the first failed run.
That is the whole of the guard. A group's concurrency of one governs the runs
a trigger or the API starts, which is how the session's tenth step runs them,
and nothing serialises the scheduled runs of two checks within or across the
three groups, nor the runs of a test session. The session group holds the
monitors under the same single-location policy. No check runs from parallel
locations anywhere. Every group is tagged `demo` and with its own name, and
every story check with `demo` and its story's identifier.

Two alert channels are constructed when their variables are set and never
otherwise. The email channel sends on failure, on recovery and on a
certificate nearing expiry, and not on a degraded run. The webhook channel
posts a JSON body carrying the alert's title and type, the check's name, the
start time and a link to the result. A session that sets neither has an empty
channel list, and the groups alert nobody.

A public dashboard shows every check tagged `demo`, with its 95th and 99th
percentiles, on pages of fifteen that refresh and turn every minute, at
`https://<slug>.checklyhq.com`. A status page shows one card, the demo, with
four services, Viewer, Document, Router and Session, at
`https://<slug>.checkly-status-page.com`. A service carries no reference to a
check: its status is set by an incident, by hand or by the automation rules,
and nothing in the code links a check to a service. Both are constructed only
when their slug or the account id is set, and a slug not given is
`sysml-federation-` followed by the first eight characters of the account id,
with `-status` appended for the page, because a slug is unique across every
account of the service and two accounts deploying this project must not claim
one address.

Three constructs are gated behind flags because they are paid features, so
that the project deploys on the free tier as it stands. `CHECKLY_INCIDENTS=1`
constructs one automation rule per service, tagged with its group's name so
that a failing check opens a partial-outage incident on its own service and no
other. `CHECKLY_MAINTENANCE=1` constructs a weekly window of five minutes on
Sunday at 03:00 UTC, from 6 September 2026, over every check tagged `demo`.
`CHECKLY_PRIVATE_LOCATION_SLUG` constructs a private location and moves every
group onto it and off its public locations, and the compose file carries the
container for it behind the `checkly-private` profile, with its own key in
`CHECKLY_PL_API_KEY`, which the session script does not start.

## The tunnel and streamed responses

The named tunnel is the route for a session that runs all twelve stories. It
needs a token, created on the owner's own zone, and the public hostname routed
to that tunnel, and the route from the hostname to `http://demo:8080` is
configured where the tunnel was created, so nothing of it is in the
repository. The
hostname is a CNAME to the tunnel, flattened at the edge, which is why the DNS
monitor asks for an A record. The quick tunnel needs nothing. It is given a
hostname under the tunnel provider's own domain for the life of the session,
and the runner reads it from the tunnel's metrics endpoint on port 2000 inside
the compose network.

The two differ in one thing that matters here. Quick tunnels are documented as
carrying no streamed response, so a page behind one is expected never to hear
the server-sent event that would redraw it, and the named tunnel is the route
for all twelve stories. The runner's fourth step settles it for each session
on a quick tunnel: it holds a subscription open, sends a mutation beside it,
and sets `SSE_STREAMS=1` if an event frame arrives within ten seconds and
`SSE_STREAMS=0` otherwise. Every
check whose page must learn of an edit from the stream, its own page or a
second one, reads that verdict and skips with the reason
`live updates need the named tunnel` rather than fail. Seven of the twelve
story checks do: `us03`, `us04`, `us05`, `us08`, `us09`, `us10` and `us12`.
The other five, and the walks, read their own page or the router directly.

A quick tunnel can also fail to open at all. The tunnel container asks the
provider's service for a hostname over HTTPS, and an access provider whose
resolver answers that service's name with a block page leaves the request
refused on a certificate that names the provider's own hosts. The record below
carries one such block. On such a network the named tunnel is the route.

## Traces

The router runs with tracing off and no exporter, which is what the image
sets. The session names a configuration file in
`SYSML_FEDERATION_ROUTER_CONFIG_PATH`, mounted at `/otel/router.yaml`, and the
router's own rule makes a file win over the environment. The file,
`otel/router.yaml`, turns tracing on with one exporter, the collector on the
compose network, samples every request, keeps the trace id of a request that
arrives with trace context, and defining that exporter is what keeps the
router's default one off. The router logs the file it read and then one line
for the tracer:

```
{"level":"info","msg":"Tracer enabled","service":"@wundergraph/router","service_version":"0.343.1","exporter":"http","endpoint":"http://otel-collector:4318","path":""}
```

That is the only exporter line, and the endpoint is the collector on the
compose network and nothing outside it.

The collector has two configurations. `otel/collector-local.yaml`, the
default, forwards every span to Jaeger and writes a line per batch to its own
log, and Jaeger's interface is published on the host at `localhost:16686`. A
trace arrives there a few seconds after the request: a lookup six seconds
after the request answered not found once, and the trace was there a moment
later. `otel/collector-checkly.yaml`, chosen with
`OTEL_COLLECTOR_CONFIG=collector-checkly.yaml`, does the same and adds a second
pipeline that keeps only the spans whose trace state carries the mark the
monitoring service's runners set, `checkly=true`, and forwards them to the
service's ingest endpoint with `CHECKLY_OTEL_API_KEY` as the whole value of the
authorization header. Each such span is then shown beside the check result
that caused it. The endpoint in the file is a data-residency host, and the
owner reads the one shown on their integration page.

The second configuration also opens a third receiver, on port 4320 inside the
compose network, under a bearer token, for spans that arrive from outside the
network and join the viewer's pipeline. That is the optional inbound side, and
it needs a second tunnel route to be reachable at all. The token is
`OTEL_INGEST_TOKEN`, and it must be set to any secret string whenever
`collector-checkly.yaml` is chosen, whether or not the inbound route is used,
because the token guards port 4320 and the collector refuses to start without
one. The session script says so and refuses the combination.

## The free tier's budget

Checkly's free tier, as its pricing page read on 14 September 2026, includes
10,000 API check runs and 1,000 browser check runs a month, ten uptime
monitors billed by number rather than by run, and round-robin scheduling
across at most three locations per check. The runs a check consumes in a month
are its executions multiplied by its locations when it runs in parallel, and
its executions alone when it does not, and every group here runs round robin.
Three multipliers apply on top: each request of a multistep check counts as a
run, a Playwright suite counts one run for every 30 seconds of its execution,
and every retry counts.

Over a month of 30 days the seven hourly API checks cost 5,040 runs. The two
deployed multistep checks run every six hours, 120 executions each: the
refusals walk sends four requests a run, 480 more, and the arithmetic walk
nine, 1,080 more, which puts the API side at 6,600 of 10,000, and at 6,820 in
a month of 31 days. The router group's two retries add runs only in an hour
that fails. On the browser side the ten deployed story checks, eight daily and
two every six hours, cost 480 runs, and the suite's 30 executions cost 30 runs
if each stays inside half a minute and 60 if each stays inside a minute, so
the browser side sits between 510 and 540 of 1,000. The monitors are four
without a hostname and seven with one, of the ten allowed.

## What is not done

The private location has not been run. Its construct, the group assignments
and the container are in place behind their flag and profile, and the route
needs a paid plan. Incident automation on the status page and the maintenance
window are written and gated for the same reason, and neither has been
deployed. Visual regression is not on the free tier and no check here takes a
screenshot to compare.

The checks carry no secret and the project stores none, because the demo has
none: no login, no key, no value a visitor could not read from the page.

While a session runs the demo answers at a public hostname without
authentication, for the session's duration. The demo holds no secret, so what
a visitor can do in that window is edit a value, and an edit that lands while
a check is reading can fail that check.

Two mutating checks scheduled at the same frequency can overlap, because the
service does not serialise them. A check that meets another check's edit
fails, and its reset restores the state. The first session's record is where
this shows, if it does.

## Verification record

The rows on the compose file and the stack were run from the repository root
with the demo image built from source with `make image`, and the first of
them with the published image as well. The rows on the project itself, its
construction, its type check, its specs and its API requests, were run from
`checkly/`, against a local container of the same build where one was
needed. Every row is of the branch as committed, except where a row says
which correction came after its run. The rows with a date have been observed.
The pending rows need an account, a token or a network that resolves the
tunnel service, and are the owner's to add when they have run the setup
steps.

| date | what | how | observed |
|---|---|---|---|
| 2026-09-13 | the demo alone through the compose file | `docker compose -f checkly/compose.yml up -d demo`, once with `DEMO_IMAGE=sysml-federation:dev` and once with the published image | `/health/ready` 404, `/viewer/` 200 and `/` 302 to `/viewer/` both times. The router process's environment, read from the container's process table, held nine variables with `TRACING_ENABLED=false` among them and no `CONFIG_PATH`, and the log had no configuration-file line |
| 2026-09-13 | the router with the file | the stack with `SYSML_FEDERATION_ROUTER_CONFIG_PATH=/otel/router.yaml` | the router logged the configuration-file line naming `/otel/router.yaml` and one `Tracer enabled` line, `exporter http`, `endpoint http://otel-collector:4318`, and no other exporter |
| 2026-09-13 | the stack under the quick profile | `DEMO_IMAGE=sysml-federation:dev SESSION_WITHOUT_ACCOUNT=1 bash checkly/scripts/session.sh up` | the host side printed `tunnel mode quick` and five containers started. The tunnel's request for a hostname was refused: the access provider's resolver answers the tunnel service's name with a block page whose certificate names the provider's own hosts, so the tunnel exited and the runner reported no hostname. The run recorded here waited under a count-bounded loop corrected before the commit, and the committed script, run later the same day, reported no hostname within two minutes. The collector logged its two receivers ready and the viewer its receivers and query server. `down -v` removed the containers, the network and the volume |
| 2026-09-13 | the subscription probe | the probe's own commands from the host against `http://localhost:8080` with the stack up | the mutation answered the part's id, the held stream carried a heartbeat and then `event: next` with `modelChanged` after a second, the reset answered, so `SSE_STREAMS=1` |
| 2026-09-13 | the runner's first four steps | the same stack with a stand-in for the quick tunnel's metrics endpoint answering the demo's own address over plain HTTP | `[1/12]` to `[4/12]` logged in order, the demo at `http://demo:8080`, the viewer answering 200, `SSE_STREAMS=1`, then the no-account stop, and the runner exited 0 within four seconds of the stop |
| 2026-09-13 | the join query in the viewer | the join query for `PIPE-R1` to `/graphql` with the stack up, then `/api/services` and `/api/traces?service=sysml-federation-router` on port 16686 | one service, `sysml-federation-router`. One trace of 13 spans: `query unnamed` as the server span, `HTTP - Read Body` and the five `Operation -` spans beneath it, and under `Operation - Execute` three `Engine - Fetch` spans for `model`, `document` and `capacity`, each with the client span of its HTTP call. The collector logged one batch of 13 spans |
| 2026-09-13 | the trace id round trip | the same query with a `traceparent` header carrying a trace id and a parent span id of the sender's choosing and `tracestate: checkly=true`, then `/api/traces/<that id>` | one trace under the sent id, 13 spans, the server span a child of the sent parent with the tag `w3c.tracestate = checkly=true`, and three fetch spans. The first lookup, six seconds after the request, answered not found, and the trace was there a moment later |
| 2026-09-13 | the second collector configuration | `otelcol-contrib validate` with stand-in values, then the collector started with `collector-checkly.yaml` on the compose network with its ingest exporter replaced by a debug exporter, and three spans posted as OTLP JSON, one with trace state `checkly=true`, one with none, one with `vendor=x` | validation exit 0. Port 4318 without a token 200. Port 4320 without a token 401, with a wrong token 401, with the right token 200. The viewer pipeline received all three spans and the ingest pipeline one, `marked-span` with `TraceState : checkly=true` |
| 2026-09-13 | the project's construction | the CLI's own loader and validator over the project, without credentials and with every flag set | thirty checks and monitors, five groups, no fatal diagnostic. With the flags, the rules, the window, the private location and its five group assignments as well |
| 2026-09-13 | the twelve browser specs | `DEMO_URL=http://localhost:18080 SSE_STREAMS=1 npx playwright test --project chromium` against a local container | 12 passed in 6.8 s on one worker, and the container at capacity 1200, bottleneck parse, afterwards |
| 2026-09-13 | the specs without a stream | the same with `SSE_STREAMS=0` | 7 skipped with the reason `live updates need the named tunnel`, `us03`, `us04`, `us05`, `us08`, `us09`, `us10` and `us12`, and 5 passed |
| 2026-09-13 | the three multistep specs | under a copy of the configuration matching `*.multistep.spec.ts` against the same container | 3 passed in 1.4 s, the container back at its shipped state |
| 2026-09-13 | the seven API checks | each request sent with `curl` against a local container on port 18080, with the check's assertions applied to the answer | every assertion held: the version 1, the join answer's text, verdict `FAIL` and document number `1`, `PIPE-P1` at capacity 1200 with the bottleneck `PIPE-S2`, the four type names in the schema, `/playground` 200 with `<html`, `/` 302 to `/viewer/`, `/health/ready` 404 |
| 2026-09-14 | the contributor commands as this file gives them | `npx tsc --noEmit`, the twelve specs with `SSE_STREAMS=1` and with `SSE_STREAMS=0`, the multistep specs under `multistep.config.ts`, a copy of `playwright.config.ts` differing in its `testMatch` line alone, and the seven API requests, all against a local container on port 18080 | the type check clean. 12 passed in 6.9 s, then 7 skipped and 5 passed in 2.7 s, then 3 passed in 1.2 s. Every API assertion held. The container answered capacity 1200, bottleneck parse, and the shipped verdicts at model version 26 afterwards |
| 2026-09-14 | the checks after the bottleneck and introspection assertions changed, and the collector configurations after their exporters were renamed | the twelve specs with `SSE_STREAMS=1`, the three multistep specs under the copied configuration, and the seven API requests, all against a local container on port 18080, then `validate` with the pinned collector image over both configurations with stand-in values, and each configuration started once | 12 passed in 6.8 s, 3 passed in 1.3 s, and every API assertion held, `PIPE-P1`'s first bottleneck `PIPE-S2` and the whole name field for `Node` and `Model` among them. The container answered capacity 1200, bottleneck parse, afterwards. Both configurations validated at exit 0, and each started to `Everything is ready` with no deprecation line in its log, where the `otlp` exporter alias had drawn one before |
| 2026-09-14 | a quick-tunnel session | `bash checkly/scripts/session.sh up` from a network whose resolver answers the tunnel service | the hostname read from the tunnel, the viewer through it, and the probe's verdict |
| 2026-09-14 | a named-tunnel session | `TUNNEL_TOKEN` and `DEMO_HOSTNAME` set | the viewer through the hostname, `SSE_STREAMS=1`, all twelve story checks run |
| 2026-09-14 | the recorded test session and the suite session | steps 7 and 8 with an account | the two sessions in the account, every check once, the test-only checks among them |
| 2026-09-14 | the deploy and destroy cycle | steps 9 and 12 with an account | the project deployed for the session, the heartbeat's ping address read back, the dashboard and the status page reachable, and nothing left in the account after the stop |
| 2026-09-14 | the alerts received | a failing check with `CHECKLY_ALERT_EMAIL` or `CHECKLY_WEBHOOK_URL` set | the failure and the recovery arriving on the channel |
| 2026-09-14 | the trace beside a check result | `CHECKLY_OTEL_API_KEY` set and `OTEL_COLLECTOR_CONFIG=collector-checkly.yaml` | the router's spans beside the check result in the account |

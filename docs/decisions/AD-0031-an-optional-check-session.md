# AD-0031 An optional check session through a tunnel, with traces

Status: accepted. Date: 2026-09-13.

## Context

The demo is one container that answers on one port of the machine it runs on.
A monitoring service runs elsewhere, on runners of its own, and cannot reach a
port on a laptop. The two had never met.

Until this record the repository carried no browser automation. Behaviour that
lives in browser JavaScript was demonstrated against a scripted checklist
rather than tested (AD-0017), and the two-second bound on a live update was
demonstrated for the same reason (AD-0014). SC-01 forbade a dependency that
drives a browser in the product code, which is the right rule for a Go
repository meant to be read without setup. It left the twelve stories with no
check that ran on anyone's schedule.

SR-03 kept the router off the network. The image sets the variables that turn
the router's telemetry off (AD-0013), so a request's trace went nowhere, and
when a check failed there was nothing to read beyond the check's own log. The
router is the one place a request is seen end to end, with the three fetches it
fans out into. It is also the only service in the container that can emit a
span without a new dependency.

## Decision

We will add a compose profile beside the demo, under `checkly/`, that starts a
Cloudflare tunnel, an OpenTelemetry collector, a trace viewer and a runner. The
runner tests, deploys and destroys a Checkly project against this instance. It
runs the browser specs and records the run as a session, deploys the project
for as long as the stack runs, and destroys it when the stack stops. The tunnel
is a named one on the operator's own hostname when a token is supplied, and a
quick one on a hostname the tunnel provider assigns otherwise.

The router's tracing becomes an operator opt-in. A configuration file is handed
to the router only when `SYSML_FEDERATION_ROUTER_CONFIG_PATH` names one, and
the file's values then govern the router's telemetry, because the router's own
rule makes a file win over the environment. The session names a file that
exports to the collector beside the demo, and the collector forwards every span
to the viewer. When an ingest key is supplied, it also forwards to the
monitoring service, but only the spans a check marked as its own.

The session runs on the operator's own Checkly account, supplied as an API key
and an account id in `checkly/.env`, because the maintainer's credentials
cannot ship inside a public image.

Without credentials nothing changes. `docker run` runs the demo as before, the
variable is unset, the router's environment is what the supervisor sets, and
nothing under `checkly/` is built, run or read.

## Alternatives considered

A hosted demo, one instance running somewhere public that the monitoring
service could reach without a tunnel. It would have made the session a fixture
rather than a profile. It loses because it needs an account and a bill nobody
asked for, and because the repository's claim is that one `docker run` on one
machine is the whole demo.

A private location, the monitoring service's runners placed beside the demo so
that no tunnel is needed. That is the cleaner topology, and the compose file
carries the container for a private location behind a profile of its own. It
loses as the route
because it needs a paid plan, and the project is kept deployable on the free
one.

Instrumenting the Go services, so that the three subgraphs carried spans of
their own and a trace showed each resolver. It loses because it is a dependency
SC-01 forbids in the product code, and because the router already sees every
request and every fetch it makes. The router's own exporter is the source of
spans, and the trace of one request shows the three fetches as the router made
them.

## Consequences

SR-03 and SC-01 are amended. SR-03 carries the configuration-file opt-in as a
clause of its statement and a fourth acceptance criterion, and SC-01 permits
`go-yaml` for the one test that reads the compose file and places the check
project outside the product.

AD-0013, AD-0014 and AD-0017 are amended. The telemetry record says how a file
governs what the environment turned off, and the two records that said the
repository carried no browser automation now say where it lives and what runs
it.

A Node project lives under `checkly/`, outside the product. Browser automation
exists in the repository, and it runs on the monitoring service's runners
rather than in `make check`, so the Go gate keeps its shape and its timing.

`go-yaml` becomes a direct dependency of the module, for one test that reads
the compose file and holds the demo service outside every profile.

The session on a quick tunnel probes whether a subscription's events cross the
tunnel before the checks run. When none arrives, the seven story checks that
depend on an event reaching the page skip, with the reason in their log. Quick
tunnels are documented as carrying no streamed response, so on one the seven
are expected to skip and the named tunnel on the operator's own zone is the
route for all twelve. The record in `checkly/README.md` says which route each
recorded run used.

While a session runs the demo answers at a public hostname without
authentication, for the session's duration. The demo holds no secret, so what
a visitor can do in that window is edit a value and fail a check.

## Requirements affected

SR-03, SR-48, SC-01

## Sources

[The check session](https://github.com/Roarge/sysml-federation/blob/main/checkly/README.md) for the setup, the checks and the verification record. [The model of the demo](https://github.com/Roarge/sysml-federation/blob/main/model/README.md) for the session composite and the check cases, in which every live check is a case of its own. The telemetry record (AD-0013), the version events record (AD-0014) and the vanilla web apps record (AD-0017), each amended by this decision.

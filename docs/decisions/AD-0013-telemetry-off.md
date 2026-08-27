# AD-0013 Telemetry disabled by environment baked into the image

Status: accepted. Date: 2026-08-27.

## Context

The README says that the router can be started from a pre-built configuration
file with no network dependency at all, and presents that as what makes the
stack air-gappable for defence, rail, energy and medical device work. The
research sweep of 2026-08-26 found the sentence false as the router ships.
Since router 0.215.0 an anonymous usage tracker is on by default and is gated
on neither the graph token nor the static execution config. It sends
`router_base_config`, `router_execution_config` and a per-minute
`cosmo_router_uptime` event to PostHog at eu.i.posthog.com. The payload
carries the host name, the router version, commit and build date, an instance
id, the cluster name and the output of `git remote get-url origin` run in the
router's working directory (C-04). It is switched off only by `DO_NOT_TRACK=1`
or `COSMO_TELEMETRY_DISABLED=true` in the environment, or by a Go option that
a router run as a child process cannot reach. There is no YAML key, no
documentation page mentions it, and the only mention in the vendor's
repository is `router/.env.example`.

Two further settings matter less. `telemetry.tracing.enabled` and
`telemetry.metrics.otlp.enabled` both default to true, with default exporters
pointing at cosmo-otel.wundergraph.com when none is configured. With an empty
token the router itself marks those exporters disabled, so setting
`TRACING_ENABLED` and `METRICS_OTLP_ENABLED` to false is belt and braces
rather than a requirement (C-02, C-05).

Whether a router with all of these set opens no outbound connection at all
follows from the code paths read. It was never confirmed by running the image
with networking removed, and posthog-go's feature flag polling was not examined
(C-07). Failures to reach PostHog are debug-logged and non-fatal, so the router
reaches readiness whether or not it is trying to reach the tracker, and
readiness therefore proves nothing about attempts.

The router runs as a child process from the copied vendor binary, driven by
configuration and environment alone (AD-0010, P2). The plan chose to set the
variables in the image and the brief recorded that as P5. C-85 queues the
README correction for the public phase: the sentence gains the qualification
and names the image's variables. What remained to record is the mechanism and
where it lives.

## Decision

We will disable every outbound path of the router by setting `DO_NOT_TRACK=1`,
`COSMO_TELEMETRY_DISABLED=true`, `TRACING_ENABLED=false` and
`METRICS_OTLP_ENABLED=false` as `ENV` instructions in the Dockerfile, so that
the variables are part of the image rather than of any launch command, and we
will set no graph token, so that the router's own token check disables the
Cosmo Cloud exporters, schema usage tracking and persisted operations as well
(C-02). The supervisor passes the four variables into the router child's
environment beside `LISTEN_ADDR`, `PLAYGROUND_PATH` and
`EXECUTION_CONFIG_FILE_PATH`, and SR-03 states them as an obligation on the
image and the router rather than as a rationale.

## Alternatives considered

Relying on the empty token alone, with the vendor's defaults otherwise in
place. That disables the default OTLP exporters and the control plane
self-registration but leaves the usage tracker running, so the air-gap claim
would be false as shipped (C-02, C-04, C-85).

A key in the router's configuration file. None exists for the usage tracker,
so a YAML file could not carry the setting even if the image shipped one
(C-04).

The Go option `WithDisableUsageTracking()`. It is reachable only from a router
embedded as a library, which AD-0010 rejected on the module's pseudo-versions,
its required `replace` block and its two API breaks (C-18).

Supplying the variables on the launch command. D5 fixes the launch line as one
`docker run` with one flag, and a variable the visitor has to remember is one
that will be forgotten, which is what "baked into the image" in P5 rules out.

## Consequences

The container makes no outbound connection while it runs, which is what V1
Context states and SR-03 requires, and the README sentence can be qualified
rather than withdrawn. The four names appear in the Dockerfile, in the
supervisor's environment for the child and in the public text that describes
static composition (C-85, C-88), so an organisation building its own image
around the adapter knows what to set.

The cost is a dependency on two variable names the vendor does not document.
They were found by reading `router.go` and `track.go` at 0.343.1. A later
release could rename them, add a second tracker or change the default, and
nothing in the vendor's documentation would announce it. The router version
is pinned together with wgc (C-16), and every bump means rereading the
tracker's code path, maintenance the demo would not otherwise carry.

The claim is only as good as its verification, and readiness is not it. SR-03
therefore has three parts: a test that the router process's environment
contains the four variables and that the execution configuration names only
loopback addresses, an analysis of the code paths in C-07, and a demonstration
that runs the container under `docker run --network none` at debug log level
and watches for connection attempts. The demonstration is C-07's spike, and it
must pass before the air-gap sentence goes public.

wgc itself sends usage events to the same host unless the same two variables
are set (C-06). Composition is a maintainer step on a connected machine
(AD-0012), so that is a habit for the maintainer rather than a property of the
image, and this record does not cover it.

## Requirements affected
SR-03

## Sources
The design brief P5 and D5. The constraints list C-02, C-04, C-05, C-06, C-07, C-16, C-18, C-85 and C-88. The requirements list SR-03. The design-phase plan, decision 5. The engineering log, design phase planning, the findings that changed the design. The sceptic verdicts on the research, the offline verdict. The architecture description V1 Context and V4 Deployment.

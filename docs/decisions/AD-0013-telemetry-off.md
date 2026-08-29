# AD-0013 Telemetry disabled by environment baked into the image

Status: accepted, amended once the image was built. Date: 2026-08-27.

Amendment, 2026-08-29: the decision as first accepted named four `ENV` lines
for the image and left `PROMETHEUS_ENABLED=false` to the router child's
environment alone, so the scrape endpoint stayed open for anyone who ran the
router binary out of the image directly. The image now carries that variable
as a fifth `ENV` line beside the four, and it is no part of SR-03.

## Context

The README says that the router can be started from a pre-built configuration
file with no network dependency at all, and presents that as what makes the
stack air-gappable for defence, rail, energy and medical device work. As the
router ships, the sentence is false. Since router 0.215.0 an anonymous usage
tracker is on by default and is gated on neither the graph token nor the
static execution config. It is switched off only by `DO_NOT_TRACK=1`
or `COSMO_TELEMETRY_DISABLED=true` in the environment, or by a Go option that
a router run as a child process cannot reach. There is no YAML key, no
documentation page mentions it, and the only mention in the vendor's
repository is `router/.env.example`.

Two further settings matter less. `telemetry.tracing.enabled` and
`telemetry.metrics.otlp.enabled` both default to true, with default exporters
pointing at the vendor's collector when none is configured. With an empty
token the router itself marks those exporters disabled, so setting
`TRACING_ENABLED` and `METRICS_OTLP_ENABLED` to false is belt and braces
rather than a requirement.

Whether a router with all of these set opens no outbound connection at all
was inferred rather than observed. It was not confirmed by running the image
with networking removed. Failures to reach the tracker are debug-logged and
non-fatal, so the router reaches readiness whether or not it is trying to
reach it, and readiness therefore proves nothing about attempts.

The router runs as a child process from the copied vendor binary, driven by
configuration and environment alone (AD-0010). The variables are set in the
image. The README's air-gap sentence gains the qualification and names the
image's variables. What remained to record is the mechanism and
where it lives.

## Decision

We will disable every outbound path of the router by setting `DO_NOT_TRACK=1`,
`COSMO_TELEMETRY_DISABLED=true`, `TRACING_ENABLED=false` and
`METRICS_OTLP_ENABLED=false` as `ENV` instructions in the Dockerfile, so that
the variables are part of the image rather than of any launch command, and we
will set no graph token, so that the router's own token check disables the
Cosmo Cloud exporters, schema usage tracking and persisted operations as
well. The supervisor passes the four variables into the router child's
environment beside `LISTEN_ADDR`, `PLAYGROUND_PATH` and
`EXECUTION_CONFIG_FILE_PATH`, and SR-03 states them as an obligation on the
image and the router rather than as a rationale.

A fifth `ENV` line sits beside those four, so the Dockerfile carries five.
`PROMETHEUS_ENABLED=false` closes the scrape endpoint the router opens on
loopback by default, which the supervisor also closes in the child's
environment. It is a listener rather than an outbound path, so it is no part
of SR-03, and its place in the image is for anyone who runs the router binary
out of the image directly, where nothing builds an environment on their behalf.

## Alternatives considered

Relying on the empty token alone, with the vendor's defaults otherwise in
place. That disables the default OTLP exporters and the control plane
self-registration but leaves the usage tracker running, so the air-gap claim
would be false as shipped.

A key in the router's configuration file. None exists for the usage tracker,
so a YAML file could not carry the setting even if the image shipped one.

The Go option `WithDisableUsageTracking()`. It is reachable only from a router
embedded as a library, which AD-0010 rejected on the module's pseudo-versions,
its required `replace` block and its two API breaks.

Supplying the variables on the launch command. The launch line is one
`docker run` with one flag, and a variable the visitor has to remember is one
that will be forgotten, which is what baking them into the image rules out.

## Consequences

The container makes no outbound connection while it runs, which is what the
context view states and SR-03 requires, and the README sentence can be
qualified rather than withdrawn. The four names appear in the Dockerfile, in
the supervisor's environment for the child and in the public text that
describes static composition, so an organisation building its own image
around the adapter knows what to set.

The router also serves a Prometheus scrape endpoint, on `127.0.0.1:8088` unless
it is told otherwise. That is a different mechanism from the exporters this
record disables, since a scrape endpoint waits to be read and opens nothing
outbound, so the air-gap claim never rested on it either way. It is switched off
by `PROMETHEUS_ENABLED=false` in the child's environment and again in the
image's `ENV` block, because nothing here reads it, and a port with no reader is
one more thing a port table has to explain.

The cost is a dependency on variable names the vendor does not document, and
there are three of them: `DO_NOT_TRACK`, `COSMO_TELEMETRY_DISABLED` and the
`PROMETHEUS_ENABLED` above. No documentation page names any of the three, and
the router binary's help lists no environment variables at all, so a release
that renames one, changes a default or adds a second tracker would announce it
nowhere. Nor would anything here notice. The test that guards SR-03 asserts the
environment the supervisor builds for the child, not what the router does with
it, so a name the router has stopped reading passes it unchanged while the
container goes back to reporting usage or to opening a scrape port. The router
version is pinned together with wgc, and every bump means reading the release
notes and `router/.env.example` for all three names, then starting the
container at debug level and confirming from its log that usage tracking is
off and that nothing listens on `127.0.0.1:8088`. AD-0010 sets a fourth
undocumented name for a purpose of its own and carries the same obligation, so
the two are checked in one pass. This is maintenance the demo would not
otherwise carry.

The claim is only as good as its verification, and readiness is not it. SR-03
therefore has three parts: a test that the router process's environment
contains the four variables and that the execution configuration names only
loopback addresses, an analysis of the paths that could open a connection, and
a demonstration that runs the container under `docker run --network none` at
debug log level and watches for connection attempts. The demonstration is a
spike, and it must pass before the air-gap sentence goes public.

wgc itself sends usage events unless the same two variables
are set. Composition is a maintainer step on a connected machine
(AD-0012), so that is a habit for the maintainer rather than a property of the
image, and this record does not cover it.

## Requirements affected
SR-03

## Sources
The vendor's `router/.env.example`, which is the one place `DO_NOT_TRACK` and `COSMO_TELEMETRY_DISABLED` are named, and the vendor's pages on tracing and metrics exporters and on running without a graph token. [What the research overturned](../articles/03-what-the-research-overturned.md) for the finding as it stands and for what remains inferred rather than run, and [The demo as it shipped](../articles/10-the-demo-being-built.md) for the five `ENV` lines and the air-gap demonstration.

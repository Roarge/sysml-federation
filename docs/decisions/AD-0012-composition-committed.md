# AD-0012 Composition as a maintainer step with committed output and a drift test

Status: accepted. Date: 2026-08-27.

## Context

The router needs an execution configuration composed from the three
subgraph schemas. `wgc router compose -i graph.yaml -o config.json`
produces one locally, and the vendor's page says the command "does not
interact with the control plane and completely runs locally". The router
loads the file through `EXECUTION_CONFIG_FILE_PATH` and takes the
static-config branch before any poller is consulted. The vendor's
compose page says "it is recommended to not use this for production", its
CLI source directs production users to `router fetch`, and the router logs
"Not recommended for Production" at every start without a token. The demo
has no control plane to
fetch from, and the README presents the static path as what makes the
stack air-gappable, so the vendor's wording is quoted wherever the public
text describes it.

The question is where and when composition runs. The Go composition
library was removed from the Cosmo repository on 2026-05-06, so composition
means wgc or the TypeScript package, both of which need Node. The
packaging research had suggested composing at container start or in a Go
build step, and both suggestions died with that removal.
CI runs `go build` and `go test` with read-only permissions and
image publishing on version tags, and no Node toolchain (SC-07). wgc sends
usage events unless `DO_NOT_TRACK=1` or `COSMO_TELEMETRY_DISABLED=true` is
set, it declares no Node range, and whether it needs no network at all with
schema files only was not established.

The output carries a compatibility version that the router checks at
start, wgc hard-codes it at 1, and the router 0.313.0 subscription overhaul
renamed configuration keys, so the router tag and the wgc version are
pinned and bumped together. The README's third condition for a
projection, that the contract between producer and consumer is checked
mechanically before deployment, is met by the repository rather than by the
demo, through a test that fails when a subgraph schema and the composed
configuration drift apart.

## Decision

We will run composition as a maintainer step on a connected machine with
the two telemetry variables set, from the compose input at
`examples/pipeline/graph.yaml` to `examples/pipeline/config.json`, commit
both, copy the configuration into the image at `/app/config.json`, and
guard the committed output with a Go test that parses it and compares each
embedded subgraph schema with the schema file it came from (SR-42). The
compose input names the three subgraphs with
loopback routing URLs on ports 3011 to 3013, a schema file for each, and
`ws` with subprotocol `graphql-transport-ws` for the two that carry
subscriptions, which is the subprotocol the subgraphs serve. The third is
left to the tool's default, since it has no subscriptions to negotiate.

## Alternatives considered

Composing at container start, so the supervisor could template the routing
URLs itself. The packaging research raised it with `composition-go`, which
no longer exists, and the remaining composers need Node, which the
distroless image does not have.

Composing in the build, as a Node stage in the Dockerfile running
`npx wgc router compose`. The packaging research listed it as one of two
places composition could run in CI. It would put Node into every image
build, with a tool that declares no Node range and sends usage events
unless the two variables are set in CI. A maintainer step
on a connected machine is what makes wgc's unverified offline behaviour
and Node range not load-bearing, which is why no spike is opened for them.

Fetching the configuration from a control plane with `router fetch`, the
vendor's production recommendation. There is no control plane in
the demo, and a fetch at start would be a network dependency in a stack
presented as air-gappable.

## Consequences

CI stays Go only. The drift test is an ordinary unit test, so a schema
change without a recompose fails on the pull request without Node in CI
(SC-07), and the maintainer's recompose is a demonstration rather
than a test (SR-42). The committed configuration is what the image ships,
so what the router runs is reviewable in a pull request as a diff.

The maintainer's machine becomes the one place composition happens, with
Node, a pinned wgc and the two telemetry variables. A contributor who edits
a schema and cannot run wgc sees the drift test fail and has to ask the
maintainer to recompose, which is a cost the demo accepts. The two file
locations are fixed by the allowlist: `examples/*/*.yaml` and
`examples/*/*.json` reach one level below `examples/`, so the files sit at
`examples/pipeline/` rather than in a `router/` subdirectory (SC-04).

The router tag and the wgc version move together, so a router bump
is a recompose and a rebuild, not a tag change alone. The configuration
bakes in the loopback routing URLs of the single-container layout, so any
other layout needs its own compose input, which is one reason the compose
file among AD-0011's alternatives was not taken.

Whether Cosmo composition accepts the capacity service's nested `@requires`
is the first spike of the implementation phase, and the compose
step is where it is exercised. The public text carries the vendor's
production caveat beside its description of static composition and says
what the demo does about it: no control plane exists to fetch from, and
the configuration is committed and tested for drift.

## Requirements affected
SR-42, SC-07

## Sources
The vendor's pages on `wgc router compose` and on the router's static execution configuration, and the commit that removed `composition-go` from wundergraph/cosmo. [Five spikes before the first line](../articles/09-five-spikes-before-the-first-line.md) for the composition runs and where the embedded schema sits in the output, and [Five views and twenty-six decisions](../articles/06-five-views-and-twenty-six-decisions.md) for the composition view.

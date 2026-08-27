# AD-0010 The router as a child process from the copied binary

Status: accepted. Date: 2026-08-27.

## Context

The demo ships as one image started by one `docker run`, and the
Cosmo router has to be inside it. The vendor publishes the router as an
image, `ghcr.io/wundergraph/cosmo/router`, built with `CGO_ENABLED=0` onto
distroless static for linux/amd64 and linux/arm64, with the binary at
`/router` and about 40 MB compressed. It is also a Go module, and
the module can in principle be embedded:
`core.NewRouter`, `WithStaticExecutionConfig` and `Router.Start` are
exported, and the documented custom-binary pattern is `routercmd.Main()`
with blank-imported modules.

What the module does not offer is a stable footing. Its release tags are of
the form `router@x.y.z`, which Go tooling cannot resolve, so only commit
pseudo-versions exist. The official router-examples README requires a block
of OpenTelemetry `replace` directives and says compatibility is not
guaranteed without them. The module API broke at 0.188.0 and 0.278.0, a new
module system is announced for the next major release, `WithModulesConfig`
takes `map[string]interface{}`, and no docs page describes direct
construction through `core.NewRouter`. No Go API stability promise is
published anywhere. Embedding would also pull graphql-go-tools, the process extension library and
the NATS, Kafka and Redis clients into this module's dependency graph, and
the empty-interface signature would sit in hand-written code, which the
repository's own rule forbids outside a reviewed exception.

Copying the binary has its own conditions. `COPY --from` accepts an image
reference, the router is statically linked, and copying `/router` out of
the pinned image is legitimate under Apache 2.0 provided the LICENSE
travels with it. The router binary has no probe subcommand, so
a container built on it alone could carry no `HEALTHCHECK`. Whether
`COPY --from` resolves the target platform's variant in a multi-platform
build was unverified when this record was written. A Go supervisor that is
PID 1 and the router's direct parent reaps it through `os/exec` without tini.
Reaping of grandchildren is only likely, and the two known sources of
grandchildren are the usage tracker and the router's extension mechanism.

## Decision

We will run the Cosmo router as a child process of the supervisor, from
the binary copied out of `ghcr.io/wundergraph/cosmo/router:0.343.1` with a
`COPY --from` stage, driven by environment and the committed execution
configuration alone. The image copies `/router` and fetches the
vendor's LICENSE at the pinned tag with a pinned checksum to sit beside it,
the supervisor starts the child with
`LISTEN_ADDR=127.0.0.1:3002`, `EXECUTION_CONFIG_FILE_PATH=/app/config.json`,
`PLAYGROUND_PATH=/playground` and the four telemetry variables of SR-03,
and waits for `/health/ready` before opening the published port.

## Alternatives considered

Embedding the router as a Go library through `core.NewRouter` and
`WithStaticExecutionConfig`, with the subgraphs as in-process servers. One
static binary, one process, the smallest image, and programmatic control of
playground, telemetry and listen address. It lost on the module's footing:
pseudo-versions only, the `replace` block, two API breaks, an
announced rewrite, an undocumented construction path and no stability
promise, plus an empty-interface signature and a dependency graph that
would carry message brokers the demo never uses.

The vendor's custom-binary form, `routercmd.Main()` with custom modules,
which is the supported shape of embedding. It takes
over `main`, reads its configuration from file and environment as the stock
binary does, and offers the demo nothing the copied binary does not, while
carrying every dependency and version problem of the library route.

The router's extension mechanism under Cosmo Connect, or standalone gRPC
services, as the way to co-locate the subgraphs with the router. Officially
supported, and the router would supervise the extensions itself. Neither
supports subscriptions, and the two apps depend on live version events
(AD-0014), so this route would have replaced live push with polling.

## Consequences

The router is a vendor artefact inside the image, pinned by tag and bumped
by rebuilding. Nothing in the Go module depends on it, so the router's Go
API can change without touching this repository, and a router upgrade is a
change to the Dockerfile and to the wgc version composed with it, which
must move together. The licence obligation is explicit: the image
carries the router's LICENSE next to `/router` (SR-07), and `NOTICE` names
the router with its version (SR-08).

The cost is a second process to start, watch and stop. The supervisor is
the parent of a child it cannot inspect from the inside, so readiness is
read from `/health/ready` and the `healthcheck` subcommand of
AD-0011 probes on the router's behalf. What `/health/ready`
reports before the subgraphs answer is a spike, and the supervisor's
ordering, subgraphs first, router last, published port after readiness, is
chosen so as not to depend on the answer.

The router sets the floor of the image size, around 40 MB compressed of
SR-06's 80 MB budget. SR-05 rests on another spike: if `COPY --from`
does not resolve the target platform, the arm64 image would carry an amd64
router and fail on Apple silicon, which is why the arm64 layer is inspected
before the first tag. A third settles whether the copied binary starts
from environment alone in a distroless image with no YAML file present.
All three are reported in
[Five spikes before the first line](../articles/09-five-spikes-before-the-first-line.md).

Grandchild reaping stays likely rather than verified. The design
removes the two known sources, since the usage tracker is disabled by SR-03
and no router extension is loaded, and no tini is added.

## Requirements affected
SR-03, SR-05, SR-06, SR-07

## Sources
The Cosmo router image and its LICENSE at 0.343.1, the router module's release tags and the `replace` block its examples repository requires, and the vendor's pages on running the router from environment and a static execution configuration. [Five views and twenty-six decisions](../articles/06-five-views-and-twenty-six-decisions.md) for the runtime and deployment views, and [What the research overturned](../articles/03-what-the-research-overturned.md) for the embedding claim as it was checked.

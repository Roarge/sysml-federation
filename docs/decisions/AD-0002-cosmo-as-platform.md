# AD-0002 Cosmo as the federation platform

Status: accepted. Date: 2026-08-27.

## Context

AD-0001 chose federation, and the README is plain that federation is a
platform problem rather than a library problem. Composition, breaking-change
checks, a registry that knows what is deployed and a planner in front all
have to come from somewhere, and a plain GraphQL server supplies none of
them. The README also fixes who the platform has to suit, organisations
with fewer than twenty-five engineers, and among them the defence, rail,
energy and medical device work that has to run air-gapped.

Cosmo's licence is the first fact. The wundergraph/cosmo monorepo carries
one licence file, the standard Apache License 2.0 text, with no
per-directory licence and no enterprise, commercial or source-available
exception, and the router module, wgc 0.130.1 and `@wundergraph/composition`
0.63.3 all declare Apache-2.0 in their package metadata (C-19). The vendor
states the platform runs without licence fees or feature gates, and "will
always remain so" is a promise rather than a licence term, so the versions
are pinned. The README describes the main alternative as being under a
licence the OSI does not recognise as open source, and the vendor
comparison the research quotes names it as Apollo's router under ELv2.

The second fact is that the router runs without a control plane. `wgc
router compose` composes locally and "does not interact with the control
plane", the router loads the output from `EXECUTION_CONFIG_FILE_PATH`, and
`Start()` takes the static-config branch before any poller is consulted
(C-01). No graph token is required with a static config, and with the
token empty the default Cosmo Cloud exporters are disabled (C-02). Two
vendor caveats sit beside that. The compose page says "it is recommended
to not use this for production" and the router logs "Not recommended for
Production" at every start without a token (C-03). Since router 0.215.0 an
anonymous usage tracker has been on by default, gated on neither the token
nor the static config and switched off only by `DO_NOT_TRACK=1` or
`COSMO_TELEMETRY_DISABLED=true` (C-04). The README's "no network dependency
at all" is true only with those set, and C-85 and C-88 queue the public
wording that carries both caveats rather than going around them.

Two more facts shape how the platform is used rather than whether. The Go
composition library was removed from Cosmo on 2026-05-06, so composition
needs Node and happens on the maintainer's machine with its output
committed (C-17, P4). The router's Go module has only commit
pseudo-versions and has broken its API twice, so the router is a copied
binary driven by configuration rather than a library (C-18, P2). The
README's deeper reason for Cosmo is that its subgraphs can now be compiled
to protobuf and served over gRPC with GraphQL kept as the schema language,
which the README reads as an admission that the valuable part was never the
wire format.

## Decision

We will use WunderGraph Cosmo as the federation platform, pinned to router
0.343.1 and wgc 0.130.1 and bumped together, run from a static execution
configuration with no control plane, no graph token and telemetry switched
off by `DO_NOT_TRACK=1`, `COSMO_TELEMETRY_DISABLED=true`,
`TRACING_ENABLED=false` and `METRICS_OTLP_ENABLED=false` baked into the
image (P5). The router binary is copied out of the vendor's image with its
Apache-2.0 licence text beside it and started as a child process (P2),
composition is a maintainer step whose output is committed and guarded by a
Go test that compares each embedded schema with the schema file it came
from (P4), and all three subgraphs stay standard GraphQL subgraphs so that
subscriptions work.

## Alternatives considered

A plain GraphQL server with one schema. The README's answer is that it is
not federation at all, and AD-0001 records why a single service lost.

Apollo's router, the main alternative platform. The README does not name it
and rejects it on licence alone, one the OSI does not recognise as open
source, which for an argument aimed at small organisations is not a
footnote. The research has only the vendor's comparison for the ELv2 label
and examined Apollo no further.

Cosmo's gRPC services and router plugins as the subgraph shape, which the
Cosmo research listed as its fourth and fifth options. Both keep the GraphQL
SDL as the contract and are the direction the README admires, and both
explicitly do not support subscriptions (C-14), so the two-UI live update
would fall back to polling. The research keeps the plugin route for a
later second example.

## Consequences

The demo can be launched with nothing but Docker and run with the network
removed, once the four variables are set. Composition runs on any machine
with Node and no account anywhere, the router starts from a file, and the
playground is where a visitor sees three services answer one query
(SR-43). Everything checked is Apache-2.0, and the pinned versions stay so.

The public text has to carry the vendor's position. Wherever it describes
static composition it quotes "it is recommended to not use this for
production" and says what the demo does about it, which is that no control
plane exists to fetch from and the configuration is committed and tested
for drift (C-88), and the README's air-gap sentence gains the qualification
and names the image's variables (C-85). CI has no Node, so the committed
configuration is all it can check, and SR-42's drift test is the third
condition's whole evidence. Router and wgc are pinned together because the
router refuses to start on an execution config whose compatibility version
exceeds its threshold (C-16), and every router bump is a rebuild of the
demo image (C-51).
The router binary is about 40 MB compressed and sets the floor of SR-06's
budget (C-08). The licence text travels with the binary (SR-07) and the
repository carries a `NOTICE` (SR-08).

Four spikes belong to this decision. The first, C-07, runs the image under
`docker run --network none` at debug log level and watches for connection
attempts before the air-gap claim goes public, which is SR-03's
demonstration. The second, C-15, composes the three draft schemas with wgc
and runs gqlgen over the nested `@requires`, which the research never ran.
The third, C-20, starts the copied binary from the supervisor with
environment alone in a distroless image with no YAML file present. The
fourth, C-57, observes what `/health/ready` reports with the subgraph
ports closed.

## Requirements affected
SR-03, SR-07, SR-42, SR-43

## Sources
README "Why Cosmo, and not simply GraphQL", the constraints list C-01 to C-05, C-07, C-08, C-14 to C-20, C-51, C-57, C-81, C-85, C-88, the design brief P2, P4, P5, the requirements list SR-03, SR-06, SR-07, SR-08, SR-42, SR-43, the research notes on Cosmo, options and licence findings, the sceptic verdicts on the research, licence and offline verdicts, the engineering log's 2026-08-26 planning entry (findings that changed the design).

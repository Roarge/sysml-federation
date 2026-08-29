# The query pipeline example

## What this is

A SysML v2 model of a query processing pipeline of five servers, served as a
live GraphQL projection by the adapter and joined at a federation router by two
services that know nothing about SysML. One computes the pipeline's capacity
from the wiring and returns a verdict for every requirement it can judge. The
other holds the structure of a requirements document, its order, its numbering
and what it leaves out. Two small web apps sit in front of the router, a model
viewer and the document itself, and the whole arrangement runs from one
container image. The repository README states the claim this example exists to
test: "This repository is an argument that federation is the missing
integration layer for open MBSE."

## Run it

    docker run --rm -p 8080:8080 ghcr.io/roarge/sysml-federation

Then open `http://localhost:8080/`. The image is published from a version tag,
so the line above works from the first tag onwards. From a checkout there is
nothing to wait for: `make image && make run` builds the same Dockerfile for the
host's own architecture and runs the result on the same port.

Port 8080 carries four paths and nothing else. `/viewer/` is the model viewer,
`/document/` the requirements document, `/graphql` the router's endpoint and
`/playground` the router's own query editor. A request to `/` is redirected to
`/viewer/`. The three services and the router listen on loopback inside the
container and are not reachable from outside it, and the router's health paths
are not proxied either.

Nothing is written to disk. An edit changes the served text and a version
counter held in memory, so stopping the container and starting it again returns
the shipped state, model version 1 and document version 1. A Reset button in
either app puts the shipped model and the shipped tree back without a restart,
although the two counters go on rising rather than returning to 1.

## What to try

The demo is arranged around a bottleneck that moves, and four edits in the
viewer are enough to watch it move.

Open `/viewer/`. It draws the served model text with the language marked, a
sketch built from the wiring, an edit panel of six controls, and one block per
requirement carrying the verdict the capacity service returned for it. In the
shipped state the caption under the sketch reads `capacity 1200, bottleneck
parse`, the parse box is outlined red, and PIPE-R1 reads `FAIL capacity 1200
against 1500, limited by parse`.

Raise ingest from 2000 to 3000. The literal changes in the text pane and the
derived requirement on ingest, PIPE-R1.1, goes to `PASS throughput 3000 against
1500`. Nothing else moves. Capacity is still 1200 and parse is still the
bottleneck, because a chain carries no more than its weakest stage and ingest
was never it.

Raise parse from 1200 to 1700. Capacity goes to 1400 and the caption names
indexA and indexB, both now outlined red. PIPE-R1 still fails. Raising the
bottleneck moved it rather than removed it: the two index servers at 700 each
add to 1400, and that pair is now the cheapest place to cut the wiring.

Raise indexA from 700 to 900. Capacity reaches 1600 and PIPE-R1 passes, `PASS
capacity 1600 against 1500, limited by indexA, indexB`. PIPE-R1.4, the derived
requirement on indexB, still fails at `throughput 700 against 750`. The pipeline
meets its rate while one of its servers does not meet the share allocated to it,
because indexA is now carrying more than half of the parallel stage. The
document's first paragraph warns of exactly that before anyone reads a verdict.

The other direction is the limit. Press Reset, then set PIPE-R1's limit to 1000.
It passes at capacity 1200 with no server touched, and the constraint literal in
the text pane reads 1000 where it read 1500.

Now open `/document/` beside the viewer. The same requirements appear as a
numbered document, an unnumbered paragraph first, then PIPE-R1 as 1 with its
five derived requirements nested 1.1 to 1.5 in server order, then PIPE-R2 as 2.
Each row shows what the model holds, what the analysis returned, and the one
thing the document decides for itself, which is the number. Edit a value here
and the viewer follows without a reload, and the reverse holds as well.

Drag PIPE-R1.5 by its grip above PIPE-R1.1. It renumbers to 1.1 and the others
follow in their former order. The header's document version rises and its model
version does not, because order and numbering belong to the document and nothing
about the model changed. Excluding a requirement behaves the same way: it leaves
the tree, a tray offers to restore it, and the viewer goes on listing it.

## How it is put together

Three services stand behind the router. None of them imports, calls or is
configured with the address of either of the others, and a test walks the import
graph to keep it so.

The adapter serves the model. It parses `model.sysml` and publishes parts with
their attributes, ports, children and connections, requirements with the
quantity, comparison and limit read from their constraints and the satisfy and
derive relationships around them, the verification cases, the model's own text
and a version counter. Nothing in it is named after this example, and a test
scans its sources for the example's words.

The capacity service computes the rollup. It is handed two names when it starts,
`capacity` for the quantity it computes and `throughput` for the attribute it
reads from each child part, and neither word appears in its own sources. What
reaches it over the wire is a subject part with its children, their attribute
values and the connections between them, carried across from the adapter's
fields by the router. It keeps no copy of the model and computes afresh on every
read.

The document service holds the tree: the ordering, the numbering, the headings
and paragraphs someone added, and the requirements someone excluded. Its shipped
tree is the one place in a service where the example's identifiers appear,
because a document is about particular requirements and cannot be written
without naming them.

One query shows the join. `{ requirement(id: "PIPE-R1") { text verdict
verdictReason documentNumber } }` comes back as a single object whose text is the
adapter's, whose verdict and reason are the capacity service's, and whose number
is the document service's. All three declare `Requirement` with the same entity
key, `id`. The router works out which service holds which field, calls each one
and merges the answers on that key.

### The rollup

The capacity of the pipeline is the largest sustained query rate the wiring can
carry from the servers that receive queries to the servers that deliver results.
As arithmetic that is a minimum along a chain and a sum across parallel branches,
which is the version a reader can check by hand. The service computes a maximum
flow instead. Each child part becomes two nodes joined by an edge of its
attribute value, each connection becomes an edge with no limit of its own, a
super-source feeds every child nobody feeds and a super-sink drains every child
that feeds nobody, and the capacity is the maximum flow between the two. The
bottleneck is the minimum cut taken on the source side, which is the saturated
servers nearest the entry and is the same set for every maximum flow, so the
answer does not depend on the order the algorithm found its paths in. Flow and
the two rules agree wherever the two rules apply, and flow still answers where
they do not, such as a fan-in from branches that were never forked at the same
point.

The arithmetic is exact for an idealised pipeline and for nothing else. Work is
taken to be evenly partitionable across parallel branches, load balancing to be
perfect, load to be stationary and connections to be unlimited, and there is no
queueing anywhere in the model. A real pipeline that neither drops nor
duplicates queries departs from those assumptions in one direction only, so the
number then reads as an upper bound. It must not be used for capacity planning.
The arithmetic was chosen to make a point about federation.

### The router

The router is the vendor's own binary, copied unchanged out of
`ghcr.io/wundergraph/cosmo/router:0.343.1` and run as a child process with an
environment the supervisor builds in full and nothing inherited. Its
configuration is composed outside the build and committed to the repository,
which is the path the vendor advises against. The composition page says "it is
recommended to not use this for production", and at every start without a graph
token the router itself logs "No graph token provided. The following Cosmo Cloud
features are disabled. Not recommended for Production."

The demo takes that at face value rather than around it. There is no control
plane here to fetch a configuration from, the composed file is committed where
anyone can read it, and a test fails when it drifts from the subgraph schemas it
was built from. That answers the caveat for a demo, which is what this is. It
does not answer it for a deployment.

### The image

`examples/pipeline/Dockerfile` builds in three stages. Go 1.27 cross-compiles
the binary for the target architecture with `CGO_ENABLED=0` and
`GOTOOLCHAIN=local`, so the toolchain in the image is the toolchain that builds.
The router stage names its source by tag and digest together, the tag saying
which release this is and the digest being what actually gets pulled, and since
the digest is of the multi-platform index both architectures still resolve from
it. The last stage is `gcr.io/distroless/static-debian13:nonroot`. It receives
the router binary, the router's Apache licence fetched from the vendor's
repository by checksum and left readable at mode 644, the Go binary, and
`config.json` and `model.sysml` under `/app`. No stage after the first runs a
command, so there is no shell in the image, and the `HEALTHCHECK` runs the
binary's own `healthcheck` subcommand rather than anything that would need one.
One port is published.

Five environment variables are set in the image, `DO_NOT_TRACK=1`,
`COSMO_TELEMETRY_DISABLED=true`, `TRACING_ENABLED=false`,
`METRICS_OTLP_ENABLED=false` and `PROMETHEUS_ENABLED=false`. They turn off the
vendor's anonymous usage tracker, its tracing and metrics exporters, and the
Prometheus scrape endpoint the router would otherwise open. On the ordinary path
they are inert, because the supervisor hands the router child a complete
environment of its own that sets the same five. They are in the image for anyone
who runs `/router` out of it directly.

Compressed as a registry counts them, the image comes to 44,845,388 bytes for
amd64 and 41,470,306 bytes for arm64, against a published ceiling of 80,000,000.
Those two figures are read from a two-platform build exported as a local OCI
layout rather than from a manifest in a registry. The publishing workflow pushes
the version tag, reads both platforms back from the registry, fails if either is
over the ceiling, and moves `latest` only after that.

## The model

`model.sysml` is a package `PIPE` holding five servers wired in series and in
parallel, a throughput requirement on the pipeline derived into one requirement
per server, a latency requirement with a verification case, and no rollup
arithmetic anywhere. Capacity is declared without a value and computed by the
capacity service from the wiring. Every published element carries a short name in
the `PIPE` family, which is the key the adapter publishes and the string the
other two services join on.

The file leans on a small set of constructs: an abstract part definition
specialised by two others, an item definition, port definitions carrying a
directed item with typed port usages joined by `connect`, plain numeric bindings
on `Real` attributes, a division inside a bound expression, a duration written
`200[ms]` against the model's own `<ms>` declaration, `satisfy` statements inside
the pipeline part, a subject in a requirement usage and again in a verification
usage, an `objective { verify latencyLimit; }`, a `#derivation connection` with
one `#original` end and five `#derive` ends, and `doc` bodies at package level
and across several lines. Both reference tools accepted every one of them, the
OMG pilot implementation release 2026-07 with kernel 0.61.0 and OpenSysML
v0.2.1. The record is at the foot of this file.

### Validating the model

The adapter's parser accepts a subset of the notation and cannot prove
that the file is valid SysML v2, so the file is checked with the two
reference tools before it is used as a fixture. The check runs locally and
never in CI.

OMG pilot implementation, release 2026-07, kernel 0.61.0 (Java 21 or
later). The kernel's batch entry point reads one model from standard input
between `%` markers:

    curl -sSfLO https://github.com/Systems-Modeling/SysML-v2-Pilot-Implementation/releases/download/2026-07/jupyter-sysml-kernel-0.61.0.zip
    unzip -q jupyter-sysml-kernel-0.61.0.zip -d pilot
    { printf '%%\n'; cat model.sysml; printf '\n%%\n%%exit\n'; } \
      | java -cp pilot/sysml/jupyter-sysml-kernel-0.61.0-all.jar \
             org.omg.sysml.interactive.SysMLInteractive pilot/sysml/sysml.library

OpenSysML v0.2.1 (Go 1.25 or later):

    go install github.com/Open-MBEE/OpenSysML/cmd/sysml@v0.2.1
    sysml -validate -strict model.sysml

The pilot accepts the file when its output carries no `ERROR:` or
`WARNING:` diagnostic and prints the root element line
`1> Package <PIPE> QueryPipeline (<uuid>)`. A diagnostic follows the
`1> ` prompt on the same line. OpenSysML accepts it when the command
exits 0, printing `✓ package QueryPipeline` and
`✓ model.sysml: no errors`.

### What the adapter refuses

The parser accepts exactly the subset of the notation the adapter projects and
refuses anything else with the file, the line and the column of the first
offending token. It has no opinion about meaning: resolving names and evaluating
expressions happen after it, and a model that parses can still be refused there,
again at a position in the source.

Editing is narrower than reading. A value is editable only where a numeric
literal in the source binds it and no other value shares that literal. An
expression-bound limit such as `globalThroughput.requiredRate / 2` is refused
with `the value is not a literal in the source`, and so is a value nothing binds
at all. A value that is not a finite, non-negative number is refused before
anything is patched. Either refusal names the element it concerns, leaves the
served text as it was and does not move the version counter, so nothing
downstream is told anything happened.

## Limits

Expression-bound values are read-only. The five derived requirements take their
limits from PIPE-R1 by arithmetic, so each of them follows an edit to PIPE-R1 and
none of them can be edited on its own. Both apps offer a limit control only where
the analysis reached a verdict, so PIPE-R2, whose latency no service computes,
shows its limit as text.

The viewer's edit panel reaches the attributes of the parts directly inside a
root part, which here is the five servers. An editable attribute belonging to a
root part itself gets no control in the viewer. This model has no such attribute
and the adapter's second fixture, `adapter/model/testdata/warehouse.sysml`, does:
its root part carries `shifts = 2`. The limit is the viewer's. The adapter
projects such an attribute like any other and the document app offers a control
for every editable attribute of a requirement's subject, at whatever depth the
subject sits.

The requirements document fetches its tree six levels deep. Anything
below that is not shown, and the deepest row visible says so rather than
offering a place to add more.

A refetch redraws a page from the answer the router gave. Focus returns to the
control it was on, but a value typed into that control and not yet sent is
replaced by the served one, so an edit arriving from elsewhere at the wrong
moment costs a keystroke.

Neither app computes anything a service could have answered. Capacity,
bottleneck, verdicts, document numbers and the model text all arrive over the
wire, and the only thing either app works out for itself is where to draw a box
in the sketch.

The router is the vendor's binary, copied into the image unchanged. Nothing here
patches it, and its behaviour under this configuration is the vendor's.

## Licences

This repository is under the Apache License, Version 2.0, whose text is in
`LICENSE` at the repository root.

The Cosmo router is the vendor's work under the same licence. Its binary is
copied into the image at `/router`, and the vendor's own licence text sits beside
it at `/router.LICENSE`, fetched at build time from the vendor's repository at
the router's release tag and pinned by checksum.

SortableJS is under the MIT License. The minified file is vendored at
`examples/pipeline/ui/shared/Sortable.min.js` with its licence text beside it at
`LICENSE.SortableJS`.

`NOTICE` at the repository root names the copyright holder and both third-party
works, with the tags and the checksums they were taken from.

## Verification record

| Date | Tool | Versions | Command | Result |
|---|---|---|---|---|
| 2026-08-27 | OMG pilot implementation | release 2026-07, kernel 0.61.0, OpenJDK 21.0.12 | `PILOT=$HOME/.local/share/sysml-pilot/sysml` then `{ printf '%%\n'; cat examples/pipeline/model.sysml; printf '\n%%\n%%exit\n'; } \| java -cp "$PILOT/jupyter-sysml-kernel-0.61.0-all.jar" org.omg.sysml.interactive.SysMLInteractive "$PILOT/sysml.library"` | accepted, no errors, no warnings |
| 2026-08-27 | OpenSysML | v0.2.1, built with Go 1.27.0 | `sysml -validate -strict model.sysml` | accepted, exit 0 |
| 2026-08-27 | OMG pilot implementation | release 2026-07, kernel 0.61.0, OpenJDK 21.0.12 | `PILOT=$HOME/.local/share/sysml-pilot/sysml` then `{ printf '%%\n'; cat adapter/model/testdata/warehouse.sysml; printf '\n%%\n%%exit\n'; } \| java -cp "$PILOT/jupyter-sysml-kernel-0.61.0-all.jar" org.omg.sysml.interactive.SysMLInteractive "$PILOT/sysml.library"` | accepted, no `ERROR:` or `WARNING:` diagnostic, root element line `1> Package <WH> Warehouse (<uuid>)` |
| 2026-08-27 | OpenSysML | v0.2.1, built with Go 1.27.0 | `sysml -validate -strict adapter/model/testdata/warehouse.sysml` | accepted, exit 0, printing `✓ package Warehouse` and `✓ adapter/model/testdata/warehouse.sysml: no errors` |

The first two rows cover this example. The last two cover the adapter's
second fixture, `adapter/model/testdata/warehouse.sysml`, run from the
repository root on the same day. That fixture carries other names, a wiring
with fan-in, a constraint written with the subject last, a requirement
without a short name and a literal limit inside a requirement definition.

### Constructs confirmed

| Construct | Pilot | OpenSysML |
|---|---|---|
| plain numeric binding on a `Real` attribute, `attribute :>> throughput = 2000;` | accepted | accepted |
| division in a bound expression, `= globalThroughput.requiredRate / 2` | accepted | accepted |
| `200[ms]` with the model's own `<ms>` declaration | accepted | accepted |
| `port def` with a directed item, a typed port usage and `connect a.output to b.input;` | accepted | accepted |
| `abstract part def` | accepted | accepted |
| `item def Query;` | accepted | accepted |
| `satisfy X by Y;` inside the pipeline part | accepted | accepted |
| `subject target :> pipeline;` in a requirement usage | accepted | accepted |
| `subject target :> pipeline;` in the verification usage | accepted | accepted |
| `objective { verify latencyLimit; }` | accepted | accepted |
| `#derivation connection` with `#original` and `#derive` ends | accepted | accepted |
| package-level and multi-line `doc` bodies | accepted | accepted |

No construct was rejected, so no fallback was applied.

### The running stack

The three services, the router and the user interface server were run together
on x86-64 Linux on 2026-08-27, with Go 1.27.0 and the router binary taken from
the pinned image `ghcr.io/wundergraph/cosmo/router:0.343.1`, image id
`sha256:cdbfd46b79b4e3fb309785f90162b8261dce75dc1b238acf7f9416351ad0fcf3`. The
composed configuration came from `wgc@0.130.1`, which printed no warning of any
kind when it composed it.

The router started from its nine environment variables with no configuration
file beside the composed one, and named no host outside loopback in its log. It
bound `127.0.0.1:3002` for the graph and nothing else. Left to its defaults it
also opens a Prometheus scrape endpoint on `127.0.0.1:8088`, which nothing here
reads, and `PROMETHEUS_ENABLED=false` closes it. With the variable set, the log
line announcing the endpoint is absent and `ss` finds no listener on that port.
Without it, both come back. The three subgraphs bound `127.0.0.1:3011`, `:3012` and
`:3013`. Only the user interface port, 8080, was bound on a wildcard address,
so the running tree has five listeners.

| Date | Requirement | What was run | What was observed |
|---|---|---|---|
| 2026-08-27 | SR-04 | `GET /`, `/viewer/`, `/document/`, `/playground` and `/health/ready` on port 8080 | `302` to `http://127.0.0.1:8080/viewer/` in one hop, then `200`, `200`, `200` and `404`. The router answers `200` on `/health/ready` when asked directly on its own port, so the `404` is the user interface server declining to proxy it |
| 2026-08-27 | SR-43 | `{ requirement(id: "PIPE-R1") { text verdict verdictReason documentNumber } }` through the router | `{"data":{"requirement":{"text":"The pipeline shall sustain the required query rate","verdict":"FAIL","verdictReason":"capacity 1200 against 1500, limited by parse","documentNumber":"1"}}}`, one answer drawn from all three services. `{ part(id: "PIPE-P1") { capacity bottleneck { name } } }` gave `1200` and `parse`, and introspection listed `Model`, `Verdict`, `Document` and `Node` in one schema |
| 2026-08-27 | SR-26 | A subscription on `subscription { modelChanged }` over server-sent events, then `setAttribute(partId: "PIPE-S2", name: "throughput", value: 1700)` | `:heartbeat` while idle, then `event: next` with `data: {"data":{"modelChanged":2}}` within a second of the mutation. A second edit gave `data: {"data":{"modelChanged":3}}`, and the verdict moved to `PASS` with reason `capacity 1600 against 1500, limited by indexA, indexB` |
| 2026-08-27 | SR-36 | `{ model { version } }`, then `moveNode`, `excludeRequirement` and `insertHeading` through the router, then `{ model { version } }` again | The document version rose 2, 3, 4, 5 while the model version read 4 before and 4 after. As a control, one model edit then moved the model version from 4 to 5 and left the document version at 5, so the reading is capable of changing |
| 2026-08-27 | SR-44 | `mutation { resetModel { version } resetDocument { version } }` | A further `modelChanged` event, and the shipped verdict again: `FAIL`, `capacity 1200 against 1500, limited by parse`, document number `1`. A restart of the whole stack gave model version `1`, document version `1` and the same verdict |
| 2026-08-27 | SR-24 | `mutation { setLimit(requirementId: "PIPE-R1.3", value: 900) { id } }` | `{"errors":[{"message":"the value is not a literal in the source: limit of requirement \"PIPE-R1.3\"","path":["setLimit"],"extensions":{"code":"DOWNSTREAM_SERVICE_ERROR","serviceName":"model"}}],"data":null}`, so the refusal is the whole of `errors[0].message`. No change to the model version and no event on the subscription |
| 2026-08-27 | SR-25 | `mutation { setAttribute(partId: "PIPE-S1", name: "throughput", value: -5) { id } }` | `{"errors":[{"message":"the value must be a finite, non-negative number: got -5","path":["setAttribute"],"extensions":{"code":"DOWNSTREAM_SERVICE_ERROR","serviceName":"model"}}],"data":null}`, the same shape. No change to the model version and no event on the subscription |
| 2026-08-27 | SR-09 | `SIGTERM` to the supervisor while a subscription was held open, then a restart of the whole stack after the edits recorded above | `stopping`, the router's own shutdown lines, exit status 0 after 2.00 seconds, an empty `ss` table on all five ports and no router process left. The held subscription was cut rather than waited on. The stack came back on model version `1`, document version `1` and the shipped verdict `FAIL`, `capacity 1200 against 1500, limited by parse`, so no edited value and no document change outlived the process |
| 2026-08-27 | SR-41 | An import sweep over `adapter`, `examples/pipeline/capacity` and `examples/pipeline/document`, test files excluded | Each package imports only its own subpackages: `capacity` to `capacity/flow`, `document` to `document/tree`, `serve` to `projection`, `projection` to `model` and `model` to `syntax`. No service imports another |

The router's readiness endpoint answers 200 as soon as its execution
configuration is loaded, whether or not the subgraphs are reachable, so a
readiness probe against it proves that the router process is listening with its
configuration loaded and proves nothing about the graph answering.

A refusal from a service reaches the client unwrapped, because the router is
configured with `SUBGRAPH_ERROR_PROPAGATION_MODE=pass-through` to pass a
subgraph's own error through. The service's message is `errors[0].message`, the
field it names is at `errors[0].path`, and `errors[0].extensions` carries `code`
as `DOWNSTREAM_SERVICE_ERROR` and the name of the service that refused. Left to
its default the router puts `Failed to fetch from Subgraph 'model'.` at the top
instead and nests the service's message under `errors[0].extensions.errors[0]`,
two levels below where a client reads a message. Passing the error through
changes nothing about a subgraph that cannot be reached at all. Pointed at a
dead port the router answers `{"errors":[{"message":"Failed to fetch from
Subgraph 'model'."}],"data":null}` in either mode, so no address, port or
internal error text is disclosed.

Two of the three services serve subscriptions and so register a WebSocket
transport, the model service and the document service. The capacity service
registers only `POST` and has no WebSocket transport to check. Both of the two
keep the library's same-origin check, and both were probed. The handshake the
router sends carries no `Origin` header, on either service, so the check never
fires on the router's own connection and both answered `101 Switching
Protocols`. The check is live all the same: the same handshake sent by hand with
`Origin: http://evil.example` was refused with `403 Forbidden` and `request
Origin "evil.example" is not authorized for Host "127.0.0.1:3011"`, and the
document service refused it the same way naming its own port.

### The two web apps

Both apps were driven by hand in a browser on 2026-08-28 against the stack
described above, starting from the shipped model and the shipped document. The
browser was Microsoft Edge 151 on Chromium 151, on x86-64 Linux reached from
Windows, and the drags and the recovery of a stale page were run in Google
Chrome 151 on the same host. Where a row records something other than the
expected result, the row says what happened instead.

A drag cannot be started from injected pointer input in this browser family, so
the drag rows were driven by dispatching the drag events the library reads,
along the path a pointer would take, the first drag in five moves and the later
walk three pixels at a time. The library cannot tell the difference, because it
never reads an event's `isTrusted` flag, so a dispatched event reaches the same
handlers a pointer's would. What happens at the foot of a list was settled by
measuring the page as well as by reading the vendored library. Its
`emptyInsertThreshold` defaults to five pixels, so an empty child list claims a
drop from that far outside itself on every side, and a list that ends on the
same pixel as its last row leaves nowhere to aim past that row. The remedy is
the strip a list keeps below its last row, 20 pixels against that five-pixel
reach, which is geometry rather than event handling, and a pointer meets it
the same way.

| Date | Requirement | What was run | What was observed |
|---|---|---|---|
| 2026-08-28 | SR-11 | `/viewer/` opened, then the tokeniser called from the console on `part <'PIPE-S2'> parse : Server { attribute throughput : Real = 1200; }` | The text pane draws the served file with `package`, `part`, `attribute`, `requirement`, `connect` and `satisfy` in bold ink, numbers and quoted names in blue `oklch(0.52 0.12 250)` and `doc` bodies in pencil. The call returned 28 tokens reading keyword, space, operator, string, operator, space, identifier and on to number, operator, space, operator, contiguous from 0 to 71 with no gap |
| 2026-08-28 | SR-10 | The network list over a reload of `/viewer/` | Eight requests and all of them to `localhost:8080`: the page, `style.css`, `app.js`, `shared/graphql.js`, `tokeniser.js`, `sketch.js` and two to `/graphql`. No font and no other host. The reload with the machine taken off the network was not run, so this row covers the request list alone |
| 2026-08-28 | SR-12 | The served sketch, then `render` called from the console on a four-part wiring with fan-in from two depths | Five boxes left to right at x 16, 188, 360 and 532 with indexA above indexB, arrows ingest to parse, parse to both index servers and both to serve, each box reading `throughput N`, the caption `capacity 1200, bottleneck parse` and parse stroked red. The console wiring drew a and b in the first column, c in the second and d in the third, with d receiving an arrow from a and from c. The next event redrew the served model |
| 2026-08-28 | SR-14, SR-15 | The requirements column as served | PIPE-R1 red and reading `FAIL capacity 1200 against 1500, limited by parse`, PIPE-R2 reading `INCONCLUSIVE PIPE-VC1 is declared and no service runs it`, and the five derived blocks reading `PASS throughput 2000 against 1500`, `FAIL throughput 1200 against 1500`, `FAIL throughput 700 against 750` twice and `PASS throughput 1800 against 1500` |
| 2026-08-28 | SR-13, SR-38 | A count of the controls on each page | Six in the viewer and six in the document, the five server throughputs and PIPE-R1's limit. No control for PIPE-R2's 200 ms, for a derived requirement's limit, or for capacity or latency |
| 2026-08-28 | SR-12, SR-14, SR-22 | Values edited in the viewer's panel | ingest to 3000 put `3000` in the served text where `2000` was, left capacity at 1200 and parse the bottleneck and moved PIPE-R1.1 to `PASS throughput 3000 against 1500`. parse to 0 gave `capacity 0, bottleneck parse` and `FAIL capacity 0 against 1500, limited by parse`. parse to 1700 gave capacity 1400 with indexA and indexB both red, and indexA to 900 then gave `PASS capacity 1600 against 1500, limited by indexA, indexB` with PIPE-R1.4 still `FAIL throughput 700 against 750`. PIPE-R1's limit at 1000 gave PASS, and at 2500 gave a red `FAIL capacity 1200 against 2500, limited by parse` with `2500` in the served text where `1500` was |
| 2026-08-28 | SR-25 | `abc`, then `-5`, then an emptied field, each followed by Tab, in the viewer and again in the document | This browser refuses the letters. With the caret in the field and nothing selected, `abc` produced three `beforeinput` events, no `input` event, no change to the value and no change event on Tab, so the app said nothing and nothing moved. With the value selected first, the refused keystroke still cleared the selection, the field reported an empty value with `badInput` false and the app answered as it does for an emptied field. `-5` gave `"-5" is not a finite, non-negative number, the served value stands` and `-1` in the document gave the same sentence, and an emptied field gave it with an empty value quoted. Every refusal put the served value back and changed nothing else |
| 2026-08-28 | SR-39, SR-44 | Reset pressed in each app | The shipped values came back in both apps within two seconds, and a Reset in one app was picked up by the other. A reset raises both counters rather than returning them to 1, so the header reads `version 1` only on a stack that has just started |
| 2026-08-28 | SR-39 | The stack stopped under an open viewer, then started again | The status line read `live updates: network error` and then `live updates: Failed to fetch`, and the retries went out about 1, 2, 4, 8 and 8 seconds apart, each attempt taking about 2.4 seconds to fail against the closed port. After the restart the next attempt succeeded, and an edit made in the playground reached the viewer with no reload. With a control focused, an edit from the playground redrew the page and the same control kept focus |
| 2026-08-28 | SR-39 | A viewer left in a background tab across two model changes and then read again, and the same page brought back from a stale state | The page was drawing the model as it stood two changes earlier, with an empty status line and no reconnection to make, which is the browser freezing a tab it thinks nobody is watching. Both apps ask the router for the current state whenever the page becomes visible, and a page holding a model two changes out of date redrew the served version, values, wiring caption and verdicts within a second of that, cleared its status line and went on taking live changes with no reload. Every browser reachable from here holds a page at one visibility state, so the event itself was raised in the page rather than by switching to the tab |
| 2026-08-28 | SR-33, SR-37 | `/document/` opened on a freshly started stack | The unnumbered prose paragraph sits first, PIPE-R1 is `1` with PIPE-R1.1 to PIPE-R1.5 nested `1.1` to `1.5` in server order, PIPE-R2 is `2` and the header reads `document version 1, model version 1`. PIPE-R1's row carries its short name and text, `limit` as a control holding 1500, `derives PIPE-R1.1, PIPE-R1.2, PIPE-R1.3, PIPE-R1.4, PIPE-R1.5`, `satisfied by pipeline`, `current value capacity 1200` and a red `FAIL capacity 1200 against 1500, limited by parse`. PIPE-R1.2 carries `derived from PIPE-R1`, `satisfied by parse`, `current value capacity 1200`, `parse throughput` as a control holding 1200 and `FAIL throughput 1200 against 1500`. PIPE-R2 carries `limit latency <= 200 ms` as text, `verified by PIPE-VC1`, no current value and no control |
| 2026-08-28 | SR-34 | PIPE-R1.5 dragged by its grip above PIPE-R1.1, then PIPE-R2 dragged to the foot of PIPE-R1's list, then dragged again into PIPE-R1.5's own empty child list | PIPE-R1.5 read `1.1` and the others `1.2` to `1.5` in their former order, with the prose and PIPE-R2 unchanged. PIPE-R2 then landed as the last child of PIPE-R1 reading `1.6`, keeping `verified by PIPE-VC1` and its INCONCLUSIVE verdict, on one `moveNode` that raised the document counter and left the model counter alone. The third drag put the same item at `1.5.1`. A list keeps a strip below its last row, so the foot of a list and the last row's own empty child list are two places a drag can be aimed at separately |
| 2026-08-28 | SR-33, SR-34 | `Heading above` on PIPE-R1, then `Add prose` on the new heading, then the heading text edited in place | The heading read `1`, PIPE-R1 read `1.1` and its children `1.1.1` to `1.1.5`. The paragraph appeared as the heading's last child, dashed and unnumbered. The heading text changed to `Performance budget` on Enter and came back from the service with that text. `Exclude` on PIPE-R1.4 took it out of the tree, moved PIPE-R1.5 to `1.1.4`, listed `PIPE-R1.4 indexBThroughput` in the tray with `Restore` and left PIPE-R1.4 in the viewer. `Restore` put it back as the last child at `1.1.5` |
| 2026-08-28 | SR-36 | The header read through the moves, the heading, the paragraph and the exclusion above | The document counter rose through every one of them and the model counter did not move |
| 2026-08-28 | SR-39 | PIPE-R1.2's throughput set to 1700 and PIPE-R1's limit to 1600 in the document, then parse to 1700 and indexA to 900 in the viewer | The document gave PIPE-R1.2 `PASS`, PIPE-R1 `FAIL capacity 1400 against 1500, limited by indexA, indexB`, and the viewer showed 1700 in the served text and `capacity 1400, bottleneck indexA, indexB` with no reload. The limit at 1600 put `1600` in the viewer's constraint literal. The edits made in the viewer gave the document `PASS capacity 1600 against 1500, limited by indexA, indexB`, PIPE-R1.3 `PASS` and PIPE-R1.4 `FAIL throughput 700 against 750` |
| 2026-08-28 | Refusal text | `moveNode(id: "no-such-node", parentId: null, index: 0)` sent from the document's own client | The client threw with the message `no such node: no-such-node`, which is the whole of `errors[0].message` as the service wrote it |
| 2026-08-28 | Depth limit | `Heading above` pressed four times on PIPE-R1, taking its children to the sixth level | Each row at the sixth level reads `The document is shown 6 levels deep, so this item cannot be nested any deeper.` and offers `Exclude` alone. Both `Add prose` and `Heading above` are withheld there, and they are present at every level above |

A page left in a background tab for several minutes stops being updated, because
the browser freezes the tab, and it shows no error while that lasts. Both apps
ask the router for the current state whenever the page becomes visible, so a tab
that missed changes while it was away is drawing the served model again by the
time anyone reads it. A page loaded into a background tab and left there receives
its updates normally, so the freeze is the browser's own idle handling rather
than the subscription failing.

Neither app logged anything to the console over the whole session, and no
JavaScript error was raised.

### The container image

The image was built from `examples/pipeline/Dockerfile` with the repository
root as the build context and run on x86-64 Linux on 2026-08-28, with Docker
29.7.2 and buildx 0.36.0. The two cross-architecture builds came from the same
buildx on the same host. Nothing was pushed, so every figure below is read from
a local build.

| Date | Requirement | What was run | What was observed |
|---|---|---|---|
| 2026-08-28 | SR-04 | `GET /`, `/viewer/`, `/document/` and `/playground` on port 8080 of the running container, a `{ model { version } }` query on `/graphql`, and `subscription { modelChanged }` held open over server-sent events | `302` with `Location: /viewer/`, which resolves to `http://localhost:8080/viewer/` on the same port, then `200`, `200` and `200`, and `{"data":{"model":{"version":1}}}` from the query. The subscription was held to the client's own twelve-second limit and carried repeated `:heartbeat` lines, three in one run and two in another. The one published port answers all of it |
| 2026-08-28 | SR-02 | A container started detached from the built image and polled every 200 ms by the same command that started it, both for `/viewer/` and for a real `{ model { version } }` query through the router | The viewer answered `200` after 0.75 s and the query after 0.77 s, well inside the ten-second budget. The query figure covers the whole stack coming up rather than the user interface server alone. A first attempt was discarded because the poller was started separately from the container and so began reading a container that had been up for several seconds |
| 2026-08-28 | SR-07 | `/router.LICENSE` read from inside a container running the image and as the user the image runs as, then `/router -version` from the same image | The licence is the Apache License version 2.0 text, mode 644 owned by root, which uid 65532 can read and did. The binary reports version 0.343.1, Go 1.25.14, linux/amd64, built 2026-08-26 from revision `6a8da18`, so the certificate and the binary name one release |
| 2026-08-28 | SR-06 | `docker buildx build --platform linux/amd64,linux/arm64` exported as an OCI layout, then the per-platform manifests read out of it | One image manifest per platform under a manifest list, with an attestation manifest for each. Compressed layers come to 44,845,388 bytes on amd64 and 41,470,306 on arm64, or 44,850,223 and 41,475,142 with the configuration blob added. The OCI exporter gzips as a registry push does, so these are what a registry manifest would report, and the router binary is 39,987,546 of the amd64 total |
| 2026-08-28 | C-53 | `docker buildx build --platform linux/arm64` exported to a local directory, with the router image pinned by digest as well as by tag | `/router` and `/sysml-federation` are both `ELF 64-bit LSB executable, ARM aarch64`. The digest names the multi-platform index rather than one architecture, so the arm64 router still resolves from it |

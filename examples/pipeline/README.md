# The query pipeline example

`model.sysml` is a SysML v2 model of a query processing pipeline: five
servers wired in series and in parallel, a throughput requirement on the
pipeline derived into one requirement per server, a latency requirement
with a verification case, and no rollup arithmetic. Capacity is declared
without a value and computed by the analysis service from the wiring.
Every published element carries a short name in the `PIPE` family, which
is the key the adapter publishes.

How to run the demo and what to try arrive with the services.

## Validating the model

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

## Limits

The adapter reads the subset its `adapter/syntax` package documents and
refuses any other construct with the file, line and column of the first
one it meets. The model is edited only through the literals the adapter
publishes as editable.

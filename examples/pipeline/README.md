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

## Limits

The adapter reads the subset its `adapter/syntax` package documents and
refuses any other construct with the file, line and column of the first
one it meets. The model is edited only through the literals the adapter
publishes as editable.

The requirements document fetches its tree six levels deep. Anything
below that is not shown, and the deepest row visible says so rather than
offering a place to add more.

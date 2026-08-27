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

The pilot accepts the file when its output carries no `ERROR:` or `WARNING:`
diagnostic, the first of which follows the `1> ` prompt on the same line, and prints the root element line `1> Package <PIPE> QueryPipeline (<uuid>)`. OpenSysML accepts it when the command exits 0, printing `✓ package QueryPipeline` and `✓ model.sysml: no errors`.

## Verification record

| Date | Tool | Versions | Command | Result |
|---|---|---|---|---|
| 2026-08-27 | OMG pilot implementation | release 2026-07, kernel 0.61.0, OpenJDK 21.0.12 | the pilot command above | accepted, no errors, no warnings |
| 2026-08-27 | OpenSysML | v0.2.1, built with Go 1.27.0 | `sysml -validate -strict model.sysml` | accepted, exit 0 |

The adapter's second fixture, `adapter/model/testdata/warehouse.sysml`,
passed the same two runs on the same day. It carries other names, a wiring
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

## Limits

The adapter reads the subset its `adapter/syntax` package documents and
refuses any other construct with the file, line and column of the first
one it meets. The model is edited only through the literals the adapter
publishes as editable.

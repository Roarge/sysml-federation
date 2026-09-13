# A model of the demo itself

*Roar Georgsen, 13 September 2026*

Part 13 of 13 in [Federating a systems model](../README.md).

The [light requirements scheme](../decisions/AD-0023-light-requirements-scheme.md) put the repository's own requirements in Markdown tables and deferred a model of them. A reader judging a SysML adapter is not judging a way of working, the record said, and the question would return once the requirement count outgrew a screen. It outgrew a screen before the first line of code was written, and the question stayed deferred through the build.

What brought it back was drift. The published description came apart from the code more than once, a number, a file name, a count of tests, each true when written and wrong by the time a reader met it, and each found by somebody reading the two side by side. Reading is not a mechanism. A repository arguing that a systems model belongs at the centre of an organisation's engineering, while keeping its own description in prose that nothing checks, was making its case with the wrong example.

The demo is now [modelled in SysML v2](../decisions/AD-0029-the-demos-own-model-in-sysml-v2.md), under [`model/`](https://github.com/Roarge/sysml-federation/tree/main/model), with a unit test that fails when the model and the repository disagree and both reference tools reading the model on every change. A check session came with it, because a model in which every story has a verification case and no story has anything that runs it on a schedule is a description with more structure.

## Two models, one repository

The pipeline model is what the adapter serves. It lives in `examples/pipeline/model.sysml`, describes five servers wired in series and in parallel with a throughput each, a requirement on the whole pipeline and one derived from it for each server, and the adapter reads it at startup. Every board and every screenshot that shows `ingest`, `parse` or the model requirement `PIPE-R1` is showing that file's content.

The model of the demo describes the thing that serves it: the adapter, the capacity and document services beside it, the router that composes the three into one graph, the two web apps that read the graph, and the image they all ship in. Nothing of the pipeline is declared in it, and no service reads it.

A reader tells them apart by the shape of the name. A short name beginning `PIPE-` belongs to the example, is written in code font, and is called a model requirement or a model element wherever it appears in prose. The demo's own identifiers are the light scheme's, `US-nn` for a story, `SR-nn` for a system requirement, `SC-nn` for a design constraint and `AD-nnnn` for a decision record, and in the model each is a short name, `<'SR-22'>`, over a qualified name that carries the number and a phrase, `SR_22_EditsPatchTheSource`, so that a satisfy or a verify reads as a sentence. The numbers were in the articles and the records before the model existed, and they stay as they are.

A decision record is named by a metadata tag on the element it shaped, and a story's rationale cites the records it rests on, so a reader can go from an element to its reason without leaving the file.

## Stories, then system stories

The twelve stories of the storyboard are the stakeholder stories, each a requirement usage with a role, a capability, a benefit, nested acceptance criteria and a status carried as metadata. Seven were added with the model, one for each requirement that traced to an obligation the project had set itself rather than to a use case: that the licences travel with what is copied, that nothing of the example is in the adapter, that a served model is never silently wrong, that the contract is checked before deployment, that the example is accepted by the reference tools, that the demo's own model is the record, and that the stories are exercised against a running instance.

Before the model those requirements traced to an obligation and not to anyone who wanted it, and a story with a role and a benefit says who does.

The forty-five requirements of [From use cases to requirements](05-from-use-cases-to-requirements.md) are restated as system stories, with three more that came with the model: a unit test fails when the model and the repository disagree, both reference tools accept the model and continuous integration runs them, and the demo, given credentials, runs the check suite against itself through a tunnel and is unchanged given none. Each keeps its EARS statement as an attribute, `attribute statement : String = "..."`, so that the text a case verifies is the text the story carries and never a paraphrase.

Three statements were amended with the model, the no-outbound-network requirement for the configuration-file opt-in the session needs, the constraint on dependencies for the check project, and the constraint on continuous integration for the model validation, and each amendment has a record behind it. Article 05 keeps its count of forty-five, which is the count at the second gate and reads as a statement about that gate.

The traceability tables of that article, one hop each and maintained by hand, are now derivation connections, nineteen of them, one per stakeholder story, with the stakeholder story as the original end and every system story that follows from it as a derived end. Nothing was renumbered. A system story whose status is done must be a derived end of one of those connections, satisfied by a part in the architecture and verified by a case, and the `coverage` agreement names any that is not.

## The tests are in the model

A verification case per system story and per constraint, and inside each case one action per piece of evidence, with the action's short name the evidence's own name. A Go test is named by its function, `action <'TestSR02_ReadyWithinTenSeconds'> readyWithinTenSeconds`, with the kind and the file on an evidence tag, and the file matters because two functions share a name across two packages. Sixty-nine actions name a Go test.

A recorded run of the demo is named by its row in the example README's verification record, `<'record: 2026-08-28 SR-02'>`, and one row stands under every story it bears on. The two validators are actions of kind `validator`, a make target is an action named as it is typed, `<'make check-tracked'>`, a workflow is named by its file and step, `<'publish.yml: Read the manifest back'>`, and what is read rather than run, the router's outbound paths or the module file, is an analysis or an inspection.

A case names only evidence that exists. Four stories ask for more than the repository holds, one platform recorded of three, an offline render never run, two stories naming a test among their methods with no test of their own, and each of their cases claims what is there and nothing more.

Every live check is a verification case of its own. The check project reads a manifest of thirty entries, and the check register carries thirty cases whose attributes equal the entries, the check's id, its kind, its frequency in minutes, its locations, whether it is deployed and whether it mutates, and whose actions are the check's own steps: post the join query for `PIPE-R1`, expect the three subgraphs to answer. A thirty-first case has no kind and is the session's own record, the trace of a check request found in the viewer beside the demo.

The session is a composite in the logical architecture, the runner, the two tunnels of which one runs, the collector, the trace viewer, the optional container for a private location and a reference to the demo under test, each part carrying the name of the compose service it maps to. The runner's twelve steps are an action definition, and so is the path of one traced request, nine steps from the runner adding trace context to the trace shown beside the check result.

`TestSR46_ModelAndRepositoryAgree` in `internal/trace` is what keeps all of this true. It reads the registers as text and compares them with the repository around them, one subtest per agreement, and there are eight.

Every short name the model writes is declared exactly once, every decision record it names exists and every record in the directory is named. Every requirement a record names as affected is a short name the model declares. The Go test functions carrying a requirement identifier and the ones the verification register names are the same set, counted per file. The check files under `checkly/__checks__` and the file names the case registers quote are the same set, and every browser spec is exercised by exactly one validation case. Every image under `docs/img` is named by exactly one view, and every image a view names exists. A system story that is done is verified, satisfied and derived, and a stakeholder story that is done has exactly one validation case. The check register and the manifest carry the same checks with the same fields, each constructed once. And the services the compose file starts are the parts the session composite carries, and nothing else.

It runs in `make check` in a fraction of a second, and from the day it existed every edit to a check file, a manifest entry or a compose service was a model edit as well, or the gate went red.

## The boards and the views

Every published board corresponds to a view. The five architecture views, the two A3 sheets, the storyboard, the five-views board and the app screenshots have a view each, ten in all, and an eleventh, the session, names no image. A view exposes what its board draws, `expose Federation_LogicalArchitecture::demo::**` for the parts of the demo, `expose Federation_FunctionalArchitecture::EditPropagation::**` for an edit's seven steps, and nothing is redrawn. The boards stay the hand-drawn PDFs the design phase produced, and no rendering is declared on any view, so the model says what each board is about and leaves the drawing where it was.

The `images` subtest holds the naming both ways. What it cannot hold is that the drawing and the exposed elements agree, so each of the thirty-six images was opened once and read against its view's expose lines, and the model's README carries the correspondence as a table with the date of each reading, 12 September 2026.

Two patterns recur. The storyboard frames and the app screenshots draw the example's servers and requirements, which are content of the example model and no element of this one, and those rows say so. And several boards draw the parts of the demo on a page whose own view is about a functional package or a set of stories, so each of those views exposes the demo as well.

The reading found three things drawn that the model did not carry. The runtime board draws three panels beneath the seven steps of an edit, and two of them, a document edit and a reset, had no action definition, so they have one now beside the one for nothing moving. The deployment board draws the image's HEALTHCHECK block, now the `healthcheck` attribute of the demo, with the subcommand it runs among the supervisor's. And the last screenshot shows the document's six-level depth limit, now an attribute of the requirements document, exposed by the view, with three validation cases recording a run against it. Each is a small thing. Each is also exactly what a description checked by reading had not caught.

## Two tools on every change

`make model-check` puts the whole tree to both reference tools. The OMG pilot implementation, release 2026-07, runs in batch with the files concatenated in path order so that the imports between registers resolve, and passes on a root element line for every file and no line matching an error or a warning. OpenSysML v0.6.0 then runs `sysml -validate -strict` over the same files and passes on exit status zero with no warning line. A validator that is not installed fails the target, and nothing is skipped.

The same target runs [in continuous integration](../decisions/AD-0030-model-validation-in-continuous-integration.md) on every pull request and push that changes a file under `model/`, in a workflow of its own, with each tool installed from its release archive against a recorded checksum. The job is gated by paths and is not a required check, because a required check that never reports leaves a pull request pending for ever. The example model keeps its local check under `make example-model-check`, with its record in the example README, because the adapter's tests parse that file on every run and a broken example is caught in seconds anyway.

Nineteen forms the model meant to use were put to both tools in a probe file before the registers were written, and eighteen were accepted as written. The one refused by both was a view definition stating conformance with `satisfy viewpoint`, so a view definition owns a viewpoint usage instead.

Five more refusals came while the registers were written. `public` and `connector` are reserved words, so the tunnel edge's port is `publicSide` and the tunnel link's end is `tunnelConnector`. An interface whose two ports are not conjugate draws a warning from OpenSysML, and the gate treats a warning as a refusal, so the tunnel port carries the conjugate features of the HTTP client port and a twelfth port definition was added for the alerting interface. The pilot refused the satisfy that allocates the session requirement to the runner until the runner's part definition specialised the element type the story's subject has, because bound features must have conforming types.

A validator accepting a form is not a validator checking what the author meant by it, and the model's README says which acceptances carry that caveat. A `require constraint` holding a documentation comment where an expression would go is accepted by both and asserts nothing beyond its doc.

## A session, when asked for

The [check session](../decisions/AD-0031-an-optional-check-session.md) is a compose profile beside the demo, under `checkly/`. With a Checkly account and `bash checkly/scripts/session.sh up`, the stack that starts is the demo, a tunnel, an OpenTelemetry collector, a trace viewer and a runner. The tunnel is a named one on the operator's own hostname when a token is supplied and a quick one on a hostname the tunnel provider assigns otherwise.

The runner walks twelve steps, each logged under its number and each an action in the model. It installs the project, resolves the demo's public address, waits for the viewer through the tunnel, probes whether a subscription's events cross it, reads the account's plan, publishes the address as an account variable, records a test session of every check against this instance, records the browser suite as a second session, deploys the project, triggers every deployed check once so the dashboard and the status page fill, pings a heartbeat every five minutes, and on stop destroys the project and exits with the test session's code.

Thirty checks and monitors sit in five groups, and the three groups that edit the demo's one in-memory state run from one location with no retries and no parallel locations, because a retried or a parallel run would edit it twice. Two of their checks due at the same moment can still overlap, since the service does not serialise them, and a check that meets another's edit fails and resets the state. Without credentials `docker run` runs the demo exactly as before, and nothing under `checkly/` is built, run or read.

The trace is the part that was missing when a check failed. The router's tracing is an operator opt-in: a configuration file is handed to the router only when `SYSML_FEDERATION_ROUTER_CONFIG_PATH` names one, the session names a file whose one exporter is the collector on the compose network, and defining that exporter is what keeps the router's default one off. The collector forwards every span to Jaeger at `localhost:16686` and, with a tracing key, only the spans a check marked as its own to the monitoring service, where each is shown beside the check result that caused it.

One join query through the router gives a trace of thirteen spans: one server span at the root, six internal spans for reading the body and parsing, normalising, validating, planning and executing the operation, and three fetch spans named for the model, document and capacity subgraphs, each with the client span of its HTTP call beneath it. That is one request across the router and the three services, seen end to end, and the router is the only service in the container that could show it without a dependency the design constraints forbid.

What has been run without credentials is on record in the [check session's README](https://github.com/Roarge/sysml-federation/tree/main/checkly). The demo alone through the compose file, with the published image and with a local build, answers as `docker run` does, and the router's environment carries no configuration path and its log no configuration-file line. With the file named, the router logs `Tracer enabled` once, with the collector's address and no other exporter.

The stack came up under the quick profile, five containers, on a network whose access provider's resolver answers the tunnel service's name with a block page, so the tunnel exited and no hostname was assigned, and the collector and the viewer were verified with the demo reached directly on its port instead. The subscription probe's commands from the host saw the event frame after a second, and the runner's first four steps ran inside the container against a stand-in for the tunnel's metrics endpoint and stopped where a session without an account stops.

The join query's trace was read from Jaeger with the spans above. A request sent with a trace parent and a trace state carrying the mark the monitoring service's runners set came back under the id it was sent with and the mark on the router's span, and the second collector configuration passed the one marked span of three and refused a missing or a wrong token on its public port. The rows with credentials, the named tunnel, the two recorded sessions, the deploy and the destroy, the alerts and the trace beside a check result, are the owner's to add when the account exists.

## What is not there

The model carries no risk register and no priorities, and its stories carry a status and nothing else of a lifecycle. Allocation is expressed by `satisfy`, ninety-four of them in the components register, because the allocation keyword has no recorded validation run against either reference tool and a form neither has been seen to accept does not go into the model. No view declares a rendering, so nothing generates a diagram from the model and the boards are as they were drawn. The schema the services share is not modelled as elements. The interface definitions carry its payloads and the projection's doc names its types.

Quick tunnels are documented as carrying no streamed response, so a page behind one is not expected to hear the server-sent event that would redraw it, and seven of the twelve story checks depend on that event: the three cross-app stories, and four whose own page must learn of an edit from the stream, the raise that moves nothing, the raise that moves the bottleneck, the tightened limit and the shaped document. On a quick-tunnel session they skip with the reason `live updates need the named tunnel` rather than fail, and a named tunnel runs all twelve.

The private location, the monitoring service's runners placed beside the demo so that no tunnel is needed, is in the compose file behind a profile of its own and has not been run, because it needs a paid plan. Incident automation on the status page and a weekly maintenance window are written and gated behind flags for the same reason, so that the project deploys on the free tier as it stands.

And while a session runs the demo answers at a public hostname without authentication. It holds no secret, so what a visitor can do in that window is edit a value and fail a check.

Decision records: [docs/decisions](../decisions/README.md) · Repository: https://github.com/Roarge/sysml-federation

---

Previous: [What shipped, and what did not](11-what-shipped-and-what-did-not.md) · Index: [Federating a systems model](../README.md)

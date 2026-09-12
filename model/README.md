# The model of the demo

The demo is what is modelled here: the SysML v2 adapter, the capacity and
document services beside it, the router that composes the three into one graph,
the two web apps that read that graph, and the image they all ship in. The
pipeline the example describes, five servers and their requirements, is a
different thing. It is modelled in
[`examples/pipeline/model.sysml`](../examples/pipeline/model.sysml), the adapter
reads that file at startup, and nothing of it is declared here. Where a board or
a screenshot shows `ingest`, `parse` or `PIPE-R1`, it is showing the example's
content, and the correspondence table below says so.

One register per file, and the root imports every one of them.

| File | What it holds |
|---|---|
| [`library/`](library/) | the story definitions, a story as a requirement usage with a role, a capability, a benefit, nested acceptance criteria and a status carried as metadata, and beside them the definitions particular to this model, the scheme's enumerations, the design-constraint form and the metadata for decision records and evidence |
| [`core/core.sysml`](core/core.sysml) | the root package, which imports every register below |
| [`core/stakeholders/`](core/stakeholders/stakeholders.sysml) | six people, each a stakeholder of the description and an actor in the context |
| [`core/concerns/`](core/concerns/concerns.sysml) | one concern per stakeholder question, in the words the boards ask them in |
| [`core/context/`](core/context/architecture-context.sysml) | the demo among what it touches, the external systems, and the elements the requirements are allocated to |
| [`core/domain/`](core/domain/domain.sysml) | the items that move, pages, queries, version events, the image, the router configuration, spans and alerts |
| [`core/stories/stakeholder/`](core/stories/stakeholder/stakeholder-stories.sysml) | nineteen stakeholder stories, US-01 to US-19 |
| [`core/stories/system/`](core/stories/system/system-stories.sysml) | forty-eight system stories, SR-01 to SR-48, each restating one requirement, and their derivation from the stakeholder stories |
| [`core/constraints/`](core/constraints/design-constraints.sysml) | seven design constraints, SC-01 to SC-07, each with a plain statement |
| [`core/use-cases/`](core/use-cases/use-cases.sysml) | thirteen use cases, one per storyboard story and one for the check session |
| [`core/functional-architecture/`](core/functional-architecture/functional-architecture.sysml) | what the demo does, an edit's seven steps, the verdict arithmetic, the argument in eight steps and the check session's behaviour |
| [`core/logical-architecture/interface-types/`](core/logical-architecture/interface-types/interfaces.sysml) | the port and interface definitions, each conjugate written out |
| [`core/logical-architecture/components/`](core/logical-architecture/components/components.sysml) | what the demo is made of, the project around it, the check session, every allocation as a `satisfy` and the decision records as tags |
| [`core/verification-validation/verification-cases/`](core/verification-validation/verification-cases/verification-cases.sysml) | one verification case per system story and per constraint, with the evidence as actions |
| [`core/verification-validation/validation-cases/`](core/verification-validation/validation-cases/validation-cases.sysml) | one validation case per stakeholder story, with the recorded runs as actions |
| [`core/views/`](core/views/views.sysml) | the viewpoints, the view definitions and one view per published board or sheet |

The [articles](../docs/README.md) are the narrative and the
[decision records](../docs/decisions/README.md) are the rationale. The model is
the record of structure and traces (AD-0029), and neither of the other two is
repeated in it. A part carries the identifiers of the records that shaped it,
and a story's rationale cites the records it rests on, so a reader can go from
an element to its reason.

## Identifiers

Every identifier of the light requirements scheme is a short name here,
`<'US-01'>`, `<'SR-22'>`, `<'SC-04'>`, and a decision record is named by its
identifier in a `@DecisionRecord` tag, `AD-0018`. The qualified name behind a
short name carries the number and a phrase, `SR_22_EditsPatchTheSource`, so
that a `satisfy` or a `verify` reads as a sentence. `SR_` names a system story,
in the role a heavier scheme gives to a `SYS_` prefix. The numbers were already
in the articles and the records before the model existed, and they stay as they
are.

US-01 to US-12 are the storyboard stories. US-13 to US-19 were added with the
model, one for each requirement that traced to an obligation the project set
itself rather than to a use case: the licences that travel with what is copied
(US-13), nothing of the example in the adapter (US-14), a served model that is
never silently wrong (US-15), the contract checked before deployment (US-16),
the example accepted by the reference tools (US-17), the demo's own model as the
record (US-18), and the stories exercised against a running instance (US-19).
Three requirements came with them. SR-46 states that a unit test fails when the
model and the repository disagree on an identifier, a test name, a check file or
a published image. SR-47 states that both reference tools accept the model and
that continuous integration runs them on every change to it. SR-48 states that
the demo, given credentials, runs the check suite against itself through a
tunnel and is unchanged when given none.

The requirements are restated as system stories, and each story keeps its
statement as an attribute, `attribute statement : String = "...";`, so that
the text a case verifies is the text the story carries. Three statements were
amended with the model: SR-03 for the configuration-file opt-in, SC-01 for the
check project and SC-07 for the model validation.

A story's status is carried as `@StoryMeta` and is `done` or `inProgress`.
`done` is asserted as of the merge of the pull request that carries the story's
evidence, and not before. Three stories are in progress. US-19 and SR-48 wait
for the check session, which a later pull request adds. SR-03 waits for the
configuration-file opt-in its statement describes, which is not yet
implemented, so its case names the evidence that exists and the story stays
open until the rest does.

## Validation

`make model-check` puts the whole tree to both reference tools. The OMG pilot
implementation, release 2026-07 with kernel 0.61.0, is run in batch: the
seventeen files are concatenated in path order between `%` markers and read on
standard input as one block, so that the imports between registers resolve. The
target reads its verdict out of the output rather than the exit status, and
passes on a root element line for every file handed to it and no line matching
`ERROR:` or `WARNING:`. OpenSysML v0.6.0 then runs
`sysml -validate -strict` over the same files and passes on exit status zero
with no `warning:` line. A validator that is not installed fails the target with
a pointer to its install instructions, and nothing is skipped.

[`model.yml`](../.github/workflows/model.yml) runs the same target on every pull
request and every push to `main` that changes a file under `model/`, the
Makefile or the workflow itself, installing the pilot from its release archive
and OpenSysML from its release tarball, each checked against a recorded
checksum. It is not a required check, because a job gated by paths reports
nothing on a pull request that leaves the model alone, and a required check
that never reports leaves a pull request pending for ever (AD-0030).
`make example-model-check` puts the two example models, the pipeline and the
adapter's second fixture, to both validators one file at a time. It never runs
in continuous integration: the adapter's tests parse the example on every run,
and the record of the validator runs is in the
[example README](../examples/pipeline/README.md#verification-record).

### Verification record

Both commands were run from the repository root over the seventeen files as
committed, in the order `git ls-files` sorts them.

| Date | Tool | Version | What was run | What was observed |
|---|---|---|---|---|
| 2026-09-12 | OMG pilot implementation | release 2026-07, kernel 0.61.0, OpenJDK 21.0.12 | `PILOT="$HOME/.local/share/sysml-pilot/sysml"` then `{ printf '%%\n'; cat $(git ls-files -- 'model/*.sysml' 'model/**/*.sysml' \| sort); printf '\n%%\n%%exit\n'; } \| java -cp "$PILOT/jupyter-sysml-kernel-0.61.0-all.jar" org.omg.sysml.interactive.SysMLInteractive "$(cd "$PILOT/sysml.library" && pwd)"` | accepted: seventeen root element lines after the `1>` prompt, `Package Federation_Concerns (<uuid>)` first and `Package VSE_Library (<uuid>)` last, with `Package <SF> Federation_Core (<uuid>)` among them, and no line matching `ERROR:` or `WARNING:` |
| 2026-09-12 | OpenSysML | v0.6.0, built with Go 1.25.14 | `sysml -validate -strict $(git ls-files -- 'model/*.sysml' 'model/**/*.sysml' \| sort)` | accepted, exit 0, seventeen `✓ package` lines from `Federation_Concerns` to `VSE_Library`, then one `✓` line naming all seventeen files and ending `no errors`, and no `warning:` line |

## Forms the reference tools accepted

Before the registers were written, every form the model intended to use was put
to both reference tools in a probe file, once with every form together and once
with each form on its own, with every other form commented out and the ones it
depends on kept. Nineteen forms were probed. Eighteen are accepted by both
validators as written, and none of those fell back to a simpler spelling. One is
refused by both, and the model uses the alternative its row names.

| Form | Pilot | OpenSysML | Form used |
|---|---|---|---|
| `UserStory` usage owning `attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test, VerificationMethod::Demonstration);` and `attribute statement : String = "...";` | accepted | accepted, exit 0 | the probed form |
| nested acceptance criteria (`requirement :>> acceptance { requirement criterionOne { doc } requirement criterionTwo { doc } }`) and `verify US_P02_Nested.acceptance;` in a verification objective | accepted | accepted, exit 0 | the probed form |
| one `#derivation connection` with one `#original` end and twelve `#derive` ends, each `#derive` annotating a requirement usage | accepted | accepted, exit 0 | the probed form |
| two `satisfy` of one story by two parts, story subject `DemoElement`, both parts specialising it | accepted | accepted, exit 0 | the probed form |
| `action <'TestSR22_Foo'> checkTheStoryRegister { @Evidence { kind = "go-test"; location = "..."; } doc /* */ }` and the short name `<'record: 2026-08-28 SR-39'>` | accepted | accepted, exit 0 | the probed form |
| `viewpoint def VP { frame concern : C; }`, `view def V { viewpoint conformsTo : VP; }`, `view v : V { expose ModelProbe::**; doc /* */ }` | accepted | accepted, exit 0 | the probed form |
| `metadata def DecisionRecord { attribute id : String; attribute title : String; }` applied as `@DecisionRecord { id = "AD-0010"; title = "..."; }`, and `@InformationalOnly;` applied bare | accepted | accepted, exit 0 | the probed form |
| `concern def` carrying `require constraint { doc /* */ }` | accepted | accepted, exit 0 | the probed form |
| `@StoryMeta { status = StoryStatus::done; }` with no other attribute | accepted | accepted, exit 0 | the probed form |
| use case with `objective :> US_P10_Browse { subject :>> system = demo; }` and `first stepOpen then stepRead;` | accepted | accepted, exit 0 | the probed form |
| `attribute listenAddress : String = "127.0.0.1:3011";` on a part def, `part def Demo :> DemoElement` owning a `port` | accepted | accepted, exit 0 | the probed form |
| `flow versionEvents of VersionEvent from adapter.listen.events to router.subgraphs.events;` in a composite part usage | accepted | accepted, exit 0 | the probed form |
| `requirement readinessSla { require constraint { readySeconds <= 10.0 } }` nested in a story usage declaring `attribute readySeconds : Rational = 10.0;` | accepted | accepted, exit 0 | the probed form |
| `verify SR_02_X.readinessSla;` beside `verify SR_02_X.acceptance;` in a `verification def` objective | accepted | accepted, exit 0 | the probed form |
| story usage with `stakeholder :>> role : Visitor;` and a second `stakeholder modelOwner : ModelOwner;` | accepted | accepted, exit 0 | the probed form |
| `requirement <'SC-01'> X : DesignConstraint` with `subject :>> system : DemoElement;`, `attribute :>> statement = "...";`, `frame concern : C;` and `requirement :>> acceptance { requirement plainReading { doc } }` | accepted | accepted, exit 0 | the probed form |
| `ref part demo :> demoInstance;` in a composite, `part t : TunnelConnector[0..1] { :>> mode = "named"; }`, a `verification def` with `actor checkly : ChecklyCloud;` and `verify US_P02_Nested.acceptance.criterionOne;`, and `attribute locations : String[1..3] = ("eu-central-1");` | accepted | accepted, exit 0 | the probed form, with the qualified cross-package spelling of the `ref part` settled by the last row |
| `view def ReadabilityRendering { satisfy viewpoint ReadabilityViewpoint; }`, a view definition expressing conformance with a satisfy statement rather than with an owned viewpoint usage | refused: `ERROR:no viable alternative at input 'viewpoint'` | refused, exit 2: `error: expected '{' or ';'`, caret under `ReadabilityViewpoint` | the owned viewpoint usage of the sixth row, `view def V { viewpoint conformsTo : VP; }` |
| `ref part demoFromElsewhere :> ModelProbeOther::demoInstanceElsewhere;` in a composite, the target being a part usage in a second top-level package, named through that package's qualified name | accepted | accepted, exit 0 | the probed form |

A validator accepting a form is not the same as a validator checking what the
author meant by it, and four of the acceptances carry a caveat. `require
constraint { doc /* */ }` holds a documentation comment where a boolean
expression would go, so both validators accept it and neither has anything to
evaluate, and the constraint asserts nothing beyond what its doc says. The
single-element list `("eu-central-1")` is accepted against a multiplicity of one
to three, but the probe did not establish whether either validator reads it as a
one-element sequence or as a parenthesised scalar, whereas the two-element list
in the first row is a genuine sequence. The model gives that attribute no
default, and where the distinction comes to matter the register will write two
elements or drop the parentheses. `viewpoint conformsTo : VP;` declares a
viewpoint usage inside the view definition, which both validators accept and
resolve, but it is a usage rather than the language's conformance relation and
nothing checks that the view satisfies the viewpoint. The refused row asked
whether a view definition may state conformance with `satisfy viewpoint`
instead. Both validators refuse it, `satisfy` of a viewpoint definition is
refused by OpenSysML because a satisfy target must be a usage, and
`satisfy requirement X` exits zero even for a name declared nowhere, because
that spelling declares a new requirement usage rather than referring to one,
which is a trap and not an answer. What both validators accept and resolve is a
`satisfy` whose target is a viewpoint usage. The last row settles the spelling
the model relies on for its session composite: a `ref part` subsetting a part
usage of another top-level package through that package's qualified name is
accepted by both, and neither objects to the two packages referring to each
other.

Five more forms were refused while the registers were written rather than
while they were probed, and each is recorded here with the form used. `public`
and `connector` are reserved words, so the tunnel edge's port is `publicSide`
and the tunnel link's end is `tunnelConnector`. An item and an interface cannot
both be named `HeartbeatPing` under one wildcard import, so the interface is
`Heartbeating`. An interface whose two ports are not conjugate draws a warning,
so `TunnelPort` carries the conjugate features of `HttpClientPort`, and a
twelfth port definition, `AlertClientPort`, was added for the alerting
interface. The OMG pilot refused `satisfy SR_48_AnOptionalCheckSession by
runner` until `SessionRunner` specialised `DemoElement`, because bound features
must have conforming types. And both validators refused
`Federation_LogicalArchitecture::session`, because the composite lives in the
nested package and is reached as
`Federation_LogicalArchitecture::CheckSession::session`.

## Correspondence with the published boards

Every image under `docs/img` is named by exactly one view, and the `images`
subtest of the agreement test holds that both ways. What the subtest cannot
hold is that the drawing and the exposed elements agree, so each image was
opened once and read against the `expose` lines of its view. The table records
what was seen, and the `Inspected` column carries the date of the reading.

Two patterns recur and are stated here rather than in every row. The
storyboard frames and the screenshots of the apps draw the example's servers
and requirements, `ingest`, `parse`, `PIPE-R1` and the rest, which are content
of the example model and no element of this one, and those rows say "example
content". And several boards draw the parts of `demo`, the viewer, the document
app, the router and the three services, on a board whose own view is about a
functional package or a set of stories. Each of those views exposes `demo` as
well, so that what a board draws is what its view names, and the row says so.

| Image | View | What the board shows | Inspected |
|---|---|---|---|
| [`v1-context.png`](../docs/img/v1-context.png) | `v1Context` | the demo among what it touches | 2026-09-12. The pipeline box corresponds to `ExampleModel`, the file that describes it, whose doc says the pipeline runs nowhere. The services and files inside the demo box are parts of `demo`, and `config.json` is `RouterConfiguration` in the domain register. The view exposes all three. |
| [`v2-composition.png`](../docs/img/v2-composition.png) | `v2Composition` | services, ports, schemas, the merged graph | 2026-09-12. The schema types and fields are not elements: the projection part's doc names the seven types and the interface definitions carry the payloads. `config.json` is `RouterConfiguration` in the domain register, exposed by the view. |
| [`v2-subgraph-schemas.png`](../docs/img/v2-subgraph-schemas.png) | `v2Composition` | the three subgraph schemas | 2026-09-12. The types and fields are not elements: the projection part's doc names the types and the interface definitions carry the payloads. The three services are parts of `demo`. |
| [`v2-merged-requirement.png`](../docs/img/v2-merged-requirement.png) | `v2Composition` | one requirement, three fetches | 2026-09-12. The fields are not elements: the projection part's doc names the type and the interface definitions carry the payloads. The three fetches are described in the doc of `demo`. |
| [`v3-runtime.png`](../docs/img/v3-runtime.png) | `v3Runtime` | an edit's seven steps | 2026-09-12. The seven steps are `PropagateAnEdit`, and the three panels beneath them are `NothingMoves`, `ADocumentEdit` and `Reset`, the four action definitions of `EditPropagation`. |
| [`v3-nothing-moves.png`](../docs/img/v3-nothing-moves.png) | `v3Runtime` | ingest raised, nothing moves | 2026-09-12 |
| [`v4-deployment.png`](../docs/img/v4-deployment.png) | `v4Deployment` | container, image layers, publish chain | 2026-09-12. The HEALTHCHECK block is the `healthcheck` attribute of `demo`, and the subcommand it runs is one of the supervisor's `subcommands`. |
| [`v4-container.png`](../docs/img/v4-container.png) | `v4Deployment` | one container, one process tree | 2026-09-12 |
| [`v4-image-layers.png`](../docs/img/v4-image-layers.png) | `v4Deployment` | the image, layer by layer | 2026-09-12 |
| [`v5-adapter.png`](../docs/img/v5-adapter.png) | `v5Adapter` | four packages, loop back, refusal path | 2026-09-12. The router at the edge is `Router`, exposed by the view. The fixtures box corresponds to the second fixture named in the doc of `VC_SR_45`. |
| [`a3-l0-model-side.png`](../docs/img/a3-l0-model-side.png) | `l0Sheet` | the argument in eight steps | 2026-09-12. The registry, the browser and the machine the compose step runs on are `ContainerRegistry`, `Browser` and `MaintainerMachine`, exposed by the view. The counts of servers and requirements are the example's. |
| [`a3-l0-legend.png`](../docs/img/a3-l0-legend.png) | `l0Sheet` | four decisions and the legend | 2026-09-12. The four records are tags on `demo` and `Router`. |
| [`a3-l2b-model-side.png`](../docs/img/a3-l2b-model-side.png) | `l2bSheet` | how the number is made | 2026-09-12. The five servers are the example's. The services under "where it runs" are parts of `demo`, exposed by the view, and the four decision records are tags on two of them. |
| [`a3-l2b-wiring-states.png`](../docs/img/a3-l2b-wiring-states.png) | `l2bSheet` | the wiring in three states | 2026-09-12. The servers are the example's. |
| [`a3-l2b-arithmetic.png`](../docs/img/a3-l2b-arithmetic.png) | `l2bSheet` | four states, their cuts and verdicts | 2026-09-12 |
| [`a3-l2b-physical.png`](../docs/img/a3-l2b-physical.png) | `l2bSheet` | three services, one router | 2026-09-12. Parts of `demo`, exposed by the view. |
| [`stories-journey.png`](../docs/img/stories-journey.png) | `storyboard` | twelve stories in order | 2026-09-12. The sketch beneath the stories draws parts of `demo`, exposed by the view. |
| [`stories-personas.png`](../docs/img/stories-personas.png) | `storyboard` | the three personas | 2026-09-12 |
| [`us01-launch.png`](../docs/img/us01-launch.png) | `storyboard` | one command, then two tabs | 2026-09-12. Example content in both tabs. |
| [`us03-nothing-moves.png`](../docs/img/us03-nothing-moves.png) | `storyboard` | ingest raised, nothing else moved | 2026-09-12. Example content. |
| [`us04-bottleneck-moves.png`](../docs/img/us04-bottleneck-moves.png) | `storyboard` | parse raised, the cut migrates | 2026-09-12. Example content. |
| [`us06-document.png`](../docs/img/us06-document.png) | `storyboard` | the document, shipped state | 2026-09-12. Example content. |
| [`us09-edit-from-document.png`](../docs/img/us09-edit-from-document.png) | `storyboard` | document edit, the viewer follows | 2026-09-12. Example content. |
| [`us10-viewer-to-document.png`](../docs/img/us10-viewer-to-document.png) | `storyboard` | viewer edit, the document follows | 2026-09-12. Example content. |
| [`us11-query.png`](../docs/img/us11-query.png) | `storyboard` | one query, three services | 2026-09-12. Example content. |
| [`us12-reset.png`](../docs/img/us12-reset.png) | `storyboard` | both apps after a reset | 2026-09-12. Example content. |
| [`architecture-five-views.png`](../docs/img/architecture-five-views.png) | `fiveViews` | five views and their questions | 2026-09-12 |
| [`overview-sketch.png`](../docs/img/overview-sketch.png) | `fiveViews` | two apps, one graph, three services | 2026-09-12. Parts of `demo`, exposed by the view. |
| [`app-viewer-shipped.png`](../docs/img/app-viewer-shipped.png) | `appShots` | the viewer, shipped state | 2026-09-12. Example content. |
| [`app-viewer-bottleneck-moved.png`](../docs/img/app-viewer-bottleneck-moved.png) | `appShots` | the viewer, bottleneck moved | 2026-09-12. Example content. |
| [`app-viewer-passing.png`](../docs/img/app-viewer-passing.png) | `appShots` | the viewer, `PIPE-R1` passing | 2026-09-12. Example content. |
| [`app-viewer-limit-edit.png`](../docs/img/app-viewer-limit-edit.png) | `appShots` | the viewer, the limit edited | 2026-09-12. Example content. |
| [`app-viewer-refusal.png`](../docs/img/app-viewer-refusal.png) | `appShots` | the viewer refusing a value | 2026-09-12. Example content. |
| [`app-document-tree.png`](../docs/img/app-document-tree.png) | `appShots` | the document tree | 2026-09-12. Example content. |
| [`app-both-exclusion.png`](../docs/img/app-both-exclusion.png) | `appShots` | an exclusion both apps honour | 2026-09-12. Example content. |
| [`app-document-depth.png`](../docs/img/app-document-depth.png) | `appShots` | the depth limit | 2026-09-12. The six-level limit is the `maxDepth` attribute of `RequirementsDocument`, exposed by the view, and three validation cases record a run against it. Example content. |

## Tests in the model

A verification case owns one action per piece of evidence, and the action's
short name is the evidence's own name. A Go test is named by its function,
`action <'TestSR02_ReadyWithinTenSeconds'> readyWithinTenSeconds`, and the
`@Evidence` tag on it gives the kind, `go-test`, and the file,
`cmd/sysml-federation/main_test.go`. Sixty-six actions name a Go test this
way, and the file matters, because two functions share a name across two
packages and the agreement test compares name and file together. The
sixty-seventh, the second `TestSR25_InvalidValuesAreRefused` in the projection
package, carries no short name because the name is taken by the first, and the
agreement test reads the function name from that action's doc. A recorded run
of the demo is named by its row in the example README's verification record,
`<'record: 2026-08-28 SR-02'>`, with the record's anchor as its location. A row
names one or more requirements, and it stands under every stakeholder story
those requirements derive from, so one row stands under several cases when it
bears on several stories. The two validators are actions of kind `validator`
named `<'omg-pilot 2026-07'>` and `<'opensysml v0.6.0'>`, with the make target
that runs them as their location. A make target is an action of kind
`make-target` named as it is typed, `<'make check-tracked'>`. A workflow is
named by its file, `<'model.yml'>`, or by its file and step,
`<'publish.yml: Read the manifest back'>`. What is read rather than run, the
router's outbound paths or the module file, is an action of kind `analysis` or
`inspection`.

A case names only evidence that exists. Four stories ask for more than the
repository holds. SR-01's criterion `recordedOnThreePlatforms` has one platform
recorded of three. SR-10's second criterion, the apps rendering with the host
offline, has no recorded run. SR-40 and SR-43 name `Test` among their methods
and have no test function of their own, so their cases carry the inspection and
the recorded run alone. Each case claims what is there and nothing more.

The check session, which a later pull request adds, brings its own cases with
it. Every live check will be a verification case of its own, carrying the
check's steps, and the session is already a composite with its behaviour:
`Federation_LogicalArchitecture::CheckSession::session` holds the runner, the
two tunnels, the collector, the viewer, the optional private-location agent
and the external services they talk to. The six session parts carry the name
of the compose service each maps to, and so does the reference to the demo
under test, while the three `ref part`s for the external services, the
monitoring service, the tunnel edge and the alert receiver, carry none. The
nested `package CheckSession` in the logical architecture mirrors
`Federation_Context::CheckSession`, which holds those externals. The two share
a simple name, so every reference to either is qualified. The check suite's
decomposition is `part def CheckSuiteProject :> CheckSuite`, and `RunASession`
in the functional architecture holds the runner's twelve steps.
Until the session lands, `VC_SR_48` carries its objective and no action, and the
view `session` names no image.

`TestSR46_ModelAndRepositoryAgree` in
[`internal/trace`](../internal/trace/trace_test.go) is the evidence for SR-46.
It reads the registers as text and compares them with the repository around
them, one subtest per agreement.

- `identifiers`: every short name the model writes is declared exactly once,
  every decision record the model names exists under `docs/decisions`, and
  every record there is named in the model.
- `requirementsAffected`: every requirement a decision record names as affected
  is a short name the model declares.
- `goTests`: the Go test functions of the requirement scheme and the ones the
  verification register names are the same set, counted per file.
- `checkFiles`: the check files of the check project and the file names the two
  case registers quote are the same set, and every suite file is exercised by
  exactly one validation case.
- `images`: every image under `docs/img` is named by exactly one view, and every
  image a view names exists.
- `coverage`: a system story that is done is verified by a case, satisfied in
  the logical architecture and a derived end of a derivation connection, a
  stakeholder story that is done has exactly one validation case, and a story in
  progress is reported by name and exempt from the first two.
- `checkInventory`: the check register and the manifest the check project reads
  carry the same checks with the same fields, each constructed once.
- `sessionParts`: the services the compose file of the check session starts are
  the parts the session composite carries, and nothing else.

Three of these read the check project. `checkFiles` and `checkInventory`
compare two empty sides until it exists, which is agreement rather than absence,
and `sessionParts` skips, naming the compose file it waits for.

## Tailoring

Profile: light. Status only on stories, no priorities, no risk register, no allocation keyword, no process documents, no copy of the practice's own specification.

The layout follows a story-driven practice of agile model-based systems
engineering, trimmed to what a demo of this size can carry. Seven directories
of the fuller layout are absent: `base-architecture/`, `product-architecture/`,
`parametrics/`, `processes/`, `logical-architecture/allocations/`, `variations/`
and `sandbox/`. Two were added: `constraints/`, because a design constraint
carries a plain statement and belongs to no story, and `views/`, because the
published boards needed a place to be named. The particulars of the tailoring:

- Sixteen system stories carry more than one stakeholder feature: SR-12,
  SR-13, SR-14, SR-15, SR-16, SR-22, SR-23, SR-26, SR-27, SR-28, SR-29, SR-30,
  SR-32, SR-37, SR-39 and SR-41. The definition gives a story one role, and
  these add a second or a third stakeholder because the requirement serves more
  than one person.
- One concern is framed by no story, `ReadBuildChange`, the maintainer's and
  the contributor's question. It is carried by the design constraints SC-01,
  SC-05 and SC-06, which frame it, and it is tagged `@InformationalOnly` to say
  so.
- Three stories formalise their benefit as a nested requirement with a
  constraint: SR-02 (ready within ten seconds), SR-06 (each platform's layers
  within 80 MB) and SR-39 (reflected within two seconds). The rest keep an
  informal benefit and therefore supply no trade-study criteria.
- No rendering is declared on any view. The images are drawn by hand, and a
  view names the elements each one is about.
- The GraphQL schema the services share, drawn on the composition boards, is
  not modelled as elements. The interface definitions carry its payloads and
  the projection's doc names its types.
- Allocation is expressed by `satisfy`, ninety-four of them in the components
  register, because the allocation keyword has no recorded validation run
  against either reference tool, and a form neither has been seen to accept does
  not go into the model. The elements allocated to are twelve. The nine of the
  requirement scheme are `Adapter`, `CapacityService`, `DocumentService`,
  `Router`, `ModelViewer`, `RequirementsDocument`, `Demo` (the image),
  `ExampleModel` and `Repository`. `DemoModel` and `CheckSuite` came with the
  model. `SessionRunner` is the twelfth, because SR-48 is satisfied by it beside
  the check suite and the repository.

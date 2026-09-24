# Planning the build

*Roar Georgsen, 27 August 2026*

Part 9 of 13 in [Federating a systems model](../README.md), written for release v0.3.0.

> [!IMPORTANT]
> **The story so far**
>
> The demo is a SysML v2 model of a five-server query pipeline, served as one subgraph of a federated GraphQL graph. Beside it sit a capacity service that works out where the pipeline's bottleneck is, and a document service that owns an editorial ordering of its requirements, all shipped as one container image. A subgraph is a GraphQL service that owns part of a shared schema. The router is the process that composes the subgraphs into one graph and answers queries against it. [The architecture in one sitting](01-the-architecture-in-one-sitting.md) walks through the parts. The design phase is over. This part is about the plan that turned the design record into code, and the decisions the plan had to take that the record had left open.

## From record to plan

The design phase ended with the fourth gate approved, twenty-six decision records accepted, and an architecture description that names every type, every field and every process. What it didn't have was a sequence of tasks. The plan had to be detailed enough to implement from without re-deciding anything, and it opens on the finish line:

> Build the pipeline demo the design phase specified: a generic SysML v2 adapter subgraph, a capacity service, a document service, two vanilla web apps and a Cosmo router, shipped as one container image launched by one `docker run`

Each of the forty-five requirements was to be verified in the way its own text says.

The records had already fixed the stack, and the plan restates it in one line:

| Part | Choice |
|---|---|
| Language | Go 1.27, with the toolchain pinned locally |
| The three subgraphs | <span class="term" data-term="gqlgen">gqlgen</span> 0.17.94, with federation v2 and WebSocket subscriptions over the `coder/websocket` package gqlgen already requires |
| Router | WunderGraph Cosmo router 0.343.1, as the child process |
| Composition | wgc 0.130.1, run only on the maintainer's machine, since [composition is a maintainer step with committed output](../decisions/AD-0012-composition-committed.md) |
| Web apps | vanilla <span class="term" data-term="es-modules">ES modules</span> with one vendored file, SortableJS 1.15.7, which is the whole of [the vanilla web apps decision](../decisions/AD-0017-vanilla-web-apps.md) |
| Image | a <span class="term" data-term="distroless">distroless</span> static base, [published to the GitHub Container Registry by a tag-triggered workflow](../decisions/AD-0020-publish-on-tags.md) |

The plan has three layers:

1. A top-level document carries the global constraints, the working method and one line per task.
2. One detail document per phase carries the full text of each task: every file, every test, every command and the output it should print.
3. A coverage map closes the set. It lists every requirement and constraint against the task that meets it and the test or demonstration that verifies it.

The top-level document and the one task in hand are enough to build from.

## Five phases, one pull request each

| Phase | What it builds |
|---|---|
| 0 | repository policy, and the spikes (the fifth waits for the image) |
| 1 | the example model and the adapter core |
| 2 | the three subgraphs, composed |
| 3 | the supervisor, with one binary, one process tree and one port |
| 4 | the two web apps |
| 5 | the image, and the first release |

### Phase 0: policy and spikes

Before the five phases of building comes a phase zero. Like every other phase it gets its own branch and its own pull request. It changes the repository's policy and settles the spikes.

The allowlist `.gitignore` gains rules for the test helpers, the gqlgen configuration, the web app sources, the vendored licence, the composed router configuration and the NOTICE file, each proved with `git check-ignore`. The Makefile gains `generate`, `compose`, `image` and `run` <span class="term" data-term="make-target">targets</span>, and the empty-interface rule stops reporting generated files, as [the generated code decision](../decisions/AD-0016-generated-code-exempt.md) says.

Then come the five spikes the architecture names. Two concern the model and the graph:

- the ports and connections syntax quoted from the OMG training material, checked in the example model's own shape
- a nested-list `@requires`, composed and served through the router into gqlgen

Three concern the container:

- whether the router starts from environment variables alone on a distroless image, and what its readiness path proves when the subgraphs are down
- whether a `COPY --from` of the multi-platform router image resolves the target platform
- whether the container reaches anything outside itself under `--network none`

The first four run ahead of any code that depends on them. The fifth needs the built image, and waits for Phase 5. A spike that fails switches the plan to the fallback the architecture names for it, and the switch is recorded before the next phase starts. [Five spikes before the first line](09-five-spikes-before-the-first-line.md) has the outcomes.

### Phase 1: the adapter core

The example model and the adapter core come next. That means the lexer with byte spans, the AST, the structural parser, and expressions with precedence. Then port definitions, connections, requirements and verification cases, and finally the model package that projects the AST onto the types the schema shows.

Phase 1 closes the ten requirements on the adapter. They start from the generic projection, the plainly typed set of GraphQL types the adapter serves in place of the metamodel, proved against a second fixture model. They run through refusal of unsupported syntax, connection direction, and short names as identifiers. Edits patch the source, expression-bound values are evaluated and read-only, invalid values are rejected, and the reference tools accept the model.

The phase ends with the second fixture, a warehouse model with other names and wiring, and the walk that proves no identifier from the example appears in the adapter.

### Phase 2: three subgraphs, composed

Phase 2 builds the three subgraphs and composes them.

The capacity service gets its `flow` package first, [the rollup as maximum flow with the source-side cut](../decisions/AD-0007-rollup-as-maximum-flow.md). The cut is the set of servers whose combined throughput bounds the pipeline, which is what the demo calls the bottleneck. A differential test checks the result against the naive minimum and sum. Then come the verdict precedence and every reason template. (A verdict is the capacity service's judgement on one requirement: PASS, FAIL, INCONCLUSIVE or ERROR.) The gqlgen schema and the entity resolvers follow.

The document service gets its `tree` package with dotted-decimal numbering, then its service with WebSocket subscriptions. The adapter gets its schema, its store and its server.

To close the phase there's the committed router configuration, and four tests:

- the drift test, which fails if a schema file and the composed configuration disagree
- a loopback test
- a WebSocket composition test
- the isolation test, which proves document operations leave the model's version alone

Seventeen requirements close here, more than in any other phase.

### Phase 3: one binary, one port

With the subgraphs composed, Phase 3 brings the supervisor, [one binary, one process tree, one port](../decisions/AD-0011-one-binary-one-port.md). The `serve` subcommand runs the subgraphs as goroutines and [the router as a child process](../decisions/AD-0010-router-as-child-process.md). A UI server serves both apps and proxies the router on the one published port. The tests need neither Docker nor the vendor's binary.

The phase's second task is a set of demonstrations against the running stack, with the router extracted from the pinned image. They cover:

- the four paths
- one query across three services
- a version event over server-sent events
- refusals through the router
- the playground
- a clean stop on SIGTERM, the signal Docker sends to stop a container
- a restart
- an inspection of imports that proves the services share no code

### Phase 4: the two apps

Phase 4 replaces the placeholder pages with the two apps and the shared client module. The viewer gets its tokeniser, its sketch drawn from the wiring, its edit panel, and its requirement blocks with their verdicts and reasons. The document app gets its numbered tree, drag and drop, headings and prose, exclude and restore, and its inputs for the current value and the limit. NOTICE and the vendored licence go beside them.

### Phase 5: the image

Phase 5 builds the image: cross-compiled Go, the router copied from the pinned image, the licence fetched by `ADD --checksum`, a nonroot distroless base, and a `HEALTHCHECK`. The publish workflow runs on `v*` tags and reads the manifest back, failing if either platform's image exceeds 80,000,000 bytes.

The example's documentation is completed, and the air-gap run happens. Then `v0.1.0` is tagged, and the last requirement, that one command runs the demo, is verified from a logged-out host.

![The layers of the shipped image, from the distroless base through the router binary to the supervisor and its embedded apps](../img/v4-image-layers.png)

*The layers of the shipped image, from the distroless base through the router binary to the supervisor and its embedded apps. Cut from the [architecture views](../architecture/architecture-views.pdf).*

## The working method

Test first, with the test as the specification. A new implementation file never appears in a package that has no test yet, and every task creates the test file before the implementation file. The tests use the two tracked helper packages, [assert and tabletest](../decisions/AD-0022-track-internal-helpers.md), and not a library that takes the empty interface throughout. A test that verifies a numbered requirement is named for it.

`make check` runs after every task. It covers the toolchain, formatting, vet across five operating system and architecture pairs, lint, the <span class="term" data-term="race-detector">race detector</span> and the tracked-files check. `make preflight` runs before every push, and adds a 70 per cent coverage floor.

Every task ends with one commit. Every phase ends with:

- `make preflight` green
- a push, and a pull request whose body lists the requirements the phase satisfies
- a review against its tasks and the requirements it claims, then a merge
- one entry in the engineering log
- the next branch cut from `main`

Requirements verified by demonstration have checklists in the tasks, fourteen items for the viewer and twelve for the document app. Their outcomes go into the example's verification record, with the date and host.

The end-to-end proof is the worked example of [the capacity model](05-from-use-cases-to-requirements.md), observed through the router and both apps:

1. The shipped state fails at parse.
2. Raising ingest to 3000 changes nothing.
3. Raising parse to 1700 moves the bottleneck to the index pair at 1400.
4. Raising indexA to 900 passes at 1600, while the model requirement `PIPE-R1.4` still fails.
5. Back in the shipped state, dropping the limit of `PIPE-R1` to 1000 lets the same 1200 pass.
6. A reorder in the document leaves the model's version at 1.
7. Reset from either app restores the shipped state within two seconds.

### Line figures

The plan carries line figures, because the constraint that the adapter be readable in an afternoon means little unless somebody counts. The first estimate for hand-written Go under `adapter/` was 2000 lines. It lasted until the detail document for Phase 1 had every file's text written out in full, and the mandated text came to more in total:

| Adapter package | First estimate | Measured in the detail document | Revised figure |
|---|---|---|---|
| `syntax` | 800 | about 1261 | 1300 |
| `model` | 650 | about 932 | 1000 |
| `projection` and `serve` together | 550 | about 389 | 450 |
| Total | 2000 | about 2582 | 2750 |

The revised figure is the measured counts plus a margin. The two example services get 600 lines each and the supervisor 700. Each web app gets 900 lines of JavaScript and 300 of CSS, with the shared client counted against the document app. Tests and generated files are outside the figures.

Having to revise the figure as soon as the real text existed settled what kind of number it is. These are estimates of the expected scale, and not limits. Correctness comes first. A component that needs more lines to be right takes them, and the figure is then revised to what the work turned out to be. The counts are still measured and reported at each phase close, because the scale is worth seeing, but nothing is trimmed and no fix is refused to stay under one.

A figure written as a limit reaches for the wrong things first. The cheapest lines to give up are the doc comments, the refusals that give a position, and the guards that make a refusal honest. Those are the parts a reader with an afternoon most needs.

## Decisions the plan made

A plan detailed enough to implement from has to decide things the design record didn't, and it has to say so instead of burying them in task text. The adapter core fixed fourteen, the subgraphs ten, the supervisor six and the web apps nine, most of them at the head of the phase's detail document, with a reason.

### In the example model

In the example model, the abstract shared definition is `Component`, not `Node`. The merged graph already has a `Node` type from the document service, and the viewer draws nodes, so `Component` collides with nothing in any schema.

Port direction is carried by directed items inside the port definitions: `in item queries : Query` in the input definition and `out item queries : Query` in the output one. That's the shape the language itself gives ports, and the plan chose it over the conjugation operator or a direction prefix on the usage. The port names are `input` and `output`, because `in` and `out` are reserved words. A port usage's projected direction is derived from its definition's items. This is how [connection direction from the order of the ends](../decisions/AD-0009-connection-direction.md) is realised in text.

The model declares its own millisecond. The shipped library declares seconds, minutes, hours and days as duration units, and no `ms`. So the example carries an `<ms>` declaration copied from the library's own pattern for the millimetre. If a validation tool's own library declares `ms`, the line is deleted and nothing else changes.

### In the adapter

A model is immutable once built. Patching a literal returns a new model with the version incremented, and the store swaps a pointer. That's what stops a query from ever seeing the text and the projection disagree under concurrent reads. The projected types are `Model`, `Part`, `Attribute`, `Port`, `Connection`, `Requirement` and `VerificationCase`, the set [the curated projection decision](../decisions/AD-0005-curated-generic-projection.md) chose over generated metamodel types.

The constraint's operand order is free. [The constraint shape](../decisions/AD-0008-quantity-from-constraint.md) fixes that a requirement's quantity, comparison and limit are read from its constraint. It doesn't fix which side the subject sits on, so the parser accepts either, and flips the comparison when the chain rooted at the subject is on the right. The warehouse fixture writes it the other way round on purpose.

Identifiers follow [the short-name decision](../decisions/AD-0018-short-names-as-keys.md): the short name if declared, else the qualified name, and duplicates make the model refuse to load.

<details markdown="1">
<summary>Under the bonnet: two choices about Go code</summary>

Sum types in the AST, meaning values that can be one of several kinds, are typed slices per member kind everywhere except expressions. There, one small named interface with three implementations does the work.

Both packages panic internally with a typed error and recover at their one entry point. The plan reckons that saves roughly 120 lines of error plumbing.

</details>

### On the federation side

gqlgen is pinned as a tool directive in `go.mod`, and its generated files are committed, so continuous integration needs no generator.

The capacity service uses gqlgen's explicit requires strategy, the only one documented to support nested and array fields. Its two populate functions carry the only two permitted empty-interface lines in the hand-written code. That took the tracked baseline from 2 to 4, with the justification in the pull request.

Pure code lives in `flow` and `tree` sub-packages beside the generated code, because the generated enum is also called `Verdict` and the generated document type is also called `Node`.

The adapter's store owns the version counter, and sets it on every accepted mutation and on reset. So the counter only ever grows, even across a reset, whatever version a fresh parse of the text would carry. Every operation takes a snapshot of the current model, so one query never mixes two versions.

Excluding a requirement that has children promotes the children into its place. Restoring it puts it back, childless, as the last child of its former parent.

### In the viewer

The largest departure is in the viewer. [The viewer decision](../decisions/AD-0026-viewer-shows-text-and-wiring.md) said the editable numbers would appear as inline inputs, at the literals' positions in the text pane.

The projection carries each attribute's editability, value and unit, but no source spans. So an inline input could only be placed by a client-side search for the literal, by part name and attribute name. That's a second, weaker parser of a syntax the adapter has already parsed. The limit of a requirement is no better placed. The projection publishes the quantity its constraint reads and the limit's own value, but not the name of the attribute the limit binds, so a search for its literal has even less to go on. Neither search is impossible. Both read the notation a second time and less well, and that's the objection.

So the plan puts the editable literals in a panel inside the text pane, above the model text. It has a control for each editable attribute of the parts directly inside a root part, and one for a requirement's limit when that limit is editable and the analysis has reached a verdict. The served text still shows the edited number where the literal was, because the adapter patches it. I still think the panel is the right call for a demo whose point is the model text, and the record now carries it as the decision.

![The adapter's packages, from the source text through the parser and the model to the served projection](../img/v5-adapter.png)

*The adapter's packages, from the source text through the parser and the model to the served projection.*

## Validation tools

Two tools serve the requirement that the reference tools accept the model, and [the decision on the language target](../decisions/AD-0019-sysml-2-0-target.md) fixed which. One is the <span class="term" data-term="pilot-implementation">OMG pilot implementation</span>, release 2026-07, run in batch on Java 21. The other is the <span class="term" data-term="opensysml">OpenSysML</span> command line at version 0.2.1, with strict validation. Both are installed once in Phase 0, and run on the probe model and then on both fixtures. A third tool, a commercial validator, was considered and dropped on its licence terms.

Validation of the example runs locally and never in continuous integration. That still holds at release v0.3.0. The demo's own model, which came later, is a different case, and is validated in CI on every change, as [A model of the demo itself](12-a-model-of-the-demo-itself.md) describes.

## Open points carried into the build

- The escape hatch for elements nobody has projected, which [the curated projection decision](../decisions/AD-0005-curated-generic-projection.md) leaves open, is still open. A model that uses a construct outside the subset doesn't load, and the plan builds nothing to soften that.
- Inline inputs in the viewer would need byte-range span fields for attributes and limits on the adapter's schema. The plan adds none, and the panel stands until somebody decides the schema change is worth it.
- The A3 set published so far is the L0 and L2b sheets, described in [the A3 article](07-an-a3-sheet-for-a-fifteen-minute-reader.md). The L1 sheet, where the contract between the adapter and the capacity service is meant to fit as its quantification block, and the L2a sheet are still to be drawn.

---

Previous: [An A3 sheet for a fifteen-minute reader](07-an-a3-sheet-for-a-fifteen-minute-reader.md) · Index: [Federating a systems model](../README.md) · Next: [Five spikes before the first line](09-five-spikes-before-the-first-line.md)

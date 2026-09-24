# Five views and twenty-six decisions

*Roar Georgsen, 27 August 2026*

Part 7 of 13 in [Federating a systems model](../README.md), written for release v0.3.0.

> [!IMPORTANT]
> **The story so far**
>
> The demo is a SysML v2 model of a five-server query pipeline, published as one federated GraphQL graph by three services in one container, with a capacity analysis and a requirements document joined to the model. Part 6 turned the use cases into requirements. This part is the third gate: the architecture that answered them, drawn as five views for five kinds of reader, with twenty-six decisions behind it.

## Five readers, five views

An architecture description is for people who want different things from the same system. I borrowed the vocabulary of [ISO/IEC/IEEE 42010](https://en.wikipedia.org/wiki/ISO/IEC_42010) for it: stakeholders, their concerns, and views that answer them. I followed its concepts without claiming to conform to the standard, and nobody has audited the description against it.

> [!NOTE]
> **View and viewpoint**
>
> In ISO/IEC/IEEE 42010, a view is one picture of a system drawn for a particular set of concerns, such as what runs where or how a request flows. A viewpoint is the recipe for that picture: which concerns it answers, for whom, and in what notation. Five views of one system can drift apart, so a good description also says how they correspond.

Five readers are named, each with a question:

- **The visitor**: does it launch, and does it show something I wouldn't have believed?
- **The model owner**: does the model stay the source, and does an edit anywhere land in it?
- **The document owner**: can I shape the document without touching the model?
- **A maintainer**: can I read it in an afternoon, build it with Go alone, and change one service at a time?
- **An organisation deciding whether it fits**: what do the services agree on, and what would we replace?

The last question is the reason the series exists. The reading at the third gate found the description had promised that reader an answer and never given one, and the answer it gained is the adoption contract in [The architecture in one sitting](01-the-architecture-in-one-sitting.md).

You'll find the text of each view, with the field ownership table and the process tree, on [the architecture page](../architecture/README.md), which together with the decision records is the record of the architecture. Here's what each view is for.

## The five views

### Context: what's inside and what isn't

![The context view: the demo as one box, with the browser, the Docker host and the registry outside it](../img/v1-context.png)

*The context view draws the demo as one box, and the browser, the Docker host and the container registry outside it. All the boards are in the [architecture views PDF](../architecture/architecture-views.pdf).*

Everything the demo needs runs inside one container. Outside it are a browser, the Docker host, and the registry the image is pulled from once, and nothing else. The demo is built to make no connection beyond that boundary while it runs.

### Composition: who answers for which field

![The composition view: three subgraph schemas, their entity keys and the merged Requirement](../img/v2-composition.png)

*The composition view shows which service answers for every field, and the merged graph in which one requirement carries fields from all three.*

This is the view the deciding organisation reads. Each of the three subgraphs answers for its own fields, and they never overlap. They meet on one entity key, the SysML short name. The capacity service asks the router, through `@requires`, for the parts and connections it needs, and the document service answers for an id it has never heard of by saying the requirement isn't in the document.

### Runtime: one edit, end to end

The runtime view follows one edit through the router, the adapter and the capacity service to both apps, and [The architecture in one sitting](01-the-architecture-in-one-sitting.md) walks through its seven steps. What the view adds is the asymmetry. An edit to the document goes to the document service alone, changes the document's version and leaves the model's untouched, which is the whole of what the document owner is promised.

### Deployment: one container, one process tree

![The deployment view: the process tree, the Dockerfile stages and the tag-triggered publish](../img/v4-deployment.png)

*The deployment view shows one container with one process tree, the build stages, and the workflow that publishes on a version tag.*

One Go binary supervises three services and the vendor's router inside one container, and only one port is published. The image is built for two processor architectures and published when I push a version tag.

### The adapter: text in, projection out

![The adapter view: files through lexer, parser, resolver and projection to gqlgen, with a patch arrow back to the AST](../img/v5-adapter.png)

*The adapter view is a pipeline of four packages, with the patch path from an edit back to the syntax tree.*

The adapter parses the model text into a tree in which every node remembers where it came from in the source. An edit replaces one number at its position and parses again, so the served text and the projection can never disagree, and the adapter never holds a second copy of the model.

## Twenty-six decisions

The main choices behind those views each have a <span class="term" data-term="decision-record">decision record</span>, with its context, the alternatives that lost, and the consequences. Twenty-four came from decisions already taken in the brief, the plan and the engineering log. Two more turned up when the traceability showed requirements with no decision behind them. The [decision index](../decisions/README.md) lists thirty-one by release v0.3.0, because five more came during the build and after it.

The twenty-six fall into five groups:

- **The shape of the system**: [federation over a single service](../decisions/AD-0001-federation-over-single-service.md), [Cosmo as the platform](../decisions/AD-0002-cosmo-as-platform.md), [one binary, one port](../decisions/AD-0011-one-binary-one-port.md), [the router as a child process](../decisions/AD-0010-router-as-child-process.md), [composition committed and tested for drift](../decisions/AD-0012-composition-committed.md) and [telemetry off](../decisions/AD-0013-telemetry-off.md).
- **The analysis**: [an idealised capacity model](../decisions/AD-0006-idealised-capacity-model.md), [the rollup as a maximum flow](../decisions/AD-0007-rollup-as-maximum-flow.md), [the quantity read from the constraint](../decisions/AD-0008-quantity-from-constraint.md), [connection direction](../decisions/AD-0009-connection-direction.md) and [reasons from templates](../decisions/AD-0024-reason-templates.md).
- **The adapter**: [files instead of a repository](../decisions/AD-0003-adapter-reads-files.md), [editing as scaffolding](../decisions/AD-0004-editing-as-scaffolding.md), [a curated projection](../decisions/AD-0005-curated-generic-projection.md), [a hand-written subset parser](../decisions/AD-0015-hand-written-subset-parser.md), [short names as keys](../decisions/AD-0018-short-names-as-keys.md) and [SysML 2.0 as the target](../decisions/AD-0019-sysml-2-0-target.md).
- **The two apps**: [events that carry a version number](../decisions/AD-0014-version-events.md), [plain web apps with no build step](../decisions/AD-0017-vanilla-web-apps.md), [the document owning its structure](../decisions/AD-0025-document-owns-its-structure.md) and [the model's text beside its wiring in the viewer](../decisions/AD-0026-viewer-shows-text-and-wiring.md).
- **The repository**: [generated code exempt from the empty-interface rule](../decisions/AD-0016-generated-code-exempt.md), [publishing on tags](../decisions/AD-0020-publish-on-tags.md), [the Markdown record and the A3 overview](../decisions/AD-0021-architecture-record.md), [tracked test helpers](../decisions/AD-0022-track-internal-helpers.md) and [a light requirements scheme](../decisions/AD-0023-light-requirements-scheme.md).

The reading at the third gate overturned none of them. Its fixes were to mechanisms and to gaps in the description, which told me the decisions had settled before the description was drawn.

For a reader with only a quarter of an hour, I cut all of this down to single sheets of paper, which is where [An A3 sheet for a fifteen-minute reader](07-an-a3-sheet-for-a-fifteen-minute-reader.md) picks up.

---

Previous: [From use cases to requirements](05-from-use-cases-to-requirements.md) · Index: [Federating a systems model](../README.md) · Next: [An A3 sheet for a fifteen-minute reader](07-an-a3-sheet-for-a-fifteen-minute-reader.md)

# Why federate a systems model

*Roar Georgsen, 27 August 2026*

Part 1 of 13 in [Federating a systems model](../README.md), written for release v0.3.0.

> [!IMPORTANT]
> **Where this starts**
>
> With a problem I keep running into. Engineering organisations describe one system in several documents that drift apart, and the modelling tools meant to fix that lock the model away. This part sets out the idea I think gets round it, and the small demo I built to show it. The rest of the series follows that demo from design to a published container.

## One system, many documents

Every engineering organisation I've worked with keeps its system in several places at once. The requirements live in one file and the interface definitions in another. The power budget sits in a spreadsheet, and the hazard analysis is somewhere a third team controls. All of them describe the same machine. Keeping them in agreement is manual work that nobody enjoys and everybody puts off until a review forces the issue.

[Model-based systems engineering](https://www.sebokwiki.org/wiki/Model-Based_Systems_Engineering_(MBSE)) was supposed to end that. The documents give way to a model, one structured description holding the system's parts, the properties they carry, the requirements they must meet and the relationships between all of it. Documents become views of the model rather than the place the information lives. Consistency stops being something you check in a review. Change a component's power draw and everything downstream can be recomputed, including which requirements now fail and which tests need running again.

That promise has been five years from mainstream adoption for about 25 years. People disagree about why, and the least disputed explanation is the tooling.

Models live inside proprietary environments. They're stored in binary formats, or in a vendor's own flavour of <span class="term" data-term="xmi">XMI</span> that only that vendor reads back reliably, and you reach them through an API that is either missing or shaped differently in every product. A model that only its author's tool can read is a document with extra ceremony, and people treat it like one.

Version control is worse. Several of the big environments still assume [Subversion](https://en.wikipedia.org/wiki/Apache_Subversion) or [TFS](https://en.wikipedia.org/wiki/Team_Foundation_Server), treat the model as a binary file you lock rather than something you merge, and offer nothing a build server can drive. Branching becomes an administrative chore. <span class="term" data-term="ci">Continuous integration</span> is not a concept the tool has heard of, so the model can't join the workflow every other discipline now takes for granted.

Vendors will tell you integration is a solved problem, and they aren't exactly lying. Connectors to requirements tools and test management exist, and so do bridges to <span class="term" data-term="plm">PLM</span>. They are also, with tiresome regularity, a separately licensed product on top of the one you already bought. What arrives is often a bridge written against a Java desktop stack that has aged badly. You configure it through a dialogue box rather than a file, and it's liable to break at the next upgrade of either end.

Two-way synchronisation exists as well, and where somebody has set it up properly it works. Setting it up takes a different skill from the one the engineer who needs it has, so more often than not it stays unconfigured and everyone falls back to export.

That is what lock-in looks like in practice. It's rarely a clause in a contract. It's a pile of small daily frictions that make export always the cheapest option in the moment. Requirements go to Word for review. Interface definitions reach the software team in a spreadsheet, exported where the tool allows it and retyped where it doesn't. The power budget lives in Excel because that's where the person who owns it works.

Every one of those copies starts drifting the moment it is made.

## We've tried this before

The industry has had two serious goes at the problem. The <span class="term" data-term="service-bus">service bus</span> generation put a central integration platform in the middle, with adapters at the edges, a canonical data model in the centre and transformation logic owned by an integration team. <span class="term" data-term="oslc">OSLC</span> took a lighter route through linked data. It identifies resources by URI and describes them with resource shapes, so tools can refer to each other's objects without importing them.

Both beat having nothing, and both put the integration in the middle, where it needs a team of its own. That team owns the mapping between systems it didn't build. A large organisation runs hundreds of them, across domains no single group can hold in its head, so the mapping ends up written by the people furthest from the thing being mapped. Headcount isn't the constraint, because a large organisation can afford the team. The trouble is that the knowledge and the responsibility have ended up in different places.

A small organisation can't staff that middle at all, so the older answers were never affordable below a certain size. Most engineering happens in organisations with fewer than 25 engineers, and they are who this work is for.

## What SysML v2 settles

[SysML](https://en.wikipedia.org/wiki/Systems_Modeling_Language) is the standard modelling language for systems engineering. Its first version was defined as a profile of [UML](https://en.wikipedia.org/wiki/Unified_Modeling_Language), a set of stereotypes layered onto a language built for software design. That bought an installed base on day one, and it imposed limits that never went away. A pump was a stereotyped class with mass and material hung off it as tags. Energy crossing an interface borrowed relationships designed for messages passing between software objects. It worked, but systems engineers had to describe their systems in a vocabulary shaped for somebody else's problems.

Semantics defined by a profile are semantics defined by convention. Two tools could apply the same profile differently and both still be valid, so exchanging models was unreliable in principle. That's part of why the export habit outlived every attempt to standardise around it.

Version 2 is not a profile. It sits on a foundation of its own called [KerML](https://www.omg.org/spec/KerML/), with semantics defined formally at the base rather than borrowed from somewhere they were never meant to reach. The [OMG](https://en.wikipedia.org/wiki/Object_Management_Group) announced its adoption on 21 July 2025 ([press release](https://www.omg.org/news/releases/pr2025/07-21-25.htm)), and the language specification is published as formal/26-03-02, which is [the version this work targets](../decisions/AD-0019-sysml-2-0-target.md).

Cutting the UML dependency is what makes everything after it possible. Two kinds of lock-in go, and they're worth keeping apart because people usually run them together.

The first is authoring. The language has a textual notation, so a model is a set of text files. They live in [Git](https://en.wikipedia.org/wiki/Git), show their differences in a pull request and can be written in any editor. Your choice of tool becomes a preference rather than a commitment. This is the change everyone talks about, and it's the one that puts a systems model into the same review workflow as the software.

The second is access, and for this series it matters more. SysML v2 comes with a standard API and a standard JSON serialisation. Any conforming repository exposes projects, commits and elements the same way, and any conforming tool can read the same serialised model. For the first time the model has a defined interface that isn't a vendor's.

## What it leaves open

Neither of those fixes the shape of what comes back. The API hands you the <span class="term" data-term="metamodel">metamodel</span>. Elements are generically typed, relationships are themselves elements, and you reach a requirement's text by walking owned relationships to documentation comments. That's correct, complete and the right design for a modelling API. It also means every consumer of the model has to learn the metamodel before it can do anything useful.

So the compliance dashboard team learns SysML. The reliability tool learns SysML. The people building the customer-facing status page learn SysML badly, and write the same traversal again with a different set of bugs. In effect, the ontology gets pushed out onto people who never asked for it. That's what happened with OSLC resource shapes, and it's why lifecycle integration stayed the preserve of organisations that could staff it.

## A projection instead

Nobody outside systems engineering should need to know what SysML is.

My proposal is to let the model's owners publish a <span class="term" data-term="projection">projection</span>. Systems engineers understand the metamodel, and they're the right people to decide that, for everyone downstream, a requirement is an identifier, a name, some text, a limit and a set of things that satisfy it. Everything else stays inside the model. A consumer meets a small, plainly typed object and never runs into KerML, usages or owned relationships.

Requirements are first-class objects in a SysML model rather than a separate document. A requirement can be satisfied by a part, derived from another requirement and verified by a test case, and those are typed relationships you can query and check. The projection publishes whichever of them a consumer needs, and leaves the machinery underneath where it belongs.

Three conditions make such a projection worth having:

- It has to be **live** rather than exported, so it can't drift.
- It has to be **joinable**. A service holding data the model doesn't contain should be able to attach that data to the model's objects without either side knowing about the other.
- The contract between producer and consumer has to be **checked mechanically** before deployment, instead of being discovered in production.

Separately from those three, a projection is a read path. Writes belong to the model's own API, and a projection that accepts them has stopped being one.

Federation meets all three conditions. The idea grew up inside GraphQL and still speaks its vocabulary, which is why this series uses GraphQL. I chose it for the maturity of the tooling, and I have no attachment to the protocol. What federation supplies is two things, combining schemas that different teams own and planning queries across them, and both can be separated from whatever carries the bytes. The claim underneath this whole series is that federation is the missing integration layer for open MBSE. I've recorded the choice of [federation over a single service](../decisions/AD-0001-federation-over-single-service.md), with the alternatives that lost.

## Federation, for a systems engineer

[GraphQL](https://en.wikipedia.org/wiki/GraphQL) is a style of API where the server publishes a schema describing its types and their fields, and a client sends a query naming exactly the fields it wants. There is one endpoint, and the response has the same shape as the request.

Federation stretches that across services. Several independently owned services each publish a fragment of a schema, called a subgraph. A build step merges the fragments into one schema. At run time a component called the router accepts a query written against the merged schema, works out which service holds which field, calls each one and assembles the result. The client sees one coherent graph and has no idea how many services stand behind it.

> [!NOTE]
> **Entity key**
>
> The field that tells federation two services are talking about the same object. One service declares that a type is identified by a particular field. Any other service can then add fields to that same type by declaring the same key. When a client asks for fields from both, the router fetches from each and joins the answers on the key.

The entity key is what makes the merge work. In this demo the adapter says a requirement has an identifier and some text (both snippets are trimmed to the fields that matter here):

```graphql
type Requirement @key(fields: "id") {
  id: ID!
  text: String
}
```

and the capacity service, which runs the analysis, separately says a requirement has a verdict, its own answer on whether that requirement is met:

```graphql
type Requirement @key(fields: "id") {
  id: ID!
  verdict: Verdict!
}
```

Neither service imports the other, and neither calls the other. A client asks for a requirement's text and its verdict in one query, and the router fetches from both and joins them on the key. To a systems engineer, that's a cross-tool trace link that resolves by itself and gets checked before it ships. What the key should be for an element of a SysML model is a question of its own, and I answered it with [the short name, with the qualified name as fallback](../decisions/AD-0018-short-names-as-keys.md).

[![The adapter and the capacity service each declare Requirement with the key id. The adapter adds text and the capacity service adds verdict. A client sends one query for the text and verdict of PIPE-R1 to the router, which fetches from both services and returns one answer joined on the id.](../figures/entity-key-join.svg)](../figures/entity-key-join.svg)

*One query, two services that never call each other, one answer joined on the key.*

Two properties follow, and they matter more than the syntax.

**Nobody writes the merged schema.** It is computed, so nobody owns it, which is the opposite of a canonical data model negotiated by committee. That's why the approach needs no central integration team. The people who build each service, and understand its domain, write its schema. So the integration sits where the knowledge already is, rather than in a layer that has to pick it up second hand. If two services define incompatible things, the merge fails in the pipeline of whoever pushed the change, with a message naming the conflict.

**The router holds no logic.** It transforms nothing, orchestrates nothing and enforces no rules of its own. It resolves a typed graph and stops there. Everything the service bus generation put in the middle stays in the services, which keeps the middle from becoming a bottleneck.

## A model of models

Once that arrangement is in place, the word model starts to mean something bigger.

So far "the model" has meant the SysML model. The merged schema is a different object. It can carry requirements from a systems model, verification results from a test management system, part costs from PLM and field behaviour from an operational data store. Each is a model of some aspect of the same thing, held in whatever formalism its own discipline settled on. Some will be SysML. Most won't, and there's no reason they should be. A test management system already models tests properly, and making it speak SysML would be a step backwards.

What the merged schema describes is a [system of systems](https://www.sebokwiki.org/wiki/Systems_of_Systems_(SoS)). In systems engineering that means an arrangement whose constituent systems are useful in their own right, are owned and run by different parties, and together do something none of them does alone. That describes the services behind a federated graph without adjustment, and it describes the models they carry just as well.

The engineering benefit is a kind of [separation of concerns](https://en.wikipedia.org/wiki/Separation_of_concerns) that MBSE has found hard to achieve. Each team models what it understands, in the notation that suits it, at whatever pace its work runs. Nothing needs importing, translating or keeping in step by hand. Adding a twelfth model doesn't mean renegotiating a schema with the other eleven, because keys are the only thing anyone has to agree on. When a model outgrows itself it gets split, and nothing downstream notices.

The alternative, which the industry keeps trying, is one very large model covering everything. Those are hard to build and harder to maintain, and they concentrate authority in whoever owns the schema. A federated model of models grows by adding constituents rather than by growing a monolith, which is the same reason we prefer that architecture for the systems themselves.

## Why Cosmo, and not plain GraphQL

A plain GraphQL server publishes one schema owned by one team. To get the projection described above, you would put the model, the analysis and the document structure behind a single service, and you'd be back to a central integration component that every team has to change together. Federation exists to avoid exactly that, and it's a platform problem rather than a library problem. You need composition, checks for breaking changes, a registry that knows what's deployed, and a planner in front.

Of the platforms that do this, I chose [Cosmo](https://github.com/wundergraph/cosmo), mostly for reasons that have little to do with GraphQL. The record for [Cosmo as the platform](../decisions/AD-0002-cosmo-as-platform.md) sets them against the alternatives.

It's under the [Apache 2.0](https://en.wikipedia.org/wiki/Apache_License) licence, where the main alternative uses a licence the [OSI](https://en.wikipedia.org/wiki/Open_Source_Initiative) doesn't recognise as open source. For an argument aimed at small organisations, that isn't a footnote. Composition runs locally with no connection to any <span class="term" data-term="control-plane">control plane</span>, and the router can start from a pre-built configuration file.

As the vendor ships it, the router still sends anonymous usage data. So the image bakes in the two variables that turn the usage tracker off, `DO_NOT_TRACK=1` and `COSMO_TELEMETRY_DISABLED=true`, alongside `TRACING_ENABLED=false` and `METRICS_OTLP_ENABLED=false` for the tracing and metrics exporters. With [telemetry off](../decisions/AD-0013-telemetry-off.md) and no graph token set, nothing the router does reaches outside the container on any code path I've read. I ran the image under `docker run --network none` to show that rather than just claim it. Defence, rail, energy and medical device work all need that, and those are the industries most likely to be modelling in SysML in the first place.

<details markdown="1">
<summary>Under the bonnet: the fifth variable</summary>

A fifth variable, `PROMETHEUS_ENABLED=false`, closes the scrape endpoint the router would otherwise open on `127.0.0.1:8088`. An endpoint like that waits to be read and opens nothing outbound, so the air-gap claim never rested on it. It's off because nothing here reads it.

</details>

The vendor is candid that the static path isn't the one it recommends. The composition page says "it is recommended to not use this for production", and the router logs "Not recommended for Production" when it starts from a file. I take the warning at face value and answer it on the demo's own terms. There's no control plane to fetch from, the composed configuration lives in [a committed file](../decisions/AD-0012-composition-committed.md), and a test fails when that file drifts from the schemas it was built from. That test is the third condition made concrete, and of the three it's the one you can check by running something.

The deeper reason is that Cosmo itself is moving away from GraphQL as the transport. Its subgraphs can now be compiled to [protobuf](https://en.wikipedia.org/wiki/Protocol_Buffers) and served over [gRPC](https://en.wikipedia.org/wiki/GRPC), with GraphQL kept as the schema language and the protocol at the edge. I read that as an admission that the valuable part was never the wire format. It's the composition algebra (the rules by which schemas merge), the entity key and the query planner.

That's my position too. I'm agnostic about protocols. What I want is a layer that lets several independently owned services contribute to one coherent view of a system, with the contract checked at build time. Cosmo is the most credible open implementation of that layer I've found, and its direction of travel suggests it will still be one when the wire format changes again.

## What the demo is meant to show

The demo publishes a SysML v2 model of a query processing pipeline as a federated service. A throughput analysis and a requirements document join it, and neither knows anything about SysML.

The pipeline is five servers wired as stages in series and in parallel, each with a throughput. Where stages follow one another, the capacity is the smallest of them. Where a stage is split across parallel servers, it is their sum. The capacity service computes this as [a maximum flow with the source-side minimum cut](../decisions/AD-0007-rollup-as-maximum-flow.md), which means the smallest total throughput of any set of servers that cuts every path through the wiring. A model requirement on the whole pipeline states the query rate it must sustain, and a requirement for each server is derived from it.

Three services stand behind one router, an arrangement [The architecture in one sitting](01-the-architecture-in-one-sitting.md) takes apart:

- the **adapter** serves the model
- the **capacity service** computes the rollup, meaning the pipeline-wide capacity that follows from the servers, and returns a verdict
- the **document service** holds [the document's structure](../decisions/AD-0025-document-owns-its-structure.md) and nothing else

A requirements document, here, is a live view over the model plus the editorial choices about ordering, numbering and what to include, which the model doesn't hold. Two small web apps sit in front, a model viewer and a requirements document, and you can make an edit in either.

Change one server's throughput and the requirements document responds at once. The rolled-up capacity moves, the requirement passes or fails, and where it fails, the reason names the server that limits it. Nothing is exported, and nobody reruns an analysis to reissue the document.

What *doesn't* happen is the instructive part. Raise the throughput of a server that isn't the bottleneck and nothing moves, because a serial chain is governed by its worst link. Raise the bottleneck and the capacity rises, but the requirement still fails, because the bottleneck has moved to the next weakest stage in the wiring. Raise one of the servers there and the requirement passes. With the shipped values the pipeline sustains 1200 queries per second against a limit of 1500, and the bottleneck sits at the parse stage. It's obvious once you've seen it and reliably surprising before. [Twelve use cases and one moving bottleneck](04-twelve-use-cases-and-one-moving-bottleneck.md) walks through it.

None of the three services knows about the other two. The analysis and the document service are written against what the adapter publishes, and an organisation adopting this would put its own in their place. The verdict you see against a requirement comes from a service that has never parsed a model file, sitting beside text from a service that has never computed anything.

## The demo described in its own language

I also keep a SysML v2 model of the demo itself, [the model](https://github.com/Roarge/sysml-federation/tree/main/model) under `model/`. It's a different thing from the pipeline model the demo serves. The pipeline model describes five servers and their requirements, and the adapter reads it at startup. This one describes the adapter, the two services beside it, the router, the two web apps and the image they ship in. It holds:

- the stakeholders and their concerns
- the twelve stories of the storyboard, plus seven more that came with the model
- the requirements, restated as system stories with their statements kept
- the design constraints
- the architecture with its interfaces
- the tests as verification actions, so that a Go test, a recorded demonstration, a validator run, a make target, a workflow and a live check are each an action of a case
- views that name the published boards, one view per board, each exposing what its board draws

Both reference tools, the OMG pilot implementation and OpenSysML, accept it on every change: locally through one make target, and in continuous integration on every pull request that touches it. A unit test keeps it in step with the documents and the tests. It fails when the model and the repository disagree on an identifier, a decision record, a test name, a check file, a published image, a check's inventory or a compose service.

I'd argued for a model at the centre of an organisation's engineering while describing my own demo in prose that nothing checked. I should have known better. [A model of the demo itself](12-a-model-of-the-demo-itself.md) owns up to that, and describes what the model holds and what it leaves out.

The testing around the demo comes in four layers:

1. **Unit tests** run under the <span class="term" data-term="race-detector">race detector</span> on every pull request.
2. **A demonstration record** in the example's README carries the runs against the container that no unit test can make, each dated and each named by the requirements it bears on.
3. **Two validators** read both models, the example's before any parser test uses it and the demo's own on every change to it.
4. **[The check session](https://github.com/Roarge/sysml-federation/tree/main/checkly)** runs every story against your own live instance. It needs a <span class="term" data-term="checkly">Checkly</span> account, and mine can't ship inside a public image, so you bring your own. You then watch a browser on Checkly's runners exercise every story through a tunnel. Every request the router sees is traced in a viewer beside the demo, and with a tracing key, beside the check result that caused it. Without an account none of that runs, and the demo is one `docker run` as before.

## Stand-ins, and what replaces them

Three parts of the demo are stand-ins.

Two belong to the adapter. There's no model repository behind it, only [a directory of files](../decisions/AD-0003-adapter-reads-files.md), so versioning is a counter rather than the commit history a conforming SysML v2 repository would give. Editing the model through the projection is [scaffolding](../decisions/AD-0004-editing-as-scaffolding.md) too. It contradicts the position I took above, that a projection is a read path, and a real deployment would write through the SysML v2 API instead.

The third belongs to the example. Its [idealised capacity model](../decisions/AD-0006-idealised-capacity-model.md) assumes work splits evenly across parallel branches, load balancing is perfect and nothing queues anywhere. It's arithmetic chosen to make a point about federation, and nobody should plan capacity with it.

Replacing the adapter's two stand-ins means a fuller mapping from the model to the graph. SysML v2 has <span class="term" data-term="view">views</span> and viewpoints of its own, and I intend to use them as the way a systems engineer chooses which parts of a model to federate and how those parts look to the services outside. A viewpoint names a concern and the stakeholder who holds it, and a view is the part of the model rendered for that concern. A subgraph, seen from its consumer, is the same thing. It's the slice of the model one downstream party needs, shaped for that party's vocabulary. If the model already says which stakeholder sees what, the projection should be read from it rather than written out a second time in a schema.

That will almost certainly need a parser for the whole language rather than [the strict subset the adapter reads today](../decisions/AD-0015-hand-written-subset-parser.md), and it's the largest piece of work on the horizon. The goal is coverage of the language, so that any conforming model can be served without the adapter knowing what it's modelling. What exists today is a fraction of that. It covers parts with their attributes, ports and connections, requirements with the satisfy and derive relationships between them, and the verification cases that reference them. Nothing in the adapter names the example.

Full coverage pulls against the projection, though, towards a few hundred types generated mechanically from KerML. That lands the consumer right back in the abstract syntax the projection was meant to spare them. My current view is that coverage should grow as [curated projections](../decisions/AD-0005-curated-generic-projection.md), one concern at a time, with a generic escape hatch for elements nobody has projected yet. I'm not certain of that, and it's the part of the design most likely to change.

None of this is a product, and it isn't a SysML v2 API implementation. The adapter reads files rather than fronting a repository, and its coverage of the language is a fraction of what the goal needs. You'll find the adapter and the example in [the repository](https://github.com/Roarge/sysml-federation), and both are small enough to read in an afternoon. Part 2 opens the box.

---

Index: [Federating a systems model](../README.md) · Next: [The architecture in one sitting](01-the-architecture-in-one-sitting.md)

# Why federate a systems model

*Roar Georgsen, 27 August 2026*

Part 1 of 11 in [Federating a systems model](../README.md).

## The documents that describe one system

Every engineering organisation I have worked with keeps its system in several places at once. Requirements in one file, interface definitions in another, the power budget in a spreadsheet, the hazard analysis somewhere a third team controls. All of them describe the same machine, and keeping them in agreement is manual work that nobody enjoys and everybody defers until a review forces it.

[Model based systems engineering](https://www.sebokwiki.org/wiki/Model-Based_Systems_Engineering_(MBSE)) was supposed to end that. The documents give way to a model, one structured description holding the system's parts, the properties they carry, the requirements they have to meet and the relationships between all of it. Documents become views onto the model rather than the place information lives. Consistency stops being a review activity. Change a component's power draw and everything downstream can be recomputed, including which requirements now fail and which tests need rerunning.

That promise has been five years from mainstream adoption for about 25 years. Explanations for the delay differ, and the least disputed of them is tooling.

Models live inside proprietary environments. They are stored in binary formats, or in a vendor flavour of [XMI](https://en.wikipedia.org/wiki/XML_Metadata_Interchange) that only the vendor round-trips reliably, reachable through an API that is absent or shaped differently in every product. A model that only its author tool can read is a document with extra ceremony, and it gets treated accordingly.

Version control is worse. Several of the major environments still assume [Subversion](https://en.wikipedia.org/wiki/Apache_Subversion) or [TFS](https://en.wikipedia.org/wiki/Team_Foundation_Server), treat the model as a lockable binary rather than something mergeable, and offer nothing a build server can drive. Branching becomes an administrative operation. [Continuous integration](https://en.wikipedia.org/wiki/Continuous_integration) is not a concept the tool has heard of, so the model cannot join the engineering workflow every other discipline now takes for granted.

Vendors will tell you integration is a solved problem, and they are not exactly lying. Connectors to requirements tools and test management exist, and so do [PLM](https://en.wikipedia.org/wiki/Product_lifecycle_management) bridges. They are also, with tiresome regularity, a separately licensed product on top of the one you already bought. What arrives is often a bridge written against a Java desktop stack that has aged badly, configured through a dialogue box rather than a file, and liable to break on the next upgrade of either end.

Two-way synchronisation exists as well, and where somebody has set it up properly it works. Standing it up is a different skill from the one the engineer who needs it has, so more often than not it stays unconfigured and the default settles back to export.

That is the practical shape of lock-in. It is rarely a clause in a contract. It is an accumulation of daily friction that makes export always the cheapest option in the moment. Requirements go to Word for review. Interface definitions reach the software team through a spreadsheet, exported where the tool allows it and retyped where it does not. The power budget lives in Excel because that is where the person who owns it works.

Every one of those copies starts drifting the moment it is made.

## The promise has been made before

The industry has attacked this twice. The [service bus](https://en.wikipedia.org/wiki/Enterprise_service_bus) generation put a central integration platform in the middle, with adapters at the edges, a canonical data model in the centre and transformation logic owned by an integration team. [OSLC](https://en.wikipedia.org/wiki/Open_Services_for_Lifecycle_Collaboration) took a lighter route through linked data, identifying resources by URI and describing them with resource shapes, so that tools could reference each other's objects without importing them.

Both were an improvement on having nothing, and both put the integration in the middle, where it needs a team of its own. That team owns the mapping between systems it did not build. A large organisation runs hundreds of them, across domains no single group can hold in its head, so the mapping ends up written by the people furthest from the thing being mapped. Headcount is not the constraint, and a large organisation can afford the team. The constraint is that the knowledge and the responsibility have been put in different places.

A small organisation cannot staff that middle at all, so the older answers were never affordable below a certain size. Most engineering happens in organisations with fewer than 25 engineers, and they are who this work is for.

## What SysML v2 settles

[SysML](https://en.wikipedia.org/wiki/Systems_Modeling_Language) is the standard modelling language for systems engineering, and its first version was defined as a profile of [UML](https://en.wikipedia.org/wiki/Unified_Modeling_Language), a set of stereotypes layered onto a language built for software design. That bought an installed base on day one and imposed limits that never went away. A pump was a stereotyped class with mass and material hung off it as tags. Energy crossing an interface borrowed relationships designed for messages passing between software objects. It worked, but systems engineers had to describe their systems in a vocabulary shaped for somebody else's problems.

Semantics defined by a profile are semantics defined by convention. Two tools could apply the same profile differently and both remain valid, so interchange was unreliable in principle, which is part of why the export habit outlasted every attempt to standardise around it.

Version 2 is not a profile. It sits on a foundation of its own called [KerML](https://www.omg.org/spec/KerML/), with semantics defined formally at the base rather than inherited from a place they were never meant to reach. The [OMG](https://en.wikipedia.org/wiki/Object_Management_Group) announced its adoption on 21 July 2025 ([press release](https://www.omg.org/news/releases/pr2025/07-21-25.htm)) and the language specification is published as formal/26-03-02, which is [the version this work targets](../decisions/AD-0019-sysml-2-0-target.md).

Cutting the UML dependency is what makes everything after it possible. Two lock-ins go, and they are worth keeping apart because they are usually run together.

The first is authoring. The language has a textual notation, so a model is a set of text files that live in [Git](https://en.wikipedia.org/wiki/Git), diff in a pull request and can be written in any editor. Tool choice becomes a preference rather than a commitment. This is the change everyone talks about, and it is the one that puts a systems model into the same review workflow as the software.

The second is access, and it matters more here. SysML v2 comes with a standard API and a standard JSON serialisation. Any conforming repository exposes projects, commits and elements the same way, and any conforming tool can read the same serialised model. For the first time the model has a defined interface that is not a vendor's.

## What it leaves open

Neither of those fixes the shape of what comes back. The API hands you the metamodel. Elements are generically typed, relationships are themselves elements, and a requirement's text is reached by walking owned relationships to documentation comments. That is correct, complete and the right design for a modelling API. It also means every consumer of the model has to learn the metamodel before it can do anything useful.

So the compliance dashboard team learns SysML. The reliability tool learns SysML. The people building the customer-facing status page learn SysML badly and reimplement the same traversal with a different set of bugs. The ontology gets pushed outward onto people who never asked for it, which is what happened with OSLC resource shapes, and it is why lifecycle integration stayed the preserve of organisations that could staff it.

## A projection instead

Nobody outside systems engineering should need to know what SysML is.

Let the model's owners publish a projection. Systems engineers understand the metamodel and are the right people to decide that, for the purposes of everyone downstream, a requirement is an identifier, a name, some text, a limit value and a set of things that satisfy it. Everything else stays inside the model. A consumer meets a small, plainly typed object and never encounters KerML, usages or owned relationships.

Requirements are first-class objects in a SysML model rather than a separate document. A requirement can be satisfied by a part, derived from another requirement and verified by a test case, and those are typed relationships you can query and check. The projection publishes whichever of them a consumer needs, and leaves the machinery underneath where it belongs.

Three conditions make such a projection worth anything. It has to be live rather than exported, so that it cannot drift. It has to be joinable, so that a service holding data the model does not contain can attach that data to the model's objects without either side knowing about the other. And the contract between producer and consumer has to be checked mechanically, before deployment, rather than discovered in production. A projection is a read path. Writes belong to the model's own API, and a projection that accepts them has stopped being one.

Federation does all three. The concept grew up inside GraphQL and still speaks its vocabulary, which is why this series uses GraphQL, though the reason is the maturity of the tooling rather than any attachment to a protocol. What the approach supplies is the composition of independently owned schemas and the planning of queries across them, and both of those are separable from whatever carries the bytes. The claim underneath this series is that federation is the missing integration layer for open MBSE. The choice of [federation over a single service](../decisions/AD-0001-federation-over-single-service.md) is recorded with the alternatives that lost.

## Federation, for a systems engineer

[GraphQL](https://en.wikipedia.org/wiki/GraphQL) is an interface style where the server publishes a schema describing types and their fields, and a client sends a query naming the specific fields it wants. One endpoint, and the response mirrors the shape of the request.

Federation extends that across services. Several independently owned services each publish a fragment of a schema, called a subgraph. A build step merges the fragments into one unified schema. At runtime a component called the router accepts a query written against the merged schema, works out which service holds which field, calls each one and assembles the result. The client sees a single coherent graph and has no idea how many services stand behind it.

The mechanism that makes the merge work is the entity key. A service declares that a type is identified by a particular field, and any other service may then contribute fields to that same type by declaring the same key. The adapter says a requirement has an identifier and some text:

```graphql
type Requirement @key(fields: "id") {
  id: ID!
  text: String!
}
```

and the analysis service, separately, says a requirement has a verdict, its own answer on whether that requirement is met:

```graphql
type Requirement @key(fields: "id") {
  id: ID!
  verdict: Verdict!
}
```

Neither service imports the other. Neither calls the other. A client asks for a requirement's text and its verdict in one query, and the router fetches from both and merges on the key. The systems engineering reading of that is a cross-tool trace link that resolves automatically and is validated before it ships. What the key should be for an element of a SysML model is a question of its own, answered here by [the short name, with the qualified name as fallback](../decisions/AD-0018-short-names-as-keys.md).

Two properties follow that matter more than the syntax.

The merged schema is computed rather than authored. Nobody writes it and nobody owns it, which is the opposite of a canonical data model negotiated by committee, and it is why the approach needs no central integration team. Each service's own schema is written by the people who build that service and understand its domain, so the integration sits where the knowledge already is rather than in a layer that has to acquire it second hand. If two services define incompatible things the merge fails, in the pipeline of whoever pushed the change, with a message naming the conflict.

The router holds no logic. It transforms nothing, orchestrates nothing and enforces no rule of its own, and it resolves a typed graph and stops there. Everything the service bus generation put in the middle stays in the services here, which is what keeps the middle from becoming a bottleneck.

## A model of models

Something shifts once that arrangement is in place, and it changes what the word model refers to.

So far the model has been the SysML model. The merged schema is a different object. It can carry requirements from a systems model, verification results from a test management system, part costs from PLM and field behaviour from an operational data store, each of them a model of some aspect of the same thing, held in whatever formalism its own discipline settled on. Some will be SysML. Most will not, and there is no reason they should be. A test management system already models tests properly, and making it speak SysML would be a step backwards.

What the merged schema describes is a [system of systems](https://www.sebokwiki.org/wiki/Systems_of_Systems_(SoS)), which in systems engineering means an arrangement whose constituents are useful in their own right, are owned and run by different parties, and together do something none of them does alone. That describes the services behind a federated graph without adjustment, and it describes the models they carry just as well.

The engineering benefit is [separation of concerns](https://en.wikipedia.org/wiki/Separation_of_concerns) of a kind MBSE has found hard to achieve. Each team models what it understands, in the notation suited to it, at whatever cadence its work runs to. Nothing needs importing, translating or keeping in step by hand. Adding a twelfth model does not mean renegotiating a schema with the other eleven, because the only thing anyone has to agree on is keys. When a model outgrows itself it gets split, and the split is invisible to everything downstream.

The alternative, which the industry keeps trying, is one very large model covering everything. Those are hard to build, harder to maintain, and they concentrate authority in whoever owns the schema. A federated model of models grows by adding constituents rather than by growing a monolith, which is the same reason the architecture is preferred for the systems themselves.

## Why Cosmo, and not plain GraphQL

A plain GraphQL server publishes one schema owned by one team. To get the projection described above you would put the model, the analysis and the document structure behind a single service, and you would be back to a central integration component every team has to change together. Federation exists to avoid that, and it is a platform problem rather than a library problem. You need composition, breaking-change checks, a registry that knows what is deployed and a planner in front.

Among the platforms that do this, [Cosmo](https://github.com/wundergraph/cosmo) is chosen here for reasons mostly not about GraphQL, set out against its alternatives in the record for [Cosmo as the platform](../decisions/AD-0002-cosmo-as-platform.md).

It is [Apache 2.0](https://en.wikipedia.org/wiki/Apache_License), where the main alternative is under a licence the [OSI](https://en.wikipedia.org/wiki/Open_Source_Initiative) does not recognise as open source. For an argument aimed at small organisations, that is not a footnote. Composition runs locally with no connection to any control plane, and the router can be started from a pre-built configuration file. As the vendor ships it the router still sends anonymous usage data, so the image bakes in the two variables that disable the usage tracker, `DO_NOT_TRACK=1` and `COSMO_TELEMETRY_DISABLED=true`, alongside `TRACING_ENABLED=false` and `METRICS_OTLP_ENABLED=false` for the tracing and metrics exporters. With [telemetry off](../decisions/AD-0013-telemetry-off.md) and no graph token set, nothing the router does reaches outside the container on any code path that has been read, and the image phase demonstrates that under `docker run --network none` rather than asserting it. Defence, rail, energy and medical device work all need that, and those are the industries most likely to be modelling in SysML in the first place.

The vendor is candid that the static path is not the one it recommends. The composition page says "it is recommended to not use this for production", and the router logs "Not recommended for Production" when it starts from a file. The demo takes the warning at face value and answers it in its own terms. There is no control plane to fetch from, the composed configuration lives in [a committed file](../decisions/AD-0012-composition-committed.md), and a test fails when it drifts from the schemas it was built from. That test is the third condition made concrete, and it is the part of the three a reader can check by running something.

The deeper reason is that Cosmo is itself moving away from GraphQL as the transport. Its subgraphs can now be compiled to [protobuf](https://en.wikipedia.org/wiki/Protocol_Buffers) and served over [gRPC](https://en.wikipedia.org/wiki/GRPC), with GraphQL kept as the schema language and the edge protocol. I read that as an admission that the valuable part was never the wire format. It is the composition algebra, the entity key and the query planner.

Which is the position I take as well. I am agnostic about protocols. What I want is a layer that lets several independently owned services contribute to one coherent view of a system, with the contract checked at build time. Cosmo is the most credible open implementation of that layer I have found, and its direction of travel suggests it will still be one when the wire format changes again.

## What the example is meant to show

The demo publishes a SysML v2 model of a query processing pipeline as a federated service, joined to a throughput analysis and a requirements document that know nothing about SysML. The pipeline is five servers wired as stages in series and in parallel, each carrying a throughput. Where stages follow one another the capacity is the smallest of them, and where a stage is split across parallel servers it is their sum, computed as [a maximum flow with the source-side minimum cut](../decisions/AD-0007-rollup-as-maximum-flow.md), the smallest total throughput of any set of servers that cuts every path through the wiring. A model requirement on the whole pipeline states the query rate it must sustain, and a requirement for each server is derived from it.

Three services stand behind one router, an arrangement [The architecture in one sitting](01-the-architecture-in-one-sitting.md) takes apart. One serves the model, one computes the rollup, meaning the pipeline-wide capacity that follows from the servers, and returns a verdict, and one holds [the document's structure](../decisions/AD-0025-document-owns-its-structure.md) and nothing else. A requirements document, here, is a live view over the model plus the editorial decisions about ordering, numbering and what to include that the model does not contain. Two small web apps sit in front, a model viewer and a requirements document, and an edit can be made in either.

![Two apps, one graph, three services that never meet](../img/overview-sketch.png)

*Two apps, one graph, three services that never meet. Cut from the [use case storyboard](../stories/use-cases.pdf).*

Change one server's throughput and the requirements document responds at once. The rolled-up capacity moves, the requirement passes or fails, and where it fails the reason names the server that limits it. Nothing is exported and nobody reruns an analysis to reissue the document.

What does not happen is the instructive part. Raise the throughput of a server that is not the bottleneck and nothing moves, because a serial chain is governed by its worst link. Raise the bottleneck and the capacity rises, but the requirement still fails, because the bottleneck has moved to the next weakest stage in the wiring. Raise one of the servers there and the requirement passes. With the shipped values the pipeline sustains 1200 queries per second against a limit of 1500, and the bottleneck sits at the parse stage. That behaviour is obvious once seen and reliably surprising before. [Twelve use cases and one moving bottleneck](04-twelve-use-cases-and-one-moving-bottleneck.md) walks through it.

None of the three services knows about the other two. The analysis and the document service are written against what the adapter publishes, and an organisation adopting this would put its own in their place. The verdict a reader sees against a requirement comes from a service that has never parsed a model file, sitting beside text from a service that has never computed anything.

## Placeholders, and what replaces them

Three parts of the demo are stand-ins.

Two belong to the adapter. There is no model repository behind it, only [a directory of files](../decisions/AD-0003-adapter-reads-files.md), so versioning is a counter rather than the commit history a conforming SysML v2 repository would give. Editing the model through the projection is [scaffolding](../decisions/AD-0004-editing-as-scaffolding.md) as well. It contradicts the position taken above, that a projection is a read path, and a real deployment would write through the SysML v2 API instead.

The third belongs to the example. Its [idealised capacity model](../decisions/AD-0006-idealised-capacity-model.md) assumes evenly partitionable work across parallel branches, perfect load balancing and no queueing anywhere. It is arithmetic chosen to make a point about federation, and nobody should plan capacity with it.

Replacing the adapter's two stand-ins means a fuller mapping from the model to the graph. SysML v2 has views and viewpoints of its own, and my intention is to use them as the way a systems engineer chooses which parts of a model to federate and how those parts appear to the services outside. A viewpoint names a concern and the stakeholder who holds it, and a view is the part of the model rendered for that concern. A subgraph, seen from its consumer, is the same thing, the slice of the model one downstream party needs, shaped for that party's vocabulary. If the model already says which stakeholder sees what, the projection should be read from it rather than authored a second time in a schema.

That will almost certainly need a parser for the whole language rather than [the strict subset the adapter reads today](../decisions/AD-0015-hand-written-subset-parser.md), and it is the largest piece of work on the horizon. The goal is coverage of the language, so that any conforming model can be served without the adapter knowing what it is modelling. What exists today is a fraction of that, parts with their attributes, ports and connections, requirements with the satisfy and derive relationships between them, and the verification cases that reference them. Nothing in the adapter names the example. Full coverage of the language pulls against the projection, towards a few hundred types generated mechanically from KerML, which lands the consumer back in the abstract syntax the projection was meant to spare them. My current view is that coverage should grow as [curated projections](../decisions/AD-0005-curated-generic-projection.md), one concern at a time, with a generic escape hatch for elements nobody has projected yet. I am not certain of that, and it is the part of the design most likely to change.

It is not a product, and not a SysML v2 API implementation. The adapter reads files rather than fronting a repository, and its coverage of the language is a fraction of what the goal requires. The adapter and the example are in [the repository](https://github.com/Roarge/sysml-federation), and both are small enough to read in an afternoon.

---

Index: [Federating a systems model](../README.md) · Next: [The architecture in one sitting](01-the-architecture-in-one-sitting.md)

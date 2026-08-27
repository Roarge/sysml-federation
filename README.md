# Federating a systems model

An adapter that publishes a SysML v2 model as a federated service, and a worked example joining it to a throughput analysis and a requirements document that know nothing about SysML.

## The integration problem

Engineering organisations have always described their systems in documents. A requirements specification in one file, interface definitions in another, a power budget in a spreadsheet, a hazard analysis somewhere a third team owns. Every one of those describes the same system, and keeping them consistent with each other is manual work that nobody enjoys and everybody defers.

[Model based systems engineering](https://www.sebokwiki.org/wiki/Model-Based_Systems_Engineering_(MBSE)) replaces the documents with a model. One structured description holding the system's parts, the properties they carry, the requirements they have to meet and the relationships between all of it. Documents become views onto the model rather than the place information lives. The promise is that consistency stops being a review activity. Change a component's power draw and everything downstream can be recomputed, including which requirements now fail and which tests need rerunning.

That promise has been five years from mainstream adoption for about 25 years. There are several explanations. The least disputed is tooling.

Models live inside proprietary environments. They are stored in binary formats, or in a vendor flavour of [XMI](https://en.wikipedia.org/wiki/XML_Metadata_Interchange) that only the vendor round-trips reliably, reachable through an API that is absent or shaped differently in every product. A model that only its author tool can read is a document with extra ceremony, and it gets treated accordingly.

Vendors will tell you integration is a solved problem, and they are not exactly lying. Connectors to requirements tools and test management exist, and so do [PLM](https://en.wikipedia.org/wiki/Product_lifecycle_management) bridges. They are also, with tiresome regularity, a separately licensed product on top of the one you already bought. What arrives is often a bridge written against a Java desktop stack that has aged badly, configured through a dialogue box rather than a file, and liable to break on the next upgrade of either end.

Version control is worse. Several of the major environments still assume [Subversion](https://en.wikipedia.org/wiki/Apache_Subversion) or [TFS](https://en.wikipedia.org/wiki/Team_Foundation_Server), treat the model as a lockable binary rather than something mergeable, and offer nothing a build server can drive. Branching is an administrative operation. [Continuous integration](https://en.wikipedia.org/wiki/Continuous_integration) is not a concept the tool has heard of, which means the model cannot participate in the engineering workflow that every other discipline now takes for granted.

That is the practical shape of lock-in. It is rarely a clause in a contract. It is an accumulation of daily friction that makes export always the cheapest option in the moment. Requirements go to Word for review. Interface definitions reach the software team through a spreadsheet, exported where the tool allows it and retyped where it does not. The power budget lives in Excel because that is where the person who owns it works.

Some tools do offer two way synchronisation, and where someone has set it up properly it works. Standing it up, however, is a different skill from the one the engineer who needs it has, so more often than not it stays unconfigured and the default settles back to export.

Every one of those copies starts drifting the moment it is made.

The industry has attacked this before. The [service bus](https://en.wikipedia.org/wiki/Enterprise_service_bus) generation put a central integration platform in the middle, with adapters, canonical data models and transformation logic owned by an integration team. [OSLC](https://en.wikipedia.org/wiki/Open_Services_for_Lifecycle_Collaboration) took a linked data approach, identifying resources by URI and describing them with shapes, so tools could reference each other's objects without importing them. Both improved on nothing, and both carried an assumption that turns out to be the binding constraint: someone employed to do integration full time.

Organisations with fewer than 25 engineers do not have that person. They are where most engineering actually happens, and they are the ones for whom the existing answers are unaffordable. That is the audience this repository is aimed at.

## What SysML v2 fixes, and what it leaves open

[SysML](https://en.wikipedia.org/wiki/Systems_Modeling_Language) is the standard modelling language for systems engineering. Its first version was defined as a profile of [UML](https://en.wikipedia.org/wiki/Unified_Modeling_Language), a set of stereotypes layered onto a language built for software design. That bought an installed base on day one but imposed limitations. A pump was a stereotyped class with mass and material hung off it as tags. Energy crossing an interface borrowed relationships that had been designed for messages passing between software objects. It worked, but systems engineers had to describe their systems in a vocabulary shaped for someone else's problems.

Semantics defined by a profile are semantics defined by convention. Two tools could apply the same profile differently and both remain valid, so interchange was unreliable even in principle, which is part of why the export habit described above outlasted every attempt to standardise around it.

Version 2 is not a profile. It sits on a foundation of its own called [KerML](https://www.omg.org/spec/KerML/), with semantics defined formally at the base rather than inherited from a place they were never meant to reach, and the [OMG](https://en.wikipedia.org/wiki/Object_Management_Group) announced its adoption on 21 July 2025 ([press release](https://www.omg.org/news/releases/pr2025/07-21-25.htm)), with the specification published as formal/26-03-02. Cutting the UML dependency is what makes the rest of this possible. It removes two further lock-ins, and they are worth keeping apart because they are usually run together.

The first is authoring. The language has a textual notation, so a model is a set of text files that live in [Git](https://en.wikipedia.org/wiki/Git), diff in a pull request, and can be written in any editor. Tool choice becomes a preference rather than a commitment. This is the change everyone talks about, and it is genuinely large.

The second is access, and it matters more here. SysML v2 comes with a standard API and a standard JSON serialisation. Any conforming repository exposes projects, commits and elements the same way, and any conforming tool can read the same serialised model. For the first time the model has a defined interface that is not a vendor's.

What neither of those fixes is the shape of what comes back. The SysML v2 API hands you the metamodel. Elements are generically typed, relationships are themselves elements, and a requirement's text is reached by walking owned relationships to documentation comments. That is correct and complete, and it is the right design for a modelling API. It also means every consumer of the model has to learn the metamodel before it can do anything useful.

So the compliance dashboard team learns SysML. The reliability tool learns SysML. The people building the customer-facing status page learn SysML, badly, and reimplement the same traversal with a different set of bugs. The ontology gets pushed outward onto people who never asked for it, which is exactly what happened with OSLC resource shapes, and it is why lifecycle integration stayed the preserve of organisations that could staff it.

## Nobody outside systems engineering should need to know what SysML is

The alternative is to let the model's owners publish a projection.

Systems engineers understand the metamodel. They are the right people to decide that, for the purposes of everyone downstream, a requirement is an identifier, a name, some text, a limit value and a set of things that satisfy it. Everything else stays inside the model. A consumer sees a small, plainly typed object and never encounters KerML, usages, or owned relationships.

That projection has to satisfy three conditions to be worth anything. It must be a live view rather than an export, so it cannot drift. It must be joinable, so a service holding data the model does not contain can attach that data to the model's objects without either side knowing about the other. And the contract between producer and consumer must be checked mechanically, before deployment, rather than discovered in production.

Federation does all three. The concept grew up inside GraphQL and still speaks its vocabulary, which is why this repository uses GraphQL, though the reason is the maturity of the tooling rather than any attachment to any protocol. What the approach actually supplies is the composition of independently owned schemas and the planning of queries across them. Both are separable from whatever carries the bytes, and that separation is already under way.

This repository is an argument that federation is the missing integration layer for open MBSE.

## Federation, for the systems engineers

[GraphQL](https://en.wikipedia.org/wiki/GraphQL) is an interface style where the server publishes a schema describing types and their fields, and the client asks for the specific fields it wants. One endpoint, and the response mirrors the shape of the request.

Federation extends that across services. Several independently owned services each publish a fragment of a schema. A build step merges the fragments into one unified schema. At runtime a component called the router accepts a query against the merged schema, works out which service holds which field, calls each one, and assembles the result. The client sees a single coherent graph and has no idea how many services stand behind it.

The mechanism that makes the merge work is the entity key. A service declares that a type is identified by a particular field, and any other service may then contribute fields to that same type by declaring the same key. Sketched at its simplest, the adapter says a requirement has an identifier and some text:

```graphql
type Requirement @key(fields: "id") {
  id: ID!
  text: String!
}
```

and the analysis service, separately, says a requirement has a verdict:

```graphql
type Requirement @key(fields: "id") {
  id: ID!
  verdict: Verdict!
}
```

Neither service imports the other. Neither calls the other. A client asks for a requirement's text and its verdict in one query, and the router fetches from both and merges on the key. The systems engineering reading of that is a cross-tool trace link that resolves automatically and is validated before it ships.

Two properties follow that matter more than the syntax.

The merged schema is computed rather than authored. Nobody writes it, and nobody owns it. That is the opposite of a canonical data model negotiated by committee, and it is why the approach does not need a central integration team. If two services define incompatible things, the merge fails, in the pipeline of whoever pushed the change, with a message naming the conflict.

The router holds no logic. No transformation, no orchestration, no business rules. It resolves a typed graph and nothing else. Everything the service bus generation put in the middle stays in the services here, which is what keeps the middle from becoming a bottleneck.

Something shifts once that arrangement is in place, and it changes what the word model refers to.

So far the model has been the SysML model. The merged schema is a different object. It can carry requirements from a systems model, verification results from a test management system, part costs from PLM, field behaviour from an operational data store, and each of those is a model of some aspect of the same thing, held in whatever formalism its own discipline settled on. Some will be SysML. Most will not, and there is no reason they should be. A test management system already models tests properly, and making it speak SysML would be a step backwards.

What the merged schema describes is a [system of systems](https://www.sebokwiki.org/wiki/Systems_of_Systems_(SoS)), which in systems engineering means an arrangement whose constituents are useful in their own right, are owned and run by different parties, and together do something none of them does alone. That describes the services behind a federated graph without adjustment. It describes the models they carry just as well.

The engineering benefit is [separation of concerns](https://en.wikipedia.org/wiki/Separation_of_concerns) of a kind MBSE has found hard to achieve. Each team models what it understands, in the notation suited to it, at whatever cadence its work runs to. Nothing needs importing, translating, or keeping in step by hand. Adding a twelfth model does not mean renegotiating a schema with the other eleven, because the only thing anyone has to agree on is keys. When a model outgrows itself it gets split, and the split is invisible to everything downstream.

The alternative, which the industry has tried repeatedly, is one very large model that covers everything. Those are hard to build, harder to maintain, and they concentrate authority in whoever owns the schema. A federated model of models scales by adding constituents rather than by growing a monolith, which is the same reason the architecture is preferred for the systems themselves.

## Requirements and relationships

A SysML model describes a system as parts, the properties those parts carry, and the relationships between them. Requirements are first class objects in the model rather than a separate document. A requirement can be satisfied by a part, derived from another requirement, and verified by a test case, and those are typed relationships you can query and check.

A requirements document, in this project, is a live view over the model, plus the editorial decisions about ordering, numbering and what to include that the model does not contain.

## Why Cosmo, and not simply GraphQL

A plain GraphQL server publishes one schema owned by one team. To get the projection described above you would put the model, the analysis and the document structure behind a single service, and you would be back to a central integration component that every team has to change together. Federation exists precisely to avoid that, and federation is a platform problem rather than a library problem. You need composition, breaking-change checks, a registry that knows what is deployed, and a planner in front.

Among the platforms that do this, [Cosmo](https://github.com/wundergraph/cosmo) is chosen here for reasons that are mostly not about GraphQL.

It is [Apache 2.0](https://en.wikipedia.org/wiki/Apache_License), where the main alternative is under a licence the [OSI](https://en.wikipedia.org/wiki/Open_Source_Initiative) does not recognise as open source. For an argument aimed at small organisations, that is not a footnote. Composition runs locally with no connection to any control plane, and the router can be started from a pre-built configuration file. As the vendor ships it the router still sends anonymous usage data, so the image sets `DO_NOT_TRACK=1` and `COSMO_TELEMETRY_DISABLED=true`, and with those set nothing leaves the container, which makes the whole stack air-gappable. Defence, rail, energy and medical device work all need that, and those are the industries most likely to be modelling in SysML in the first place.

The vendor is candid that the static path is not the one it recommends. The composition page says "it is recommended to not use this for production", and the router logs "Not recommended for Production" when it starts from a file. The demo takes the warning at face value and answers it in its own terms. There is no control plane to fetch from, the composed configuration is committed to the repository, and a test fails when it drifts from the schemas it was built from.

The deeper reason is that Cosmo is itself moving away from GraphQL as the transport. Its subgraphs can now be compiled to [protobuf](https://en.wikipedia.org/wiki/Protocol_Buffers) and served over [gRPC](https://en.wikipedia.org/wiki/GRPC), with GraphQL kept only as the schema language and the edge protocol. That is an admission that the valuable part was never the wire format. It is the composition algebra, the entity key, and the query planner.

Which is the position this project takes as well. I am agnostic about protocols. What I want is a layer that lets several independently owned services contribute to one coherent view of a system, with the contract checked at build time. Cosmo is the most credible open implementation of that layer I have found, and its direction of travel suggests it will still be one when the wire format changes again.

## The adapter and the examples

The repository is laid out the usual way. A reusable adapter, and examples that use it. Both start as proof of concept and both will keep evolving.

The adapter reads a SysML v2 model and publishes it as a subgraph, with model elements identified so that anything else in the graph can attach to them. Nothing in it is specific to any one example. The goal is coverage of the language, so that any conforming model can be served without the adapter knowing what it is modelling. What exists today is a fraction of that. Requirements, parts, the properties they carry, and the satisfy and derive relationships between them, which is enough to prove the shape works and nowhere near enough for a real model.

There is an unresolved design question underneath that goal. A projection is valuable because it is small, and the consumer of a requirement should meet five fields rather than the metamodel. Full language coverage pulls the other way, towards generating a few hundred types mechanically from KerML and landing the consumer back in the abstract syntax they were being spared. My current view is that coverage should grow as curated domain projections, one at a time, rather than as a generated whole, and that a generic escape hatch for elements nobody has projected yet is the compromise. I am not certain of that, and it is the part of the design most likely to change.

An example is one scenario chosen to test the adapter. The analysis and document services in the first one are written against what the adapter publishes, and an organisation adopting this would write its own in their place. They exist to show that a service can contribute fields to a model's objects while knowing nothing about SysML, which is the claim the argument rests on.

Pull requests are welcome on both parts. Extending the projection to cover more of the language is the obvious place to start, and an example built on a different kind of model would be worth more than another one built on this one.

## The pipeline example

The model is a query processing pipeline of five servers, wired as stages in series and in parallel. Each server carries a throughput. Where stages follow one another the capacity is the smallest of them, and where a stage is split across parallel servers it is their sum. A global requirement states the query rate the pipeline must sustain, and a requirement for each server is derived from it.

Change one server's throughput and the requirements document responds at once. The rolled-up capacity moves, the requirement passes or fails, and a failing requirement is marked as failing with the server responsible named. No export, no regeneration step, no waiting for someone to rerun an analysis and reissue a document.

Three services stand behind that. One serves the model. One computes the rollup and returns a verdict. One holds the document's structure. The change is made in either of two small web apps, a model viewer and a requirements document, and the argument does not depend on which.

The instructive part is what does not happen. Raise the throughput of a server that is not the bottleneck and nothing moves, because a serial chain is governed by its worst link. Raise the bottleneck instead and the capacity rises, but the requirement still fails, because the bottleneck has moved to the next weakest stage in the wiring. Raise one of the servers there and the requirement passes. That behaviour is obvious once seen and reliably surprising before.

None of the three services knows about the other two. The verdict a reader sees against a requirement comes from a service that has never parsed a model file, sitting beside text from a service that has never computed anything.

## Placeholders

Three parts of the demo are stand-ins, kept deliberately small so that the argument can be seen before the full mapping exists. Two belong to the adapter. There is no model repository behind it, so versioning is a counter rather than the commits of a conforming SysML v2 repository. Editing the model through the projection is a stand-in as well. It contradicts the position that a projection should be read-only, and a real deployment would write through the SysML v2 API instead. The third belongs to the example. Its capacity model is idealised, with parallel branches assuming evenly partitionable work and perfect load balancing, and no queueing effects represented anywhere. It is arithmetic chosen to make a point about federation rather than a performance model anyone should plan capacity with.

The plan is to replace the adapter's two stand-ins with a more complete mapping from the model to the graph. SysML v2 has views and viewpoints of its own, and my intention is to use them as the way a systems engineer chooses which parts of a model to federate and how those parts appear to the services outside. That will almost certainly need a full parser for the language rather than the strict subset the adapter reads today, and it is the largest piece of work on the horizon.

## Reading further

The design is written up as a series of articles under [docs/](docs/README.md). They start with the motivation and an overview of the architecture, then follow the design from the first research through to the version now being built, and they link the decision records behind every choice. The same articles are published at https://roarge.github.io/sysml-federation/.

## What this is not

It is not a product, and not a SysML v2 API implementation. The adapter reads files rather than fronting a repository, and its coverage of the language is a fraction of what the goal requires. What it is meant to prove is that the connecting layer can exist, in something small enough to read in an afternoon.

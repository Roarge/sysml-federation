# Why federate a systems model

*Roar Georgsen, 27 August 2026*

Part 1 of 11 in [Federating a systems model](../README.md).

## The documents that describe one system

Every engineering organisation I have worked with keeps its system in several places at once. Requirements in one file, interface definitions in another, the power budget in a spreadsheet, the hazard analysis somewhere a third team controls. All of them describe the same machine, and keeping them in agreement is manual work that gets deferred until a review forces it.

[Model based systems engineering](https://www.sebokwiki.org/wiki/Model-Based_Systems_Engineering_(MBSE)) was supposed to end that, and it has been five years away from mainstream adoption for about a quarter of a century. The [repository's README](https://github.com/Roarge/sysml-federation) makes that case at length, and I will not make it twice. Its short form is that models live in proprietary environments, that version control treats them as lockable binaries, that export is therefore always the cheapest option in the moment, and that the two industry answers to the problem, the [service bus](https://en.wikipedia.org/wiki/Enterprise_service_bus) generation and [OSLC](https://en.wikipedia.org/wiki/Open_Services_for_Lifecycle_Collaboration), both assumed an organisation with somebody employed to do integration full time. Firms with fewer than 25 engineers do not have that person. They are where most engineering happens and they are who this work is for.

What follows is the part of the argument the README asserts rather than defends: why federation is the right shape for the connecting layer, and what a running demo of it looks like.

## What SysML v2 leaves open

[SysML](https://en.wikipedia.org/wiki/Systems_Modeling_Language) v2 stopped being a profile of [UML](https://en.wikipedia.org/wiki/Unified_Modeling_Language) and acquired a formal foundation of its own, [KerML](https://www.omg.org/spec/KerML/). Two consequences follow, usually run together and worth keeping apart. Authoring is freed, because the language has a textual notation and a model becomes a set of files that live in [Git](https://en.wikipedia.org/wiki/Git) and diff in a pull request. Access is freed, because the standard carries an API and a JSON serialisation, so any conforming repository exposes projects, commits and elements the same way. For the first time the model has an interface that is not a vendor's.

What neither fixes is the shape of what comes back. The API hands you the metamodel. Elements are generically typed, relationships are themselves elements, and a requirement's text is reached by walking owned relationships to documentation comments. That is the right design for a modelling API and the wrong thing to hand a consumer. Every downstream tool has to learn the metamodel before it can do anything useful, so the compliance dashboard learns SysML, the reliability tool learns SysML, and the status page team learns it badly and reimplements the same traversal with a different set of bugs. The ontology gets pushed outward onto people who never asked for it, which is what happened with OSLC resource shapes.

## A projection instead

Nobody outside systems engineering should need to know what SysML is.

Let the model's owners publish a projection, a small and plainly typed view of the model. Systems engineers understand the metamodel and are the right people to decide that, for everyone downstream, a requirement is an identifier, a name, some text, a limit value and a set of things that satisfy it. Everything else stays inside the model.

Three conditions make such a projection worth anything. It has to be live rather than exported, so it cannot drift. It has to be joinable, so a service holding data the model does not contain can attach that data to the model's objects without either side knowing about the other. And the contract between producer and consumer has to be checked mechanically, before deployment, and not discovered in production.

Federation does all three, which is the whole reason it is here. A service declares the field by which a type is identified, any other service may contribute fields to the same type by declaring the same field, and a build step merges the fragments into one schema that nobody wrote and nobody owns. The README shows that as two `@key` declarations side by side, the adapter's and an analysis service's. What matters for the argument is what follows from it: the merged schema is computed, so there is no canonical data model to negotiate and no integration team to employ, and an incompatible change fails the merge in the pipeline of whoever pushed it. The choice of [federation over a single service](../decisions/AD-0001-federation-over-single-service.md) is recorded with the alternatives that lost.

The concept grew up inside GraphQL and still speaks its vocabulary, which is why this repository uses GraphQL. The tooling is mature there and nowhere else. Composition and query planning are separable from whatever carries the bytes, and I would rather have them over gRPC tomorrow than not have them today.

## Why Cosmo

A plain GraphQL server publishes one schema owned by one team, which is the central component again. Federation is a platform problem rather than a library problem, and it needs composition, breaking-change checks and a planner in front. Among the platforms that supply those, [Cosmo](https://github.com/wundergraph/cosmo) is chosen for reasons mostly not about GraphQL, set out against its alternatives in the record for [Cosmo as the platform](../decisions/AD-0002-cosmo-as-platform.md). It is [Apache 2.0](https://en.wikipedia.org/wiki/Apache_License) where the main alternative is not, composition runs locally with no control plane, and the router starts from a pre-built configuration file. As the vendor ships it the router still sends anonymous usage data, so [telemetry off](../decisions/AD-0013-telemetry-off.md) is baked into the image and no graph token is set, which is what makes the air gap that defence, rail, energy and medical device work all need.

The vendor is candid that it does not recommend the static path, and says of composing locally that "it is recommended to not use this for production". The demo answers that in its own terms rather than arguing with it. There is no control plane to fetch from, the composed configuration lives in [a committed file](../decisions/AD-0012-composition-committed.md), and a test fails when it drifts from the schemas it was built from. That test is the third condition made concrete, and it is the only part of the three that a reader can check by running something.

Underneath all of it, Cosmo is itself moving away from GraphQL as the transport. Its subgraphs can now be compiled to [protobuf](https://en.wikipedia.org/wiki/Protocol_Buffers) and served over [gRPC](https://en.wikipedia.org/wiki/GRPC), with GraphQL kept as the schema language and the edge protocol. I read that as an admission that the valuable part was never the wire format. It is the composition algebra, the entity key and the query planner. I am agnostic about protocols, and Cosmo's own direction suggests it will still be the layer I want when the wire format changes again.

## What the example is meant to show

The demo publishes a SysML v2 model of a query processing pipeline as a federated service, joined to a throughput analysis and a requirements document that know nothing about SysML. The pipeline has five servers wired in series and in parallel, each with a throughput. Capacity is the smallest stage in series and the sum across parallel servers. A requirement on the whole pipeline states the query rate it must sustain, and a requirement for each server is derived from it.

Three services stand behind one router, an arrangement [The architecture in one sitting](01-the-architecture-in-one-sitting.md) takes apart. One serves the model, one computes the rollup and returns a verdict, and one holds the document's structure. Two small web apps sit in front, a model viewer and a requirements document, and an edit can be made in either.

![Two apps, one graph, three services that never meet.](../img/overview-sketch.png)

*Two apps, one graph, three services that never meet.* The strip is cut from the [use case storyboard](../stories/use-cases.pdf).

What does not happen is the instructive part. Raise the throughput of a server that is not the bottleneck and nothing moves, because a serial chain is governed by its worst link. Raise the bottleneck and the capacity rises, but the requirement still fails, because the bottleneck has moved to the next weakest stage. Raise one of the servers there and the requirement passes. With the shipped values the pipeline sustains 1200 queries per second against a limit of 1500, and the bottleneck sits at the parse stage. That behaviour is obvious once seen and reliably surprising before. [Twelve use cases and one moving bottleneck](04-twelve-use-cases-and-one-moving-bottleneck.md) walks through it.

The verdict a reader sees against a requirement comes from a service that has never parsed a model file, sitting beside text from a service that has never computed anything.

## Placeholders and the road ahead

Three parts of the demo are stand-ins. Two belong to the adapter. There is no model repository behind it, only [a directory of files](../decisions/AD-0003-adapter-reads-files.md), so versioning is a counter, not the commit history a conforming repository would give. Editing the model through the projection is [scaffolding](../decisions/AD-0004-editing-as-scaffolding.md) as well, since a projection should be read-only and a real deployment would write through the SysML v2 API. The third stand-in belongs to the example. Its [idealised capacity model](../decisions/AD-0006-idealised-capacity-model.md) assumes evenly partitionable work across parallel branches, perfect load balancing and no queueing anywhere. It is arithmetic chosen to make a point about federation, and nobody should plan capacity with it.

Next comes a more complete mapping from the model to the graph, replacing the adapter's two stand-ins. That will almost certainly need a full parser, well beyond the strict subset the adapter reads today, and it is the largest piece of work on the horizon.

The interesting question underneath it is where the projection should come from. A viewpoint in SysML names a concern and the stakeholder who holds it, and a view is the part of the model rendered for that concern. A subgraph, seen from its consumer, is the same thing, the slice of the model one downstream party needs, shaped for that party's vocabulary. If the model already says which stakeholder sees what, the projection should be read from it and not authored a second time in a schema. Full language coverage pulls the other way, towards a few hundred types generated mechanically from KerML, which lands the consumer back in the abstract syntax the projection was meant to spare them. My current view is that coverage should grow as [curated projections](../decisions/AD-0005-curated-generic-projection.md), one concern at a time. I am not certain of that, and it is the part of the design most likely to change.

---

Index: [Federating a systems model](../README.md) · Next: [The architecture in one sitting](01-the-architecture-in-one-sitting.md)

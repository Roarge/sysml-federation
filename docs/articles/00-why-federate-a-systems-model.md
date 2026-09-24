# Why federate a systems model?

*Roar Georgsen, 27 August 2026*

Part 1 of 13 in [Federating a systems model](../README.md), written for release v0.3.0.

## One system, many documents

Every engineering organisation I've worked with keeps its system in several places at once. The requirements live in one file and the interface definitions in another. The power budget sits in a spreadsheet, and the hazard analysis is somewhere a third team controls. All of them describe the same machine. Keeping them in agreement is manual work that nobody enjoys and everybody puts off until a review forces the issue.

[Model-based systems engineering](https://www.sebokwiki.org/wiki/Model-Based_Systems_Engineering_(MBSE)) was supposed to end that. The documents give way to a model, one structured description holding the system's parts, the properties they carry, the requirements they must meet and the relationships between all of it. Documents become views of the model. Change a component's power draw and everything downstream can be recomputed, including which requirements now fail.

That promise has been five years from mainstream adoption for about 25 years, and the least disputed explanation is the tooling.

Models live inside proprietary environments. They're stored in binary formats, or in a vendor's own flavour of <span class="term" data-term="xmi">XMI</span> that only that vendor reads back reliably, and you reach them through an API that is either missing or shaped differently in every product. Several of the big environments still treat the model as a binary file you lock rather than something you merge, and offer nothing a build server can drive. So the model can't join the <span class="term" data-term="ci">continuous integration</span> workflow every other discipline now takes for granted.

Vendors will tell you integration is a solved problem, and they aren't exactly lying. Connectors to requirements tools, test management and <span class="term" data-term="plm">PLM</span> exist. They are also, with tiresome regularity, a separately licensed product, configured through a dialogue box and liable to break at the next upgrade of either end. Two-way synchronisation exists too, but setting it up takes a different skill from the one the engineer who needs it has, so more often than not it stays unconfigured and everyone falls back to export.

That is what lock-in looks like in practice. It's rarely a clause in a contract. It's a pile of small daily frictions that make export always the cheapest option in the moment. Requirements go to Word for review, interface definitions reach the software team in a spreadsheet, and the power budget lives in Excel because that's where the person who owns it works.

Every one of those copies starts drifting the moment it is made.

## We've tried this before

The industry has had two serious goes at the problem. The <span class="term" data-term="service-bus">service bus</span> generation put a central integration platform in the middle, with a canonical data model owned by an integration team. <span class="term" data-term="oslc">OSLC</span> took a lighter route through linked data, so tools could refer to each other's objects without importing them.

Both put the integration in the middle, where it needs a team of its own. That team owns the mapping between systems it didn't build, so the mapping ends up written by the people furthest from the thing being mapped. A large organisation can afford the team, and still ends up with the knowledge and the responsibility in different places.

A small organisation can't staff that middle at all. Most engineering happens in organisations with fewer than 25 engineers, and they are who this work is for.

## What SysML v2 settles

[SysML](https://en.wikipedia.org/wiki/Systems_Modeling_Language) is the standard modelling language for systems engineering. Its first version was a profile of [UML](https://en.wikipedia.org/wiki/Unified_Modeling_Language), a language built for software design, so systems engineers described pumps and power lines in a vocabulary shaped for somebody else's problems. Semantics defined by a profile are semantics defined by convention, and two tools could apply the same profile differently and both still be valid.

Version 2 sits on a foundation of its own, [KerML](https://www.omg.org/spec/KerML/), with semantics defined formally. The [OMG](https://en.wikipedia.org/wiki/Object_Management_Group) announced its adoption on 21 July 2025 ([press release](https://www.omg.org/news/releases/pr2025/07-21-25.htm)), and the language specification is published as formal/26-03-02, which is [the version this work targets](../decisions/AD-0019-sysml-2-0-target.md).

That removes two kinds of lock-in, and they're worth keeping apart.

The first is authoring. The language has a textual notation, so a model is a set of text files that live in [Git](https://en.wikipedia.org/wiki/Git), show their differences in a pull request and can be written in any editor. That's the change everyone talks about.

The second is access, and for this series it matters more. SysML v2 comes with a standard API and a standard JSON serialisation. Any conforming repository exposes a model the same way, and any conforming tool can read the same serialised model. For the first time the model has a defined interface that isn't a vendor's.

## What it leaves open

Neither fixes the shape of what comes back. The API hands you the <span class="term" data-term="metamodel">metamodel</span>. Elements are generically typed, relationships are themselves elements, and you reach a requirement's text by walking owned relationships to documentation comments. That's the right design for a modelling API. It also means every consumer has to learn the metamodel before it can do anything useful.

So the compliance dashboard team learns SysML. The reliability tool learns SysML. The people building the customer-facing status page learn SysML badly, and write the same traversal again with a different set of bugs. That's what happened with OSLC, and it's why lifecycle integration stayed the preserve of organisations that could staff it.

## A projection instead

Nobody outside systems engineering should need to know what SysML is.

My proposal is to let the model's owners publish a <span class="term" data-term="projection">projection</span>. Systems engineers understand the metamodel, and they're the right people to decide that, for everyone downstream, a requirement is an identifier, a name, some text, a limit and a set of things that satisfy it. Everything else stays inside the model. A consumer meets a small, plainly typed object and never runs into KerML.

Three conditions make such a projection worth having:

- It has to be **live** rather than exported, so it can't drift.
- It has to be **joinable**. A service holding data the model doesn't contain should be able to attach that data to the model's objects without either side knowing about the other.
- The contract between producer and consumer has to be **checked mechanically** before deployment, instead of being discovered in production.

A projection is also a read path. Writes belong to the model's own API.

Federation meets all three conditions. The idea grew up inside GraphQL, which is why this series uses GraphQL. I chose it for the maturity of the tooling, and I have no attachment to the protocol. The claim underneath this whole series is that federation is the missing integration layer for open MBSE. I've recorded the choice of [federation over a single service](../decisions/AD-0001-federation-over-single-service.md), with the alternatives that lost.

## Federation, for a systems engineer

[GraphQL](https://en.wikipedia.org/wiki/GraphQL) is a style of API where the server publishes a schema describing its types and their fields, and a client sends a query naming exactly the fields it wants. There is one endpoint, and the response has the same shape as the request.

Federation stretches that across services. Several independently owned services each publish a fragment of a schema, called a subgraph. A build step merges the fragments into one schema. At run time a component called the router accepts a query written against the merged schema, works out which service holds which field, calls each one and assembles the result. The client sees one graph and has no idea how many services stand behind it.

> [!NOTE]
> **Entity key**
>
> The field that tells federation two services are talking about the same object. One service declares that a type is identified by a particular field. Any other service can then add fields to that same type by declaring the same key. When a client asks for fields from both, the router fetches from each and joins the answers on the key.

In this demo the adapter says a requirement has an identifier and some text (both snippets are trimmed to the fields that matter here):

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

Neither service imports the other, and neither calls the other. A client asks for a requirement's text and its verdict in one query, and the router fetches from both and joins them on the key. To a systems engineer, that's a cross-tool trace link that resolves by itself and gets checked before it ships. For an element of a SysML model, I made the key [the short name, with the qualified name as fallback](../decisions/AD-0018-short-names-as-keys.md).

[![The adapter and the capacity service each declare Requirement with the key id. The adapter adds text and the capacity service adds verdict. A client sends one query for the text and verdict of PIPE-R1 to the router, which fetches from both services and returns one answer joined on the id.](../figures/entity-key-join.svg)](../figures/entity-key-join.svg)

*One query, two services that never call each other, one answer joined on the key.*

Two properties follow, and they matter more than the syntax.

**Nobody writes the merged schema.** It's computed, so nobody owns it, which is the opposite of a canonical data model negotiated by committee. The people who build each service, and understand its domain, write its schema. If two services define incompatible things, the merge fails in the pipeline of whoever pushed the change, with a message naming the conflict.

**The router holds no logic.** It transforms nothing and enforces no rules of its own. Everything the service bus generation put in the middle stays in the services, which keeps the middle from becoming a bottleneck.

## A model of models

Once that arrangement is in place, the word model starts to mean something bigger. The merged schema can carry requirements from a systems model, verification results from a test management system, part costs from PLM and field behaviour from an operational data store. Each is a model of some aspect of the same thing, held in whatever formalism its own discipline settled on. Most won't be SysML, and there's no reason they should be.

What the merged schema describes is a [system of systems](https://www.sebokwiki.org/wiki/Systems_of_Systems_(SoS)): constituent systems that are useful in their own right, owned by different parties, and together do something none of them does alone. Each team models what it understands, at its own pace. Adding a twelfth model doesn't mean renegotiating a schema with the other eleven, because keys are the only thing anyone has to agree on. The alternative the industry keeps trying, one very large model covering everything, concentrates authority in whoever owns its schema.

## Why Cosmo

A plain GraphQL server publishes one schema owned by one team, which puts us straight back in the middle. Federation needs a platform: composition, checks for breaking changes, and a planner in front. Of the platforms that do this, I chose [Cosmo](https://github.com/wundergraph/cosmo), mostly for reasons that have little to do with GraphQL ([Cosmo as the platform](../decisions/AD-0002-cosmo-as-platform.md)).

It's under the [Apache 2.0](https://en.wikipedia.org/wiki/Apache_License) licence, where the main alternative uses a licence the [OSI](https://en.wikipedia.org/wiki/Open_Source_Initiative) doesn't recognise as open source. For an argument aimed at small organisations, that isn't a footnote. Composition runs locally with no <span class="term" data-term="control-plane">control plane</span>, and the router can start from a pre-built configuration file.

As the vendor ships it, the router sends anonymous usage data. The image turns that off with environment variables ([telemetry off](../decisions/AD-0013-telemetry-off.md)), and with no graph token set, nothing the router does reaches outside the container on any code path I've read. I ran the image with its network removed to check. Defence, rail, energy and medical device work all need that, and those are the industries most likely to be modelling in SysML in the first place.

The vendor is candid that starting the router from a file, with no control plane, isn't the path it recommends. The composition page says "it is recommended to not use this for production", and the router logs "Not recommended for Production" when it starts from a file. I take the warning at face value and answer it on the demo's own terms. There's no control plane to fetch from, the composed configuration lives in [a committed file](../decisions/AD-0012-composition-committed.md), and a test fails when that file drifts from the schemas it was built from. That test is the third condition made concrete.

The deeper reason is that Cosmo itself is moving away from GraphQL as the transport. Its subgraphs can now be compiled to [protobuf](https://en.wikipedia.org/wiki/Protocol_Buffers) and served over [gRPC](https://en.wikipedia.org/wiki/GRPC), with GraphQL kept as the schema language. I read that as an admission that the valuable part was never the wire format. It's the rules by which schemas merge, the entity key and the query planner, which is my position too.

## What the demo is meant to show

The demo publishes a SysML v2 model of a query processing pipeline as a federated service. A throughput analysis and a requirements document join it, and neither knows anything about SysML.

The pipeline is five servers wired as stages in series and in parallel, each with a throughput. A model requirement on the whole pipeline states the query rate it must sustain, and a requirement for each server is derived from it. Three services stand behind one router, an arrangement [The architecture in one sitting](01-the-architecture-in-one-sitting.md) takes apart:

- the **adapter** serves the model
- the **capacity service** works out the pipeline's capacity from the servers and returns a verdict for each requirement
- the **document service** holds [the document's structure](../decisions/AD-0025-document-owns-its-structure.md) and nothing else

Two small web apps sit in front, a model viewer and a requirements document, and you can make an edit in either. Change one server's throughput and the document responds at once. The capacity moves, the requirement passes or fails, and where it fails, the reason names the server that limits it.

What *doesn't* happen is the instructive part. Raise a server that isn't the bottleneck and the capacity stays put, because a chain is governed by its worst link. Raise the bottleneck and the capacity rises, but the requirement still fails, because the bottleneck has moved to the next weakest stage. With the shipped values the pipeline sustains 1200 queries per second against a limit of 1500, and the bottleneck sits at the parse stage. It's obvious once you've seen it and reliably surprising before. [Twelve use cases and one moving bottleneck](04-twelve-use-cases-and-one-moving-bottleneck.md) walks through it.

The verdict you see against a requirement comes from a service that has never parsed a model file, sitting beside text from a service that has never computed anything.

I also keep a SysML v2 model of the demo itself, which I should have done from the start. [A model of the demo itself](12-a-model-of-the-demo-itself.md) owns up to that.

## Stand-ins

Three parts of the demo are stand-ins. The adapter reads [a directory of files](../decisions/AD-0003-adapter-reads-files.md) where a real deployment would front a SysML v2 repository. Editing through the projection is [scaffolding](../decisions/AD-0004-editing-as-scaffolding.md) that contradicts the read-path rule above, and a real deployment would write through the SysML v2 API. The capacity arithmetic is [deliberately simple](../decisions/AD-0006-idealised-capacity-model.md), enough to make the point and not meant for real-world use.

The next step is a fuller mapping from model to graph. SysML v2 has <span class="term" data-term="view">views</span> and viewpoints of its own: a viewpoint names a concern and the stakeholder who holds it, and a view is the part of the model rendered for that concern. A subgraph, seen from its consumer, is the same thing. If the model already says which stakeholder sees what, the projection should be read from it rather than written out a second time in a schema. Covering the whole language would pull the other way, towards a few hundred types generated mechanically from KerML, which is exactly the abstract syntax the projection exists to spare its consumers. My current view is that coverage should grow as [curated projections](../decisions/AD-0005-curated-generic-projection.md), one concern at a time. I'm not certain of that, and it's the part of the design most likely to change.

The adapter and the example are in [the repository](https://github.com/Roarge/sysml-federation), and both are small enough to read in an afternoon. Part 2 opens the box.

---

Index: [Federating a systems model](../README.md) · Next: [The architecture in one sitting](01-the-architecture-in-one-sitting.md)

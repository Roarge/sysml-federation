# Architecture views

The five views and the overview board are one PDF, [architecture-views.pdf](architecture-views.pdf), one page per board. This page is the text form of each view. Together with [the decision records](../decisions/README.md) it is the record of the architecture, as [the architecture record decision](../decisions/AD-0021-architecture-record.md) sets out. Around it, the article series tells the story, starting with [The architecture in one sitting](../articles/01-the-architecture-in-one-sitting.md) and [Five views and twenty-six decisions](../articles/06-five-views-and-twenty-six-decisions.md).

This description follows the concepts of ISO/IEC/IEEE 42010 (stakeholder, concern, viewpoint, view, decision, rationale) without a claim of conformance. The entity of interest is the demo: the adapter, the example stack behind it and the image they ship in. By contrast, the pipeline the example models is a system the demo describes, and runs nowhere.

## Stakeholders and concerns

| Stakeholder | Concern | Answered by |
|---|---|---|
| Visitor | Does it launch, and does it show something I would not have believed? | Context, runtime |
| Model owner | Does the model stay the source, and does an edit anywhere land in it? | Runtime, adapter |
| Document owner | Can I shape the document without touching the model, and still see what the model knows? | Composition, runtime |
| Maintainer or contributor | Can it be read in an afternoon, built with Go alone and changed one service at a time? | Deployment, adapter |
| Organisation deciding whether it fits | What do the services agree on, and what would we replace? | Composition |

The last concern is answered by the adoption contract. The services agree on the entity key (a SysML short name), the `@requires` field set the capacity service declares, and that service's two configured names. An adopter keeps the adapter, the compose step and the supervisor, and replaces the model file, the two example services, the two apps and the shipped document tree.

## How the views correspond

Each view is a model of one kind in one notation, with this text form and a board. Across the views, the three subgraphs carry the names `model`, `capacity` and `document` into composition. They run as the three goroutines of the deployment view's process tree, on the ports that view names, and are the packages `adapter/` and `examples/pipeline/{capacity,document}/`. A drift test guards the field ownership table below. Step 5 of the runtime view carries the field sets the capacity service requires between services, and the adapter view's packages are the adapter's single box in the other views.

## Context

Inside the boundary are three subgraphs, one router, one UI server and two web apps, all in one container started by `docker run --rm -p 8080:8080 ghcr.io/roarge/sysml-federation`. Outside it are a browser with two tabs or two browsers, the Docker host, and the GitHub container registry the image is pulled from once. The demo makes no connection outside the container while it runs.

Two things sit at the edge. The example model file is content the adapter reads, written in the repository, and the Cosmo router is a vendor binary copied into the image and driven only by configuration and environment. What crosses the boundary is HTTP from the browser to one published port and the image pull. The composition step and the tags that publish the image run outside the container.

## Composition

The three subgraphs resolve fields that do not overlap.

| Subgraph | Resolves |
|---|---|
| `model` (the adapter) | everything that comes from the source: `Model`, `Part`, `Attribute`, `Port`, `Connection`, `Requirement`, `VerificationCase`, the query roots, the mutations `setAttribute`, `setLimit` and `resetModel`, and the `modelChanged` subscription |
| `capacity` | `Part.capacity`, `Part.bottleneck`, `Requirement.verdict`, `Requirement.verdictReason`, computed on read from the fields it requires |
| `document` | the document query and its tree of nodes, `Requirement.documentNumber`, `Requirement.included`, seven mutations and the `documentChanged` subscription |

Three types are entities, `Part`, `Requirement` and `VerificationCase`, keyed on `id`, whose value is the SysML short name or, where an element declares none, its qualified name ([short names as keys](../decisions/AD-0018-short-names-as-keys.md)). The entity resolvers generated from the `@key` directives are the contract between services as it exists in code.

The adapter's schema is generic and names nothing from the example ([a curated generic projection](../decisions/AD-0005-curated-generic-projection.md)). `setAttribute` and `setLimit` accept any value that is a literal in the source and refuse the rest with an error naming the element and the reason.

The capacity service carries two configured names, `capacity` and `throughput`, and no other word of the model. Its requires selections read the subject's attributes, its child parts with their attributes, and the connections between them, and part names travel in every selection because the reasons name the servers in the cut:

```graphql
capacity: Float
  @requires(fields: "attributes { name value } parts { id name attributes { name value } } connections { from to }")
```

The value types carry `@external` on the fields the selections read and no `@shareable`. `VerificationCase` is keyed `resolvable: false`, because the capacity service only reads it. The first spike of the build confirmed that Cosmo composition and gqlgen's `explicit_requires` carry this nested shape intact.

The document service's shipped tree names the example's requirement ids in a configuration file, the one place the example's identifiers enter a service. It never reads the model. For an id it has never heard of, its entity resolver answers `included: false` and `documentNumber: null` ([the document owns its structure](../decisions/AD-0025-document-owns-its-structure.md)).

In the merged schema, `Requirement` carries text and limit from the adapter, verdict and reason from the capacity service, and number and inclusion from the document service. The router resolves `requirement(id: "PIPE-R1") { text verdict verdictReason documentNumber }` with three fetches. It asks the adapter for the requirement and every field the capacity service requires, then the capacity service's entity resolver with those fields as the representation, then the document service's entity resolver with the key.

Both the compose input and its output are committed at `examples/pipeline/`, and a Go test compares each schema embedded in the output with the file it came from ([composition committed](../decisions/AD-0012-composition-committed.md)). The vendor says of this static path that "it is recommended to not use this for production", and the router logs "Not recommended for Production" at every start without a graph token. The demo has no control plane to fetch from, so the static path is the only one, and the answer to the caveat is the drift test and pinning the router and the compose tool together.

Both apps are clients of the router and of nothing else. The viewer subscribes to `modelChanged`, the document app to both events, and each refetches its whole query on every event. Neither renders from a mutation's response or computes anything it shows.

## Runtime

One value edit takes seven steps:

1. The viewer sends `setAttribute` to `/graphql`.
2. The UI server passes it to the router, which plans it to the adapter.
3. The adapter validates the value, patches the literal at its span, rebuilds the projection, increments the version and emits `modelChanged`.
4. The router fans the event out to every subscribed browser as server-sent events.
5. Both apps refetch. The router plans the document's query across all three services, with representations built from the adapter's answer.
6. The capacity service builds the flow network, runs the flow, finds the cut and returns capacity, bottleneck, verdict and reason.
7. Both apps render.

From the version change to both apps rendering is bounded at two seconds. Raising a server outside the cut runs the same seven steps and changes only that server's figure.

A document edit goes to the document service alone. It reorders its tree, renumbers, increments its version and emits `documentChanged`. The viewer is not subscribed to document events, and the model version is unchanged.

Reset from either app is one mutation document with both root fields, `mutation { resetModel { version } resetDocument { version } }`. The router plans the two fields to the two services, each restores its shipped state, increments its counter and emits its event, and both apps refetch. The apps send both because neither service knows the other exists.

At start-up, `serve` starts the three subgraphs as goroutines on loopback and waits for each health endpoint, starts the router as a child process, waits for the router's `/health/ready`, and only then opens the published port. `/health/ready` reports that the router has loaded its configuration, not that the subgraphs answer, so the ordering does not depend on it.

## Deployment

```
PID 1  sysml-federation serve            Go, supervisor, PID 1
  |-- goroutine  adapter        127.0.0.1:3011   gqlgen subgraph "model"
  |-- goroutine  capacity       127.0.0.1:3012   gqlgen subgraph "capacity"
  |-- goroutine  document       127.0.0.1:3013   gqlgen subgraph "document"
  |-- child      /router        127.0.0.1:3002   Cosmo router 0.343.1
  \-- goroutine  ui server      0.0.0.0:8080     static apps, proxy
```

The UI server serves `/viewer`, `/document` and `/shared` from an embedded filesystem, proxies `/graphql` and `/playground` to the router, and redirects `/` to `/viewer`. A `healthcheck` subcommand probes the router's `/health/ready` on loopback and `/viewer` on the published port, because the distroless base has no shell or curl. The router's health path is not proxied. Four further subcommands, `adapter`, `capacity`, `document` and `ui`, run one component each on a given address ([one binary, one port](../decisions/AD-0011-one-binary-one-port.md), [the router as a child process](../decisions/AD-0010-router-as-child-process.md)).

The Dockerfile has three stages:

1. a Go stage that cross-compiles the binary for the target platform
2. a stage named `router`, which is the vendor's image at the pinned version
3. a distroless static base running as nonroot, which receives the router binary, the demo binary, the committed configuration and the model file

At build time the router's licence is fetched at the pinned tag with a pinned checksum. The image is published for `linux/amd64` and `linux/arm64` by a workflow triggered on a `v*` tag, which reads the manifest back and fails if either platform is missing or over the 80,000,000-byte budget ([publish on tags](../decisions/AD-0020-publish-on-tags.md)).

## The adapter

```
files --> lexer --> parser --> AST with spans --> resolver --> projection --> gqlgen
                                   |                              ^
                                   +---- patch literal at span ----+   (setAttribute, setLimit)
```

`syntax` turns the text into a tree in which every node carries its byte range in the source. It accepts a strict subset of the textual notation and refuses anything else with the file, line and column of the first offending token ([a hand-written subset parser](../decisions/AD-0015-hand-written-subset-parser.md)).

`model` resolves qualified names, short names and feature chains. It evaluates the bound expressions the subset allows: a literal with an optional unit, a feature chain, the four arithmetic operators and parentheses. It reads each `require constraint` into a quantity, a comparison and a limit ([quantity from the constraint](../decisions/AD-0008-quantity-from-constraint.md)), and checks port directions against the order of each connection's ends ([connection direction](../decisions/AD-0009-connection-direction.md)). `Patch(span, newLiteral)` replaces the text and re-parses, so the served text and the projection are rebuilt together.

`projection` maps resolved elements to the schema's types and holds nothing the schema does not show. `serve` is the gqlgen server with WebSocket subscriptions, the entity resolvers, a store that hands each operation one snapshot of the current model, and the version counter.

A second fixture model with other names and wiring is part of the adapter's tests. A construct outside the subset is refused rather than served generically. The example model is checked against the two reference tools locally before it becomes a fixture ([the SysML 2.0 target](../decisions/AD-0019-sysml-2-0-target.md)).

---

Index: [Federating a systems model](../README.md) · Repository: https://github.com/Roarge/sysml-federation

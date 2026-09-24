# The architecture in one sitting

*Roar Georgsen, 27 August 2026*

Part 2 of 13 in [Federating a systems model](../README.md), written for release v0.3.0.

> [!IMPORTANT]
> **The story so far**
>
> Part 1 argued that a systems model should publish a plain projection of itself, and that federation lets other services attach their own data to it, joined on an entity key. This part opens the demo I built to show that. A SysML v2 model of a five-server query pipeline goes out through a federated GraphQL router. A capacity analysis and a requirements document attach their own data to its parts and requirements, and neither of them has ever read a line of SysML. All of it runs in one container.

![The L0 sheet: the argument in eight steps, what is agreed, and what is in the box](../img/a3-l0-model-side.png)

*The L0 sheet: the argument in eight steps, what is agreed, and what is in the box. Cut from the [L0 A3 sheet](../a3/L0-federating-a-systems-model.pdf).*

## What's in the box

One command starts it. The first tagged release published the image that command names, with a manifest for amd64 and one for arm64. The package is public, and a host holding no registry credentials pulls it, and the untagged form below resolves to `latest`.

> [!TIP]
> **Run it yourself**
>
> ```
> docker run --rm -p 8080:8080 ghcr.io/roarge/sysml-federation
> ```
>
> Then open `http://localhost:8080` in a browser. I build and test on Ubuntu under WSL, and [The demo as it shipped](10-the-demo-as-it-shipped.md) says what else it has been tried on.

Inside, one Go binary runs as PID 1, the first process in the container, and supervises everything else. Three subgraphs run as <span class="term" data-term="goroutine">goroutines</span> of that binary, on <span class="term" data-term="loopback">loopback</span> ports 3011, 3012 and 3013. Each is a GraphQL service that owns one slice of the merged schema. The Cosmo router merges the three schemas into one and plans every incoming query across them. It runs as a child process on loopback port 3002, from a binary copied out of the vendor's image.

The only thing the outside world can reach is a UI server on port 8080. It serves the two web apps at `/viewer` and `/document` from embedded assets. At `/shared` it serves the client module both apps import, along with the drag library vendored beside it. It passes `/graphql` and `/playground` through to the router, and it redirects `/` to the viewer. That's five paths on one port, and the router is never addressable on its own.

![One container, one process tree](../img/v4-container.png)

*One container, one process tree. Cut from the [architecture views](../architecture/architecture-views.pdf).*

The router is about 40 MB compressed, the demo binary a few MB and the base image 2 MB, against a budget of 80 MB.

<details markdown="1">
<summary>Under the bonnet: start-up order, the health check and the subcommands</summary>

The supervisor fixes the order of start-up:

1. The subgraphs come up first, and the supervisor polls each one's own health endpoint.
2. It starts the router with the configuration path and the telemetry variables in its environment.
3. It waits on the router's `/health/ready`.
4. Only then does it open port 8080.

A `healthcheck` subcommand probes the router's readiness path and `/viewer` on behalf of the image's HEALTHCHECK. It has to, because the <span class="term" data-term="distroless">distroless</span> base has no shell and no curl to call.

Four further subcommands, `adapter`, `capacity`, `document` and `ui`, each run one component on a given address, for anyone who wants to see the services apart.

</details>

The two web apps on that port are what a visitor meets, and neither of them works anything out for itself. The wiring, the arithmetic and the document ordering all arrive from services that share no code.

![The model viewer in the shipped state, with parse outlined red and the pipeline requirement failing beneath it](../img/app-viewer-shipped.png)

*The model viewer. The sketch is drawn from the model's own connect statements, parse holds the pipeline to 1200, and PIPE-R1 fails against its limit of 1500.*

![The same requirements as a numbered document, with five derived requirements nested under the first](../img/app-document-tree.png)

*The requirements document. Its numbering, headings and prose belong to a service that computes nothing, and the verdict on each row comes from one that has never read a model file.*

## Three services and what each owns

![Two apps, one graph, three services that never meet](../img/overview-sketch.png)

*Two apps, one graph, three services that never meet. Cut from the [use case storyboard](../stories/use-cases.pdf).*

**The adapter owns the model.** It publishes the parts, with their attributes, ports and connections and the requirements they satisfy. It publishes the requirements themselves, each with its text, subject, constrained quantity, comparison, limit and derivation links. It also publishes the verification cases, and the model's own text and version number.

The adapter is generic. Nothing in its schema or its code names the example, and its tests include a second fixture model with other names and other wiring. Its schema carries three <span class="term" data-term="mutation">mutations</span>, and they are the whole of the model's edit path:

```graphql
setAttribute(partId: ID!, name: String!, value: Float!): Part!
setLimit(requirementId: ID!, value: Float!): Requirement!
resetModel: Model!
```

**The capacity service owns four fields** and nothing else: `capacity` and `bottleneck` on a part, and `verdict` and `verdictReason` on a requirement. It computes all four on every read, from fields the router brings in because of a `@requires` declaration.

> [!NOTE]
> **@requires**
>
> A federation directive a service puts on a field it computes. It says, in effect, "before you ask me for this field, fetch these other fields from whoever owns them and hand them to me". The router does the fetching, so the service needs no copy of anyone else's data and never calls another service. It's how the capacity service gets the wiring and throughputs it needs without ever reading the model.

The service holds no copy of the model. It is configured with two names, `capacity` for the quantity it computes and `throughput` for the attribute it reads from each child part, and it never sees the words "server" or "pipeline". A verdict, the service's answer for a requirement, is one of PASS, FAIL, INCONCLUSIVE and ERROR. Those are the four words of the SysML v2 Systems Library, so a reader of the model meets the same four in the analysis.

**The document service owns an ordered tree** and its numbering. It ships one structure. At the top sits an unnumbered paragraph of prose explaining why an allocated limit on a server can fail while the pipeline as a whole passes. Then comes `PIPE-R1`, the model requirement that the pipeline shall sustain the required query rate, with its five derived requirements nested beneath it in server order and numbered 1.1 to 1.5. Last comes `PIPE-R2`, the latency requirement, as 2.

The tree names those seven requirement ids in a configuration file, which is the one place the example's identifiers legitimately enter a service. The document service never reads the model. Ask its <span class="term" data-term="entity-resolver">entity resolver</span> about an id it has never heard of, and it answers that the requirement isn't included and has no number.

[![Three services, each with what it owns. The adapter owns the model, the capacity service owns four computed fields and keeps no copy of the model, and the document service owns an ordered tree and never reads the model. Below them, the only things they agree on: the entity key, the fields the capacity service requires, and its two configured names.](../figures/what-the-services-share.svg)](../figures/what-the-services-share.svg)

*Nothing else needs agreeing.*

The three share one thing above all, the entity key, which is the field by which two subgraphs agree they're describing the same object. Here it is `id` on `Part`, `Requirement` and `VerificationCase`, and its value is the SysML short name.

So the adoption contract, the whole agreement between the services, fits in one sentence. It is the entity key (a SysML short name), the `@requires` field set the capacity service declares, and that service's two configured names. An organisation adopting the approach keeps the adapter, the compose step and the supervisor. It replaces the model file, the two example services, the two apps and the shipped document tree. If two services define the same field, or a `@requires` names a field the adapter no longer projects, composition fails in the pipeline of whoever pushed the change.

## One query, three fetches

Composition merges the three schemas into one. In it, a requirement carries its text and limit from the adapter, its verdict and reason from the capacity service, and its document number and inclusion from the document service. The query a visitor runs in the playground, the in-browser query editor at `/playground`, asks for fields from all three, two of them from the capacity service.

```graphql
{
  requirement(id: "PIPE-R1") {
    text
    verdict
    verdictReason
    documentNumber
  }
}
```

![Requirement as the router serves it, each field chipped with the service that resolves it](../img/v2-merged-requirement.png)

*Requirement as the router serves it, each field chipped with the service that resolves it.*

The router plans it as three fetches:

1. The first goes to the adapter, for the requirement and for every field the capacity service's `@requires` names. For a verdict, that means the subject part, its children with their names and attributes, and the connections between them.
2. The second goes to the capacity service's entity resolver, with those fields packed up as the representation, meaning the key plus the fields the service asked for.
3. The third goes to the document service's entity resolver, with nothing but the key.

The shipped answer is the text "The pipeline shall sustain the required query rate", the verdict FAIL, the reason "capacity 1200 against 1500, limited by parse", and the document number "1". Ask the same endpoint for the pipeline part's capacity and bottleneck and you get 1200 and `parse`.

That reason string deserves a second look. Both "capacity" and "parse" come from the model, and neither is written into the template. One arrives as the service's configured quantity name, the other as a part name the router carried in.

## One edit, seven steps

The bottleneck moving is the demo's one memorable moment.

![One edit, seven steps](../img/v3-runtime.png)

*One edit, seven steps.*

Here's what happens when a visitor raises `parse` from 1200 to 1700 in the viewer:

1. The app sends `setAttribute` for part `PIPE-S2`, attribute `throughput`, value 1700, to `/graphql`.
2. The UI server passes it to the router, which plans it to the adapter.
3. The adapter checks that the value is a finite, non-negative number and that the attribute is a literal in the source. It replaces the literal at its recorded span in the in-memory text and rebuilds the projection, the plainly typed GraphQL view of the model it serves. Then it bumps the version and emits `modelChanged: 2` on its <span class="term" data-term="subscription">subscription</span>.
4. The router holds the one upstream subscription and fans the event out to every subscribed browser as <span class="term" data-term="server-sent-events">server-sent events</span>. They pass through the UI server's <span class="term" data-term="reverse-proxy">reverse proxy</span>, which flushes on `text/event-stream`.
5. Both apps refetch their own query. The router plans the document's query across all three services, with representations built from the adapter's answer.
6. The capacity service builds the flow network from each representation, runs the flow and finds the cut, the cheapest set of servers that limits it. It answers capacity 1400, bottleneck `indexA` and `indexB`, verdict FAIL, and the reason "capacity 1400 against 1500, limited by indexA, indexB".
7. Both apps render. The viewer moves its red outline from `parse` to the index pair, and the document flips `PIPE-R1.2` to PASS and rewrites the reason under `PIPE-R1`.

A requirement bounds the whole thing. Once the version changes, each app has to show the change within two seconds and without a page reload.

Neither app renders from the mutation's response. The mutation goes out, and the event that follows triggers the same refetch as any other event, so there's only one rendering path. Neither app computes anything either. The viewer's red block, its capacity figure and its highlighted bottleneck all arrive through the router, and so does the document's reason text.

Raise `ingest` to 3000 instead and the same seven steps run. The cut is still `parse`, every number and verdict stays the same, and the only visible change is the figure on `ingest`. The event still fires and both apps still refetch. Nowhere does the demo detect that nothing changed, and it doesn't try to.

Reset, from either app, sends one mutation document carrying two root fields, `resetModel` and `resetDocument`. The router plans the two fields to the two services. Each restores its shipped state, bumps its own counter and emits its own event. The apps have to ask for both because neither service knows the other exists.

## Inside the adapter

![Inside the adapter](../img/v5-adapter.png)

*Inside the adapter.*

The adapter is a pipeline of four packages, `syntax` to `model` to `projection` to `serve`, with a patch path from a mutation back to the syntax tree. Every node carries its byte range in the source. That's what lets an edit replace one literal in place and rebuild the served text and the projection together. [Five views and twenty-six decisions](06-five-views-and-twenty-six-decisions.md) takes the four packages apart.

The part most worth having here is the refusal path, which works in both directions.

At start-up, a construct outside the subset stops the load with a file, line and column, instead of being served generically. That includes a numeric literal the parser can't read as a number.

At run time, a mutation against a value bound by an expression is refused with an error saying the value isn't a literal in the source. A negative or non-finite number is refused with an error saying the value must be a finite, non-negative number. Neither refusal emits an event. So you can't set a derived limit through the playground, and nonsense typed into a field leaves the previous value in place.

## The decisions that shaped it

Twenty-six decision records stand behind the design, and twelve of them explain most of the picture above. Open any note below for the short version, or follow its link for the full record.

<details markdown="1">
<summary>Federation, and not one service</summary>

Choosing [federation over a single service](../decisions/AD-0001-federation-over-single-service.md) put three independently owned subgraphs behind one router, where the alternative was one GraphQL service with everything behind it. The rest of the argument rests on this. A single service is the central integration component every team has to change together, which is what federation exists to avoid.

The service bus tradition lost on the audience. A canonical data model needs an owner, and organisations with fewer than 25 engineers don't employ one. The merged schema is computed rather than written, so nobody owns it, and keys are the only thing anyone has to agree on.

</details>

<details markdown="1">
<summary>Cosmo as the platform</summary>

[Cosmo](../decisions/AD-0002-cosmo-as-platform.md) is the platform, pinned to router 0.343.1 and wgc 0.130.1, and the two are bumped together. Everything I checked is Apache-2.0, and the router runs from a static configuration with no control plane and no graph token. The main alternative lost on licence alone, since the OSI doesn't recognise its licence as open source.

The vendor's own caveat comes with the choice. On the compose page it says "it is recommended to not use this for production", and the router logs "Not recommended for Production" at every start without a token. The demo has no control plane to fetch from, so the static path is the only one. My answer to the caveat is to commit the configuration and test it for drift.

</details>

<details markdown="1">
<summary>One binary, one process tree, one port</summary>

[One binary, one process tree, one port](../decisions/AD-0011-one-binary-one-port.md) beat Docker Compose, which the vendor's own demos use and which shows every subgraph as a visibly separate service. With Compose, "one command" would become `docker compose up` after fetching a file, and the routing URLs would need a second configuration.

Two published ports, with the router beside the apps, lost on the launch line and on the cross-origin rules that a single origin avoids. s6-overlay and supervisord need a userland the distroless base doesn't have.

The cost is that a crash anywhere takes the whole demo down. Nothing on port 8080 shows where the router's responsibility ends and the UI server's begins, either.

</details>

<details markdown="1">
<summary>The router as a child process</summary>

The router runs as a [child process from the copied binary](../decisions/AD-0010-router-as-child-process.md) and not as a Go library. Embedding it is possible in principle, and it lost on footing. The module has only commit pseudo-versions and needs a block of `replace` directives. It broke its API at 0.188.0 and 0.278.0, has a rewrite announced, and offers no stability promise. It would also pull message broker clients into a dependency graph that never uses them.

The router's own extension mechanism would have let it supervise the subgraphs itself, but that route doesn't support subscriptions.

</details>

<details markdown="1">
<summary>The rollup as a maximum flow</summary>

For the rollup I chose [maximum flow with the source-side minimum cut](../decisions/AD-0007-rollup-as-maximum-flow.md). Each server becomes an in-node and an out-node joined by an edge of its throughput, and each connection becomes an unlimited edge. The capacity is the maximum flow from the servers nobody feeds to the servers that feed nobody.

The bottleneck is a minimum cut, the cheapest set of servers whose removal would sever every path from entry to exit. Several minimum cuts can exist, so the reported one is defined as the source-side canonical cut, which is the same for every maximum flow.

Rollup in the adapter would have turned a projection into an expression evaluator. Rollup in the apps would have meant two copies of the logic, and no single answer that both apps read through the router. The simpler arithmetic, minimum over serial stages and sum over parallel ones, survives in two places. It's the explanation a reader can check by hand, and it's a differential test that checks the flow against it.

</details>

<details markdown="1">
<summary>A curated projection</summary>

The projection is a [curated set of plainly typed GraphQL types](../decisions/AD-0005-curated-generic-projection.md), chosen by what the downstream services need, and not a few hundred types generated from KerML. Generated types would be correct and complete, and would hand every consumer the abstract syntax the projection exists to spare them.

A generic type for elements nobody has projected yet would be the escape hatch, and I haven't built it. A construct outside the subset is refused instead, which is the smaller first cut, and the record's status says the question is open.

</details>

<details markdown="1">
<summary>Events that carry a version number</summary>

Polling was the obvious alternative, and it lost because the delay before anything shows up equals the poll interval. Subscriptions carry a [version number and nothing else](../decisions/AD-0014-version-events.md), and each app refetches its whole query when one arrives.

A payload carrying verdicts would need the router to resolve another subgraph's entity fields inside a subscription response, and nobody had checked that it could. A version event sidesteps the question, because the refetch is an ordinary query. Adding a broker with Cosmo Streams would have meant another process.

To be honest about the cost, a refetch means a full query per event per open tab. That suits a demo, and it wouldn't suit a fleet of tabs.

</details>

<details markdown="1">
<summary>Composition, committed and tested for drift</summary>

Composition is a [maintainer step whose output is committed](../decisions/AD-0012-composition-committed.md) and copied into the image. A Go test guards it by parsing the configuration and comparing each embedded schema with the schema file it came from.

Composing at container start died when the Go composition library was removed from Cosmo. A Node stage in the Dockerfile would put a tool that declares no Node range into every build. So CI stays Go only, and a schema change without a recompose fails on the pull request.

The drift test is the whole evidence for the third condition a projection has to meet, from Part 1. The contract between producer and consumer is checked mechanically before deployment, instead of being discovered in production.

</details>

<details markdown="1">
<summary>Short names as keys</summary>

Where two subgraphs have to agree they're describing the same requirement, the key is a [SysML short name](../decisions/AD-0018-short-names-as-keys.md), with the qualified name as fallback where an element declares none.

The API-level `elementId` is a UUID that the tool assigns, and no tool stands behind the adapter. A key nobody writes down can't be quoted in the document service's configuration or typed into the playground either. Renaming a short name breaks every service that stores it, and the demo doesn't handle renames.

</details>

<details markdown="1">
<summary>Files instead of a model repository, and editing as scaffolding</summary>

The adapter [reads files](../decisions/AD-0003-adapter-reads-files.md), where a real deployment would front a SysML v2 repository over the standard API. Standing one up means a JVM, a database and a second runtime, inside an image whose whole promise is one command. A counter stands in for commits, and every edit is lost when the container stops.

[Editing through the projection](../decisions/AD-0004-editing-as-scaffolding.md) contradicts my own position that a projection should be read-only, and the record says so. It's scaffolding, needed because nothing would move without it, and a real deployment writes through the API instead. Rewriting the file on disk lost, because persistence is a non-goal and the launch line uses `--rm` in any case.

</details>

<details markdown="1">
<summary>A hand-written parser for a strict subset</summary>

The parser is [hand-written in Go and reads a strict subset](../decisions/AD-0015-hand-written-subset-parser.md) of the language. Three Go parsers for the notation exist, and none of them can be used here. One keeps its code under `internal/`. One was too new to depend on and carries a grammar descended from a much older release. One is GPL-3.0.

A tree-sitter binding needs <span class="term" data-term="cgo">cgo</span>, which would end the `CGO_ENABLED=0` cross-compilation the image depends on. The OMG JSON serialisation lost because it needs Java 17 at build time and hands back the flat metamodel the adapter exists to hide.

The record carries its own condition for replacement. When OpenSysML exposes a public package or freezes its gRPC contract, the hand-written parser is the part of the adapter to retire first.

</details>

> [!WARNING]
> **One assumption left for the build**
>
> One thing in this design hadn't been exercised when I wrote the architecture. The capacity service's schema depends on `@requires` over nested lists of objects, and nobody knew yet whether Cosmo composition and gqlgen, the Go library that generates the services' GraphQL code, would accept that. I left it to the first spike of the build, with a flat JSON scalar on the part as the fallback. [Five spikes before the first line](09-five-spikes-before-the-first-line.md) tells how that went.

## Where to go next

The capacity model in full, with its assumptions, edge cases and the requirements it drove, is in [From use cases to requirements](05-from-use-cases-to-requirements.md). All five architecture views and all twenty-six decisions, including the fourteen I haven't touched here, are in [Five views and twenty-six decisions](06-five-views-and-twenty-six-decisions.md). What the build produced, set against the design described here, is [The demo as it shipped](10-the-demo-as-it-shipped.md).

---

Previous: [Why federate a systems model](00-why-federate-a-systems-model.md) · Index: [Federating a systems model](../README.md) · Next: [How the design was run](02-how-the-design-was-run.md)

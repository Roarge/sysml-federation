# The architecture in one sitting

*Roar Georgsen, 27 August 2026*

Part 2 of 13 in [Federating a systems model](../README.md), written for release v0.3.0.

> [!IMPORTANT]
> **The story so far**
>
> Part 1 argued that a systems model should publish a plain projection of itself, and that federation lets other services attach their own data to it, joined on an entity key. This part opens the demo I built to show that. A SysML v2 model of a five-server query pipeline goes out through a federated GraphQL router. A capacity analysis and a requirements document attach their own data to its parts and requirements, and neither of them has ever read a line of SysML. All of it runs in one container.

## What's in the box

One command starts it. The image is public, built for both amd64 and arm64, and needs no registry login.

> [!TIP]
> **Run it yourself**
>
> ```
> docker run --rm -p 8080:8080 ghcr.io/roarge/sysml-federation
> ```
>
> Then open `http://localhost:8080` in a browser. I build and test on Ubuntu under WSL, and [What shipped, and what did not](11-what-shipped-and-what-did-not.md) says what else it has been tried on.

Inside, one Go binary runs as PID 1, the first process in the container, and supervises everything else. Three subgraphs run as <span class="term" data-term="goroutine">goroutines</span> of that binary, on <span class="term" data-term="loopback">loopback</span> ports 3011, 3012 and 3013. Each is a GraphQL service that owns one slice of the merged schema. The Cosmo router holds the merged schema of all three and plans every incoming query across them. It runs as a child process on loopback port 3002, from a binary copied out of the vendor's image.

The only thing the outside world can reach is a UI server on port 8080. It serves the two web apps at `/viewer` and `/document`, and passes `/graphql` and the `/playground` query editor through to the router. The router itself is never addressable on its own, so everything a browser does goes through one port and one origin.

![One container, one process tree](../img/v4-container.png)

*One container, one process tree. Cut from the [architecture views](../architecture/architecture-views.pdf).*

The two web apps on that port are what a visitor meets, and neither of them works anything out for itself. The wiring, the arithmetic and the document ordering all arrive from services that share no code.

![The model viewer in the shipped state, with parse outlined red and the pipeline requirement failing beneath it](../img/app-viewer-shipped.png)

*The model viewer. The sketch is drawn from the model's own connect statements, parse holds the pipeline to 1200, and PIPE-R1 fails against its limit of 1500.*

![The same requirements as a numbered document, with five derived requirements nested under the first](../img/app-document-tree.png)

*The requirements document. Its numbering, headings and prose belong to a service that computes nothing, and the verdict on each row comes from one that has never read a model file.*

## Three services and what each owns

![Two apps, one graph, three services that never meet](../img/overview-sketch.png)

*Two apps, one graph, three services that never meet. Cut from the [use case storyboard](../stories/use-cases.pdf).*

**The adapter owns the model.** It publishes the parts with their attributes and connections, the requirements with their text, limits and derivation links, the verification cases, and the model's own text and version number.

The adapter is generic. Nothing in its schema or its code names the example, and its tests include a second model with other names and other wiring. Its schema carries three <span class="term" data-term="mutation">mutations</span>, and they are the whole of the model's edit path:

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

So the adoption contract, the whole agreement between the services, fits in one sentence. It is the entity key (a SysML short name), the `@requires` field set the capacity service declares, and that service's two configured names. An organisation adopting the approach keeps the adapter, the compose step that merges the schemas, and the supervisor. It replaces the model file, the two example services, the two apps and the shipped document tree. If two services define the same field, or a `@requires` names a field the adapter no longer projects, composing the schemas fails. A schema changed without recomposing fails a drift test on the pull request.

## One query, three fetches

Composition merges the three schemas into one. In it, a requirement carries its text and limit from the adapter, its verdict and reason from the capacity service, and its document number and inclusion from the document service. The query a visitor runs in the playground asks for fields from all three, two of them from the capacity service.

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
6. The capacity service works out the pipeline's capacity from the wiring, and finds the cut, the cheapest set of servers that limits it. [From use cases to requirements](05-from-use-cases-to-requirements.md) explains how. It answers capacity 1400, bottleneck `indexA` and `indexB`, verdict FAIL, and the reason "capacity 1400 against 1500, limited by indexA, indexB".
7. Both apps render. The viewer moves its red outline from `parse` to the index pair, and the document flips `PIPE-R1.2` to PASS and rewrites the reason under `PIPE-R1`.

A requirement bounds the whole thing. Once the version changes, each app has to show the change within two seconds and without a page reload.

Neither app renders from the mutation's response. The mutation goes out, and the event that follows triggers the same refetch as any other event, so there's only one rendering path. Neither app computes anything either. The viewer's red block, its capacity figure and its highlighted bottleneck all arrive through the router, and so does the document's reason text.

Raise `ingest` to 3000 instead and the same seven steps run. The cut is still `parse`, every number and verdict stays the same, and the only visible change is the figure on `ingest`. The event still fires and both apps still refetch. Nowhere does the demo detect that nothing changed, and it doesn't try to.

Reset, from either app, sends one mutation document carrying two root fields, `resetModel` and `resetDocument`. The router plans the two fields to the two services. Each restores its shipped state, bumps its own counter and emits its own event. The apps have to ask for both because neither service knows the other exists.

## Refusing what it can't serve

The adapter is a pipeline of four packages, `syntax`, `model`, `projection` and `serve`. Every node in its syntax tree carries its byte range in the source, which is what lets an edit replace one literal in place and rebuild the served text and the projection together.

The part most worth having is the refusal path, which works in both directions. The parser reads a strict subset of SysML v2. At start-up, a construct outside that subset stops the load with a file, line and column, instead of being served generically. At run time, an edit to a value bound by an expression, or a negative or non-finite number, is refused with an error that says why, and no event goes out. So you can't set a derived limit through the playground, and nonsense typed into a field leaves the previous value in place.

## The decisions behind it

Twenty-six decision records stand behind this design, and [Five views and twenty-six decisions](06-five-views-and-twenty-six-decisions.md) walks through them. The ones that shape what you've just read are [federation over a single service](../decisions/AD-0001-federation-over-single-service.md), [Cosmo as the platform](../decisions/AD-0002-cosmo-as-platform.md), [one binary, one process tree, one port](../decisions/AD-0011-one-binary-one-port.md), [maximum flow for the rollup](../decisions/AD-0007-rollup-as-maximum-flow.md) and [events that carry only a version number](../decisions/AD-0014-version-events.md).

One thing in this design hadn't been tried when I wrote it: whether `@requires` over nested lists of objects would survive composition and the router. That became the first spike of the build, and [Five spikes before the first line](09-five-spikes-before-the-first-line.md) tells how it went.

---

Previous: [Why federate a systems model](00-why-federate-a-systems-model.md) · Index: [Federating a systems model](../README.md) · Next: [How the design was run](02-how-the-design-was-run.md)

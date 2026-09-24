# What the research overturned

*Roar Georgsen, 27 August 2026*

Part 4 of 13 in [Federating a systems model](../README.md), written for release v0.3.0.

> [!IMPORTANT]
> **The story so far**
>
> The demo serves a SysML v2 model of a query pipeline as a GraphQL subgraph, one service's own slice of a larger schema. A capacity service and a requirements document service sit beside it. The Cosmo router joins the three slices into one graph and answers queries, and two browser apps sit in front, all from one image and one `docker run`. Part 3 showed how I ran the design in gates. This part covers the research the design phase began with, when I wrote down what I believed about my tools and tried to knock each belief over.

## Beliefs, written down

Before any of the demo was designed, the plan rested on beliefs about its tools. Some came from memory and some from vendor pages I'd read months earlier. The design phase began by writing them down and trying to prove them wrong.

The research covered six topics:

- A3 architecture overviews
- WunderGraph Cosmo
- SysML v2 syntax and the state of its parsers
- browser techniques for the two apps
- packaging for the GitHub Container Registry
- requirements practice

Each produced a report of options, open questions and facts. Every fact carried a confidence (verified, likely or unverified) and a source URL with a version. Two of the findings below concern repositories that had changed in the fortnight before I read them, so both describe the state they were in when the report was written.

Next, I checked each load-bearing claim against its primary sources, with the aim of knocking it over. A final pass read the six reports against each other and asked what none of them had examined.

## Seven claims, four refuted

| The claim | What the checking found |
|---|---|
| No maintained Go parser for SysML v2 exists | Refuted |
| The A3 method puts model views on the front and text on the back | Refuted |
| The router can be embedded as a Go library | Confirmed, then narrowed until it was useless here |
| The router serves subscriptions without a message broker | Confirmed |
| Requirements can hold constraints, and derivation has its own keyword | Half right, recorded as refuted |
| The router runs offline with no outbound telemetry | Refuted, as it ships |
| The router, `wgc` and the composition library are Apache 2.0, with no source-available exception | Confirmed |

### No maintained Go parser exists

The claim was that no maintained Go parser for the SysML v2 textual notation exists on GitHub or pkg.go.dev. It was wrong.

Open-MBEE/OpenSysML, under Apache-2.0, was a few weeks old when I read it. It had shipped ten releases in the fortnight before, and it reports that all 95 bundled library files parse cleanly: the 94 official standard library files plus one non-normative extension. All its library code sits under `internal/`, though, so nothing can be imported from it.

The one importable pure-Go library on pkg.go.dev, dVoo/gosysml2, is GPL-3.0 and can't be used here whatever its state. The completeness pass added a third, mycr0ft/gosysml, under MIT, but it was new when I read it and its grammar descends from a 2023 release.

So the corrected statement is narrower than the claim it replaces. No importable, permissively licensed, pure-Go parser is indexed on pkg.go.dev, and the one that exists off it is too young to depend on.

The decision didn't change, but its justification did. The [hand-written strict subset parser](../decisions/AD-0015-hand-written-subset-parser.md) stands, with OpenSysML watched as the natural replacement if it ever exposes a public package. OpenSysML also became one of the two tools that validate the example model against the [SysML 2.0 target](../decisions/AD-0019-sysml-2-0-target.md).

### A3 prescribes a front and a back

> [!NOTE]
> **A3 architecture overview**
>
> A way of summarising one aspect of a system's architecture on a single two-sided sheet of A3 paper, from Borches and Bonnema, developed at Philips Healthcare. One side carries models of the system, such as its functions, the numbers that matter and its physical layout. The other side carries a short text summary.

The claim was that Borches' A3 Architecture Overview method prescribes a specific layout, with model views on the front and text on the back. That was wrong as well.

The [INCOSE 2010 paper](https://web.mst.edu/~lib-circ/files/Special%20Collections/INCOSE2010/A3%20Architecture%20Overviews.pdf) by Borches and Bonnema, and the thesis abstract, say "one side" and "the other side", and never front or back. The one document Borches wrote that uses those words is the 2009 [cookbook](https://www.gaudisite.nl/BorchesCookbookA3architectureOverview.pdf). Its template labels the text sheet "FRONT (Summary)" and the model sheet "BACK (Model)", which is the reverse of the claim.

Borches doesn't prescribe, either. The cookbook says its guidelines "are not fix" as long as the structure is kept. It offers the placements (functional view on the left, quantification top right, physical view bottom right) as recommendations.

The A3 topic report had written "front = model, back = summary" into all four planned sheet specifications. I struck that wording before it reached public text. The sheets keep two sides and the cookbook's placements, and call them a model side and a summary side.

### The router embeds as a Go library

The claim was that the Cosmo router can be embedded as a Go library inside a custom binary, in a supported way. That turned out true, and then it was narrowed until it was worth nothing to this project.

The supported shape is the custom-module form. You depend on the router module, register modules from `init()`, and call `routercmd.Main()`. Constructing the router directly with `core.NewRouter` is exported, but no documentation page describes it. The rest of the evidence went against it.

<details markdown="1">
<summary>Under the bonnet: why embedding was a bad bet</summary>

- The module's release tags have the form `router@0.343.1`, which Go tooling can't resolve, so only commit pseudo-versions exist.
- The official examples repository requires a block of OpenTelemetry `replace` directives, and says compatibility isn't guaranteed without them.
- No Go API stability promise is published anywhere. The module API broke at 0.188.0 and 0.278.0.
- The docs announce that "The new module system will be available in the next major release of the router".
- `WithModulesConfig` takes `map[string]interface{}`, which the repository forbids in [hand-written code, generated files excepted](../decisions/AD-0016-generated-code-exempt.md).

</details>

So the router runs as a [child process from the binary copied out of the official image](../decisions/AD-0010-router-as-child-process.md). It's pinned by tag and driven by its environment and a committed execution configuration, so its Go API can change without touching this repository.

### Subscriptions without a broker

The claim was that the router serves subscriptions from subgraphs over <span class="term" data-term="server-sent-events">SSE</span> or <span class="term" data-term="websocket">WebSocket</span>, and to browsers over SSE or WebSocket, without Kafka or NATS. That one held.

Per subgraph, the transport is `ws`, `sse` or `sse_post`, set in the compose input. Towards the browser the router accepts WebSocket, multipart, and SSE on an ordinary GraphQL POST carrying `Accept: text/event-stream`. Brokers belong only to Cosmo Streams, which is an alternative design and not a prerequisite.

One detail stayed loose. The challenge marked SSE over GET, which a native `EventSource` would need, as unverified, while the browser report had found a router test that exercised it. The design uses `fetch` with a `ReadableStream` reader, the path the vendor documents.

Live push survived as the mechanism behind the two apps. Subscriptions carry [version events](../decisions/AD-0014-version-events.md), and clients refetch instead of taking payloads.

### The `derive` keyword

The claim had two halves. A requirement usage can declare `require constraint { ... }` over its attributes, and derivation between requirements is written with a specific keyword. Half of it was right, and I recorded it as refuted.

The constraint half holds. I checked it against the SysML 2.0 language specification, whose grammar gives a requirement usage the same body as a requirement definition.

The keyword half is false. SysML v2 has no `derive` keyword, and `deriveReqt` occurs nowhere in Part 1 of the specification. <span class="term" data-term="derivation">Derivation</span> lives in the Requirement Derivation domain library, clause 9.6, as semantic metadata applied with the generic `#` prefix. After importing `RequirementDerivation::*`, the form is:

```
#derivation connection {
    end #original ::> globalThroughput;
    end #derive ::> parseThroughput;
}
```

It has one `#original` end and one or more `#derive` ends, and every end is a requirement usage, as in the OMG's own Annex A vehicle model.

The example model's per-server requirements are derived through that library connection, and the construct is inside the parser's subset. The browser report had advised taking the viewer's keyword table from the specification rather than from memory. Now that advice had a concrete reason.

### Runs offline with no outbound telemetry

The claim was that the router runs entirely offline from a static execution configuration produced by `wgc router compose`, with no <span class="term" data-term="control-plane">control plane</span>, no graph token and no outbound telemetry. As the vendor ships it, that's false.

The offline part holds. `wgc router compose` "does not interact with the control plane and completely runs locally". `Start()` takes a static-config branch that never polls, and with an empty token the router disables the Cosmo Cloud exporters and self-registration.

The telemetry part doesn't hold. Since router 0.215.0 the binary has contained an anonymous usage tracker. It's on by default, and it depends on neither the token nor the static configuration. Only `DO_NOT_TRACK=1` or `COSMO_TELEMETRY_DISABLED=true` switches it off, and the vendor's own `router/.env.example` is the one place that names them. There's no YAML key, and no docs page mentions it. The `wgc` command-line tool, `router compose` included, sends the same events unless the same variables are set.

The vendor also documents `wgc router compose` as "recommended to not use this for production", and the router logs "No graph token provided ... Not recommended for Production." at every start.

So telemetry is turned off by [environment variables baked into the image](../decisions/AD-0013-telemetry-off.md). I had claimed the router runs from a pre-built configuration file with no network dependency at all, and that holds only once those two variables are set. The claim now names them, and the vendor's production caveat goes into the public text.

At this stage, whether the router then makes zero outbound connections was inferred from code paths and hadn't been run under `docker run --network none`. It stood as likely, not verified. One of the spikes ran it later, and [Five spikes before the first line](09-five-spikes-before-the-first-line.md) has the result.

### Apache 2.0 throughout

The claim was that the router, `wgc` and the composition library are all Apache 2.0, with no source-available exception. That held. At the root of the monorepo sits one LICENSE file, with no licence per directory. The package metadata of router 0.343.1, `wgc` 0.130.1 and `@wundergraph/composition` 0.63.3 all declare Apache-2.0. The vendor's own pages say you can run it in your own infrastructure at any scale "without license fees or feature gates".

I couldn't run a code-level search for licence gating, so that absence rests on the vendor's word. Apache 2.0 is irrevocable for released versions and silent about future ones, hence the pin.

The side finding turned out the useful one. The Go composition library, `composition-go`, was removed from the repository on 6 May 2026 in PR #2830, whose note said it was "the last place where it was used".

## What the completeness reading asked

In the cross-document pass I listed missing facts, contradictions between reports, unverified claims the design leaned on, and open questions. One missing fact decided the shape of the schema.

The capacity service computes the rollup, the throughput the pipeline can sustain given each server's capacity and the wiring between them. It also returns a verdict, the pass or fail for a requirement against that number. The whole argument rests on none of the three services knowing about the other two. The gap was how the capacity service gets the wiring and the leaf numbers without knowing the adapter.

In GraphQL federation, a service uses `@requires` to declare which fields of an entity must be fetched from elsewhere before it can resolve its own. An entity is a type the router joins across services by its entity key, the field that identifies the same element everywhere. But a `@requires` selection is a field set of fixed depth, and you can't write a recursively nested serial and parallel tree as one. That left two ways out. Either the adapter's <span class="term" data-term="projection">projection</span> flattens the stages, each carrying a parent id and a kind, or the capacity service queries the composed graph as a client of the router, and stops being a member of the federation. No topic report had looked at this, and it decides whether the claim survives.

The plan's answer is in [the article on the use cases](04-twelve-use-cases-and-one-moving-bottleneck.md), and the record for [maximum flow](../decisions/AD-0007-rollup-as-maximum-flow.md) holds the decision. Maximum flow is the largest rate the wiring can carry from end to end. The rollup runs on read and keeps no state, through `@requires`, so the capacity service holds no copy of the model and can't go stale.

The one integration risk it takes is whether Cosmo composition and gqlgen, the Go GraphQL server library, accept `@requires` over nested lists of objects. That became the first implementation spike, with a JSON scalar as the fallback, and [the article on the spikes](09-five-spikes-before-the-first-line.md) covers it.

The contradictions list was long. Four of its entries mattered, and two of the four, the telemetry tracker and the A3 layout, are settled above.

The packaging report recommended `composition-go` for composing at build time or at start-up. By then, the licence challenge had found it removed on 6 May 2026. So both the "compose at container start" and the "Go build step" options were dead. Composition became a maintainer step that needs Node, with its output committed and guarded by a [drift test](../decisions/AD-0012-composition-committed.md).

The SysML report's minimal example encoded the rollup and the verdict inside the model, with `sum(branches.capacity)`, a `->reduce min` and a `require constraint { stage.capacity >= requiredRate }`. My own claim put the arithmetic in an analysis service that "has never parsed a model file". That's two owners for the same semantics, and the adapter would have had to parse expressions it never evaluates. The plan settled it. The model declares `capacity` without a value and states the requirement as a constraint over it, and the arithmetic lives in the capacity service alone.

## Facts that shaped the stack

Some findings refuted nothing and still fixed decisions.

**The router image.** The official router image is built on `gcr.io/distroless/static-debian13` from a static `CGO_ENABLED=0` build, and published for linux/amd64 and linux/arm64. Measured from the registry manifest, its compressed layers total 40.8 MB for amd64 and 37.8 MB for arm64, nearly all of it the binary. That number set the floor for the demo's image against an 80 MB budget, and made copying the binary with `COPY --from` the obvious route.

<span class="term" data-term="distroless">Distroless</span> static has no shell, so a `HEALTHCHECK` has nothing to run unless the binary carries its own probe, and the router's doesn't. So the demo's [one binary](../decisions/AD-0011-one-binary-one-port.md) has a `healthcheck` subcommand that asks the router's `/health/ready` on its behalf.

![The image, layer by layer, base at the bottom](../img/v4-image-layers.png)

*The image, layer by layer, base at the bottom. Cut from the [architecture views](../architecture/architecture-views.pdf).*

**The registry.** The GitHub Container Registry has one trap. A package first published under a personal account is private by default, even when a workflow publishes it from a public repository. The launch line then fails for everyone else until someone flips the visibility by hand, and GitHub warns that the change is one way: "Once you make a package public, you cannot make it private again."

Cross-compilation covers both processor architectures without <span class="term" data-term="qemu">QEMU</span>, because every process in the image is Go. Whether `COPY --from` picks the target platform's variant of the router image in a multi-platform build wasn't confirmed, and went on the spike list. Hence the [tag-triggered publish workflow](../decisions/AD-0020-publish-on-tags.md).

**The viewer's syntax colouring.** No browser-side renderer for SysML v2 text exists that can be vendored as one static file under a permissive licence.

<details markdown="1">
<summary>Under the bonnet: the renderers that didn't fit</summary>

- Prism's 297 languages contain no SysML grammar.
- The OMG pilot implementation's visualisation is a Java PlantUML bundle that needs GraphViz.
- Syside Editor is free but closed source, with a licence server check.
- The tree-sitter grammars would need about 355 KB of JavaScript and WASM, plus a build step.

</details>

A regex tokeniser of roughly 40 lines, using a sticky-flag alternation, is enough for a viewer that shows one small model, so that's what the viewer has.

**Drag and drop.** SortableJS 1.15.7 (MIT, no dependencies, 45 KB minified) is the smallest single file I could vendor whose official nested demo does what the document needs. Together with the tokeniser, this fixed the [web app decision](../decisions/AD-0017-vanilla-web-apps.md): ES modules with no build step, served from the binary's embedded files, plus one vendored MIT file.

**The shape of the requirements.** The requirements practice report settled this. <span class="term" data-term="ears">EARS</span>, the Easy Approach to Requirements Syntax published in 2009, has five patterns with fixed keyword templates: ubiquitous, event-driven, state-driven, unwanted behaviour and optional feature. It also has a complex form that combines two of them. It's the right form for the demo's functional requirements.

It's the wrong form for the rollup arithmetic, which Mavin's own guidance says belongs in a formula or a table. It's wrong for architectural constraints too, which stay as plain "shall" statements. ISO/IEC/IEEE 29148:2018 defines a traceability matrix as an artefact that links requirements to higher-level needs and lower-level implementation. So a two-column Markdown table per hop is what the standard describes.

"Requirement" also means two things in this repository: the demo's requirements on itself, and the requirement elements inside the pipeline model. So the [light requirements scheme](../decisions/AD-0023-light-requirements-scheme.md) gives them different identifier shapes, and the prose always calls the second kind a "model requirement". [The article on the requirements](05-from-use-cases-to-requirements.md) has the requirements themselves.

## What the A3 literature settled

The A3 report read Borches' 2010 thesis record, the Borches and Bonnema <span class="term" data-term="incose">INCOSE</span> 2010 paper, the 2009 cookbook, and the published cases up to 2025. Those run from Thales naval systems and subsea workover, through Mercedes-Benz eDrive (Pesselse and others, 2019) and an IoT tender (Hidle and Kjørstad, 2024), to a Norwegian company of about 70 people (Bergtun and Engen, 2025).

The method came out of Philips Healthcare MRI, where roughly 250 developers spanned five disciplines. Borches estimated 40 to 60 sheets would replace a 200-page system design specification. He found that readers came to meetings having read the A3, when they hadn't read the equivalent document.

The rules that carried into this project are the cookbook's:

- Two sides, a model side and a summary side.
- On the model side:
  - a functional flow of verb-and-noun boxes on the left, with a numbered reading path
  - a visual aid beside it
  - quantification top right, with a formula and colour-coded values
  - a physical view bottom right
  - a legend
- No more than five colours, combined with shading, because "people will not remember more than five colors".
- Fonts of 30 pt for the title, 18 pt for subtitles and 14 pt for the rest, at A3.
- One system aspect per sheet, and "you can't put everything you know about this topic in this A3! So do not try to do it". The cookbook's answer is to link to another sheet instead.
- No formal notation on the sheet, because in Borches' own SysML experiments "most of the meetings with experts was spent discussing the notation itself rather than the content".

That last one is a dry thing to find in the literature behind a project about SysML.

The cases add what the cookbook couldn't know. Pesselse's five sheets went through three to eight versions each. Readers scored "empty spaces" negatively on every sheet, and circled reading numbers were "highly appreciated". Bergtun's interviewees spent 10 to 15 minutes reading one, and over 90 per cent reported better understanding.

Practice settled on a hierarchy of two or three levels: L0 for context, L1 for a technical overview, and L2 for topics. The demo adopted that hierarchy, with four sheets planned, starting with L0 and L2b, the argument and the demonstrator. The cookbook estimates about 20 hours per sheet before review, which is why there are only four.

Two things the literature left open. Nobody has published guidance for producing an A3 as SVG or HTML. And a sheet duplicates whatever document holds the same numbers, which Pesselse's interviewees flagged as a consistency risk. So the [Markdown architecture description](../decisions/AD-0021-architecture-record.md) is the record and the sheets are the overview. How I drew the first sheet, and what a fifteen-minute reader gets from it, is in [the article on the A3 sheet](07-an-a3-sheet-for-a-fifteen-minute-reader.md).

---

Previous: [How the design was run](02-how-the-design-was-run.md) · Index: [Federating a systems model](../README.md) · Next: [Twelve use cases and one moving bottleneck](04-twelve-use-cases-and-one-moving-bottleneck.md)

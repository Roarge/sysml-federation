# What the research overturned

*Roar Georgsen, 27 August 2026*

Part 4 of 11 in [Federating a systems model](../README.md).

The demo takes a SysML v2 model of a query pipeline and serves it as a GraphQL subgraph (one service's own slice of a larger schema) beside a capacity service and a requirements document service, composes the three with the Cosmo router (which joins the slices into one graph and answers queries), and puts two browser apps in front of the result, all from one image and one `docker run`. Before any of that was designed, the plan rested on beliefs about its tools, some from memory, some from vendor pages read months earlier. The design phase began by writing them down and trying to knock them over.

## Six topics

The research covered six topics: A3 architecture overviews, WunderGraph Cosmo, SysML v2 syntax and the state of its parsers, browser techniques for the two apps, packaging for the GitHub Container Registry, and requirements practice. Each produced a report of options, open questions and facts, every fact carrying a confidence (verified, likely or unverified) and a source URL with a version. Two of the findings below concern repositories that had changed in the fortnight before they were read, so both carry the state they were in when the report was written.

Each load-bearing claim was then checked against its primary sources with the aim of knocking it over, and a final pass read the six reports against each other and asked what none of them had examined.

## Seven claims, four refuted

### No maintained Go parser exists

The claim was that no maintained Go parser for the SysML v2 textual notation exists on GitHub or pkg.go.dev. Refuted. Open-MBEE/OpenSysML, Apache-2.0, was a few weeks old, had shipped ten releases in the fortnight before it was read, and reports that all 95 bundled library files parse cleanly, the 94 official standard library files plus one non-normative extension. All its library code sits under `internal/`, so nothing can be imported from it. The one importable pure-Go library on pkg.go.dev, dVoo/gosysml2, is GPL-3.0 and cannot be used here whatever its state. The completeness pass added a third, mycr0ft/gosysml, MIT, but new at the time of reading and with a grammar descended from a 2023 release. The corrected statement is narrower than the claim it replaces: no importable, permissively licensed, pure-Go parser is indexed on pkg.go.dev, and the one that exists off it is too young to depend on.

The decision did not change. Its justification did. The [hand-written strict subset parser](../decisions/AD-0015-hand-written-subset-parser.md) stands, with OpenSysML watched as the natural replacement if it ever exposes a public package, and OpenSysML became one of the two tools that validate the example model against the [SysML 2.0 target](../decisions/AD-0019-sysml-2-0-target.md).

### A3 prescribes a front and a back

The claim was that Borches' A3 Architecture Overview method prescribes a specific layout, model views on the front and text on the back. Refuted. Borches and Bonnema's INCOSE 2010 paper and the thesis abstract say "one side" and "the other side" and never front or back. The one Borches-authored document that uses those words is the 2009 cookbook, whose template labels the text sheet "FRONT (Summary)" and the model sheet "BACK (Model)", the reverse of the claim. Nor does Borches prescribe. The cookbook says its guidelines "are not fix" as long as the structure is kept, and offers the placements (functional view left, quantification top right, physical view bottom right) as recommendations.

The A3 topic report had written "front = model, back = summary" into all four planned sheet specifications. That wording was struck before it reached public text. The sheets keep two sides and the cookbook's placements, and call them a model side and a summary side.

### The router embeds as a Go library

The claim was that the Cosmo router can be embedded as a Go library inside a custom binary in a supported way. Confirmed, and narrowed until it was worth nothing to this project. The supported shape is the custom-module form: depend on the router module, register modules from `init()`, and call `routercmd.Main()`. Constructing the router directly with `core.NewRouter` is exported but described on no documentation page. The module's release tags have the form `router@0.343.1`, which Go tooling cannot resolve, so only commit pseudo-versions exist. The official examples repository requires a block of OpenTelemetry `replace` directives and says compatibility is not guaranteed without them, and no Go API stability promise is published anywhere. The module API broke at 0.188.0 and 0.278.0, the docs announce that "The new module system will be available in the next major release of the router", and `WithModulesConfig` takes `map[string]interface{}`, which the repository forbids in [hand-written code, generated files excepted](../decisions/AD-0016-generated-code-exempt.md).

The router therefore runs as a [child process from the binary copied out of the official image](../decisions/AD-0010-router-as-child-process.md), pinned by tag and driven by environment and a committed execution configuration, so its Go API can change without touching this repository.

### Subscriptions without a broker

The claim was that the router serves subscriptions from subgraphs over SSE or WebSocket, and to browsers over SSE or WebSocket, without Kafka or NATS. Confirmed. Per subgraph the transport is `ws`, `sse` or `sse_post`, set in the compose input. Towards the browser it accepts WebSocket, multipart, and SSE on an ordinary GraphQL POST carrying `Accept: text/event-stream`. Brokers belong only to Cosmo Streams, an alternative design rather than a prerequisite. One detail stayed loose. The challenge marked SSE over GET, which a native `EventSource` would need, as unverified, while the browser report had found a router test exercising it. The design uses `fetch` with a `ReadableStream` reader, the path the vendor documents.

Live push survived as the mechanism behind the two apps, with subscriptions carrying [version events](../decisions/AD-0014-version-events.md) and clients refetching rather than taking payloads.

### The `derive` keyword

The claim was that a requirement usage can declare `require constraint { ... }` over its attributes, and that derivation between requirements is written with a specific keyword. Half right, recorded as refuted. The constraint half is confirmed against the SysML 2.0 language specification, whose grammar gives a requirement usage the same body as a requirement definition. The keyword half is false. SysML v2 has no `derive` keyword, and `deriveReqt` occurs nowhere in Part 1 of the specification. Derivation lives in the Requirement Derivation domain library, clause 9.6, as semantic metadata applied with the generic `#` prefix. After importing `RequirementDerivation::*` the form is

```
#derivation connection {
    end #original ::> globalThroughput;
    end #derive ::> parseThroughput;
}
```

with one `#original` end and one or more `#derive` ends, every end a requirement usage, as in the OMG's own Annex A vehicle model.

The example model's per-server requirements are derived through the library connection, the construct is inside the parser's subset, and the browser report's advice to take the viewer's keyword table from the specification rather than from memory acquired a concrete reason.

### Runs offline with no outbound telemetry

The claim was that the router runs entirely offline from a static execution config produced by `wgc router compose`, with no control plane, no graph token and no outbound telemetry. Refuted by default. The offline part holds. `wgc router compose` "does not interact with the control plane and completely runs locally", `Start()` takes a static-config branch that never polls, and with an empty token the router disables the Cosmo Cloud exporters and self-registration. The telemetry part does not hold. Since router 0.215.0 the binary has contained an anonymous usage tracker that is on by default and gated on neither the token nor the static config. It is switched off only by `DO_NOT_TRACK=1` or `COSMO_TELEMETRY_DISABLED=true`, which the vendor's own `router/.env.example` is the one place to name. There is no YAML key and no docs page mentions it. The `wgc` CLI, `router compose` included, sends the same events unless the same variables are set. The vendor also documents `wgc router compose` as "recommended to not use this for production", and the router logs "No graph token provided ... Not recommended for Production." at every start.

So telemetry is off through [environment variables baked into the image](../decisions/AD-0013-telemetry-off.md). The project had claimed the router runs from a pre-built configuration file with no network dependency at all, and that holds only once those two variables are set. The claim now names them, and the vendor's production caveat goes into the public text. Whether the router then makes zero outbound connections was inferred from code paths and never run under `docker run --network none`, so it stands as likely, not verified.

### Apache 2.0 throughout

The claim was that the router, `wgc` and the composition library are Apache 2.0 with no source-available exception. Confirmed. One LICENSE file at the root of the monorepo, no per-directory licence, and the package metadata of router 0.343.1, `wgc` 0.130.1 and `@wundergraph/composition` 0.63.3 all declare Apache-2.0. The vendor's own pages say you can run it in your own infrastructure at any scale "without license fees or feature gates". A code-level search for licence gating could not be run, so that absence rests on the vendor's word. Apache 2.0 is irrevocable for released versions and silent about future ones, hence the pin.

The side finding was the useful one. The Go composition library, `composition-go`, was removed from the repository on 6 May 2026 in PR #2830, whose note said it was "the last place where it was used".

## What the completeness reading asked

The cross-document pass listed missing facts, contradictions between reports, unverified claims the design leaned on, and open questions. One missing fact decided the shape of the schema.

The capacity service computes the rollup, the throughput the pipeline can sustain given each server's capacity and the wiring between them, and returns a verdict, the pass or fail for a requirement against that number. The whole argument rests on none of the three services knowing the other two. The gap was how the capacity service gets the wiring and the leaf numbers without knowing the adapter. In GraphQL federation a service declares with `@requires` which fields of an entity (a type the router joins across services by its entity key, the field that identifies the same element everywhere) must be fetched from elsewhere before it can resolve its own. But a `@requires` selection is a finite-depth field set, and a recursively nested serial/parallel tree cannot be written as one. Either the adapter's projection (the GraphQL view it serves of the model) flattens the stages, each carrying a parent id and a kind, or the capacity service queries the composed graph as a client of the router rather than a member of the federation. No topic report had examined this, and it decides whether that claim survives.

The plan's answer is described in the [article on the use cases](04-twelve-use-cases-and-one-moving-bottleneck.md) and recorded as [maximum flow](../decisions/AD-0007-rollup-as-maximum-flow.md), the largest rate the wiring can carry end to end. The rollup runs on read, stateless, through `@requires`, so the capacity service holds no copy of the model and cannot be stale, and the one integration risk it takes is whether Cosmo composition and gqlgen, the Go GraphQL server library, accept `@requires` over nested lists of objects. That became the first implementation spike, with a JSON scalar as the fallback, and the [article on the spikes](09-five-spikes-before-the-first-line.md) covers it.

The contradictions list was long. Four of its entries mattered, and two of the four were the telemetry tracker and the A3 layout, both settled above.

The packaging report recommended `composition-go` for build-time or start-up composition. The licence challenge had found it removed on 6 May 2026. Both the "compose at container start" and the "Go build step" options were dead, composition is a maintainer step that needs Node, and its output is committed, guarded by a [drift test](../decisions/AD-0012-composition-committed.md).

The SysML report's minimal example encoded the rollup and the verdict inside the model, with `sum(branches.capacity)`, a `->reduce min` and a `require constraint { stage.capacity >= requiredRate }`. The project's own claim put the arithmetic in an analysis service that "has never parsed a model file". Two owners of the same semantics, and the adapter would have had to parse expressions it never evaluates. The plan settled it: the model declares `capacity` without a value and states the requirement as a constraint over it, and the arithmetic lives in the capacity service alone.

## Facts that shaped the stack

Some findings refuted nothing and still fixed decisions.

The official router image is built on `gcr.io/distroless/static-debian13` from a static `CGO_ENABLED=0` build, published for linux/amd64 and linux/arm64. Measured from the registry manifest, the compressed layers total 40.8 MB for amd64 and 37.8 MB for arm64, nearly all of it the binary. That number set the floor of the demo's image against an 80 MB budget and made copying the binary with `COPY --from` the obvious route. Distroless static has no shell, so a `HEALTHCHECK` has nothing to execute unless the binary carries its own probe, which the router's does not. The demo's [one binary](../decisions/AD-0011-one-binary-one-port.md) therefore has a `healthcheck` subcommand that asks the router's `/health/ready` on its behalf.

![The image, layer by layer, base at the bottom](../img/v4-image-layers.png)

*The image, layer by layer, base at the bottom. Cut from the [architecture views](../architecture/architecture-views.pdf).*

GHCR has one trap. A package first published under a personal account is private by default, even when a workflow publishes it from a public repository, so the launch line fails for everyone else until visibility is flipped by hand, a change GitHub warns is one way: "Once you make a package public, you cannot make it private again." Cross-compilation covers both architectures without QEMU, because every process in the image is Go. Whether `COPY --from` picks the target platform's variant of the router image in a multi-platform build was not confirmed and went on the spike list. Hence the [tag-triggered publish workflow](../decisions/AD-0020-publish-on-tags.md).

No browser-side renderer for SysML v2 text exists that can be vendored as one static file under a permissive licence. Prism's 297 languages contain no SysML grammar, the OMG pilot implementation's visualisation is a Java PlantUML bundle that needs GraphViz, Syside Editor is free but closed source with a licence server check, and the tree-sitter grammars would need about 355 KB of JavaScript and WASM plus a build step. A regex tokeniser of roughly 40 lines using a sticky-flag alternation is enough for a viewer that shows one small model, and that is what it has. For drag and drop, SortableJS 1.15.7 (MIT, zero dependencies, 45 KB minified) is the smallest vendorable single file whose official nested demo does what the document needs. Together these fixed the [web app decision](../decisions/AD-0017-vanilla-web-apps.md): ES modules with no build step, served from the binary's embedded files, and one vendored MIT file.

The requirements practice report settled the shape of the requirements. EARS, the Easy Approach to Requirements Syntax published in 2009, has five patterns with fixed keyword templates, ubiquitous, event-driven, state-driven, unwanted behaviour and optional feature, plus a complex form built by combining two of them, and it is the right form for the demo's functional requirements. It is not the right form for the rollup arithmetic, which Mavin's own guidance says belongs in a formula or table, nor for architectural constraints, which stay as plain shall statements. ISO/IEC/IEEE 29148:2018 defines a traceability matrix as an artefact linking requirements to higher-level needs and lower-level implementation, so a two-column Markdown table per hop is what the standard describes. And because "requirement" means two things in this repository, the demo's requirements on itself and the requirement elements inside the pipeline model, the [light requirements scheme](../decisions/AD-0023-light-requirements-scheme.md) gives them different identifier shapes and always calls the latter "model requirement" in prose. The [article on the requirements](05-from-use-cases-to-requirements.md) has the requirements themselves.

## What the A3 literature settled

The A3 report read Borches' 2010 thesis record and the Borches and Bonnema INCOSE 2010 paper, the 2009 cookbook, and the published cases up to 2025, from Thales naval systems and subsea workover through Mercedes-Benz eDrive (Pesselse and others 2019) and an IoT tender (Hidle and Kjørstad 2024) to a Norwegian company of about 70 people (Bergtun and Engen 2025).

The method came out of Philips Healthcare MRI, where roughly 250 developers spanned five disciplines and Borches estimated 40 to 60 sheets to replace a 200-page system design specification. Borches found readers came to meetings having read the A3 when they had not read the equivalent document.

The rules that carried into this project are the cookbook's. Two sides, a model side and a summary side. On the model side, a functional flow of verb-and-noun boxes on the left with a numbered reading path, a visual aid beside it, quantification top right with a formula and colour-coded values, a physical view bottom right, and a legend. No more than five colours, combined with shading, because "people will not remember more than five colors". Fonts of 30 pt for the title, 18 pt for subtitles and 14 pt for the rest at A3. One system aspect per sheet, and "you can't put everything you know about this topic in this A3! So do not try to do it", which the cookbook answers by linking to another sheet instead. No formal notation on the sheet, because in Borches' own SysML experiments "most of the meetings with experts was spent discussing the notation itself rather than the content". That is a dry thing to find in the literature behind a project about SysML.

The cases add what the cookbook could not know. Pesselse's five sheets went through three to eight versions each, readers scored "empty spaces" negatively on every sheet, and circled reading numbers were "highly appreciated". Bergtun's interviewees spent 10 to 15 minutes reading one and over 90 per cent reported better understanding. Practice converged on a hierarchy of two or three levels, L0 context, L1 technical overview, L2 topics, which is the hierarchy the demo adopted: four sheets planned, starting with L0 and L2b, the argument and the demonstrator. The cookbook's estimate of about 20 hours per sheet before review is why the count is four.

Two things the literature left open. Nobody has published guidance for producing an A3 as SVG or HTML. And a sheet duplicates whatever document holds the same numbers, which Pesselse's interviewees flagged as a consistency risk, so the [Markdown architecture description](../decisions/AD-0021-architecture-record.md) is the record and the sheets are the overview. How the first sheet was drawn, and what a fifteen-minute reader gets from it, is the [article on the A3 sheet](07-an-a3-sheet-for-a-fifteen-minute-reader.md).

---

Previous: [How the design was run](02-how-the-design-was-run.md) · Index: [Federating a systems model](../README.md) · Next: [Twelve use cases and one moving bottleneck](04-twelve-use-cases-and-one-moving-bottleneck.md)

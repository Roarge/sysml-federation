# What the research overturned

*Roar Georgsen, 27 August 2026*

Part 4 of 13 in [Federating a systems model](../README.md), written for release v0.3.0.

> [!IMPORTANT]
> **The story so far**
>
> The demo serves a SysML v2 model of a query pipeline as a GraphQL subgraph, beside a capacity service and a requirements document service. The Cosmo router joins the three into one graph, and two browser apps sit in front. Part 3 showed how I ran the design in gates. This part covers what came before the first gate, when I wrote down what I believed about my tools and tried to knock each belief over.

## Beliefs, written down

Before any of the demo was designed, the plan rested on beliefs about its tools. Some came from memory and some from vendor pages I'd read months earlier. So the design phase began by writing them down and trying to prove them wrong. The research covered six topics, from Cosmo and the state of SysML v2 parsers to packaging and requirements practice, and every fact in it carried a confidence and a source.

Then I took each belief the plan leaned on and checked it against its primary sources, with the aim of breaking it.

## Six claims, three refuted

| The claim | What the checking found |
|---|---|
| No maintained Go parser for SysML v2 exists | Refuted |
| The router can be embedded as a Go library | True, but useless here |
| The router serves subscriptions without a message broker | Confirmed |
| Derivation between requirements has its own keyword | Refuted |
| The router runs offline with no outbound telemetry | Refuted, as it ships |
| The router, its tool and the composition library are Apache 2.0 | Confirmed |

### A Go parser did exist, just not one I could use

I'd believed nobody had written a maintained Go parser for the SysML v2 text notation. Wrong. Open-MBEE/OpenSysML, under Apache 2.0, had shipped ten releases in the fortnight before I read it, and reports that the whole standard library parses cleanly. But all its code sits under Go's `internal/` directory, so nothing in it can be imported. The one importable library on pkg.go.dev is GPL-3.0, and a third, found elsewhere, was too young to depend on.

So the decision stood, and only its reason changed. The adapter has [a hand-written parser for a strict subset](../decisions/AD-0015-hand-written-subset-parser.md) of the language, with OpenSysML watched as its natural replacement if it ever exposes a public package. OpenSysML also became one of the two tools that check the example model is valid SysML.

### The router could be embedded, on paper

The Cosmo router can be embedded as a Go library in a supported way. In practice its module tags its releases in a form Go's tooling can't resolve, and its examples warn that compatibility isn't guaranteed without a block of dependency overrides. It has broken its API twice, promises no stability, and takes a configuration map of the untyped kind this repository forbids. So the router runs as [a child process, from the binary in the official image](../decisions/AD-0010-router-as-child-process.md), and its Go API can change without touching this repository.

### Live updates without a broker

The router serves subscriptions from the subgraphs and passes them to browsers over <span class="term" data-term="server-sent-events">server-sent events</span> or <span class="term" data-term="websocket">WebSocket</span>, with no Kafka or NATS in between. One detail stayed loose, whether a browser's native `EventSource` would work, so the apps use `fetch` with a stream reader instead, the path the vendor documents. Live push survived as the mechanism behind the two apps, with subscriptions carrying [a version number and nothing else](../decisions/AD-0014-version-events.md).

### There is no `derive` keyword

I'd assumed SysML v2 writes derivation between requirements with a keyword of its own. It doesn't. <span class="term" data-term="derivation">Derivation</span> lives in a small library in the standard, applied as metadata with a `#` prefix:

```
#derivation connection {
    end #original ::> globalThroughput;
    end #derive ::> parseThroughput;
}
```

The example model derives its per-server requirements that way, and the parser reads it. A small thing, but exactly the kind of thing that ends up wrong when written from memory.

### The router talks home unless you tell it not to

I'd believed the router runs entirely offline from a pre-built configuration, with no control plane and no telemetry. The offline part holds. The telemetry part doesn't. Since router 0.215.0 the binary has carried an anonymous usage tracker that is on by default, and the only place the vendor names the two variables that switch it off is an example environment file.

So the image [bakes those variables in](../decisions/AD-0013-telemetry-off.md), and every public claim that the demo runs offline names them. At this stage, whether the router then made no outbound connection at all was inferred from reading its code, and not run. One of the spikes ran it later, and [Five spikes before the first line](09-five-spikes-before-the-first-line.md) has the result.

### Apache 2.0, and a Go library that had just vanished

The router, its command-line tool and the composition library are all Apache 2.0, with no source-available exception in the package metadata. I couldn't search the code itself for licence gating, so that part rests on the vendor's word. Apache 2.0 also says nothing about future versions, which is why the versions are pinned ([Cosmo as the platform](../decisions/AD-0002-cosmo-as-platform.md)).

The useful find was on the side. The Go composition library had been removed from the Cosmo repository on 6 May 2026. It decided where composition could happen, as the next section shows.

## What nobody had examined

A final pass read all six reports against each other and asked what none of them had looked at. One missing fact decided the shape of the schema.

The whole argument rests on none of the three services knowing about the others. The gap was how the capacity service would get the wiring and throughputs it needs without knowing the adapter. In federation, a service uses `@requires` to ask the router to fetch fields from another service first. But a `@requires` selection has a fixed depth, and a pipeline of stages nested inside stages doesn't. Either the adapter would flatten the nested stages, or the capacity service would have to query the router as a client and stop being a member of the federation.

The plan's answer was to hand the capacity service one level at a time: a part's children, their attributes and the connections between them, which a fixed-depth `@requires` can carry. From those it computes the rollup on every read, as a [maximum flow](../decisions/AD-0007-rollup-as-maximum-flow.md), so it keeps no copy of the model and can't go stale. Whether Cosmo and the Go GraphQL library would accept `@requires` over nested lists of objects became the first spike of the build.

Two contradictions between the reports mattered as well. The packaging report had recommended composing the schemas inside the image at build or start-up, using the Go composition library the licence check had just found deleted. So composition became a step I run by hand, with its output [committed and guarded by a drift test](../decisions/AD-0012-composition-committed.md). And the SysML report's example had put the rollup arithmetic inside the model, while the whole claim was that the analysis lives in a service that has never parsed a model file. The model now declares `capacity` without a value, and the arithmetic lives in the capacity service alone.

## How the requirements would be written

The research on requirements practice settled their shape. <span class="term" data-term="ears">EARS</span>, a small set of sentence patterns, suits the demo's functional requirements. It doesn't suit the rollup arithmetic, which belongs in a formula, or architectural constraints, which stay plain "shall" statements. Traceability is a plain table per hop, which is what the ISO/IEC/IEEE 29148 standard describes. "Requirement" also means two things here: the demo's requirements on itself, and the requirement elements inside the pipeline model. So the two get different identifier shapes in a [light scheme](../decisions/AD-0023-light-requirements-scheme.md), and the second kind is always called a model requirement. [From use cases to requirements](05-from-use-cases-to-requirements.md) has the requirements themselves.

---

Previous: [How the design was run](02-how-the-design-was-run.md) · Index: [Federating a systems model](../README.md) · Next: [Twelve use cases and one moving bottleneck](04-twelve-use-cases-and-one-moving-bottleneck.md)

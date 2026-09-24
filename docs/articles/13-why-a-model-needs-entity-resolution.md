# Why a model needs entity resolution

*Roar Georgsen, 24 September 2026*

Part 1 of 6 in [Automating traceability](../README.md).

## One stage, four hunts

Take the parse stage from the pipeline in my first series, [Federating a systems model](00-why-federate-a-systems-model.md). In the example model it's `PIPE-S2`, a server with a throughput of 1200, and the model requirement `PIPE-R1.2` says it must sustain 1500. The pipeline was only ever modelled. Suppose it had been built, and a colleague asks which of the running things are parse, and which of them would tell us if `PIPE-R1.2` stopped holding.

In this made-up case, a search of the built system turns up four candidates:

- in the <span class="term" data-term="opentofu">OpenTofu</span> configuration, a resource at `module.pipeline.aws_ecs_service.stage["parse"]`
- in the <span class="term" data-term="opentelemetry">OpenTelemetry</span> traces, a service called `query-parser`
- in <span class="term" data-term="checkly">Checkly</span>, a check called `parse-p95-latency`
- in the test suite, `TestParserRejectsEmptyQuery`

None of them says `PIPE-S2`. So you open each one, compare it with what the model says the stage does, and make a call. The OpenTofu address is easy, since it carries the stage's own name. `query-parser` needs a look at what the service actually does before you'd trust the resemblance. The check measures latency against a target, which puts it closer to the requirement than to the part. And the test could belong to the stage or to the query language, which only its body will settle.

That routine is the whole task, done by hand. Everything in this series is about doing it without the hand, and knowing how far to trust the result.

[![A model element on the left, PIPE-S2 parse, has dashed lines to four objects of an imagined built system on the right. They are an OpenTofu resource whose address contains parse, an OpenTelemetry service named query-parser, a Checkly check named parse-p95-latency and a Go test named TestParserRejectsEmptyQuery. None of the four carries the key PIPE-S2, and each box names the only evidence that connects it.](../figures/four-candidates.svg)](../figures/four-candidates.svg)

*The parse stage and four made-up counterparts. The model holds the key, and the built system holds only descriptions.*

## Entity resolution

> [!NOTE]
> **Entity resolution**
>
> Deciding that two records held in different places describe the same thing. In this series one record is a model element and the other is an object in the built system, such as an OpenTofu resource, a traced service, a check or a test. Resolving them tells you where a trace link belongs.

The problem is old. [Getoor and Machanavajjhala](http://vldb.org/pvldb/vol5/p2018_lisegetoor_vldb2012.pdf) trace its earliest work back to the 1950s and list the other names it goes by, from record linkage and deduplication to reference reconciliation. Software engineering has its own name for the wider job of finding trace links after the fact, traceability link recovery, which [Guo and colleagues](https://arxiv.org/abs/2405.10845) describe as identifying the relations within a set of artefacts that already exist.

The two sides aren't symmetric. On one side, the model holds the keys: short names such as `PIPE-S2`, which I chose, and which the federation in the first series joins on. The built system holds descriptions. Addresses, service names, tags, check names and test names are all written by whoever built that part, for their own purposes. Resolving the parse stage means deciding which descriptions belong to its key.

The first series used a similar-sounding term, and the two shouldn't be confused. A federation's <span class="term" data-term="entity-resolver">entity resolver</span> is handed a key and returns the fields its service adds for that object. It never decides anything, because the services already agree on the key. Entity resolution is the step before that: working out the key for an object that never carried one. Once that's done, the federation can join the two as it joins everything else. That's why I think of automating traceability and automating federation as one job seen from its two ends.

## Where the keys went missing

My own demo has both kinds. Some keys were designed in. The requirement `SR-22`, that an edit patches the model's source text, appears as the short name `SR-22` and as the model's longer name `SR_22_EditsPatchTheSource`. It's also the prefix of three Go tests, such as `TestSR22_SetAttributePatchesTextAndProjectionTogether`. Stories reach Checkly as tags in lower case, so `US-03` arrives there as `us-03`. The test described in part 13 of the first series fails if the model and the Go test names drift apart.

Other keys never existed. The router is `router` in the demo's model and `sysml-federation-router` in its traces, a name the model never mentions. The <span class="term" data-term="adapter">adapter</span> is `adapter` in the model and `model` in the composition file that names the <span class="term" data-term="subgraph">subgraphs</span>, so the traces call it that too. And of the 146 top-level Go test functions in the repository, 69 carry a requirement's key in their names. The other 77 are invisible to anything that joins on keys.

None of this is carelessness. Each tool has its own idea of identity, and none of them has any reason to use mine. OpenTofu tracks a resource by its [address](https://opentofu.org/docs/cli/state/resource-addressing/), which the author chooses and which only means something inside that configuration. Rename it and OpenTofu, by default, reads the change as an [intent to destroy the old object and create a new one](https://opentofu.org/docs/language/modules/develop/refactoring/), unless a `moved` block says otherwise. OpenTelemetry's `service.name` is a [logical name the team picks](https://opentelemetry.io/docs/specs/semconv/resource/service/), and when nobody sets one it falls back to `unknown_service:` followed by the executable's name. Checkly knows a check by its `logicalId`, and [changing that ID](https://www.checklyhq.com/docs/cli/constructs-reference/) tells Checkly one check was removed and another created.

Designing the keys in helps, but it leaks. In a study of open-source projects whose convention was to put an issue key in every commit message, [Rath and colleagues](https://arxiv.org/abs/1804.02433) found only 60% of commits linked, and a single mistyped key was enough to break the join.

## Why a model needs the links

Picture a check failing at three in the morning, in the made-up pipeline. Without a link to the model, the engineer on call knows that a URL got slow. With one, they know `PIPE-R1.2` is at risk, and the model shows them everything else that depends on parse. The link turns an alert into a question about the system.

That's the everyday case. The formal one is certification. Safety standards ask for trace links, and Cleland-Huang and colleagues' [review of the field](http://selab.netlab.uky.edu/homepage/publications/2014-ICSE-FOSE.pdf) quotes DO-178C, for airborne software, on tracing source code to low-level requirements to show there's no undocumented code. The same review found traceability "often conducted in an ad-hoc, after-the-fact manner", at a cost that "can be extremely high". In submissions to the US Food and Drug Administration it found trace data that was incomplete, incorrect and conflicting, with links created at the very end, specifically for certification.

The field's stated goal is the opposite. The [grand challenge](https://arxiv.org/abs/1710.03129) set out by Antoniol, Cleland-Huang, Hayes and Vierhauser is traceability that's "always there, without ever having to think about getting it there". Defence engineering has a name for the connected result, the <span class="term" data-term="digital-thread">digital thread</span>, and the US Department of Defense's [digital engineering strategy](https://ac.cto.mil/wp-content/uploads/2019/06/2018-Digital-Engineering-Strategy_Approved_PrintVersion.pdf) expects its authoritative source of truth to "provide traceability as the system of interest evolves".

The first series argued that a model belongs at the centre of an organisation's engineering. A centre with no links out to what was built is still only a description with more structure. The links are what let the model answer for the system.

## Links go out of date

Much of the thinking about linking records assumes the records sit still. A census return or a published paper is written once. A system model is not, and neither is the system it describes. Stages get split, services renamed, tests moved, requirements reworded, and a link that was right last month can be wrong today without anybody touching it.

The traceability literature calls this decay. Links "become stale when source and/or target artifacts are modified", as the same review puts it, and [Mäder and Gotel](https://europepmc.org/article/MED/23471308) argue that links need maintaining as the system evolves "in order to prevent their decay". The stakes are plain in Rahimi and Cleland-Huang's line, in the grand challenges report: "Outdated trace links invalidate safety-cases."

So entity resolution for traceability can't be a one-off clean-up. It has to run again whenever either side changes, the way a test suite does. The demo's own model has a small version of this already. The test from part 13 runs on every change and fails when a named test or check disappears. What it can't do is find a link nobody wrote down.

## Recommend, or add and repair later

There are two ways to put a resolver to work, and they fail in different directions.

The first works at the moment something is made. While an engineer writes the OpenTofu resource for parse, the resolver suggests `PIPE-S2`, the engineer accepts, and the key goes in as a tag at the source. Every link is confirmed by a person before it exists. The second works over what's already there. The resolver reads the configuration, the traces, the checks and the tests, adds the links it's confident about, and people remove the wrong ones when they find them.

Statisticians settled on three outcomes for this long ago. In [Fellegi and Sunter's model](http://www2.stat.duke.edu/~rcs46/linkage/presentations/01-baiLi_FelleigSunter1969.pdf) of record linkage, each pair is a link, a possible link or a non-link, and the middle band is what a person reviews. How wide that band is decides which of the two ways of working you're really running.

It also decides which mistakes you live with. A wrong link is a false positive, and a missing one a false negative. With a person confirming, the resolver can lean towards finding everything. A wrong suggestion only costs a moment's rejection. [Dekhtyar and Hayes](https://arxiv.org/abs/1807.11454) make the point directly, that "detecting a false positive is much simpler and faster than discovering an error of omission". Adding links unsupervised needs the opposite lean. A wrong link sits there, looking authoritative, until someone trips over it. Rath and colleagues' study shows both settings on one classifier: about 96% recall at 33% precision when recommending, and over 89% precision at 50% recall when adding links on its own.

A person in the loop is usually better, though not always. When [Cuddeback, Dekhtyar and Hayes](https://digitalcommons.calpoly.edu/csse_fac/108/) had analysts vet candidate trace matrices, the analysts pulled them towards the point where recall equals precision. Poor matrices improved and good ones got worse, and the review above sums it up: "the higher the quality of the starting trace matrix, the worse the decisions the analyst makes". Decay cuts across both ways of working, too. A link a person confirmed last year goes stale just as fast as one a machine added.

## When the built thing is physical

Cyber-physical systems make every part of this harder, because the counterpart of a model element is an object on a bench or a wall. The [Asset Administration Shell](https://industrialdigitaltwin.org/wp-content/uploads/2024/06/IDTA-01001-3-0-1_SpecificationAssetAdministrationShell_Part1_Metamodel.pdf), the industrial standard for describing such assets, notes that one asset can have several identifiers, among them a serial number, the manufacturer's part number, the customers' own part numbers and an RFID code. None of those is a model key either. Messages about the asset bring their own naming. <span class="term" data-term="mqtt">MQTT</span> topic names are [case-sensitive and never normalised](https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html), so in a made-up plant a subscriber to `site1/line2/pumpA/#` hears nothing from a pump publishing as `PumpA`. The problem is the same, and there are fewer hints to resolve it with. This series stays mostly with software, where the built system can at least be queried.

## What I'm after

I want a model whose links out are found without anyone typing a key, checked again whenever either side changes, and served beside the model through the same federation the first series built. Later in the series I'll build a small resolver for the demo to test these ideas against.

Coming up next: part 2 takes a single resolution task apart, from what goes in to what comes out, and asks how precisely a task has to be stated before anyone tries to solve it.

---

Previous: [A model of the demo itself](12-a-model-of-the-demo-itself.md), the last part of Federating a systems model · Index: [Automating traceability](../README.md)

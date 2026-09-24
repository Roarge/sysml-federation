# Why a model needs entity resolution

*Roar Georgsen, 24 September 2026*

Part 1 of 6 in [Automating traceability](../README.md).

## Four candidates for one stage

Take the parse stage from the pipeline in my first series, [Federating a systems model](00-why-federate-a-systems-model.md). In the example model it's `PIPE-S2`, a server with a throughput of 1200. The model requirement `PIPE-R1.2` asks it to sustain 1500, and in the model it falls short. The pipeline was only ever modelled. Suppose it had been built, and a colleague sends me a list of what they think belongs to parse in the running system.

In this made-up case the list has four entries:

- in the <span class="term" data-term="opentofu">OpenTofu</span> configuration, a resource at `module.pipeline.aws_ecs_service.stage["parse"]`
- in the <span class="term" data-term="opentelemetry">OpenTelemetry</span> traces, a service called `query-parser`
- in <span class="term" data-term="checkly">Checkly</span>, a check called `parse-sustained-qps`
- in the test suite, `TestParserRejectsEmptyQuery`

None of them says `PIPE-S2`, and each asks for a different kind of judgement. The OpenTofu address carries the stage's own name, which is about as good as evidence gets without a key. `query-parser` resembles it, and I'd want to know what the service actually does before trusting the resemblance. The check measures sustained throughput, so it's evidence about the requirement more than about the part. The test could belong to the stage or to the query language, and only its body will say.

A real system turns one list like this into thousands. Every model element has counterparts in several tools, and every tool renames things on its own schedule. This series is about deciding which of those candidates are the element, without a colleague doing it by hand, and about knowing how far to trust the answer.

[![A model element on the left, PIPE-S2 parse, has dashed lines to four objects of an imagined built system on the right. They are an OpenTofu resource whose address contains parse, an OpenTelemetry service named query-parser, a Checkly check named parse-sustained-qps and a Go test named TestParserRejectsEmptyQuery. None of the four carries the key PIPE-S2, and each box names the only evidence that connects it.](../figures/four-candidates.svg)](../figures/four-candidates.svg)

*The parse stage and four made-up counterparts. The model holds the key, and the built system holds only names that point towards it.*

## Entity resolution

> [!NOTE]
> **Entity resolution**
>
> Deciding that two records held in different places describe the same thing. In this series one record is a model element and the other is an object in the built system, such as an OpenTofu resource, a traced service, a check or a test. Resolving them tells you where a trace link belongs.

The problem is old. [Getoor and Machanavajjhala](http://vldb.org/pvldb/vol5/p2018_lisegetoor_vldb2012.pdf) trace its earliest work back to the 1950s and list the other names it goes by, from record linkage and deduplication to reference reconciliation. Software engineering has its own name for the wider job of finding trace links after the fact. Guo and colleagues [call it trace link recovery](https://arxiv.org/abs/2405.10845), the task of identifying the relations within a set of artefacts that already exist.

The two sides aren't symmetric. On one side, the model holds the keys, short names such as `PIPE-S2` that I chose and that the federation in the first series joins on. On the other, the built system holds names written by whoever built each part, for their own purposes. Addresses, service names, tags, check names and test names all point somewhere, and resolving the parse stage means deciding which of them point at its key.

The first series used a similar-sounding term, and the two shouldn't be confused. A federation's <span class="term" data-term="entity-resolver">entity resolver</span> is handed a key and returns the fields its service adds for that object. It never decides anything, because the services already agree on the key. Entity resolution is the step before that, working out the key for an object that never carried one. Once that's done, the federation can join the two as it joins everything else. That's why I think of automating traceability and automating federation as one job seen from its two ends.

## Where the keys went missing

My own demo has both kinds. Some keys were designed in. The requirement `SR-22`, that an edit patches the model's source text, appears as the short name `SR-22` and as the model's longer name `SR_22_EditsPatchTheSource`. It's also the prefix of six Go tests, such as `TestSR22_SetAttributePatchesTextAndProjectionTogether`. Stories reach Checkly as tags in lower case, so the check for `US-04` carries the tag `us-04`. The test described in part 13 of the first series fails if the model and the Go test names drift apart.

Some bridges I wrote by hand. The model records that the <span class="term" data-term="adapter">adapter</span> is the <span class="term" data-term="subgraph">subgraph</span> called `model`, in an attribute that nothing checks. Other keys never existed at all. The router is `router` in the demo's model and `sysml-federation-router` in the traces from the check session, a name the model never mentions. And of the 146 top-level Go test functions in the repository, 69 carry a requirement's key in their names. The other 77 are invisible to anything that joins on keys.

Each tool has its own idea of identity, and none of them has any reason to use mine. OpenTofu tracks a resource by its [address](https://opentofu.org/docs/cli/state/resource-addressing/), which the author chooses and which only means something inside that configuration. Rename it and OpenTofu, by default, reads the change as an [intent to destroy the old object and create a new one](https://opentofu.org/docs/language/modules/develop/refactoring/), unless a `moved` block says otherwise. OpenTelemetry's `service.name` is the service's [logical name](https://opentelemetry.io/docs/specs/semconv/resource/service/), and when nobody sets one it falls back to `unknown_service:` followed by the executable's name. Checkly knows a check by its `logicalId`, and [changing that ID](https://www.checklyhq.com/docs/cli/constructs-reference/) tells Checkly one check was removed and another created.

Designing the keys in helps, but it leaks. In six open-source projects that largely followed the practice of tagging commits with issue keys, [Rath and colleagues](https://arxiv.org/abs/1804.02433) found that on average only 60% of commits were linked. One mistyped key was enough to break a join.

## Why a model needs the links

Picture a check on the made-up pipeline failing at three in the morning. Without a link to the model, the engineer on call knows that a number fell below a line. With one, they know it's the throughput `PIPE-R1.2` asks of parse, and the model shows them everything else that depends on parse. The link turns an alert into a question about the system.

That's the everyday case. The formal one is certification. Safety standards ask for trace links, and Cleland-Huang and colleagues' [review of the field](http://selab.netlab.uky.edu/homepage/publications/2014-ICSE-FOSE.pdf) quotes DO-178C, for airborne software, on tracing source code to low-level requirements to show there's no undocumented code. The same review found traceability "often conducted in an ad-hoc, after-the-fact manner", at a cost that "can be extremely high". It also reports an analysis of submissions to the US Food and Drug Administration. In many cases the trace data was incomplete, incorrect and conflicting, with clear signs that links had been created at the very end, for certification.

The goal the field set itself points the other way. Antoniol, Cleland-Huang, Hayes and Vierhauser restated it in 2017 as the [grand challenge](https://arxiv.org/abs/1710.03129) of traceability that's "always there, without ever having to think about getting it there". Defence engineering has a name for the connected result, the <span class="term" data-term="digital-thread">digital thread</span>, and the US Department of Defense's [digital engineering strategy](https://ac.cto.mil/wp-content/uploads/2019/06/2018-Digital-Engineering-Strategy_Approved_PrintVersion.pdf) expects its authoritative source of truth to "provide traceability as the system of interest evolves".

None of that comes free. A resolver takes effort to build and tune, and a poor one produces confident mistakes at scale. I think the links earn it anyway. The first series argued that a model belongs at the centre of an organisation's engineering, and a centre that can't point at what was built can't answer for it.

## Links go out of date

Much of the thinking about linking records assumes the records sit still. A census return or a published paper is written once, while a system model and the system it describes change all the time. Stages get split, services renamed, tests moved and requirements reworded. A link that was right last month can be wrong today without anybody touching it.

The traceability literature calls this decay. Links "become stale when source and/or target artifacts are modified", as the same review puts it, and [Mäder and Gotel](https://europepmc.org/article/MED/23471308) argue that links need maintaining as the system evolves "in order to prevent their decay". Rahimi and Cleland-Huang, in the grand challenges report, put the stakes plainly: "outdated trace links invalidate safety-cases".

So entity resolution for traceability has to run again whenever either side changes, the way a test suite does. The demo's own model has a small version of this already. The test from part 13 runs on every change and fails when a named test or check disappears. What it can't do is find a link nobody wrote down.

## Recommend, or add and repair later

A resolver can be wrong in two directions. A wrong link is a false positive, and a missing one is a false negative. Which of the two you'd rather live with decides how the resolver should be put to work.

One way puts a person in front of every link. While an engineer writes the OpenTofu resource for parse, the resolver suggests `PIPE-S2`, the engineer accepts, and the key goes in as a tag at the source. Here the resolver can lean towards finding everything, since a wrong suggestion costs a moment's rejection. Dekhtyar and Hayes [make the same point](https://arxiv.org/abs/1807.11454), that "detecting a false positive is much simpler and faster than discovering an error of omission".

The other way works over what's already there. The resolver reads the configuration, the traces, the checks and the tests, adds the links it's confident about, and people remove the wrong ones when they find them. It's far faster and cheaper, and it's the only way to cover a system too big to review link by link. It needs the opposite lean, though. A wrong link sits there looking authoritative until someone trips over it, so the resolver has to favour precision. Rath and colleagues ran one classifier at two thresholds in their study. On average it reached about 96% recall at 33% precision when recommending, and over 89% precision at 50% recall when adding links on its own.

Record linkage has a name for the ground between the two. In the Fellegi and Sunter model, as [Winkler](https://www.census.gov/content/dam/Census/library/working-papers/2006/adrm/rrs2006-02.pdf) sets it out, two thresholds divide every pair into a link, a non-link and a possible link held for clerical review. How wide that middle band is decides which way of working you're really running.

Review by a person helps most when the candidates are poor. When [Cuddeback, Dekhtyar and Hayes](https://digitalcommons.calpoly.edu/csse_fac/108/) had analysts vet candidate trace matrices, the analysts pulled them towards the point where recall equals precision, and poor matrices improved sharply. Follow-up studies found the other half of that, which the review above sums up: "the higher the quality of the starting trace matrix, the worse the decisions the analyst makes". Decay cuts across both ways of working, too. A link a person confirmed last year goes stale just as fast as one a machine added.

## When the built thing is physical

Cyber-physical systems make every part of this harder, because the counterpart of a model element is an object on a bench or a wall. The [Asset Administration Shell](https://industrialdigitaltwin.org/wp-content/uploads/2024/06/IDTA-01001-3-0-1_SpecificationAssetAdministrationShell_Part1_Metamodel.pdf), an industrial specification for describing such assets, notes that one asset can have several identifiers, among them a serial number, the manufacturer's part number, the customers' own part numbers and an RFID code. None of those is a model key either. Messages about the asset bring their own naming. <span class="term" data-term="mqtt">MQTT</span> topic names are [case-sensitive and never normalised](https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html), so in a made-up plant a subscriber to `site1/line2/pumpA/#` hears nothing from a pump publishing as `PumpA`. The problem is the same, and there are fewer hints to resolve it with. This series stays mostly with software, where the built system can at least be queried.

## What I'm after

I want a model whose links out are found without anyone typing a key, and checked again whenever either side changes. I want those links served beside the model, through the same federation the first series built. Later in the series I'll build a small resolver for the demo to test these ideas against.

Part 2, coming up next, takes a single resolution task apart, from what goes in to what comes out, and asks how precisely a task has to be stated before anyone tries to solve it.

---

Previous: [A model of the demo itself](12-a-model-of-the-demo-itself.md), the last part of Federating a systems model · Index: [Automating traceability](../README.md)

# Why does a systems model need entity resolution?

*Roar Georgsen, 24 September 2026*

Part 1 of 6 in [Automating traceability](../README.md).

A <span class="term" data-term="systems-model">systems model</span> describes a system, usually long before anyone builds it. It names the parts, says what each of them must do and records how they fit together. In <span class="term" data-term="sysml-v2">SysML v2</span> it does all of that in a form a program can read. If you come from software or data, it's a different animal from the models you may know. A data model or a database schema describes the shape of records, and a machine learning model is something trained on data, while a systems model describes the system being engineered. It's rarely the only description of that system, though. While a team builds the system, it writes several more without ever meaning to. The infrastructure code says which servers to start. The running services report their own names in their telemetry, and the checks and tests that keep an eye on the system carry names of their own. All of them describe the same system, and hardly any of them mention the systems model.

The links between the systems model and those other descriptions are called trace links, and keeping them is what engineers mean by <span class="term" data-term="traceability">traceability</span>. With them, you can start from a failing check and find the requirement at risk, or start from a requirement and find everything that shows it's met. Without them, the systems model slowly turns into a nicely drawn picture of a system nobody can point at.

Keeping trace links by hand works for a while. I know, because in an [earlier series](00-why-federate-a-systems-model.md) I built a small demo that serves a SysML v2 systems model to other tools, and I wrote every link between that systems model and the rest of the demo myself. It's a fine way to spend an evening and a poor way to run an engineering organisation. This series is about finding those links automatically, and keeping them right while the systems model and the built system both change underneath them. Later on I'll build a small tool for the demo to try the ideas out. You won't need the earlier series to follow along.

## Try it yourself

Here's a made-up case. Say a team has modelled a pipeline that takes search queries in at one end and serves results at the other. The queries pass through five stages, and the second one, parse, is in the team's systems model under the key `PIPE-S2`. According to the model, parse can handle 1,200 queries per second. A requirement with the key `PIPE-R1.2` asks it to handle 1,500, so on paper it falls short. The pipeline comes from the example systems model in my demo, and it was only ever modelled, so I'm free to imagine the rest.

Now suppose the team has built it, and a colleague sends me a list of what they think belongs to parse in the running system:

- a resource in the <span class="term" data-term="opentofu">OpenTofu</span> code that deploys the pipeline, at `module.pipeline.aws_ecs_service.stage["parse"]`
- a service called `query-parser` in the <span class="term" data-term="opentelemetry">OpenTelemetry</span> traces the system sends while it runs
- a check called `parse-sustained-qps` in <span class="term" data-term="checkly">Checkly</span>, which tests the live system on a schedule
- a Go test called `TestParserRejectsEmptyQuery`

Before reading on, decide for yourself which of these you'd link to `PIPE-S2`.

[![An element of the systems model on the left, PIPE-S2 parse, has dashed lines to four objects of an imagined built system on the right. They are an OpenTofu resource whose address contains parse, an OpenTelemetry service named query-parser, a Checkly check named parse-sustained-qps and a Go test named TestParserRejectsEmptyQuery. None of the four carries the key PIPE-S2, and each box names the only evidence that connects it.](../figures/four-candidates.svg)](../figures/four-candidates.svg)

*The parse stage and four made-up counterparts. The systems model holds the key, and the built system holds only names that point towards it.*

Here's how I'd call it. None of the four says `PIPE-S2`, and each needs a different kind of judgement. The OpenTofu address has the stage's own name in it, which is about as good as evidence gets without a key. `query-parser` looks right, but I'd want to know what the service actually does before trusting a resemblance. The check measures sustained throughput, so it says more about the requirement than about the part, and I'd rather link it to `PIPE-R1.2`. The test could be about the parse stage or about the query language, and only its code will tell.

Each of those calls rested on names chosen by people who weren't thinking about the systems model at all. A real system turns a list like this into thousands of entries. Every element of a systems model has counterparts in several tools, and every tool renames things on its own schedule. Nobody keeps a colleague's list up to date for long.

## Entity resolution

> [!NOTE]
> **Entity resolution**
>
> Deciding that two records held in different places describe the same thing. In this series one record is an element of the systems model, such as a part or a requirement. The other is an object in the built system, such as an OpenTofu resource, a traced service, a check or a test. Resolving them tells you where a trace link belongs.

The problem is far older than software. [Getoor and Machanavajjhala](http://vldb.org/pvldb/vol5/p2018_lisegetoor_vldb2012.pdf) trace the first work on it back to the 1950s. They also list some of the names it has gone by since, from record linkage and deduplication to reference reconciliation. Software engineering has its own name for the wider job of finding trace links after the fact. Guo and colleagues [call it trace link recovery](https://arxiv.org/abs/2405.10845), the task of identifying the relations within a set of artefacts that already exist. I'll use both terms, with "resolve" as the verb.

The two sides of the problem aren't equal. On one side, the systems model holds the keys, short names such as `PIPE-S2` that its author picked so that tools could refer to each element. On the other, the built system holds names that each builder picked for their own purposes. Addresses, service names, tags, check names and test names all point somewhere, and resolving the parse stage means deciding which of them point at `PIPE-S2`.

There's a bridge here to the earlier series. The demo joins several services into one graph, a technique called federation, and in it an <span class="term" data-term="entity-resolver">entity resolver</span> is handed a key and returns whatever its service knows about that object. It never has to decide anything, because every service already agrees on the key. Entity resolution is the step before that, working out the key for an object that never had one. Once that's done, the graph can join the two like anything else, and the links can sit beside the systems model where every tool can reach them. So I've come to think of automating traceability and automating federation as one job seen from its two ends.

## Why the names don't line up

You'd think a team that cares about traceability would just put the systems model's keys everywhere. My demo has a systems model of its own, describing the demo itself, and it shows how far that gets you. Some keys were designed in from the start. The requirement `SR-22`, that an edit made through the graph patches the source text of the systems model it came from, appears as the short name `SR-22` and as the longer name `SR_22_EditsPatchTheSource`. It's also the start of six Go test names, such as `TestSR22_SetAttributePatchesTextAndProjectionTogether`. User stories reach Checkly as tags in lower case, so the check for `US-04` carries the tag `us-04`. A test in the demo fails if its systems model and the Go test names drift apart.

Other links I wrote by hand. The demo's systems model records that the <span class="term" data-term="adapter">adapter</span>, the service that publishes the example pipeline's systems model to the graph, is the <span class="term" data-term="subgraph">subgraph</span> called `model`, in an attribute nothing checks. And some keys never existed at all. The router that joins the demo's services is `router` in the demo's systems model and `sysml-federation-router` in the traces it sends, a name the systems model never mentions. I chose both names, and I still managed to make them disagree. Of the 146 top-level Go test functions in the repository, 69 carry a requirement's key in their names. The other 77 are invisible to anything that joins on keys.

Each tool also has its own idea of identity, and none of them has any reason to use mine. OpenTofu knows a resource by its [address](https://opentofu.org/docs/cli/state/resource-addressing/), which the author chooses and which only means something inside that configuration. Rename it and, by default, OpenTofu reads the change as an [intent to destroy the old object and create a new one](https://opentofu.org/docs/language/modules/develop/refactoring/), unless a `moved` block says otherwise. OpenTelemetry's `service.name` is the service's [logical name](https://opentelemetry.io/docs/specs/semconv/resource/service/), and when nobody sets one it falls back to `unknown_service:` followed by the name of the executable. Checkly knows a check by its `logicalId`, and [changing that ID](https://www.checklyhq.com/docs/cli/constructs-reference/) tells Checkly that one check was removed and another created.

Even keys that were designed in leak. [Rath and colleagues](https://arxiv.org/abs/1804.02433) studied six open-source projects that largely followed the practice of tagging each commit with an issue key, and found that on average only 60% of commits were linked. One mistyped key is enough to break a join.

Hardware makes all of this harder, because the counterpart of an element of the systems model is an object on a bench or a wall. The [Asset Administration Shell](https://industrialdigitaltwin.org/wp-content/uploads/2024/06/IDTA-01001-3-0-1_SpecificationAssetAdministrationShell_Part1_Metamodel.pdf), an industrial specification for describing such assets, notes that one asset can have several identifiers, among them a serial number, the manufacturer's part number, the customers' own part numbers and an RFID code. None of those is a key from the systems model either. Messages about the asset bring their own naming. <span class="term" data-term="mqtt">MQTT</span> topic names are [case-sensitive and never normalised](https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html), so in a made-up plant, a subscriber to `site1/line2/pumpA/#` hears nothing at all from a pump publishing as `PumpA`. It's the same problem with fewer clues, and this series stays mostly with software, where the built system can at least be asked questions.

## What the links are for

Picture a check on the made-up pipeline failing at three in the morning. Without a link to the systems model, the engineer on call knows that a number fell below a line. With one, they know it's the throughput `PIPE-R1.2` asks of parse, and the model shows them everything else that depends on parse. The link turns an alert into a question about the system, which is a much better thing to be woken up by.

That's the everyday case. Certification is the formal one. Safety standards ask for trace links, and Cleland-Huang and colleagues' [review of the field](http://selab.netlab.uky.edu/homepage/publications/2014-ICSE-FOSE.pdf) quotes DO-178C, the standard for airborne software, on tracing source code to low-level requirements to show there's no undocumented code. The same review found traceability "often conducted in an ad-hoc, after-the-fact manner", at a cost that "can be extremely high". It also reports an analysis of submissions to the US Food and Drug Administration. In many cases the trace data was incomplete, incorrect and conflicting, with clear signs that the links had been created at the very end, for certification.

The field has long wanted the opposite. Antoniol, Cleland-Huang, Hayes and Vierhauser restated the goal in 2017 as the [grand challenge](https://arxiv.org/abs/1710.03129) of traceability that's "always there, without ever having to think about getting it there". Defence engineering has a name for the connected result, the <span class="term" data-term="digital-thread">digital thread</span>, and the US Department of Defense's [digital engineering strategy](https://ac.cto.mil/wp-content/uploads/2019/06/2018-Digital-Engineering-Strategy_Approved_PrintVersion.pdf) expects its authoritative source of truth to "provide traceability as the system of interest evolves".

None of that comes free. A resolver takes effort to build and tune, and a poor one makes confident mistakes at scale. I think the links earn it anyway. If a systems model is going to sit at the centre of an organisation's engineering, it had better be able to point at what was built.

## When links go stale

Much of the thinking about linking records assumes the records sit still. A census return or a published paper is written once and then left alone. A systems model and the system it describes change all the time. Stages get split, services renamed, tests moved and requirements reworded. A link that was right last month can be wrong today, and nobody has to touch it for that to happen.

The traceability literature calls this decay. Links "become stale when source and/or target artifacts are modified", as the same review puts it, and [Mäder and Gotel](https://europepmc.org/article/MED/23471308) argue that links need maintaining as the system evolves "in order to prevent their decay". Rahimi and Cleland-Huang, in the grand challenges report, put the stakes plainly: "outdated trace links invalidate safety-cases".

So finding the links once isn't enough. Resolution has to run again whenever either side changes, the way a test suite does. My demo's own systems model has a small version of this already, in the test that fails when a named Go test or check disappears. What it can't do is find a link nobody wrote down.

## Suggesting links, or adding them

A resolver can be wrong in two ways. It can make a link that shouldn't exist, a false positive, or miss one that should, a false negative. Which of the two you'd rather live with decides how the resolver should be put to work.

One way puts a person in front of every link. While an engineer writes the OpenTofu resource for parse, the resolver suggests `PIPE-S2`, the engineer accepts, and the key goes into the code as a tag. This resolver can lean towards finding everything, because a wrong suggestion costs a moment to reject. Dekhtyar and Hayes [make the same point](https://arxiv.org/abs/1807.11454), that "detecting a false positive is much simpler and faster than discovering an error of omission".

The other way works over what's already there. The resolver reads the configuration, the traces, the checks and the tests, adds the links it's confident about, and leaves people to remove the wrong ones when they find them. It's far faster and cheaper, and it's the only way to cover a system too big to review link by link. It needs the opposite lean, though. A wrong link sits there looking authoritative until someone trips over it, so this resolver has to be careful about what it adds, even if that means missing some.

Rath and colleagues tried both in one study, with one classifier set two ways. Set to suggest, it found about 96% of the true links on average, although only a third of its suggestions were right. Set to add links on its own, more than 89% of what it added was right, and it found half of the links.

Record linkage has a name for the ground in between. In Fellegi and Sunter's method, as [Winkler](https://www.census.gov/content/dam/Census/library/working-papers/2006/adrm/rrs2006-02.pdf) sets it out, two thresholds divide every pair into a link, a non-link and a possible link held for a person to review. How wide you make that middle band decides which way of working you're really running.

People reviewing links aren't a perfect safety net either. When [Cuddeback, Dekhtyar and Hayes](https://digitalcommons.calpoly.edu/csse_fac/108/) had analysts vet candidate trace matrices, the analysts moved them towards a balance of about as many wrong links as missing ones, and poor matrices improved sharply. Follow-up studies found the other half of the story, which the review above sums up: "the higher the quality of the starting trace matrix, the worse the decisions the analyst makes". And decay cuts across both ways of working. A link a person confirmed last year goes stale just as fast as one a machine added.

[Part 2](14-the-parts-of-a-resolution-task.md) takes a single resolution task apart, from what goes in to what comes out, and asks how exactly a task has to be stated before anyone tries to solve it.

---

Previous: [A model of the demo itself](12-a-model-of-the-demo-itself.md), the last part of Federating a systems model · Index: [Automating traceability](../README.md) · Next: [The parts of a resolution task](14-the-parts-of-a-resolution-task.md)

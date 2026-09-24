# What makes a resolver worth running

*Roar Georgsen, 24 September 2026*

Part 5 of 6 in [Automating traceability](../README.md).

> [!IMPORTANT]
> **The story so far**
>
> [Part 4](16-measuring-a-resolver.md) measured a resolver against a gold set that moves with the system, using an F-score picked for the way the resolver is used. A resolver that measures well can still be the wrong one to run. This part asks what else decides it.

Suppose the made-up pipeline had been built, a resolver had added the link from `PIPE-S2` to `module.pipeline.aws_ecs_service.stage["parse"]` on its own, and six weeks later an engineer decides the link is wrong. Before removing it, they'll want to know a few things. Which run added it, and which version of the model and which <span class="term" data-term="opentofu">OpenTofu</span> plan did that run read? Which rule proposed the link, and did a person ever confirm it? If all they have is a row in a table, they can delete the row and still not know whether the next run will put it back, or whether a hundred other links came from the same fault.

## Where a link came from, and where it lives

> [!NOTE]
> **Provenance**
>
> The record of how a piece of data came to be: what produced it, from which inputs, when, and who has vouched for it since. For a trace link, it's what lets a person judge how far to trust the link, and undo a whole run's links if the run went wrong.

The W3C's [PROV ontology](https://www.w3.org/TR/prov-o/) already has words for all of that, and a trace link fits them without strain. Here's a made-up record for the link in question:

```turtle
@prefix prov: <http://www.w3.org/ns/prov#> .
@prefix xsd:  <http://www.w3.org/2001/XMLSchema#> .
@prefix :     <https://example.org/trace#> .

:link-7f3a a :TraceLink, prov:Entity ;
  :from "PIPE-S2" ;
  :to "module.pipeline.aws_ecs_service.stage[\"parse\"]" ;
  :state "added" ;
  prov:wasGeneratedBy :run-1412 ;
  prov:generatedAtTime "2026-10-02T14:12:09Z"^^xsd:dateTime .

:run-1412 a prov:Activity ;
  prov:used :model-9c1e2d4, :tofu-plan-118 ;
  prov:qualifiedAssociation [
    prov:agent :resolver-0.1 ;
    prov:hadPlan :stage-name-rule-v3 ] .

:resolver-0.1 a prov:SoftwareAgent .
:stage-name-rule-v3 a prov:Plan .
```

In words, the run used a model commit and an OpenTofu plan, the resolver was a piece of software, and the rule it followed was, in PROV's terms, a plan. That's two different kinds of plan in one record, and I'm blaming the standards for it.

A few more statements complete the story over time. When a person confirms the link, that's an activity of its own, associated with them. When someone removes the link, that's an invalidation, which [PROV-DM](https://www.w3.org/TR/prov-dm/) defines as "the start of the destruction, cessation, or expiry of an existing entity by an activity". The link stays in the record, marked invalid by a named activity, so the next run can see that someone rejected it. That's the remembered rejection from part 3, in a standard vocabulary. Regulated work asks for the same thing in its own words. The FDA's rule on electronic records, [21 CFR 11.10(e)](https://www.ecfr.gov/current/title-21/chapter-I/subchapter-A/part-11/subpart-B/section-11.10), requires time-stamped audit trails, and says that "record changes shall not obscure previously recorded information". A suggested link needs its evidence on show at the moment someone decides. A link added automatically needs its run on record, so that a bad run can be found and undone in one go.

The record also needs a home. In my demo, each <span class="term" data-term="subgraph">subgraph</span> owns its own fields, and the <span class="term" data-term="router">router</span> joins them on the model's key. Resolved links fit the same pattern. A trace subgraph would extend the model's element types on their short names and add the link fields, each carrying its provenance, and the rest of the graph wouldn't need to know that a resolver made them. That's where part 1's two ends meet. Entity resolution works out the keys, and federation joins on them.

Ownership comes with it. By default, Federation 2 doesn't let two subgraphs resolve the same field. [Apollo's documentation](https://www.apollographql.com/docs/graphos/schema-design/federated-schemas/sharing-types) says a single object field "can't be defined or resolved by more than one subgraph schema". [Cosmo's documentation](https://cosmo-docs.wundergraph.com/federation/directives/shareable) treats sharing one without the `@shareable` directive as a composition error. If the trace subgraph owns the link fields, no other service can publish its own version of them without the graph refusing to compose.

The join rests on the short names, and my own records are honest about the weak spot. The decision that made short names the keys says that "renaming a short name is a breaking change for every service that stores it, and the demo does not handle it". A trace subgraph would store a great many short names, so renaming one is the first thing it would have to handle, and the demo doesn't handle it yet.

## Explaining a link

The capacity service in my demo explains every verdict with one of a few fixed templates, such as `<quantity> <value> against <limit>, limited by <cut>`. Nobody has to wonder where a reason came from, because there's only one place it can have come from. A rule that links a test to `SR-22` because its name starts with `TestSR22_` can do the same, in the same words on every run. Anyone who disagrees can see which rule to change.

A learnt score gives less away. Say a made-up learnt strategy links the <span class="term" data-term="checkly">Checkly</span> browser spec `refusals.multistep.spec.ts`, whose file name carries no key, to a user story. It gives a score of 0.81 against a threshold of 0.75, and lists the words that weighed most. That's something, but both numbers belong to one version of the strategy. [Sculley and colleagues](https://papers.nips.cc/paper_files/paper/2015/hash/86df7dcfd896fcaf2674f757a2463eba-Abstract.html) point out that when a model updates on new data, "the old manually set threshold may be invalid". The explanation a reviewer read last month may not be the one the resolver would give today.

A good explanation also tells the engineer what to change. "No link, because this resource has no `stage` tag" is a problem they can fix in a minute, and the next run gets it right. A score of 0.62 tells them nothing they can act on.

Work that ends in certification wants more again. The certification scenario in Cleland-Huang and colleagues' [review of the field](http://selab.netlab.uky.edu/homepage/publications/2014-ICSE-FOSE.pdf) asks for "supporting rationales explaining how these artifacts mitigate the hazard", and an assessor will want the reason for a link as much as the link. Language models change the picture, because they write fluent explanations of their decisions. In one study, [Rodriguez, Dearstyne and Cleland-Huang](https://arxiv.org/abs/2308.00229) found that a model "could provide an in-depth analysis of its decision". They also said plainly that whether such explanations "are accurate reflections of the reasoning behind the model's decision is beyond the scope of this paper". Part 6 picks that question up.

A reviewer confirming candidates needs a reason for each one at the moment of deciding, or they're guessing. For a link added automatically, the reason matters later, when someone disputes it, and then it has to be the reason the resolver actually had. That's one more thing provenance should keep.

## Running on every change

My demo's trace test reads the whole model and the repository, and its package finishes in a few hundredths of a second. Every pull request runs it afresh, with Go's test cache out of the way. That's the bar I'd hold a resolver to, since part 1 argued that resolution has to rerun whenever either side changes, the way a test suite does.

Most of that speed comes from reading keys. Part 2 noted that comparing everything with everything grows with the product of the two sides. In a made-up system of 500 model elements and 20,000 built objects, an all-pairs scorer faces 10 million pairs, while a key rule needs 20,000 lookups in an index of 500 keys. Blocking, as [Papadakis and colleagues](https://arxiv.org/abs/1905.06167) put it, "trades slightly lower effectiveness for significantly higher efficiency", and in a pipeline that runs on every commit, that trade usually pays. Cost adds up the same way. Compute and memory are paid on every run, retraining on every retrain, and the most accurate resolver is no use if the team won't pay for it a hundred times a week. I'd keep an eye on those numbers over time, too, since they grow with the system.

Each way of working sets its own deadline. Suggesting a key while an engineer writes an OpenTofu resource has to happen while they're still typing, in well under a second. Adding links in bulk after a merge can take minutes, as long as it finishes before the next merge. Recomputing only what changed is cheap and safe for the machine, but part 3 found that reviewing only what changed costs quality, so I'd still show a reviewer each link in its context.

## When the tools change their formats

Say a made-up strategy tells the parse stage's write path from its reads by the `http.method` attribute on its spans. One day the instrumentation switches to OpenTelemetry's stable names, `http.method` becomes `http.request.method`, and the strategy finds nothing. No test fails, and nobody notices. That exact rename is in <span class="term" data-term="opentelemetry">OpenTelemetry's</span> [HTTP migration guide](https://opentelemetry.io/docs/specs/semconv/non-normative/http-migration/), which has instrumentations emit the old names, the new ones or both, depending on an environment variable. OpenTelemetry's [telemetry schemas](https://opentelemetry.io/docs/specs/otel/schemas/) exist for this. They define the transformations from one version to the next, and a resolver that normalises through them first would keep working.

OpenTofu is kinder. `tofu providers schema -json` [versions each provider's schema](https://opentofu.org/docs/cli/commands/providers/schema/), and the [dependency lock file](https://opentofu.org/docs/language/files/dependency-lock/) records the exact provider version chosen, so a resolver can at least tell when the ground has moved. Part 3's split shows the difference. OpenTofu can record a move in a `moved` block, while the telemetry keeps no record that `query-parser` became `query-frontend`.

New objects are the easier half of change. A rule or a key index picks up a new resource the moment it appears. A learnt model may need retraining before it knows about something new, although some can update as they go. For a resolver meant to run for years, that's one of the first things I'd ask about.

My demo has a silent pass of the same kind on record. The decision that turns the router's telemetry off depends on three environment variables the vendor doesn't document, and notes that a release renaming one "would announce it nowhere. Nor would anything here notice." A resolver built on names it doesn't own should expect that, and check the versions of the formats it reads as carefully as the links it makes.

## Who may read the model

My demo's image makes no outbound connection on its own. I ran it with its network removed to check. Its optional check session is a different matter, and deliberately so. Checkly's hosted runners reach the demo through a public tunnel and read the example model, which is fine for a made-up pipeline. A confidential model would need Checkly's private location, which puts the runners beside the demo, and I've never run it because it needs a paid plan.

A resolver faces the same choice, and in some industries the rules make it for you. In US defence work, a system model is likely to count as technical data, which the export rules in [22 CFR 120.33](https://www.ecfr.gov/current/title-22/chapter-I/subchapter-M/part-120/subpart-C/section-120.33) define as information "required for the design, development, production, … or modification of defense articles". The rules exempt some data sent with end-to-end encryption, under conditions that include, in [120.54](https://www.ecfr.gov/current/title-22/chapter-I/subchapter-M/part-120/subpart-C/section-120.54), that "the means of decryption are not provided to any third party". A hosted resolver has to read the model in clear, so as far as I can tell it can't use that exemption. Whether it counts as an export then turns on where it runs and who can reach it. I'm not a lawyer, and this is my reading of the text. When in doubt, I'd run the resolver where the model lives.

## Open and reproducible

An auditor asks why `PIPE-S2` is linked to the parse service. The best answer is to rerun the resolver on the same inputs and get the same link. That's the trace-link version of a [reproducible build](https://reproducible-builds.org/docs/definition/), one where "given the same source code, build environment and build instructions, any party can recreate bit-by-bit identical copies of all specified artifacts". For a resolver, the same model commit, the same plan and the same resolver version should give the same links, which rules out any randomness that isn't pinned down.

Being able to read the resolver's code makes that much easier to check, and open code brings more. Other people test it on their own systems, fix what breaks and improve it, and you're free to change it yourself when your needs differ. The [Open Source Definition](https://opensource.org/osd) adds a clause that matters here. An open licence "must not restrict anyone from making use of the program in a specific field of endeavor", and the fields in question include defence and medical devices. My decision records have a case of the opposite. A SysML tool I considered for the demo was closed source and licence-gated, "with air-gapped use needing a Business plan", and air-gapped is exactly how defence and medical work tends to run. Open code needs documentation and tests beside it, or nobody else can check what it does, and checking is most of the point of being able to read it.

## Choosing one

Here's how I'd decide for the parse task, with made-up candidates. First I'd add a line to the task card from part 2, set before looking at any results:

```
Gate:         F0.5 >= 0.8, under 60 s in CI, on premises
```

Then I'd run the candidates against it. A hosted service scores best, but the model can't leave the building, so it's out. A learnt model scores well too, but retraining it takes an hour after every merge, so it's out as well. That leaves the rule cascade from part 2 at an F0.5 of 0.84, and the same cascade with a learnt model added at 0.85. I'd run the cascade. One point doesn't pay for a fourth moving part. Each strategy is another thing to test and keep current, and when a link goes wrong, fewer parts mean the fault is found faster.

[Sculley and colleagues](https://papers.nips.cc/paper_files/paper/2015/hash/86df7dcfd896fcaf2674f757a2463eba-Abstract.html) describe the machine learning version of this as CACE, "Changing Anything Changes Everything". They also warn that improving one model in an ensemble can make the whole system worse when the errors that remain line up with the other components'. They wrote about learnt models, and applying it to a cascade of rules is my own extension, but I think it carries over.

If nothing passes the gate, the task goes back to suggesting links for a person to confirm, or waits. Sometimes the card itself is wrong, and a task that no resolver can do well is really two tasks, or one aimed at the wrong search space. None of these answers lasts, since the gold set and the tools both move, so I'd ask again whenever the measurements from part 4 shift.

## Where machine learning fits

None of this rules out learnt resolvers. In 2017, [Guo, Cheng and Cleland-Huang](https://arxiv.org/abs/1804.02438) trained a recurrent network on existing trace links from the Positive Train Control domain and tried 360 configurations of it. It significantly outperformed the classic information retrieval methods. They also retrained it on a larger share of the links, a step a growing project would have to repeat. [Lin and colleagues](https://arxiv.org/abs/2102.04411) describe deep trace models as "restricted by availability of labeled data and efficiency at runtime". And large models are costly to train "both financially … and environmentally", as [Strubell, Ganesh and McCallum](https://arxiv.org/abs/1906.02243) measured.

Language models come at the task from the other side. [LiSSA](https://publikationen.bibliothek.kit.edu/1000178348), presented at ICSE 2025, recovers trace links with retrieval-augmented generation, prompting a model with no training on the task itself. It covers three kinds of link, among them architecture documentation to architecture models. That removes the retraining cost, although the retrieval index still has to be kept current. It also brings back two questions from this article, whether a hosted model may read the system model at all, and whether its fluent explanations are the reasons it really had.

Part 6 comes after a small resolver for the demo has been modelled and built. It weighs language models for this job, including open-weights models that can run where the model lives.

---

Previous: [Measuring a resolver](16-measuring-a-resolver.md) · Index: [Automating traceability](../README.md)

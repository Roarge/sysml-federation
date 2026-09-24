# The parts of a resolution task

*Roar Georgsen, 24 September 2026*

Part 2 of 6 in [Automating traceability](../README.md).

> [!IMPORTANT]
> **The story so far**
>
> [Part 1](13-why-a-model-needs-entity-resolution.md) called the job entity resolution: deciding which objects in the built system are the same entity as a model element, so the trace links between them can be found and kept. It followed the parse stage of the first series' pipeline into four tools, with a made-up counterpart in each, and found no shared key in any of them. This part takes one such task apart, and asks how exactly a task has to be stated before anyone builds something to do it.

## Two questions that sound alike

"Which OpenTofu resources are the parse stage?" and "Which model element does this span belong to?" sound like one question asked from two ends. The first starts from one model element and searches an <span class="term" data-term="opentofu">OpenTofu</span> configuration for any number of resources. The second starts from a single object in the running system and has to pick at most one element out of a whole model. A wrong answer to the first is an extra resource or a missing one, and a wrong answer to the second files a span under the wrong part.

So a task needs a name that says which way it runs and which relationship it records. The traceability community's own [glossary](http://sarec.nd.edu/coest/glossary.html) builds direction into its definition of a trace link. A link associates a source with a target, and it has a primary direction and a reverse one, so that if A tests B then B is tested by A. It may also carry a type, such as tests, implements or refines. By that measure, "resolve the parse stage" is still a topic, and "find the OpenTofu resources that deploy each part of the pipeline" is a task.

## A task card

Writing the task down in a fixed shape is the quickest way I know to find out whether I understand it. Here is the card for that second, better-named task, filled in for the made-up pipeline:

```
Task:         pipeline part -> OpenTofu resources deploying it
Used by:      engineers tracing a requirement to its deployment
Direction:    model -> built system
Search space: resources in module.pipeline only
Out of scope: outside module.pipeline, no link and no error
Input:        the plan or state, from tofu show -json
Strategy:     model_element tag, else stage name in address
Score:        1.0 for a tag, lower for a name in the address
Output:       zero or more resource instances per part
Link type:    deploys (a resource deploys a part)
Time:         one plan at a time, rerun when the plan changes
```

Every line on it is a decision somebody would otherwise make by accident.

## What counts as a candidate

A resolver first has to decide what it's choosing from. Entity resolution research splits the work into [two broad steps](https://arxiv.org/abs/1905.06167). The first chooses which pairs are worth comparing, and the second compares them. The first step exists because comparing every model element with every object in a real system grows with the product of the two, and almost all of those comparisons are pointless. In the made-up pipeline, the card has already limited the search space to `module.pipeline`. A resolver that then compares the parse stage only with the module's `aws_ecs_service` instances, and skips its log groups and security groups, has taken that first step.

> [!NOTE]
> **Candidate**
>
> A possible link that a strategy has proposed and not yet decided on, usually with a score. A task narrows its search space down to candidates, compares each one with the input, and then decides which become links.

The search space also decides what the resolver must refuse. The demo from the first series has two models in it, the demo's own model and the example pipeline it serves. One of the demo's own <span class="term" data-term="checkly">Checkly</span> checks queries `requirement(id: "PIPE-R1")`, a real key that resolves perfectly, but in the wrong model. A task that resolves checks against the demo's model has to say what happens to it. My answer is that it produces no link and no error, because the key belongs to a model the task doesn't cover.

## Reading the input

In the made-up pipeline, one block of configuration deploys three of the five stages:

```hcl
resource "aws_ecs_service" "stage" {
  for_each = toset(["ingest", "parse", "serve"])
  name     = "pipeline-${each.key}"
  tags     = { stage = each.key }
}
```

Read as configuration, that's one resource. Read from the plan or the state with `tofu show -json`, it's three instances, each with its own [address](https://opentofu.org/docs/cli/state/resource-addressing/), and the parse one is `module.pipeline.aws_ecs_service.stage["parse"]`. OpenTofu's [JSON format](https://opentofu.org/docs/internals/json-format/) tells callers to treat those addresses as opaque strings they may compare in full. The `stage` tag is a firmer hint than the address, and still no model key. A `model_element` tag, had anyone written one, would be the easy case.

Some inputs need work before resolution can start. A made-up commit message such as `fix(adapter): keep patches atomic under load (see SR 22)` holds a key only after someone extracts "SR 22" and spells it `SR-22`. A [trailer](https://git-scm.com/docs/git-interpret-trailers) such as `Refs: SR-22` at the end of the message is structured already. I count extraction as part of the task, because its mistakes should count against the task. A resolver that can't read its own input hasn't resolved anything.

## Comparing

I'd start with the cheapest rule that's certain and fall back from there. A key in a tag comes first, then a key in a name, then a comparison of the names themselves, with whatever each rule finds merged into one list of candidates. The demo already ships a rule of the second kind. The test that holds its model in step with the repository finds a requirement's Go tests by a name prefix such as `TestSR22_`, and by nothing else. Its package comment is honest about how far rules like that go. A form the model could take that its patterns don't match "is a hole in the check rather than a pass".

## Scoring and deciding

A strategy can attach a score to each candidate, and it's tempting to read a score between 0 and 1 as a probability. It is one only if it's [calibrated](https://arxiv.org/abs/1706.04599). Of all the candidates a calibrated strategy scores about 0.8, about 80% really are links. A cosine similarity between two names is good for ranking candidates, but on its own it says nothing about how often a 0.8 is wrong. Cleland-Huang and colleagues' [review of the field](http://selab.netlab.uky.edu/homepage/publications/2014-ICSE-FOSE.pdf) calls confidence scores on trace links appealing. I'd add that they're only as good as the calibration behind them.

The score is also where the two ways of working from part 1 land on the card. The record linkage tradition, as [Winkler](https://www.census.gov/content/dam/Census/library/working-papers/2006/adrm/rrs2006-02.pdf) sets it out, puts two thresholds on the score. Above the upper one a pair is a link, below the lower one it isn't, and in between it's held for clerical review. Recommending links for a person to confirm widens that middle band. Adding links automatically narrows it. The card's "Score" line is where that choice gets written down.

## What comes out, and how many

Depending on the direction, the output is a set of model elements or of built-system objects. How many can come out, the task's cardinality, is where the made-up examples stop being tidy. The `for_each` block above is one input producing three instances, each belonging to a different part. Turned the other way, one part can run as many instances, and <span class="term" data-term="opentelemetry">OpenTelemetry</span> expects exactly that. Its `service.name` [must be the same](https://opentelemetry.io/docs/specs/semconv/resource/service/) for every instance of a horizontally scaled service, and `service.instance.id` tells the instances apart. The [deployment environment](https://opentelemetry.io/docs/specs/semconv/resource/deployment-environment/) doesn't count towards a service's identity at all, so staging and production traces from `query-parser` both belong to `PIPE-S2`. Resolved from the model's side, then, `PIPE-S2` has one service name and any number of instance IDs, and the card has to say which of them a link points at.

The demo has real cases too. One verification case lists six Go tests as evidence for the requirement `SR-22`, and `SR-22` itself is derived from four stakeholder stories. There are two test functions called `TestSR25_InvalidValuesAreRefused`, in two different packages, so a search space of test names alone would quietly merge them into one. The model tells them apart by file. And one of my own commits has a subject that begins "model: US-19, SR-03 and SR-48 done", three keys in one input. Whether that commit yields three links, or none until a person splits it, has to be decided before the task runs, and written on the card.

## Where the link lands

Once a task has decided, the link has to live somewhere, and SysML v2 offers more than one kind. A requirement is satisfied by a part and verified by a case. A requirement derives from another. An [allocation](https://www.omg.org/spec/SysML/2.0/Language/PDF) says a target is responsible for realising some or all of a source's intent, and a plain dependency can connect any number of clients to any number of suppliers. Each of those relationships is a model element in its own right. Choosing among them is part of stating the task, since "deploys" and "verifies" aren't the same claim.

For links that leave the model, the OMG's [Systems Modeling API](https://www.omg.org/spec/SystemsModelingAPI/1.0/PDF) defines an external relationship from a model element to data elsewhere, named by an IRI, the internationalised form of a URI. That's a natural home for a link to an OpenTofu resource or a Checkly check. Every model element also has an element ID that never changes once it's created, which would make the stable key underneath. The [text notation](https://www.omg.org/spec/SysML/2.0/Language/PDF) has no way to write one, though, and the first series joined on short names, which an author writes and a reader can type.

## When the target moves

Time belongs on the card because both sides change, and the tools disagree about what a change means. Rename a resource in OpenTofu and add a [`moved` block](https://opentofu.org/docs/language/modules/develop/refactoring/), and OpenTofu treats the existing object as belonging to the new address. The plan's JSON even records the `previous_address`. The entity survived the rename. Change a Checkly check's `logicalId`, and Checkly [reads it](https://www.checklyhq.com/docs/constructs/overview/) as one check removed and another created. As far as Checkly is concerned, the old check is gone.

A resolver that follows OpenTofu's rule can carry a link across a rename. One that follows Checkly's has to find the link again from nothing. Neither tool is wrong. Each knows what its users need from it, and a task that reads both has to say which rule it follows for each. A link is resolved against one version of each side, and the card should say which.

## One task or two

Services in a Compose file and services in OpenTelemetry both have to resolve to model parts. OpenTofu configuration and OpenTofu state describe the same resources in two formats. Each pair could be one task or two, and the card's "Used by" line is what settles it for me. The question is whether the people using the links would treat a mistake from one source the same way as a mistake from the other.

A Compose service and a traced service are different kinds of evidence, even when they name the same part. A wrong link from telemetry sends the engineer on call to the wrong place, and a wrong link from the Compose file breaks a deployment view. I'd keep them apart, with their own cards and their own scores. Reading OpenTofu configuration or state is a job for the input reader, and the links it produces mean the same thing either way, so I'd call that one task.

Part 3, coming up next, asks why no task gets every link right however carefully it's stated, and why that's less of a problem than it sounds.

---

Previous: [Why a model needs entity resolution](13-why-a-model-needs-entity-resolution.md) · Index: [Automating traceability](../README.md)

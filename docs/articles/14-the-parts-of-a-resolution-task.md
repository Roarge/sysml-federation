# The parts of a resolution task

*Roar Georgsen, 24 September 2026*

Part 2 of 6 in [Automating traceability](../README.md).

> [!IMPORTANT]
> **The story so far**
>
> [Part 1](13-why-a-model-needs-entity-resolution.md) called the job entity resolution: deciding which things in a built system are the same as an element of its model, so that trace links between them can be found and kept. Its example was a made-up pipeline whose parse stage turned up under a different name in each of four tools. This part takes a single resolution task apart, and asks how exactly it needs stating before anyone builds something to do it.

"Link the parse stage to the running system" sounds like a clear enough request, right up until someone sits down to do it. One engineer might start from the model and go through the infrastructure code for anything that deploys parse. Another might start from the traces and work out which part of the model each service belongs to. Both are doing entity resolution, and they're doing different jobs. The first can come back with any number of resources. The second has to pick at most one model element for each service, out of a whole model. A wrong answer to the first is an extra resource or a missing one, and a wrong answer to the second files a service under the wrong part.

So a task needs a name that says which way it runs and which relationship it records. The traceability community's own [glossary](http://sarec.nd.edu/coest/glossary.html) builds direction into its definition of a trace link. A link joins a source to a target, and it reads both ways, so that if A tests B then B is tested by A. It may also carry a type, such as tests, implements or refines. By that measure, "resolve the parse stage" is still a topic. "Find the <span class="term" data-term="opentofu">OpenTofu</span> resources that deploy each part of the pipeline" is a task.

## A card for one task

The quickest way I know to find out whether I understand a task is to write it down in a fixed shape. Here's the card I'd fill in for that second, better-named task on the made-up pipeline:

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

Every line on it is a decision somebody would otherwise make by accident, and a few of them deserve more than a moment's thought.

## What goes in

A resolver first has to decide what it's choosing from. Entity resolution research splits the work into [two broad steps](https://arxiv.org/abs/1905.06167). The first picks out the pairs worth comparing, and the second compares them. The first step exists because comparing every model element with every object in a real system grows with the product of the two, and almost all of those comparisons are pointless. In the made-up pipeline, the card has already limited the search space to `module.pipeline`. A resolver that then compares the parse stage only with the module's `aws_ecs_service` instances, and skips its log groups and security groups, has taken that first step. Whatever survives it is a candidate.

> [!NOTE]
> **Candidate**
>
> A possible link that a strategy has proposed and not yet decided on, usually with a score. A task narrows its search space down to candidates, compares each one with the input, and then decides which become links.

The search space also decides what a resolver should refuse, and my own demo has a neat example. The demo holds two models, its own and the example pipeline it serves. One of its <span class="term" data-term="checkly">Checkly</span> checks asks for `requirement(id: "PIPE-R1")`, which is a real key that resolves perfectly, just in the wrong model. A task that links checks to the demo's own model has to say what happens to it. My answer is no link and no error, because the key belongs to a model the task doesn't cover.

Then there's the input itself, which rarely arrives in the shape you'd like. In the made-up pipeline, one block of configuration deploys three of the five stages:

```hcl
resource "aws_ecs_service" "stage" {
  for_each = toset(["ingest", "parse", "serve"])
  name     = "pipeline-${each.key}"
  tags     = { stage = each.key }
}
```

Read as configuration, that's one resource. Read from the plan or the state with `tofu show -json`, it's three instances, each with its own [address](https://opentofu.org/docs/cli/state/resource-addressing/), and the parse one is `module.pipeline.aws_ecs_service.stage["parse"]`. OpenTofu's [JSON format](https://opentofu.org/docs/internals/json-format/) tells callers to treat those addresses as opaque strings they may compare in full. The `stage` tag is a firmer hint than the address, and it's still no model key. A `model_element` tag, had anyone written one, would be the easy case.

Some inputs need work before resolution can even start. A made-up commit message such as `fix(adapter): keep patches atomic under load (see SR 22)` only holds a key once someone extracts "SR 22" and spells it `SR-22`. A [trailer](https://git-scm.com/docs/git-interpret-trailers) such as `Refs: SR-22` at the end of the message is structured already. I count extraction as part of the task, so its mistakes count against the task too. A resolver that can't read its own input hasn't resolved anything.

## How candidates are compared

I'd start with the cheapest rule that's certain and fall back from there. A key in a tag comes first, then a key in a name, then a comparison of the names themselves, with whatever each rule finds merged into one list of candidates. The demo already ships a rule of the second kind. The test that holds its model in step with the repository finds a requirement's Go tests by a name prefix such as `TestSR22_`, and by nothing else. Its package comment is honest about how far rules like that go. A form the model could take that its patterns don't match "is a hole in the check rather than a pass".

Rules further down the list come with a score, and it's tempting to read a score between 0 and 1 as a probability. It is one only if it's [calibrated](https://arxiv.org/abs/1706.04599). Of all the candidates a calibrated strategy scores about 0.8, about 80% really are links. A similarity score between two names is good for putting candidates in order, but on its own it says nothing about how often a 0.8 is wrong. Cleland-Huang and colleagues' [review of the field](http://selab.netlab.uky.edu/homepage/publications/2014-ICSE-FOSE.pdf) calls confidence scores on trace links appealing, and I'd agree, with the rider that they're only as good as the calibration behind them.

The score is also where part 1's two ways of working land on the card. The record linkage tradition, as [Winkler](https://www.census.gov/content/dam/Census/library/working-papers/2006/adrm/rrs2006-02.pdf) sets it out, puts two thresholds on the score. Above the upper one a pair is a link, below the lower one it isn't, and anything in between waits for a person to review it. Suggesting links for a person to confirm widens that middle band. Adding links automatically narrows it. The card's "Score" line is where that choice gets written down.

## What comes out

Depending on the direction, the output is a set of model elements or a set of objects from the built system. How many can come out, the task's cardinality, is where the made-up examples stop being tidy. The `for_each` block above is one input producing three instances, each belonging to a different part. Turned the other way, one part can run as many instances as it likes, and <span class="term" data-term="opentelemetry">OpenTelemetry</span> expects exactly that. Its `service.name` [must be the same](https://opentelemetry.io/docs/specs/semconv/resource/service/) for every copy of a service running side by side, and `service.instance.id` tells the copies apart. The [deployment environment](https://opentelemetry.io/docs/specs/semconv/resource/deployment-environment/) doesn't count towards a service's identity at all, so traces from `query-parser` in staging and in production both belong to `PIPE-S2`. Seen from the model, `PIPE-S2` has one service name and any number of instance IDs, and the card has to say which of them a link points at.

My demo is less tidy still. One verification case lists six Go tests as evidence for the requirement `SR-22`, and `SR-22` itself derives from four user stories. I also managed to give two test functions the same name, `TestSR25_InvalidValuesAreRefused`, in two different packages, so a search space of test names alone would quietly merge them into one. The model tells them apart by file. And one of my own commit subjects, clearly not written with a resolver in mind, begins "model: US-19, SR-03 and SR-48 done", which is three keys in one input. Whether that commit yields three links, or none until a person splits it, has to be decided before the task runs, and written on the card.

Once a task has decided, the link has to be kept somewhere, and SysML v2 offers several kinds. A requirement is satisfied by a part and verified by a case. One requirement can derive from another. An [allocation](https://www.omg.org/spec/SysML/2.0/Language/PDF) says a target is responsible for realising some or all of a source's intent, and a plain dependency can connect any number of elements to any number of others. Each of those relationships is a model element in its own right. Choosing among them is part of stating the task, since "deploys" and "verifies" aren't the same claim.

Links that leave the model need a home too. The OMG's [Systems Modeling API](https://www.omg.org/spec/SystemsModelingAPI/1.0/PDF) defines an external relationship from a model element to data elsewhere, named by an IRI, the internationalised form of a URI. That's a natural place for a link to an OpenTofu resource or a Checkly check. Underneath, every model element also has an element ID that never changes once it's created, which would make the ideal key. The [text notation](https://www.omg.org/spec/SysML/2.0/Language/PDF) has no way to write one, though, so my demo joins on short names, which an author writes and a reader can type.

## When things get renamed

Both sides change, and the tools disagree about what a change means. Rename a resource in OpenTofu and add a [`moved` block](https://opentofu.org/docs/language/modules/develop/refactoring/), and OpenTofu treats the existing object as belonging to the new address. The plan's JSON even records the `previous_address`. As far as OpenTofu is concerned, the thing survived its rename. Change a Checkly check's `logicalId`, and Checkly [reads it](https://www.checklyhq.com/docs/constructs/overview/) as one check removed and another created. As far as Checkly is concerned, the old check is gone.

A resolver that follows OpenTofu's rule can carry a link across a rename. One that follows Checkly's has to find the link again from nothing. Neither tool is wrong, since each knows what its own users need, but a task that reads both has to say which rule it follows for each. A link is always resolved against one version of each side, and the card should say which.

## One task or two

Services in a Compose file and services in OpenTelemetry both have to resolve to parts of the model. OpenTofu configuration and OpenTofu state describe the same resources in two formats. Either pair could be one task or two, and the card's "Used by" line is what settles it for me. The question is whether the people using the links would treat a mistake from one source the same way as a mistake from the other.

A Compose service and a traced service are different kinds of evidence, even when they name the same part. A wrong link from telemetry sends the engineer on call to the wrong place, and a wrong link from the Compose file breaks a deployment view. I'd keep them apart, with their own cards and their own scores. Reading OpenTofu configuration or state is a job for the input reader, and the links it produces mean the same thing either way, so I'd call that one task.

[Part 3](15-why-no-resolver-gets-every-link-right.md) asks why no task gets every link right however carefully it's stated, and why that's less of a problem than it sounds.

---

Previous: [Why does a model need entity resolution?](13-why-a-model-needs-entity-resolution.md) · Index: [Automating traceability](../README.md) · Next: [Why no resolver gets every link right](15-why-no-resolver-gets-every-link-right.md)

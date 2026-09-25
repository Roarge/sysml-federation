# The parts of a resolution task

*Roar Elias Georgsen, 25 September 2026*

Part 2 of 6 in [Automating traceability](../README.md).

> [!IMPORTANT]
> **The story so far**
>
> [Part 1](13-why-a-model-needs-entity-resolution.md) introduced entity resolution, deciding which things in a built system are the same as an element of its systems model, so that trace links between them can be found and maintained. Its example was a made-up pipeline whose parse stage turned up under a different name in each of four tools. This part takes a single resolution task apart, and asks how exactly it needs stating before anyone builds something to do it.

"Link the parse stage in the system model to the running parsing service" sounds like a clear enough request, right up until someone sits down to do it. One engineer might start from the <span class="term" data-term="systems-model">systems model</span> and go through the infrastructure code for anything that deploys parse. Another might start from the traces and work out which element of the systems engineering model each service belongs to. Both are doing entity resolution, and they're doing different jobs. The first can come back with any number of resources. The second has to pick, for each service, one or more elements out of the whole systems model, such as a <span class="term" data-term="part">part</span> or a requirement. A wrong answer to the first is an extra resource or a missing one, and a wrong answer to the second files a service under the wrong part.

So a task needs a name that says which way it runs and which relationship it records. The traceability community's own [glossary](http://sarec.nd.edu/coest/glossary.html) builds direction into its definition of a trace link. A link joins a source to a target, and it reads both ways, so that if A tests B then B is tested by A. It may also carry a type, such as tests, implements or refines. By that measure, "resolve the parse stage" is still a topic. "Find the <span class="term" data-term="opentofu">OpenTofu</span> resources that deploy each part of the pipeline" is a task.

## A card for one task

The quickest way I know to find out whether I understand a task is to write it down in a fixed shape. Here's the card I'd fill in for that second, better-named task on the made-up pipeline:

```
Task:         pipeline part -> OpenTofu resources deploying it
Used by:      engineers tracing a requirement to its deployment
Direction:    systems model -> built system
Search space: resources in module.pipeline only
Out of scope: outside module.pipeline, no link and no error
Input:        the plan or state, from tofu show -json
Strategy:     key in a model_element tag, else stage in address
Score:        1.0 for a tag, lower for a name in the address
Output:       zero or more resource instances per part
Link type:    deploys (a resource deploys a part)
Time:         one plan at a time, rerun when the plan changes
```

Every line on it is a decision somebody would otherwise make by accident, and a few of them deserve more than a moment's thought.

## What goes in

A resolver first has to decide what it's choosing from. Researchers in entity resolution split the work into [two broad steps](https://arxiv.org/abs/1905.06167). The first, which they call blocking, picks out the pairs worth comparing, and the second compares them. The first step exists because comparing every element of the systems model with every object in a real system grows with the product of the two, and almost all of those comparisons are pointless. In the made-up pipeline, the card has already limited the search space to `module.pipeline`. A resolver that then compares the parse stage only with the module's `aws_ecs_service` instances, and skips its log groups and security groups, has taken that first step. Whatever survives it is a candidate.

> [!NOTE]
> **Candidate**
>
> A possible link that a strategy has proposed and not yet decided on, usually with a score. A task narrows its search space down to candidates, compares each one with the input, and then decides which become links.

The search space also decides what a resolver should refuse, and my own demo has a useful example. The demo holds two systems models, its own and the example pipeline's. One of its <span class="term" data-term="checkly">Checkly</span> checks asks for `requirement(id: "PIPE-R1")`, which is a real key that resolves perfectly, just in the wrong systems model. A task that links checks to the demo's own systems model has to say what happens to that check. My answer is no link and no error. Each systems model works like a namespace for its keys, and `PIPE-R1` belongs to one the task doesn't cover.

Then there's the input itself, which rarely arrives in the shape you'd like. In the made-up pipeline, one block of configuration deploys three of the five stages:

```hcl
resource "aws_ecs_service" "stage" {
  for_each = toset(["ingest", "parse", "serve"])
  name     = "pipeline-${each.key}"
  tags     = { stage = each.key }
}
```

Read as configuration, that's one resource. Read from the plan or the state with `tofu show -json`, it's three instances, each with its own [address](https://opentofu.org/docs/cli/state/resource-addressing/), and the parse one is `module.pipeline.aws_ecs_service.stage["parse"]`. OpenTofu's [JSON format](https://opentofu.org/docs/internals/json-format/) tells callers to treat those addresses as opaque strings they may compare in full. The `stage` tag is a firmer hint than the address, and it's still not a key from the systems model. A `model_element` tag carrying the key `PIPE-S2`, had anyone written one, would be the easy case.

Some inputs need work before resolution can even start. A made-up commit message such as `fix(adapter): keep patches atomic under load (see SR 22)` only holds a key once someone extracts "SR 22" and spells it `SR-22`. A [trailer](https://git-scm.com/docs/git-interpret-trailers) such as `Refs: SR-22` at the end of the message is structured already. I count extraction as part of the task, so its mistakes count against the task too. A resolver that can't read its own input hasn't resolved anything.

## How candidates are compared

I'd start with the cheapest rule that's certain and fall back from there. A key in a tag comes first, then a key in a name, then a comparison of the names themselves, with whatever each rule finds merged into one list of candidates. The demo already ships a rule of the second kind. The test that holds its systems model in step with the repository finds a requirement's Go tests by a name prefix such as `TestSR22_`, and by nothing else. Its package comment says how far rules like that go. A form the systems model could take that the test's patterns don't match "is a hole in the check rather than a pass".

Rules further down the list come with a score, and it's tempting to read a score between 0 and 1 as a probability. It is one only if it's [calibrated](https://arxiv.org/abs/1706.04599). Of all the candidates a calibrated strategy scores about 0.8, about 80% really are links. A similarity score between two names is good for putting candidates in order, but on its own it says nothing about how often a 0.8 is wrong. Cleland-Huang and colleagues' [review of the field](http://selab.netlab.uky.edu/homepage/publications/2014-ICSE-FOSE.pdf) calls confidence scores on trace links appealing, and I'd agree, with the rider that they're only as good as the calibration behind them.

The score is also where part 1's two ways of working land on the card. [Winkler's](https://www.census.gov/content/dam/Census/library/working-papers/2006/adrm/rrs2006-02.pdf) account of record linkage puts two thresholds on the score. Above the upper one a pair is a link, below the lower one it isn't, and anything in between waits for a person to review it. Suggesting links for a person to confirm widens that middle band. Adding links automatically narrows it. The card's "Score" line is where that choice gets written down.

## What comes out

Depending on the direction, the output is a set of elements of the systems model or a set of objects from the built system. How many can come out, the task's cardinality, is where the made-up examples stop being tidy. The `for_each` block above is one input producing three instances, each belonging to a different part. Turned the other way, one part can run as many instances as it likes, and <span class="term" data-term="opentelemetry">OpenTelemetry</span> expects exactly that. Its `service.name` [must be the same](https://opentelemetry.io/docs/specs/semconv/resource/service/) for every copy of a service running side by side, and `service.instance.id` tells the copies apart. The [deployment environment](https://opentelemetry.io/docs/specs/semconv/resource/deployment-environment/) doesn't count towards a service's identity at all, so traces from `query-parser` in staging and in production both belong to `PIPE-S2`. Seen from the systems model, `PIPE-S2` has one service name and any number of instance IDs, and the card has to say which of them a link points at.

My demo is less tidy still. One <span class="term" data-term="verification-case">verification case</span> lists six Go tests as evidence for the requirement `SR-22`, and `SR-22` itself <span class="term" data-term="derivation">derives</span> from four user stories. I also managed to give two test functions the same name, `TestSR25_InvalidValuesAreRefused`, in two different packages, so a search space of test names alone would quietly merge them into one. The demo's systems model tells them apart by file. And one of my own commit subjects for the systems model, clearly not written with a resolver in mind, begins "model: US-19, SR-03 and SR-48 done", which is three keys in one input. Whether that commit yields three links, or none until a person splits it, has to be decided before the task runs, and written on the card.

Once a task has decided, the link has to be recorded somewhere, and SysML v2 offers several kinds of relationship to record it with. A requirement is <span class="term" data-term="satisfy">satisfied</span> by a part and verified by a verification case. One requirement can derive from another. An [allocation](https://www.omg.org/spec/SysML/2.0/Language/PDF) says a target is responsible for realising some or all of a source's intent, and a plain dependency can connect any number of elements to any number of others. Choosing among them is part of stating the task, since "deploys" and "verifies" aren't the same claim.

Each of those relationships is itself an element of the systems model, and a systems model is dense with them long before any resolver runs. Whoever built it modelled which part satisfies which requirement, which verification case verifies it and where each requirement came from. A resolver that has found one element can follow those relationships to its neighbours, which a list of names can't offer.

Links that leave the systems model need a home too. The Object Management Group's [Systems Modeling API](https://www.omg.org/spec/SystemsModelingAPI/1.0/PDF) defines an external relationship from an element of a systems model to data elsewhere, named by an IRI, the internationalised form of a URI. That's a natural place for a link to an OpenTofu resource or a Checkly check.

Underneath, every element of a SysML systems model also has an element ID, a globally unique identifier that never changes once the element exists. It would make the ideal shared key, and it shows a line SysML v2 draws between a systems model and the text you read it in. The [language specification](https://www.omg.org/spec/SysML/2.0/Language/PDF) calls a systems model's underlying structure its abstract syntax. The textual and graphical notations present that structure, and a tool can parse them to create or update it. Element IDs live in the abstract syntax. The [KerML specification](https://www.omg.org/spec/KerML/1.0/PDF), for the Kernel Modeling Language that SysML v2 builds on, leaves them to the modelling tool. The textual notation has no way to state one. Parse a file into a tool, and the tool assigns the IDs. Store only the text, and there are none to store.

That's the gap between two camps. Enthusiasts for SysML v2 like the textual notation because a systems model in text can live in Git, where, as one [introductory course](https://sensmetry.com/advent-of-sysml-v2-lesson-6-version-control-with-git/) puts it, "you can see exactly what changed". Sceptics answer, in my experience, that the syntax is not the model, and element IDs are part of what they mean. My demo sits firmly in the first camp. It reads its systems models straight from text files, with no modelling tool behind them to assign IDs, so it joins on short names instead, which an author writes and a reader can type.

## When things get renamed

Both sides change, and the tools disagree about what a change means. Rename the pipeline's `stage` resource in OpenTofu and add a [`moved` block](https://opentofu.org/docs/language/modules/develop/refactoring/), and OpenTofu treats the existing objects as belonging to the new address. The plan's JSON even records each one's `previous_address`. As far as OpenTofu is concerned, the parse instance survived its rename. Change the `logicalId` of the made-up check `parse-sustained-qps`, and Checkly [reads the change](https://www.checklyhq.com/docs/constructs/overview/) as one check removed and another created. As far as Checkly is concerned, the old check is gone.

A resolver that follows OpenTofu's rule can carry a link across a rename. One that follows Checkly's has to find the link again from nothing. Neither tool is wrong, since each knows what its own users need, but a task that reads both has to say which rule it follows for each. A link is always resolved against one version of each side, and the card should say which.

## One task or two

Services in a Docker Compose file and services in OpenTelemetry both have to resolve to parts of the systems model, and they could be one task or two. The card's "Used by" line is what settles it for me. The question is whether the people using the links would treat a mistake from one source the same way as a mistake from the other.

A Compose service and a traced service are different kinds of evidence, even when they name the same part. A wrong link from telemetry sends the engineer on call to the wrong place, and a wrong link from the Compose file breaks a deployment view. I'd keep them apart, with their own cards and their own scores. Reading OpenTofu's configuration and its state is a different matter, since both describe the same resources and the links mean the same thing either way, so I'd call that one task.

Part 3, coming up next, asks why no task gets every link right however carefully it's stated, and why that's less of a problem than it sounds.

---

Previous: [Why does a systems model need entity resolution?](13-why-a-model-needs-entity-resolution.md) · Index: [Automating traceability](../README.md)

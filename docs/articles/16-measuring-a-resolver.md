# Measuring a resolver

*Roar Georgsen, 24 September 2026*

Part 4 of 6 in [Automating traceability](../README.md).

> [!IMPORTANT]
> **The story so far**
>
> [Part 3](15-why-no-resolver-gets-every-link-right.md) argued that every resolver gets some links wrong, and that which errors matter more depends on how it's put to work. Whether one resolver is good enough for one job is a question of measurement. This part asks what to measure it against, and with which numbers, when the right answers keep changing.

My demo's own <span class="term" data-term="systems-model">systems model</span> came with a ready-made test for a resolver, and I'm a little embarrassed by how well any resolver would do on it. Its verification cases name 69 Go tests as evidence, spread over 32 cases, and I wrote every one of those links in by hand. Each of the 69 test names carries, right after `Test`, the key of the requirement its case verifies, so `TestSR22_SetAttributePatchesTextAndProjectionTogether` sits under the case for `SR-22`.

A resolver that reads the requirement key in test names would find all 69 and nothing else. That's a perfect score, and it would prove nothing, because the links were findable by the very rule being scored. The 77 test functions without a key, which part 1 counted, are the ones a resolver is actually needed for. A score on the 69 says nothing about them, and nothing at all about <span class="term" data-term="opentofu">OpenTofu</span> resources or traces.

Links that someone has already made are a tempting thing to measure against anywhere, since the work is paid for. They're also a biased sample. [Bird and colleagues](https://www.cs.ucdavis.edu/~filkov/papers/biasbusters.pdf) looked at the bug-fix datasets that defect research relies on. They found that "only a fraction of bug fixes are actually labelled", and "strong evidence of systematic bias" in which ones. [Rath and colleagues](https://arxiv.org/abs/1804.02433) saw a hint of the same thing with trace links. Their classifier, trained on commits that developers had linked to issues, did worse on commits nobody had linked. For one project's improvements, precision fell from 0.32 to 0.11, although that second figure comes from a small hand-checked sample.

> [!NOTE]
> **Gold set**
>
> The links a resolver's output is measured against, each one checked by people. Traceability research calls it an answer set. It includes the elements that should get no link as well as those that should, and it holds only for the version of the system it was built from.

In the terms of part 1, a gold set is the join the federation would make if every object had carried the systems model's keys from the start. Measuring a resolver means running it on every input the gold set covers and comparing what it proposes with what the gold set says, link by link. The comparison then gets turned into numbers. The rest is choosing which numbers, and building a gold set worth measuring against.

## Why accuracy flatters

Accuracy is the first number most people reach for, and for trace links it's close to useless. Take the demo's tests and verification cases as a grid. Any of the 146 test functions could in principle verify any of the 55 cases, which makes 8,030 pairs, and 69 of them are links. A resolver that links nothing at all gets 7,961 of the 8,030 pairs right, for an accuracy of 99.14%. A tool advertised as 99% accurate might be doing exactly that, so a bare figure like that means nothing until you know what was counted.

Record linkage reached the same verdict a long time ago. [Christen and Goiser](https://users.cecs.anu.edu.au/~christen/publications/qmdm2007christen.pdf) say plainly that accuracy "is therefore not a good quality measure for data linkage and deduplication, and should not be used". Possible pairs grow with the product of the two sides, while true links grow roughly with their size, so on any real system the empty cells swamp everything else. Precision and recall, from part 3, leave the empty cells out. Precision looks only at the links a resolver proposed, and recall only at the links that ought to exist.

## Picking the F-score to suit the job

Two numbers are awkward when you want to compare resolvers, so people combine them into an F-score, a weighted average of precision and recall. The weight is the part that needs a decision. In [van Rijsbergen's](http://www.dcs.gla.ac.uk/Keith/Chapter.7/Ch.7.html) formulation, the parameter β describes a user who attaches β times as much importance to recall as to precision. So F1 treats the two equally, while F2 cares more about finding everything and F0.5 more about being right.

Here's part 3's made-up telemetry task again, at its two thresholds, with the three F-scores added:

| Threshold | Precision | Recall | F0.5 | F1 | F2 |
|---|---|---|---|---|---|
| 0.8 | 0.92 | 0.58 | 0.82 | 0.71 | 0.62 |
| 0.5 | 0.60 | 0.90 | 0.64 | 0.72 | 0.82 |

F1 can barely tell the two settings apart. F0.5 clearly prefers the strict one, which is the setting I'd trust to add links on its own. F2 clearly prefers the loose one, which suits a team that suggests links for a person to confirm. Their reviewer sees more of the right links and rejects the others cheaply. Rath and colleagues, whose two settings part 1 described, made the same choice and scored their automatic mode with F0.5, "because the objective of this scenario is to achieve high precision".

[Hand and Christen](https://link.springer.com/article/10.1007/s11222-017-9746-6) go further. They argue that the relative weight of precision and recall "should be an aspect of the problem and the researcher or user, but not of the particular linkage method being used". They treat it as a flaw of the F-measure that its effective weights shift with the method. That's why I'd report precision and recall beside any F-score. I'd also write the chosen β on the task card from part 2, next to its "Used by" line, before anyone has seen a result.

Measuring also answers a question that comes before any of this, which is whether to automate a task at all. If no threshold ever gets the telemetry task to the precision a coverage report needs, that's a measured reason to leave it to suggesting links, however good its recall.

## Measuring a reviewer's time

When a person confirms every link, the order of the suggestions matters as much as the set. A reviewer works down a ranked list. A resolver that puts the right test third saves more of their time than one that puts it thirtieth, even when a threshold would count both. Hayes, Dekhtyar and Sundaram's [study of methods](http://selab.netlab.uky.edu/homepage/publications/tse_jan_2006_as_appeared.pdf) measures this with lag, the number of wrong candidates ranked above a true link, averaged over the true links. It measures the reviewer's load with selectivity, the share of all possible pairs they had to look at. Rath and colleagues report recall within a short list, and found on average 96% of the right issues within three suggestions.

Say, as a made-up example, that a resolver shows the top three tests for each of the demo's 55 cases. That's 165 of the 8,030 pairs, a selectivity of about 2%. Suppose 56 of the 69 real links turn up among them, for a recall of 0.81 and a precision of 0.34. It could never have found all 69, because four of the cases have more than three tests each, and cardinality from part 2 caps a top-three list at 59.

The same study offered bands for judging such numbers, with recall of 80% or more counted as excellent and precision between 30% and 49% as good. That puts the made-up resolver in decent company. The authors called the bands their "first attempt to 'draw a line in the sand'", based on the first author's industrial experience "in performing and validating many traces". The methods they judged read textual documents, so I'd treat the bands as a starting point. A resolver reading structured names from OpenTofu plans ought to beat them comfortably.

In this way of working, the resolver's list is only half the story. What reaches the systems model is the list after a person has been through it, and part 1 described how people move what they're given in both directions. A simulation by [Hayes and colleagues](https://link.springer.com/article/10.1007/s00766-016-0260-8), which assumed analysts never make mistakes, still found that "the lowest effort does not always yield the highest quality matrix". So I'd measure the confirmed links as well as the suggested ones.

## What a useful gold set looks like

Here's the start of a gold set for part 2's OpenTofu task on the made-up pipeline, just after part 3's split:

| Input, as read from the plan | Expected link |
|---|---|
| `module.pipeline.aws_ecs_service.stage["ingest"]` | `PIPE-S1` |
| `module.pipeline.aws_ecs_service.stage["parse"]` | `PIPE-S2` |
| `module.pipeline.aws_ecs_service.stage["lex"]` | none, and flag the systems model as stale |

Each row is a real input in the form the resolver reads it, beside what a careful person decided it should get. The third row matters as much as the other two. A gold set that only lists links can't catch a resolver that invents one, and an object that ought to resolve to nothing needs saying so. Some inputs should get several links, and unless the gold set records that, the cardinality on the task card goes untested.

The demo's 69 tests fail as a gold set in two ways. They come from a single tool, and they're the ones that already carried keys. I'd build a gold set for each task, and put the hard cases in on purpose. For the made-up pipeline, that means `query-frontend` in the telemetry task's set after the rename, and `TestParserRejectsEmptyQuery` in the test task's.

It also has to be big enough, and small gold sets say less than they seem to. Precision is a proportion, and a proportion from a small sample comes with a wide interval. Take the [Wilson interval](http://www-stat.wharton.upenn.edu/~lbrown/Papers/2001a%20Interval%20estimation%20for%20a%20binomial%20proportion%20(with%20T.%20T.%20Cai%20and%20A.%20DasGupta).pdf), which Brown, Cai and DasGupta recommend for small samples. By that measure, an observed precision of 0.8 from 10 checked links is consistent with anything from about 0.49 to 0.94 at 95% confidence. From 100 links it narrows to about 0.71 to 0.87, and from 1,000 to about 0.77 to 0.82. Even if the demo's 69 Go-test links were a fair sample, a perfect recall on all of them would only show that recall is above about 0.95. Precision and recall have different denominators, the links proposed and the links that ought to exist, so each needs its own count. Going by those intervals, the demo's gold sets are too small to say much, and I'd rather admit that now than have a figure from them quoted later.

A bigger one is expensive. [Dekhtyar, Hayes and Antoniol](http://selab.netlab.uky.edu/homepage/publications/tefse-2007-dekhtyar-hayes-antoniol.pdf) wrote in 2007 that their experience of creating ground truth for even moderate-size datasets showed that "a very significant effort (perhaps unachievable by a single research group) is required". Eleven years later, [Dekhtyar and Hayes](https://arxiv.org/abs/1807.11454) gave a section of a paper the title "No one believes our ground truth".

## When the gold set is missing links

A resolver will sometimes find a link the gold set lacks, and a plain comparison scores it as wrong. That's likely wherever people built the gold set by hand. When few of a dataset's artefacts appear in any link, [Hey and colleagues](https://publikationen.bibliothek.kit.edu/1000140404/134392428) count among the possible reasons that "the gold standard is incomplete". Information retrieval met the same problem with relevance judgements. [Buckley and Voorhees](https://tsapps.nist.gov/publication/get_pdf.cfm?pub_id=150469) found that "current evaluation measures are not robust to substantially incomplete relevance judgments", and proposed a measure that counts only what someone judged.

For trace links, I'd keep three groups apart when scoring. A proposed link can be confirmed by the gold set, ruled out by it, or never judged at all. The third group needs a sample checked by a person, which is what Rath and colleagues did. Beside the pairs their classifier suggested, they had raters judge a group it hadn't, "added to mitigate evaluation bias". Every link a person confirms that way goes into the gold set, and the next evaluation is a little less wrong.

## A gold set that moves

My demo's gold set is less than a fortnight old and has already moved. In the first two days of its systems model, the Go-test links grew from 67 to 69 as new tests arrived. On the evening of 13 September, the manifest that ties each <span class="term" data-term="checkly">Checkly</span> check to a story went from 15 entries to 30 in 31 minutes. The systems model itself was still being written at the time, which only makes the point more strongly.

A score against last month's gold set describes last month's system. For a system that changes daily, I'd measure at every commit, against the gold set as it stood at that commit. The trend matters more than any single figure. Defect prediction research tests its prediction models the same way. [Bangash and colleagues](https://arxiv.org/abs/1911.06348) evaluate so that "models are trained only on the past, and evaluations are executed only on the future". They also warn that broad claims "might be contradicted by the next upcoming release". For trace links, [Rahimi and Cleland-Huang](https://link.springer.com/article/10.1007/s10664-017-9561-x) start from "a tendency for trace links to degrade over time as the system continually evolves". They then evaluated a tool that evolves links across 27 releases of the Cassandra database.

The gold set decays along with the links it describes, so it needs the same care. After part 3's split, the OpenTofu task's gold set gains the `lex` row above, and the telemetry task's gains `query-frontend` resolving to `PIPE-S2`. Without those updates, the next measurement marks the resolver down for being right.

Part 5, coming up next, looks at what else decides whether a resolver is worth running, once its links are good enough.

---

Previous: [Why no resolver gets every link right](15-why-no-resolver-gets-every-link-right.md) · Index: [Automating traceability](../README.md)

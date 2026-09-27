# Measuring a resolver

*Roar Elias Georgsen, 27 September 2026*

Part 4 of 7 in [Automating traceability](../README.md).

> [!IMPORTANT]
> **The story so far**
>
> [Part 3](15-why-no-resolver-gets-every-link-right.md) told correct, incorrect and missing links apart, and argued that every resolver proposes some incorrect links and misses some correct ones. Which mistake matters more depends on how the resolver is put to work. This part asks what to measure a resolver against, and with which numbers, when the set of correct links changes along with the system.

My demo's own <span class="term" data-term="systems-model">systems model</span> came with a ready-made test for a resolver, and I'm a little embarrassed by how well any resolver would do on it. Of its 56 <span class="term" data-term="verification-case">verification cases</span>, 32 name Go tests as evidence, 69 tests in all, and I added every one of those links by hand. Each of the 69 test names carries, straight after `Test`, the key of the requirement it's evidence for, so `TestSR22_SetAttributePatchesTextAndProjectionTogether` sits under the verification case for `SR-22`.

A resolver that reads the requirement key in test names would find all 69 and nothing else. That's a perfect score, and it would prove nothing, because the links were findable by the very rule being scored. The obvious next step is to hide the keys and see what a resolver makes of what's left. Even then, the 77 test functions without a key, which part 1 counted, are the ones a resolver is actually needed for. A score on the 69 says nothing about them, and nothing at all about <span class="term" data-term="opentofu">OpenTofu</span> resources or traces.

Links that someone has already made are a tempting thing to measure against anywhere, since the work is paid for. They're also a biased sample. [Bird and colleagues](https://www.cs.ucdavis.edu/~filkov/papers/biasbusters.pdf) looked at the bug-fix datasets that defect research relies on. They found that "only a fraction of bug fixes are actually labelled", and "strong evidence of systematic bias" in which ones. [Rath and colleagues](https://arxiv.org/abs/1804.02433) saw a hint of the same thing with trace links. Their classifier, trained on commits that developers had linked to issues, did worse on commits nobody had linked. For one project's improvements, precision fell from 0.32 to 0.11, although that second figure comes from a small hand-checked sample.

> [!NOTE]
> **Gold set**
>
> The correct links that a resolver's output is measured against, each one checked by a person. Traceability research calls it an answer set. It lists the inputs that should get no link as well as those that should, and it holds only for the versions of the systems model and the built system it was made from.

In the terms of part 1, a gold set is the join a federation would make if every object in the built system had carried a shared key from the start. Measuring a resolver means running it on every input the gold set covers and comparing what it proposes with what the gold set says, link by link. A proposed link the gold set holds counts as correct, one it lacks counts as incorrect, and a link in the gold set that the resolver didn't propose counts as missing. The rest is choosing which numbers to make of those counts, and building a gold set worth measuring against.

## Why accuracy flatters

Accuracy is the first number most people reach for, and for trace links it's close to useless. Take the demo's tests and verification cases as a grid. Any of the 146 test functions in the demo's own code could in principle verify any of the 56, which makes 8,176 pairs, and 69 of them are links. A resolver that links nothing at all classifies 8,107 of the 8,176 pairs correctly, for an accuracy of 99.16%, without finding a single link.

Record linkage reached the same verdict a long time ago. [Christen and Goiser](https://users.cecs.anu.edu.au/~christen/publications/qmdm2007christen.pdf) say plainly that accuracy "is therefore not a good quality measure for data linkage and deduplication, and should not be used". Possible pairs grow with the product of the two sides, while correct links grow roughly with their size, so on a system with more than a handful of parts the empty cells swamp everything else. Precision and recall, from part 3, leave the empty cells out. Precision looks only at the links a resolver proposed, and recall only at the links that ought to exist.

## Picking the F-score to suit the job

Two numbers are awkward when you want to compare resolvers, so people combine them into an F-score, a weighted harmonic mean of precision and recall that drops sharply when either is low. The weight is the part that needs a decision. In [van Rijsbergen's](http://www.dcs.gla.ac.uk/Keith/Chapter.7/Ch.7.html) formulation, the parameter β describes a user who attaches β times as much importance to recall as to precision. So F1 treats the two equally, while F2 cares more about missing links and F0.5 more about incorrect ones.

The demo's keyed tests can't show what that weight changes, because a rule that reads keys has no threshold to set. Part 3 gave example figures for a telemetry task on a bigger version of its hypothetical pipeline, where a name strategy links traced services to the pipeline's stages. Here they are again, at the task's two thresholds, with three F-scores added:

| Threshold | Precision | Recall | F0.5 | F1 | F2 |
|---|---|---|---|---|---|
| 0.8 | 0.92 | 0.58 | 0.82 | 0.71 | 0.62 |
| 0.5 | 0.60 | 0.90 | 0.64 | 0.72 | 0.82 |

F1 can barely tell the two settings apart. F0.5 clearly prefers the strict one, which is the setting I'd trust to add links on its own. F2 clearly prefers the loose one, which suits a team that suggests links for a person to confirm. The team's reviewer sees more of the correct links and rejects the incorrect ones cheaply. Rath and colleagues, whose two settings part 1 described, made the same choice and scored their automatic mode with F0.5, "because the objective of this scenario is to achieve high precision".

[Hand and Christen](https://link.springer.com/article/10.1007/s11222-017-9746-6) go further. They argue that the relative weight of precision and recall "should be an aspect of the problem and the researcher or user, but not of the particular linkage method being used". They treat it as a flaw of the F-measure that its effective weights shift with the method. That's why I'd report precision and recall beside any F-score. I'd also write the chosen β on the task card from part 2, next to its "Used by" line, before anyone has seen a result.

Measuring also answers a question that comes before any of this, which is whether to automate a task at all. If no threshold ever gets the telemetry task to the precision a coverage report needs, that's a measured reason to leave it to suggesting links, however good its recall.

## Measuring a reviewer's time

When a person confirms every link, the order of the suggestions matters as much as the set. A reviewer works down a ranked list. A resolver that puts a correct link third saves more of their time than one that puts it thirtieth, even when a threshold would count both. Hayes, Dekhtyar and Sundaram's [study of methods](http://selab.netlab.uky.edu/homepage/publications/tse_jan_2006_as_appeared.pdf) measures this with lag, the number of incorrect candidates ranked above a correct link, averaged over the correct links. It measures the reviewer's load with selectivity, the share of all possible pairs they had to look at. Rath and colleagues report recall within a short list, and found on average 96% of the correct issues within three suggestions.

Say, as an example, that a resolver shows the top three tests for each of the demo's 56 verification cases. That's 168 of the 8,176 pairs, a selectivity of about 2%. Suppose 56 of the 69 recorded links turn up among them, for a recall of 0.81 and a precision of 0.33. It could never have found all 69, because four of the verification cases have more than three tests each, and cardinality from part 2 caps a top-three list at 59.

The same study offered bands for judging such numbers, with recall of 80% or more counted as excellent and precision between 30% and 49% as good. That puts such a resolver in decent company. The authors called the bands their "first attempt to 'draw a line in the sand'", based on the first author's industrial experience "in performing and validating many traces". The methods they judged read textual documents, so I'd treat the bands as a starting point. A resolver reading structured names from OpenTofu plans ought to beat them comfortably.

In this way of working, the resolver's list is only half the story. What reaches the systems model is the list after a person has been through it, and part 1 described how people move what they're given in both directions. A simulation by [Hayes and colleagues](https://link.springer.com/article/10.1007/s00766-016-0260-8), which assumed analysts never make mistakes, still found that "the lowest effort does not always yield the highest quality matrix". So I'd measure the confirmed links as well as the suggested ones.

## What a useful gold set looks like

The demo's 69 tests fail as a gold set in two ways. They come from a single tool, and they're the ones that already carried keys. The demo has a second tool whose names carry none: an <span class="term" data-term="opentofu">OpenTofu</span> configuration that runs the demo on a Proxmox host, a server for virtual machines and containers, and publishes it through a Cloudflare tunnel. I left the systems model's keys out of its resource names on purpose. Here's the start of a gold set for linking those resources to the demo's systems model, from a plan with the demo switched on:

| Input, as read from the plan | Expected link |
|---|---|
| `proxmox_oci_image.demo[0]` | `demo` |
| `proxmox_virtual_environment_container.demo[0]` | `demo` |
| `proxmox_virtual_environment_container.tunnel[0]` | `hostDeployment.connector` |
| `cloudflare_dns_record.demo` | `TunnelEdge` |
| `data.cloudflare_zero_trust_tunnel_cloudflared_token.demo` | none |

Each row is an input in the form the resolver reads it, beside what a careful person decided it should get. In the demo's systems model, `demo` is the part that stands for the whole image, and `TunnelEdge` is the tunnel provider's edge, which answers for the public hostname.

The last row matters as much as the others. A plan holds things the systems model has no reason to hold, such as the credential the connector runs with. A gold set that lists only links can't catch a resolver that invents one for them. An input can also resolve to nothing because the systems model lacks an element, as part 3's new `lex` stage did, and then its row carries a flag that the systems model is stale.

The first two rows show cardinality from the element's side. The image and the container that runs it both belong to `demo`, and unless the gold set records both, the cardinality on the task card goes untested.

I'd build a gold set for each task, and put the hard inputs in on purpose. The connector on the host is one. The systems model has a second tunnel connector, the container the check session's compose file starts. A resolver that links the connector on the host to that one has found the correct kind of element and the incorrect instance. For the telemetry task on the hypothetical pipeline, the hard input is `query-frontend`, once the renamed parse service reports under that name.

It also has to be big enough, and small gold sets say less than they seem to. Precision is a proportion, and a proportion from a small sample comes with a wide interval. Take the [Wilson interval](http://www-stat.wharton.upenn.edu/~lbrown/Papers/2001a%20Interval%20estimation%20for%20a%20binomial%20proportion%20(with%20T.%20T.%20Cai%20and%20A.%20DasGupta).pdf), which Brown, Cai and DasGupta recommend for small samples. By that measure, an observed precision of 0.8 from 10 checked links is consistent with anything from about 0.49 to 0.94 at 95% confidence. From 100 links it narrows to about 0.71 to 0.87, and from 1,000 to about 0.77 to 0.82. Even if the demo's 69 Go-test links were a fair sample, a perfect recall on all of them would only show that recall is above about 0.95. Precision and recall have different denominators, the links proposed and the links that ought to exist, so each needs its own count. Going by those intervals, the demo's gold sets are too small to say much, and I'd rather admit that now than have a figure from them quoted later.

The same caution applies when one resolver is compared with another. Take an example. Two resolvers are scored on the demo's 69 tests with the keys hidden, and one puts the correct requirement first for 52 tests where the other manages 47. Five tests sounds like a win. The paired way to judge it, McNemar's test, looks only at the tests where the two disagree, since the tests on which both are correct, or both incorrect, say nothing about which resolver is better. Suppose they disagree on nine tests, seven going to the first resolver and two to the second. Even if the two were equally good, a split at least that uneven would turn up by chance about 18% of the time. In 1998 [Dietterich](https://doi.org/10.1162/089976698300017197) compared five tests for this job and concluded that "for algorithms that can be executed only once, McNemar's test is the only test with acceptable type I error". A type I error is a difference reported where there is none, and a resolver scored once against one gold set is in exactly that position.

A bigger gold set would narrow both figures, and it's expensive. [Dekhtyar, Hayes and Antoniol](http://selab.netlab.uky.edu/homepage/publications/tefse-2007-dekhtyar-hayes-antoniol.pdf) wrote in 2007 that their experience of creating ground truth for even moderate-size datasets showed that "a very significant effort (perhaps unachievable by a single research group) is required". Eleven years later, [Dekhtyar and Hayes](https://arxiv.org/abs/1807.11454) gave a section of a paper the title "No one believes our ground truth".

## When the gold set is missing links

A resolver will sometimes find a correct link the gold set lacks, and a plain comparison scores it as incorrect. That's likely wherever people made the gold set by hand. When few of a dataset's artefacts appear in any link, [Hey and colleagues](https://publikationen.bibliothek.kit.edu/1000140404/134392428) count among the possible reasons that "the gold standard is incomplete". Information retrieval met the same problem with relevance judgements. [Buckley and Voorhees](https://tsapps.nist.gov/publication/get_pdf.cfm?pub_id=150469) found that "current evaluation measures are not robust to substantially incomplete relevance judgments", and proposed a measure that counts only what someone judged.

For trace links, I'd score three groups separately. A proposed link can be confirmed by the gold set, ruled out by it, or never judged at all. The third group needs a sample checked by a person, which is what Rath and colleagues did. Beside the pairs their classifier suggested, they had raters judge a group it hadn't, "added to mitigate evaluation bias". Every link a person confirms that way goes into the gold set, and the next measurement starts from a more complete one.

The links a gold set rules out aren't all the same failure, either. Say a resolver, with the keys hidden, links `TestSR22_SetAttributePatchesTextAndProjectionTogether` to the <span class="term" data-term="adapter">adapter</span>, the part the demo's systems model says <span class="term" data-term="satisfy">satisfies</span> `SR-22`, and not to `SR-22` itself. Scored plainly, that's one incorrect link and one missing link, the same score as a link to a requirement the test has nothing to do with. Yet the link to the adapter is correct in part 3's sense, since the test lives in the adapter's code and exercises it. It isn't the link a coverage report needs, and it sits one relationship away from the one that is. A reviewer can repair it in a moment, where a link to a stranger means starting the search again.

I haven't found traceability research that scores these near misses apart, so take this as my own argument. Ontology matching has a precedent. [Ehrig and Euzenat](https://exmo.inria.fr/files/publications/ehrig2005a.pdf) point out that with classical precision and recall, correspondences that aren't exactly the expected ones "score zero", and they propose relaxed versions that give credit for being close. I'd count the proposed links that land within one or two relationships of the gold set's link in the systems model. That count goes beside precision and recall, with no partial credit folded into them, so the plain figures stay comparable across resolvers.

## A gold set that moves

My demo's gold set has moved more than once already. In the first two days of its systems model, the Go-test links grew from 67 to 69 as new tests arrived. On the evening of 13 September, the manifest that ties each <span class="term" data-term="checkly">Checkly</span> check to a story went from 15 entries to 30 in 31 minutes. The systems model itself was still being built at the time, which only makes the point more strongly.

A score against last month's gold set describes last month's system. For a system that changes daily, I'd measure at every commit, against the gold set as it stood at that commit. The trend matters more than any single figure. Defect prediction research evaluates its machine learning models the same way. [Bangash and colleagues](https://arxiv.org/abs/1911.06348) evaluate so that "models are trained only on the past, and evaluations are executed only on the future". They also warn that broad claims "might be contradicted by the next upcoming release". For trace links, [Rahimi and Cleland-Huang](https://link.springer.com/article/10.1007/s10664-017-9561-x) start from "a tendency for trace links to degrade over time as the system continually evolves". They then evaluated a tool that evolves links across 27 releases of the Cassandra database.

A gold set decays along with the links it describes, so it has to be maintained like them. The table above changed while I was writing it. Its connector row first read "none, and flag the systems model", because the only tunnel connector the systems model held was the check session's. I modelled the connector on the host, and the row became a link. A plan made with the demo switched off holds none of the first three rows, so the table holds only for a plan with the switch on. On the hypothetical pipeline, the telemetry task's gold set gains `query-frontend` resolving to `PIPE-S2` once the parse service is renamed. Without updates like these, the next measurement marks the resolver down for proposing a correct link.

[Part 5](17-what-makes-a-resolver-worth-running.md) looks at what else decides whether a resolver is worth running, once its links are good enough.

---

Previous: [Why no resolver gets every link right](15-why-no-resolver-gets-every-link-right.md) · Index: [Automating traceability](../README.md) · Next: [What makes a resolver worth running?](17-what-makes-a-resolver-worth-running.md)

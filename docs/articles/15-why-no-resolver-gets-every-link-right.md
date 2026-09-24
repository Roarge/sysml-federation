# Why no resolver gets every link right

*Roar Georgsen, 24 September 2026*

Part 3 of 6 in [Automating traceability](../README.md).

> [!IMPORTANT]
> **The story so far**
>
> [Part 2](14-the-parts-of-a-resolution-task.md) wrote a resolution task down as a card, from who uses the links to the version of each side a link is resolved against. However carefully that card is filled in, the resolver that carries it out will get some links wrong. This part asks what can fairly be expected of one.

Suppose the made-up pipeline from the first two parts has been running happily for a while. The task from part 2 has linked the parse stage, `PIPE-S2`, to the <span class="term" data-term="opentofu">OpenTofu</span> resource `module.pipeline.aws_ecs_service.stage["parse"]`, and the link is right. Then the team splits parsing in two. Tokenising moves into a new `lex` stage, and the list of stages in the configuration gains a fourth name. In the same change, the service that reported to <span class="term" data-term="opentelemetry">OpenTelemetry</span> as `query-parser` starts reporting as `query-frontend`. Nobody touches the <span class="term" data-term="systems-model">systems model</span>, so `PIPE-S2` is still one part.

Nobody edits a link either, and three of them change meaning anyway. The OpenTofu link still resolves, because the parse instance kept its address, but it now covers half of what the part used to do. In telemetry, the old link points at a service that has stopped sending anything. The new `query-frontend` traces resolve to nothing, since nothing in the service name says parse. Meanwhile the <span class="term" data-term="checkly">Checkly</span> check `parse-sustained-qps` goes on measuring throughput, now through two services where the systems model says one.

That's the decay from part 1, and it cuts both ways. [Mäder and Gotel](https://europepmc.org/article/MED/23471308) describe links that, left unmaintained, "get lost or represent false dependencies", so the same neglect produces missing links and wrong ones. Tools that keep trace matrices mark a link as suspect when either end changes. Cleland-Huang and colleagues' [review of the field](http://selab.netlab.uky.edu/homepage/publications/2014-ICSE-FOSE.pdf) notes that "it is not uncommon to see an industrial trace matrix populated with a high percentage of suspect links".

My own records managed it on a small scale. Two of the demo's decision records gave the same detail about how the compose input sets up subscriptions, and they agreed with each other. When I corrected the first on 27 August, they stopped agreeing. The second went on saying the old thing for 28 days, until 24 September, and nobody touched it in between. I'd like to say I noticed sooner. A link that was right once only tells you it was right then.

## How many are still right?

After a change like the split, the obvious question is how many of the pipeline's links are still right, and nobody is going to check every one of them by hand. A fixed benchmark won't settle it either. One built from last year's system decays along with the system, and a small one may never have looked much like the real thing.

Sampling does the job. Tepping proposed a way in 1968, as [Binette and Steorts](https://arxiv.org/abs/2008.04443) describe it. Group the pairs into classes by which of their fields agree, check a sample from each class by hand, and estimate the error rate class by class. Then split further the classes that add most to the expected cost. The estimate is only as good as the sample, and it has to be taken again as both sides change. Part 4 is about measuring properly, and it comes back to this.

## Names are written for people

A good enough resolver might seem able to get every link right, if only someone built it carefully. My demo suggests otherwise. Here are eight of the names in it that contain the word router:

- `Router`, the part definition in the demo's systems model
- `router`, the part in the same systems model that uses it
- ROUTER, the allocation element that definition's documentation names
- `sysml-federation-router`, the service name in the router's telemetry configuration
- `routerVersion`, an attribute holding the router's version number, 0.343.1
- `router-version`, the logical ID of a Checkly check
- "router: the model version", the same check's display name
- `CHK_RouterVersion`, the case in the systems model that describes that check

The first four are the router, in four different spellings, and the other four aren't. What makes this a trap is the pair in the middle. `routerVersion` holds the version of the router, while the check called `router-version` asks the demo for the version of the systems model it serves, through the router. A name strategy sees "router" in all eight and "version" in four. I chose every one of these names myself, which rather spoils the option of blaming anyone else.

Even a careful person can go round in circles on one of them. The check's own file calls it one of "the seven API checks on the router", and it sits in the router's group. The systems model's case for the same check gives its subject as the whole demo. It says the check verifies `SR-43`, which asks that one response carry a requirement's text, verdict and document number from a schema all three services contribute to. My two records disagree about what the check is about, and I wrote both.

Names in a built system are written for the people who work on it, and those people read them with context a resolver doesn't have. Text-based recovery has leaned from the start on the premise, in [Antoniol and colleagues'](https://doi.org/10.1109/TSE.2002.1041053) words, "that programmers use meaningful names for program items". Mostly they do, and the names mean something to whoever chose them. The 2014 review found that the gains from those methods "seem to have plateaued", mostly because of term mismatches between the documents being traced.

Resolvers still differ a lot. One that reads the `stage` tag on the OpenTofu resource gets parse right after the split, because the tag still says parse. One that compares service names is stuck the moment `query-frontend` appears. Any resolver can be improved, but never to the point where nobody needs to check, and that's why building one takes more work than most people expect.

Sometimes the fault isn't in the resolver at all. After the split, the new `lex` instance had nothing to resolve to, because the systems model has no `lex`. A resolver that trusts the systems model completely drops that instance without comment, and yet it's the most useful thing the run turned up. The systems model is out of date, and the unresolved object is the evidence.

Software architecture research met this in the 1990s. [Murphy and Notkin's](https://www.cs.ubc.ca/~murphy/papers/rm/rm-case-study.pdf) reflexion models compare a high-level diagram of a software architecture with the code. An engineer writes a map from the code to the diagram's boxes by hand, often with regular expressions over file names, and a tool then reports where the two agree and where they don't. In their case study, the first comparison found 15 convergences, 83 divergences and four absences. Where the code had an interaction the diagram lacked, the engineer updated the diagram. My own systems model has been the stale side too. When I read each of the demo's 36 images against the views in its systems model on 12 September, I found three things on the boards the systems model didn't yet hold. So an object that resolves to nothing is a report on the systems model as much as on the resolver, and whichever way of working you choose, it should reach a person.

## Fixing the case in front of you

The morning after the split, in this made-up case, someone notices that the `query-frontend` traces link to nothing and asks for them to be linked to the parse stage. It's a fair request, and there are two quick fixes.

The first is a rule saying that `query-frontend` is `PIPE-S2`. It fixes exactly the reported case and nothing else, so on today's links it costs nothing. It's also a key written by hand, the kind part 1 found holding my demo together in places, and it decays like any other hand-written link. The next time the team renames the service, the rule finds nothing and the miss comes back. A resolver patched one case at a time ends up as a list of hand-written links with a scoring function bolted on.

The second fix lowers the threshold on the name strategy until `query-frontend` scores high enough. Its cost is harder to see, and it needs two measures.

> [!NOTE]
> **Precision and recall**
>
> Two measures of the links a resolver proposes, taken against the links that ought to exist. Precision is the share of proposed links that are right, so every false positive lowers it. Recall is the share of the links that ought to exist that the resolver found, so every false negative lowers it.

Say the telemetry task on a bigger version of the made-up pipeline ought to produce 40 links, and its name strategy scores each candidate between 0 and 1. The numbers below are made up to show the shape of the trade.

| Threshold | Links proposed | Of which right | Precision | Recall |
|---|---|---|---|---|
| 0.8 | 25 | 23 | 0.92 | 0.58 |
| 0.5 | 60 | 36 | 0.60 | 0.90 |

Lowering the threshold from 0.8 to 0.5 to catch one reported miss brings in 35 more links. Of those, 13 are right and 22 are wrong, and nobody will report the 22, because the person who asked for the fix was looking at something else. It works the other way too. Raise the threshold to get rid of a reported wrong link, and right links go with it. Fixes like these pile up, each tuned to a case somebody happened to notice, and each needing testing against everything the resolver already does. The cases nobody noticed pay for them. The thing to aim for is the best balance across all the links, and which balance is best depends on who pays for each kind of mistake.

## What a wrong link costs

Suppose `PIPE-R1.2`, the throughput the parse stage must sustain, ended up linked to a made-up latency check called `parse-p95-latency`. A coverage report would show the requirement verified, by a check that never measures throughput. A missing link would have shown it unverified, which is a gap somebody would ask about. I haven't found a source that says this in so many words, so take it as my argument. It fits what [Mäder and colleagues](https://doi.org/10.1109/MS.2013.60) report of the traceability sent to regulators, which "is often weak, casting doubt rather than confidence".

While a person is reviewing candidates, it's the other way round. A wrong candidate takes a moment to reject, and part 1 quoted Dekhtyar and Hayes on how much faster that is than finding an omission. They go on to call recall in tracing ["significantly more important than precision"](https://arxiv.org/abs/1807.11454). [Rodriguez and colleagues](https://arxiv.org/abs/2306.10972) add that "many safety-critical domains require near-perfect recall when using NLP techniques to automatically generate links, which can not be consistently achieved". In practice that pushes safety work towards suggesting links for a person to confirm.

Record linkage has a formal version of the argument. Tepping's decision rule, in Binette and Steorts's account, gives each action one cost when the pair is a true link and another when it isn't, and picks the action with the lowest expected cost. Unequal costs move the thresholds. So my position is part 1's, with a reason attached. When a person will see every candidate, favour recall. When links go straight into anything a coverage report or a safety case reads, favour precision, because that's where a wrong link hides.

The person seeing every candidate isn't a perfect filter, either. Part 1 reported that analysts improve poor candidate sets and decide worse on good ones, and Cleland-Huang and colleagues add that human feedback on trace links "is incorrect approximately 25% of the time". Continuous resolution brings a trap of its own. A resolver that runs on every change will want to show a reviewer only what changed since the last run, which saves a great deal of effort. The same review reports that showing analysts only new or changed links "negatively impacts the quality" of the links. A resolver that works this way should at least say when each link was last seen by a person in its full context.

## Letting a resolver add links on its own

Now say someone adds `lex` to the systems model, and the resolver is sure the new OpenTofu instance belongs to it. I'd let it add that link without asking anyone, under a few conditions. A blanket rule against it compares automatic links with a perfect hand-made set that doesn't exist. [Rath and colleagues](https://arxiv.org/abs/1804.02433) found that on average about 40% of commits carried no issue key, in projects that largely followed the practice of tagging every one. The links already in any systems model came from people and from whatever tools those people used, and nobody vetted those either.

The first condition is that every link says where it came from. Removal depends on it, since you can't clean up after a bad run you can't find. [Guo and colleagues](https://arxiv.org/abs/2405.10845) argue that maintenance has to handle "a mixture of automatically generated and manually generated trace links and leave the manually created ones untouched", and part 5 comes back to this under the name provenance.

The second is that removing a wrong link is cheap, and that the removal is remembered. Otherwise a resolver that runs on every change adds the same wrong link back on its next run, and the person who removed it soon stops bothering. A rejected link is a result, and it needs keeping as carefully as an accepted one. The third is that the automatic mode favours precision, for the reasons above. My fourth, which I haven't seen written down elsewhere, is that a link goes suspect the moment either of its ends changes, and stays suspect until something checks it again.

Taken together, those conditions make a link more than yes or no. It has a state, such as proposed, accepted, rejected, added automatically or suspect, and a score that says how sure the resolver was. A link with a state and a score can be trusted as far as its history supports, which is further than a bare cell in a matrix.

## More links aren't better

A resolver judged by how many links it adds will add too many. [Heindl and Biffl](https://doi.org/10.1145/1081706.1081717) found that tracing requirements by their value took around 35% of the effort of tracing all of them in full. The risky and volatile requirements were the ones that warranted more detail. At a 2015 [Dagstuhl seminar](https://drops.dagstuhl.de/storage/04dagstuhl-reports/volume05/issue04/15162/DagRep.5.4.76/DagRep.5.4.76.pdf) on traceability, Cleland-Huang argued that the traceability certifiers prescribe "tends to be overly extensive". For the parse stage, the link that earns its keep runs from `PIPE-R1.2` to the check that measures throughput. A link from `PIPE-S2` to every log line that mentions parsing adds nothing a search couldn't.

[Part 4](16-measuring-a-resolver.md) is about measuring a resolver against a set of right answers that changes along with the system.

---

Previous: [The parts of a resolution task](14-the-parts-of-a-resolution-task.md) · Index: [Automating traceability](../README.md) · Next: [Measuring a resolver](16-measuring-a-resolver.md)

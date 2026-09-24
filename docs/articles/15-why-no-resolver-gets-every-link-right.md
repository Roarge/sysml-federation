# Why no resolver gets every link right

*Roar Georgsen, 24 September 2026*

Part 3 of 6 in [Automating traceability](../README.md).

> [!IMPORTANT]
> **The story so far**
>
> [Part 2](14-the-parts-of-a-resolution-task.md) wrote a resolution task down as a card of eleven decisions, from the search space to the version of each side that a link is resolved against. However carefully the card is filled in, the resolver that carries it out will get some links wrong. This part asks what can fairly be expected of a resolver once that's accepted.

## The parse stage, a week later

Suppose the made-up pipeline has been running for a while, and the task from part 2 has linked `PIPE-S2` to the <span class="term" data-term="opentofu">OpenTofu</span> resource `module.pipeline.aws_ecs_service.stage["parse"]`. The link is right. Then the team splits parsing in two. Tokenising moves into a new `lex` stage and the `for_each` set gains a fourth name. In the same change, the service that reported to <span class="term" data-term="opentelemetry">OpenTelemetry</span> as `query-parser` starts reporting as `query-frontend`. Nobody touches the model, so `PIPE-S2` is still one part.

Nobody edits a link either, and three of them change meaning anyway. The OpenTofu link still resolves, because the parse instance kept its address, but it now covers half of what the part used to do. In telemetry, the link points at a service that has stopped sending spans. The new `query-frontend` spans resolve to nothing, since nothing in the name says parse. And the <span class="term" data-term="checkly">Checkly</span> check `parse-sustained-qps` goes on measuring throughput, now through two services where the model says one.

Traceability research calls this decay, and it cuts both ways. [Mäder and Gotel](https://europepmc.org/article/MED/23471308) describe links that, without maintenance, "get lost or represent false dependencies", a missing link and a wrong one from the same cause. Tools that keep trace matrices mark a link suspect when either end changes. Cleland-Huang and colleagues' [review of the field](http://selab.netlab.uky.edu/homepage/publications/2014-ICSE-FOSE.pdf) notes that "it is not uncommon to see an industrial trace matrix populated with a high percentage of suspect links".

My own records show it on a small scale. Two of the demo's decision records gave the same wrong detail about how the compose input sets up subscriptions. I corrected the first on 27 August. The second went on saying the old thing until 24 September, in a demo whose whole argument is that its records should agree.

A link that was right once is evidence that it was right then, and nothing more.

## Fixing the case in front of you

The next morning, in this made-up case, someone notices that the `query-frontend` traces link to nothing and asks for them to be linked to the parse stage. It's a fair report, and there are two quick fixes. One lowers the threshold on the name strategy until `query-frontend` scores high enough. The other adds a rule saying that `query-frontend` is `PIPE-S2`. Both close the report, and both cost something the report won't show.

The cost of the first takes two measures to see.

> [!NOTE]
> **Precision and recall**
>
> Two measures of a set of proposed links, taken against the links that ought to exist. Precision is the share of proposed links that are right, so every wrong link lowers it. Recall is the share of the links that ought to exist that were proposed, so every missing link lowers it. Loosening a resolver usually raises recall at the cost of precision, and tightening it does the reverse.

Say the pipeline, built out in full, has 40 links that ought to exist between model elements and objects in the built system, and the name strategy scores its candidates between 0 and 1. The numbers below are made up to show the shape of the trade.

| Threshold | Links proposed | Of which right | Precision | Recall |
|---|---|---|---|---|
| 0.8 | 25 | 23 | 0.92 | 0.58 |
| 0.5 | 60 | 36 | 0.60 | 0.90 |

Lowering the threshold from 0.8 to 0.5 to catch one reported miss brings in 35 more links. Of those, 13 are right and 22 are wrong, and none of the 22 will be reported by the person who asked for the fix, because they were looking at something else. Fixes like this also pile up. Each is tuned to a case somebody noticed, and the cases nobody noticed pay for it.

The rule is subtler. It fixes exactly the reported case and nothing else, so neither measure gets worse. It's also a key written by hand, of the kind part 1 found holding the demo together in places, and it decays like any other hand-written link. When the team renames the service again, the rule finds nothing and the miss comes back. A resolver patched one case at a time ends up as a list of hand-written links with a scoring function attached, which is more or less where my demo started.

What's worth aiming for is the best balance over all the links, and which balance is best depends on who pays for each kind of mistake.

## Names are written for people

It's tempting to think a good enough resolver would get every link right, and that the failures above come from a poor one. The demo says otherwise, and I wrote all of it. Eight names in it contain the word router:

- `Router`, the part definition in the model
- `router`, the part in the model that uses it
- ROUTER, the allocation element that definition's documentation names
- `sysml-federation-router`, the service name in the router's telemetry configuration
- `routerVersion`, an attribute of the router that holds its version number
- `router-version`, the logical ID of a Checkly check
- "router: the model version", the same check's display name
- `CHK_RouterVersion`, the case in the model that describes that check

The first four are the router. The other four are about something else, and the check named after the router is evidence for `SR-43`, that one response answers a query with types from all three subgraphs. A name strategy sees "router" in all eight. I wrote the telemetry configuration myself, and the model still never mentions the name I gave the router there.

Names in a built system are written for the people who work on it, and those people read them with context a resolver doesn't have. Text-based recovery has leaned from the start on the premise, in [Antoniol and colleagues'](https://doi.org/10.1109/TSE.2002.1041053) words, "that programmers use meaningful names for program items". Mostly they do, and the names mean something to the people who chose them. The 2014 review found the gains from those methods had "plateaued", mostly because of term mismatches between the documents being traced.

Neither side is clean, either. A model is seldom complete and current at once, and the built system has its own noise, such as a made-up Java service that reports as `unknown_service:java` because nobody set its name. Telling a random difference from a real one would take knowing both sides almost completely, and anyone who knew that much wouldn't need a resolver. Some cases defeat a careful person too. Part 1's `TestParserRejectsEmptyQuery` could belong to the stage or to the query language, and I can't tell without reading its body.

So some wrong links and some missing ones are certain, whatever the resolver. That doesn't make resolvers equal. Some get far more right than others, and any of them can be improved, only never to the point where nobody needs to check.

## When the model is the stale side

Back in the split, the new `lex` instance had nothing to resolve to, because the model has no `lex`. A resolver that trusts the model completely drops that instance without comment, and yet it's the most useful thing the run turned up. The model is out of date, and the unresolved object is the evidence.

Software architecture research met this in the 1990s. In [Murphy and Notkin's](https://www.cs.ubc.ca/~murphy/papers/rm/rm-case-study.pdf) reflexion models, an engineer maps source code onto a high-level model with regular expressions over names, which is entity resolution by hand, and a tool reports where the two agree and where they don't. In their case study, the first comparison found 15 convergences, 83 divergences and four absences. Where the code had an interaction the model lacked, the engineer changed the model.

My own model has been the stale side too. When I read each of the demo's 36 images against the model's views on 12 September, I found three things on the boards that the model didn't yet hold. An object that resolves to nothing is a report on the model as much as on the resolver, and both ways of working from part 1 should pass it to a person, whatever they do with the links.

## What a wrong link costs

Precision and recall are often reported side by side, as though the two kinds of error cost the same. Which costs more depends on when it's found.

While a person reviews candidates, a wrong one is cheap. It takes a moment to reject, and part 1 quoted Dekhtyar and Hayes on how much faster that is than finding an omission. They go on to call recall in tracing ["significantly more important than precision"](https://arxiv.org/abs/1807.11454). Safety work pushes the same way, and [Rodriguez and colleagues](https://arxiv.org/abs/2306.10972) note that "many safety-critical domains require near-perfect recall".

Once a link is accepted, the costs flip. Suppose `PIPE-R1.2`, the throughput the parse stage must sustain, ended up linked to a made-up latency check called `parse-p95-latency`. A coverage report would show the requirement verified, by a check that never measures throughput. A missing link would have shown it unverified, which is a gap somebody would ask about. I haven't found a source that says this in so many words, so take it as my argument. It fits what [Mäder and colleagues](https://doi.org/10.1109/MS.2013.60) report of the traceability sent to regulators, which "is often weak, casting doubt rather than confidence".

Record linkage has a formal version of this. Tepping's decision rule, as [Binette and Steorts](https://arxiv.org/abs/2008.04443) describe it, gives each action one cost when the pair is a true link and another when it isn't, and picks the action with the lowest expected cost. Unequal costs move the thresholds. So my position is the one part 1 set out, now with the reason attached. When a person will see every candidate, favour recall. When links go straight into anything a coverage report or a safety case reads, favour precision, because that's where a wrong link hides.

## A person in the loop

A person checking every candidate sounds like the way to put things right. Part 1 showed the limit of that, since analysts improve poor candidate sets and make good ones worse. Cleland-Huang and colleagues add that human feedback on trace links "is incorrect approximately 25% of the time".

Continuous resolution brings a trap of its own. A resolver that runs on every change will want to show a reviewer only what changed since the last run, which saves a great deal of effort. The same review reports that showing analysts only new or changed links "negatively impacts the quality" of what they decide. A resolver that works this way should at least say when each link was last seen by a person in its full context.

## Links the resolver adds on its own

The opposite instinct is never to let a resolver add a link that no person has seen. Where a wrong link costs most, that's a reasonable rule. As a general rule, it compares automatic links with a perfect hand-made set that doesn't exist. Part 1 found 77 of the demo's 146 top-level Go test functions joined to nothing, and Rath and colleagues found about 40% of commits unlinked in projects that meant to link them all. Adding links at the made-up 0.92 precision above can leave a set like that better than it was.

I'd let a resolver add links on its own under three conditions.

1. The automatic mode favours precision, for the reasons above.
2. Removing a wrong link is cheap, and the removal is remembered. Otherwise a resolver that runs on every change adds the same wrong link back on its next run, and the person who removed it soon stops bothering. A rejected link is a result, and it needs keeping as carefully as an accepted one.
3. Every link says where it came from. [Guo and colleagues](https://arxiv.org/abs/2405.10845) make it a requirement of maintenance, which has to handle "a mixture of automatically generated and manually generated trace links and leave the manually created ones untouched". Part 5 comes back to this as provenance.

Those conditions make a link more than yes or no. It has a state, such as proposed, accepted, rejected, added automatically, or suspect after one of its ends changed, and a score that says how sure the resolver was. A link with a state and a score can be trusted as far as its history supports, which is further than a bare cell in a matrix.

## More links

A resolver judged by how many links it adds will add too many. [Heindl and Biffl](https://doi.org/10.1145/1081706.1081717) found that tracing requirements by their value took around 35% of the effort of tracing all of them in full. The risky and volatile requirements were the ones that warranted more detail. In 2015 a [Dagstuhl seminar](https://drops.dagstuhl.de/storage/04dagstuhl-reports/volume05/issue04/15162/DagRep.5.4.76/DagRep.5.4.76.pdf) on traceability found that prescribed traceability "tends to be overly extensive". For the parse stage, the link that earns its keep runs from `PIPE-R1.2` to the check that measures throughput. A link from `PIPE-S2` to every log line that mentions parsing adds nothing a search couldn't.

## How good the links are

It's sometimes said that nobody can know how good a resolver's links are without checking every one, which nobody will. A fixed benchmark doesn't settle it either, because a benchmark built from last year's system decays along with the system. Sampling does work. Record linkage has used it since Tepping in 1968, as Binette and Steorts describe. A person checks a sample of pairs in each score band, the error rate of each band is estimated from the sample, and review goes where the estimate is worst. The estimate is only as good as the sample, and it has to be taken again as both sides change.

Part 4, coming up next, is about measuring a resolver against a set of right answers that changes along with the system.

---

Previous: [The parts of a resolution task](14-the-parts-of-a-resolution-task.md) · Index: [Automating traceability](../README.md)

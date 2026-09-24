# From use cases to requirements

*Roar Georgsen, 27 August 2026*

Part 6 of 13 in [Federating a systems model](../README.md), written for release v0.3.0.

> [!IMPORTANT]
> **The story so far**
>
> The demo is one container. It runs a SysML v2 adapter, a capacity service and a document service as three GraphQL subgraphs behind a single router, with a model viewer and a requirements document as two web apps in front. The [twelve use cases](04-twelve-use-cases-and-one-moving-bottleneck.md) say what a visitor sees. This part covers the second gate, which turned them into requirements the code can be tested against, and the capacity model at the heart of the demo.

## Forty-five requirements

The second gate produced the requirements, the <span class="term" data-term="traceability">traceability</span> between them and the use cases, and a page on the capacity model. The requirements are written in EARS form.

> [!NOTE]
> **EARS**
>
> The Easy Approach to Requirements Syntax, a small set of sentence patterns for writing requirements. There are five: ubiquitous, event-driven, state-driven, unwanted behaviour and optional feature, plus a complex form that combines two of them. An event-driven one from this demo reads: "When the model version or the document version changes, each app shall reflect the change within two seconds and without a page reload."

There are forty-five requirements, each with one statement, a rationale, the use cases it traces up to, and a verification method, plus seven design constraints written as plain statements. The traceability is one table per hop, generated once from the requirements and kept by hand after that, and it runs in both directions. Every use case traces to at least one requirement, and every requirement either traces to a use case or names the obligation it comes from, such as a licence.

Most requirements are verified by a Go test named after them. Behaviour that lives in the browser is **demonstrated** instead: a scripted checklist, run by hand and recorded in the example's verification record, because nothing in the repository drove a browser at the time. That's a smaller promise than a test, and the requirements say so.

Model requirements, the ones inside the example model such as `PIPE-R1`, are demo content, not obligations on this code. They keep their SysML short names.

## Three holes in the design

I read the documents against their sources and against each other. Three things that came back were holes in the design, not the wording, and each went back to the brief as a decision.

**Which quantity a requirement constrains.** A verdict is the capacity service's answer for one requirement: PASS, FAIL or INCONCLUSIVE, with a reason beside it. Both `PIPE-R1` (throughput) and `PIPE-R2` (latency) have the pipeline as their subject. So a verdict rule keyed on the subject couldn't tell them apart, and would have passed the latency requirement at 1200 queries per second against a limit of 200 ms. The fix was to read each requirement's own constraint, such as `target.capacity >= requiredRate`, as a quantity, a comparison and a limit. The capacity service judges only requirements whose quantity it computes, and says INCONCLUSIVE for the rest. Reading the constraint, instead of looking for an agreed attribute name such as `requiredRate`, keeps the adapter free of any convention taken from the example ([quantity read from the constraint](../decisions/AD-0008-quantity-from-constraint.md)).

**Which way a connection points.** A SysML v2 `connect` has two ordered ends and no direction of its own, and the capacity calculation needs to know which way queries go. The direction is now the order of the ends, and the adapter refuses to start if a connection's first end isn't an output port or its second isn't an input. A wiring mistake becomes a model error with a line number, instead of a wrong capacity with nothing to flag it ([connection direction from the order of the ends](../decisions/AD-0009-connection-direction.md)).

**What a reason may say.** The reason beside every verdict is built from fixed templates that carry no word of the model, because the analysis can't know, for instance, that a requirement is derived ([reasons built from templates](../decisions/AD-0024-reason-templates.md)).

## The capacity model

The capacity of the pipeline is the largest query rate the wiring can carry from the servers that receive queries to the servers that deliver results. For a chain of stages, that's the slowest one. For stages split across parallel servers, it's their sum. That's the version you can check by hand: min(2000, 1200, 700 + 700, 1800) = 1200.

The capacity service computes it a more general way, as a [maximum flow](../decisions/AD-0007-rollup-as-maximum-flow.md). Each server becomes an edge limited to its throughput, each connection an edge with no limit, and the capacity is the most that can flow from the start of the pipeline to the end.

[![The shipped pipeline as a flow network. Each server is split into an in-node and an out-node joined by an edge whose capacity is its throughput: ingest 2000, parse 1200, indexA 700, indexB 700, serve 1800. Connections become edges with no limit. A super-source feeds ingest and serve feeds a super-sink. The minimum cut is the parse edge, at 1200.](../figures/flow-network.svg)](../figures/flow-network.svg)

*The shipped wiring as the capacity service sees it. The cheapest set of server edges that cuts every path is parse, at 1200.*

Flow and the hand rule agree wherever the hand rule applies, and a test checks that they do. Flow still answers where the hand rule has nothing to say, such as wiring that forks and joins in odd places. The bottleneck falls out of the same calculation. It's the minimum cut, the cheapest set of servers whose removal would cut every path, and where several tie the service reports the one nearest the entry.

Each requirement then gets a verdict: PASS or FAIL against its limit, INCONCLUSIVE when it constrains a quantity the service doesn't compute, and ERROR when a server's value is missing or negative. The four words are the ones the SysML v2 Systems Library uses. A reason comes with each verdict, such as `capacity 1200 against 1500, limited by parse`, where "capacity" and "parse" come from the service's configuration and the model and not from the template.

The derived limits come from the model: the full 1500 for each server in the chain, and half each for the two index servers, which share the load. Here's the whole script with every verdict:

| State | Capacity | `PIPE-R1` | Cut | Derived: ingest, parse, indexA, indexB, serve |
|---|---|---|---|---|
| Shipped | 1200 | FAIL | parse | PASS, FAIL, FAIL, FAIL, PASS |
| `ingest` to 3000 | 1200 | FAIL | parse | PASS, FAIL, FAIL, FAIL, PASS |
| `parse` to 1700 | 1400 | FAIL | indexA, indexB | PASS, PASS, FAIL, FAIL, PASS |
| then `indexA` to 900 | 1600 | PASS | indexA, indexB | PASS, PASS, PASS, FAIL, PASS |

The last row is the derived failure: the pipeline passes while indexB fails its share. That's how allocated requirements behave wherever a budget is split over parts, and the shipped document explains it in its opening paragraph.

> [!WARNING]
> **A demo model, not a planning tool**
>
> The arithmetic is deliberately simple. It assumes, among other things, that work splits evenly across parallel servers, load balancing is perfect and nothing ever queues. That's enough to show where a number gets computed in a federated graph, which is the point of the demo. It isn't a model of any real pipeline, and it isn't meant for real-world use ([an idealised capacity model](../decisions/AD-0006-idealised-capacity-model.md)).

The capacity service never sees SysML. On every query that asks for a capacity or a verdict, the router hands it the parts, their values and the connections, and it recomputes from scratch each time, so it can't go stale. The full capacity model page, with the flow network, the rule for the reported cut, the verdict precedence, every reason template, the assumptions and the edge cases, lives beside the code in [the example's README](https://github.com/Roarge/sysml-federation/blob/main/examples/pipeline/README.md#the-capacity-model).

---

Previous: [Twelve use cases and one moving bottleneck](04-twelve-use-cases-and-one-moving-bottleneck.md) · Index: [Federating a systems model](../README.md) · Next: [Five views and twenty-six decisions](06-five-views-and-twenty-six-decisions.md)

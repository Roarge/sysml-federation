# From use cases to requirements

*Roar Georgsen, 27 August 2026*

Part 6 of 13 in [Federating a systems model](../README.md), written for release v0.3.0.

> [!IMPORTANT]
> **The story so far**
>
> The demo is one container. It runs a SysML v2 adapter, a capacity service and a document service as three GraphQL subgraphs, each a service that owns part of a shared schema. They sit behind a single router, the process that composes those parts into one graph and answers every query against it. A model viewer and a requirements document are two web apps reading the one graph. The [twelve use cases](04-twelve-use-cases-and-one-moving-bottleneck.md) say what a visitor does with it. This part covers the approval gate after them, the review the requirements had to pass before the architecture began, which says what the code owes the visitor in return.

## Four documents

The gate produced four documents.

**A constraints card of 95 entries**, compiled from the research reports and the repository's own policy files. Each entry states a fact or a limit in one to three sentences. Then it gives its source, with a confidence in brackets: verified, likely or unverified, as the research assigned it. The bracket can also say "repository policy" where the source is the repository's own configuration, "decision" where the entry records a choice the design made and not a fact, and "editorial" for a wording choice. Last comes the decision it serves. For an unverified item the design leans on, it names a spike instead, to be settled early in the build before anything depends on it.

**A capacity model page**, stating what the service computes and how far the number can be trusted.

**Forty-five requirements in EARS form**, so each statement takes one of its patterns.

> [!NOTE]
> **EARS**
>
> The Easy Approach to Requirements Syntax, a small set of sentence patterns for writing requirements. There are five: ubiquitous, event-driven, state-driven, unwanted behaviour and optional feature, plus a complex form that combines two of them. An event-driven one from this demo reads: "When the model version or the document version changes, each app shall reflect the change within two seconds and without a page reload."

Each requirement carries one statement with one system name. It also has a rationale citing the decisions and constraints behind it, a verification method, the use cases it traces up to, and the elements it is allocated to. The seven design constraints that follow the requirements carry a plain statement instead, because EARS suits behaviour and not architectural limits.

Requirements are allocated to nine elements:

- the adapter
- the capacity service
- the document service
- the router, with its committed execution configuration
- the model viewer
- the requirements document app
- the image, meaning the supervisor binary, the one UI server that serves both apps and proxies the router, and the publishing workflow
- the example model file
- the repository's build, tests and continuous integration

Verification is by one of four methods:

- **Test** means a Go test named after the requirement, run by the repository's full check.
- **Analysis** and **Inspection** are what they sound like.
- **Demonstration** means a scripted checklist run in a local browser or on a Docker host, and recorded in the example's own verification record.

The checklist is a manual one because nothing in the repository drives a browser, and the language and licence constraint permits no dependency that does. So behaviour that lives in browser JavaScript is demonstrated rather than tested. That's a smaller promise, and it's stated as one.

**The traceability**, the fourth document, is generated from the requirements and checked in both directions. Every one of the twelve use cases traces to at least one requirement. Thirty-eight of the forty-five requirements trace up to a use case. The other seven trace to an obligation the project sets itself, whether in its stated goals, in the brief or in a licence, and the tables say so instead of leaving a blank. Every requirement names a verification method. Once the [decision records](../decisions/README.md) existed, the same tables gained a column showing that every requirement is affected by at least one of the twenty-six.

Model requirements, the ones that live inside the example model such as `PIPE-R1`, are demo content and not obligations on this code. They keep their SysML short names.

## Three findings that reached back into the design

I read the four documents against their sources and against each other. Three of the things that came back were holes in the design, not defects in the wording. Each went back to the brief as a decision before the documents were revised.

### Which quantity a requirement constrains

The example has two requirements on the pipeline, `PIPE-R1` on throughput and `PIPE-R2` on latency, and both have the pipeline as their subject. A verdict is the answer returned for one requirement: PASS, FAIL, INCONCLUSIVE or ERROR. Every document had assumed a verdict rule keyed on the subject. That rule couldn't tell the two requirements apart, and would have passed the latency requirement at 1200 against 200.

The same pass found a second gap in the same place. No document said how an adapter with nothing in it specific to one model would know which of a requirement's attributes is its limit.

One rule answers both. The adapter reads each requirement's `require constraint` as a comparison. On one side is a feature chain rooted at the requirement's subject, a dotted path that starts at the subject and ends at the quantity. On the other is either an attribute of the requirement or a literal. The chain's last segment is projected as the constrained quantity, the operator as the comparison, and the other operand's value as the limit. In the example that gives `<subject>.capacity >= requiredRate` for the six throughput requirements, and `pipeline.latency <= maxLatency` for the latency one.

A constraint of any other shape is refused at start with file, line and column, the same treatment any construct outside the supported subset gets. The capacity service then evaluates only requirements whose quantity is the one it's configured to compute, and returns INCONCLUSIVE for every other.

The alternative was a naming convention on the requirement's attributes, an agreed name such as `requiredRate` for the adapter to look for. It lost because the convention would have been an assumption taken from the example, and the whole claim is that the adapter is specific to no model. The record is [quantity, comparison and limit read from the constraint](../decisions/AD-0008-quantity-from-constraint.md).

### Which way a connection points

A SysML v2 `connect` has two ordered ends and no direction of its own. The rollup, a maximum flow over the wiring, needs directed edges, and the direction had been assumed and never stated.

It is now the end order, from the first end to the second. The adapter refuses to start on a connect whose first end isn't an `out` port, or whose second isn't an `in` port. The end order is the source of the direction and the port declarations are the check. So a wiring error becomes a model error with a position, instead of a wrong capacity with nothing to flag it. The record is [connection direction from the order of the ends](../decisions/AD-0009-connection-direction.md).

### Reasons, and where the UI server lives

The third settled two smaller things at once. Every verdict's reason string is built from fixed templates that carry no word of the model, and the record is [verdict reasons built from templates](../decisions/AD-0024-reason-templates.md). The single UI server that serves both apps and proxies the router belongs to the image element. So the two apps are HTML, CSS and JavaScript and nothing else, which [one binary, one port](../decisions/AD-0011-one-binary-one-port.md) records.

## What else the review changed

Nine requirements named two systems or two obligations in one statement, and were split. With the four added below, the set reached forty-five, above the plan's target of 20 to 25. I accepted the overrun for that reason.

Four use case criteria had no requirement behind them, and each gained one:

- the viewer showing a verdict and its reason
- document edits leaving the model version unchanged
- the apps referencing no resource at another origin
- short names as identifiers

The last of those matters more than it looks. The short name is the entity key, the field every service uses to say "this requirement", so that the router can join what three subgraphs know about it.

Originally, the requirement against outbound network traffic listed the four environment variables that silence the router's telemetry in its rationale. They are now the obligation itself. Its demonstration runs the container with the network removed at debug log level and looks for connection attempts. The router's usage tracker fails quietly, so readiness on its own would prove nothing.

The allowlist constraint had listed two gaps in the repository's tracking rules that weren't gaps, and missed the two that were. First, the committed router configuration and its compose input have to sit at `examples/pipeline/` for the existing one-level rules to reach them. Second, the one vendored JavaScript file has to stay out of any directory named `vendor/`.

<details markdown="1">
<summary>Under the bonnet: smaller fixes to the capacity page and the constraints card</summary>

On the capacity page:

- The explanation of the tie the example avoids had placed the tie in the wrong state. It now says where it would have arisen.
- The edge cases were brought into line with the verdict rules.
- Flow conservation, and inferring the entry and exit servers from the wiring, were added to the assumptions. Both had been used but not stated.
- The limits paragraph no longer claims both "upper bound" and "invalidated by any violation" of the same number.

On the constraints card:

- The OMG training folders for ports and connections are 09 and 10, not 14 and 16.
- The function that fills in a subgraph's required fields takes a map, not a slice of maps.
- Three references to the plan's candidate requirement numbers now point at the real identifiers.

</details>

A second cross-document pass over the revised set found two gaps that the first fix had opened. Two changes closed them, both described below. One was the abstract part definition that lets `Server` declare `capacity` as well as `throughput`, and the other was a reordered verdict precedence.

The same pass made four smaller changes. First, the pipeline gained a `latency` attribute without a value, which the latency constraint needs in order to resolve at all. The reason templates had carried the words "capacity" and "server", which the service isn't allowed to know, and now use the two configured names and the word "part". The limit's unit as written in the source is now projected beside the limit, so the document can show "200 ms". Two requirements that had traced to use cases which don't observe them now trace to stated obligations.

## The capacity model

The decision to keep [an idealised capacity model](../decisions/AD-0006-idealised-capacity-model.md) requires a page stating what is computed, the assumptions, the limits of validity and the absence of an uncertainty estimate. That page is published beside the verdicts a reader sees, and this section follows it.

### What is computed

A pipeline part owns servers, which are its child parts, and a wiring, which is a set of directed connections between those servers. Each server carries one numeric attribute, `throughput`, in queries per second. The capacity of the pipeline is the largest sustained query rate the wiring can carry, from the servers that receive queries to the servers that deliver results. The capacity service computes that number, names the servers that limit it, and returns a verdict for every requirement that constrains a quantity it computes.

Only two names configure the service: the quantity it computes, `capacity`, and the attribute it reads from each child, `throughput`. It never sees the words "server" or "pipeline". I use them here only because this section describes the example.

From the pipeline, the service builds a flow network:

- Every server becomes two nodes, an in-node and an out-node, joined by one edge whose capacity is the server's throughput.
- Every connection becomes an edge with no capacity limit, from the out-node of its first end to the in-node of its second, which is the direction the adapter took from the order of the ends.
- A super-source is joined by unlimited edges to every server with no incoming connection.
- Every server with no outgoing connection is joined by an unlimited edge to a super-sink.

[![The shipped pipeline as a flow network. Each server is split into an in-node and an out-node joined by an edge whose capacity is its throughput: ingest 2000, parse 1200, indexA 700, indexB 700, serve 1800. Connections become edges with no limit. A super-source feeds ingest and serve feeds a super-sink. The minimum cut is the parse edge, at 1200.](../figures/flow-network.svg)](../figures/flow-network.svg)

*The shipped wiring as the capacity service sees it. The cheapest set of server edges that cuts every path is parse, at 1200.*

So entry and exit servers are inferred from the wiring. A server nobody feeds receives queries from outside, and a server that feeds nobody delivers results. A server with no connections at all is neither, and is left out of the network. So is any server the flow can't reach from an entry, or from which it can't reach an exit. Such servers are ignored for capacity and aren't named in any reason.

Capacity is the [maximum flow](https://en.wikipedia.org/wiki/Maximum_flow_problem) from the super-source to the super-sink. [Dinic's algorithm](https://en.wikipedia.org/wiki/Dinic%27s_algorithm) computes it in a few dozen lines of Go.

Because the connection edges carry no limit, every finite cut of the network is made of server edges alone. The [max-flow min-cut theorem](https://en.wikipedia.org/wiki/Max-flow_min-cut_theorem) then gives the same number a second way:

```
capacity(P) = max flow from s to t in N(P)
            = min over server sets C that separate every entry server
              from every exit server of  sum of throughput(v), v in C
```

In words, the capacity is the smallest total throughput of any set of servers whose removal would cut every path from an entry server to an exit server. Entries and exits are taken from the full wiring, before any removal. That set is a minimum cut, and it's what the demo calls the bottleneck.

Put without the flow, the rollup is a minimum over serial stages and a sum over parallel ones, and the flow agrees with that rule wherever it applies.

<details markdown="1">
<summary>Under the bonnet: why the flow agrees with minimum and sum</summary>

Flow through a chain can't exceed the smallest throughput in it, and a flow of exactly that size can be routed through, so the capacity of a chain is the minimum.

For parallel branches that share one fork and one join, any flow splits into one flow per branch, each bounded by that branch's capacity. The branch maxima can all be reached at once because the branches share no server, so the capacity of the group is the sum.

Applied recursively, replacing each chain or parallel group with one server of the group's capacity, the two rules give the maximum flow of any series-parallel wiring. A differential test in the capacity package checks the flow against them on such wirings.

</details>

Wiring that isn't series-parallel leaves the flow intact and the two rules without an answer. Fan-in from servers that weren't forked at the same point, a cycle, or several entry or exit servers all have a well-defined maximum flow. The question of whether a given stage is serial or parallel has none. So flow is the mechanism, and minimum over serial with sum over parallel is the explanation a reader can check by hand. The alternatives that lost (the rollup evaluated in the adapter, in each app, or by a service that queries the supergraph it belongs to) are in [rollup as maximum flow](../decisions/AD-0007-rollup-as-maximum-flow.md).

### The bottleneck

A network can have several minimum cuts, so the reported one has to be defined. The service reports the source-side canonical cut. That's the set of servers whose in-node is reachable from the super-source in the <span class="term" data-term="residual-network">residual network</span> of a maximum flow, but whose out-node isn't. Those are the saturated servers closest to the entry.

The set of nodes reachable from the super-source in the residual network is the same for every maximum flow. So the reported cut doesn't depend on the order in which Dinic's algorithm happened to find its augmenting paths.

![The example wiring in its three states, cut from the L2b sheet on capacity and verdicts](../img/a3-l2b-wiring-states.png)

*The example wiring in its three states, cut from the [L2b sheet on capacity and verdicts](../a3/L2b-pipeline-example-capacity-and-verdicts.pdf).*

A visitor sees the consequences directly. Raising a server outside the cut changes nothing, because the same cut still limits the flow to the same value. Raising a server inside the cut raises the capacity by the same amount, since the members of a cut add, until some other set of servers becomes the cheapest cut. At that point the bottleneck moves to that set, and further increases to the first server stop having any effect.

The demo's values are chosen so that no tie occurs anywhere on the visitor's path. Had `parse` been raised to 1600 and not 1700, the second step would still have moved the bottleneck to the index pair at 1400. The third step, `indexA` to 900, would then have left `parse` at 1600 and the pair at 1600, two minimum cuts of equal weight. The source-side rule would report `parse`, which is correct, and confusing to a visitor who has just edited `indexA`. At 1700, the pair at 1400 is the unique minimum cut after the second step, and after the third step the pair at 1600 is still strictly below `parse`.

### Verdicts and reasons

A requirement reaches the service with its subject, the quantity its constraint names, the comparison operator, the limit and, where the model declares one, the short name of its verification case. All of that comes from the adapter's projection, the generic set of fields the adapter serves for any model, carried through the router. The service evaluates a requirement only when the named quantity is the one it computes. For the example that's `capacity`, so `PIPE-R1` and its five derived requirements are evaluated, and `PIPE-R2`, which constrains `latency`, isn't.

![The quantification zone of the L2b sheet, with the formula, the four states and the derived verdicts](../img/a3-l2b-arithmetic.png)

*The quantification zone of the L2b sheet, with the formula, the four states and the derived verdicts.*

For an evaluated requirement, the verdict is PASS when the subject's capacity satisfies the projected comparison against the limit, and FAIL otherwise. The comparison is the constraint's own, `>=` for every throughput requirement in the example. A requirement written the other way round would be evaluated the other way round.

A leaf is a part that carries the configured attribute and has no children. Its capacity is its own attribute value and its bottleneck set is empty. That way the derived requirements on single servers are evaluated by the same rule, with no second code path. For that to hold in the model, the servers have to declare `capacity` as well as `throughput`. The example does that through an abstract part definition shared by the pipeline and the servers, so every throughput requirement constrains `<subject>.capacity`. A part with neither children nor the attribute is empty and has no capacity.

The four verdicts mirror `VerdictKind` from the SysML v2 Systems Library, so a reader of the model meets the same four words in the analysis. The service decides between them in this order:

1. **INCONCLUSIVE** if the constrained quantity isn't the one the service computes. This comes before anything else is looked at, so the latency requirement never reports a bad child value.
2. **ERROR** if a child's attribute is missing or negative, with a reason that names the child.
3. **INCONCLUSIVE** if the subject is empty, or the wiring has no entry part or no exit part.
4. **PASS** or **FAIL** otherwise.

In every INCONCLUSIVE case the capacity is absent, not zero.

Every verdict carries a reason string built from one of seven templates. A cut of several servers is listed in the order the router delivers the children, separated by commas. The wording avoids any verb that would have to agree in number with the cut.

| Case | Template |
|---|---|
| PASS or FAIL, subject with children | `<quantity> <value> against <limit>, limited by <cut>` |
| PASS or FAIL, leaf subject | `<attribute> <value> against <limit>` |
| INCONCLUSIVE, other quantity, verification case declared | `<verification case> is declared and no service runs it` |
| INCONCLUSIVE, other quantity, no verification case | `no service computes <quantity>` |
| INCONCLUSIVE, empty subject | `no children to analyse` |
| INCONCLUSIVE, no entry or no exit | `no entry part` or `no exit part` |
| ERROR | `<child> has <missing / negative> <attribute>` |

The words in angle brackets are the only parts that vary, and `<quantity>` and `<attribute>` are the service's two configured names, so a template never carries a word of the model. For the example the first template renders as `capacity 1200 against 1500, limited by parse`.

The leaf template is chosen because the subject has no children, and not because the requirement is derived, which the service can't know. Words such as "allocated" belong to the document. It does know the derivation relationship, and may add them beside the reason.

### Derived limits and the worked example

The derived limits are written in the model as expressions over the limit of `PIPE-R1`, and the adapter evaluates them. The capacity service never sees the allocation rule, only the number that results. That rule gives the full rate on the serial path and an equal share per parallel branch. So `ingest`, `parse` and `serve` are each allocated the whole of the limit, and the two index servers half each.

A derived requirement can fail while the pipeline passes. The pipeline's capacity counts what a parallel group delivers in total, and one branch delivering more than its share covers another delivering less. That's how allocated requirements behave wherever a budget is split over parts, and the two verdicts side by side are consistent with each other. The shipped requirements document says so in one unnumbered paragraph of prose above `PIPE-R1`. It's the one caveat the idealised-model decision insists a reader meets in the document itself.

As shipped, the values are `ingest` 2000, `parse` 1200, `indexA` 700, `indexB` 700 and `serve` 1800. The wiring runs from `ingest` to `parse`, from `parse` to both index servers, and from both index servers to `serve`. `PIPE-R1` requires 1500 of the pipeline. The derived limits are 1500 for `ingest`, `parse` and `serve`, and 750 for each index server.

| State | Capacity | `PIPE-R1` | Cut | Derived: ingest, parse, indexA, indexB, serve |
|---|---|---|---|---|
| Shipped | 1200 | FAIL | parse | PASS, FAIL, FAIL, FAIL, PASS |
| `ingest` to 3000 | 1200 | FAIL | parse | PASS, FAIL, FAIL, FAIL, PASS |
| `parse` to 1700 | 1400 | FAIL | indexA, indexB | PASS, PASS, FAIL, FAIL, PASS |
| then `indexA` to 900 | 1600 | PASS | indexA, indexB | PASS, PASS, PASS, FAIL, PASS |

The second row is the case where nothing moves, with `ingest` raised from 2000 to 3000 in the shipped state. Capacity stays at 1200 and the cut stays at `parse`. `PIPE-R1` still fails, with the reason `capacity 1200 against 1500, limited by parse`, and every derived verdict is unchanged. `PIPE-R1.1` on `ingest` already passed at 2000.

![The note panel on the nothing-moves case, cut from the architecture views](../img/v3-nothing-moves.png)

*The note panel on the nothing-moves case, cut from the [architecture views](../architecture/architecture-views.pdf).*

The last row shows the derived failure. `PIPE-R1` passes at 1600 against 1500, with the reason `capacity 1600 against 1500, limited by indexA, indexB`. Meanwhile `PIPE-R1.4` on `indexB` fails with `throughput 700 against 750`, and `indexA` at 900 is what makes up the difference.

### Assumptions, limits and edge cases

The number is exact for a pipeline that meets eight assumptions:

1. Every query traverses exactly one path from an entry server to an exit server, and no server duplicates, drops or multiplies queries. So flow is conserved at every server, and a fork splits the queries between branches instead of sending each to every branch.
2. Work can be split evenly across parallel branches, so a group of branches can be loaded to the sum of their throughputs.
3. Load balancing is perfect, so queries reach whichever branch has capacity.
4. There's no queueing and no coupling through latency, so throughput is the only quantity that limits anything.
5. Load is stationary, a sustained rate with no bursts.
6. Connections have unlimited capacity, so only servers limit the flow.
7. A server's throughput doesn't depend on the mix of queries it receives.
8. Entry and exit are read from the wiring, so a feedback connection into an exit server would make it stop being one.

> [!WARNING]
> **Not for capacity planning**
>
> For a real pipeline that meets the first assumption, the usual departures from the others (queueing, uneven balancing, bursts and imperfect routing) only lower the rate it can sustain, so the number then reads as an upper bound. A server that duplicates or drops queries breaks the first assumption and can move the real rate either way. The service can't detect any departure, because it sees only throughputs and connections. No quantitative estimate of the uncertainty is available, and the results must not be used for capacity planning. The arithmetic is exact for the idealised model, and I chose the idealised model to make a point about federation.

The edge cases follow from the rules above, and add nothing new:

- Zero throughput is a valid value. A zero on the serial path gives capacity 0 and FAIL, with that server as the cut.
- A negative or missing throughput in the model file gives ERROR, with the server named. The adapter refuses a negative or non-numeric value on edit, so ERROR can only come from a literal in the source.
- A cycle is handled by the flow with no special treatment.
- A wiring left with no exit or no entry gives INCONCLUSIVE, with the reason `no exit part` or `no entry part`, and no capacity value.
- An empty subject gives INCONCLUSIVE, with the reason `no children to analyse`.

![The physical view from the L2b sheet, where the service holds no copy of the model and recomputes on every read](../img/a3-l2b-physical.png)

*The physical view from the L2b sheet, where the service holds no copy of the model and recomputes on every read.*

The service never sees SysML. It gets no model file, no KerML, no usage, no relationship beyond a from and a to, and nothing about derivation. The router carries the parts, attributes, connections and requirement fields to it on every query that asks for a capacity or a verdict. The service recomputes from scratch each time, so it can't be stale and needs no change feed. Its agreement with the adapter is the entity key (a SysML short name), the field set of the generic projection it declares in its `@requires`, and the two names it's configured with.

## Two things deliberately not done

The document's row for `PIPE-R2` shows no current value. No service computes latency, so there's none to show. The requirement on what a row displays now says the value appears only where the analysis returns one, instead of inventing a figure so the row looks complete. The row's verdict is INCONCLUSIVE with the reason that `PIPE-VC1` is declared and no service runs it, which is a more useful sentence than a number would have been.

I left the second question open on purpose. At the time, the repository's tracking policy listed the internal helper packages as local only, and yet told tests to import one of them. So a tracked test that did would fail on a fresh clone and in continuous integration. The choice was to track the two helper packages and amend the policy, or to keep tracked tests free of the import. The gate recorded the question as one for the next gate, where it became [track the internal helpers](../decisions/AD-0022-track-internal-helpers.md). I still think that was the right place for it, since a tracking decision belongs with the other decisions and not in a constraint's rationale.

---

Previous: [Twelve use cases and one moving bottleneck](04-twelve-use-cases-and-one-moving-bottleneck.md) · Index: [Federating a systems model](../README.md) · Next: [Five views and twenty-six decisions](06-five-views-and-twenty-six-decisions.md)

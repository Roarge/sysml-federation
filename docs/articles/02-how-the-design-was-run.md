# How the design was run

*Roar Georgsen, 27 August 2026*

Part 3 of 13 in [Federating a systems model](../README.md), written for release v0.3.0.

> [!IMPORTANT]
> **The story so far**
>
> The demo publishes a SysML v2 model of a five-server query pipeline through a federated <span class="term" data-term="router">router</span>. Two services that have never parsed a model file attach a throughput verdict (one service's judgement on one requirement) and a requirements document to it, and [The architecture in one sitting](01-the-architecture-in-one-sitting.md) describes the result. This part goes back to before any of that code existed, to show how I ran the design and what each round of reading caught.

## Why design before code

At the start, the repository held a Go module, a Makefile, a gitignore that works as an allowlist, and a set of coding rules. There was no product code at all. I ran the whole design before writing the first line of it, in a fixed order:

1. use cases, drawn as a storyboard rather than written as a list
2. requirements, derived from those use cases and from the technical constraints of the platform
3. an architecture in views, with a decision record behind each choice
4. a design for the A3 sheets that give a reader the whole thing on one page

The order did more work than the documents. Requirements written before the use cases would have described an adapter someone imagined, and not a demo someone would watch.

It also brought the rollup into view early. The throughput rollup, the pipeline-wide figure computed from the servers, is the only global function in the system. Working out what it needs became a section of the plan instead of a detail inside the analysis service. Its inputs fix the minimum <span class="term" data-term="projection">projection</span> the adapter has to publish. They put the analysis on the read path, with no copy of the model. They also leave capacity declared in the model without a value, so that every requirement is a constraint over it. In the same working out, [maximum flow](../decisions/AD-0007-rollup-as-maximum-flow.md) beat series-parallel reduction, because flow handles any wiring and explains the bottleneck through the cut.

## Four gates

> [!NOTE]
> **Gate**
>
> A checkpoint where the work stops until its output is approved. Each of the four gates here ended with a set of documents, a reading that tried to find what was wrong with them, and my approval. The four ran in a fixed order.

| Gate | What it produced |
|---|---|
| 1 | the design brief, twelve use cases and their storyboard |
| 2 | the constraints card, the requirements, the traceability table and the capacity model |
| 3 | the architecture description in five views, draft schemas, the decision records and one board per view |
| 4 | the A3 design and drafts of two of the four sheets |

### Gate 1: what the demo would show

Gate 1 produced the design brief, twelve use cases as text, and a storyboard with one board per use case plus an overview. [Twelve use cases and one moving bottleneck](04-twelve-use-cases-and-one-moving-bottleneck.md) goes through it board by board. The brief fixed the values of the example model, the shipped structure of the requirements document and the set of six numbers that can be edited. Approving it meant the storyboard was what the demo would show.

Before that approval, I read the argument, the plan, the brief and the use cases against each other. Every arithmetic claim held. What came back were gaps, and the one that mattered most was a tie.

The pipeline runs `ingest` into `parse`, then `parse` into `indexA` and `indexB` in parallel, and both index servers into `serve`. A cut across the index pair carries the sum of the two. In the draft, with `parse` raised to 1600 and `indexA` to 900, the stages stood like this:

| Stage | Throughput in that state |
|---|---|
| `ingest` | 2000 |
| `parse` | 1600 |
| `indexA` and `indexB` in parallel | 900 + 700 = 1600 |
| `serve` | 1800 |

`parse` and the index pair each stood at 1600 queries per second. So the minimum cut, the set of servers whose combined throughput bounds the pipeline and the thing the demo calls the bottleneck, wasn't unique. The board would have highlighted whichever one the implementation happened to pick. Use case 4, raise the bottleneck, now raises `parse` to 1700, so the cut stays at the index pair.

The brief had also promised a pass on the first raise, which the chosen limit can't deliver, and it named no subject on any requirement. A second reading of the thirteen boards found no wrong number.

### Gate 2: the requirements

Gate 2 produced four documents:

- a constraints card of 95 entries, each with a source and a confidence
- the requirements
- a <span class="term" data-term="traceability">traceability</span> table, generated from the requirements by a script that checks coverage in both directions
- the capacity model page

I read each against its own sources, then all four against each other. A second reading went back over everything the first had called serious. All of it held.

Three things the reading turned up said the design itself was wrong. Those went back to the brief as new decisions, instead of being patched over in the requirements:

- **The verdict rule.** Both the throughput requirement and the latency requirement have the pipeline as their subject. A verdict rule keyed on the subject would have passed the latency requirement at a capacity of 1200 queries per second against a limit of 200 ms. The adapter now reads the [quantity, comparison and limit from the constraint itself](../decisions/AD-0008-quantity-from-constraint.md), and the analysis returns an inconclusive verdict for any quantity it doesn't compute.
- **Connection direction.** A SysML v2 `connect` has ordered ends and no direction of its own. So direction became [the order of the ends](../decisions/AD-0009-connection-direction.md), and a wiring whose port directions disagree with it is refused as a model error.
- **Reason strings.** The verdict's reasons are [built from fixed templates](../decisions/AD-0024-reason-templates.md) that avoid the model's words, because the analysis can't know that a requirement is derived.

The rest changed the documents. Nine requirements named two obligations in one statement and were split. Four use case criteria had no requirement behind them and gained one each. Together that took the count from thirty-one to forty-five.

A second pass across the revised set found two gaps that the first fix had opened. One was that a derived requirement's constraint has to read its own server's capacity, so the model gained an abstract part definition shared by the pipeline and its servers. Part 6, [From use cases to requirements](05-from-use-cases-to-requirements.md), covers both the requirements and the capacity model.

### Gate 3: the architecture

By the end of gate 3 the architecture description existed in five views: context, composition, runtime, deployment and adapter. [Five views and twenty-six decisions](06-five-views-and-twenty-six-decisions.md) walks through them. The gate also produced draft schemas for the three <span class="term" data-term="subgraph">subgraphs</span>, the decision records, and the architecture boards, one per view.

![The five views, each with its stakeholders and its question](../img/architecture-five-views.png)

*The five views, each with its stakeholders and its question. Cut from the [architecture views](../architecture/architecture-views.pdf).*

Twenty-four records came from decisions already made in the brief, the plan and the log. The traceability then showed six requirements with no decision behind them, and two more records followed. The [document's own structure](../decisions/AD-0025-document-owns-its-structure.md) and the [viewer's text beside a sketch of its wiring](../decisions/AD-0026-viewer-shows-text-and-wiring.md) were real decisions that nobody had written down.

Nothing in the reading of the description overturned a decision, and every fix was local to one mechanism. The checks on the records did correct a run of misattributions. One claimed the whole analysis service was a few dozen lines, when only the flow algorithm is.

### Gate 4: the A3 sheets

Gate 4 produced the A3 design, which [An A3 sheet for a fifteen-minute reader](07-an-a3-sheet-for-a-fifteen-minute-reader.md) covers, and drafts of two of the four sheets: the one that carries the argument and the one that carries the moving bottleneck.

Reading the design back against the research corrected its method paragraph. It had presented placements from the A3 cookbook the design draws on as rules, where [the research](03-what-the-research-overturned.md) records them as guidance. The safeguard on the sheet's numbers became a unit test. It reads the shipped model through the adapter, runs the rollup, and checks every number in the table against the SVG.

Approving gate 4 closed the design phase. The implementation plan came next, in phases built from the approved record. It has a top-level document with one line per task, and one detail document per phase. Each phase is one pull request, and [Planning the build](08-planning-the-build.md) describes it.

## Working documents and the record

Three kinds of document carry the history, and each has a different job.

The **engineering log** is append-only, with one entry per gate. Every entry has the same parts: what was done, the decisions taken with the alternatives that lost, the findings that changed the design, the open questions and what comes next. It's the narrative the gates leave behind.

The **brief** is the ledger. Every decision has a number and a row there, and the alternatives that lost carry over into the decision records. So if you want to know why polling lost to [live push](../decisions/AD-0014-version-events.md), you can find the argument. Rows and records aren't one to one. Some rows settle a question too small to publish, some records gather several rows, and three of the records exist because the reading at gate 2 forced a decision the ledger had never had to make.

The **<span class="term" data-term="decision-record">decision records</span>** are the public form, in Michael Nygard's shape: context, decision, alternatives considered and consequences, then the requirements affected and the sources. A number is never reused, and a decision that replaces an earlier one gets its own number.

One rule holds the three together. A working document stops being the record once its content is promoted. The same rule makes [the Markdown architecture description the record and the A3 sheets the overview](../decisions/AD-0021-architecture-record.md), so a sheet that disagrees with the description is wrong by definition.

## One identifier scheme, kept light

Every page talks about two systems of interest at once. One is the demo, with its adapter, its services and its image. The other is the pipeline inside the example model, whose requirements are declared in the `.sysml` file with short names. You have to be able to tell them apart from the shape of an identifier alone.

The [light scheme](../decisions/AD-0023-light-requirements-scheme.md) gives the repository's own artefacts a short prefix and a two-digit number. There's one prefix each for use cases, system requirements, design constraints and entries on the constraints list, and decision records get a four-digit number. A requirement inside the model keeps its SysML short name, in code font, and is always called a model requirement. Model requirement `PIPE-R1` is the pipeline's throughput requirement, and nothing about the demo is ever written that way. Traceability runs one hop per table. I generated it once at gate 2 and kept it up by hand in Markdown, until the demo got [a model of its own](12-a-model-of-the-demo-itself.md) in September. The model has been the trace source since.

A fuller systems engineering methodology, with its own identifier shapes and file rules, lost twice. Used throughout, it would have shown on every page of a repository whose reader is judging a SysML adapter, not a way of working. Used privately, it would have left two schemes to drift apart. A single shared prefix for both kinds of requirement lost too, because a shared shape hides exactly the mix-up a distinct shape shows.

## Rules I hold the repository to

The gitignore is an allowlist. Nothing gets committed unless a rule names it. That's why [tracking the two test helper packages](../decisions/AD-0022-track-internal-helpers.md) needed a decision, since a tracked test that imports one would otherwise fail on a fresh clone and in CI.

Hand-written Go may not use an empty interface in a value position, meaning neither `interface{}` nor `any` as the type of a value. A value of either type can hold anything, and the compiler can no longer check what it is. `any` is allowed only as a type-parameter constraint. [Generated code](../decisions/AD-0016-generated-code-exempt.md) is exempt. The only credible Go library for federated subgraphs with subscriptions emits code full of `interface{}`, and an exception written into a file doesn't survive the next regeneration. The exemption rests on a one-line header, which is a weaker guarantee than the rule has anywhere else.

Tests come first, with the test as the specification. The `make cover` <span class="term" data-term="make-target">target</span> enforces a coverage floor of 70 per cent, and commits reach `main` only through a pull request. `make check` runs the rest in one command, from formatting through to the tests under the <span class="term" data-term="race-detector">race detector</span>.

## What the method cost and bought

It cost requirements. The plan aimed for 20 to 25 and gate 2 produced forty-five, because a statement that names two obligations is two requirements. I accepted the count for that reason.

It bought misattributions caught before they became the public account of anything, and three design errors caught in the requirements instead of in the code.

It also bought a broken pass criterion. The first spike of the build ran a probe model through the <span class="term" data-term="pilot-implementation">OMG pilot implementation</span>. The plan's criterion for a clean pass was a grep for error lines, and that grep couldn't see an error the pilot printed on the interactive shell's leading `1> ` line. One deliberate breakage produced exactly one such error, the grep printed nothing, and nothing was the plan's own signal for a pass. My pass criterion would have passed a broken file. [Five spikes before the first line](09-five-spikes-before-the-first-line.md) has the rest.

I still think the gates were the right shape for a repository one person maintains. What I'd change is the shape of the plan. One detail document per phase should have been the first draft's form, and not its repair.

---

Previous: [The architecture in one sitting](01-the-architecture-in-one-sitting.md) · Index: [Federating a systems model](../README.md) · Next: [What the research overturned](03-what-the-research-overturned.md)

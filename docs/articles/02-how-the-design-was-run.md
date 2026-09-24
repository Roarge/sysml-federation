# How the design was run

*Roar Georgsen, 27 August 2026*

Part 3 of 13 in [Federating a systems model](../README.md), written for release v0.3.0.

> [!IMPORTANT]
> **The story so far**
>
> The demo publishes a SysML v2 model of a five-server query pipeline through a federated <span class="term" data-term="router">router</span>. Two services that have never parsed a model file attach a throughput verdict and a requirements document to it, and [The architecture in one sitting](01-the-architecture-in-one-sitting.md) describes the result. This part goes back to before any of that code existed, to show how I ran the design and what each round of reading caught.

## Why design before code

At the start, the repository held a Go module, a Makefile and a set of coding rules, and no product code at all. I ran the whole design before writing the first line of it, in a fixed order:

1. use cases, drawn as a storyboard
2. requirements derived from them
3. an architecture, with a decision record behind each choice
4. a design for the one-page summaries that give a newcomer the whole thing at a glance

The order did more work than the documents. Requirements written before the use cases would have described an adapter someone imagined, and not a demo someone would watch.

It also brought the one piece of real arithmetic into view early. The throughput rollup, the pipeline-wide figure computed from the servers, is the only global function in the system. Working out what it needs fixed the minimum <span class="term" data-term="projection">projection</span> the adapter has to publish, and put the analysis on the read path with no copy of the model. It also settled that [maximum flow](../decisions/AD-0007-rollup-as-maximum-flow.md) would compute it, because flow handles any wiring and explains the bottleneck.

## Four gates

> [!NOTE]
> **Gate**
>
> A checkpoint where the work stops until its output is approved. Each of the four gates here ended with a set of documents, a reading that tried to find what was wrong with them, and my approval.

| Gate | What it produced |
|---|---|
| 1 | the design brief, twelve use cases and their storyboard |
| 2 | the requirements, the traceability table and the capacity model |
| 3 | the architecture description in five views, draft schemas, the decision records and one board per view |
| 4 | the design of the one-page A3 summaries, and drafts of two of them |

The reading is the point. I read each gate's documents against their sources and against each other, trying to break them, and what the reading found went back into the design before the next gate opened. The first three gates each made a catch worth telling.

**Gate 1 found a tie.** In the first draft of the script, one step left two different sets of servers limiting the pipeline by exactly the same amount. The demo's memorable moment, the bottleneck highlighted on screen, would have depended on which one the code happened to pick. One number in the script changed to avoid it, and [Twelve use cases and one moving bottleneck](04-twelve-use-cases-and-one-moving-bottleneck.md) tells that story.

**Gate 2 found three design errors.** Two requirements with the same subject would have been judged by the same rule, which would have passed a latency requirement against a throughput figure. A SysML connection has no direction of its own, and the rollup needs one. And the reason strings would have used words the analysis can't know. Each went back to the brief as a decision, not a patch: [the quantity read from the constraint](../decisions/AD-0008-quantity-from-constraint.md), [direction from the order of the ends](../decisions/AD-0009-connection-direction.md) and [reasons from fixed templates](../decisions/AD-0024-reason-templates.md). The same reading split nine requirements that each named two obligations, and added four the use cases needed, which took the count to forty-five. [From use cases to requirements](05-from-use-cases-to-requirements.md) covers them.

**Gate 3 found two decisions nobody had written down.** The <span class="term" data-term="traceability">traceability</span> showed six requirements with no decision behind them. That turned up two real decisions that had been taken and never recorded: [the document owning its own structure](../decisions/AD-0025-document-owns-its-structure.md), and [the viewer showing the model's text beside a sketch of its wiring](../decisions/AD-0026-viewer-shows-text-and-wiring.md). Nothing in the reading overturned a decision. [Five views and twenty-six decisions](06-five-views-and-twenty-six-decisions.md) walks through the architecture.

![The five views, each with its stakeholders and its question](../img/architecture-five-views.png)

*The five views, each with its stakeholders and its question. Cut from the [architecture views](../architecture/architecture-views.pdf).*

**Gate 4** designed the one-page A3 summaries and drafted two of them, which [An A3 sheet for a fifteen-minute reader](07-an-a3-sheet-for-a-fifteen-minute-reader.md) covers.

Approving gate 4 closed the design phase. The plan that followed, one pull request per phase, is in [Planning the build](08-planning-the-build.md).

## Where the record lives

An engineering log kept the narrative, one entry per gate. A brief kept the ledger of decisions, each with the alternatives that lost. What was worth publishing became a <span class="term" data-term="decision-record">decision record</span>, the short form Michael Nygard proposed for context, decision and consequences, to which mine add the alternatives that lost. One rule holds them together. A working document stops being the record once its content is promoted, so when two documents disagree, the record wins.

A few house rules shaped the code from the first commit. The gitignore is an allowlist, so nothing gets committed unless a rule names it. That's why [tracking two test helper packages](../decisions/AD-0022-track-internal-helpers.md) needed a decision of its own, since tests import them and would otherwise fail on a fresh clone. Hand-written Go may not use an empty interface as the type of a value, because a value of that type can hold anything and the compiler can no longer check it. [Generated code](../decisions/AD-0016-generated-code-exempt.md) is exempt. Tests come first, and commits reach `main` only through a pull request.

## What the method cost and bought

It cost requirements. The plan aimed for 20 to 25 and got forty-five, because a statement that names two obligations is two requirements.

It bought three design errors caught in the requirements instead of in the code, and misattributed claims caught before they reached anything public.

It also bought a broken pass criterion, which I only caught because I tested the test. The check I'd written for the first spike of the build would have reported a broken model file as a pass. [Five spikes before the first line](09-five-spikes-before-the-first-line.md) has that story.

I still think the gates were the right shape for a project one person maintains. What I'd change is the shape of the plan. It ended up with one detail document per phase, and that should have been its form from the first draft, not a repair.

---

Previous: [The architecture in one sitting](01-the-architecture-in-one-sitting.md) · Index: [Federating a systems model](../README.md) · Next: [What the research overturned](03-what-the-research-overturned.md)

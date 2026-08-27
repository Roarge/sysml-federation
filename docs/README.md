# Federating a systems model

*Roar Georgsen, 27 August 2026*

An adapter publishes a SysML v2 model as a federated service, and a worked example joins it to a throughput analysis and a requirements document that know nothing about SysML. In the example, a query pipeline of five servers rolls its capacity up from their throughputs, and a requirement passes or fails as the numbers change. These articles describe the design from the problem through to the version being built, and they are written to be read by someone deciding whether the approach fits their organisation.

## The articles

Read in order, or start with the overview and come back.

1. [Why federate a systems model](articles/00-why-federate-a-systems-model.md). The integration problem MBSE never solved, what SysML v2 changes, and the claim this repository makes.
2. [The architecture in one sitting](articles/01-the-architecture-in-one-sitting.md). Three services, one router, two apps and one container, with the twelve decisions that shaped them.
3. [How the design was run](articles/02-how-the-design-was-run.md). Four gates before any code, every document read back against its sources before approval, and the repository rules that held throughout.
4. [What the research overturned](articles/03-what-the-research-overturned.md). Six topics investigated, every load-bearing claim checked against its primary sources, and four of them refuted.
5. [Twelve use cases and one moving bottleneck](articles/04-twelve-use-cases-and-one-moving-bottleneck.md). The brief, the personas, the example model and the storyboard, and what the first review changed.
6. [From use cases to requirements](articles/05-from-use-cases-to-requirements.md). Constraints, 45 requirements, traceability and the capacity model, with the three findings that reached back into the design.
7. [Five views and twenty-six decisions](articles/06-five-views-and-twenty-six-decisions.md). The architecture description, its viewpoints, the decision records and the review that changed mechanisms without changing decisions.
8. [An A3 sheet for a fifteen-minute reader](articles/07-an-a3-sheet-for-a-fifteen-minute-reader.md). Why the overview is a printed sheet, the method it follows, and the two sheets drafted.
9. [Planning the build](articles/08-planning-the-build.md). Five phases, one pull request each, test first, and the decisions the plan had to make on its own.
10. [Five spikes before the first line](articles/09-five-spikes-before-the-first-line.md). The syntax the reference tools accept, nested requires through the router, a router with no config file whose readiness lies, and cross-platform copying.
11. [The demo being built](articles/10-the-demo-being-built.md). What it does once it runs, package by package and service by service, and what a visitor sees in fifteen minutes.

## Documents

- [Architecture views](architecture/architecture-views.pdf), the five views and the overview board as one PDF, one page per board.
- [L0, Federating a systems model](a3/L0-federating-a-systems-model.pdf), an A3 sheet for the reader asking "What does this demo claim, what is in the box, and what would I keep or replace if I adopted it?"
- [L2b, Pipeline example: capacity and verdicts](a3/L2b-pipeline-example-capacity-and-verdicts.pdf), an A3 sheet for the reader asking "Why does raising one server change nothing and raising another change everything?"
- [Use cases](stories/use-cases.pdf), the storyboard as one PDF, one page per use case after the overview.
- [Decision records](decisions/README.md), the 26 decisions with their alternatives and consequences.

## The repository

The code is at https://github.com/Roarge/sysml-federation. The command that starts the demo arrives when the first release is tagged. The same articles are rendered as a site at https://roarge.github.io/sysml-federation/.

---

Licence: Apache 2.0 · Repository: https://github.com/Roarge/sysml-federation

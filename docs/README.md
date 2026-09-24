# The missing integration layer for open MBSE?

*Roar Georgsen, 27 August 2026*

The goal of this project is to make SysML v2 models easier to integrate with the domain models of the organisation's other tools. The proposal is to adopt federation. The model's owners publish a small projection of it (SysML Views and Viewpoints), and other services attach their own data to its objects by agreeing on an identifier.

So far this is a proof of concept, a SysML v2 adapter behind a Cosmo router with one worked example. Next is a larger example and an adapter driven by SysML v2's own views and viewpoints.

## The idea

Why models stay locked in their tools, and what federation changes.

1. [Why federate a systems model](articles/00-why-federate-a-systems-model.md)  
   Models were meant to end the drift between an organisation's documents, and closed tools kept them out of reach. SysML v2 opens the model up, and federation lets other tools join it without learning SysML.
2. [The architecture in one sitting](articles/01-the-architecture-in-one-sitting.md)  
   How one query collects its answer from three services that never talk to each other, and the decisions that put them all in one container.

## The design

How the demo was designed before any code was written, and what each round of review caught.

3. [How the design was run](articles/02-how-the-design-was-run.md)  
   Four approval gates stood between the idea and the first line of code. Reading each document back against its sources caught three design errors early.
4. [What the research overturned](articles/03-what-the-research-overturned.md)  
   Before designing anything, I wrote down what I believed about the tools and tried to knock each belief over. Four of the seven claims that mattered fell.
5. [Twelve use cases and one moving bottleneck](articles/04-twelve-use-cases-and-one-moving-bottleneck.md)  
   The example pipeline and the twelve short stories a visitor walks through. Their one surprise is that raising the bottleneck moves it somewhere else.
6. [From use cases to requirements](articles/05-from-use-cases-to-requirements.md)  
   Turning the stories into forty-five testable requirements exposed three holes in the design, and pinned down how the capacity number is computed.
7. [Five views and twenty-six decisions](articles/06-five-views-and-twenty-six-decisions.md)  
   The architecture seen from five angles, each for a different reader, with a written record behind every choice that shaped it.
8. [An A3 sheet for a fifteen-minute reader](articles/07-an-a3-sheet-for-a-fifteen-minute-reader.md)  
   Why the overview is one printed sheet, what the method behind it fixes and what it only suggests, and the two sheets drawn so far.

## The build

From the approved design to a published container image.

9. [Planning the build](articles/08-planning-the-build.md)  
   How the approved design became five phases of work, one pull request each and tests first, and the choices the design had left open.
10. [Five spikes before the first line](articles/09-five-spikes-before-the-first-line.md)  
    Five small experiments tested the riskiest assumptions before any product code existed. None failed outright, and four of them corrected the plan.
11. [The demo as it shipped](articles/10-the-demo-as-it-shipped.md)  
    What the finished demo does once it runs, package by package, and the fifteen-minute walk a visitor takes through it.

## Looking back

What the finished demo proves and leaves open, and a model of the demo itself.

12. [What shipped, and what did not](articles/11-what-shipped-and-what-did-not.md)  
    What the published image weighs and how its release is guarded, with a list of what the running container does not prove.
13. [A model of the demo itself](articles/12-a-model-of-the-demo-itself.md)  
    The repository takes its own advice. The demo is described as a SysML v2 model that a test keeps in step with the code, and a Checkly session runs every story against the live demo.

## Documents

- [Architecture views](architecture/architecture-views.pdf), the five views and the overview board as one PDF, one page per board.
- [L0, Federating a systems model](a3/L0-federating-a-systems-model.pdf), an A3 sheet for the reader asking "What does this demo claim, what is in the box, and what would I keep or replace if I adopted it?"
- [L2b, Pipeline example: capacity and verdicts](a3/L2b-pipeline-example-capacity-and-verdicts.pdf), an A3 sheet for the reader asking "Why does raising one server change nothing and raising another change everything?"
- [Use cases](stories/use-cases.pdf), the storyboard as one PDF, one page per use case after the overview.
- [Decision records](decisions/README.md), the 31 decisions with their alternatives and consequences.
- [The model of the demo](https://github.com/Roarge/sysml-federation/tree/main/model), the demo itself in SysML v2, with its stories, requirements, architecture, tests and views, validated by the two reference tools on every change.
- [The check session](https://github.com/Roarge/sysml-federation/tree/main/checkly), the Checkly project that runs every story against a live instance through a tunnel, with a collector and a trace viewer. It needs your own Checkly account, since mine cannot ship with the image.

## The repository

The code is at https://github.com/Roarge/sysml-federation, under the Apache 2.0 licence. The same articles are rendered as a site at https://sysml-federation.org/.

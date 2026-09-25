# The missing integration layer for open MBSE?

*Roar Elias Georgsen, 27 August 2026*  
[LinkedIn](https://www.linkedin.com/in/roareliasgeorgsen) · [Bluesky](https://bsky.app/profile/roarge.bsky.social) · [Google Scholar](https://scholar.google.com/citations?user=_IxcvnQAAAAJ&hl=en) · [ORCID](https://orcid.org/0000-0003-2868-497X)

I want SysML v2 models to be easier to connect to everything else an engineering organisation runs, down to the spreadsheets that quietly hold the place together. My proposal is federation. The model's owners publish a small projection of it, chosen with SysML's own views and viewpoints, and other services attach their own data to the model's elements by agreeing on an identifier. Nobody else has to learn SysML to join in.

This site is where I write up the experiments. It is a personal blog, so expect opinions and the occasional admission that I should have known better. The first series builds a proof of concept, a SysML v2 adapter behind a Cosmo router with one worked example, and follows it from the first sketch to a published container image. It is a demo and not a product. The second series, still being written, asks how a systems model's links to everything else that gets built could be found and kept without anyone typing them.

## Federating a systems model

Complete, in thirteen parts, written for release v0.3.0. If a later release changes something an article says, that article gets a note saying so.

### The idea

Why do models stay locked in their tools, and what does federation change?

1. [Why federate a systems model?](articles/00-why-federate-a-systems-model.md)  
   Models were meant to end the drift between an organisation's documents, and closed tools kept them out of reach. SysML v2 opens the model up, and federation lets other tools join it without learning SysML.
2. [The architecture in one sitting](articles/01-the-architecture-in-one-sitting.md)  
   How does one query collect its answer from three services that never talk to each other, and which decisions put them all in one container?

### The design

How was the demo designed before any code was written, and what did each round of review catch?

3. [How the design was run](articles/02-how-the-design-was-run.md)  
   Four approval gates stood between the idea and the first line of code. Reading each document back against its sources caught three design errors early.
4. [What the research overturned](articles/03-what-the-research-overturned.md)  
   Before designing anything, I wrote down what I believed about the tools and tried to knock each belief over. Three of the six claims that mattered fell.
5. [Twelve use cases and one moving bottleneck](articles/04-twelve-use-cases-and-one-moving-bottleneck.md)  
   The example pipeline and the twelve short stories a visitor walks through. Their one surprise is that raising the bottleneck moves it somewhere else.
6. [From use cases to requirements](articles/05-from-use-cases-to-requirements.md)  
   Turning the stories into forty-five testable requirements exposed three holes in the design, and pinned down how the capacity number is computed.
7. [Five views and twenty-six decisions](articles/06-five-views-and-twenty-six-decisions.md)  
   The architecture seen from five angles, each for a different reader, with a written record behind every choice that shaped it.
8. [An A3 sheet for a fifteen-minute reader](articles/07-an-a3-sheet-for-a-fifteen-minute-reader.md)  
   Why is the overview one printed sheet, what does the A3 method force on it, and what do the two sheets drawn so far show?

### The build

How did the approved design become a published container image?

9. [Planning the build](articles/08-planning-the-build.md)  
   How did the approved design become five phases of work, one pull request each and tests first? And which choices did the design leave for the plan?
10. [Five spikes before the first line](articles/09-five-spikes-before-the-first-line.md)  
    Five small experiments tested the riskiest assumptions, four of them before any product code existed. None failed outright, and each of those four corrected the plan.
11. [The demo as it shipped](articles/10-the-demo-as-it-shipped.md)  
    What does the finished demo show once it runs, how does a visitor's fifteen-minute walk through it go, and what did the build come to?

### Looking back

What does the finished demo prove and leave open, and why does it now have a systems model of its own?

12. [What shipped, and what did not](articles/11-what-shipped-and-what-did-not.md)  
    How big is the published image, how is its release guarded, what does a run with no network prove, and where has the demo been run?
13. [A model of the demo itself](articles/12-a-model-of-the-demo-itself.md)  
    I take my own advice at last. This article describes the demo as a SysML v2 model that a test keeps in step with the code, with a Checkly session that runs every storyboard story against the live demo.

### The documents behind it

These are the boards, sheets and records the articles draw on, for anyone who wants the source material.

- [Architecture views](architecture/architecture-views.pdf), the five views and the overview board as one PDF, one page per board.
- [L0, Federating a systems model](a3/L0-federating-a-systems-model.pdf), an A3 sheet for the reader asking "What does this demo claim, what is in the box, and what would I keep or replace if I adopted it?"
- [L2b, Pipeline example: capacity and verdicts](a3/L2b-pipeline-example-capacity-and-verdicts.pdf), an A3 sheet for the reader asking "Why does raising one server change nothing and raising another change everything?"
- [Use cases](stories/use-cases.pdf), the storyboard as one PDF, one page per use case after the overview.
- [Decision records](decisions/README.md), the 32 decisions with their alternatives and consequences.
- [The model of the demo](https://github.com/Roarge/sysml-federation/tree/main/model), the demo itself in SysML v2, with its stories, requirements, architecture, tests and views, validated by the two reference tools on every change.
- [The check session](https://github.com/Roarge/sysml-federation/tree/main/checkly), the Checkly project that runs every story against a live instance through a tunnel, with a collector and a trace viewer. It needs your own Checkly account, since mine cannot ship with the image.

## Automating traceability

In progress. The second series takes on automating the federation itself. Before the rest of a system's tools can join the graph, something has to decide which of their objects are which elements of the systems model, and in the demo that's still me, by hand. New parts appear here as they're published.

### The task

What is entity resolution between a systems model and a built system, and why does a systems model need it?

1. [Why does a systems model need entity resolution?](articles/13-why-a-model-needs-entity-resolution.md)  
   Hardly anything a team builds carries the keys from its systems model. What does it take to find the counterparts of each element of the systems model anyway, and why is that job never quite finished?

## The repository

The code is at https://github.com/Roarge/sysml-federation, under the Apache 2.0 licence.

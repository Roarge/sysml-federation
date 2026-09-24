# An A3 sheet for a fifteen-minute reader

*Roar Georgsen, 27 August 2026*

Part 8 of 13 in [Federating a systems model](../README.md), written for release v0.3.0.

> [!IMPORTANT]
> **The story so far**
>
> The demo publishes a SysML v2 model, a capacity analysis and a requirements document as three GraphQL services, served as one graph through a router, with two web apps in front, all in one container. Part 7 introduced the architecture, drawn as five views with twenty-six decisions behind it. This part is about the opposite problem: getting the important part onto a single sheet of paper.

## Who the sheet is for

The reader I have in mind is an engineering manager, or a systems engineer new to federation, deciding in a quarter of an hour whether the approach deserves a closer look. That reader isn't going to open an [architecture description](../architecture/README.md) with five views and twenty-six decision records. What they need fits on one sheet of paper.

So the design phase made the Markdown description and the records the record, and the sheets the overview ([the architecture record decision](../decisions/AD-0021-architecture-record.md)). When the design changes, the record changes first and a sheet follows at its next re-issue, so a sheet can lag behind, and its status says so.

## Why A3

> [!NOTE]
> **A3 architecture overview**
>
> A way of summarising one aspect of a system's architecture on a single sheet of A3 paper, from Borches and Bonnema, developed at Philips Healthcare. The sheet carries simple models of the system, such as what it does, the numbers that matter and how it's built, with a short text summary.

I adopted A3 for what it forces, not for its rituals. Four things make it worth the effort.

**It gets read.** Borches found that people came to meetings having read the A3, when they hadn't read the equivalent document ([the 2010 paper](https://web.mst.edu/~lib-circ/files/Special%20Collections/INCOSE2010/A3%20Architecture%20Overviews.pdf)). A 2025 study at a Norwegian company, by Bergtun and Engen, found readers spending ten to fifteen minutes on a sheet, and over 90 per cent said they understood the system better for it.

**It forces a choice.** One aspect per sheet, and nothing that doesn't fit. Borches' [cookbook for the method](https://www.gaudisite.nl/BorchesCookbookA3architectureOverview.pdf) puts it bluntly: "you can't put everything you know about this topic in this A3! So do not try to do it." Anything else goes on another sheet, or stays in the record.

**It keeps notation out.** No formal modelling notation on the sheet, because in Borches' own experiments with SysML "most of the meetings with experts was spent discussing the notation itself rather than the content". That's a dry thing to find in the literature behind a project about SysML, and a fair warning to me.

**It layers.** Practice settled on a small hierarchy: an L0 sheet for context, an L1 for the technical overview, and L2 sheets for topics. A reader starts at the top and goes only as deep as they need.

The rest of what the method offers, I took as guidance, and used where it helped.

## Four sheets, two drawn

Each sheet has one goal question its reader should be able to answer afterwards:

| Sheet | Goal question |
|---|---|
| L0, Federating a systems model | What does this demo claim, what is in the box, and what would I keep or replace if I adopted it? |
| L1, Three subgraphs, one graph | What do the three services agree on, and what happens when a number changes? |
| L2a, Adapter: projecting SysML v2 | How does a model file become a subgraph, and where does the adapter stop? |
| L2b, Pipeline example: capacity and verdicts | Why does raising one server change nothing and raising another change everything? |

I drew L0 and L2b first, because they carry the argument and the memorable moment. L1 and L2a were to follow once the schemas stopped moving. At release v0.3.0 neither has been drawn, and [the sheet index](../a3/README.md) says so.

## L0, the argument on one page

![The model side of the L0 sheet](../img/a3-l0-model-side.png)

*The L0 sheet: the argument as eight numbered boxes on the left, what the services agree on and what's in the box top right, the container bottom right. The whole sheet is [L0, Federating a systems model](../a3/L0-federating-a-systems-model.pdf).*

The reading path is [the argument for federating a systems model](00-why-federate-a-systems-model.md) in eight boxes:

1. Author the model as text, in Git.
2. Publish a projection of it, plain types and no metamodel.
3. Attach an analysis that knows nothing about SysML.
4. Attach editorial structure that knows nothing about it either.
5. Compose the three into one graph, checked before deployment.
6. Serve two apps from the one graph.
7. Change a number in either app.
8. Watch the verdict and the document follow.

Beside it sit the numbers a manager asks about: one image, one command, one port, and how little the services have to agree on. The container view marks what an adopter would keep (the adapter, the compose step, the supervisor) and what they'd replace (the model, the example services, the apps and the document tree).

I first drew L0 with the numbers the design could only estimate, and re-issued it on 29 August once the build had measured them. It now shows the image at 45 MB for amd64 and 41 MB for arm64 against an 80 MB ceiling, and the two-second bound on an edit as met.

## L2b, how the number is made

![The model side of the L2b sheet](../img/a3-l2b-model-side.png)

*The L2b sheet: seven numbered boxes, the wiring in three states beside them, the arithmetic and its table top right. The whole sheet is [L2b, Pipeline example: capacity and verdicts](../a3/L2b-pipeline-example-capacity-and-verdicts.pdf).*

L2b is for anyone who has run the demo and wants to know why the bottleneck moved. Its seven boxes go from reading the five servers and their wiring, through the flow and the cut, to the verdicts. They end on the two lessons: raise a server outside the cut and nothing moves, raise one inside it and the cut migrates. Its table is the worked example from [From use cases to requirements](05-from-use-cases-to-requirements.md), row for row, and a note under the boxes spells out the limits of the idealised model it rests on.

The next part leaves the design behind, and plans the build.

---

Previous: [Five views and twenty-six decisions](06-five-views-and-twenty-six-decisions.md) · Index: [Federating a systems model](../README.md) · Next: [Planning the build](08-planning-the-build.md)

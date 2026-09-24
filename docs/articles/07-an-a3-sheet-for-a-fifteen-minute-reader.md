# An A3 sheet for a fifteen-minute reader

*Roar Georgsen, 27 August 2026*

Part 8 of 13 in [Federating a systems model](../README.md), written for release v0.3.0.

> [!IMPORTANT]
> **The story so far**
>
> The demo takes a SysML v2 model, a capacity analysis and a requirements document, and publishes each as a GraphQL subgraph, a service that owns part of one shared schema. It composes the three into one graph and serves two web apps from it through a router, the one endpoint that plans a query across the subgraphs and merges their answers. All of it runs from one container. [The architecture in one sitting](01-the-architecture-in-one-sitting.md) describes it for someone who will build or change it, and part 7 walked through the full architecture description. This part is about squeezing the important bits onto a single sheet of paper.

## Who the sheet is for

The reader I have in mind is an engineering manager, or a systems engineer new to federation, deciding in a quarter of an hour whether the approach deserves a closer look. That reader isn't going to open an architecture description with [five views and twenty-six decision records](06-five-views-and-twenty-six-decisions.md). What they need fits on one sheet of paper. That's why the design phase made [the Markdown description the record and the A3 sheets the overview](../decisions/AD-0021-architecture-record.md). When the design changes, the description changes first and the sheet at its next re-issue. So a sheet may lag, and its status says so.

The top sheet condenses the argument I make for the whole project, which [Why federate a systems model](00-why-federate-a-systems-model.md) sets out at length, and the decision accepts that overlap on purpose. That article is prose for a reader at a desk, and the sheet is for a meeting.

## The method, its rules and its recommendations

The format is the <span class="term" data-term="a3-overview">A3 architecture overview</span> of Borches and Bonnema, developed at Philips Healthcare on an MRI scanner and published from the University of Twente. Borches found that readers came to meetings having read the A3, when they hadn't read the equivalent text document. His [cookbook](https://www.gaudisite.nl/BorchesCookbookA3architectureOverview.pdf), a draft of December 2009 that is itself an A3 overview, holds most of the rules. The [2010 symposium paper](https://web.mst.edu/~lib-circ/files/Special%20Collections/INCOSE2010/A3%20Architecture%20Overviews.pdf) argues the method.

The design keeps what the cookbook fixes apart from what it recommends. These are fixed:

- One system aspect per sheet.
- A model side with a summary side, because models without supporting text ended up buried in other documents.
- The order of the summary side's sections.
- At most five colours plus shading, since people won't remember more.
- Type at 30, 18 and 14 points on A3.
- No formal modelling notation, since in Borches' SysML experiments the notation took up most of each meeting.
- Condensing is validated when nothing has been added and nothing removed.
- A sheet can't hold everything its author knows about the topic, so the author shouldn't try.

Everything else the cookbook offers I adopted as a recommendation: where each view sits, the visual aid beside the functional view, a legend, decisions marked as stars, and a numbered reading path. The cookbook's own template labels the summary side the front and the model side the back, the reverse of how the layout rules are usually summarised. The design steps around that by calling them the model side and the summary side, and saying nothing about which is printed first.

Two things come from later practice and not from Borches. One is the hierarchy of an L0 context sheet, an L1 technical overview and L2 topic sheets. Pesselse and others settled on it at Mercedes-Benz in 2019, after Muller in 2015 and Viken and Muller in 2018. The other is the reading time of ten to fifteen minutes that Bergtun and Engen measured on industry readers in 2025.

Where the cookbook types its stars by concern, the design simplifies a star to a decision record id and one line of rationale, and the legend says so.

## Four sheets on three levels

Each sheet has a goal question its reader should be able to answer afterwards.

| Sheet | Title | Reader | Goal question |
|---|---|---|---|
| L0 | Federating a systems model | an organisation deciding | "What does this demo claim, what is in the box, and what would I keep or replace if I adopted it?" |
| L1 | Three subgraphs, one graph | a developer evaluating it | "What do the three services agree on, and what happens when a number changes?" |
| L2a | Adapter: projecting SysML v2 | someone extending it | "How does a model file become a subgraph, and where does the adapter stop?" |
| L2b | Pipeline example: capacity and verdicts | anyone who has run the demo | "Why does raising one server change nothing and raising another change everything?" |

Four sheets on three levels sits inside Borches' one to five per aspect, and far below the forty to sixty he estimated an MRI scanner would need. L0 and L2b were drafted first, because they carry the argument and the memorable moment.

Every sheet keeps both sides, even though the research records readers who wanted no summary side, and practitioners who shipped one-sided sheets with a linked page. The summary side is where a sheet points back to the record, names who is responsible and carries the concerns table. A sheet that does none of those is a poster.

## The model side

The sheet is A3 landscape at 1587 by 1123 px, laid out like this:

- **A title band** across the top carries the sheet id and title, the goal question and a scope line in pencil. On the right it has the sheet's owner, status and date.
- **The functional view** takes the left 58 per cent. It has verb-plus-noun boxes joined by labelled arrows, at most nine, each with a circled reading-path number and one line of text.
- **The visual aid** sits wherever the flow leaves room in that column. It's a picture the reader recognises and never notation, though one small snippet of code or schema is allowed.
- **The right column** holds quantification above the physical view. Quantification is a formula and a table of values marked by confidence. The physical view shows components, with interfaces labelled by protocol and port.
- **Decision stars** run along the bottom of the functional column, with a references footer under them, and the legend sits bottom right.

The type is Source Sans 3 throughout, and nothing is below 19 px, the cookbook's 14-point floor with no exception for legends or labels. Elsewhere in this series the diagrams use a handwriting face for their titles, because they're sketches for a maintainer. The sheets are read by engineering managers, and the research records a customer who asked for standardised illustrations rather than hand-drawn ones, so the sheets drop the handwriting. The same research records readers who preferred life-like figures, so which style suits an audience that reads code repositories is untested.

There are five colours plus shading, over paper with shade for the shading:

- ink
- pencil
- blue for the router and whatever passes through it
- red for a bottleneck or a failing verdict, meaning the capacity service's judgement that a requirement's limit isn't met
- amber for an estimate

The cookbook codes confidence as blue for known, orange for estimated and red for unknown. Blue here already means the router and red a failure, so the design recodes it:

| A value that is | Cookbook | These sheets |
|---|---|---|
| known | blue | ink |
| estimated | orange | amber |
| unknown | red | a pencil question mark |

A requirement's bound, such as the two seconds from an edit to both apps rendering, is none of these. It shows in ink with its id, and the measured value beside it is a question mark until one exists. The legend says so on every sheet.

## The summary side

The summary side is text in the cookbook's order, in two columns, eleven sections:

1. definitions
2. introduction
3. system partition
4. the functional view as a paragraph
5. the physical view as a paragraph
6. a concerns table, with the cells it doesn't address hatched in pencil
7. key parameters and requirements
8. an owner block, naming who is responsible and who has read the sheet
9. design strategies, with assumptions and known issues
10. a roadmap
11. references

The cookbook remarks that the size of the text boxes doesn't matter, but the order does.

## L0, the argument in eight steps

![The model side of the L0 sheet](../img/a3-l0-model-side.png)

*The L0 model side: the argument as eight numbered boxes on the left, the agreement and the box's contents top right, the container bottom right. The whole sheet is [L0, Federating a systems model](../a3/L0-federating-a-systems-model.pdf).*

The reading path is [the argument for federating a systems model](00-why-federate-a-systems-model.md) in eight boxes:

1. Author the model as text, in Git.
2. Publish a projection of it, plain types and no metamodel. (A projection is the curated set of GraphQL types the adapter derives from the model.)
3. Attach an analysis that knows nothing about SysML.
4. Attach editorial structure that knows nothing about it either.
5. Compose the three into one graph, checked before deployment.
6. Serve two apps from the one graph.
7. Change a number in either app.
8. Watch the verdict and the document follow.

For the visual aid there's the overview sketch: two apps, one router, three services that never meet.

Quantification marks every value's confidence. The agreement between the services is one entity key (the field by which the router recognises the same element across subgraphs), one declared field set and two configured names, all known. The demo is one image, one command and one port, about 40 MB compressed for the router and under 80 MB in all, the latter amber until measured. In the example there are five servers, seven requirements and one verification case. The two seconds from an edit to both apps rendering is a requirement in ink, with a question mark beside it.

That's the sheet as I first drew it. I re-issued it on 29 August, once the build had measured what the design could only estimate. It now shows the image at 45 MB for amd64 and 41 MB for arm64 against the 80 MB ceiling, and the two seconds as met, measured on 28 August. The picture above is the re-issued sheet.

The physical view is the container as one box, with the three subgraphs, the router and the UI server inside and the browser and the registry outside. What an adopter keeps (the adapter, the compose step, the supervisor) is shaded. What it replaces (the model, the example services, the apps, the shipped document tree) is hatched.

Four decisions are starred:

- [federation over a single GraphQL service](../decisions/AD-0001-federation-over-single-service.md)
- [Cosmo as the platform](../decisions/AD-0002-cosmo-as-platform.md)
- [one binary, one port, one UI server](../decisions/AD-0011-one-binary-one-port.md)
- [telemetry off by baked-in environment](../decisions/AD-0013-telemetry-off.md)

Beside the router box, in pencil, sits the vendor's caveat with the demo's answer. Every public description of static composition in this project has to carry that note. The compose page says "it is recommended to not use this for production". The demo has no control plane to fetch from, and its configuration is committed and tested for drift.

![The L0 decisions strip, references footer and legend](../img/a3-l0-legend.png)

*The bottom of L0: four stars with one line of rationale each, the footer that keeps the corner full, and the legend where the confidence code differs from the cookbook's.*

## L2b, how the number is made

![The model side of the L2b sheet](../img/a3-l2b-model-side.png)

*The L2b model side, from [L2b, Pipeline example: capacity and verdicts](../a3/L2b-pipeline-example-capacity-and-verdicts.pdf): seven boxes on the left, the wiring in three states beside them, the arithmetic and its four-row table top right.*

This sheet has seven boxes:

1. Read five servers and their wiring from the projection.
2. Build the flow network.
3. Find the maximum flow, which is the capacity.
4. Find the source-side minimum cut, the saturated servers nearest the entry, which is the bottleneck.
5. Compare against each requirement's limit to get the verdicts.
6. Raise a server outside the cut, and nothing moves.
7. Raise a server inside the cut, and the cut migrates.

The visual aid is the wiring, with its numbers in each state of the worked example. [From use cases to requirements](05-from-use-cases-to-requirements.md) explains the capacity model behind it.

The quantification block gives the formula in words: capacity is the maximum flow, equal to the minimum cut, min over a chain and sum over parallel branches. Then comes the worked example as a four-row table. The rows are the shipped state, the ingest raise that changes nothing, the parse raise that moves the cut, and the indexA raise that makes the pipeline pass. Each row carries the cut, the capacity, the verdict on the model requirement `PIPE-R1`, and the five derived verdicts in server order. The numbers are the ones in [From use cases to requirements](05-from-use-cases-to-requirements.md), which owns the capacity model, and the sheet takes them from the running example instead of restating them from anywhere else.

`PIPE-R1` has a limit of 1500, allocated to five derived requirements, one per server: 1500 on the serial path and 750 for each index server. The second row of the table is box 6, and the fourth is box 7. In that last row `PIPE-R1.4` still fails, with indexB at 700 against its allocated 750, while `PIPE-R1` passes at 1600. That's what the document's shipped paragraph of prose explains. All the values are known, and none estimated.

In the physical view, the capacity service sits beside the adapter and the document service, with the router between them and the two apps above. The router carries the subject's children, wiring, quantity, comparison and limit to the capacity service through `@requires`. The capacity box says "holds no copy of the model, recomputes on every read".

Four decisions are starred here:

- [an idealised capacity model](../decisions/AD-0006-idealised-capacity-model.md)
- [rollup as maximum flow with the source-side minimum cut](../decisions/AD-0007-rollup-as-maximum-flow.md)
- [verdict reasons built from templates](../decisions/AD-0024-reason-templates.md)
- [the document owning its own structure](../decisions/AD-0025-document-owns-its-structure.md)

The known-issues slot carries the limits of validity. The number is exact for the idealised model, an upper bound for a real one that meets the assumptions, and not for capacity planning.

## Production, and the review that shaped it

Each sheet is drawn as one SVG with a `viewBox` of 0 0 1587 1123 and literal colour values, not stylesheet variables. So the drawing lifts out unchanged, and the PDF exported from it, which is what's published, prints as one true A3 page.

Fonts are the production risk. Source Sans 3 loads from Google Fonts, which GitHub's SVG rendering can't fetch, so a sheet viewed there would reflow in a fallback face and lose its fit. An SVG published beside a PDF will have to embed the face or convert its text to outlines. Whether `oklch()` colours survive the export is checked on the first one.

Numbers on a sheet are of three kinds, and the summary side says which each is:

- A requirement or a constraint, such as the two seconds or the 80 MB, cites its id.
- So does a fact about the vendor, such as the 40 MB router image.
- Numbers from the running example are guarded by a unit test. It reads the shipped model through the adapter, runs the rollup, and checks every number in the L2b table against the SVG. So a change to the example fails a test before it quietly dates the sheet.

The capacity service's own tests wouldn't catch that. They build their representations by hand and never read the model file.

The sheets reuse the cookbook's layout rules and none of its artwork, whose licence is unverified. The [sheet index](../a3/README.md) records that against each sheet.

Most of what reads above as a careful distinction came from reading the design back against the research and the record. The method paragraph had presented the cookbook's placements and the L0, L1, L2 hierarchy as rules, where the research records the first as guidance and the second as later practice. It had also missed that the cookbook's template calls the summary side the front. L2b's worked example gained its fourth row, the ingest-to-3000 state that box 6 needs, and its derived verdicts. L0 gained the demo's answer beside the vendor's quote. The safeguard on the numbers had been left to the capacity service's tests.

Beside the method's own test of the condensing sits the reader's test. A reader of the kind the sheet names as its audience answers its goal question in fifteen minutes, without the record. Every published case ran a review loop of two to four weeks, and a public repository maintained by one person has nothing like it. So the design's mechanism is an issue template asking such a reader for the answer and the time it took.

A sheet is finished when it answers its goal question, and it ships at that point without waiting, because the method's documented failure mode is a sheet that is never finished at all. What the issue template collects afterwards is what sends a sheet back for re-issue.

Both drafted sheets are published on those terms, approved without external review, and the sheet index says so in its status column. L1 and L2a were to follow the first implementation phases, once the schemas and the projection stopped moving. By release v0.3.0 neither had been drafted, and the sheet index lists both as outstanding work. So much for the documented failure mode.

---

Previous: [Five views and twenty-six decisions](06-five-views-and-twenty-six-decisions.md) · Index: [Federating a systems model](../README.md) · Next: [Planning the build](08-planning-the-build.md)

# Planning the build

*Roar Georgsen, 27 August 2026*

Part 9 of 13 in [Federating a systems model](../README.md), written for release v0.3.0.

> [!IMPORTANT]
> **The story so far**
>
> The demo is a SysML v2 model of a five-server query pipeline, served as one subgraph of a federated GraphQL graph. Beside it sit a capacity service that finds the pipeline's bottleneck and a document service that owns a requirements document, all in one container. The design phase is over. This part is about the plan that turned the design into code, and the things the plan had to decide for itself.

## From record to plan

The design ended with four gates approved, twenty-six decision records and an architecture description that names every type, field and process. What it didn't have was a sequence of tasks. The plan had to be detailed enough to build from without deciding anything again, and it opens on the finish line:

> Build the pipeline demo the design phase specified: a generic SysML v2 adapter subgraph, a capacity service, a document service, two vanilla web apps and a Cosmo router, shipped as one container image launched by one `docker run`

The records had already fixed the stack:

| Part | Choice |
|---|---|
| Language | Go 1.27 |
| The three services | <span class="term" data-term="gqlgen">gqlgen</span> 0.17.94, with federation and WebSocket subscriptions |
| Router | Cosmo router 0.343.1, run as a child process |
| Composition | wgc 0.130.1, run by hand, with its output [committed](../decisions/AD-0012-composition-committed.md) |
| Web apps | plain <span class="term" data-term="es-modules">ES modules</span> and one vendored file, SortableJS ([no build step](../decisions/AD-0017-vanilla-web-apps.md)) |
| Image | a <span class="term" data-term="distroless">distroless</span> base, [published on a version tag](../decisions/AD-0020-publish-on-tags.md) |

The plan came in three layers:

1. a top-level document with one line per task
2. one detail document per phase, with the full text of every task
3. a map of every requirement against the task that meets it and the test or demonstration that proves it

With the top-level document and the one task in hand, I had enough to work from.

## Five phases, one pull request each

Before the five phases of building came a phase zero, with its own pull request like the rest.

| Phase | What it builds |
|---|---|
| 0 | repository policy, and the spikes that settle what the documentation couldn't |
| 1 | the example model and the adapter core |
| 2 | the three services, composed into one graph |
| 3 | the supervisor: one binary, one process tree, one port |
| 4 | the two web apps |
| 5 | the image, and the first release |

Phase 0 runs the spikes, small experiments that answer the questions the architecture couldn't settle from reading, each with a pass criterion and a fallback written down first. Four of the five run there. The fifth, a run with no network, needs the built image and waits for Phase 5. [Five spikes before the first line](09-five-spikes-before-the-first-line.md) has what they found.

Every phase is one pull request. Tests come first, with the test as the specification, and a test that verifies a numbered requirement is named for it. `make check` runs after every task, and a coverage floor of 70 per cent is checked before every push. Requirements about browser behaviour get checklists, run by hand and recorded with the date and host.

The end-to-end proof is the worked example from [From use cases to requirements](05-from-use-cases-to-requirements.md), observed through the router and both apps. It runs from the shipped state failing at parse, through the bottleneck moving and the pipeline passing while one server fails its share, to reset putting everything back.

## Lines are estimates, not limits

One of the design's constraints is that the repository can be read in an afternoon, which means little unless somebody counts. So the plan carried line figures. The first estimate for the adapter was 2000 lines of hand-written Go. It lasted exactly as long as it took to write the adapter's tasks out in full:

| Adapter package | First estimate | In the written-out tasks | Revised |
|---|---|---|---|
| `syntax` | 800 | about 1261 | 1300 |
| `model` | 650 | about 932 | 1000 |
| `projection` and `serve` | 550 | about 389 | 450 |
| Total | 2000 | about 2582 | 2750 |

The other components got estimates too: 600 lines for each example service, 700 for the supervisor, and 900 lines of JavaScript and 300 of CSS for each web app.

Having to revise the figure the moment real text existed settled what kind of number it is ([the line figures are estimates](../decisions/AD-0028-line-figures-are-estimates.md)). Correctness comes first, and a component that needs more lines to be right takes them. The counts are still measured at each phase's close, because the scale is worth seeing, but nothing is trimmed to fit.

A figure written as a limit reaches for the wrong things first. The cheapest lines to give up are the doc comments, the error messages that say where the problem is, and the guards that make a refusal honest. Those are exactly the parts a reader with an afternoon most needs.

## What the plan decided for itself

A plan detailed enough to build from has to decide things the design didn't, and say so. Most were small. Two are worth telling.

**The model is immutable once built.** An edit returns a new model with the version bumped, and the store swaps one pointer. So a query can never see the served text and the projection disagree, even with many readers at once.

**The editable numbers moved out of the text.** [The viewer decision](../decisions/AD-0026-viewer-shows-text-and-wiring.md) had put inline inputs at each number's position in the model text. But the projection carries values, not source positions, so placing an input there would have meant the browser searching the text for the number, a second and weaker parser of a language the adapter has already parsed. The plan put the editable numbers in a panel above the text instead. The served text still shows each edited number where it was, because the adapter patches it there. I still think the panel is right for a demo whose point is the model text, and the decision record was amended to match.

Two points stayed open into the build. The escape hatch for elements nobody has projected yet, which [the curated projection decision](../decisions/AD-0005-curated-generic-projection.md) leaves open, is still unbuilt, so a model using a construct outside the subset doesn't load. And inline inputs would need source positions in the adapter's schema, which nobody has decided are worth adding.

---

Previous: [An A3 sheet for a fifteen-minute reader](07-an-a3-sheet-for-a-fifteen-minute-reader.md) · Index: [Federating a systems model](../README.md) · Next: [Five spikes before the first line](09-five-spikes-before-the-first-line.md)

# Five spikes before the first line

*Roar Georgsen, 27 August 2026*

Part 10 of 13 in [Federating a systems model](../README.md), written for release v0.3.0.

> [!IMPORTANT]
> **The story so far**
>
> The demo serves a SysML v2 model of a query pipeline through a federated GraphQL graph: one service reads the model, a second computes capacity and verdicts, a third owns a requirements document, and a router joins them. Part 9 laid out the plan to build it. Before any of that plan's code, four questions had to be answered by experiment, and a fifth waited for the finished image.

## Why spikes first

> [!NOTE]
> **Spike**
>
> A small, throwaway experiment that answers one question before anything is built on the answer. Each spike here had a pass criterion and a fallback written down before it ran, so a failure would switch the plan to the fallback instead of starting an argument.

The architecture left a handful of questions that reading the documentation couldn't settle. Five became spikes. Four ran in a phase of their own ahead of the first line of product code, and the fifth, which needed the built image, ran at the end.

| Spike | The question | What it found |
|---|---|---|
| Syntax | Which port and connection syntax do the reference validators accept? | Accepted after a one-token change, and my pass criterion was broken |
| Nested requires | Does `@requires` over nested lists survive composition and the router? | Yes, end to end |
| A router with no configuration file | Does the router start from environment variables alone? | Yes, but its readiness check says less than it seems |
| Copying across platforms | Does copying the router out of the vendor's image pick the right processor architecture? | Yes |
| No network | Does the container reach anything outside itself? | No |

None failed outright, and each of the first four corrected something in the plan.

## The syntax, and a test that couldn't fail

At that point the design had left ports and connections off the list of constructs the parser accepts, because nobody had yet checked the official examples. The spike found the shapes in the OMG's training material and wrote a small probe model: the example cut down to two servers, one connection and three requirements. The two reference tools, the <span class="term" data-term="pilot-implementation">OMG pilot implementation</span> and <span class="term" data-term="opensysml">OpenSysML</span>, were to check it ([SysML 2.0 as the target](../decisions/AD-0019-sysml-2-0-target.md)).

They disagreed. OpenSysML accepted the probe. The pilot refused it at both `satisfy` lines, because the two satisfied requirements had each bound their subject once and the `satisfy` statement bound it a second time. The fix was one token: declaring the subject as `subject target :> pipeline;` (a subsetting) rather than `= pipeline` (a binding). With that change both tools accepted the probe. They took port definitions carrying directed items, such as `in item queries : Query;`, and connections written as `connect ingest.output to parse.input;`. They took the plain numeric binding the edit panel changes, `attribute :>> throughput = 2000;`, as well as the arithmetic that binds a derived limit and the millisecond unit the model declares for itself. The design had committed to both tools from the start, and this disagreement is why [the demo's own model is checked by both in CI](../decisions/AD-0030-model-validation-in-continuous-integration.md) too.

Then came the part I'm least proud of. The pilot runs as an interactive shell, and writes its prompt `1> ` without a newline, so the first error of a run lands on the same line as the prompt. My pass criterion was a search for lines starting with `ERROR` or `WARNING`, and it couldn't see that one. So I broke the probe on purpose to test the test. The breakage produced exactly one error, the search printed nothing, and nothing was the plan's own signal for a pass. As written, my pass criterion would have passed a broken file.

The spike's result still stood, because the criterion also required the tool to print the model's root element. The corrected search allows for the prompt. In all, three deliberate breakages, that first one among them, confirmed that a clean pass means something, and both tools caught every one.

## Nested requires through the router

The capacity service computes its fields from data the adapter owns. It uses `@requires` to ask the router to fetch a part's children, their attributes and the connections between them from the adapter first. That's a selection nested two levels deep, and nothing in the research had confirmed Cosmo and the Go GraphQL library would carry it intact. The plan's fallback was to flatten the subtree into one JSON string.

For the spike I built the two services with fixed data and ran the query through the router. The answer came back byte for byte as predicted, with every part and connection delivered to the capacity service. So the fallback was never built.

The corrections came from the edges. Each service's schema is embedded in the composed output, at `.engineConfig.datasourceConfigurations[n].customGraphql.federation.serviceSdl`, and extracting it with `jq -j`, which adds no trailing newline, is what makes the drift test's comparison exact ([composition committed](../decisions/AD-0012-composition-committed.md)). And one file the code generator writes, the stub for `@requires`, carries no "generated" header. Its one function takes a map of the untyped kind the repository's rule against empty interfaces forbids, and generated files escape that rule only by carrying the header. So the stub needs an explicit allowance of its own, and since the allowance is tied to a line, it has to be put back after every regeneration ([generated code exempt](../decisions/AD-0016-generated-code-exempt.md)).

## A router with no configuration file

The design runs the router as [a child process](../decisions/AD-0010-router-as-child-process.md) of the demo's own binary, set up entirely through environment variables, with no YAML configuration file anywhere. The spike built a cut-down image and confirmed that works: the router starts, serves the playground, and answers queries. It also showed the playground answers a plain request with an empty page, so any check on it has to ask for HTML.

Readiness was the surprise.

> [!WARNING]
> **Ready doesn't mean answering**
>
> The router's `/health/ready` returned 200 with every service down, and a query then failed. Readiness means the router has loaded its configuration, and says nothing about the services behind it. The demo's start-up ordering doesn't depend on it, and the documentation says so instead of letting the container's health check imply more ([one binary, one port](../decisions/AD-0011-one-binary-one-port.md)).

## Copying across platforms

The image is built for both amd64 and arm64, and the router binary is copied out of the vendor's image rather than downloaded per platform. A build for both confirmed that each platform gets its own statically linked router, with no extra flags ([publish on tags](../decisions/AD-0020-publish-on-tags.md)). The plan's command for inspecting the binaries needed a small fix to its search.

## A container with no network

The last spike is the demonstration behind the claim that the demo opens no connection of its own. I ran the image with `docker run --network none` and debug logging, and left it alone for nine minutes, long enough for the router's usage tracker to try several times if anything were still tracking. The log never grew past its start-up lines, and searches for telemetry hosts, name lookups and timeouts found nothing.

In the end the interface list carries the claim more than the quiet log does. Inside the container the only network interface is the loopback, so there's nowhere for a packet to go, and an empty log on its own is only an empty log.

What that run couldn't show is the browser, since a container with no network publishes no port. So the check that both apps still load with the host itself taken off the network stayed open until 14 September. Then I reloaded both apps on Ubuntu under WSL with the host offline, and both drew in full, as the [example's verification record](https://github.com/Roarge/sysml-federation/blob/main/examples/pipeline/README.md#the-two-web-apps) shows.

## What changed in the plan

No decision was reversed and no new decision record was needed. The plan took a run of small corrections instead: a subject declared by subsetting, a search that tolerates the prompt, an exact schema extraction, a re-placed allowance after regeneration, and a few commands that were slightly wrong. None of them would have surfaced from reading. The broken pass criterion is the kind of defect that stays hidden until the first real refusal, which is exactly why it's worth breaking a check on purpose before trusting it.

---

Previous: [Planning the build](08-planning-the-build.md) · Index: [Federating a systems model](../README.md) · Next: [The demo as it shipped](10-the-demo-as-it-shipped.md)

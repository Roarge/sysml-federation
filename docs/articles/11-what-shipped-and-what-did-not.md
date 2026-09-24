# What shipped, and what did not

*Roar Georgsen, 29 August 2026*

Part 12 of 13 in [Federating a systems model](../README.md), written for release v0.3.0.

> [!IMPORTANT]
> **The story so far**
>
> The demo shipped as one container. Inside it was a SysML v2 model, a capacity analysis and a requirements document, answering as one graph behind a single router, with two web apps in front. [The demo as it shipped](10-the-demo-as-it-shipped.md) walked through it. This part gives the image's size, says how a release is guarded, and says what running it does and doesn't prove.

## How big it is

`docker run --rm -p 8080:8080 ghcr.io/roarge/sysml-federation` pulls about 45 MB on amd64, or 41.5 MB on arm64, and answers on port 8080 roughly two seconds after the container starts. There's no account to make and no login to run.

Read back from the registry, the first release came to 44,850,689 bytes on amd64 and 41,475,216 on arm64, and release 0.2.0 came in about 39,000 bytes heavier on each. The ceiling the publishing job enforces is 80,000,000 bytes per platform, so both sit a little over half way into their budget.

![The image, layer by layer, base at the bottom](../img/v4-image-layers.png)

*The image, layer by layer, base at the bottom. Cut from the [architecture views](../architecture/architecture-views.pdf).*

Almost all of it is somebody else's binary. The vendor's router accounts for nearly 40 MB of the amd64 total, which leaves under 5 MB for the base image, the Go supervisor with both web apps inside it, the configuration and the model. A size budget for this demo is really a budget for the router, and the number moves when the vendor's next release does.

One small trap catches anyone who types a version. The git tag is `v0.1.0`, but the image tag is `0.1.0`, because the publishing step drops the leading letter. So `ghcr.io/roarge/sysml-federation:v0.1.0` finds nothing. The untagged launch line avoids the question.

## A gate, not a report

The design as first accepted moved the `latest` tag in the same push as the version tag, and measured the image afterwards. That's the wrong order, and it's an easy mistake to write into a release job that looks correct. The registry's size can only be read after a push, so with `latest` already moved, an oversized image is what an untagged pull returns from the moment it goes up. The run that goes red later changes nothing about what a stranger has already downloaded. A check that runs after the thing it guards is a report, not a gate.

So what ships works the other way round. The version tag goes up alone. Both platforms are measured at the exact digest the build reported, and from v0.3.0 each is then pulled with no credential and started on a runner of its own architecture. Only once both have passed does `latest` move onto that digest ([publishing on tags](../decisions/AD-0020-publish-on-tags.md)). A run that fails leaves `latest` pointing at the last release that passed.

## What the no-network run proves

[Five spikes before the first line](09-five-spikes-before-the-first-line.md) ran the image with its network removed, and it went quiet: nothing attempted, nothing retried, nothing timed out, and the only network interface inside was the loopback.

That's a real result, and it's narrower than it sounds. It shows the stack starts, reports itself healthy and stays quiet with no route anywhere, which readiness alone never showed. It doesn't show what the image does when a network *is* present, and silence from a process with nowhere to send anything is weaker evidence than silence from one that could. For that case the demo relies on the environment variables that switch the router's telemetry off, and the router does log that usage tracking is disabled when they're set.

## Two checks, run late

Two checks were still open when the rest had been run, and I ran both on 14 September on Ubuntu under WSL. Both are in the [example's verification record](https://github.com/Roarge/sysml-federation/blob/main/examples/pipeline/README.md#the-two-web-apps).

The first is the offline reload. A reload of the viewer makes all its requests to `localhost:8080`, with no font and no other host among them. What had never been watched was the page still drawing with the machine itself taken off the network. With the host offline, both apps reloaded and drew in full.

The second is the same-origin check on the services' subscription socket. It had been proved by hand, with a handshake sent from a foreign origin coming back `403 Forbidden` from both services that serve subscriptions. A page served from another origin, opened in a browser, has now tried to open the socket, and the socket was refused.

## Where it runs

I build and test the demo on Ubuntu under WSL, and nowhere else by hand. The image is built for amd64 and for arm64, which is the one Macs with Apple silicon need. From release v0.3.0, continuous integration also pulls the published image and starts it on GitHub's Linux runners, one amd64 and one arm64, before `latest` moves. I haven't tried it on a real Mac of my own, and I make no promises beyond what those runs show.

## What it is

What's here is one command. It puts a model file, an analysis that has never read a model file and a document that has never computed anything behind a single endpoint, with two pages in front that work nothing out for themselves.

The demo isn't a product, and it isn't an implementation of the SysML v2 API. It reads a file instead of fronting a repository, its parser covers a fraction of the notation, it forgets every edit when it stops, and its capacity arithmetic is deliberately simple. The claim it was built to test is narrower than any of that, and it's the one thing the running image does demonstrate. Three services that share no code and know nothing of each other answer one query about one requirement, and a stranger with Docker can watch them do it.

---

Previous: [The demo as it shipped](10-the-demo-as-it-shipped.md) · Index: [Federating a systems model](../README.md) · Next: [A model of the demo itself](12-a-model-of-the-demo-itself.md)

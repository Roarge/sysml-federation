# AD-0011 One binary, one process tree, one port, one UI server

Status: accepted. Date: 2026-08-27.

## Context

D5 fixed the launch shape: one image, one
`docker run --rm -p 8080:8080 ghcr.io/roarge/sysml-federation`, with the
apps at `/viewer` and `/document`, the playground at `/playground` and `/`
redirecting to the viewer. Behind that line stand three subgraphs, a
router, two web apps and whatever serves them, and the decision here is how
they are arranged inside the container. The router is a child process from
the copied binary (AD-0010), which fixes one of the processes. D4 fixed the
apps as vanilla JavaScript served from a Go binary with embedded assets.

The base image constrains the arrangement. Distroless static has no shell
(C-50) and no Python runtime, so s6-overlay, supervisord and a `wait -n`
wrapper cannot run in it, and a Dockerfile `HEALTHCHECK` has nothing to
execute unless the check is built into the image's own binary (C-55). The router
binary has no probe subcommand (C-56). Docker's guidance calls one process
per container a rule of thumb rather than a hard rule and allows several
processes where the container has one concern (C-54).

The port question is separate. The router listens on `localhost:3002` by
default with the playground at `/` and the endpoint at `/graphql` (C-09).
Go's `httputil.ReverseProxy` flushes `text/event-stream` immediately and
proxies `101 Switching Protocols`, so SSE and WebSocket both pass through
it, and one origin on one port needs no CORS configuration, whereas a
two-port topology would depend on the router's CORS defaults, whose
acceptance by browsers for credentialed requests was not checked (C-62).
The research left the one-port or two-port choice as an open question,
with the remark that one port simplifies the README line and hides the
router as a separately addressable component (the research notes on
packaging).

The gate 2 review's third design-level finding settled where the UI server
lives. P17 places the single UI server that serves both apps and proxies
the router in the image element, so UI-MODEL and UI-DOC are the two apps
and nothing else (the engineering log, gate 2). SC-06 gives the supervisor
with its UI server and command dispatch a budget of 700 lines.

## Decision

We will ship one Go binary, `sysml-federation`, whose default `serve`
subcommand is PID 1 and runs the three subgraphs as goroutines on loopback
ports 3011, 3012 and 3013, starts the router as a child on 127.0.0.1:3002,
and runs one UI server on 0.0.0.0:8080 that serves both apps from
`embed.FS`, proxies `/graphql` and `/playground` to the router and
redirects `/` to `/viewer` (P3, P17). The subcommands `adapter`,
`capacity`, `document` and `ui` run one component each on a given address
for anyone who wants the services apart, and `healthcheck` GETs the
router's `/health/ready` on its loopback port and `/viewer` on the
published port, exiting non-zero if either fails, for the image's
`HEALTHCHECK` (architecture V4). The router's health path is not proxied.

## Alternatives considered

Docker Compose with one container per service against the official router
image, the shape the vendor's own demos use and the lowest-risk one in the
Cosmo research. Every subgraph is then a visibly separate service, which
suits the README's argument. It lost because "one command" would be
`docker compose up` after obtaining a file, which needs Compose and either
a clone, a git-backed URL or an OCI artefact, and the routing URLs in the
execution configuration would differ from the single-container layout, so
a second configuration would be needed (the research notes on packaging).
The engineering log's D5 entry records that one command is literally true
only with a single image.

Both shapes at once, the same image started several times by a compose
file beside the official router container. The packaging research
recommends it as a second entry point and the engineering log lists it as an
alternative offered under D5. It carries the second configuration and the
file-fetching problem of the compose route and was not taken.

One image per service, four images plus the router. The cleanest
separation and closest to how an adopter would deploy, and the most
artefacts to build, publish and carry across an air gap (the research notes
on packaging). Not taken for a demo.

s6-overlay, supervisord or a shell wrapper as the in-container supervisor.
Each needs a userland the distroless base does not have (C-55), and s6
would add a supervision layer to explain in a repository meant to be read
in an afternoon. tini or `docker run --init` supervises exactly one child
and would put a flag in the README line (C-58, the research notes on
packaging).

Two published ports, the router on 3002 and the apps on 8080. It keeps the
router addressable on its own. It lost on the launch line, which would
carry two `-p` flags, and on the cross-origin rules of C-62 that one origin
avoids.

## Consequences

The visitor gets one port and four paths (SR-04), the apps reference
nothing outside that origin (SR-10), and every request from an app reaches
the router through the proxy (SR-40). Startup is one process tree with an
ordering the supervisor controls, so SR-02's ten seconds is measured in one
place. The `healthcheck` subcommand gives the image a `HEALTHCHECK` that a
router container on its own could not carry (C-56).

The proxy hides the router. A visitor who wants to see the three services
as separate processes has the subcommands, and nothing on the published
port shows where the router's responsibility ends and the UI server's
begins. A crash in any goroutine or in the child takes the whole demo down,
which the packaging research calls acceptable for a demo while noting that
it hides the process boundaries the argument is about.

The supervisor carries the ordering risk of C-57: what `/health/ready`
reports with the subgraphs down is unverified, so the subgraphs are started
first and the published port is opened last, and the spike settles what
the health check proves. C-20's spike settles whether the child starts from
environment alone. The UI server also sets `Cache-Control: no-cache` and
answers 404 for directory paths, because `embed.FS` sends no cache headers
and lists directories by default (C-63).

## Requirements affected
SR-01, SR-02, SR-04, SR-10, SR-40

## Sources
The design brief D5, P3, P17 (with D4), the constraints list C-09, C-50, C-54, C-55, C-56, C-57, C-58, C-62, C-63, the requirements list SR-02, SR-04, SR-10, SR-40, SC-06, the architecture description V4, the engineering log's D5 and gate 2 entries, the research notes on packaging, options and open questions, the research notes on Cosmo, options.

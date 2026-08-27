# AD-0014 Subscriptions as version events with client refetch

Status: accepted. Date: 2026-08-27.

## Context

D2 chose live push through the router over polling and manual refresh. The
bottleneck moving in one app while the edit is made in the other is the moment
that sells the demo, and the README's "responds at once" only holds with push.
D2 left open what is pushed, who pushes it and how it reaches the browser.

The router subscribes to a subgraph over `ws` with subprotocol `auto`, `sse`
or `sse_post`, and serves subscriptions to clients over WebSocket, over SSE on
a GET or POST that sends `Accept: text/event-stream`, or as multipart (C-11,
C-12). None of that needs a broker. Cosmo Streams with Kafka, NATS or Redis is
an alternative design in which subgraphs stay stateless, and nothing on the
local compose path references it (C-13). Router plugins and remote gRPC
services keep the SDL as the contract but do not support subscriptions, so a
subgraph on either route could push nothing to the two apps (C-14).

Whether the router resolves entity fields owned by another subgraph inside a
subscription payload is a question the plan records as unverified, and a
subscription that carried a whole requirement with its verdict would depend
on it. There is no model repository behind the adapter, so versioning is
stood in for by a counter: `Model.version` carries it, every accepted
mutation increments it, and the document service keeps its own (P10, C-92).

On the browser side the vendor's SSE page documents a fetch POST with
`Accept: text/event-stream` read through a `ReadableStream`, with no retry
logic in its example, and says of `EventSource` that a compatible interface
"is not very difficult to add, but we haven't seen the need for it yet"
(C-61). One SSE subscription is one HTTP connection, and an HTTP/1 browser
allows about six per origin, so two open tabs, the viewer with one
subscription and the document with two, hold three against it (C-66). Both
apps share one origin through the UI server's reverse proxy, which flushes
`text/event-stream` at once and proxies WebSocket upgrades (C-62, AD-0011).

## Decision

We will make each subscription carry nothing but a version number and make
each app refetch its whole query when one arrives. The adapter exposes
`Subscription.modelChanged: Int!` and emits the new `Model.version` after every
accepted mutation, the document service exposes
`Subscription.documentChanged: Int!` and does the same for its own counter, and
on every event the viewer and the document re-run their ordinary query against
the router, which plans it across the three subgraphs as it would any other
read. Towards the router both services serve subscriptions through gqlgen's
WebSocket transport, registered in the compose input with protocol `ws` and
subprotocol `auto`. Towards the browser the shared `graphql.js` module opens
SSE over fetch POST with `Accept: text/event-stream`, reads the stream with a
`ReadableStream` reader that parses `event:` and `data:` lines, and reconnects
on drop.

## Alternatives considered

Polling every one or two seconds, or manual refresh, the two alternatives the
engineering log records as losing to D2. Polling is trivially simple and needs
no Subscription root anywhere, but latency equals the poll interval and
"responds at once" reads weaker, and a manual refresh reads weaker still. The
UI research keeps polling as the fallback if no subgraph exposes a
Subscription root.

Refetching after one's own mutation with no subscription at all, an open
question in the Cosmo research. The app that made the edit would be current
and the other tab would not, which is exactly what D2 rules out.

Subscriptions carrying the changed data. A payload with verdicts or document
numbers in it needs the router to resolve other subgraphs' entity fields inside
a subscription response, which is unverified. A version event sidesteps the
question because the refetch is an ordinary query (plan decision 6).

Cosmo Streams with an embedded or sidecar NATS. Subgraphs would stay stateless
and any service could publish, but it adds a broker process to a demo built
on as few processes as possible and EDFS directives to the schemas (C-13).

SSE from the subgraphs to the router, `sse` or `sse_post`, in place of
WebSocket. Simpler subgraph code with no WebSocket library, but one way only
and needing the POST variant for large queries. gqlgen's WebSocket transport is
the documented default and the router has multiplexed upstream connections
since 0.313.0 (the research notes on Cosmo).

Native `EventSource` over GET in the browser. Built-in reconnect and the fewest
lines, but disclaimed by the vendor's docs and never executed against a router
in the research (C-61, the research notes on the UI). A browser WebSocket with a
`graphql-transport-ws` handler is bidirectional, but its handshake state
machine is more code than two apps with one or two subscriptions each need.

## Consequences

The update path is one mechanism. An accepted mutation increments a counter,
the counter is pushed, the app refetches, and the capacity service recomputes
from the representation the router carries through `@requires` (AD-0007). No
service needs a change feed and nothing is kept between requests, which is
what SR-32 requires and what makes the view live rather than exported. Reset
travels the same path, since `resetModel` and `resetDocument` each emit their
version event (SR-44).

A refetch costs a full query per event per open tab. At demo scale that is
microseconds of recomputation and a few kilobytes on the wire. V3 Runtime's
"nothing moves" case shows the honest side of it: raising `ingest` fires the
event and refetches both apps although only one number changes, because the
demo detects "nothing changed" nowhere. Many tabs would multiply the cost, so
the design suits a demo and not a fleet of tabs (C-66), and the scaling
story, a cache per model version, stays one sentence on the A3 sheet.

The browser side carries its own reconnect logic because the documented path
has none. It is written once in the shared module and used by both apps. The
two-second bound of SR-39 is demonstrated rather than tested, because the
repository carries no browser automation.

Choosing `ws` towards the subgraphs makes gqlgen's WebSocket transport a
dependency and means both pushing subgraphs hold connections from the router.
It also fixes what C-14 rules out: the adapter and the document service must
stay standard GraphQL subgraphs for as long as they push, while the capacity
service, which has no subscription, is untouched by that limit.

## Requirements affected
SR-26, SR-27, SR-32, SR-39

## Sources
The design brief D2, P6 and P10. The constraints list C-11, C-12, C-13, C-14, C-61, C-62, C-66 and C-92. The design-phase plan, decision 6. The engineering log, design phase planning, D2 and the rollup evaluation. The research notes on Cosmo, the live updates option and its open questions. The research notes on the UI, the transport options. The sceptic verdicts on the research, the subscriptions verdict. The architecture description V2 Composition and V3 Runtime.

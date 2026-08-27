# AD-0017 Vanilla web apps with one vendored file

Status: accepted. Date: 2026-08-27.

## Context

The demo puts two web apps in front of the router so that the sharing of
data is seen rather than asserted: a model viewer that renders the SysML v2
source with editable numbers beside a sketch drawn from the wiring
(AD-0026), and a requirements document whose owner reorders, nests and
annotates it (AD-0025).
The README says of the interface for making a change that it "is not
settled yet and does not matter to the argument", and it closes with the
promise of "something small enough to read in an afternoon". Every
line of browser code counts against that promise.

The repository is a Go module with one dependency and no Node toolchain.
CI runs the unit tests and an image publish on version tags, and nothing
in the build knows how to bundle, transpile or minify JavaScript. The apps
also have to work with the host offline, because a font or script fetched
from another origin would break the launch story and the air-gap claim
without any service being at fault (SR-10).

The research settled what a browser can do here without a library. No SysML
v2 text renderer exists that can be vendored as one permissively licensed
file. Drag and drop over nested lists is the one place where
hand-written code would run to sixty lines at the least and a hundred and
fifty at the most, and SortableJS 1.15.7 covers it in a single MIT file of
45,478 bytes with no dependencies. Live updates reach a browser from
the Cosmo router over SSE on a fetch POST, the path the vendor documents.
Go's standard library serves an embedded directory with correct MIME
types and no cache headers, and its reverse proxy carries event streams and
WebSocket upgrades, so one origin needs no CORS configuration.
Native ES modules do not load from file URLs, HTML5 drag and drop is
missing from two mobile browsers and the brief states desktop first,
and one SSE subscription holds one of the six HTTP/1 connections a browser
allows per origin.

## Decision

We will write both apps as plain HTML, CSS and native ES modules with no
build step and no node_modules, embed them in the Go binary and serve them
from `embed.FS` through a wrapping handler that sets `Cache-Control:
no-cache`, send queries and mutations by fetch POST, read live updates
over SSE on a fetch POST with a reconnecting ReadableStream reader, draw
the sketch as inline SVG, render the model text with a hand-written
tokeniser of about forty lines, and vendor exactly one third-party file,
SortableJS 1.15.7 under MIT, for drag and drop in the document app. The UI
server that serves both apps and proxies the router belongs to the image
rather than to either app, and the apps are clients of the router
and of nothing else (SR-40).

## Alternatives considered

Go templates with HTMX. It lost for the reason the whole choice went the
other way: the repository is meant to be read in an afternoon and carries no
Node toolchain.

A framework with a build step. It lost on the same
grounds. A bundler brings a Node toolchain into a Go repository, its output
is not what a reader reads, and CI would either have to run it or the
committed output would drift.

For the model text, Prism core or highlight.js core vendored with a
hand-written grammar, or tree-sitter compiled to WASM. Either library still
needs the grammar written here, so it
replaces about thirty lines of tokeniser at the cost of a second vendored
file and its notice, and the tree-sitter route adds a build step of its
own. Reusing an existing SysML v2 renderer was rejected outright because
none is a permissively licensed static browser asset.

For drag and drop, hand-written HTML5 drag and drop at about sixty lines, or
a pointer-events implementation at a hundred to a hundred and fifty. The
first inherits the mobile-browser gaps, the second
must handle auto-scroll and keyboard access itself, and the research
recommends it only if no third-party JavaScript becomes a rule.

## Consequences

No Node, no bundler and no node_modules anywhere in the repository, which
SC-05 states as a design constraint. A reader opens `examples/pipeline/ui/`
and reads the files the browser runs. The figures in SC-06, 900 lines of
JavaScript and 300 of CSS per app with a shared module counted once, say
what that reading should cost, and the vendored file is not hand-written and
sits outside them.

The tokeniser is a tokeniser and not a parser, and it owns its
keyword table, which cites the specification's reserved-word list rather
than the pilot's EPL-2.0 grammar.

SortableJS is 45 KB that a reader will not read, and it uses native HTML5
drag and drop underneath with its own fallback for the browsers that lack
it. Its MIT text sits beside the file, the root `NOTICE`
names it, and the file must stay out of any directory named `vendor/` or
the allowlist ignores it twice over (SR-08, SC-04). The implementation
phase also adds `*.html`, `*.css` and `*.js` under `examples/` to the
allowlist, since no rule tracks them today.

The apps cannot be opened from disk, so the Go binary is the only way to run
them. Reconnect logic for the SSE reader is written here because the
vendor's example has none. A viewer tab and a document tab hold three
long-lived connections against the HTTP/1 limit of about six per origin,
which suits a demo and not many tabs. Embedded files carry a zero
modification time, so the handler's `no-cache` is what stops a browser
serving a stale app.

Behaviour that lives in browser JavaScript is verified by demonstration
against a scripted checklist, because the repository carries no browser
automation and SC-01 permits none, so SR-11, SR-12 and SR-34 are
demonstrated rather than tested. No spike belongs to this decision.

## Requirements affected

SR-10, SR-11, SR-12, SR-34, SC-05, SC-06

## Sources

SortableJS 1.15.7 and its MIT licence. The Cosmo documentation on receiving subscriptions in a browser over SSE. Go's `embed` and `net/http/httputil` documentation. The SysML 2.0 language specification's reserved-word list. [The demo being built](../articles/10-the-demo-being-built.md) for the two apps as they ship.

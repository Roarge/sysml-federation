# AD-0017 Vanilla web apps with one vendored file

Status: accepted. Date: 2026-08-27.

## Context

The demo puts two web apps in front of the router so that the sharing of
data is seen rather than asserted: a model viewer that renders the SysML v2
source with editable numbers beside a sketch drawn from the wiring (D1), and
a requirements document whose owner reorders, nests and annotates it (D7).
The README says of the interface for making a change that it "is not
settled yet and does not matter to the argument", and it closes with the
promise of "something small enough to read in an afternoon" (C-79). Every
line of browser code counts against that promise.

The repository is a Go module with one dependency and no Node toolchain
(C-76). CI runs the unit tests and, after D10, an image publish, and nothing
in the build knows how to bundle, transpile or minify JavaScript. The apps
also have to work with the host offline, because a font or script fetched
from another origin would break the launch story and the air-gap claim
without any service being at fault (SR-10).

The research settled what a browser can do here without a library. No SysML
v2 text renderer exists that can be vendored as one permissively licensed
file (C-59). Drag and drop over nested lists is the one place where
hand-written code would run to sixty lines at the least and a hundred and
fifty at the most, and SortableJS 1.15.7 covers it in a single MIT file of
45,478 bytes with no dependencies (C-60). Live updates reach a browser from
the Cosmo router over SSE on a fetch POST, the path the vendor documents
(C-61). Go's standard library serves an embedded directory with correct MIME
types and no cache headers, and its reverse proxy carries event streams and
WebSocket upgrades, so one origin needs no CORS configuration (C-62, C-63).
Native ES modules do not load from file URLs (C-64), HTML5 drag and drop is
missing from two mobile browsers and the brief states desktop first (C-65),
and one SSE subscription holds one of the six HTTP/1 connections a browser
allows per origin (C-66).

The choice was put to the owner during planning and recorded as D4, with
SortableJS named separately as P13.

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
rather than to either app (P17), and the apps are clients of the router
and of nothing else (SR-40).

## Alternatives considered

Go templates with HTMX, offered to the owner beside the chosen option (the
engineering log, D4). It records the reason the whole choice went the
other way: the repository is meant to be read in an afternoon and carries no
Node toolchain.

A framework with a build step (the engineering log, D4). It lost on the same
grounds. A bundler brings a Node toolchain into a Go repository, its output
is not what a reader reads, and CI would either have to run it or the
committed output would drift.

For the model text, Prism core or highlight.js core vendored with a
hand-written grammar, or tree-sitter compiled to WASM (C-59, research on
the UI). Either library still needs the grammar written here, so it
replaces about thirty lines of tokeniser at the cost of a second vendored
file and its notice, and the tree-sitter route adds a build step of its
own. Reusing an existing SysML v2 renderer was rejected outright because
none is a permissively licensed static browser asset.

For drag and drop, hand-written HTML5 drag and drop at about sixty lines, or
a pointer-events implementation at a hundred to a hundred and fifty (C-65,
research on the UI). The first inherits the browser gaps of C-65, the second
must handle auto-scroll and keyboard access itself, and the research
recommends it only if no third-party JavaScript becomes a rule.

## Consequences

No Node, no bundler and no node_modules anywhere in the repository, which
SC-05 states as a design constraint. A reader opens `examples/pipeline/ui/`
and reads the files the browser runs. The budgets in SC-06, 900 lines of
JavaScript and 300 of CSS per app with a shared module counted once, bound
what that reading costs, and the vendored file is not hand-written and
sits outside them.

The tokeniser is a tokeniser and not a parser (C-59), and it owns its
keyword table, which cites the specification's reserved-word list rather
than the pilot's EPL-2.0 grammar (C-82).

SortableJS is 45 KB that a reader will not read, and it uses native HTML5
drag and drop underneath with its own fallback for the browsers that lack
it (C-60, C-65). Its MIT text sits beside the file, the root `NOTICE`
names it, and the file must stay out of any directory named `vendor/` or
the allowlist ignores it twice over (SR-08, SC-04). The implementation
phase also adds `*.html`, `*.css` and `*.js` under `examples/` to the
allowlist and to `check-allowlist`, since no rule tracks them today.

The apps cannot be opened from disk, so the Go binary is the only way to run
them (C-64). Reconnect logic for the SSE reader is written here because the
vendor's example has none (C-61). A viewer tab and a document tab hold three
long-lived connections against the HTTP/1 limit of about six per origin,
which suits a demo and not many tabs (C-66). Embedded files carry a zero
modification time, so the handler's `no-cache` is what stops a browser
serving a stale app (C-63).

Behaviour that lives in browser JavaScript is verified by demonstration
against a scripted checklist, because the repository carries no browser
automation and SC-01 permits none, so SR-11, SR-12 and SR-34 are
demonstrated rather than tested. None of C-59 to C-66 opens a spike.

## Requirements affected

SR-10, SR-11, SR-12, SR-34, SC-05, SC-06

## Sources

The design brief D1, D4, D7, P13, P17. The constraints list C-59 to C-66, C-76, C-79, C-82. The engineering log, design phase planning entry, D4. The requirements list SR-10, SR-11, SR-12, SR-34, SR-40, SC-01, SC-04, SC-05, SC-06. The architecture description V2 The two web apps, V4 Deployment. The research notes on the UI, options. The design-phase plan, D4 and decision 13.

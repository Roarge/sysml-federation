# AD-0026 The viewer shows the model's text beside a sketch of its wiring

Status: accepted. Date: 2026-08-27.

## Context

The demo is about SysML v2, and the first thing a visitor sees decides
whether the language reads as approachable. The owner was offered four
forms of viewer at planning (D1): the SysML v2 text with editable numbers,
the text plus a small sketch of the wiring, a structured outline without
notation, and a diagram first with no notation. The bottleneck moving is
the memorable moment (the design brief, "Purpose"), and it is visible at a
glance only on a drawing of the wiring, while the notation is the point of
showing SysML at all. No permissively licensed browser renderer for SysML v2
exists (C-59), so whatever shows the text has to be written here, and the
apps are plain HTML and ES modules with no build step (D4, AD-0017).

## Decision

We will make the model viewer a text pane and a sketch side by side. The
text pane renders the model's source with keywords, names, literals,
strings and comments distinguished by a small tokeniser whose keyword table
cites the specification's reserved words (SR-11, C-82), and renders the
editable set, the throughput of each server and the limit of the global
throughput requirement, as inline inputs at the literals' positions, with
every other value read-only (SR-13). The sketch is drawn from the model's
connections as a left-to-right graph showing each server's throughput, the
pipeline's capacity and the bottleneck servers marked in red (SR-12). Each
requirement shows its verdict and reason (SR-14), and a failing requirement
block is red, the only colour accent in the viewer (SR-15). Everything the
viewer shows arrives through the router, and it computes nothing.

## Alternatives considered

Text alone, with the numbers editable inline. Simplest, and the strongest
signal that this is SysML v2, but the migration of the bottleneck would be
read off numbers rather than seen, and the README's "reliably surprising
before it is seen" depends on seeing it.

A structured outline of packages, parts and attributes with input fields
and no notation. Easiest for a reader who has never seen SysML, and it
hides the language the demo is about.

A diagram first, with requirements listed beside it and no notation. The
most visual, the furthest from a SysML viewer, and the most drawing work.

## Consequences

The viewer carries two renderers, a tokeniser of roughly forty lines and a
layered left-to-right layout of a small graph, both hand-written and both
within SC-06's budget. The text pane and the projection can never disagree,
because the adapter patches edited literals into the served text (SR-22)
and the viewer renders that text. The sketch shows the capacity and the
bottleneck exactly as the capacity service returns them, so the viewer is a
faithful client and never a second implementation of the rollup. The
tokeniser is not a parser and is verified by demonstration and inspection
rather than by a test (SR-11), which is the price of keeping browser
automation out of the repository (SC-01).

## Requirements affected

SR-11, SR-12, SR-13, SR-14, SR-15

## Sources

The design brief D1, D3, D4, "The two apps", "The example model". The
requirements list SR-11 to SR-15, SR-22, SC-01, SC-06. The constraints list
C-59, C-82. The engineering log, planning entry D1. The research notes on
the UI, option (a).

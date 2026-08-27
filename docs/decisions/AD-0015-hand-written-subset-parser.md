# AD-0015 A hand-written strict subset parser

Status: accepted. Date: 2026-08-27.

## Context

The adapter reads a SysML v2 model from files rather than fronting a
repository (AD-0003), so something has to turn the textual notation into
elements the projection can serve. The README says the adapter's coverage of
the language is a fraction of the goal and leaves open whether coverage grows
as curated projections or as a generated whole (AD-0005). Underneath that sits
the question this record answers: what reads the text, and what happens to
text it does not understand.

The research sweep set out to confirm that no Go parser for the notation
existed, and the verdicts on the research refuted that. As of 2026-08-26 there
are three. Open-MBEE/OpenSysML is Apache-2.0, a hand-written recursive-descent
parser with semantic resolution, an LSP, a REPL and a gRPC server, created
2026-08-01 and at v0.2.1 by 2026-08-25, with all its library code under
`internal/` (C-38). mycr0ft/gosysml is MIT and ANTLR-based, created
2026-08-25 with three commits and a grammar descended from a 2023 release
(C-39). dVoo/gosysml2 is GPL-3.0 and describes itself as "purely vibe-coded
without further checks" (C-40). Beside them sit the antlr/grammars-v4 sysml-v2
grammar, MIT, which would generate a very large parser from a KEBNF that two
downstream projects document as ambiguous (C-41), and
nomograph-ai/tree-sitter-sysml, MIT, whose Go binding is cgo and would end the
`CGO_ENABLED=0` cross-compilation the image depends on (C-42). Syside is
closed source and licence-gated, with air-gapped use needing a Business
plan (C-44).

Two more routes read the model without parsing it in Go. The OMG pilot emits
the standard JSON serialisation through a Java 17 tool, the same shape a
conforming repository returns over the API, at the cost of a JVM at build time
and a flat element graph in which even memberships are elements (C-43).
OpenSysML's gRPC server could run as a sidecar and return its own `SymbolInfo`
tree, which is not the OMG JSON and whose API is pre-1.0 (C-38).

The repository's own rules bear on the choice. Product code is Go with no empty
interface in a value position (C-71), it is test-first (C-73), and it has to
be readable in an afternoon within a line budget per component (C-79, SC-06).
Edits through the projection patch a literal at its recorded source span (P10,
C-93), so whatever reads the text has to keep byte ranges. And the example has
to be accepted by the reference tools in any case, since a subset parser
cannot prove conformance (C-35).

## Decision

We will write the parser by hand in Go as a strict subset of the textual
notation, with a lexer, a recursive-descent parser and an AST in which every
node carries its byte range in the source. It accepts exactly the constructs
of the syntax card (C-23 to C-37, with ports and connections filled in by
C-34's spike) and refuses anything else with the file, line and column of the
first offending token. The parser has no opinion about meaning. Name
resolution, expression evaluation and the reading of `require constraint`
belong to the model package above it. The example model is validated
separately with the OMG pilot 2026-07 and the OpenSysML command line before it
becomes a fixture (P11, SR-45), and OpenSysML becomes the replacement when its
maintainers expose a public Go package or freeze the gRPC contract (C-45).

## Alternatives considered

Consuming the OMG JSON serialisation produced by the pilot, checked into the
example or produced in a build stage. The engineering log names it as one of
D8's two losing alternatives. It needs Java 17 at build time, its output is a
flat element graph the adapter would have to walk in the very metamodel it
exists to hide, and `elementId` values are tool-assigned UUIDs whose stability
across runs is unverified (C-43). The research keeps it as the best second
input path, because its shape is what a real repository returns.

An OpenSysML gRPC sidecar, D8's other losing alternative. It delegates
conformance to a maintained implementation, but it adds a second process to
the one-command container, its API is pre-1.0 and its output is its own
`SymbolInfo` model rather than the OMG JSON (C-38, the research notes on
SysML).

Importing an existing Go parser. None is importable under a compatible licence
today: OpenSysML's code is under `internal/`, gosysml was a day old with a
grammar from 2023, gosysml2 is GPL-3.0, tree-sitter needs cgo, and a parser
generated from grammars-v4 would be huge and inherit the grammar's documented
ambiguities (C-38 to C-42). Syside is excluded by its licensing (C-44).

## Consequences

The adapter stays pure Go with no dependency for parsing, cross-compiles with
`CGO_ENABLED=0`, fits the empty-interface and test-first rules, and owns a
small AST so the projection is plainly typed. Source spans make
`Patch(span, newLiteral)` possible, so the served text and the projection are
rebuilt together and never disagree (SR-22). A construct outside the subset is
refused, never skipped, which is what SR-18 requires, and it is the smaller
and more honest first cut of the escape hatch the README leaves open: nothing
is served generically in this phase (AD-0005).

The costs are the research's and are repeated here. The parser is not
conformant and never will be. Expressions need a real grammar even for the
subset, which is why the adapter evaluates only a literal with an optional
unit, feature reference chains, the four arithmetic operators and parentheses
(SR-23, P14). Name resolution, imports and redefinition are reimplemented in
miniature. Every new language feature is adapter work, and SC-06's budget of
2000 hand-written lines under `adapter/` bounds how far that goes before a
decision record has to raise it.

A silent mis-parse is the main risk and strictness is the mitigation. The
parser fails on an unknown token rather than skipping it, and one fixture per
rejection it distinguishes is the test (SR-18). Conformance of the example is
proven outside the adapter (SR-45). The validation run also settles the plain
numeric binding (C-26) and the duration literal (C-36), the ports and
`connect` syntax is quoted from OMG training folders 09 and 10 before the
example is written (C-34), and the 2.1 Beta 2 change list is C-22's spike,
since the pilot implements 2.1 Beta 2 and the adapter claims 2.0 formal.

The decision carries its own replacement condition. OpenSysML's `SymbolInfo`
tree is already close to what a projection needs, and if its maintainers
expose a package or freeze the gRPC contract the hand-written parser is the
part of the adapter to retire first. Keying the internal model on declared
short names rather than tool-assigned UUIDs (AD-0018) is what the research
says keeps a later swap a source change rather than a schema change.

## Requirements affected
SR-16, SR-18, SR-22, SR-23, SR-45

## Sources
The design brief D8, P10, P11 and P14. The constraints list C-22, C-23 to C-37, C-38 to C-45, C-71, C-73, C-79 and C-93. The requirements list SR-16, SR-18, SR-22, SR-23, SR-45 and SC-06. The engineering log, design phase planning, D8 and the findings that changed the design. The research notes on SysML, the options. The sceptic verdicts on the research, the parser verdict. The architecture description V5 The adapter.

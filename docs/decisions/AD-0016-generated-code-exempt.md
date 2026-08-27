# AD-0016 Generated code exempt from the empty-interface rule

Status: accepted. Date: 2026-08-27.

## Context

The coding rules forbid the empty interface in any value position: not
`interface{}` and not `any`, in parameters, results, struct fields, variables,
map, slice, array, channel and pointer element types, type arguments, named
types, aliases or constraint type sets, with `any` permitted only as a
type-parameter constraint. The `nointerface` analyser, kept outside the
tracked tree, enforces the rule from `make vet` across linux, darwin and
windows on amd64 and arm64, a reviewed exception is `//nointerface:allow` on
the line, and `make any-baseline` fails if the count of exceptions grows
beyond `.nointerface-baseline` (C-71). The rule governs what this code
declares and not what a dependency accepts, which is why go-cmp is allowed
despite its own signatures (C-77).

The three subgraphs are federated GraphQL services, two of them with
subscriptions (AD-0001, AD-0014). gqlgen is the only credible Go library for
that combination, and its generated code is full of `interface{}` (the
engineering log, D9). Left as it stands, the rule would report every
generated file on every vet run.

The analyser already has a `-generated` flag, default true, described as
"also report generated files (the header is one line and proves nothing)".
When it is false the analyser skips files whose header comment starts
`// Code generated ` and ends ` DO NOT EDIT.`, which is the marker gqlgen
writes. The `vet` target does not pass the flag today (C-72). One hand-written
exception is foreseeable: gqlgen's `explicit_requires` populate function,
whose signature carries `reps map[string]interface{}` and which the capacity
service needs for `@requires` over nested fields (C-15, C-72).

The rule is enforced on the maintainer's machine only, since the analyser is
not tracked and neither a fresh clone nor CI runs it (C-71). What the coding
rules left open was whether generated files fall under it at all.

## Decision

We will exempt generated files from the empty-interface rule and keep every
hand-written file under it. The Makefile passes `-generated=false` on each
`go vet -vettool` line so that the analyser skips files carrying gqlgen's
`// Code generated ... DO NOT EDIT.` header, and a forced empty interface in a
hand-written signature carries `//nointerface:allow` on the line with a
baseline bump and a justification in the pull request. SC-02 states the rule
in that form.

## Alternatives considered

No code generation at all. The engineering log records it as D9's first
losing alternative. Without gqlgen the subgraphs would take the gRPC or
plugin route, and neither supports subscriptions (C-14), so the live push of
D2 would be lost.

Per-line allowances inside generated files, D9's second losing alternative,
which the engineering log calls unworkable. An allowance written into a file
that `go generate` rewrites does not survive the next regeneration.

## Consequences

The rule keeps its force where it was meant to apply. Every hand-written file
under `adapter/`, `examples/` and `cmd/` is analysed as before, the baseline
stays small, and the one foreseeable exception is named in advance, so its
appearance in a pull request is expected rather than negotiated. gqlgen's
output carries the standard header, so no file has to be listed by name and
regeneration does not disturb the configuration.

The exemption rests on a one-line header, and the flag's own description says
that header proves nothing. A hand-written file that copied it would escape
the rule, and only review would notice. That is a weaker guarantee than the
analyser gives elsewhere.

A second cost is the shape of gqlgen's federation output. With
`explicit_requires` the populate function's `map[string]interface{}` lands in
hand-written code, and the nested-list `@requires` that forces it is C-15's
spike, the first of the implementation phase. If the spike fails and the flat
`Part.wiring: String` fallback is taken, the exception may not be needed and
the baseline stays where it is.

The Makefile edit is listed with the other repository policy additions the
implementation phase makes (architecture V4). Because the analyser is not
tracked, the change is invisible to a fresh clone and to CI, which is the
known limit of the whole rule and not a new one.

## Requirements affected
SC-02

## Sources
The design brief D9 and D2. The constraints list C-14, C-15, C-71, C-72 and C-77. The requirements list SC-02. The engineering log, design phase planning, D9. The design-phase plan, D9 and its ownership table. The architecture description V4 Deployment, additions to repository policy.

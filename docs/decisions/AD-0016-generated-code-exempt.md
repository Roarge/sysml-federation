# AD-0016 Generated code exempt from the empty-interface rule

Status: accepted. Date: 2026-08-27.

## Context

The coding rules forbid the empty interface in any value position: not
`interface{}` and not `any`, in parameters, results, struct fields, variables,
map, slice, array, channel and pointer element types, type arguments, named
types, aliases or constraint type sets, with `any` permitted only as a
type-parameter constraint. The rule governs what this code
declares and not what a dependency accepts, which is why go-cmp is allowed
despite its own signatures.

The three subgraphs are federated GraphQL services, two of them with
subscriptions (AD-0001, AD-0014). gqlgen is the only credible Go library for
that combination, and its generated code is full of `interface{}`. Left as it
stands, the rule would flag every generated file.

Generated files carry a header comment that starts `// Code generated ` and
ends ` DO NOT EDIT.`, which is the marker gqlgen writes. One hand-written
exception is foreseeable: gqlgen's `explicit_requires` populate function,
whose signature carries an empty-interface map and which the capacity
service needs for `@requires` over nested fields.

What the coding rules left open was whether generated files fall under the
rule at all.

## Decision

We will exempt generated files from the empty-interface rule and keep every
hand-written file under it. A file carrying the
`// Code generated ... DO NOT EDIT.` header is not checked, and a forced
empty interface in a hand-written signature carries an explicit allowance on
the line with a justification in the pull request. SC-02 states the rule
in that form.

## Alternatives considered

No code generation at all. Without gqlgen the subgraphs would take the gRPC or
extension-module route, and neither supports subscriptions, so live push
would be lost.

Per-line allowances inside generated files. An allowance written into a file
that `go generate` rewrites does not survive the next regeneration.

## Consequences

The rule keeps its force where it was meant to apply. Every hand-written file
under `adapter/`, `examples/` and `cmd/` is checked as before, and the one
foreseeable exception is named in advance, so its
appearance in a pull request is expected rather than negotiated. gqlgen's
output carries the standard header, so no file has to be listed by name and
regeneration does not disturb the configuration.

The exemption rests on a one-line header, which proves nothing on its own. A
hand-written file that copied it would escape the rule, and only review would
notice. That is a weaker guarantee than the rule has anywhere else.

A second cost is the shape of gqlgen's federation output. With
`explicit_requires` the populate function's empty-interface map lands in
hand-written code, and the nested-list `@requires` that forces it is the first
spike of the implementation phase. If that spike fails and the flat
`Part.wiring: String` fallback is taken, the exception may not be needed at
all.

## Requirements affected
SC-02

## Sources
gqlgen's generated output, its `// Code generated ... DO NOT EDIT.` header and its `explicit_requires` populate function. Go's own convention for the generated-code header. [Five spikes before the first line](../articles/09-five-spikes-before-the-first-line.md) for the one generated file that does not carry the header, and [How the design was run](../articles/02-how-the-design-was-run.md) for the rule the exemption sits inside.

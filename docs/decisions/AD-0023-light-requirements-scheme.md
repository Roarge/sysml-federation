# AD-0023 The light requirements scheme

Status: accepted. Date: 2026-08-27.

## Context

Two systems of interest share the page. One is the demo, with its adapter,
its services, its image and its repository. The other is the pipeline the
example model describes, whose requirements are declared inside the `.sysml`
file with short names such as `PIPE-R1`. The research notes on requirements
practice observe that a reader meets both in the same document and that a
mix-up has to be visible from the shape of the identifier alone. The README
says nothing about how the repository's own requirements are written or
numbered.

A heavier systems engineering methodology, with its own identifier shapes and
file rules, was the alternative. The repository is public and meant to be read
in an afternoon (SC-06), and `docs/` is written for someone deciding whether
the approach fits their organisation, which argues against a page of
conventions before the first requirement.

The research recommended persona-first stories with Given/When/Then
criteria, EARS statements for behaviour with a plain shall form for
constraints, shape-distinct identifiers plus the fixed phrase "model
requirement", decision records with four-digit numbers and Nygard sections,
and hand-maintained one-hop Markdown traceability tables. Gate 2 added the
`SC-nn` prefix for design constraints, after the requirements document
had separated design constraints, which take a plain statement, from
behaviour, which takes EARS.

## Decision

We will number the repository's own artefacts with a light scheme: `US-nn`
for user stories, `SR-nn` for system requirements written in EARS form,
`SC-nn` for design constraints written as plain statements, `AD-nnnn` for
decision records, and `C-nn` for the entries on the technical constraints
card, while a requirement that lives inside the example model is shown by its
SysML short name, in code font, and is always called a model requirement.
Traceability is one hop per table, an identifier is never reused once
retired, and a withdrawn constraint keeps its number with a note saying why.

## Alternatives considered

The heavier methodology throughout, with its own identifier shapes, file
rules and tooling. It lost because a reader judging a SysML adapter is not
judging a way of working, and the methodology's conventions would show on
every page.

The heavier methodology behind the scenes with the light scheme in `docs/`.
It lost because two schemes drift, and the mapping between them
would be one more thing to maintain and one more place for a number to be
wrong.

Prefixing both the repository's requirements and the model's with `REQ-` and
relying on prose to keep them apart. The research rejected it because a
shared shape hides a mix-up where a distinct shape shows it.

A docs-as-code requirements tool, StrictDoc or Sphinx-needs, or the SysML
model as the trace source. The research deferred both. A toolchain does not
belong in a Go repository meant to be read without setup, and the demo's own
requirements are not in a SysML model, so the model cannot be their source.
The question returns only if the requirement count outgrows one screen.

## Consequences

The scheme needs no explanation beyond one paragraph, and the identifier's
shape tells the reader which system a requirement
belongs to. Every SR carries an EARS statement with one system name, a
rationale, a verification method, the stories
it traces to and its allocation, which is what the reading at gate 2 checked.
Splitting the statements that named two obligations took the count from
thirty-one to forty-five, above the plan's target of twenty to twenty-five,
and that was accepted for the reason the split gives.

Traceability is hand-maintained Markdown, generated at gate 2 by a script
that checks coverage both ways. The research suggested a Go unit test that
parses the documents and fails on a dangling identifier, which the
unit-tests-only CI rule (SC-07) allows, and the plan turns the script into
that test, reading `docs/`.

No implementation spike depends on this decision.

## Requirements affected

none

## Sources

Mavin's EARS patterns and ISO/IEC/IEEE 29148:2018 on traceability, which fix the requirement syntax and the shape of the tables. [What the research overturned](../articles/03-what-the-research-overturned.md) for the practice this scheme is drawn from, and [From use cases to requirements](../articles/05-from-use-cases-to-requirements.md) for the forty-five requirements and the traceability as published.

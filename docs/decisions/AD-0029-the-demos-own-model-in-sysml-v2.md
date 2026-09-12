# AD-0029 The demo's own model in SysML v2

Status: accepted. Date: 2026-09-12.

## Context

The light requirements scheme (AD-0023) deferred a model as the trace source
because the demo's own requirements were not in a model, and said the question
would return once the requirement count outgrew one screen. The count outgrew a
screen long ago.

Two decisions taken alongside it bear on this one. The Markdown architecture
description, published in the article series, and these records were made the
record of the architecture, with the A3 sheets as the overview for a newcomer
(AD-0021). Separately, `internal/` was opened for two helper packages so that
tracked tests could import them (AD-0022).

The demo has never been modelled. Its stories live as prose and as boards
published beside the articles, its use cases as a PDF, its requirements and
design constraints as tables in article 05, its decisions as the records in
this directory. Nothing joins those but a reader's attention. The traceability
self-check AD-0023 planned, a Go test reading `docs/` and failing on a dangling
identifier, was never built, so every table has been checked by being read at a
gate. The repository argues that a systems model belongs at the centre of an
organisation's engineering, and keeps its own in prose.

## Decision

We will model the demo itself in SysML v2 under `model/`, in a tailored subset
of a story-driven agile MBSE layout: stakeholders, concerns, the twelve
storyboard stories and seven added with the model, the forty-five requirements
restated as system stories with their statements kept, the seven design
constraints, the architecture and its interfaces, and a verification register
in which every Go test, every recorded demonstration, the two validators, the
make targets and the workflows are actions or cases of their own and, once the
check session lands, every live check is a case of its own. The published
boards correspond to named views. The identifiers of the light scheme are the
model's short names. A unit test in `internal/trace` fails when the model and
the repository disagree on an identifier, a test name, a check file, a
published image, a check's inventory or a compose service.

## Alternatives considered

Keep the Markdown description as the record and add nothing to it. That is what
AD-0021 settled and it carried the design phase. It loses here because nothing
checks the description against the code. Earlier rounds of corrections to the
published documents happened because a number, a file name or a count had
drifted, and each drift was found by reading. Reading is not a mechanism.

A docs-as-code requirements tool, StrictDoc or Sphinx-needs, which AD-0023 also
weighed. It loses for the reason that record gave. A toolchain does not belong
in a Go repository meant to be read without setup, and the visitor this
repository is written for would meet the toolchain before meeting the adapter.

A model of the stories, requirements and architecture with the verification
left out. It would have been smaller and quicker to write, and it would have
carried the traces that matter most to a reader judging the approach. It loses
because the trace would stop at the requirement. The repository's own claim,
that verification is a first-class relationship, would then hold for the
pipeline example and not for the demo that ships it.

## Consequences

The model is the record of structure and traces. The articles stay the
narrative and these records stay the rationale, so AD-0021 is amended rather
than replaced, and the A3 sheets keep their place as the overview.

Seven stories and three requirements were added with the model. The stories are
US-13 to US-19 and the requirements SR-46 to SR-48. Nothing that already
existed changed its identifier, because the scheme retires numbers and never
reuses them. Article 05 keeps its count of forty-five, which is the count at
gate 2 and reads as a statement about that gate.

The planning notes behind the published documents stop being a second record of
the same structure. Where a note and the model disagree, the model is right.

`internal/` holds a third package, and it is a check rather than a helper.
AD-0022 opened the tree for helper packages and is amended to say what else is
in it.

The model is put to both reference tools on every change that touches it
(AD-0030), which is a Java runtime on a runner for a repository that has no
other use for one.

## Requirements affected

SR-46, SR-47

## Sources

The light requirements scheme (AD-0023), whose identifiers the model reuses as
short names, the architecture record decision (AD-0021) for what the Markdown
description is, and the tracked helpers decision (AD-0022) for the tree the
check package joins. [From use cases to requirements](../articles/05-from-use-cases-to-requirements.md) for the forty-five requirements and the traceability as published.

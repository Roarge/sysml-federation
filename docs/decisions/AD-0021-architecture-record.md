# AD-0021 The Markdown architecture description as the record, A3 sheets as the overview

Status: accepted. Date: 2026-08-27.

## Context

The repository publishes its design documentation under `docs/`, written for
someone deciding whether the approach fits their organisation, while a local
working directory holds notes and drafts. C-67 states the rule that joins the
two: a decision worth publishing moves from the working notes to `docs/`,
after which the working copy stops being the record. The README argues for
federation as the missing integration layer for open MBSE and says nothing
about where the architecture is written down. The plan closes the design
phase with public A3 architecture overviews at gate 4, so two documents about
the same architecture will exist, a Markdown description with views and
decision records, and a set of A3 sheets. One of them has to be the record.

An A3 architecture overview (Borches 2010, the research notes on A3 sheets)
is a two-sided sheet that captures one system aspect so that a group can
reason about change after reading it for ten to fifteen minutes. The research
recommends four sheets on three levels for this repository, L0, L1, L2a and
L2b, and says to start with L0 and L2b. Its cookbook is firm about density:
at most five colours, 14 pt body at A3, "you can't put everything you know in
this A3, so do not try",
and a link to another sheet instead of cramming. The same report records
that every published case ran a review loop of two to four weeks with several
stakeholders, that a small content change can restructure a sheet, and that a
single-maintainer public repository has none of that machinery.

The report also names what the format is not for: a decision log, code-level
detail, normative or contractual description, and content that changes faster
than a sheet can be re-issued. Two of its open questions bear directly on
this repository. The L0 sheet overlaps the README, so one of the two must
become the summary of the other, and numbers on L1 and L2b drift from the
code unless they are generated.

Against that stands the architecture description, which carries five views
each with a text form, the draft subgraph schemas, the process tree, the
Dockerfile, the exact allowlist and Makefile additions the implementation
phase makes, and an index of these decision records. That is the material a
contributor or an evaluating organisation needs verbatim, and it changes at
every gate. P12 in the design brief is where the plan took its position, and
the design-phase plan's decision 12 states it in two sentences with Borches'
own scope as the ground.

## Decision

We will treat the Markdown architecture description, published in the article
series, together with these decision records, as the record of the
architecture, and we will treat the A3 sheets designed at gate 4 as the
overview for a newcomer, and each sheet points back to the record. Every view
keeps a text form in the description, every decision has an AD record with
context and consequences,
and a sheet carries neither a decision log nor normative detail. When
material moves to `docs/`, the published copy becomes the record and the
working copy stops being one (C-67).

## Alternatives considered

The two-sided A3 sheet as the record, with the cookbook's back side carrying
definitions, concerns, requirements, decisions and known issues. Borches'
scope excludes it. The sheet exists for shared understanding of one aspect,
and a decision log or a normative description on it would either break the
density rules or need re-issuing more often than a sheet allows. Pesselse
found the back side's benefit not always clear, and this repository has no
two to four week review loop to keep such a sheet honest.

A full arc42 template as the record. It maps onto ISO/IEC/IEEE 42010 and
covers constraints, quality scenarios, decisions and risks in full. Twelve
sections is heavier than a proof of concept readable in an afternoon
warrants, many of them would read as ceremony, and the template does not
compress to a wall sheet, so it would not have replaced the A3 sheets either.

C4 diagrams as the record. C4 gives a developer the software decomposition
with a review checklist and can serve as the physical view inside a sheet.
It is software-only and static structure, it is explicitly not a
documentation template, and it carries neither quantification nor the
reading-path discipline of an A3AO. The architecture canvas borrows its
notation checklist for the view artboards and nothing more.

## Consequences

The text is the source and every sheet is derived from it. A change to the
design lands in the architecture description or a decision record first and
reaches a sheet at its next re-issue, so a sheet may lag and its status block
has to say so. Two documents about one architecture still exist and the
research's drift warning stands. The L1 field-ownership table and the L2b
worked numbers are the candidates for generation from the running example
rather than typing, which gate 4 settles, and the L0 sheet condenses the
README rather than the README condensing the sheet.

The decision records carry the rationale, which is where the 42010 concepts
of decision and rationale land without a conformance claim. Each record costs
about a page and there are twenty-six of them at gate 3, more text than the
README's afternoon suggests. That cost is accepted because the alternative
was rationale spread across sheets or lost with the conversation that
produced it.

Publication is a rename plus a rewrite under C-67, with the log entry that
produced the material cited in the pull request, and the public copy says
what the design contains rather than what was left out. No implementation
spike depends on this decision.

## Requirements affected

none

## Sources

The design brief P12, the design-phase plan decision 12 and gate 4, the
research notes on A3 sheets, options (the recommended sheet set, two-sided
versus one-sided layout, when to prefer C4 or arc42) and open questions, the
constraints list C-67, the architecture description preamble and decision
index, the engineering log's planning entry.

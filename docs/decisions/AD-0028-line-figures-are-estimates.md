# AD-0028 The line figures are estimates and never a limit

Status: accepted. Date: 2026-08-27.

## Context

SC-06 asks for code small enough to read in an afternoon and expressed that as
a set of line figures, under a rule that a figure rises only by a decision
record. Building the adapter tested the rule, and the rule failed twice over
the same number. The estimates for the parser were written before the files
existed and came to about half of what the specified contents measured, so the
first package was over its share before any judgement about implementation had
been made. Raising the figure took a decision record and a correction to every
document that carried the number, and the raised figure was itself overtaken,
which took a second round of the same (AD-0027).

The bookkeeping was the smaller cost. A figure written as a limit pushes in the
wrong direction at exactly the wrong moment, because the cheapest lines to give
up are the doc comments the linter asks for, the positioned refusals that make
the adapter honest, and the guards that keep a refusal from being reached by
accident. Those are the parts a reader in an afternoon most needs, and they are
also where a correctness fix lands. A refusal that is missing is repaired by
writing more of it and never by writing less.

## Decision

We will treat the figures in SC-06 as estimates of the expected scale. They are
not a gate. Nothing is trimmed, refused or left unwritten to stay under one,
and where the work turns out larger the figure is revised to what the work is,
in the requirement itself, with no decision record required.

The totals are still measured and reported as each part of the system is
finished, because the scale is worth seeing. A total above its figure is a fact
about the work rather than a defect in it.

Correctness comes first, then clarity, then size. A reviewer may say that code
is longer than it needs to be, which is a judgement about the code. Nothing may
say that code must shrink because a number says so.

## Alternatives considered

Keeping the rule and revising each figure as it is overtaken. That is what
AD-0027 did. It costs a record and a sweep through every document carrying the
number every time an estimate turns out low, and it buys nothing that reporting
the count does not.

Dropping the figures. SC-06 would then say only that the code should be
readable in an afternoon, which is a claim nobody can check. The figures are
what make that claim falsifiable, and they earn their place on that alone.

## Consequences

SC-06 is reworded. AD-0027's raise from 2000 to 2750 remains the record of why
the original figures were wrong, and its three shares for `adapter/syntax`,
`adapter/model` and the pair of `adapter/projection` and `adapter/serve` become
estimates like the rest.

Nothing in the build is a size gate, and no test fails for size alone. The
counts for the two web apps are measured and reported rather than asserted
against a limit.

The check on the image is untouched. SR-06's 80 MB per platform is a property
of the artefact a user downloads rather than a proxy for readability, and it
stays a gate that stops the release.

A figure revised whenever it is exceeded records the scale rather than holding
it down, which is the honest thing for such a number to be and also the weaker
one. Nothing in the repository now stops the adapter growing. What stands
against that is the reading the figures invite, and a reader who finds the
count no longer matches an afternoon has the number in front of them to say so.

## Requirements affected

SC-06

## Sources

[Planning the build](../articles/08-planning-the-build.md) for the figures and for what kind of number they are, and [The demo as it shipped](../articles/10-the-demo-as-it-shipped.md) for the counts as the system ships.

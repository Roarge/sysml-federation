# AD-0027 The adapter's expected scale raised to 2750 lines

Status: accepted, superseded by AD-0028. Date: 2026-08-27.

## Context

SC-06 asks for code small enough to read in an afternoon and puts a number on
it. For hand-written Go under `adapter/`, tests and generated files excluded,
that number was 2000 lines, counted with `wc -l` so that blank lines and
comments count towards it (C-79). The figure was split three ways, 800 lines
for `adapter/syntax`, 650 for `adapter/model` and 550 for `adapter/projection`
and `adapter/serve` together, and every share rested on per-file estimates
written before any of those files existed. The syntax tree was put at about
140 lines and the structural parser at about 160.

Once every file's text had been written out in full, the size of each package
stopped being an estimate and became a count. `adapter/syntax` measures about
1261 lines, `adapter/model` about 932, and `adapter/projection` with
`adapter/serve` about 389. The adapter as specified is therefore about 2580
lines, close to a third again above the figure it had been given.

Two properties of the design account for the difference, and both were settled
before the estimates were written. Every refusal carries the file, the line and
the column of the first offending token (SR-18, AD-0015), which needs a span on
every node of the tree and a positioned failure in every production. And every
exported identifier carries a doc comment, which the linter enforces, across a
tree of some thirty node types and a projected model of about as many.

An intermediate figure of 2300 was reached the same way, from an estimate of
what remained rather than a count, and the measured text passed it too. The
figure decided below is counted.

## Decision

We will raise the expected scale of hand-written Go under `adapter/` to 2750
lines, split as 1300 for `adapter/syntax`, 1000 for `adapter/model` and 450 for
`adapter/projection` and `adapter/serve` together. Each share is the specified
text plus a margin of three to fifteen per cent for what implementation adds.
The count stays `wc -l` over non-test files. The other figures in SC-06 are
unchanged.

## Alternatives considered

Thinning the specified text. The parser's repeated body scaffold and the
comments on the tree are the only slack there is, and taking all of it saves
perhaps fifty lines in the syntax package. That reaches neither 2000 nor 2300
without cutting the doc comments the linter requires or the positions the
refusals depend on.

Counting non-blank, non-comment lines instead. It brings the number near 2000
by changing what the number means. The plainer count is the one a reader can
reproduce with a single command, and it is the count SC-06 has always named.

Cutting the size of the model package. Its size is the size of the projection,
so a smaller package is a smaller projection, and that is a change to the
syntax tree and to the model's own interface rather than to the amount of code
written against them.

## Consequences

SC-06 reads 2750 for the adapter with a pointer to this record, and the
published account of the scale gives 2750 and the reason for it.

This is the second revision the same figure needed, and what the second one
shows is not that the adapter is large but that the figure had never been
measured. AD-0028 draws that conclusion and replaces the mechanism used here.
Under it a figure the work overtakes is revised to what the work is, with no
record required to do it, so this is the last time raising one costs a
decision. What stands from this record is the arithmetic, the account of why
the first estimates were wrong, and the three shares, which AD-0028 keeps as
estimates alongside the rest.

## Requirements affected

SC-06

## Sources

The measured text of `adapter/syntax`, `adapter/model`, `adapter/projection` and `adapter/serve` as specified, counted with `wc -l`. [Planning the build](../articles/08-planning-the-build.md) for the first figures and this revision, and [The demo being built](../articles/10-the-demo-being-built.md) for the scale of the code as it ships.

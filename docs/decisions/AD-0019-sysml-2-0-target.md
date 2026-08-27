# AD-0019 SysML 2.0 formal as the target, validated with the pilot and OpenSysML

Status: accepted. Date: 2026-08-27.

## Context

The README rests part of its argument on adoption. Version 2 "is not a
profile", it sits on KerML, and "the OMG approved it for final adoption in
July 2025". SysML 2.0 Part 1 Language is OMG document formal/26-03-02,
KerML 1.0 is formal/26-03-01 and the Systems Modeling API and Services 1.0
is formal/26-03-04. The press release of 2025-07-21 announced adoption,
the OMG About page lists September 2025, and the documents carry March 2026
on the cover. Those are the vote, the listing and the publication
rather than a contradiction, and the public text picks one phrasing
with the press release as its citation and the document number beside it.

The reference tooling has already moved on. The OMG pilot release 2026-07
conforms to KerML 1.1 Beta 2
and SysML 2.1 Beta 2, and OpenSysML pins 2.1 Beta 1. The research
put the question plainly: target the formal 2.0 the README cites, or the
2.1 Beta 2 the tools implement. On the evidence of the training files it
judged the constructs the example uses unchanged between the two, and it
did not read the 2.1 change list.

The adapter is a hand-written strict subset parser (AD-0015). A subset
parser cannot prove conformance, and a SysML-literate reader will load the
example file in a real tool before believing anything the demo says about
it. The pilot is Java, Xtext and Maven under EPL-2.0 with a
GPL-2.0-or-later secondary licence and needs Java 17 or later. The
OpenSysML command line installs with `go install`, is Apache-2.0, reports
324 of 353 pilot test files in diagnostic agreement, and claims no
conformance. Syside is closed source with a licence server, and its
command line needs a paid plan.

Two details of the example are settled only by running a tool. Throughput
is `attribute throughput : Real` documented as queries per second, because
hertz is not the unit of an event rate, and the latency requirement uses an
ISQ duration whose literal form, `200[ms]` or whatever the pilot accepts,
is quoted from no official file. The plain numeric binding on a Real
attribute is grammatically a usage declaration plus a value part and was
not found in an official example either.

## Decision

We will target SysML 2.0 formal, document formal/26-03-02, in the adapter
and the example, and validate the example `.sysml` with the OMG pilot
release 2026-07 and with the OpenSysML command line before any parser test
uses it as a fixture. Both runs happen locally and
never in CI, and the example README records the releases used and the
result. SR-45 states the obligation: the file is accepted without error by
both tools.

## Alternatives considered

Target 2.1 Beta 2, the version the pilot implements. It would have matched
the pilot exactly. It lost because the
README cites the adopted standard and the design claims it, and the
research found nothing the example uses that differs between the two. The
gap is stated rather than hidden: acceptance by the pilot shows the file is
valid SysML as the reference tools read it, not 2.0 conformance in the
formal sense (SR-45).

Trust the adapter's own parser. It is ruled out in one sentence: a
subset parser cannot prove conformance. The parser accepts the constructs it
lists and refuses the rest, which says nothing about whether that list is
right.

Validate with Syside. Excluded from both the toolchain and the
validation step by its licence: the free editor has no command line, the
command line needs a paid plan, and air-gapped use needs the Business tier.

Run the validation in CI. It lost to SC-07, which keeps CI to the unit
tests and the tag publish, and to the Java runtime the pilot would need,
which the test workflow does not set up.

## Consequences

A maintainer who changes the example model runs two tools before the file
becomes a fixture, and the record of which releases accepted it goes in
the example README. The pilot is Java, so a Java runtime is needed for a
step the repository otherwise has no use for. OpenSysML is a Go binary and
costs nothing extra.

Three spikes hang off this record, and the validation run is where they
close. The first reads the 2.1 Beta 2 change list for requirement
usage, satisfy, derivation, verification, ports and connections before the
run. The second settles the `[ms]` literal for the latency requirement. The
plain numeric binding is confirmed by the same run. Until those are done
the claim that nothing the example uses changed between 2.0 and 2.1 Beta
rests on the training files alone.

The claim the public text can make is bounded. The adapter and the example
say 2.0 formal, the validation says the reference tools accept the file,
and the two are not the same statement. OpenSysML claims no conformance
and the pilot implements a later beta. The example README records the
releases used and the result, which is all SR-45 asks of it.

Nothing under `examples/` is copied from OMG. The example model is authored
here, the subset the parser accepts cites the OMG training files as its
sources, and an
OMG file placed under `testdata/` as a parser fixture carries its EPL-2.0
notice with it. The adoption date is written one way in the public
text, cited to the press release with the document number beside it, which
settles the README's July 2025 against the About page's September and the
cover's March 2026.

## Requirements affected

SR-45

## Sources

OMG documents formal/26-03-02 (SysML 2.0 Part 1), formal/26-03-01 (KerML 1.0) and formal/26-03-04 (Systems Modeling API and Services 1.0), the OMG press release of 21 July 2025, and the OMG training files. The pilot implementation's release notes and licence, and OpenSysML's README and licence. [Five spikes before the first line](../articles/09-five-spikes-before-the-first-line.md) for the validation run and what both tools accepted.

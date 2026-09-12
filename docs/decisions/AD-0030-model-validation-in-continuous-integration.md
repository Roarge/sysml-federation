# AD-0030 Model validation in continuous integration

Status: accepted. Date: 2026-09-12.

## Context

The SysML 2.0 target decision (AD-0019) settled how the example model is
validated. The OMG pilot implementation and the OpenSysML command line both
accept the file before any parser test uses it as a fixture, both runs happen
locally, and neither happens in CI. SC-07 keeps CI to the unit tests and the tag publish. Both of those
statements were made about the example model, the one the adapter parses and
serves, and both still hold for it. A broken example is caught in seconds by
the parser tests that read it.

The demo's own model (AD-0029) is a different file set with a different reason
to be checked. Nothing serves it. No fixture reads it, no service parses it,
and the unit test that holds it to the repository reads names out of the file
rather than judging its grammar. A register left half-written, a `satisfy`
pointing at a requirement that was renamed, an import that no longer resolves:
none of that shows until a validator is run by hand, and the model drifts
silently between one such run and the next.

## Decision

We will validate the demo's own model with both reference tools, the OMG pilot
implementation release 2026-07 and OpenSysML v0.6.0, on every pull request and
push to main that changes a file under `model/`, in a workflow of its own,
`model.yml`, through the same `make model-check` a maintainer runs locally. The
example model keeps its local-only validation under `make example-model-check`,
and its record stays in the example README.

## Alternatives considered

Local only, as for the example. It is the cheaper answer and it keeps one rule
for both models. It loses on the difference between the two file sets. The
example is read by the adapter's tests on every run, so a file the reference
tools would refuse is usually a file the tests refuse first. The demo's model
is read by nothing else, so a register that neither validator accepts could sit
on main for weeks and be found by a visitor rather than by the repository.

One of the two rather than both. Half the runner time and half the setup. It
loses because the two do not agree on every form, which the five spikes showed
on the example: OpenSysML accepted with no errors a probe the pilot refused at
both its `satisfy` lines. Acceptance by one of them is a weaker statement than
this record makes.

## Consequences

The runner needs a Java runtime and two downloads for a job that neither builds
nor tests the Go code. Both downloads are cached between runs, so the cost
falls on the first. The validation lives in a workflow of its own, so the test
workflow keeps its shape and its timing.

SC-07 is amended to name the model validation. The constraint as written keeps
CI to the unit tests and the tag publish, and CI now does a third thing that is
neither.

The job is gated by paths, so it reports on a pull request that touches
`model/` and stays silent on one that does not. That keeps it off the list of
required checks as it stands, because a required check that never reports
leaves a pull request pending for ever. Making it required would mean running
it on every pull request, which is the cost the path gate avoids.

## Requirements affected

SR-47, SC-07

## Sources

The SysML 2.0 target decision (AD-0019) for the two validators and the releases
they are pinned to, [Five spikes before the first line](../articles/09-five-spikes-before-the-first-line.md) for what the two disagreed about on the example, and the workflow file itself.

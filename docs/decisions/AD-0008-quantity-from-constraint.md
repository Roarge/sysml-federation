# AD-0008 Quantity, comparison and limit read from the constraint

Status: accepted. Date: 2026-08-27.

## Context

The README's projection promises that a requirement is an identifier, a
name, some text, a limit value and a set of things that satisfy it. It does
not say how an adapter that contains nothing specific to any one model finds
the limit among a requirement's attributes, or how a service that knows only
a subject and a number tells one requirement on that subject from another.
Until gate 2 neither question had an answer in any document.

The example carries a latency requirement with a verification case beside
the throughput requirement, and both have the pipeline as their
subject. The reading at gate 2 found the consequence: a verdict rule keyed on
the subject could not tell PIPE-R1 from PIPE-R2 and would have passed the
latency requirement at a capacity of 1200 against a limit of 200. A second
gap in the same place was that no document said how a generic adapter knows
which of a requirement's attributes is its limit. Both came back to planning
as design findings rather than implementation defects.

The language supplies the shape. A requirement definition or usage carries a
`require constraint` whose body is a comparison, verified from the OMG
training files as `require constraint { massActual <= massReqd }` with the
limit bound on the usage as `attribute :>> massReqd = 2000[kg]`, and
an attribute or subject may be bound to a feature chain such as
`vehicle.dryMass + vehicle.fuelMass`. It was already settled that the
model never carries the rollup arithmetic: an abstract part definition
declares `capacity` without a value, the pipeline declares `latency` without
a value, and every requirement is a constraint over one of them. The Systems
Library's `VerdictKind` has four literals, pass, fail, inconclusive and error,
which gives a requirement the service cannot evaluate a word of its
own.

## Decision

We will have the adapter read each requirement's `require constraint` as a
comparison between a feature chain rooted at the requirement's subject and
either an attribute of the requirement or a literal, and project the chain's
last segment as the constrained quantity, the operator as the comparison and
the other operand's evaluated value as the limit, with the unit as written in
the source beside it. A constraint of any other shape is refused at start
with its file, line and column, as for any construct outside the subset
(SR-18, SR-19). The capacity service evaluates only requirements whose
quantity is the name it is configured to compute, returns PASS or FAIL by the
projected comparison for those, and returns INCONCLUSIVE for every other
quantity before it looks at anything else (SR-30).

## Alternatives considered

A verdict keyed on the subject. The rule before gate 2 compared the
subject's capacity with the requirement's limit and nothing more. It lost on
the latency finding: two requirements with the same subject and different
quantities are indistinguishable to it, and the wrong one passes.

A naming convention on the requirement's attributes. The adapter could have
looked for an attribute with an agreed name, `requiredRate` in the example,
and taken its value as the limit. It was rejected because the
convention would have been an assumption taken from the example, and the
README's claim is that the adapter is specific to no example.

Examining child values before the quantity check. The first verdict
precedence put ERROR ahead of every INCONCLUSIVE case, so the latency
requirement would have reported a bad server value it never uses. The
other-quantity check moved to the front.

## Consequences

One rule evaluates all seven requirements of the example. The six throughput
requirements constrain `<subject>.capacity`, the service computes `capacity`,
and each gets PASS or FAIL by its own operator against its own limit, the
derived ones on single servers through the leaf rule with no second code
path. PIPE-R2 constrains `latency`, no service computes it, and it reports
INCONCLUSIVE with the reason that PIPE-VC1 is declared and no service runs
it. The comparison is the constraint's own, so a requirement written the
other way round is evaluated the other way round, and the limit's unit
reaches the document as written, which is how it shows "200 ms". The adapter
still contains no word of the example, and the capacity service's contract
with it stays at the entity key, the field set in its `@requires`, which now
carries quantity, comparison, limit, the verification case's short name and
the names of the subject and its children, and two configured names
(SR-31).

The rule reaches back into the model. For a leaf's constraint to read
`<server>.capacity`, every server had to declare `capacity` as well as
`throughput`, which is why the example shares an abstract part definition
between the pipeline and the servers, and the pipeline had to declare
`latency` for the latency constraint to resolve at all. A requirement whose
constraint has another shape, two feature chains, a compound expression, or
no operand rooted at the subject, refuses the whole model at start rather
than projecting without a quantity. That is the honest behaviour for a subset
and it narrows what "any conforming model" can mean until the shape is
widened.

Two spikes settle the literals the other operand may take: the plain numeric
binding on a `Real` attribute and the duration literal for the latency
limit, both confirmed by the validation run with the OMG pilot and reported
in [Five spikes before the first line](../articles/09-five-spikes-before-the-first-line.md).

## Requirements affected

SR-19, SR-30, SR-31

## Sources

The repository README, "Nobody outside systems engineering should need to know what SysML is". The SysML 2.0 language specification on the body of a requirement definition and usage, the OMG training material for `require constraint` and the bound limit, and the Systems Library's `VerdictKind`. [From use cases to requirements](../articles/05-from-use-cases-to-requirements.md) for the verdict rules and their precedence.

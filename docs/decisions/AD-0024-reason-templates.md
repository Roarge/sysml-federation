# AD-0024 Verdict reasons built from the capacity service templates

Status: accepted. Date: 2026-08-27.

## Context

Every verdict the capacity service returns carries a reason string, and both
apps show it, the viewer beside its red block and the document in each
requirement's row (SR-14, SR-37). The service is allowed exactly two model-specific
names, the quantity it computes and the attribute it reads, and it selects
from the adapter only fields of the generic projection declared in its
`@requires` (SR-31, P7). It never sees the words "server" or "pipeline", and
it never sees the derivation relationship between `PIPE-R1` and its five
derived requirements, because derivation is not in its field set
(the capacity model page). The README promises that a failing requirement is
marked as failing with the stage responsible named, so the reason has to name
the cut.

The wording had drifted before it was fixed. The story canvas at gate 1
carried the shipped failure reason four ways and was aligned on "parse is
the bottleneck at 1200" (the engineering log, gate 1 addendum). The first
templates on the capacity model page carried "capacity" and "server", which
the gate 2 second pass caught as words the service is not allowed to know,
and the derived requirements had been given a template chosen by the
requirement being derived, which the service cannot know (the engineering
log, gate 2 entry and addendum). The canvas board that quotes a reason was
then aligned with the template.

There was also the question of which service owns the sentence a document
reader sees. The document knows the derivation, its shipped prose paragraph
explains why an allocated limit on a server can fail while the pipeline as a
whole passes, and "allocated" is the brief's own word for the derived limits.
The gate 2 review settled the ownership as P17. The other half of P17, the
UI server's place in the image rather than in either app, is recorded in
AD-0011.

## Decision

We will build every verdict reason in the capacity service from one of the
fixed templates on the capacity model page, whose only variable parts are the
configured quantity and attribute names, numbers, the names of parts, the
kind of fault an ERROR reports and the short name of a verification case, so
that no template carries a word of the model. The template for a subject with children reads `<quantity>
<value> against <limit>, limited by <cut>`, the leaf template reads
`<attribute> <value> against <limit>` and is selected by the subject having
no children, each INCONCLUSIVE case and the ERROR case has a template of
its own, and a cut of several parts is listed in the order the router delivers the
children, comma separated, in wording that avoids a verb that would have to
agree in number. Words such as "allocated" belong to the document, which
knows the derivation and may add them beside the reason.

## Alternatives considered

Templates carrying the model's own words, "capacity" and "server", which the
first draft of the capacity page used and which read naturally for the
example. They lost at the gate 2 second pass because the service is
configured with two names and may know no other (SR-31), so the templates
now use the configured names and the word "part".

A separate template for derived requirements, chosen because the requirement
is derived. It lost because the service never sees derivation. The leaf
template is selected instead by the subject having no children, which yields
the same sentence for the example's five derived requirements with no second
code path, since a leaf's capacity is its own attribute value.

The reason naming the allocation, so that a derived requirement's row would
say its limit is allocated. It lost because the allocation rule is written in
the model as expressions the adapter evaluates, and the service sees only
the number that results. The document, which does know the derivation, is
where that word goes.

## Consequences

The reasons are testable in isolation. SR-30's verification is one test
case per verdict kind, per precedence rule and per reason template, and the
templates are a table a reader can check against the worked example, where
`PIPE-R1` fails with `capacity 1200 against 1500, limited by parse` and
`PIPE-R1.4` fails with `throughput 700 against 750`. Both apps show the
string unchanged and compute nothing, which is the demo's claim that the
verdict a reader sees comes from a service that has never parsed a model
file.

The cost is a plainer sentence than the example could have had. The reason
does not say bottleneck, server or allocated, and a document reader who
wants the word "allocated" finds it in the document's prose paragraph rather
than in the reason. The order of a cut of several parts follows the router's
delivery order of the children, which the service does not control, so the
index pair reads `indexA, indexB` today and the service makes no promise
about that order. For `PIPE-R2` the reason names `PIPE-VC1` as declared and
run by no service, and the row shows no current value (SR-37).

The INCONCLUSIVE reason for a requirement of another quantity depends on the
verification case's short name arriving through the nested `@requires`,
which is C-15's spike. The flat `Part.wiring` fallback keeps `verifiedBy {
shortName }` in the field set, so the template survives either outcome.

## Requirements affected

SR-30, SR-37

## Sources

The design brief P7, P15, P17 and the shipped document, the capacity model
page "Verdicts" (the template table, leaf selection and the "allocated"
paragraph) and "Allocation of derived limits", the requirements list SR-14,
SR-30, SR-31, SR-37, the engineering log's gate 1 addendum, gate 2 entry
(P17) and gate 2 addendum, the constraints list C-15, the architecture
description V2 capacity service schema and fallback.

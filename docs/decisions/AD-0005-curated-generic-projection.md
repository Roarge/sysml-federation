# AD-0005 A curated generic projection rather than generated metamodel types

Status: accepted, with the escape hatch left open. Date: 2026-08-27.

## Context

The SysML v2 API returns the metamodel. Elements are generically typed,
relationships are themselves elements, and a requirement's text is reached by
walking owned relationships to a documentation comment. The README calls that
the right design for a modelling API and the wrong thing to hand a consumer,
because every service downstream would have to learn KerML before it could
read a requirement. Its answer is a projection published by the model's
owners, in which a requirement is an identifier, a name, some text, a limit
and the things that satisfy it, and nothing else is visible.

The README also sets the adapter a goal that pulls the other way. Coverage of
the whole language, so that any conforming model can be served without the
adapter knowing what it models, points towards generating a few hundred types
mechanically from KerML, which would land the consumer back in the abstract
syntax the projection exists to spare them. The README states its current
view, that coverage should grow as curated domain projections one at a time
with a generic escape hatch for elements nobody has projected yet, and adds
that this is the part of the design most likely to change. That is the
question this record inherits.

Two facts bound the answer for this phase. The adapter reads the textual
notation through a hand-written strict subset parser (AD-0015) that owns a
small AST, which is what keeps the projection plainly typed, and the
repository forbids the empty interface in any value position of hand-written
code, so a type that holds whatever an element happens to carry has no
natural home. The only author-controlled stable identifier in the language is
the declared short name, which AD-0018 makes the entity key.

The rollup fixed the field set. To compute a capacity the analysis needs
parts with a numeric attribute, directed connections between sibling parts,
and a requirement with a subject, a constrained quantity, a comparison and a
limit, and the plan observes that this set is generic. The resulting
projection is parts, attributes, ports, connections, requirements with their
relationships and verification cases, the model's text and version, and with
it comes the whole of what the capacity service and the adapter agree on: the
entity key, the field set in the service's `@requires`, and two configured
names. The words "server" and "pipeline" stay in the example and the two apps.

## Decision

We will publish the model through a curated set of plainly typed GraphQL
types, `Model`, `Part`, `Attribute`, `Port`, `Connection`, `Requirement` and
`VerificationCase`, chosen by what the services downstream need to read
rather than generated from the metamodel, with no identifier from the example
anywhere in the adapter. The schema lives in `adapter/schema.graphql`, the
subset parser produces an AST with source spans, and the projection package
maps resolved elements onto those types and holds nothing the schema does not
show. A construct outside the subset is refused at start with its file, line
and column rather than served through a generic type, and the escape hatch
the README describes stays open as a question for a later phase.

## Alternatives considered

Generating the types from KerML. The README names this as where full
language coverage leads: a few hundred types produced mechanically, correct
and complete, and the same abstract syntax the API already returns. It lost
because the value of a projection is that it is small, and a consumer who
meets an owned relationship has been handed the ontology the argument says
nobody outside systems engineering should need.

Building the generic escape hatch now. The README offers it as the compromise
between curation and coverage, a fallback type for elements nobody has
projected. The architecture description leaves it unbuilt in this phase:
refusing a construct outside the subset is the smaller and more honest first
cut. It is deferred, which is why the status of this record is not simply
accepted.

One attribute name as the whole contract. The plan's first ownership table
had the adapter and the capacity service agree on nothing beyond the name of
the attribute the service reads. Working the rollup out during planning
replaced that with the three-part statement above, and the second gate
brought the ownership table in line. It lost because the field set
the service declares in its `@requires` is a structural dependency on the
projection, and leaving it unstated would have made the contract look
smaller than it is (SR-31).

## Consequences

A consumer of the graph meets a requirement with an id, a name, text, a
subject, a quantity, a comparison and a limit, and never a usage or an owned
membership. The adapter contains no identifier from the example (SR-17), a
second fixture model with other names and wiring is part of its tests
(SR-16), and every element is identified by its short name with the qualified
name as fallback (SR-21). The contract between the adapter and the capacity
service fits on the L1 sheet as its quantification block. Because every
projected field has a declared type, the empty-interface rule holds in the
adapter's hand-written code without an exception.

The cost is coverage. What the adapter serves today is a fraction of the
language, every new construct is adapter work, and a model that uses
anything outside the subset does not load at all (SR-18). That is the honest
behaviour for a subset parser and it is also a hard limit on the README's
goal of serving any conforming model. The escape hatch that would soften it
is undecided, and when it is designed it will have to avoid reintroducing the
metamodel through the back door, which is the tension the README already
names.

Two spikes touch this record. If Cosmo composition or gqlgen refuse the
nested-list `@requires` in the capacity schema, the fallback adds
`Part.wiring: String`, a JSON document of the children and connections behind
one scalar. That keeps the semantics and loses the plain typing for that one
field, so the spike decides whether the projection stays wholly typed. The
ports and connections syntax is quoted from the OMG training folders before
the model is written, which fills the last gap in the subset the
projection is drawn from. Both are settled in
[Five spikes before the first line](../articles/09-five-spikes-before-the-first-line.md).

## Requirements affected

SR-16, SR-17, SR-21

## Sources

The repository README, "What SysML v2 fixes, and what it leaves open", "Nobody outside systems engineering should need to know what SysML is" and "The adapter and the examples". The SysML v2 API and Services specification for what the metamodel returns. [Five views and twenty-six decisions](../articles/06-five-views-and-twenty-six-decisions.md) for the projected type set and the adapter view, and [The demo as it shipped](../articles/10-the-demo-being-built.md) for the packages that produce it.

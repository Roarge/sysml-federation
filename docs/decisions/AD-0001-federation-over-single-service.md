# AD-0001 Federation over a single GraphQL service

Status: accepted. Date: 2026-08-27.

## Context

The README opens with the integration problem MBSE was meant to solve and
has not: several descriptions of one system, kept consistent by hand, each
copy drifting from the moment it is exported. SysML v2 removes two of the
reasons for that. Its textual notation puts the model in Git, and its
standard API and JSON serialisation give the model an interface that is not
a vendor's. What the API returns is the metamodel, so every consumer has to
learn SysML before it can read a requirement. The README's answer is a
projection published by the model's owners, in which a requirement is an
identifier, a name, some text, a limit and the things that satisfy it, and
nothing else escapes the model.

Three conditions are set on that projection. It must be a live view rather
than an export, so it cannot drift. It must be joinable, so a service
holding data the model does not contain can attach that data to the model's
objects without either side knowing about the other. And the contract
between producer and consumer must be checked mechanically before
deployment. The brief restates the demo's purpose in the README's words,
that federation is the missing integration layer for open MBSE, and notes
that the demo shows the first two conditions while the repository shows
the third through a test that fails when a subgraph schema and the composed
configuration drift apart.

The audience fixes what an acceptable answer may cost. Organisations with
fewer than twenty-five engineers have nobody employed to do integration
full time, which is the assumption both the service bus generation and
OSLC carried and the reason those answers stayed the preserve of
organisations that could staff them. A central component every team
changes together fails the same test.

Three further decisions sit on top of that. Every read, write and
subscription from the two apps goes through the router. The adapter's
projection is generic, and the capacity service's whole agreement with it
is the entity key, the field set it declares in its `@requires` and two
configured names. Entity keys are SysML short names. One federation feature
the design leans on is unverified: whether Cosmo composition and gqlgen
accept `@requires` over nested lists of objects, which the first
implementation spike settles, with a flat `Part.wiring` scalar as the
fallback. Cosmo's gRPC and extension-module subgraph routes carry no
subscriptions, which is why all three services stay standard GraphQL
subgraphs. The README names Cosmo and says nothing of how three services
are packaged. The platform is AD-0002's subject and the packaging
AD-0011's.

## Decision

We will publish the SysML model, the capacity analysis and the document
structure as three independently owned federated subgraphs behind one
router, with no service importing, calling or reading the data of another,
and with every query, mutation and subscription from the two apps sent to
the router. The adapter declares `Part`, `Requirement` and
`VerificationCase` as entities keyed on `id`, the capacity service
contributes `capacity`, `bottleneck`, `verdict` and `verdictReason` to
those entities from the fields its `@requires` names, the document service
contributes `documentNumber` and `included`, and composition merges the
three schemas into the one the router serves.

## Alternatives considered

A single GraphQL service. One server, one schema, one team, with the model,
the analysis and the document structure all behind it. The README rejects
it because it is exactly the central integration component that every team
has to change together, which is what federation exists to avoid.

A central integration platform in the service bus tradition, with adapters,
a canonical data model and transformation logic owned by an integration
team. It lost on the audience. The canonical model has to be negotiated and
owned, and owning it is a full-time job the target organisations do not staff.

OSLC's linked data approach, identifying resources by URI and describing
them with shapes so that tools reference each other's objects without
importing them. The README credits it as an improvement and records that it
pushed its resource shapes outward onto consumers who never asked for them,
under the same assumption of a full-time integrator.

One very large model covering everything. The README names it as the
alternative the industry has tried repeatedly. It is hard to build, harder
to maintain, and it concentrates authority in whoever owns the schema,
where a federated arrangement scales by adding constituents.

The capacity service as a client of the router rather than a contributor to
the graph, which the plan's rollup evaluation considered. It would have the
analysis query the supergraph during its own resolution and keep a change
feed and a cache of its own. It lost to the `@requires` route, and the
detail of that choice is AD-0007's subject.

## Consequences

The merged schema is computed rather than authored, so nobody owns it, and
two services that define incompatible things fail composition in the
pipeline of whoever pushed the change, with the conflict named. The router
resolves a typed graph and holds no logic of its own. Each team models what
it understands in the notation suited to it, and the only thing anyone has
to agree on is keys. The demo makes this visible in the playground, where
one query returns a requirement's text from the adapter, its verdict from
the capacity service and its document number from the document service in
one response (SR-43), and the apps prove it by being pure clients of the
router (SR-40) that compute nothing.

The costs are real. Federation is a platform problem, so the choice commits
the project to a platform with composition, a planner and a registry, and
that is AD-0002. The agreement between adapter and capacity service is
larger than the word "keys" suggests. The field set in the service's
`@requires` is a structural dependency on the generic projection, and SR-31
states it as such so nobody mistakes the contract for smaller than it is.
The nested-list `@requires` the capacity service needs is the one
federation feature the design leans on that has not been exercised, and the
spike that settles it runs first in the implementation phase. Three services and a
router are four servers where a single service would be one, which is
what forces the single-image supervisor of AD-0011, and a router that is
absent takes both apps down with it, which SR-40's test relies on.

## Requirements affected
SR-40, SR-41, SR-43

## Sources
The repository README, "Nobody outside systems engineering should need to know what SysML is", "Federation, for the systems engineers" and "Why Cosmo, and not simply GraphQL". [Why federate a systems model](../articles/00-why-federate-a-systems-model.md) for the argument and [Five views and twenty-six decisions](../articles/06-five-views-and-twenty-six-decisions.md) for the composition view this record describes. The Cosmo documentation on entity keys, `@requires` and subscription support per subgraph transport, at router 0.343.1.

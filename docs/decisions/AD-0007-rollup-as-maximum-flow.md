# AD-0007 Rollup as maximum flow with the source-side minimum cut

Status: accepted. Date: 2026-08-27.

## Context

The README describes the rollup as a minimum over serial stages and a sum
over parallel ones. The example fixes the pipeline as a set of servers wired
in a combination of serial and parallel connections, with the throughput the
requirement checks a rollup over the servers computed from that wiring. An
earlier shape had modelled abstract stages with serial and parallel container
parts, in which the two rules apply by construction. With wiring instead of
containers, whether a given server is serial or parallel to another is a
property of the graph, and the rollup needs an algorithm that reads it.

The rollup is the only global function in the system. Its value depends on
every server and every connection, and any edit can change it, so where it
runs and what it needs decides more than the arithmetic. Its inputs fix the
minimum projection (AD-0005). If it runs on read from fields the router
carries through `@requires`, the service that computes it holds no copy of
the model and cannot be stale, which is the README's "live view, no export"
in mechanism form, and both apps become pure federated clients that compute
nothing. That places one unverified federation feature on the critical path,
`@requires` over nested lists of objects.

The bottleneck has to be a defined set. Reading the use cases at gate 1 found
that with parse raised to 1600 and indexA to 900, parse and the index pair tie
at 1600, so the minimum cut is not unique and the highlighted bottleneck would
be implementation-defined. The story now raises parse to 1700 so the demo never
shows a tie, and gate 2 was left to define which cut is reported when several
exist. The capacity model page settled it as the source-side canonical cut,
the servers whose in-node is reachable from the super-source in the residual
network of a maximum flow and whose out-node is not, a set that is the same
for every maximum flow and so does not depend on the order in which
augmenting paths were found. The arithmetic around it is idealised and says
so (AD-0006), and each connection takes its direction from the order of the
ends of its `connect` statement (AD-0009).

## Decision

We will compute the capacity of a part with children as the maximum flow
through the children's wiring, with each child split into an in-node and an
out-node joined by an edge whose capacity is the child's configured
attribute, each connection an unlimited edge from the out-node of its first
end to the in-node of its second, a super-source joined to every child with
no incoming connection and a super-sink joined from every child with no
outgoing one, and we will report as bottleneck the source-side canonical
minimum cut. A part without children has the capacity of its own attribute.
The capacity service computes both on every read from the attributes,
children and connections the router carries to it in its `@requires`, using
Dinic's algorithm, and holds nothing between requests.

## Alternatives considered

Rollup in the adapter. The adapter would evaluate the capacity as a KerML
expression over the model. That turns a projection into an expression
evaluator and contradicts the README's claim that the verdict comes from a
service that has never parsed a model file, which is the claim the whole
argument rests on.

Rollup in the apps. Each web app would fetch the throughputs and wiring and
compute the capacity and bottleneck itself. Two copies of the logic, and the
sharing story is lost: the viewer's red block and the document's reason
would no longer be one value read through the router.

The capacity service as a client of the router. The service would query the
supergraph for the model during resolution and keep its own change feed and
cache. A subgraph calling the supergraph it is part of is more machinery
than the demo needs, and the plan keeps it as a note on the sheet about how
a heavier analysis tool would integrate.

Series-parallel reduction. Applying the README's two rules recursively,
replacing each chain or parallel group with one server, gives the right
answer whenever the wiring is series-parallel and has no answer otherwise.
Flow handles fan-in from different points, cycles and several entry or exit
servers unchanged and explains the bottleneck as well through the cut, so
reduction survives only as the differential test that checks flow against
the recursive minimum and sum.

## Consequences

The rollup is defined for any wiring the adapter can serve and agrees with
the README's arithmetic wherever that arithmetic applies, so flow is the
mechanism and minimum over serial with sum over parallel is the explanation a
reader can check by hand. The bottleneck as a cut explains both halves of the
demo: raising a server outside the cut changes nothing, and raising one
inside it raises the capacity by the same amount until another set of servers
becomes the cheapest cut. The service holds no copy of the model (SR-32),
needs no subscription, and is a pure function of its inputs, covered by table
tests over the wirings SR-28 lists and by the differential test. Dinic's
algorithm is a few dozen lines of Go.

The design leans on one thing nobody has run. Whether Cosmo composition and
gqlgen accept the nested-list `@requires` in the capacity schema is the first
spike of the implementation phase, and its fallback is
`Part.wiring: String`, the same children and connections as one JSON scalar.
gqlgen's `explicit_requires` also hands the populate function an empty-interface
map, the one foreseeable allowance in
hand-written code. The service recomputes on every query, and a cache
keyed on the model version is the scaling story if it ever matters.

The source-side rule is correct and can surprise. Had the story kept parse at
1600, the third edit would have left two minimum cuts of equal weight and the
service would have reported parse to a visitor who had just edited indexA.
The shipped values avoid every tie on the visitor's path, and a visitor who
types their own values can still make one. Entry and exit are inferred from
the wiring, so a feedback connection into an exit server makes it an ordinary
server and a wiring left without an exit is INCONCLUSIVE rather than a number.

## Requirements affected

SR-28, SR-29

## Sources

The repository README, "The pipeline example". [From use cases to requirements](../articles/05-from-use-cases-to-requirements.md), which publishes the capacity model in full: the flow network, why it equals the README's arithmetic, the source-side cut and the worked example. [Twelve use cases and one moving bottleneck](../articles/04-twelve-use-cases-and-one-moving-bottleneck.md) for the tie that made the cut definition necessary. The max-flow min-cut theorem and Dinic's algorithm.

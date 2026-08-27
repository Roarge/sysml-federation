# AD-0009 Connection direction from the order of the ends

Status: accepted. Date: 2026-08-27.

## Context

The capacity service computes the capacity of a part with children as a
maximum flow over the wiring between those children (P8), and a flow needs
directed edges. The wiring in the example model is written as `connect`
statements between the ports of sibling servers, which D6 and D12 put in
scope, and the adapter projects each statement as a `Connection` with a
`from` and a `to` (P7). What neither the plan's ownership table nor the
brief said before gate 2 was where the direction of that edge came from.

A SysML v2 `connect` has two ordered ends and no direction of its own. The
gate 2 review found that the rollup's direction had been assumed rather
than stated, and the finding went back to planning (the engineering log,
gate 2). The syntax of ports and connections is not yet on the constraints
card, because no research report fetched OMG training folders 09 and 10, and
C-34 opens a spike to quote them before the example is written. The decision
therefore had to hold for whatever exact form those folders show, so long as a
connect statement has a first end and a second.

Two signals are available in the source. Every server declares an `in` port
and an `out` port (the design-phase plan, decision 10), and every connect
names one port at each end in a fixed order. The capacity model builds its
network from the connections as edges from the out-node of the first end's
part to the in-node of the second end's part, so the direction the adapter
chooses is the direction the flow runs, and a reversed edge would change the
result with no error raised anywhere.

## Decision

We will take the direction of a connection from the order of the ends of
its `connect` statement, first end to second, and the adapter will refuse
to start on a `connect` whose first end is not an `out` port or whose
second end is not an `in` port (P16, SR-20). The end order is the source of
the direction and the port directions are the check, so the two must agree,
and a disagreement is reported with file, line and column in the manner of
SR-18.

## Alternatives considered

Leaving the direction implicit. The plan's ownership table described
`connections` with a `from` and a `to` taken from the model's connect
statements and said nothing about which end was which. The gate 2 review
found that no document stated it, and a rollup whose direction rests on an
unstated reading of the syntax is not something a second model could rely
on.

Projecting the connect as written and running the rollup regardless of the
ports. This is the same end-order rule without the check. A connect written
back to front, `connect b.in to a.out`, would become an edge in the wrong
direction, the flow would run through it, and the capacity would be wrong
with nothing to say so. The engineering log's summary of P16 is that the
check makes a wiring error a model error rather than a wrong capacity.

## Consequences

A model that wires its parts the wrong way round fails at adapter start
with a positioned error, the same treatment SR-18 gives any construct
outside the subset. The model owner sees the mistake in the file rather
than in a number. The capacity service is untouched by the rule: it
receives `from` and `to` and builds its edges from them, and the direction
is settled before the router carries anything (the capacity model page).

The rule ties the projection to ports with a declared direction. A port
declared `inout`, which the projection's `Direction` enum admits, cannot
stand at either end of a connect under this rule, so the example declares
`in` and `out` on every port that takes part in the wiring. That narrows
what the adapter accepts, in line with the strict subset of D8, and a model
that needs an undirected connection is refused rather than misread.

The decision depends on one spike. C-34 requires the syntax of port
definitions, port usages and connect statements to be quoted from OMG
training folders 09 and 10 before the example is written, and the exact
form the parser accepts is fixed then. The rule itself, first end to
second, does not change with the spelling. SR-20's test covers the example
wiring and a fixture with reversed ports, and the fixture is where the
refusal is exercised.

## Requirements affected
SR-20, SR-28

## Sources
The design brief P16 (with D6, D12, P7, P8), the constraints list C-34, the requirements list SR-18 and SR-20, the capacity model page (the rollup as a maximum flow), the engineering log's gate 2 entry, the design-phase plan's decision 10 and ownership table.

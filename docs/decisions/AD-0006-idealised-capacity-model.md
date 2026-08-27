# AD-0006 An idealised capacity model

Status: accepted. Date: 2026-08-27.

## Context

The README's "Placeholders" section names three shortcuts and gives the third to
the example. Its capacity model is idealised: parallel branches assume evenly
partitionable work and perfect load balancing, and no queueing effect is
represented anywhere. The README calls the result "arithmetic chosen to make
a point about federation rather than a performance model anyone should plan
capacity with", and leaves it there. It does not say where the assumptions
are written down, how a reader of a verdict learns of them, or what the
number means for a pipeline that does not meet them.

The demo's audience is a visitor with Docker and fifteen minutes who has never
seen SysML, and its one memorable moment is a bottleneck moving. That moment
depends on arithmetic a reader can check in their head: a chain is governed
by its smallest member, parallel branches add, and raising a server outside
the bottleneck changes nothing. The brief fixed the example as servers wired
serial and parallel with the requirement checking a rollup over that wiring,
lists no queueing model among the non-goals and carries "the capacity
arithmetic is idealised and says so" among the assumptions.

The mechanism is maximum flow with node capacities and the minimum cut as the
bottleneck set (AD-0007), and the assumptions, the limits of validity and an
uncertainty statement are written in the manner NASA-STD-7009B asks of any
model used near its limits. The standard's reporting clauses are the clearest
public authority for what a systems engineering reader expects from a
computed result: explicit caveats where assumptions are violated, an
uncertainty estimate or a clear statement that none is available, and a
remark on whether the result is credible enough for its use.

The capacity model page written at gate 2 is where the statement lives. Its
assumptions are that every query traverses exactly one path from an entry
server to an exit server and no server duplicates, drops or multiplies
queries, that work partitions evenly across parallel branches, that load
balancing is perfect, that there is no queueing and no latency coupling, that
load is stationary, that connections have unlimited capacity, that a server's
throughput does not depend on the mix of queries it receives, and that entry
and exit servers are read from the wiring. Its limits paragraph says the
number is exact for the idealised pipeline, reads as an upper bound for a
real pipeline that meets the conservation assumption, can be wrong in either
direction where conservation fails, cannot detect any departure because the
service sees only throughputs and connections, carries no quantitative
uncertainty estimate, and must not be used for capacity planning.

## Decision

We will keep the example's capacity arithmetic idealised, with the
assumptions listed above and no queueing, balancing or latency model, and we
will state those assumptions, the limits of validity and the absence of an
uncertainty estimate on a capacity model page that ships with the example and
is read beside the verdicts. The page is published with the rest of the design
documentation, and the shipped requirements document carries one unnumbered
prose paragraph above PIPE-R1 that says the rollup is idealised and explains
why an allocated limit on a server can fail while the pipeline as a whole
passes.

## Alternatives considered

A performance model. A model with queueing, imperfect balancing and latency
coupling would produce a figure closer to what a pipeline sustains, and the
README rejects it as the wrong kind of model for the argument. The demo
exists to show three services joining on a key, and a capacity figure that
needs a performance engineer to interpret would spend the visitor's fifteen
minutes on the wrong lesson. The brief lists no queueing model as a non-goal.

The caveat in the README alone. The README's sentence could have stood as the
whole statement. It lost because the reader who needs the caveat is looking
at a verdict in the requirements document or the viewer and not at the
README, and because the reporting shape the research recommends wants the
assumptions, the limits and the uncertainty statement written out rather than
implied by one adjective.

Assumptions written into the model. The research notes that SysML v2's
`assume constraint` gives a place to state the same assumptions inside the
model. The projection carries no field for an assumption, so nothing
downstream would see it, and the capacity page beside the results is where
the reader of a verdict finds it.

## Consequences

The arithmetic stays checkable by hand, which is what the story of the
migrating bottleneck needs, and the example values are chosen so the visitor
never meets a tie. The page gives the number a stated meaning: exact for the
idealised pipeline, an upper bound for a real one that meets the conservation
assumption, and unusable for planning. A systems engineering reader finds the
assumptions, the limits and the uncertainty statement in the places the
standard says to look, and the shipped document repeats the one that matters
most beside the requirements it applies to.

The cost is that the number is a demonstration figure and has to keep saying
so. The service cannot detect a violated assumption, since it sees only
throughputs and connections, so the warning lives in prose that a reader can
skip. A visitor who takes the verdict as an engineering result has been told
otherwise in the document and on the page, and nowhere else. The page is a
page about arithmetic, which is more than the arithmetic deserves on its own.

No spike belongs to this record. The spike that decides how the wiring
reaches the service, nested-list `@requires` under Cosmo and gqlgen, belongs
to AD-0007.

## Requirements affected

SR-28, SR-30

## Sources

The repository README, "The pipeline example" and "Placeholders". NASA-STD-7009B on reporting a computed result and its uncertainty. [From use cases to requirements](../articles/05-from-use-cases-to-requirements.md), which publishes the capacity model page in full, its assumptions and its limits of validity.

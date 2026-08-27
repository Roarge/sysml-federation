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
the bottleneck changes nothing. D12 fixed the example as servers wired serial
and parallel with the requirement checking a rollup over that wiring, and the
brief lists no queueing model among the non-goals and "the capacity
arithmetic is idealised and says so" among the assumptions.

C-94 restates the cheat as a constraint and names the mechanism, maximum flow
with node capacities and the minimum cut as the bottleneck set (AD-0007), and
asks for the assumptions, the limits of validity and an uncertainty statement
to be written in the manner NASA-STD-7009B asks of any model used near its
limits. The research behind that constraint found the standard's reporting
clauses the clearest public authority for what a systems engineering reader
expects from a computed result: explicit caveats where assumptions are
violated, an uncertainty estimate or a clear statement that none is
available, and a remark on whether the result is credible enough for its
use. It also recorded the cost, that a page for arithmetic may look
disproportionate and that hedging can spread to the rest of the documents.

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
page about arithmetic, which is more than the arithmetic deserves on its own,
and the research warns that hedging written for it can leak into documents
that do not need it.

No spike in the constraints list is attached to C-94. The spike that decides
how the wiring reaches the service, nested-list `@requires` under Cosmo and
gqlgen (C-15), belongs to AD-0007.

## Requirements affected

SR-28, SR-30

## Sources

README "The pipeline example" and "Placeholders". The design brief D12, the shipped document, assumptions and non-goals. The constraints list C-15, C-94. The capacity model page "Assumptions" and "Limits of validity and uncertainty". The research notes on requirements practice, the capacity model page option and the NASA-STD-7009B entry. The requirements list SR-28, SR-30.

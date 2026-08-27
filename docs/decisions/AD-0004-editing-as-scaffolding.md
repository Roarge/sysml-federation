# AD-0004 Editing through the projection as scaffolding

Status: accepted. Date: 2026-08-27.

## Context

The README's position is that a projection should be read-only. Systems
engineers publish it, everyone downstream reads it, and writes go to the
model through the SysML v2 API. Under "Placeholders" the README admits
that the demo does otherwise. Editing the model through the projection is
scaffolding, it contradicts that position, and a real deployment would
write through the API instead. The pipeline section leaves the interface
for making a change unsettled and says it does not matter to the argument.

The demo cannot do without edits. Its one memorable moment is a bottleneck
moving, which takes a change to a server's throughput and a document that
responds at once, and the viewer chosen at planning shows the numbers as
inputs. The brief fixes the editable set as exactly the five server
throughput values and the limit of PIPE-R1, and its assumptions state the
rule behind that. A literal in the source is editable and a bound
expression is not, the apps expose controls only for the editable set, and
edits land in the served model text and its version counter and never on
disk. The derived limits are bound by expressions in the model so that they
follow PIPE-R1, which is what makes them read-only. The non-goals rule out
persistence across restarts, authentication and multi-user editing.

The mechanism comes from three decisions. Every read and write goes
through the router, so a value edit is a mutation the router plans to the
adapter and no app talks to a subgraph directly. State is held in memory
with a reset, and edited literals are patched into the served text. The
parser records source spans, which is what makes the patch possible. Put
together, `setAttribute`, `setLimit` and `resetModel` patch the
literal's recorded span in the in-memory text and bump the counter, so
`Model.text` and the projection never disagree, nothing is written to disk
in the container, and a restart restores the shipped values. The open
question at the time was whether a mutation rewrites the `.sysml` text on
disk, which would need a lossless syntax tree with trivia, or mutates
in-memory state lost when the container exits, and in-memory state is the
answer. Reading the use cases at gate 1 added that invalid input has to be
survived, and SR-25's rule is that a submitted value must be a finite
non-negative number.

## Decision

We will accept value edits as three GraphQL mutations on the adapter,
`setAttribute`, `setLimit` and `resetModel`, sent to the router like every
other request, and apply an accepted edit by replacing the literal at
its recorded source span in the in-memory model text, re-parsing,
rebuilding the projection and incrementing the version in one step.
The adapter accepts any value that is a literal in the source and
refuses a value bound by an expression or one that is not a finite
non-negative number, with a GraphQL error naming the element and the
reason. The two apps show edit controls only for the editable set, and the
public text calls the whole arrangement scaffolding that a real deployment
replaces with writes through the SysML v2 API.

## Alternatives considered

A read-only projection with writes made through the SysML v2 API to a
repository, which the README names as what a real deployment does. It lost
because there is no repository behind the adapter (AD-0003), so the demo
would have nothing to write to and nothing would move.

Rewriting the model file on disk, or on a volume, so that edits persist.
It was raised and the brief rules it out. Persistence across
restarts is a non-goal, a restart is the cheapest reset a demo can have,
which SR-09 relies on, and the launch line uses `--rm`, so the container's
disk is discarded at exit in any case.

Apps writing to the adapter directly rather than through the router. The
plan rejected it in one sentence: no app talks to a subgraph
directly. It would have opened the side channel SR-40 exists to forbid
and made the router optional for writes while mandatory for reads.

## Consequences

The bottleneck moves, and both apps see it move within SR-39's two seconds
because an accepted edit is one span replacement, one version increment and
one `modelChanged` event. The viewer's text and the document's projection
can never disagree, and SR-22 tests that under concurrent reads. Reset is
one mutation document carrying both root fields, `resetModel` and
`resetDocument`, and the apps ask for both because neither service knows
about the other (SR-41, SR-44). A restart is a reset (SR-09). The editable
set is the same in both apps (SR-13, SR-38), a derived limit set through
the playground is refused (SR-24), and nonsense typed into a field is
refused with the previous value left in place (SR-25).

The cost is the contradiction the README already owns. The projection is
not read-only, and the public text has to say so wherever it describes the
edit path. The adapter carries mutation resolvers and span bookkeeping that
a real deployment would remove, and the version counter it increments is
the stand-in of AD-0003 rather than a commit. The playground can reach any
literal in the source, PIPE-R2's 200 ms included, so the apps rather than
the adapter are what keep the visitor inside the editable set. There is no
authentication, no multi-user editing and no second machine, and
every edit is lost when the container stops. No spike belongs to this
decision.

## Requirements affected
SR-09, SR-13, SR-22, SR-24, SR-25, SR-38, SR-44

## Sources
The repository README, "Placeholders" and "The pipeline example". [Twelve use cases and one moving bottleneck](../articles/04-twelve-use-cases-and-one-moving-bottleneck.md) for the editable set and the criteria on invalid input, and [The demo being built](../articles/10-the-demo-being-built.md) for the three mutations and the refusals as they ship. AD-0003 for why there is no repository to write to.

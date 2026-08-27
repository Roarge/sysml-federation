# AD-0025 The document owns its structure and nothing else

Status: accepted. Date: 2026-08-27.

## Context

The README describes the requirements document as "a live view over the
model, plus the editorial decisions about ordering, numbering and what to
include that the model does not contain". The editorial
scope is reorder, nest, include or exclude a requirement, plus
editorial headings and free prose paragraphs. Two constraints shape how
that scope is held. No service may import, call or read the data of another
(SR-41), so the document service cannot ask the adapter which
requirements exist. And the numbering must be the document's own, so it
cannot be derived from the model's short names.

The use cases exercise every operation: one moves and nests, another inserts
a heading as a parent, adds prose, excludes and restores. A third fixes the
shipped structure, PIPE-R1 with its derived requirements nested as 1.1 to
1.5 and PIPE-R2 as section 2, with one unnumbered prose paragraph above
PIPE-R1.

## Decision

We will make the document service the owner of an ordered tree of heading,
prose and requirement nodes, numbered in dotted decimal from the tree alone,
siblings counted from one in tree order and prose nodes unnumbered (SR-33).
The service holds requirement keys and structure and no model data. Its
shipped tree names the example's requirement ids in a configuration file,
which is the one place the example's identifiers enter a service, and for
an id it has never heard of its entity resolver answers `included: false`
and `documentNumber: null`. Every operation of SR-34 is a mutation on the
tree (SR-35), renumbering follows each, and none touches the model
(SR-36). The reset restores the shipped tree.

## Alternatives considered

Numbering derived from the model's derivation relationships, so that
derived requirements are numbered under their original automatically. This
would have tied the document's numbering to the model and defeated the use
case where PIPE-R2 is nested under PIPE-R1 by editorial choice alone.

The document service reading the list of requirements from the graph, so
that a requirement added to the model appears in the document without
configuration. This makes a subgraph a client of the supergraph during
resolution and breaks SR-41. The README's argument rests on the services
never meeting outside the router.

A flat ordered list of requirement ids without headings or prose, the
plan's first shape. It lost because the editorial scope kept headings and
prose, so the node kinds became three.

## Consequences

The document reads like a specification and every editorial change is one
mutation, which the stories can exercise one operation at a time. The
service is small, a tree with three node kinds and a numbering function,
and holds no copy of the model (SR-36 is a one-line test). The cost is the
stated limit: a requirement not in the shipped tree is not in the document,
which the example README says plainly and the architecture's field table
repeats. An adopter replaces the shipped
tree along with the model. Excluded requirements keep their former parent
in the service's own data, which is what makes restore (SR-35) possible
without reading anything from the model.

## Requirements affected

SR-27, SR-33, SR-34, SR-35, SR-36, SR-37, SR-44

## Sources

The repository README, "Requirements and relationships". [Twelve use cases and one moving bottleneck](../articles/04-twelve-use-cases-and-one-moving-bottleneck.md) for the operations and the shipped structure, [Five views and twenty-six decisions](../articles/06-five-views-and-twenty-six-decisions.md) for the document service's schema and its answer for an unknown id, and [The demo being built](../articles/10-the-demo-being-built.md) for the tree as it ships.

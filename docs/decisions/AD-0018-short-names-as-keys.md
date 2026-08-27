# AD-0018 Short names as entity keys with the qualified name as fallback

Status: accepted. Date: 2026-08-27.

## Context

Federation joins on the entity key. The README's sketch has the adapter
declare `type Requirement @key(fields: "id")` and the analysis service
declare the same key, and the router merges on it without either service
importing the other. Adding a twelfth model, the README goes on, needs no
renegotiation with the other eleven "because the only thing anyone has to
agree on is keys". What the key is for a SysML v2 model it does not say.
The question has to be answered once, because three services share the
answer: the adapter publishes the key, the capacity service resolves
entities by it, and the document service stores it as a foreign key in its
shipped tree.

The language offers one author-controlled identifier. Any definition or
usage may carry a short name in angle brackets before its name, as in
`<'PIPE-R1'>`, and the official files use both the quoted and the unquoted
form. At the API level an element's `elementId` is a UUID assigned
by the tool, the API 1.0 OpenAPI has `alias` arrays on Project and Commit
and no `humanId` field on elements, and short names are optional per
element. Whether the pilot's serialiser produces the same
`elementId` for the same text on two runs is not verifiable from public
documents.

The brief's example table gives every server, the pipeline, every
requirement and the verification case a short name in the PIPE family, and
states that "short names are the entity keys and are shown in both apps".
The adapter reads files rather than fronting a repository (AD-0003), so no
tool-assigned identifier exists at all until a conforming repository stands
behind it.

The plan settled the key on its own. The reading at gate 2 then found that no
requirement stood behind the story criterion that short names are the
identifiers seen in both apps, and SR-21 was added.

## Decision

We will identify every projected element by its declared SysML short name,
and by its qualified name where the element declares none, and use that
identifier as the `id` field and the `@key` of `Part`, `Requirement` and
`VerificationCase` in the adapter's schema. The capacity service and the
document service declare the same key and nothing else about identity.
Every published element in the example carries a short name, so the
fallback is exercised by a test fixture rather than by the demo. Renames
are out of scope.

## Alternatives considered

The API-level `elementId`. It is the identifier a conforming SysML v2
repository would return, and the research kept the pilot's JSON
serialisation in view as a possible second loader for that reason.
It lost because the value is assigned by the tool rather than by the
author, no tool stands behind the adapter today, and its stability across
runs of the serialiser could not be verified. A key nobody
writes down cannot be quoted in the document service's configuration or
typed into the playground.

The qualified name for every element. The research placed it beside the
short name as the other author-controlled and stable identifier, and the
design keeps it as the fallback. It lost first place because the brief
shows the short name in both apps and the playground query in the
architecture asks for `requirement(id: "PIPE-R1")`, so the key a reader
meets in the document, the viewer and the playground is the short one.

An alias or human id from the API. None exists at element level in API
1.0, so there was nothing to choose.

## Consequences

The entity key is a string the model's author controls, which is the
property the README's argument needs. A document service that has never
parsed a model file ships a tree that names PIPE-R1, its five derived
requirements and PIPE-R2 in a configuration file, and the router joins the
rest. The capacity service's whole agreement with the adapter is this key,
the field set it declares in its `@requires`, and its two configured names.

Every element the example publishes must carry a short name. Leaving one
off would not break the adapter, which falls back to the qualified name,
but it would change the id the other two services see. SR-21 is verified
by a test over the example's ids and a fixture element without a short
name.

Renaming a short name is a breaking change for every service that stores
it, and the demo does not handle it. The document service's shipped tree
would name an id the adapter no longer serves. The brief lists renames as
out of scope.

The research adds that a projection keyed on a value the model text
carries, rather than on anything a tool assigns, makes a later swap to a
conforming repository a source change rather than a schema change. No spike
belongs to this record.

## Requirements affected

SR-21

## Sources

The repository README, "Federation, for the systems engineers". The SysML 2.0 language specification on short names, and the Systems Modeling API and Services 1.0 OpenAPI for `elementId` and the absence of an element-level alias. [Twelve use cases and one moving bottleneck](../articles/04-twelve-use-cases-and-one-moving-bottleneck.md) for the example's short names, and [Five views and twenty-six decisions](../articles/06-five-views-and-twenty-six-decisions.md) for the composition view and the merged graph.

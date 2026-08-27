# Decision records

Twenty-six records in the Nygard form, each giving the context, the decision,
the alternatives considered, the consequences, the requirements affected and
its sources. All twenty-six are accepted. They were written during the design
phase, before any code, and copied here so that they can be read and
challenged. A record's status is `accepted`, or `accepted, with` whatever
qualification the decision carries. A decision that replaces an earlier one
gets its own number rather than overwriting it, and a number is never reused.

Inside the records, `SR-nn` is a system requirement and `SC-nn` a design
constraint, from the set of forty-five requirements and seven constraints
that [From use cases to requirements](../articles/05-from-use-cases-to-requirements.md)
describes, and a name beginning `PIPE-` is a short name inside the example
model. `C-nn` is an entry on the technical constraints card, a design-phase
working document that is not published, so those references are there to show
what a decision rested on rather than to be looked up. The four gates the
records refer to are the design phase's own, and
[How the design was run](../articles/02-how-the-design-was-run.md) says what
each produced.

## The shape of the system

- <span class="rec-id">AD-0001</span> [Federation over a single GraphQL service](AD-0001-federation-over-single-service.md)
- <span class="rec-id">AD-0002</span> [Cosmo as the federation platform](AD-0002-cosmo-as-platform.md)
- <span class="rec-id">AD-0010</span> [The router as a child process from the copied binary](AD-0010-router-as-child-process.md)
- <span class="rec-id">AD-0011</span> [One binary, one process tree, one port, one UI server](AD-0011-one-binary-one-port.md)
- <span class="rec-id">AD-0012</span> [Composition as a maintainer step with committed output and a drift test](AD-0012-composition-committed.md)
- <span class="rec-id">AD-0013</span> [Telemetry disabled by environment baked into the image](AD-0013-telemetry-off.md)

## The analysis

- <span class="rec-id">AD-0006</span> [An idealised capacity model](AD-0006-idealised-capacity-model.md)
- <span class="rec-id">AD-0007</span> [Rollup as maximum flow with the source-side minimum cut](AD-0007-rollup-as-maximum-flow.md)
- <span class="rec-id">AD-0008</span> [Quantity, comparison and limit read from the constraint](AD-0008-quantity-from-constraint.md)
- <span class="rec-id">AD-0009</span> [Connection direction from the order of the ends](AD-0009-connection-direction.md)
- <span class="rec-id">AD-0024</span> [Verdict reasons built from the capacity service templates](AD-0024-reason-templates.md)

## The adapter

- <span class="rec-id">AD-0003</span> [The adapter reads files rather than fronting a repository](AD-0003-adapter-reads-files.md)
- <span class="rec-id">AD-0004</span> [Editing through the projection as scaffolding](AD-0004-editing-as-scaffolding.md)
- <span class="rec-id">AD-0005</span> [A curated generic projection rather than generated metamodel types](AD-0005-curated-generic-projection.md)
- <span class="rec-id">AD-0015</span> [A hand-written strict subset parser](AD-0015-hand-written-subset-parser.md)
- <span class="rec-id">AD-0018</span> [Short names as entity keys with the qualified name as fallback](AD-0018-short-names-as-keys.md)
- <span class="rec-id">AD-0019</span> [SysML 2.0 formal as the target, validated with the pilot and OpenSysML](AD-0019-sysml-2-0-target.md)

## The apps

- <span class="rec-id">AD-0014</span> [Subscriptions as version events with client refetch](AD-0014-version-events.md)
- <span class="rec-id">AD-0017</span> [Vanilla web apps with one vendored file](AD-0017-vanilla-web-apps.md)
- <span class="rec-id">AD-0025</span> [The document owns its structure and nothing else](AD-0025-document-owns-its-structure.md)
- <span class="rec-id">AD-0026</span> [The viewer shows the model's text beside a sketch of its wiring](AD-0026-viewer-shows-text-and-wiring.md)

## The repository

- <span class="rec-id">AD-0016</span> [Generated code exempt from the empty-interface rule](AD-0016-generated-code-exempt.md)
- <span class="rec-id">AD-0020</span> [The image published by a tag-triggered workflow to GHCR](AD-0020-publish-on-tags.md)
- <span class="rec-id">AD-0021</span> [The Markdown architecture description as the record, A3 sheets as the overview](AD-0021-architecture-record.md)
- <span class="rec-id">AD-0022</span> [Track internal/assert and internal/tabletest](AD-0022-track-internal-helpers.md)
- <span class="rec-id">AD-0023</span> [The light requirements scheme](AD-0023-light-requirements-scheme.md)

AD-0026 is due for amendment when the apps ship, because the editable numbers
moved to a panel above the text.

# Decision records

Twenty-six records in the Nygard form, each giving the context, the decision,
the alternatives considered, the consequences, the requirements affected and
its sources. All twenty-six are accepted. They were written during the design
phase, before any code, and copied here so that they can be read and
challenged. A record's status is `accepted`, or `accepted, with` whatever
qualification the decision carries. A decision that replaces an earlier one
gets its own number rather than overwriting it, and a number is never reused.

Inside the records, `SR-nn` is a system requirement and `SC-nn` a design
constraint, both from the requirements list of forty-five requirements and
seven constraints. `C-nn` is an entry on the technical constraints card. `D`
and `P` numbers are decisions in the design brief. `US-nn` is a use case on
the storyboard, and a name beginning `PIPE-` is a short name inside the
example model. The requirements list, the constraints card and the brief were
design-phase working documents. The articles summarise their content, and
article 05 explains the requirements and the capacity model in full.

## The shape of the system

- [AD-0001 Federation over a single GraphQL service](AD-0001-federation-over-single-service.md)
- [AD-0002 Cosmo as the federation platform](AD-0002-cosmo-as-platform.md)
- [AD-0010 The router as a child process from the copied binary](AD-0010-router-as-child-process.md)
- [AD-0011 One binary, one process tree, one port, one UI server](AD-0011-one-binary-one-port.md)
- [AD-0012 Composition as a maintainer step with committed output and a drift test](AD-0012-composition-committed.md)
- [AD-0013 Telemetry disabled by environment baked into the image](AD-0013-telemetry-off.md)

## The analysis

- [AD-0006 An idealised capacity model](AD-0006-idealised-capacity-model.md)
- [AD-0007 Rollup as maximum flow with the source-side minimum cut](AD-0007-rollup-as-maximum-flow.md)
- [AD-0008 Quantity, comparison and limit read from the constraint](AD-0008-quantity-from-constraint.md)
- [AD-0009 Connection direction from the order of the ends](AD-0009-connection-direction.md)
- [AD-0024 Verdict reasons built from the capacity service templates](AD-0024-reason-templates.md)

## The adapter

- [AD-0003 The adapter reads files rather than fronting a repository](AD-0003-adapter-reads-files.md)
- [AD-0004 Editing through the projection as scaffolding](AD-0004-editing-as-scaffolding.md)
- [AD-0005 A curated generic projection rather than generated metamodel types](AD-0005-curated-generic-projection.md)
- [AD-0015 A hand-written strict subset parser](AD-0015-hand-written-subset-parser.md)
- [AD-0018 Short names as entity keys with the qualified name as fallback](AD-0018-short-names-as-keys.md)
- [AD-0019 SysML 2.0 formal as the target, validated with the pilot and OpenSysML](AD-0019-sysml-2-0-target.md)

## The apps

- [AD-0014 Subscriptions as version events with client refetch](AD-0014-version-events.md)
- [AD-0017 Vanilla web apps with one vendored file](AD-0017-vanilla-web-apps.md)
- [AD-0025 The document owns its structure and nothing else](AD-0025-document-owns-its-structure.md)
- [AD-0026 The viewer shows the model's text beside a sketch of its wiring](AD-0026-viewer-shows-text-and-wiring.md)

## The repository

- [AD-0016 Generated code exempt from the empty-interface rule](AD-0016-generated-code-exempt.md)
- [AD-0020 The image published by a tag-triggered workflow to GHCR](AD-0020-publish-on-tags.md)
- [AD-0021 The Markdown architecture description as the record, A3 sheets as the overview](AD-0021-architecture-record.md)
- [AD-0022 Track internal/assert and internal/tabletest](AD-0022-track-internal-helpers.md)
- [AD-0023 The light requirements scheme](AD-0023-light-requirements-scheme.md)

AD-0026 is due for amendment when the apps ship, because the editable numbers
moved to a panel above the text.

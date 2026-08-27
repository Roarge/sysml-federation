# AD-0003 The adapter reads files rather than fronting a repository

Status: accepted. Date: 2026-08-27.

## Context

SysML v2 gives a model two ways out of its authoring tool. The textual
notation makes a model a set of text files that live in Git and diff in a
pull request. The standard API and JSON serialisation make any conforming
repository expose projects, commits and elements the same way. The README
values the second more for this project, since it is what gives the model
an interface that is not a vendor's, and then states under "What this is
not" that the adapter reads files rather than fronting a repository and is
not a SysML v2 API implementation. Under "Placeholders" it says there is
no model repository behind the adapter, so versioning is stood in for
rather than taken from a conforming repository's commits.

What a repository would cost is on record. The OMG pilot's API server is
EPL-2.0, Scala on Play, JDK 11 and PostgreSQL, and the research notes its
implementation has not changed since 2025-04. The launch shape is one image
and one `docker run` with nothing but Docker installed (D5), and the
brief's non-goals repeat that the demo is not a SysML v2 API
implementation. The pilot can also emit the standard JSON serialisation
from text files with a Java 17 tool, a flat element graph in which
memberships are themselves elements and every reference is a UUID whose
stability across runs is unverified (C-43). That is the shape a repository
returns over the API, and the research keeps it as the best second input
path.

The identifier question is settled independently of the source. The
declared short name is the only author-controlled stable identifier in the
language, the API-level `elementId` is a tool-assigned UUID, and the API
1.0 OpenAPI has no human id field, so every published element carries a
short name and the adapter falls back to the qualified name (C-32, P9). The
brief's P10 holds state in memory with a version counter as the versioning
stand-in, and C-92 gives that counter its mechanism. `Model.version`
carries it, every accepted mutation increments it, and
`Subscription.modelChanged` publishes it. A model that is only files also
needs a check that the files are valid SysML as the reference tools read
it, since a subset parser cannot prove conformance (C-35). Which parser
reads the files is D8 and AD-0015, and what the projection does with
elements nobody has projected is AD-0005.

## Decision

We will have the adapter read the model from SysML v2 textual notation
files on disk when it starts, hold the parsed model in memory, and stand in
for a repository's commits with a monotonically increasing counter served
as `Model.version` (P10, C-92). The adapter neither implements nor calls
the SysML v2 API. The example ships its model file inside the image, the
parser records the source span of every numeric literal so the served text
can be patched in place, and the example file is accepted by the OMG pilot
2026-07 and the OpenSysML command line before it becomes a fixture (SR-45).

## Alternatives considered

Fronting a conforming SysML v2 repository over the standard API, which the
README says a real deployment would do. It lost because the demo has no
repository to front, and standing one up means a JVM, a database and a
second runtime inside an image whose whole promise is one command. The
research's advice is to key the internal model on short names rather than
UUIDs and shape it after the API's identity fields, so that a later swap to
a repository is a source change rather than a schema change.

Consuming the standard JSON serialisation produced by the pilot at model
build time, with the JSON checked into the example or generated in a build
stage. The research ranks it the best second input path because the read
path would be reusable against a real repository. It is not the first cut,
because it needs Java 17 at build time, the adapter would have to walk the
metamodel it was meant to hide, and the `elementId` values are UUIDs whose
stability across regenerations could not be verified (C-43).

A parser run at runtime as a sidecar, whether OpenSysML's gRPC server, the
pilot's Java tooling or Syside Automator. That adds a second process and a
second language runtime to the one-command container, the OpenSysML API is
pre-1.0 and returns its own tree rather than the OMG JSON, and Syside is
excluded outright by its licence terms on air-gapped use (C-44).

## Consequences

The adapter's parser is pure Go with no dependency, the image is one
static binary beside the router, and the model is text in Git, which is
the authoring freedom the README promotes. A SysML-literate reader can
load the same file in a real tool, and SR-45 makes that a recorded check
with the releases used. Because the key is the short name, the document
service's foreign keys and the capacity service's entity lookups do not
depend on where the model came from.

What is given up is everything a repository provides. There are no commits,
no branches and no history, and the version counter returns to its shipped
value on every restart, which SR-09 turns into a requirement rather than an
accident. Edits made through the projection reach the served text and
never the disk, which is AD-0004. The public text has to say that
versioning is stood in for. The adapter's coverage of the language is
whatever its parser accepts, and a construct outside the subset is refused
at start with a file, line and column (SR-18) rather than served
generically.

No spike is named for the file path itself. The syntax the file may contain
rests on spikes that belong to the parser: ports and `connect` quoted from
the OMG training folders (C-34), the plain numeric binding and the duration
literal confirmed in the pilot (C-26, C-36), and the 2.1 Beta 2 change list
read before the validation run (C-22). The validation itself runs locally
and never in CI (C-35).

## Requirements affected
SR-22, SR-26, SR-45

## Sources
README "Placeholders" and "What this is not", the constraints list C-22, C-26, C-32, C-34 to C-36, C-43, C-44, C-92, the design brief D5, P9, P10 and non-goals, the requirements list SR-09, SR-18, SR-22, SR-26, SR-45, the research notes on SysML, options and the API services finding, the architecture description V1 and V5, the engineering log's 2026-08-26 planning entry (D8).

# AD-0034 Acceptance criteria as Gherkin scenarios, run by godog from go test

Status: accepted. Date: 2026-09-27.

## Context

Every system story nests its acceptance criteria, one for each outcome that can
be verified on its own, and the story library asks for each in given, when,
then form. The research behind AD-0023 recommended exactly that. Of the
ninety-seven criteria of the forty-nine system stories, eighty-seven were
written that way, as prose in a doc comment of the model. Nothing ran them as
written.

A criterion was verified by whatever its verification case named: a Go test named after the
requirement, a recorded run, a check, an inspection. The Go tests carry the
requirement's key in their names, and their bodies read as Go. A reader could
not see which lines of a test stood for which criterion, or tell a criterion
that a test covers from one that nothing runs.

Behaviour-driven development writes criteria as scenarios in Gherkin, the plain
text format Cucumber reads, and binds each step of a scenario to code. godog is
Cucumber's implementation for Go. Its README calls its own command-line tool
deprecated and recommends running it from `go test`.

AD-0023 and AD-0029 turned down requirement tools, because a toolchain does not
belong in a Go repository meant to be read without setup. SC-01 kept the
module's dependencies to gqlgen, go-cmp and go-yaml.

## Decision

We will give every acceptance criterion of every system story one Gherkin
scenario, and run the scenarios with godog from `go test`.

Each system story has one feature file, named after the story, such as
`sr04-four-paths-on-one-port.feature`. It sits in a `features` directory beside
the package whose tests run it. The feature carries the story's key as its one
tag, `@SR-04`. Each scenario carries the name of the criterion it gives steps
to, as the model writes it, such as `@fourPathsAnswer`. That makes forty-nine
files in fourteen directories.

Some criteria can't be checked by `go test`: a page drawn in a browser, a layer
size read back by the publishing workflow, a model put to the reference tools.
The scenario for such a criterion also carries the kind of evidence that
verifies it today, such as `@record`, `@workflow` or `@validator`. It is read
and not run, and the runner reports it as skipped with that reason. Of the
ninety-eight criteria, counting the one this record adds to SR-46, fifty-nine
run and thirty-nine are verified elsewhere.

Each package that hosts a scenario that runs has one test, `TestScenarios`. It
hands its steps to a small package, `internal/scenario`, which reads the feature
files with godog's own parser. Each criterion runs as a godog suite of its own,
inside a subtest named for the story and the criterion, such as
`TestScenarios/SR-04/fourPathsAnswer`. The steps are bound to that subtest, so
they call the helpers the Go tests already use.

The verification register names each scenario that runs as evidence of a new
kind, `scenario`, by the tag of its criterion and its feature file. A scenario
that is read and not run adds nothing to the register, since its verification
case already names the evidence that verifies it.

The agreement test of SR-46 gains a ninth check, which holds the feature files,
the stories and the register to each other. Every file must be named after the
one story it tags, and every scenario must name one criterion of that story.
Every criterion of a story that is done must have exactly one scenario. Every
kind of evidence a scenario names must be offered by its story's verification
case, and the scenarios that run must be the ones the register names.

The scenarios sit beside the Go tests. No test named after a requirement was
removed or renamed.

## Alternatives considered

Subtests named after the criteria, in Go alone. They would give each criterion
a name in the test output. They lose because nobody reads a specification out
of Go, and the point is a criterion a reader can check against its steps.

A Gherkin reader written here, as AD-0015 wrote the SysML parser. AD-0015 wrote
its own because nothing read the subset it needed. Cucumber's parser already
reads Gherkin, and a second parser kept in step with the format would be work
for nothing.

godog's command-line tool, or Cucumber's JavaScript runner. Either is a
toolchain beside `go test`, which AD-0023 and AD-0029 turned down, and the first
is deprecated.

One `features` directory at the root, with each package's runner picking its
scenarios by tag or by path. It would make the best single index. It loses
because the map from scenario to package would be a second register, and the
agreement test could not see it drift.

Registering every scenario, the ones that don't run among them. The register
names only evidence that exists, and a scenario that is read and not run
verifies nothing.

Scenarios that drive a browser. SC-01 keeps dependencies that drive a browser
out of the product, and the check session of AD-0031 already drives the apps in
one.

## Consequences

SC-01 is amended to admit godog for the scenarios. It brings seven modules:
godog, its Gherkin parser and its messages, under the MIT licence, pflag under a
BSD licence, and go-memdb, go-immutable-radix and golang-lru under the Mozilla
Public License 2.0. Only tests import them, so none reaches the image, which
builds `./cmd/sysml-federation` alone. A clone needs no more setup than before,
since `go test` downloads these modules as it downloads the rest.

SR-46 gains a ninth criterion. The ten criteria that were not in given, when,
then form, SR-22's two and SR-46's eight, were rewritten in it.

Writing the scenarios turned up one criterion with no evidence at all: that the
maintainer's compose step reproduces the committed configuration (SR-42). The
step was run, its output matched the committed file byte for byte, and the run
is recorded in the example's README. Five criteria gained evidence that runs in
`go test` where they had an inspection or a record alone. Two are the licence
files beside the vendored library (SR-08) and a capacity service with no store
(SR-32). The others are apps that fail without the router (SR-40), a served
schema with types from all three subgraphs (SR-43), and the reference tools'
releases in the example's README (SR-45).

Each criterion's wording now lives twice, in the model's doc and in its
scenario's steps, and the agreement test checks the tags, not the words. A
criterion reworded in one place has to be reworded in the other by hand.

`internal/` holds a fourth package, `internal/scenario`, and AD-0022 carries an
amendment saying so.

## Requirements affected

SR-22, SR-42, SR-46, SC-01, SC-04

## Sources

[godog's README at v0.16.0](https://github.com/cucumber/godog/blob/v0.16.0/README.md), on running godog from `go test` and on the deprecated command-line tool. The [Gherkin reference](https://cucumber.io/docs/gherkin/reference/) for the format. The model's README, under "Tests in the model", for how evidence is named in the register.

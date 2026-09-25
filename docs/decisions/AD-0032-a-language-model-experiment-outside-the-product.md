# AD-0032 A language model experiment kept outside the product

Status: accepted. Date: 2026-09-25.

## Context

The second article series asks how the trace links between a systems model and
the system built from it can be found when nothing in the built system carries
the systems model's keys. Two questions about language models were left open: whether
one may read a systems model at all, and whether the explanations it writes
for its answers are the reasons it had.

The demo's own systems model holds a set of links someone wrote down. Its verification
cases name 69 Go tests over 32 cases, and SR-46 keeps the two sides in
agreement. Each of those test names starts with its requirement's key, so the
links are easy to find as they stand, and a test of a resolver has to hide the
keys first.

A language model that could answer the question does not fit in the product. A
14B language model quantised to four bits is a file of about 9 GB that wants a graphics
card. The demo is one container meant to run on a visitor's own machine, and SC-01 allows the
product no dependency beyond the standard library, gqlgen and what the tests
need. A hosted language model would answer the first question by sending the
systems model out of the building.

## Decision

We will build the experiment as a Go module of its own under
`experiments/llm-resolution`, using the standard library only, and run it
against an Ollama server in the operator's own network. The operator runs it on
their checkout and hands back the one results file it writes. Results are
written under `results/` and never committed.

The experiment has a SysML v2 systems model of its own, in
`experiments/llm-resolution/model`, built on the demo's library. It holds
stakeholders and their concerns, six stakeholder stories, four use cases,
twenty-two system stories derived from the stakeholder stories and four design
constraints. A logical architecture records every allocation as a satisfy, and
a verification register names each Go test. The tests were written from that register
before the code that passes them. A new target, `make experiment-model-check`,
puts the experiment's systems model and the library it uses to both reference
tools, and the `model` workflow runs it.

## Alternatives considered

A resolver added to the demo's own systems model and built in the main module.
It is closer to the method the first series followed, but it adds requirements
to a demo that has no need of them, and the language model half would still
have to sit outside.

A hosted language model. It is the simplest to run, and it sends the systems
model to a third party, which is one of the two things the experiment exists to
avoid.

## Consequences

The repository has a second Go module. `go test ./...` at the root does not
reach it, so the test workflow does not run its tests, and `make
experiment-test` does. The trace test of SR-46 skips nested modules, so the
experiment carries its own agreement test, EXP-SR-16, between its systems model
and its tests.

Continuous integration validates the experiment's systems model beside the
demo's. The `model` workflow gains the experiment's path as a trigger and
`make experiment-model-check` as a step.

`.gitignore` opens `experiments/` and names the module file, the README and the
run script, and `experiments` joins the roots `make check-allowlist` searches.

Nothing in the product, the image or its requirements changes.

## Requirements affected

SR-46, SR-47, SC-01, SC-04

## Sources

[The experiment's README](https://github.com/Roarge/sysml-federation/blob/main/experiments/llm-resolution/README.md) for how to run it and what it measures, and [its systems model](https://github.com/Roarge/sysml-federation/tree/main/experiments/llm-resolution/model) for its requirements and their verification. The [Ollama chat API](https://docs.ollama.com/api/chat) for the request it makes. The language model's [model card](https://huggingface.co/Qwen/Qwen2.5-Coder-14B-Instruct) and [Ollama's listing of its quantised builds](https://ollama.com/library/qwen2.5-coder/tags) for what the operator's server holds.

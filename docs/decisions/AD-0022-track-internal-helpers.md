# AD-0022 Track internal/assert and internal/tabletest

Status: accepted, confirmed by the owner with the gate 3 approval. Date: 2026-08-27.

## Context

The tracking policy keeps the support trees local. The working notes,
`internal/`, the analyser, the install scripts and the local tooling
configuration are not tracked, on the ground that the repository is about the
SysML adapter and not about how it gets written (C-67). The same policy
directs tests to use `internal/assert` rather than a library that takes `any`
throughout, and permits go-cmp because the empty-interface rule governs what
this code declares and not what a dependency accepts (C-77). Two
packages exist on disk under `internal/`. `assert` holds assertion helpers
over type parameters and go-cmp, so a comparison between mismatched types is
a compile error at the call site, and `tabletest` is a table-driven test
runner built on `assert`. Each carries its own tests.

C-70 states the conflict. No allowlist rule names `internal/`, `git ls-files
internal` is empty, and a tracked test that imports
`github.com/Roarge/sysml-federation/internal/assert` fails on a fresh clone
and in CI, which runs `go build ./...` and the unit tests on every pull
request (C-74, SC-07). The implementation phase is test-first (SC-03), so its first
tracked test meets the conflict on its first push.

Gate 2 recorded the question as one of two things not done and left it with
the owner: track the two packages and amend the policy, or keep tracked
tests free of the import. SC-04 says the choice is made at gate 3, and the
architecture's list of repository additions already carries them,
`!/internal/` with `!/internal/**/` in `.gitignore`, `internal` in the
Makefile's `ALLOWLIST_ROOTS`, and the policy's `internal/` line amended,
each listed under this record, which the owner confirms at gate 3.

The reading taken is that the policy's rule is aimed at tooling that exists
only to constrain how the code is written, and that a typed assertion helper
constrains nobody. It is a generic test helper of the kind any Go repository
might carry. The owner confirmed that reading with the gate 3 approval on
2026-08-27.

## Decision

We will, subject to the owner's confirmation, track `internal/assert` and
`internal/tabletest` as ordinary Go packages, and tracked tests will import
them. The mechanism is the three additions the architecture lists: the two
allowlist rules, the `ALLOWLIST_ROOTS` entry so that `check-allowlist`
reports a forgotten file under `internal/`, and the policy's table amended
so that `internal/` moves from the local-only column to the tracked one.
The two allowlist rules open the whole `internal/` tree to the allowlisted
extensions, and the two packages are all it holds today. The analyser and the
install scripts stay local.

## Alternatives considered

Keep tracked tests free of the import. Tests would compare with `==` and
go-cmp directly, or each package would carry its own small helpers. The
policy's table would stand as written. The cost is the same loop and the
same comparison repeated in every table-driven test, and two packages left
on disk with no consumer, which is the opposite of what C-77 asks for. The
gate 2 entry in the engineering log names this as the other choice and does
not argue for it.

## Consequences

Tests read as the policy intended, a case whose `Want` does not match the
function's result is a compile error rather than a run-time report, and a
fresh clone and CI run the same tests the maintainer runs. The helpers become
public code under Apache 2.0 and are read by the visitor who reads the
repository in an afternoon, so they count towards the impression the code
makes, though not towards SC-06's line budgets, which name no component
under `internal/`.

The policy's rule loses its clean shape. `internal/` was one line in the
local-only column and becomes a tracked tree with an exception noted in the
policy, and any future helper that exists only to constrain how the code is
written has to live outside the tracked tree instead. The `assert` package's
own test carries two `//nointerface:allow` lines on a stand-in for `testing.TB`,
whose variadic signature the standard library imposes, and those allowances
travel with it into the tracked tree. The empty-interface rule itself is
still enforced on the maintainer's machine only, since the analyser stays
local (C-71).

No implementation spike depends on this decision. Until the owner answers,
the implementation phase cannot write its first tracked test against the
helpers, so the answer is needed before gate 3 closes.

## Requirements affected

SC-03, SC-04

## Sources

The constraints list C-67, C-70, C-71, C-74, C-77, the requirements list
SC-03, SC-04, SC-06, SC-07, the engineering log's gate 2 entry "Two things
were not done", the architecture description V4 "Additions the implementation
phase makes to repository policy", `internal/assert/assert.go` and
`internal/tabletest/tabletest.go` as on disk.

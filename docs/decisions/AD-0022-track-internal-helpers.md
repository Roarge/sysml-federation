# AD-0022 Track internal/assert and internal/tabletest

Status: accepted. Date: 2026-08-27.

## Context

Two packages sit under `internal/`. `assert` holds assertion helpers over
type parameters and go-cmp, so a comparison between mismatched types is a
compile error at the call site rather than a report at run time, and
`tabletest` is a table-driven test runner built on `assert`. Each carries its
own tests.

The repository's tests are directed to use `internal/assert` rather than a
library that takes `any` throughout, because the empty-interface rule governs
what this code declares and not what a dependency accepts. Until this
decision no allowlist rule named `internal/`, so `git ls-files internal` was
empty and a tracked test importing
`github.com/Roarge/sysml-federation/internal/assert` would fail on a fresh
clone and in CI, which runs `go build ./...` and the unit tests on every pull
request. The implementation phase is test-first (SC-03), so its first tracked
test meets that conflict on its first push.

## Decision

We track `internal/assert` and `internal/tabletest` as ordinary Go packages,
and tracked tests import them. The mechanism is three additions: `!/internal/`
with `!/internal/**/` in `.gitignore`, `internal` in the Makefile's
`ALLOWLIST_ROOTS` so that a forgotten file under `internal/` is reported, and
the tracking policy amended so that `internal/` moves from the local-only
column to the tracked one. The two allowlist rules open the whole `internal/`
tree to the allowlisted extensions, and these two packages are all it holds.

## Alternatives considered

Keep tracked tests free of the import. Tests would compare with `==` and
go-cmp directly, or each package would carry its own small helpers, and the
tracking policy would stand as written. The cost is the same loop and the
same comparison repeated in every table-driven test, and two packages left
on disk with no consumer, which is the opposite of what the rule directing
tests to a typed helper asks for.

## Consequences

Tests read as the rule intended, a case whose `Want` does not match the
function's result is a compile error rather than a run-time report, and a
fresh clone and CI run the same tests the maintainer runs. The helpers become
public code under Apache 2.0 and are read by the visitor who reads the
repository in an afternoon, so they count towards the impression the code
makes, though not towards SC-06's line budgets, which name no component
under `internal/`.

The tracking policy loses its clean shape. `internal/` was one line in the
local-only column and becomes a tracked tree with an exception noted beside
it, and any future helper that exists only to constrain how the code is
written has to live outside the tracked tree instead. The `assert` package's
own test carries two explicit allowances on a stand-in for `testing.TB`, whose
variadic signature the standard library imposes, and those allowances travel
with it into the tracked tree.

No implementation spike depends on this decision.

## Requirements affected

SC-03, SC-04

## Sources

`internal/assert/assert.go` and `internal/tabletest/tabletest.go` as they stand in the repository, and the `.gitignore` and Makefile rules named above. [How the design was run](../articles/02-how-the-design-was-run.md) for the allowlist and the empty-interface rule, and [From use cases to requirements](../articles/05-from-use-cases-to-requirements.md) for the conflict that made the decision necessary.

# Build and quality gate.
#
# Every target here fails loudly. There is deliberately no "skip if the
# toolchain is missing" path: a gate that reports success because it never ran
# is worse than no gate, because it is trusted.

SHELL := /usr/bin/env bash
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := help

MODULE  := github.com/Roarge/sysml-federation
BIN     := $(CURDIR)/bin
COVER   := $(CURDIR)/coverage.out

# ---------------------------------------------------------------- toolchain --
# Resolve go once, explicitly. Hooks and IDEs run with a PATH that does not
# include an interactive shell's rc files, so "go" may not be resolvable even
# though it works when typed by hand.
GO ?= $(shell command -v go 2>/dev/null)
ifeq ($(strip $(GO)),)
GO := $(HOME)/.local/go/bin/go
endif
ifeq ($(wildcard $(GO)),)
$(error go toolchain not found at '$(GO)' -- run 'bash scripts/install-go.sh' or pass GO=/path/to/go)
endif

GOBIN := $(BIN)
export GOBIN
# Refuse a silent toolchain download; the pinned version is the one we test.
export GOTOOLCHAIN := local

GOLANGCI_VERSION := v2.13.1
GOTESTSUM_VERSION := v1.13.0

GOLANGCI   := $(BIN)/golangci-lint
GOTESTSUM  := $(BIN)/gotestsum
NOINTERFACE := $(BIN)/nointerface

# There is deliberately no cached package list. Every recipe passes ./... to the
# go tool directly, so a broken or absent go.mod makes the recipe itself fail
# instead of resolving to an empty list that each target then skips over.

# ------------------------------------------------------------- gate inputs --
# 'override' so neither the environment nor 'make VAR=' can empty these and
# turn a check into a no-op.
# Only the directories .gitignore actually allowlists. Support trees are
# deliberately untracked, so finding source in them is the intended state, not a
# forgotten allowlist entry.
override ALLOWLIST_ROOTS := adapter cmd examples docs
override TEST_FLAGS := -race -shuffle=on -count=1 -timeout=120s

# A floor, not a decoration. 'override' for the same reason as the rest: an
# empty value from the environment would make the comparison always true.
override MIN_COVER := 70

.PHONY: help
help: ## Show this help
	@grep -hE '^[a-zA-Z0-9_-]+:.*?## ' $(MAKEFILE_LIST) \
	  | awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

# ------------------------------------------------------------------ set-up --
.PHONY: bootstrap
bootstrap: ## One-time setup for a fresh clone or worktree
	@git config --local core.hooksPath .githooks
	@chmod +x .githooks/* scripts/*.sh 2>/dev/null || true
	@printf 'bootstrap: hooks armed via core.hooksPath=.githooks\n'
	@printf 'bootstrap: set your identity if this is a new clone:\n'
	@printf '  git config --local user.name  "<name>"\n'
	@printf '  git config --local user.email "<email>"\n'

.PHONY: install-tools
install-tools: ## Install pinned dev tools into ./bin
	@mkdir -p $(BIN)
	@# Pinned to the tag being installed, not to a moving HEAD ref.
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/$(GOLANGCI_VERSION)/install.sh \
	  | sh -s -- -b $(BIN) $(GOLANGCI_VERSION)
	$(GO) install gotest.tools/gotestsum@$(GOTESTSUM_VERSION)
	@$(MAKE) --no-print-directory $(NOINTERFACE)

$(NOINTERFACE): $(wildcard tools/nointerface/*.go) tools/go.mod
	@mkdir -p $(BIN)
	cd tools && GOBIN=$(BIN) $(GO) install ./nointerface

.PHONY: toolchain
toolchain: ## Prove the toolchain is installed and reachable
	@printf 'go:         %s\n' "$$($(GO) version)"
	@printf 'GOROOT:     %s\n' "$$($(GO) env GOROOT)"
	@printf 'GOBIN:      %s\n' "$(GOBIN)"
	@test -x $(GOLANGCI)    && printf 'golangci:   %s\n' "$$($(GOLANGCI) version 2>&1 | head -1)"   || { printf 'golangci-lint MISSING -- run make install-tools\n'; exit 1; }
	@test -x $(GOTESTSUM)   && printf 'gotestsum:  %s\n' "$$($(GOTESTSUM) --version 2>&1 | head -1)" || { printf 'gotestsum MISSING -- run make install-tools\n'; exit 1; }
	@test -x $(NOINTERFACE) && printf 'nointerface: present\n'                                       || { printf 'nointerface MISSING -- run make install-tools\n'; exit 1; }

.PHONY: doctor
doctor: toolchain ## Workstation checks: can git hooks find the toolchain?
	@env -i HOME=$(HOME) PATH=/usr/bin:/bin:$(HOME)/.local/bin bash -c 'command -v go >/dev/null' \
	  && printf 'non-interactive shells resolve go: yes\n' \
	  || { printf 'non-interactive shells resolve go: NO -- git hooks will not work\n'; exit 1; }

# ------------------------------------------------------------------ format --
.PHONY: fmt
fmt: ## Apply formatting
	$(GOLANGCI) fmt ./...
	cd tools && $(GOLANGCI) fmt ./...

.PHONY: fmt-check
fmt-check: ## Fail if formatting would change anything
	$(GOLANGCI) fmt --diff ./...
	cd tools && $(GOLANGCI) fmt --diff ./...

# -------------------------------------------------------------------- vet ---
# Present only in a working copy that carries the local support tree.
HAVE_TOOLS := $(wildcard tools/nointerface/analyzer.go)

.PHONY: vet
vet: ## go vet, plus the empty-interface ban where the analyzer is available
	$(GO) vet ./...
ifeq ($(strip $(HAVE_TOOLS)),)
	@printf 'note: the empty-interface analyzer is not in this working copy, so\n'
	@printf '      that rule was not checked here.\n'
else
	@# Present but unbuilt is a hard error, never a skip.
	@$(MAKE) --no-print-directory $(NOINTERFACE)
	$(GO) vet -vettool=$(NOINTERFACE) ./...
	GOOS=windows $(GO) vet -vettool=$(NOINTERFACE) ./...
	GOOS=darwin  $(GO) vet -vettool=$(NOINTERFACE) ./...
	GOARCH=arm64 $(GO) vet -vettool=$(NOINTERFACE) ./...
	GOOS=windows GOARCH=arm64 $(GO) vet -vettool=$(NOINTERFACE) ./...
	cd tools && $(GO) vet ./... && $(GO) vet -vettool=$(NOINTERFACE) ./...
endif

.PHONY: lint
lint: ## Run golangci-lint
	$(GOLANGCI) run ./...
	cd tools && $(GOLANGCI) run ./...

# ------------------------------------------------------------------- test ---
.PHONY: test
test: $(GOTESTSUM) ## Run the suite with the race detector
	$(GOTESTSUM) --format testname -- $(TEST_FLAGS) ./...
	cd tools && $(GOTESTSUM) --format testname -- $(TEST_FLAGS) ./...

$(GOTESTSUM):
	@printf 'gotestsum missing -- run: make install-tools\n'; exit 1

.PHONY: test-watch
test-watch: $(GOTESTSUM) ## Red-green-refactor loop
	$(GOTESTSUM) --watch --format testname -- $(TEST_FLAGS) ./...

.PHONY: cover
cover: ## Run the suite with coverage and enforce the floor
	$(GO) test $(TEST_FLAGS) -covermode=atomic -coverprofile=$(COVER) -coverpkg=./... ./...
	cd tools && $(GO) test $(TEST_FLAGS) -covermode=atomic -coverprofile=$(CURDIR)/coverage-tools.out -coverpkg=./... ./...
	@pct=$$(cd tools && $(GO) tool cover -func=$(CURDIR)/coverage-tools.out | awk '/^total:/{gsub("%","",$$3); print $$3}'); \
	 printf 'tools coverage: %s%% (floor %s%%)\n' "$$pct" "$(MIN_COVER)"; \
	 awk -v p="$$pct" -v m="$(MIN_COVER)" 'BEGIN{exit !(p+0 >= m+0)}' \
	   || { printf 'tools coverage %s%% is below the floor of %s%%\n' "$$pct" "$(MIN_COVER)"; exit 1; }
	@pct=$$($(GO) tool cover -func=$(COVER) | awk '/^total:/{gsub("%","",$$3); print $$3}'); \
	 printf 'coverage: %s%% (floor %s%%)\n' "$$pct" "$(MIN_COVER)"; \
	 awk -v p="$$pct" -v m="$(MIN_COVER)" 'BEGIN{exit !(p+0 >= m+0)}' \
	   || { printf 'coverage %s%% is below the floor of %s%%\n' "$$pct" "$(MIN_COVER)"; exit 1; }

# ---------------------------------------------------- empty-interface rule ---
.PHONY: any-audit
any-audit: ## List every suppression of the empty-interface rule
	@grep -rn --include='*.go' -E '//[[:space:]]*nointerface:allow' . 2>/dev/null \
	  | grep -v '/testdata/' || printf 'no suppressions\n'

.PHONY: any-baseline
any-baseline: ## Fail if suppressions have grown since the baseline
	@n=$$(grep -rn --include='*.go' -E '//[[:space:]]*nointerface:allow' . 2>/dev/null \
	      | grep -vc '/testdata/' || true); n=$${n:-0}; \
	 b=$$(cat .nointerface-baseline 2>/dev/null || echo 0); \
	 printf 'suppressions: %s (baseline %s)\n' "$$n" "$$b"; \
	 if [ "$$n" -gt "$$b" ]; then \
	   printf 'suppressions grew from %s to %s -- justify each and update .nointerface-baseline in this PR\n' "$$b" "$$n"; exit 1; \
	 fi

# -------------------------------------------------------------- allowlist ---
.PHONY: check-tracked
check-tracked: ## No tracked path may be excluded by .gitignore (catches git add -f)
	@forced="$$(git ls-files -z | git check-ignore -z --no-index --stdin; \
	   rc=$$?; [ $$rc -le 1 ] || { echo 'git check-ignore failed' >&2; exit 2; })"; \
	 if [ -n "$$forced" ]; then \
	   printf 'tracked paths are excluded by .gitignore (force-added):\n'; \
	   printf '%s' "$$forced" | tr '\0' '\n' | sed 's/^/    /'; exit 1; \
	 fi; \
	 printf 'check-tracked: ok\n'

.PHONY: check-allowlist
check-allowlist: ## Warn about source files on disk that .gitignore would not track
	@missing="$$(git ls-files --others --ignored --exclude-standard -z -- \
	   $(addsuffix /,$(ALLOWLIST_ROOTS)) 2>/dev/null \
	   | tr '\0' '\n' | grep -E '\.(go|sysml|kerml|graphql|graphqls|proto)$$' || true)"; \
	 if [ -n "$$missing" ]; then \
	   printf 'source files on disk that .gitignore does not track:\n'; \
	   printf '%s\n' "$$missing" | sed 's/^/    /'; \
	   printf 'add the directory to the "Source tree" section of .gitignore\n'; exit 1; \
	 fi; \
	 printf 'check-allowlist: ok\n'

.PHONY: modules
modules: ## List every Go module, so none escapes the gate
	@find . -name go.mod -not -path './bin/*' -not -path '*/node_modules/*' | sed 's/^/    /'

# ------------------------------------------------------------------ gates ---
.PHONY: check
check: toolchain fmt-check vet lint test any-baseline check-tracked ## The full gate
	@printf 'check: ok\n'

.PHONY: preflight
preflight: check check-allowlist cover ## Everything, as run before a push
	@printf 'preflight: ok\n'

.PHONY: clean
clean: ## Remove build output only (never touches untracked working files)
	@rm -rf $(BIN) $(COVER) $(CURDIR)/coverage-tools.out
	@$(GO) clean -testcache 2>/dev/null || true
	@printf 'clean: ok\n'

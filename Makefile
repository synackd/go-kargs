# Use of this source code is governed by the LICENSE file in this module's root
# directory.

# go-kargs is a source-only Go library, so there are no build/install targets.
# This Makefile only provides local development checks that mirror the CI
# workflows under .github/workflows/.

# Set paths to commands, allowing overrides from the environment.
GO            ?= $(shell command -v go 2>/dev/null)
GOLANGCI_LINT ?= $(shell command -v golangci-lint 2>/dev/null)
GOVULNCHECK   ?= $(shell command -v govulncheck 2>/dev/null)

# Go toolchain version – taken from go.mod (defaults to the version
# declared in the module). Allows the Makefile to force the exact
# toolchain used by all Go-related commands (go, golangci-lint, etc.).
GO_TOOLCHAIN_VERSION ?= $(shell awk '/^go / {print $$2; exit}' go.mod)
# Exported env-var that forces Go tools to use the selected toolchain.
GOTOOLCHAIN ?= go$(GO_TOOLCHAIN_VERSION)

# Coverage profile output.
COVERPROFILE ?= coverage.out

# Function to check that a command is available and error if it is not.
#
# Arg 1: Command path (can be a variable like $(GO) or a direct path)
# Arg 2: Command name for the error message
# Usage: $(call require-command,$(GO),go)
define require-command
@if [ -z "$(1)" ]; then \
	echo "make: *** $(2) command not found" >&2; \
	exit 1; \
fi
endef

.PHONY: help
help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m[VAR=val]... <target>\033[0m\n\nTargets:\n"} \
	/^[a-zA-Z0-9_\/.-]+:.*##/ { \
		printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2 \
	}' $(MAKEFILE_LIST)

.PHONY: test
test: ## Run unit tests
	$(call require-command,$(GO),go)
	GOTOOLCHAIN=$(GOTOOLCHAIN) $(GO) test -v ./...

.PHONY: race
race: ## Run unit tests with the race detector
	$(call require-command,$(GO),go)
	GOTOOLCHAIN=$(GOTOOLCHAIN) $(GO) test -race ./...

.PHONY: coverage
coverage: ## Run unit tests and generate a coverage profile
	$(call require-command,$(GO),go)
	GOTOOLCHAIN=$(GOTOOLCHAIN) $(GO) test -covermode=atomic -coverprofile=$(COVERPROFILE) ./...
	GOTOOLCHAIN=$(GOTOOLCHAIN) $(GO) tool cover -func=$(COVERPROFILE)

.PHONY: vet
vet: ## Run go vet
	$(call require-command,$(GO),go)
	GOTOOLCHAIN=$(GOTOOLCHAIN) $(GO) vet ./...

.PHONY: lint
lint: ## Run golangci-lint
	$(call require-command,$(GOLANGCI_LINT),golangci-lint)
	GOTOOLCHAIN=$(GOTOOLCHAIN) $(GOLANGCI_LINT) run

.PHONY: govulncheck
govulncheck: ## Run govulncheck
	$(call require-command,$(GOVULNCHECK),govulncheck)
	GOTOOLCHAIN=$(GOTOOLCHAIN) $(GOVULNCHECK) ./...

.PHONY: mod
mod: ## Download and prune Go modules
	$(call require-command,$(GO),go)
	GOTOOLCHAIN=$(GOTOOLCHAIN) $(GO) mod tidy

.PHONY: check
check: test race vet lint govulncheck ## Run all local checks

.PHONY: clean
clean: ## Remove generated artifacts
	rm -f $(COVERPROFILE)

# Makefile for managing Go project tasks such as running, building, testing, and maintaining dependencies.
# NOTE: comments sit on their own column-0 lines above each command so they are never
# echoed by make and never break backslash-continued shell blocks.

# Use bash (macOS and Linux) with strict flags so a failing command aborts the recipe.
SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c
# Show help when `make` is run without a target.
.DEFAULT_GOAL := help
# Delete a target file if its recipe fails midway.
.DELETE_ON_ERROR:
# Every target is a command, not a file.
.PHONY: help dev run build build-all test test-verbose bench example check tidy deps-upgrade deps-clean-cache clean tree tree-dirs

LOG_DIR  := logs
BIN_NAME := replify
BIN_DIR  := bin
# Extra arguments forwarded to go commands (e.g. make test ARGS="-run TestWrite").
ARGS ?=
# Example folder name for `make example` (e.g. NAME=01_quickstart).
NAME ?=

# Detect OS for binary extension
ifeq ($(OS),Windows_NT)
	BIN_EXT := .exe
else
	BIN_EXT :=
endif

BIN_OUT := $(BIN_DIR)/$(BIN_NAME)$(BIN_EXT)

# ==============================================================================
# Help / housekeeping
# ==============================================================================

help: ## Show this help message
# Print the heading.
	@echo "Available targets:"
# Extract "target: ## description" lines, sort them and print them in colour.
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

# ==============================================================================
# Development
# Prints CLI help output for quick reference during development.
# ==============================================================================
dev: ## Run ./main/main.go for quick development experiments
# Ensure the main directory exists.
	@mkdir -p ./main
# Run the development entry point.
	go run ./main/main.go $(ARGS)

# ==============================================================================
# Running the main application
# Executes the main.go file, useful for development and quick testing
run: ## Run the replify CLI (./cmd/replify)
# Run the CLI entry point.
	go run ./cmd/replify $(ARGS)

# Building the application
# Compiles the main.go file into an executable, for production deployment
build: ## Build the replify binary into bin/
# Ensure the output directory exists.
	@mkdir -p $(BIN_DIR)
# Compile the CLI into the output path.
	go build -o $(BIN_OUT) ./cmd/replify

# Compiles every package (the same check CI runs) without producing binaries
build-all: ## Compile every package, as CI does
# Build all packages to verify they compile.
	go build -v ./...

# ==============================================================================
# Module support and testing
# Runs tests across all packages in the project, showing code coverage
test: ## Run all tests without cache, with coverage
# Run every test with fresh results and a coverage summary.
	go test -count=1 -cover ./... $(ARGS)

# Runs tests verbosely, the same command CI runs
test-verbose: ## Run all tests verbosely, as CI does
# Run every test and print each test name and result.
	go test -v ./... $(ARGS)

# Runs the benchmarks that exist in pkg/ (codegen, hashy, slogger); skips unit tests
bench: ## Run benchmarks under ./pkg with allocation stats
# Run only benchmarks (the empty -run pattern matches no tests).
	go test -run '^$$' -bench . -benchmem ./pkg/... $(ARGS)

# Runs one sample program from examples/ (see examples/README.md for the list)
example: ## Run one example (Usage: make example NAME=01_quickstart)
# Require NAME so the user knows which example to run.
	@if [ -z "$(NAME)" ]; then echo "Usage: make example NAME=01_quickstart (see examples/README.md)"; exit 1; fi
# Fail clearly when the example folder does not exist.
	@if [ ! -d "examples/$(NAME)" ]; then echo "Unknown example: examples/$(NAME)"; exit 1; fi
# Run the selected example program.
	go run ./examples/$(NAME) $(ARGS)

# Aggregate gate: compile everything, then run the tests
check: build-all test ## Run every validation and test
# Reached only when every prerequisite passed.
	@echo "All checks passed."

# Cleaning and maintaining dependencies
# Cleans up the module by removing unused dependencies
# Copies all dependencies into the vendor directory, ensuring reproducibility
tidy: ## Tidy go.mod and refresh the vendor directory
# Remove unused and add missing module requirements.
	go mod tidy
# Copy dependencies into vendor/.
	go mod vendor

# Upgrading dependencies
# Updates all dependencies to their latest minor or patch versions
# Cleans up the module after upgrade
# Re-vendors dependencies after upgrade
deps-upgrade: ## Upgrade dependencies, then tidy and re-vendor
# Upgrade all dependencies (including test dependencies) to the latest minor/patch.
	go get -u -t -d -v ./...
# Remove unused requirements after the upgrade.
	go mod tidy
# Re-vendor the upgraded dependencies.
	go mod vendor

# Cleaning up the module cache
# Removes all items from the Go module cache
deps-clean-cache: ## Remove everything from the Go module cache
# Wipe the module cache (shared by all Go projects on this machine).
	go clean -modcache

# Generating project file tree
# Creates a text file representing the project's directory structure, excluding certain directories
tree: ## Write the project file tree to logs/codebase.txt (requires `tree`)
# Ensure the log directory exists.
	@mkdir -p $(LOG_DIR)
# The block below checks, in order:
#   1. whether the `tree` binary is installed;
#   2. if it is missing, prints an install hint and exits non-zero.
	if ! command -v tree &> /dev/null; then \
		echo "The 'tree' command is not installed. Please install it to use this feature."; \
		exit 1; \
	fi
# Write the tree (ignoring tooling and VCS folders) to the log file.
	tree -I ".gradle|.idea|build|logs|.vscode|.git|.github|vendor" > ./$(LOG_DIR)/codebase.txt
# Display the generated tree.
	cat ./$(LOG_DIR)/codebase.txt

# Prints the directory structure only (no files), useful for a quick codebase overview
tree-dirs: ## Write the directory-only tree to logs/codebase_dirs.txt (requires `tree`)
# Ensure the log directory exists.
	@mkdir -p $(LOG_DIR)
# The block below checks, in order:
#   1. whether the `tree` binary is installed;
#   2. if it is missing, prints an install hint and exits non-zero.
	if ! command -v tree &> /dev/null; then \
		echo "The 'tree' command is not installed. Please install it to use this feature."; \
		exit 1; \
	fi
# Write the directory-only tree to the log file.
	tree -d -I ".gradle|.idea|build|logs|.vscode|.git|.github|vendor" > ./$(LOG_DIR)/codebase_dirs.txt
# Display the generated tree.
	cat ./$(LOG_DIR)/codebase_dirs.txt

# ==============================================================================
# Clean
# Removes the build directory and log directory.
clean: ## Remove bin/ and logs/ (generated files only)
# Delete the build output directory.
	rm -rf $(BIN_DIR)
# Delete the log directory.
	rm -rf $(LOG_DIR)

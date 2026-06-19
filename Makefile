# Makefile for rad

# Directories
BIN_DIR := ./bin

# Commands
.PHONY: all format build test clean generate verify-generated

all: generate format build test

generate:
	@echo "⚙️ Running generators..."
	go run "./function-metadata/extract.go"
	mv "./functions.txt" "./rts/embedded/"
	go generate ./rts

# verify-generated runs every code generator and fails if any
# generated file CHANGED as a result. CI runs this on every PR to
# block merges of stale generated files - the only reliable guard
# against shipping a binary whose codegen is out of sync with its
# source (docs/funcs/*.md today, future sources too). Local
# developers usually don't need this; `make generate` is the
# regenerate-and-go path.
#
# Implementation: snapshot the generated paths into a temp dir,
# run generate, diff against the snapshot. We deliberately don't
# use `git diff` because that compares against HEAD - which would
# false-positive on a PR that legitimately updates generated
# files. We want to know "did running generate change anything?",
# not "is the working tree dirty?".
GENERATED_PATHS := rts/signatures_gen.go rts/embedded_funcs rts/embedded/functions.txt docs-web/docs/reference/functions.md docs-web/docs/reference/errors.md core/embedded_docs
# When adding a new generator, append its OUTPUT path here so
# verify-generated catches drift. Forgetting is the failure mode -
# the generator runs and writes, but CI can't see the drift.

verify-generated:
	@echo "⚙️ Verifying generated files are up to date..."
	@tmpdir=$$(mktemp -d) && \
		trap 'rm -rf "$$tmpdir"' EXIT && \
		for p in $(GENERATED_PATHS); do \
			if [ -e "$$p" ]; then \
				mkdir -p "$$tmpdir/$$(dirname "$$p")"; \
				cp -R "$$p" "$$tmpdir/$$p"; \
			fi; \
		done && \
		$(MAKE) generate >/dev/null && \
		stale=0 && \
		for p in $(GENERATED_PATHS); do \
			if ! diff -r -q "$$tmpdir/$$p" "$$p" >/dev/null 2>&1; then \
				echo "❌ Stale: $$p"; \
				diff -r -u "$$tmpdir/$$p" "$$p" | head -40 || true; \
				stale=1; \
			fi; \
		done && \
		if [ "$$stale" = "1" ]; then \
			echo ""; \
			echo "❌ Generated files are stale. Run 'make generate' and commit the result."; \
			exit 1; \
		fi
	@echo "✅ Generated files are up to date."

format:
	@echo "⚙️ Formatting files..."
	find . -name '*.go' -exec gofmt -w {} +
	goimports -w .

# `build` depends on `generate` so a fresh `radd` binary always
# carries the latest codegen. The generators are idempotent
# (no-op when source is unchanged), so the cost in steady state
# is a sub-second scan, not a rewrite.
build: generate
	@echo "⚙️ Building the project..."
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/radd

test:
	@echo "⚙️ Running tests..."
	go test ./core/testing/... ./rts/... ./radls/lstesting/...

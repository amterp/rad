# Makefile

# Directories
OUTPUT_DIR := ./core

# Commands
GOFMT := gofmt -w
GO := go

.PHONY: all generate format build

all: generate format build

generate:
	@echo "Running ast.go to generate code..."
	@go run core/generators/ast.go

format: generate
	@echo "Formatting generated files..."
	@find $(OUTPUT_DIR) -name 'gen_*.go' -exec $(GOFMT) {} +

build: format
	@echo "Building the project..."
	@$(GO) build ./...
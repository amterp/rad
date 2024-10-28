# Makefile

# Directories
OUTPUT_DIR := ./core
BIN_DIR := ./bin

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

build:
	@echo "Building the project..."
	@mkdir -p $(BIN_DIR)
	@$(GO) build -o $(BIN_DIR)/radd

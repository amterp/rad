# Makefile for RSL/rad

# Directories
BIN_DIR := ./bin

# Commands
.PHONY: all format build test clean

all: format build test

format:
	@echo "⚙️ Formatting files..."
	find . -name '*.go' -exec gofmt -w {} +
	goimports -w .

build:
	@echo "⚙️ Building the project..."
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/radd

test:
	@echo "⚙️ Running tests..."
	go test ./core/testing

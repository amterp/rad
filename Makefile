# Makefile for rad

# Directories
BIN_DIR := ./bin

# Commands
.PHONY: all format build test clean

all: generate format build test

generate:
	@echo "⚙️ Running generators..."
	go run "./function-metadata/extract.go"
	mv "./functions.txt" "./rts/embedded/"

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
	go test ./core/testing/... ./rts/...

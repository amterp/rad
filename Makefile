# Makefile

# Directories
OUTPUT_DIR := ./core
BIN_DIR := ./bin

# Commands

.PHONY: all format build

all: format build

format:
	@echo "⚙️ Formatting files..."
	gofmt -w **/*.go
	goimports -w .

build:
	@echo "⚙️ Building the project..."
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/radd

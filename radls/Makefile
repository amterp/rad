# Makefile

# Directories
BIN_DIR := ./bin

# Commands
GOFMT := gofmt -w
GO := go

.PHONY: all format build test

all: format build test

format:
	@echo "⚙️ Formatting code...."
	goimports -w .

build:
	@echo "⚙️ Building the project..."
	mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_DIR)/radls

test:
	@echo "⚙️ Running tests..."
	$(GO) test ./lstesting/...

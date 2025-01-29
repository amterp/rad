# Makefile

# Directories
BIN_DIR := ./bin

# Commands
GOFMT := gofmt -w
GO := go

.PHONY: all format build

all: format build

format:
	@echo "⚙️ Formatting code...."
	goimports -w .

build:
	@echo "⚙️ Building the project..."
	mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_DIR)/rsl-lsp

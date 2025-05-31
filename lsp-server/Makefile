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
	$(GO) build -o $(BIN_DIR)/rls
	cp $(BIN_DIR)/rls ../vsc-extension/bin
	cd ".." && $(GO) run "./function-metadata/extract.go"
	mv "../functions.txt" "./com/embedded/"

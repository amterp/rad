.PHONY: all format build test

all: format build test

format:
	@echo "⚙️ Formatting files..."
	find . -name '*.go' -exec gofmt -w {} +
	goimports -w .

build:
	@echo "⚙️ Building the project..."
	go build

test:
	@echo "⚙️ Running tests..."
	#go test

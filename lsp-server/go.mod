module rls

go 1.24.1

toolchain go1.24.2

require (
	github.com/sanity-io/litter v1.5.8
	github.com/tree-sitter/go-tree-sitter v0.25.0
	go.uber.org/zap v1.27.0
)

require (
	github.com/mattn/go-pointer v0.0.1 // indirect
	go.uber.org/multierr v1.11.0 // indirect
)

//replace github.com/amterp/tree-sitter-rad => ../../tree-sitter-rad

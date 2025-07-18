module rls

go 1.24.1

toolchain go1.24.2

require (
	github.com/amterp/rad/rts v0.0.0
	github.com/sanity-io/litter v1.5.8
	github.com/tree-sitter/go-tree-sitter v0.25.0
	go.uber.org/zap v1.27.0
)

require (
	github.com/amterp/color v1.20.0 // indirect
	github.com/amterp/tree-sitter-rad v0.1.6 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-pointer v0.0.1 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
)

replace github.com/amterp/rad/rts => ../rts

//replace github.com/amterp/tree-sitter-rad => ../../tree-sitter-rad

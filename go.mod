module github.com/amterp/rts

go 1.23.4

require (
	github.com/amterp/tree-sitter-rsl v0.0.15
	github.com/tree-sitter/go-tree-sitter v0.24.0
)

require github.com/mattn/go-pointer v0.0.1 // indirect

replace github.com/amterp/tree-sitter-rsl => ../tree-sitter-rsl

# Navigate

Rad/RSL & related projects are spread across repositories.

## [amterp/rad](https://github.com/amterp/rad)

Contains the `rad` CLI tool & RSL interpreter + some others.

| Directory             | Description                                              |
|-----------------------|----------------------------------------------------------|
| `core`                | Core code for rad.                                       |
| `docs-web`            | MkDocs documentation website.                            |
| `rsl-language-server` | LSP Language Server for RSL, aka RLS.                    |
| `textmate-gen`        | Generator for Textmate bundles, using [RTS](#amterprts). |
| `vsc-extension`       | Implementation for Visual Studio Code extension for RSL. |

## [amterp/tree-sitter-rsl](https://github.com/amterp/tree-sitter-rsl)

Contains RSL's [tree sitter](https://github.com/tree-sitter/tree-sitter) implementation & grammar, including the generated Go bindings.

## [amterp/rts](https://github.com/amterp/rts)

A Go lib which wraps [RSL's tree sitter](#amterptree-sitter-rsl) implementation & Go bindings.

Intended as a more convenient way of interacting with the tree sitter implementation, containing various convenience methods & greater context.

## [amterp/go-tbl](https://github.com/amterp/go-tbl)

A fork of [tablewriter](https://github.com/olekukonko/tablewriter) leveraged by rad for its table formatting and writing.

## [amterp/homebrew-rad](https://github.com/amterp/homebrew-rad)

Contains the [Homebrew](https://github.com/Homebrew/brew) formula for rad.

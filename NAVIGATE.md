# Navigate

Rad & related projects are spread across repositories.

## [amterp/rad](https://github.com/amterp/rad)

Contains the `rad` CLI tool & Rad language interpreter + some others.

| Directory       | Description                                                                                    |
|-----------------|------------------------------------------------------------------------------------------------|
| `core`          | Core code for rad.                                                                             |
| `docs-web`      | MkDocs documentation website.                                                                  |
| `lsp-server`    | LSP Language Server for Rad, aka RLS.                                                          |
| `rts`           | A Go lib which wraps [Rad's tree sitter](#amterptree-sitter-rsl) implementation & Go bindings. |
| `textmate-gen`  | Generator for Textmate bundles. Not yet implemented.                                           |
| `vsc-extension` | Implementation for Visual Studio Code extension for Rad.                                       |

## [amterp/tree-sitter-rsl](https://github.com/amterp/tree-sitter-rsl)

Contains Rad's [tree sitter](https://github.com/tree-sitter/tree-sitter) implementation & grammar, including the
generated Go bindings.

## [amterp/go-tbl](https://github.com/amterp/go-tbl)

A fork of [tablewriter](https://github.com/olekukonko/tablewriter) leveraged by rad for its table formatting and
writing.

## [amterp/homebrew-rad](https://github.com/amterp/homebrew-rad)

Contains the [Homebrew](https://github.com/Homebrew/brew) formula for rad.

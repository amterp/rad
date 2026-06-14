- Where the syntax of the Rad language itself becomes relevant, see SYNTAX.md for reference.

- You have the following subagents to request input from:
  - Code Reviewer (for when you make large changes)
  - Rad Docs Maintainer (for when you make user-facing changes)

- You have the following useful commands available to you:
  - `make format` 
  - `make build`: builds the project into a local test binary `./bin/radd`
  - `make test`
  - `./dev --validate`: Runs `go mod tidy`, formats, builds, and runs tests.

- Please do not leave task-specific messages to the user via comments in the code base when making changes.

- Never commit `replace` directives in `go.mod`. These are used locally during development to point at local
  copies of dependencies, but must be removed before committing.

---

## Pre-Commit Checklist

A Claude Code hook will remind you of this checklist when you commit. Review every item; skip categories that
don't apply to your change.

### Always
- Run `./dev --validate` (formats, builds, tests). All tests pass.
- Commit messages follow conventional prefixes (`feat:`, `fix:`, `refactor:`, `docs:`, `test:`).
- Commit messages explain **why**, not just what. See `CONTRIBUTING.md` for full conventions.

### When Adding or Modifying a Built-in Function
- Function documented in `docs/funcs/<name>.md` (source of truth - see `docs/funcs/README.md` for the required format). The signature line in `## Signature` IS the type-checker's signature; there is no parallel definition in `rts/signatures.go` anymore.
- Snapshot tests added/updated in `core/testing/snapshots/functions/<name>.snap`.
- Run `make generate`. This regenerates `rts/signatures_gen.go`, mirrors `docs/funcs/` into `rts/embedded_funcs/`, regenerates the public reference at `docs-web/docs/reference/functions.md`, and refreshes `rts/embedded/functions.txt`. Commit the regenerated artifacts.

### When Changing Language Syntax or Semantics
- `SYNTAX.md` updated to reflect the change.
- Snapshot tests added in the appropriate `core/testing/snapshots/` subdirectory.
- If AST nodes were added/changed, parser snapshot tests in `rts/test/st_snapshots/` updated.
- Guide docs updated if the feature has a section in `docs-web/docs/guide/`.

### When Introducing a Breaking Change
- Commit message uses `feat!:` or `fix!:` prefix.
- Migration guide entry added to the current version's `docs-web/docs/migrations/` file.
- Migration diagnostic added (see [Breaking Changes & Migration Diagnostics](#breaking-changes--migration-diagnostics)).

### When Adding or Modifying Error Codes
- Error doc file created/updated in `core/error_docs/<code>.md`.
- Error code defined in `rts/rl/errors.go` if new.

### When Touching Platform-Specific Behavior
- Logic centralized in `core/common/platform.go`, not scattered via `runtime.GOOS` checks.
- Paths returned to user code are normalized via `NormalizePath()`.
- Platform-specific tests in `core/testing/platform_test.go` if applicable.

---

# Rad Language - LLM Quick Reference

**Rad is a modern CLI scripting language designed to replace Bash for most scripting needs.**

## Project Overview

Rad (🤙 Rad) is a lightweight CLI scripting language that makes shell scripting easier, more readable, and more
maintainable than Bash. It combines familiar Python-like syntax with powerful scripting-specific features.

### Key Features

- **Declarative argument parsing** with automatic help generation
- **Built-in JSON processing** with path expressions
- **HTTP request syntax** (`rad url`) for API interactions
- **Table formatting** and data display
- **String interpolation** with `{variable}` syntax
- **Shell command integration** while avoiding Bash pitfalls
- **Type system** with runtime type checking
- **Interactive prompts** via `pick()` function

## Project Structure

```
├── main.go                    # Entry point - creates RadRunner
├── go.mod                     # Go module definition
├── Makefile                   # Build system (generate, format, build, test)
├── README.md                  # User documentation
├── core/                      # Interpreter (evaluates AST, no tree-sitter)
│   ├── runner.go              # Main runner logic
│   ├── interpreter.go         # AST evaluation via Go type switch
│   ├── funcs.go              # Built-in functions
│   ├── rad_block.go          # Rad block syntax (HTTP requests)
│   ├── args.go               # Argument parsing
│   ├── json_*.go             # JSON processing algorithms
│   ├── type_*.go             # Type system implementation
│   └── testing/              # Comprehensive test suite
├── rts/                      # Parsing, conversion, and static analysis
│   ├── parse.go              # Tree-sitter parser wrapper
│   ├── converter.go          # CST-to-AST single-pass converter
│   ├── nodes.go              # CST node types and traversal
│   ├── signatures.go         # Built-in function signatures
│   ├── check/                # Static checker (AST-based, CST fallback)
│   └── rl/                   # AST node types, spans, typing, node kinds
├── radls/                    # Language Server Protocol implementation
├── vsc-extension/            # VS Code extension
├── docs-web/                 # Documentation website (MkDocs)
├── benchmark/                # Performance benchmarks
└── examples/                 # Example Rad scripts
```

## Architecture Overview

### 1. Entry Point (`main.go`)

- Simple entry: creates `core.RadRunner` and calls `Run()`
- All logic delegated to core package

### 2. Core Package (`core/`)

The heart of the interpreter, organized by functionality. `core/` has **no tree-sitter dependency** - it works
entirely with Go-native AST nodes from `rts/rl/`.

#### Key Files:

- **`runner.go`**: Main execution flow, argument parsing, script loading
- **`interpreter.go`**: AST evaluation via Go type switch with `EvalResult` system
- **`funcs.go`**: 50+ built-in functions (print, len, join, etc.)
- **`args.go`**: Declarative argument parsing with constraints
- **`rad_block.go`**: Special `rad url:` syntax for HTTP requests
- **`json_*.go`**: JSON path expressions and field extraction
- **`type_*.go`**: Type system (RadValue, lists, maps, strings, etc.)

#### Built-in Functions (`funcs.go`):

Common functions include:

- **I/O**: `print`, `print_err`, `debug`, `pprint`
- **Data**: `len`, `keys`, `values`, `join`, `sort`, `unique`
- **Strings**: `upper`, `lower`, `split`, `replace`, `trim`
- **Math**: `sum`, `max`, `min`, `round`, `floor`, `ceil`
- **System**: `exit`, `sleep`, `now`, `get_env`
- **Interactive**: `pick`, `pick_kv` (user selection prompts)
- **Files**: `read_file`, `write_file`, `find_paths`
- **HTTP**: `http_get`, `http_post`

### 3. Parsing & AST (`rts/`)

Tree-sitter is the **only place CGo runs**. The rest of the system works with Go-native AST nodes.

**Pipeline**: Source code -> tree-sitter CST -> `converter.go` -> Go-native AST -> `core/` evaluates AST

- **`parse.go`**: Parser wrapper around tree-sitter-rad grammar
- **`converter.go`**: Single-pass CST-to-AST transformation. Key work: delegate chain collapsing, leaf value
  pre-parsing, operator resolution to enum, compound assign/incr-decr desugaring, string escape resolution,
  eager function body conversion.
- **`nodes.go`**: CST node types and traversal (reduced post-migration)
- **`signatures.go`**: Built-in function type signatures. Defaults are pre-converted to AST at init time.
- **`check/`**: Static checker. Walks AST for structural validation (scope checks, shadowing, assignment LHS).
  Falls back to CST for tree-sitter-specific checks (invalid nodes, scientific notation).
- **`rl/`**: The leaf package imported by everything. Contains:
  - AST node types (~36 node kinds) with `Node` interface (`Kind()`, `Span()`, `Children()`)
  - `Span` type for source location tracking
  - Typing system (type definitions, resolution, compatibility)
  - Constants, error types, utilities

### 4. Language Server (`radls/`)

- Implements LSP for VS Code integration
- Provides syntax errors, diagnostics, etc.
- Currently macOS/Linux only

## Language Syntax Quick Reference

### Script Structure

```rad
#!/usr/bin/env rad
---
Script description goes here
---
args:
    name str              # Required string argument
    count int = 5         # Optional with default
    verbose v bool        # Boolean flag (can use short form)
    
    count range (0, 100]  // Constraints
    name enum ["alice", "bob"]

// Script body - comments use //
for i in range(count):
    print("Hello {name}!")
```

### Key Syntax Features

#### Arguments

- Automatic help generation from `#` comments (help text only)
- Type checking (str, int, float, bool)
- Constraints (range, enum, regex)
- Optional vs required args
- Short form flags

#### Data Types

- **Primitives**: `str`, `int`, `float`, `bool`, `null`
- **Collections**: `list[T]`, `map[K,V]`
- **Functions**: First-class functions

#### Control Flow

```rad
// If statements
if condition:
    // do something

// For loops
for item in items:
    print(item)

// While loops  
while condition:
    // do something

// Switch expressions
result = switch value:
    case "a": "Apple"
    case "b": "Banana" 
    default: "Unknown"
```

#### Rad Blocks (HTTP Requests)

```rad
// Define JSON field mappings
Name = json[].name
Email = json[].email

// Execute HTTP request and display as table
rad "https://api.example.com/users":
    fields Name, Email
    sort Name
```

#### String Interpolation

```rad
name = "world"
message = "Hello {name}!"  // Result: "Hello world!"
```

## Development Workflow

### Build Commands

```bash
make all          # generate + format + build + test
make generate     # Extract function metadata for LSP
make format       # gofmt + goimports  
make build        # Build to ./bin/radd
make test         # Run tests in core/testing
```

### Testing

- Comprehensive test suite in `core/testing/`
- Tests organized by feature (args, functions, syntax, etc.)
- Test resources in `core/testing/resources/`
- Syntax tree snapshot tests in `rts/test/` - each case captures both CST and AST dumps side-by-side
- Converter unit tests in `rts/converter_test.go`
- Regenerate snapshots with a **targeted** `-update`, e.g. `go test ./core/testing/ -run TestSnapshots -update=types/str_lexing` (path-substring match; comma-separate multiple). A mismatch in a non-targeted file still fails, so regressions aren't silently absorbed. `-update-all` rewrites everything (avoid). Write the value with `=`; a bare `-update` errors.

### Key Dependencies

- **Tree-sitter**: For parsing only (`github.com/tree-sitter/go-tree-sitter`) - confined to `rts/`
- **pflag**: Command-line flag parsing
- **go-tbl**: Table formatting
- **samber/lo**: Utility functions
- **Various amterp/***: Author's utility packages

## Common Development Tasks

### Adding Built-in Functions

**Complete TDD workflow for adding a new built-in function:**

1. **Add function signature** to `rts/signatures.go`:
    - Add `newFnSignature()` call with proper type signature
    - Place alphabetically or near related functions

2. **Write comprehensive tests first** in `core/testing/func_[name]_test.go`:
    - **Test-Driven Development**: Write tests before implementation
    - Test basic functionality with different input types
    - Test edge cases (zero, negative numbers, boundary conditions, etc.)
    - Test error conditions (wrong types, invalid inputs)
    - Use Rad testing patterns:
        - `setupAndRunCode(t, script, "--color=never")` to run Rad scripts
        - `assertOnlyOutput(t, stdOutBuffer, "expected\n")` for success cases
        - `assertError(t, 1, expected)` for error cases with exact error message
        - `assertNoErrors(t)` to ensure no stderr output
    - Follow naming: `Test_Func_[Name]_[Scenario]`
    - **Run tests to see them fail** before implementing

3. **Add function constant** in `core/funcs.go`:
    - Add `FUNC_[NAME] = "[name]"` constant with other function constants

4. **Implement function** in `core/funcs.go`:
    - Add to `GetBuiltInFuncs()` slice with `Name` and `Execute` fields
    - Use `f.GetArg()`, `f.GetFloat()`, `f.GetStr()` etc. to extract arguments
    - Return using `f.Return()` or `f.ReturnErrf()` for errors
    - Place near related functions (e.g., math functions together)

5. **Run tests to verify implementation**:
   ```bash
   go test ./core/testing -run Test_Func_[Name]  # Test specific function
   go test ./core/testing                        # Run all tests
   ```

6. **Document the function** in `docs/funcs/<name>.md`. This is the single source of truth: the codegen pipeline derives the type-checker signature (`rts/signatures_gen.go`), the embedded LSP hover docs (`rts/embedded_funcs/<name>.md`), and the aggregate public reference page (`docs-web/docs/reference/functions.md`) from it. See `docs/funcs/README.md` for required sections and format. Do NOT edit any of those derived artifacts directly - they're regenerated by `make generate`.

7. **Regenerate derived files**: Run `make generate`. This invokes all generators (function-metadata extractor for `rts/embedded/functions.txt`, plus `go generate ./rts` which runs the three docs/funcs/ generators). `make verify-generated` is the CI gate that fails if any of these are stale.

**Rad Testing Style Guide:**

- Each test function focuses on one specific scenario
- Use descriptive test names: `Test_Func_[Name]_[Scenario]`
- Scripts use multi-line strings with proper indentation
- Always use `--color=never` flag to avoid terminal codes in output
- For error tests, include the exact expected error message with proper formatting
- Test both positive and negative cases thoroughly

### Documentation

If code changes are made, invoke the Rad Docs Maintainer agent to assess whether doc updates are needed.

#### Documentation Maintenance

Keep documentation in sync with code changes. Key mappings:

| Code Change | Documentation to Update |
|-------------|------------------------|
| New/modified built-in function in `core/funcs.go` | `docs/funcs/<name>.md` (source of truth; `make generate` propagates to the embedded docs and the aggregate reference page) |
| New/modified error code in `rts/rl/errors.go` | `core/error_docs/` (surfaced via `rad explain`) |
| Language syntax changes | `SYNTAX.md` (symlinked to Language Reference) |
| New user-facing feature | Consider adding to relevant guide in `docs-web/docs/guide/` |
| Major user-facing features, project overview | `README.md` |
| Project structure, dev workflow, new patterns | `AGENTS.md` |

**Principles:**
- **Reference docs** should be authoritative and complete — if it exists in code, it should be documented
- **Guide docs** teach concepts with examples — not every feature needs a guide section
- Avoid creating reference pages that duplicate guide content; prefer one source of truth
- When in doubt, check if existing docs already cover the topic before creating new pages

When adding or updating function documentation, edit (or create) `docs/funcs/<name>.md`. The full format contract - required sections, ordering, what hover renders vs. what only the public docs page renders, optional `## Notes` and `## See also` blocks, category conventions - lives at `docs/funcs/README.md`. Run `make generate` to propagate edits to the embedded LSP docs and the aggregate reference page.

A few principles worth keeping in mind while authoring:

- **Match the signature to runtime behavior.** Check `core/funcs.go` for mutually exclusive parameters, argument interactions, and error conditions that aren't obvious from the signature alone.
- **Examples earn their keep.** Use inline comments showing results (`pow(2, 3)  // -> 8`). The first example block is what hover renders; later blocks are public-docs-only.
- **Keep prose scannable.** Reference docs are for skimming, not tutorials. Tutorials belong in `docs-web/docs/guide/`.

### Breaking Changes & Migration Diagnostics

Rad is young and we advertise that breaking changes happen in minor versions. But we still want migrations to be
as easy as possible for our users. When introducing a breaking change, provide **migration diagnostics** that detect
old usage patterns and guide users to the fix.

The goal is a three-layer help system: a concise inline hint at point-of-error, a deeper `rad explain` doc, and
a comprehensive migration guide page.

#### What to do when making a breaking change

1. **Add a migration diagnostic** that detects the old pattern and emits a helpful error (or warning, depending on
   context - usually error). The diagnostic should:
   - Clearly state what changed
   - Suggest the fix concisely
   - Link to the migration doc: `https://amterp.dev/rad/migrations/v0.X/`

2. **Add/update an error doc** in `core/error_docs/<code>.md` with a before/after example and fix steps.

3. **Add a migration guide entry** in `docs-web/docs/migrations/v0.X.md` with full context and rationale.

#### Diagnostic patterns by change type

**Renamed function** - detect the old name at runtime and emit a hint:
```go
// In the function dispatch (e.g. core/func_helpers.go or similar)
case "old_name":
    i.emitErrorWithHint(rl.ErrUnknownFunction, funcExpr,
        "Cannot invoke unknown function: old_name",
        "old_name was renamed to new_name. See: https://amterp.dev/rad/migrations/v0.X/")
```

**Removed function** - same pattern, different hint:
```go
case "get_default":
    i.emitErrorWithHint(rl.ErrUnknownFunction, funcExpr,
        "Cannot invoke unknown function: get_default",
        "get_default was removed. Use: map[\"key\"] ?? default. See: https://amterp.dev/rad/migrations/v0.8/")
```

**Changed operator/syntax behavior** - the type checker or interpreter naturally catches the new error; ensure the
error message is clear and add a hint pointing to the migration docs. If the old usage now triggers an existing error
code, update that error doc with a migration note.

**Static detection** - for patterns detectable before execution, add checks in `rts/check/` using
`NewDiagnosticError()` or `NewDiagnosticErrorWithSuggestion()`. This also benefits the LSP (editor diagnostics).

#### What users see

The diagnostic renderer produces Rust-style output:
```
error[RAD40003]: Cannot invoke unknown function: get_stash_dir
  --> script.rad:5:1
    |
  5 | get_stash_dir()
    | ^^^^^^^^^^^^^^^ Cannot invoke unknown function: get_stash_dir
    |
   = help: get_stash_dir was renamed to get_stash_path. See: https://amterp.dev/rad/migrations/v0.9/
   = info: rad explain RAD40003
```

Users can then run `rad explain RAD40003` for the full error doc, or visit the migration page for broader context.

### Debugging Tips

- Use `debug()` function in Rad scripts for debugging
- `--cst-tree`: Dump the tree-sitter CST for a script
- `--ast-tree`: Dump the Go-native AST for a script (runs converter)
- Both flags bypass arg validation, so they work on scripts with missing required args
- Check `core/testing/` for examples of every language feature

## File Extensions and Conventions

- **`.rad`**: Rad script files
- **`#!/usr/bin/env rad`**: Shebang for executable scripts
- Scripts typically have no extension when installed as CLI tools

## Status: Early Development

- Major breaking changes expected
- Core functionality working
- Missing some planned features
- Active development on language features and LSP

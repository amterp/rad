# Type System

A living reference for Rad's type system: what it does today, where the
pieces live, and what's deliberately left for later. This is an internal
doc, not user-facing - it's meant for contributors working on or near
the typing machinery.

Stub for now; sections will fill in over time as the system grows.

## Overview

Rad's type system has three layers:

- **Grammar** (`tree-sitter-rad/grammar.js`) defines the surface syntax
  for type annotations: scalars, unions, optionals, lists, tuples,
  structs, maps, function types.
- **Type internals** (`rts/rl/typing.go`) define the runtime
  representation of each type (`TypingT` interface and concrete impls)
  and the compatibility check (`IsCompatibleWith`).
- **Runtime enforcement** (`core/type_fn.go`) calls into the type
  system at function call boundaries - parameters, return values,
  variadic and default-value checks.

The static checker (`rts/check/`) and LSP layer share the parser and
AST but currently do no type analysis. See "Deferred" below.

## What is enforced today

Type annotations are checked at function call boundaries. That means:

- Each positional or named arg is checked against the declared param
  type.
- Variadic args are checked element-by-element against the variadic
  type.
- Return values are checked against the declared return type.
- Default values are checked against the param type they belong to.
- Nested collection types check inner element types at every nesting
  level (e.g. `int[][]` rejects a string anywhere inside).

Args block fields (CLI argument types) are enforced separately, at
CLI parse time, in `core/args.go`.

## Deferred / Known Gaps

These are intentional gaps - things the type system does not enforce
today. They are grouped here so the boundary is easy to find when
adding to the system.

### No typed local variables

Rad has no syntax for type annotations on local variable declarations:

```rad
x: int = 5    # not parseable today
x: int        # not parseable today
```

The grammar only carries type annotations on function parameters,
return types, and args block fields. There is no `Identifier`-with-type
AST node; locals are inferred purely from assigned values.

### Collections do not carry declared element types

A `*RadList` and `*RadMap` hold their elements but *not* the declared
element type they were constructed under. So:

- `list.append(x)` does **not** check `x` against the list's element
  type.
- `m[k] = v` does **not** check key/value types.
- Indexed assignment is unchecked.

This is fine while the only place collection types are enforced is at
function call boundaries (where we check the *whole* value against the
parameter type). Once typed locals exist, we'll likely need a type tag
on `RadList`/`RadMap` to enforce mutations against the declared type
at the mutation site.

### `TypingFnT.IsCompatibleWith` only checks "is some fn"

`TypingFnT.IsCompatibleWith` returns true if the value is *any*
function, regardless of declared param/return shape. Structural
matching between a declared `fn(int) -> bool` type and a passed
function value isn't done because `TypingCompatVal.Val` doesn't carry
function arity/return info - `RadFn` lives in the `core` package,
which `rl` cannot import.

Practical effect: `fn(int) -> bool` accepts any function. Fixing this
requires either:

- Enriching `TypingCompatVal` with optional fn-shape data, populated
  by `RadValue.ToCompatSubject()` in core, or
- Moving the structural compare into core (where `RadFn` is visible)
  and calling it from `core/type_fn.go` typeCheck instead of going
  through `IsCompatibleWith`.

### LSP does no type analysis yet

`rts/check/` (the static analyzer feeding LSP diagnostics) currently
surfaces:

- Syntax/parse errors (delegated to tree-sitter)
- Function shadowing
- Unknown function hints
- Break/continue/return scope errors
- Invalid assignment LHS
- Deprecated block keywords / no-effect Rad options
- Scientific notation in int defaults

It does **not** do:

- Type mismatch detection
- Wrong argument count
- Undefined variable detection
- Unused variable lint
- Unreachable code

All of these are appropriate for the dedicated static-typing
milestone.

### Performance note on type-checking subjects

`RadValue.ToCompatSubject()` calls `ToGoList()` / `ToGoMap()` which
deep-copy collection contents into `[]interface{}` /
`map[string]interface{}`. This happens at every typed function call
boundary. The cost is O(N) for N-element collections and could matter
on hot paths. The current architecture forces this because `rl` (where
`TypingCompatVal` lives) can't depend on `core` (where `*RadList` /
`*RadMap` live). A future change could introduce a shared-package
iteration interface to avoid the copy, but until measurements say it
matters we accept the cost.

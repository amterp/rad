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

## Scoping & name resolution

Static analysis runs a single binder pass over the AST that produces a
`Resolved` value: the scope tree, identifier-use -> symbol map, and a
list of binder-detected issues (`BindIssue`). All downstream checks
that need to ask "what does this name mean here?" go through the
resolved view rather than re-deriving their own. Code lives in
`rts/check/resolve.go` and `rts/check/binder.go`.

Resolved is a pure value over the AST - no source-text dependency, no
mutation - so the LSP can hold one per snapshot and read it lock-free.

### Scopes

Each scope chains to its parent. Lookup walks the chain. The kinds:

- `ScopeBuiltin` - ambient runtime names (`print`, `len`, ...). Sits
  above file scope; symbols are synthesized lazily on first reference
  so cold builtins don't cost.
- `ScopeFile` - script body. Holds hoisted functions and args-block
  declarations.
- `ScopeFunction` / `ScopeLambda` - function and lambda bodies. Hold
  their parameter bindings.
- `ScopeLoop` - `for` and `while` bodies. Holds loop variables and the
  optional `with` context binding.
- `ScopeListComp` - list comprehension scope. Same shape as a loop
  scope but for the comprehension.
- `ScopeBlock` - switch case body, defer body. Locals declared inside
  don't leak.
- `ScopeCmdBlock` - cmd_block body. Holds the command's arg
  declarations.

### Symbol kinds

- `SymBuiltin` - ambient name from the runtime.
- `SymHoistedFn` - top-level `fn` definition. Visible across the
  whole file regardless of source order. Nested `fn` defs aren't
  hoisted; they bind at point of declaration in their enclosing scope.
- `SymArg` - declared in the script-level `args:` block. Acts as a
  file-scope ambient local; the runtime populates it from CLI flags
  before the body executes.
- `SymCmdArg` - declared in a `cmd_block`'s args. Lives in that
  command's scope.
- `SymParam` - function or lambda parameter.
- `SymLocal` - any other name introduced by assignment.
- `SymLoopVar` - the binding from `for x in ...`.
- `SymWith` - the `with` context binding on a `for` loop or
  comprehension.

### Hoisting

Top-level `fn` definitions are hoisted into the file scope before any
statement is visited, so calls earlier in the file can refer to
definitions later in the file. The hoist pass also makes function
self-reference work for recursion: the body's scope chains up through
the file scope where the function's own name lives.

### Args block defaults

Default expressions for args-block declarations are visited *after*
every arg has been declared in file scope. Forward references across
args (`a int = b, b int = 5`) resolve at the binder level; the runtime
may still impose ordering constraints.

### Param defaults

Function and lambda parameter default expressions are visited in the
*enclosing* scope, not inside the function's own scope. A default
like `fn f(n = greeting)` looks up `greeting` where the function was
defined; it does not see sibling parameters. This matches Python and
avoids the surprise of one parameter's default referencing a later
parameter's name.

### Assignment: plain vs compound

The `Assign` AST node carries an `UpdateEnclosing` flag:

- **Plain `=`** (`UpdateEnclosing = false`): the LHS identifier is
  declared as a fresh local in the *current* scope. If a same-named
  binding exists in an enclosing scope, the new local shadows it.
- **Compound (`+=`, `++`, `--`, unpack-with-rebind)** (`UpdateEnclosing
  = true`): the LHS must resolve to an existing binding somewhere up
  the scope chain. Without one the operation has nothing to operate
  on. The binder records the target as a *use* of the existing
  binding rather than introducing a new local.

`VarPath` targets (`a.b`, `xs[i]`) mutate an existing path's contents
and never introduce new bindings; the binder visits the root
identifier as a normal expression use.

### Loops and comprehensions

For-loops, while-loops, and list comprehensions each open their own
scope. The iterable expression of a for-loop or comprehension is
visited in the *enclosing* scope so `[x for x in x]` self-reference
makes sense (the outer `x` flows into the new binding via the iter).
While-loop conditions visit inside the loop scope, but the loop header
introduces no bindings.

Loop locals don't outlive the loop body. Compound assigns inside loops
still rebind enclosing-scope bindings - so a `i += 1` inside a `while`
updates the file-scope counter rather than introducing a loop-local
shadow that no caller can see.

### Switch and defer

Each switch case body that contains statements opens a `ScopeBlock`;
locals declared in case A aren't visible to case B or to code after
the switch. The discriminant and case-match keys stay in the enclosing
scope - they're expressions, not bindings.

Defer and errdefer bodies are also `ScopeBlock`. The deferred code can
still reference everything from the surrounding function through
normal lookup; only what *it* declares stays inside.

### Cmd blocks

Each top-level `command` block has its own scope owning the command's
args. An inline-lambda callback runs inside that scope, so the lambda
body can see the args. A name-referenced callback (`calls handler`)
resolves against file scope only - the runtime threads the args to a
named function through a separate mechanism, and pretending the
function's lexical scope sees them would mislead hover/goto-def.

### Binder-emitted diagnostics

The binder records structural problems as `BindIssue` records on
`Resolved.Issues`. Today there's one: `ErrDuplicateParameter` for two
parameters in the same parameter list sharing a name. The checker
converts each issue to a `Diagnostic` using its source text.

Undefined-variable diagnostics are not yet emitted from the binder.
The existing `addUnknownFunctionHints` produces a Hint for unknown
function-call callees, via the resolved view; broader uses are
caught at runtime today with rich "did you mean" suggestions.
Promoting to a static error will require migrating the test corpus
that exercises the runtime path.

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
- Unknown function hints (Hint severity, routed through the resolved
  view; see Scoping section)
- Duplicate function/lambda parameters
- Break/continue/return scope errors
- Invalid assignment LHS
- Deprecated block keywords / no-effect Rad options
- Scientific notation in int defaults

It does **not** do:

- Type mismatch detection
- Wrong argument count
- Undefined variable detection for non-call references (caught at
  runtime today)
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

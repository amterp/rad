# For-Loop Syntax Redesign

**Date:** January 2026
**Status:** Spec
**Breaking Change:** Yes

---

## Problem

The current for-loop syntax has a confusing semantic shift when adding variables:

```rad
// One variable: item
for item in items:
    print(item)

// Two variables: FIRST becomes index, SECOND is item
for idx, item in items:
    print(idx, item)
```

Adding a second variable changes the meaning of the first variable. This is unintuitive and leads to:
- Confusion when learning the language
- Lots of `_` placeholders when you don't care about the index
- Awkward patterns when iterating parallel collections with `zip()`

---

## New Syntax

### Basic Iteration (unchanged)

```rad
for item in items:
    print(item)
```

### Multiple Variable Unpacking

Variables now always correspond to values being unpacked, with no implicit index prefix:

```rad
// Unpack each element into two values
for name, age in [["alice", 25], ["bob", 30]]:
    print(name, age)

// With zip
for name, age, city in zip(names, ages, cities):
    print(name, age, city)
```

### Context Object

Index and other metadata are accessed via an optional context object using `with`:

```rad
for item in items with ctx:
    print(ctx.idx, item)
```

The context variable name is user-defined (`ctx` is the recommended convention).

---

## Context Object Fields

| Field | Type | Description |
|-------|------|-------------|
| `idx` | int | Current 0-based index |
| `src` | any | The collection being iterated |

Other properties (`first`, `last`, `len`) can be derived:
- `ctx.idx == 0` for first
- `ctx.idx == ctx.src.len() - 1` for last
- `ctx.src.len()` for length

---

## Examples

### Simple iteration with index

```rad
for item in items with ctx:
    print("{ctx.idx}: {item}")
```

### Parallel iteration with zip

```rad
names = ["alice", "bob", "charlie"]
ages = [25, 30, 35]

for name, age in zip(names, ages) with ctx:
    print("{ctx.idx}. {name} is {age}")
```

### Accessing the source collection

```rad
for item in get_data() with ctx:
    if ctx.idx < ctx.src.len() - 1:
        next_item = ctx.src[ctx.idx + 1]
        print("{item} -> {next_item}")
```

### Map iteration

```rad
person = {"name": "alice", "age": 25}

for key in person with ctx:
    print("{ctx.idx}: {key} = {person[key]}")
```

---

## Breaking Change Migration

### Old syntax (no longer works)

```rad
for idx, item in items:  // OLD: idx was index, item was value
```

### New syntax

```rad
for item in items with ctx:
    print(ctx.idx, item)
```

### Migration Error

When we detect:
1. Two-variable unpack (`for a, b in collection:`)
2. Collection is a flat list (not list of lists)
3. First variable is literally `idx` or `index`

Emit a helpful error:

```
Error: Cannot unpack 'int' into 2 values

Note: The for-loop syntax changed in January 2026. It looks like you may
be using the old syntax where the first variable was the index.

Old: for idx, item in items:
New: for item in items with ctx:
         print(ctx.idx, item)
```

---

## Rad Block Lambda Integration

The same context concept applies to rad block `map` operations. When the lambda accepts two parameters, the second is the context object.

### Single field

```rad
timestamp:
    map fn(x, ctx) x.colorize(ctx.src, skip_if_single=true)
```

### Grouped fields (new syntax)

Multiple fields can share the same transform:

```rad
timestamp, strategy, fund, profile:
    map fn(x, ctx) x.colorize(ctx.src, skip_if_single=true)
```

### Rad Block Lambda Context Fields

| Field | Type | Description |
|-------|------|-------------|
| `idx` | int | Current row index |
| `src` | list | The full column data |
| `field` | str | The field name (e.g., "timestamp") |

### Conditional logic with `field`

```rad
timestamp, strategy, fund:
    map fn(x, ctx):
        if ctx.field == "timestamp":
            x = x.str()
        return x.colorize(ctx.src, skip_if_single=true)
```

---

## Grammar Changes

### For-loop

```
for_stmt = "for" identifier_list "in" expression ["with" identifier] ":" block
```

### Rad block field list

```
field_modifier = identifier_list ":" modifier_block
identifier_list = identifier ("," identifier)*
```

---

## Convention

The recommended convention for the context variable name is `ctx`:

```rad
// For-loops
for item in items with ctx:

// Rad block lambdas
map fn(x, ctx) x.transform(ctx.src)
```

This is a recommendation, not enforced. Users may use `loop`, `meta`, `_`, or any valid identifier.

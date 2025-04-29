# Stashes / Script State

## 2025-04-15

Every script has its own store, stored with rad's own config home.

`~/.rad/scripts/1u4HSWY2mLP/data.json`

Example 'uh' implementation (RSL impl/variation of um):

```rsl
set_script_id("1u4HSWY2mLP")

ensure_state("entries", {})
ensure_state("editor", "vim", prompt="What editor command would you like to use?")

state = load_state()
defer save_state(state)

existing = state.entries.get_default(command, "")
path = temp_file()
path.write_file(existing)

editor = state.editor
$!`{editor} {path}`
contents = path.read_file()
state.entries[command] = contents

//

cmd_file = load_script_data_file(command, "") // keys: 'contents', 'path'
$!`vim {cmd_file.path}`
```

```rsl
set_script_id("1u4HSWY2mLP")

ensure_state("entries", {})
ensure_state("editor", "vim", prompt="What editor command would you like to use?")

state = load_state()
defer save_state(state)

existing = state.entries.get_default(command, "")
path = temp_file()
path.write_file(existing)

editor = state.editor
$!`{editor} {path}`
contents = path.read_file()
state.entries[command] = contents

//

cmd_file = load_script_data_file(command, "") // keys: 'contents', 'path'
$!`vim {cmd_file.path}`
```

```
set_script_id(id: string)
load_state() -> map
save_state(data: map)
load_script_data_file(file: str, default: string) -> map { contents: string, path: string }
write_script_data_file(file: str, contents: string) -> path: string, error map
get_script_home() -> path: string

get_default(map, default) -> any
```

---

## 2025-04-26

First-time setup on script configuration.

```
args:
    redo_config bool
    // OR
    editor string

set_stash_id("J3NdHcNJxDI")

s = load_state()
defer save_state(s)
c = s.config

editor = c.load("editor", () -> input("Enter an editor > "), reload=redo_config) // forces use of loader lambda
// OR
editor = c.load("editor", () -> input("Enter an editor > "), override=editor) // overrides existing value if present, skips loader if not present, just use override and save it into 'c'
```


^^ requires lambdas. Let's tangent.

### Lambdas

```
// two args
(a, b) -> a * b
(_, b) -> b * 2
(_, _) -> 2
```

```
// no args
() -> 2
```

```
// one arg
a -> a * 2
_ -> 2
```

```
// block
foo = a ->
    b = 2 * a
    return b
```

### Changing `set_stash_id`

- Perhaps we still allow it, but I think we should offer a statically-inspectable way for rad to know the stash ID used by a script.
- This would allow us to do something like `rad state show <script>` or `rad state delete <script>`.
- I think it could make sense to put in the file header.

```
---
This script does cool stuff.
@stash_id("J3NdHcNJxDI")
---
```

This is the same syntax I was thinking about a while ago, basically a sort of 'macro', though not exactly. It's a meta syntax for Rad in the file header.

```
---
This script does cool stuff.
Example invocation: @script alice 30
---
```

That, for example, would replace `@script` with the actual script name for the help usage string.

We likely want a syntax less likely to collide with common writings in a file header, though. Or at least offer escaping e.g. `\@script` or something. Or `{{script}}` / `{{stash_id("J3NdHcNJxDI")}}`.

^ Wanna stick away from code that has to actually be evaluated. Like those `""` scare me into thinking we need to interpret. Hence just `@stash_id J3NdHcNJxDI` might be nice. 

Also, we need to exclude `@stash_id` from the help string that gets generated. Logic can be: delete the token, and if a newline follows the token, delete that too.

## 2025-04-27

### Lambdas 2

We should think about functions as well. Custom functions may really just turn out to be variable-assigned lambdas?

### Alternative 1 (Adapted Java style)

```rsl
normalize = x -> x.trim().lower()
 
normalize(mystring)
mylist.map(normalize)

normalize = x ->
    out = x.trim().lower()
    return out
```

Pro: Concise.

Con: `->` begins a block, which is unique from rest of language where `:` does that.

### Alternative 2 (Go style)

```rsl
normalize = func(x) x.trim().lower()
 
normalize(mystring)

normalize = func(x):
    out = x.trim().lower()
    return out
```

Pro: `:` begins block, like it does in most other parts of the language.

Con: Single liner is a little verbose though. Is there a best of both worlds?

### Alternative 3 (Adapted Go style)

```rsl
normalize = fn x: x.trim().lower()
 
normalize(mystring)

normalize = fn x:
    out = x.trim().lower()
    return out

provide = fn: 5
provide()  // returns 5

multiply = fn x, y: x * y  // uncertain on if x and y should be comma or just space-separated.

mylist.map(fn x: x.upper())
mylist.map(upper)  // technically, it'd need to redefine all my built-ins as function vars, so they can be passed this way
```

Pro: Easier parsing because `func x` won't get interpreted as a call, unlike `func(x)`.  

Con: Single liner still a little verbose.

### Alternative 4 (Final?)

```rsl
normalize = fn(x) x.trim().lower()
 
normalize(mystring)

normalize = fn(x):
    out = x.trim().lower()
    return out

provide = fn() 5
provide()  // returns 5

multiply = fn(x, y) x * y

mylist.map(fn(x) x.upper())
mylist.map(upper)  // technically, it'd need to redefine all my built-ins as function vars, so they can be passed this way
```

Pro: Easier parsing because `func x` won't get interpreted as a call, unlike `func(x)`.  

Con: Single liner still a little verbose.

...

Continues on [custom_functions.md](./custom_functions.md).

### Importing functions

We might want to support importing functions from other scripts?

```rsl
// script1.rsl

normalize = x -> x.trim().lower()

// script2.rsl

from script1.rsl import { normalize }   // << could omit '{ }' for single case
```

Something like that?

How would rad know where to look for script1.rsl though? Relative paths? Would mean needing to store all these together.
What if you want to not do that, and have a 'central repo' for all these helper functions you can then use in different scripts?
This goes against one of the advantages of rad so far: all batteries are already included, and you need no additional dependencies.
You could argue resources are already an external dependency, in the same way these imported functions would be. hmm.

Maybe relative importing is the way to go.
Store all your scripts in one location with a relative path `../shared/*` that you can import from? `from ../shared/script1.rsl import normalize`.

More here: [imports.md](./imports.md).

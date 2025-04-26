# Script State

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
||||||| Stash base
=======
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

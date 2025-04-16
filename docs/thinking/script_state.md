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

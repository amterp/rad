---
title: hm
---

## Preview

```rad linenums="1" hl_lines="0"
#!/usr/bin/env rad
---
A personal cheatsheet manager.
Store and retrieve help snippets for commands you forget.
@stash_id = J8xKmN3pQrT
---

command show:
    ---
    Show the help entry for a topic.
    ---
    topic str  # The topic to look up.
    calls do_show

command edit:
    ---
    Edit or create a help entry for a topic.
    ---
    topic str  # The topic to edit.
    calls do_edit

command list:
    ---
    List all stored topics.
    ---
    calls do_list

state = load_state()
defer save_state(state)

fn do_show():
    file_path = get_stash_path("files/entries/{topic}.txt")
    if not file_path.get_path().exists:
        print("No entry for '{topic}'. Use 'hm edit {topic}' to create one.")
    else:
        print(file_path.read_file().content)

fn do_edit():
    editor = state.load("editor", fn() input("Editor? > ", default="vim"))
    result = load_stash_file("entries/{topic}.txt", "")
    $`{editor} {result.full_path}`
    print("Entry for '{topic}' saved.")

fn do_list():
    entries_dir = get_stash_path("files/entries")
    if not entries_dir.get_path().exists:
        print("No entries yet. Use 'hm edit <topic>' to create one.")
        exit()
    files = entries_dir.find_paths(depth=1, relative="absolute")
    if files.len() == 0:
        print("No entries yet. Use 'hm edit <topic>' to create one.")
        exit()
    for file in files:
        name = file.get_path().base_name.replace(".txt", "")
        print(name)
```

```
> hm -h
A personal cheatsheet manager.
Store and retrieve help snippets for commands you forget.

Usage:
  hm [command] [OPTIONS]

Commands:
  edit   Edit or create a help entry for a topic.
  list   List all stored topics.
  show   Show the help entry for a topic.
```

```
> hm edit tar
Editor? > vim
# Opens vim with empty file, user writes their notes, saves and exits
Entry for 'tar' saved.

> hm show tar
tar - extract and create archives

Extract:    tar -xvf archive.tar.gz
Create:     tar -cvzf archive.tar.gz files/
List:       tar -tvf archive.tar.gz

> hm edit rsync
# No prompt this time - editor preference was saved!
Entry for 'rsync' saved.

> hm list
tar
rsync
```

## Tutorial: Building `hm`

### Motivation

Some commands have notoriously hard-to-remember syntax - `tar`, `find`, `rsync`. You look them up, use them once, and
forget again. Tools like `tldr` help, but sometimes you want your *own* notes - the specific incantations that work for
your use cases.

Let's build `hm` - a personal cheatsheet manager that stores your notes in a [stash](../guide/stashes.md), so they
persist between sessions.

### Writing the script

Let's start with a new script:

```sh
rad new hm -s
```

First, we'll add a description and set up a stash ID. The stash ID is what tells Rad where to store our persistent data.
You can generate one with `rad gen-id`:

```rad linenums="1" hl_lines="2-6"
#!/usr/bin/env rad
---
A personal cheatsheet manager.
Store and retrieve help snippets for commands you forget.
@stash_id = J8xKmN3pQrT
---
```

The `@stash_id` macro in the file header creates a dedicated storage area at `~/.rad/stashes/J8xKmN3pQrT/` for this
script.

### Defining commands

We'll use [script commands](../guide/script-commands.md) to organize our CLI into commands. Each command gets its own
description and arguments:

```rad linenums="1" hl_lines="8-26"
#!/usr/bin/env rad
---
A personal cheatsheet manager.
Store and retrieve help snippets for commands you forget.
@stash_id = J8xKmN3pQrT
---

command show:
    ---
    Show the help entry for a topic.
    ---
    topic str  # The topic to look up.
    calls do_show

command edit:
    ---
    Edit or create a help entry for a topic.
    ---
    topic str  # The topic to edit.
    calls do_edit

command list:
    ---
    List all stored topics.
    ---
    calls do_list
```

Each command block declares its own arguments (like `topic str`) and uses `calls` to specify which function handles it.
The `---` sections become the command descriptions shown in `--help`.

### Setting up state

Before implementing our commands, we need to set up state management. We'll use state to remember user preferences (like
their preferred editor):

```rad linenums="1" hl_lines="28-29"
#!/usr/bin/env rad
---
A personal cheatsheet manager.
Store and retrieve help snippets for commands you forget.
@stash_id = J8xKmN3pQrT
---

command show:
    ---
    Show the help entry for a topic.
    ---
    topic str  # The topic to look up.
    calls do_show

command edit:
    ---
    Edit or create a help entry for a topic.
    ---
    topic str  # The topic to edit.
    calls do_edit

command list:
    ---
    List all stored topics.
    ---
    calls do_list

state = load_state()
defer save_state(state)
```

The `defer save_state(state)` ensures our state is saved when the script exits, even if something goes wrong later.

### The show command

Now let's implement `do_show()`. We check if the entry exists and display it:

```rad linenums="1" hl_lines="31-36"
#!/usr/bin/env rad
---
A personal cheatsheet manager.
Store and retrieve help snippets for commands you forget.
@stash_id = J8xKmN3pQrT
---

command show:
    ---
    Show the help entry for a topic.
    ---
    topic str  # The topic to look up.
    calls do_show

command edit:
    ---
    Edit or create a help entry for a topic.
    ---
    topic str  # The topic to edit.
    calls do_edit

command list:
    ---
    List all stored topics.
    ---
    calls do_list

state = load_state()
defer save_state(state)

fn do_show():
    file_path = get_stash_path("files/entries/{topic}.txt")
    if not file_path.get_path().exists:
        print("No entry for '{topic}'. Use 'hm edit {topic}' to create one.")
    else:
        print(file_path.read_file().content)
```

We use `get_stash_path()` to build the path to where the entry file would live. The `"files/"` prefix is needed because
stash files are stored in a `files/` subdirectory. Then we chain `.get_path().exists` to check if it exists - if not, we
tell the user how to create it. If it does exist, we read and print it with `.read_file().content`.

### The edit command

The edit command opens the entry in the user's editor. Here's where it gets interesting - we use [
`load()`](../reference/functions.md#load) to handle first-run configuration:

```rad linenums="1" hl_lines="38-42"
#!/usr/bin/env rad
---
A personal cheatsheet manager.
Store and retrieve help snippets for commands you forget.
@stash_id = J8xKmN3pQrT
---

command show:
    ---
    Show the help entry for a topic.
    ---
    topic str  # The topic to look up.
    calls do_show

command edit:
    ---
    Edit or create a help entry for a topic.
    ---
    topic str  # The topic to edit.
    calls do_edit

command list:
    ---
    List all stored topics.
    ---
    calls do_list

state = load_state()
defer save_state(state)

fn do_show():
    file_path = get_stash_path("files/entries/{topic}.txt")
    if not file_path.get_path().exists:
        print("No entry for '{topic}'. Use 'hm edit {topic}' to create one.")
    else:
        print(file_path.read_file().content)

fn do_edit():
    editor = state.load("editor", fn() input("Editor? > ", default="vim"))
    result = load_stash_file("entries/{topic}.txt", "")
    $`{editor} {result.full_path}`
    print("Entry for '{topic}' saved.")
```

The `load()` function is the star here. It checks if `"editor"` exists in `state`:

- **First run**: Key doesn't exist, so it calls the loader function - which prompts the user with `input()`. The result
  gets stored in `state["editor"]`.
- **Subsequent runs**: Key exists, so it returns the cached value immediately - no prompt.

Since we have `defer save_state(state)`, the editor preference persists between sessions.

### The list command

Finally, we list all stored entries by finding files in the entries directory:

```rad linenums="1" hl_lines="44-56"
#!/usr/bin/env rad
---
A personal cheatsheet manager.
Store and retrieve help snippets for commands you forget.
@stash_id = J8xKmN3pQrT
---

command show:
    ---
    Show the help entry for a topic.
    ---
    topic str  # The topic to look up.
    calls do_show

command edit:
    ---
    Edit or create a help entry for a topic.
    ---
    topic str  # The topic to edit.
    calls do_edit

command list:
    ---
    List all stored topics.
    ---
    calls do_list

state = load_state()
defer save_state(state)

fn do_show():
    file_path = get_stash_path("files/entries/{topic}.txt")
    if not file_path.get_path().exists:
        print("No entry for '{topic}'. Use 'hm edit {topic}' to create one.")
    else:
        print(file_path.read_file().content)

fn do_edit():
    editor = state.load("editor", fn() input("Editor? > ", default="vim"))
    result = load_stash_file("entries/{topic}.txt", "")
    $`{editor} {result.full_path}`
    print("Entry for '{topic}' saved.")

fn do_list():
    entries_dir = get_stash_path("files/entries")
    if not entries_dir.get_path().exists:
        print("No entries yet. Use 'hm edit <topic>' to create one.")
        exit()
    files = entries_dir.find_paths(depth=1, relative="absolute")
    if files.len() == 0:
        print("No entries yet. Use 'hm edit <topic>' to create one.")
        exit()
    for file in files:
        name = file.get_path().base_name.replace(".txt", "")
        print(name)
```

We use [`get_stash_path()`](../reference/functions.md#get_stash_path) to get the path to our entries folder - note the
`"files/"` prefix since stash files live in a `files/` subdirectory. We chain [
`.get_path()`](../reference/functions.md#get_path) to check if it exists, and [
`.find_paths()`](../reference/functions.md#find_paths) to list all files. The `depth=1` parameter limits the search to
direct children only, and `relative="absolute"` gives us full paths so that subsequent `get_path()` calls resolve
correctly. For each file, we extract the topic name with `.get_path().base_name` and strip the `.txt` extension.

### Try it out

Create your first entry:

```
> hm edit tar
Editor? > vim
# Editor opens, you write your notes, save and exit
Entry for 'tar' saved.
```

View it later:

```
> hm show tar
tar - extract and create archives

Extract:    tar -xvf archive.tar.gz
Create:     tar -cvzf archive.tar.gz files/
List:       tar -tvf archive.tar.gz
```

Add another entry - notice no editor prompt this time:

```
> hm edit rsync
# Opens vim immediately - preference was remembered!
Entry for 'rsync' saved.
```

List all your entries:

```
> hm list
tar
rsync
```

Your notes live at `~/.rad/stashes/J8xKmN3pQrT/files/entries/` and your preferences are stored in
`~/.rad/stashes/J8xKmN3pQrT/state.json`.

## Concepts demonstrated

| Concept                                                           | Where                                              |
|-------------------------------------------------------------------|----------------------------------------------------|
| [Stash ID](../guide/stashes.md)                                   | `@stash_id = J8xKmN3pQrT`                          |
| [State persistence](../guide/stashes.md#state-storage)            | `load_state()` / `save_state()`                    |
| [Defer pattern](../guide/stashes.md#the-defer-save-state-pattern) | `defer save_state(state)`                          |
| [`load()`](../reference/functions.md#load)                        | First-run config with `state.load("editor", ...)`  |
| [`input()`](../reference/functions.md#input)                      | Prompting for editor preference                    |
| [`load_stash_file()`](../reference/functions.md#load_stash_file)  | Creating entry files in `do_edit()`                |
| [`read_file()`](../reference/functions.md#read_file)              | Reading entry content in `do_show()`               |
| [`get_stash_path()`](../reference/functions.md#get_stash_path)    | Getting entries paths                              |
| [Script commands](../guide/script-commands.md)                    | `command show:`, `command edit:`, `command list:`  |
| [Shell commands](../guide/shell-commands.md)                      | `$\`{editor} {result.full_path}\``                 |
| [`find_paths()`](../reference/functions.md#find_paths)            | Listing files in the entries directory             |
| [`get_path()`](../reference/functions.md#get_path)                | Checking if directory exists, extracting base name |
| [Custom functions](../guide/functions.md)                         | `fn do_show():`, etc.                              |

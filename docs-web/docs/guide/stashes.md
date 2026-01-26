---
title: Stashes
---

Sometimes your script needs to remember things between runs - user preferences, counters, cached data, or configuration.
Stashes are Rad's built-in mechanism for persistent, per-script storage.

## The Basics

Every script can have its own **stash** - a dedicated storage area that persists between runs.
To use a stash, you need to declare a **stash ID** in your script's file header using the `@stash_id` macro:

```rad title="counter.rad"
#!/usr/bin/env rad
---
A simple counter that remembers how many times it's been run.
@stash_id = JJRNBOSKHpa
---

state = load_state()
defer save_state(state)

count = state.get_default("count", 0)
count++
state["count"] = count

print("This script has been run {count} time(s)!")
```

Let's run it a few times:

```shell
rad counter.rad
```

<div class="result">
```
This script has been run 1 time(s)!
```
</div>

```shell
rad counter.rad
```

<div class="result">
```
This script has been run 2 time(s)!
```
</div>

The count persists because it's saved to the stash.

### Generating a Stash ID

You might wonder what that `JJRNBOSKHpa` string is. It's a unique identifier for your script's stash.
You can use any string you like, but Rad provides a built-in command to generate collision-resistant IDs:

```shell
rad gen-id
```

<div class="result">
```
K7mPqR2xNfL
```
</div>

Using `rad gen-id` is recommended, especially if you plan to share your scripts.
It ensures your stash ID won't accidentally collide with another script's stash on someone else's machine.

### The Defer Save State Pattern

The example above demonstrates the recommended pattern for working with stashes:

```rad
state = load_state()
defer save_state(state)

// ... use and modify state ...
```

The [`defer`](./defer-errdefer.md) ensures your state is saved even if the script exits early or encounters an error.
This pattern is so common that you'll want to use it almost every time you work with stashes.

## State Storage

The primary way to use stashes is through **state** - a map that gets persisted as JSON.

### Loading State

[`load_state()`](../reference/functions.md#load_state) returns your script's saved state as a map.
If no state exists yet (first run), it returns an empty map `{}`:

```rad
state = load_state()
print(state)  // {} on first run, or previously saved data
```

### Saving State

[`save_state(map)`](../reference/functions.md#save_state) persists the map to disk:

```rad
state = {"username": "alice", "theme": "dark"}
save_state(state)
```

The state is stored as JSON at `~/.rad/stashes/<stash_id>/state.json`, making it easy to inspect or debug.

### Working with State

Since state is just a map, you can use all of Rad's map operations.
The [`get_default`](../reference/functions.md#get_default) function is particularly useful for initializing values:

```rad title="preferences.rad"
#!/usr/bin/env rad
---
Remembers user preferences.
@stash_id = K7mPqR2xNfL
---
args:
    set_editor str?  # Set preferred editor
    set_theme str?   # Set color theme

state = load_state()
defer save_state(state)

// Update preferences if provided
if set_editor:
    state["editor"] = set_editor
    print("Editor set to: {set_editor}")

if set_theme:
    state["theme"] = set_theme
    print("Theme set to: {set_theme}")

// Show current preferences
editor = state.get_default("editor", "vim")
theme = state.get_default("theme", "dark")
print("Current preferences: editor={editor}, theme={theme}")
```

```shell
rad preferences.rad
```

<div class="result">
```
Current preferences: editor=vim, theme=dark
```
</div>

```shell
rad preferences.rad --set-editor nano --set-theme light
```

<div class="result">
```
Editor set to: nano
Theme set to: light
Current preferences: editor=nano, theme=light
```
</div>

```shell
rad preferences.rad
```

<div class="result">
```
Current preferences: editor=nano, theme=light
```
</div>

## File Storage

Beyond state, stashes can also store arbitrary files. This is useful for caching larger data, storing user-created content, or managing configuration files.

### Writing Files

[`write_stash_file(path, content)`](../reference/functions.md#write_stash_file) writes a file to your stash:

```rad
write_stash_file("cache.json", '{"data": [1, 2, 3]}')
write_stash_file("logs/run.log", "Script executed at {now().time}")
```

Nested paths work automatically - Rad creates any necessary directories.

### Loading Files

[`load_stash_file(path, default)`](../reference/functions.md#load_stash_file) loads a file from your stash, creating it with the default content if it doesn't exist:

```rad
result = load_stash_file("config.txt", "# Default config\nkey=value")
print("Path: {result.full_path}")
print("Was just created: {result.created}")
print("Content: {result.content}")
```

The return value is a map containing:

- `full_path` - the absolute path to the file
- `created` - `true` if the file was just created, `false` if it already existed
- `content` - the file's contents

The `created` field is particularly useful for first-time setup:

```rad title="notes.rad"
#!/usr/bin/env rad
---
A simple notes manager.
@stash_id = N4xPmK8wQrT
---
args:
    add str?  # Add a new note

result = load_stash_file("notes.txt", "")

if result.created:
    print("Created new notes file!")

if add:
    // Append the new note
    new_content = result.content + add + "\n"
    write_stash_file("notes.txt", new_content)
    print("Note added!")
else:
    if result.content:
        print("Your notes:")
        print(result.content)
    else:
        print("No notes yet. Use --add to create one.")
```

### Getting the Stash Directory

[`get_stash_dir(subpath?)`](../reference/functions.md#get_stash_dir) returns the path to your stash directory:

```rad
stash_path = get_stash_dir()
print(stash_path)  // ~/.rad/stashes/<stash_id>

file_path = get_stash_dir("data/config.json")
print(file_path)  // ~/.rad/stashes/<stash_id>/data/config.json
```

This is useful when you need to work with stash files using other Rad functions like [`read_file`](../reference/functions.md#read_file) or [`get_path`](../reference/functions.md#get_path).

## Stash Structure

Your stash lives at `~/.rad/stashes/<stash_id>/` with this structure:

```
~/.rad/stashes/<stash_id>/
├── state.json          # Your state map (from save_state)
└── files/              # Your stash files
    ├── config.txt
    ├── cache.json
    └── logs/
        └── run.log
```

State and files are kept separate - `state.json` is managed by `load_state`/`save_state`, while the `files/` subdirectory is managed by `load_stash_file`/`write_stash_file`.

## Managing Stashes

Rad provides a built-in command to inspect and manage stashes for scripts on your PATH:

```shell
rad stash myscript --state   # View the script's state
rad stash myscript --id      # Show the stash ID
rad stash myscript --delete  # Delete the stash
```

!!! info "Scripts Must Be on PATH"

    The `rad stash` command looks up scripts on your PATH.
    For scripts not on your PATH, you can inspect the stash directly at `~/.rad/stashes/<stash_id>/`.

## Summary

- Stashes provide persistent, per-script storage that survives between runs.
- Declare a stash ID in the file header with `@stash_id = <id>`.
- Use `rad gen-id` to generate collision-resistant IDs.
- **State storage**: Use `load_state()` and `save_state(map)` for map-based data.
- **File storage**: Use `load_stash_file()` and `write_stash_file()` for arbitrary files.
- The recommended pattern is `state = load_state()` followed by `defer save_state(state)`.
- Stashes live at `~/.rad/stashes/<stash_id>/`.

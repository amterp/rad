# Script Commands

Relevant: [bundling](./bundling.md).

## 2025-05-25

Right now, you can write a script and easily make it take flags/args.

```
// script: add.rsl
#!/usr/bin/env rad
args:
    file string  # File to add.
    push p bool    # Enable to push after.
```

Rad will generate the help usage, and allow users to invoke this as e.g. `add.rsl book.txt true` or `add.rsl --file book.txt -p`.

A common use case, however (and it's getting more and more popular), is it have 'commands'. So rather than a script being

```
<script> <args>
myscript alice -a 30
```

You might instead want to partition your script into commands that do different things:

```
<script> <command> <args>
myscript greet alice -a 30
```

This *can* be accomplished with Rad at the moment, but not easily. You could do something like this:

```
args:
    command string
    name string
    age a int = 30
    location string?

    command enum ["greet", "admonish", "remember"]
    
    // the following conditional relational constraints don't currently exist but could
    command requires location if command == "remember"
```

but there are some downsides to this:

- Usage string generation doesn't really understand it's a *command*, and so doesn't generate super clearly.
- Conditional argument constraints get tricky, and don't currently exist
- If you have many disparate commands, the args block will likely get very complex as they're a union and may need to mutually exclude each other.

We should think of alternatives for supporting commands, ideally as a first-class citizen.

Some ideas to keep in mind:

- Chained commands? i.e. `add file`, `add dir`, etc.
- Shareability.
- Static interpretability. Rad needs to be able to read a script and statically resolve data relevant to usage, without running any of the script.
- File header for overall command, and/or file header for individual commands? Probably good to have at least a file header description for each command.

## Named args blocks

Allow naming several mutually exclusive args blocks, and then share the  

```
args greet:
    name string
    age a int = 30

args admonish:
    name string
    age a int = 30

args remember:
    name string
    age a int = 30
    location string
```

- How do we know which args block has actually been used? Perhaps define a boolean for the arg variable.

```
> myscript greet alice -a 30
'greet' is true
'admonish' is false
'remember' is false
```

- Common args between commands here are being repeated. Can we avoid that? Perhaps nested args?

```
args:
    name string
    age a int = 30
    args greet, admonish
    args remember:
        location string
```

That particular syntax looks a bit strange, especially `greet, admonish`. Not beginner-friendly or readable.

```
args:
    name string
    age a int = 30

command greet, admonish
command remember:
    location string
```

Here, the unnamed 'args' block is implicitly inherited by all the specified commands.

You could also keep the previous nesting syntax, if you wanted:

```
args:
    name string
    age a int = 30
    command greet, admonish
    command remember:
        location string
```

There's a relatively clean hierarchy here, though. What if `remember` didn't take an `age`?

```
args:
    name string

command greet, admonish:
    age a int = 30

command remember:
    location string
```

In all these examples, if you do `myscript -h`, Rad will compile a high-level usage only listing the available commands.

If you then do `myscript <command> -h`, it will 'zoom in' and print the generated usage for that specified command.

This model for how usage should work will generally apply to any solution we come up with in this thinking doc.

After the arg/command blocks are evaluated, the rest of the script shares the same logic. If the developer wants to
partition implementation based on command, they can just do if statements i.e.

```
if greet:
    // ... greet impl here

if admonish:
    // ... admonish impl here
    // 'else' not required because commands are mutually exclusive anyway.
```

## Concatenated scripts

The idea here is to treat individual commands as effectively different Rad scripts, and just concatenate them (to keep them in one script).

I guess we keep the named args block idea?

```
#!/usr/bin/env rad
---
Greet, admonish, or remember someone.
---

---
Greet someone.
---
args greet:
    name string
    age a int = 30

print("Hello fellow {age} year old, {name}!")

---
Admonish someone.
---
args admonish:
    name string
    age a int = 30

print("You make a lot of mistakes for a {age} year old, {name}!")

---
Remember someone.
---
args remember:
    name string
    age a int = 30
    location string

print("Ah, I remember you, you're the {age} year old from {location}, right? {name}?")
```

Here, we separate individual commands via the file headers.

I guess it's implied that the top file header is for the overall script, rather than an 'empty' command.

I don't actually think this is so bad. We do go back to repeating args quite a bit between commands, but maybe this is just not that bad for the Rad use case.

Maybe the 'command-level' file headers could use e.g. `===` instead of `---`. Or `--`.

This approach does slightly complicate some things. So far, we assume that, for any given script, if it's run, then it's run from top to bottom in its entirety.
With this approach here, we now need to know that are really have three *mostly* isolated sections that get interpreted in isolation. Could complicate LSP/validation.
Would it be possible to have a shared section? Especially for constants, etc.

What about when we implement [imports](./imports.md)? Would we be able to import between each sub-script?
Or should that be pulled into an external script and we import from there for each sub script? Or just have the shared section.

## Invoking other scripts

Rather than concatenating these scripts (or perhaps as an option), simply write individual scripts from individual commands, and then have a 'joiner' script which invokes the others.

```
// greet.rsl
---
Greet someone.
---
args:
    name string
    age a int = 30

print("Hi {name}")
```

```
// admonish.rsl
---
Admonish someone.
---
args:
    name string
    age a int = 30

print("Bad {name}!")
```

```
// remember.rsl
---
Remember someone.
---
args:
    name string
    age a int = 30

print("I remember you {name}")
```

So you have these three individual scripts. Then, as the entry script, you can have:

```
// myscript.rsl
commands:
    greet "greet.rsl"
    admonish "admonish.rsl"
    remember "remember.rsl"
```

This actually seems somewhat clean?

Syntax is `<command> <path>`.

The requirement to split your script is probably good for larger scripts, but can be annoying for small <60 line scripts that still are split into 2 or 3 commands.

Perhaps we can allow this 'commands' approach still, for a single file?
Kind of a combination with the [previous section](#concatenated-scripts) on concatenated single-file scripts.

## 2025-07-08

I tried the above `commands:` block syntax, and I don't think anyone is going to be happy with that.
It may seem clean for some scripts, that have truly completely isolated commands, but I faced a couple of issues on the two scripts I tried this syntax on:

`hm`

- I wanted `hm <script>` to be a valid invocation for the 'standard' approach, but current syntax doesn't support that -- everything has to be a command.
  - Could probably be solved via some 'default' case-kinda thing.

`dot`

- I had ~20 lines of common code I would need to either copy into each command script, or define and import into each (if we supported that). In any case, the import & run lines would need to be written.
  - Could perhaps allow writing some arbitrary code *before*

Let's go back to basics. What do we actually want/need from a command framework?

1. Ability to run some shared code first
2. Hard to mess up e.g. stash_id
3. Nested commands
4. Standard `arg block` functionality for each command
5. Command *optional* i.e. *can* write commands, but if no command, the inputs just go into a 'default handle'
6. Static inspectability -- usage string needs to be able to see what commands are available.
7. Nice-to-have: ability to implement all the commands in the same file.
    - Makes it much easier to share and manage, especially if the implementations are quite small.
    - Still ability to split as well, though.

Ideas that come to mind:

1. Import the separate command scripts
    - Invoke them as functions? Each parameter is named.
2. Pass functions as command-handlers.
    - Align the structure of function signatures with args (args maybe need to be a subset?)

1 and 2 are somewhat similar, by using functions. Maybe we could even allow both.

Function handlers

```
commands:
    add add # One liner add description.
    rm rm   # One liner rm description.

fn add(file str):
    <impl>

fn rm(file str):
    <impl>
```

Actually let's stop right there real quick. I don't know that this syntax is compatible enough with requirement #5 i.e. 'optional' part of commands.
We probably need to build this into the args block, actually. That way, you can say "can take a command as first string", otherwise do all these other args.
Can we fit it in somehow? Don't wanna be too verbose. Also want to make it completely optional, which is new for positional args in the arg block.

Another downside I identify with the above approach, just to note, is the loss of a file header. Maybe we can come up with some fn docstring equivalent?

```
commands:
    add add
    rm rm
args:
    file str

fn add(file str):
    """
    Here is a docstring for add.
    """
    <impl>

fn rm(file str):
    """
    Here is a docstring for rm.
    """
    <impl>
```

Here, we're saying that `add` and `rm` will be invoked as commands if matched, otherwise we'll use the `args` block.
But hmm, the fn approach is a *little* odd. We have these args blocks, can we reuse them? Somewhat like I previously suggested in this document.

```
command add:
    file str
args:
    file str
```

Hmm, but how do we separate the actual impl for `add`? Maybe still pass a handler fn somehow? One that doesn't take args? idk

## 2025-10-05

Revisiting commands with fresh perspective. Core requirements we'd like to satisfy:

1. Run shared code first
2. Hard to mess up (good scoping, clear semantics)
3. Nested commands (e.g. `git remote add`)
4. Standard `args` block functionality per command
5. **Commands should be optional** - tool can work with or without them
6. Static inspectability for help generation
7. Single-file friendly, multi-file capable

### Default Commands

Requirement #5 (optional commands) matters for tools like `hm` (a tldr/um hybrid):

```bash
hm grep              # Shows tldr page for grep (default behavior)
hm edit grep         # Edits the page
hm list              # Lists all pages
```

Here, `hm <page>` would be equivalent to `hm view <page>` - a default command that makes the common case ergonomic.

**One possible syntax:**

```rad
args:
    verbose bool  # Global to all commands

command default:
    ---
    Shows the page for a command
    ---
    page str

command edit:
    ---
    Edit a page
    ---
    page str

command list:
    ---
    List all pages
    ---

if default:
    show_page(page)
if edit:
    edit_page(page)
if list:
    list_pages()
```

**Potential semantics:**
- `command default:` would be optional
- If defined: `tool <args>` routes to default
- If not defined: `tool` (no command) → error listing available commands
- `default` becomes a boolean like other commands
- Help behavior: `tool -h` shows all commands; when default exists, also shows default args

This could handle the `hm grep` use case cleanly without special-casing.

### Nested vs Flattened Commands

Two possible syntactic styles:

**Style 1: Nested**
```rad
command remote:
    ---
    Manage git remotes
    ---
    timeout int = 30  # Shared by all subcommands
    command add:
        ---
        Add a remote
        ---
        name str
        url str
    command remove:
        ---
        Remove a remote
        ---
        name str
```

Advantage: Could define shared args for the namespace (`timeout` available to both `add` and `remove`).

**Style 2: Flattened**
```rad
command remote add:
    ---
    Add a remote
    ---
    name str
    url str
    timeout int = 30

command remote remove:
    ---
    Remove a remote
    ---
    name str
    timeout int = 30
```

Advantage: Less nesting, flatter structure. Could work well when subcommands don't share args.

**Possibly support both.** Nested when shared args in a namespace matter, flattened otherwise.

**Routing approach - separate booleans:** `command remote add:` could define two booleans: `remote` and `add`. Check via `if remote and add:`.

**Problem with separate booleans:** Easy to write logical bugs that are hard to catch:
```rad
command aaa bbb:
    pass
command ccc ddd:
    pass

if aaa and ddd:  // Nonsense check - aaa+bbb and ccc+ddd are different commands
    pass         // This silently never executes
```

The check `if aaa and ddd:` is logically impossible since commands are mutually exclusive, but it's not a syntax error. LSP can't easily help either.

**Alternative routing approaches to consider:**
- Single variable per command level: `if remote == "add":` or similar
- Structured command variable: `if command.remote.add:` or `if command == ["remote", "add"]:`
- Keep separate booleans but add LSP warnings for nonsensical combinations

**Max nesting depth:** Probably don't enforce a hard limit, though best practices would discourage more than 2-3 levels.

### Multi-File Commands

One initial idea: make command files independent:

```rad
# tool.rsl
command add "./commands/add.rsl"
command remove "./commands/remove.rsl"
```

Where each external file would be a standalone rad script that receives remaining CLI args.

**Problem with independence:** This likely doesn't match the actual use case. Command files probably aren't standalone programs - they're more likely **part of the same logical tool**. They might need to:
- Share scope with the root file (variables, functions)
- Access shared setup code run in the root
- Be invokable 99.9% of the time via the root script, not standalone

Making them independent could create friction: you'd have to duplicate shared code, can't easily share constants/helpers, etc.

**Complexity if command files share scope with root:**
- What code runs when? Root setup code would need to run before command code
- Are variables from root visible in command files?
- Can command files call functions defined in root?
- How would LSP reason about cross-file scope?
- This bleeds into imports/modules, which is a bigger design question

**Possible path:** Defer multi-file commands to v2. Focus on nailing single-file command syntax first.

**Rationale:**
- Single-file likely covers 90% of use cases
- Multi-file scope sharing is complex and ties into imports
- Better to understand real usage patterns before designing the multi-file story
- Could add multi-file cleanly in v2 without breaking single-file syntax

### Single-File Syntax Example

One possible approach for v1:

```rad
#!/usr/bin/env rad
---
Git helper tool
---

# Global args - available to ALL commands
args:
    verbose v bool
    config str = "~/.config"

# Default command (optional)
command default:
    ---
    Process a file
    ---
    file str

# Simple command
command edit:
    ---
    Edit a file
    ---
    file str
    interactive i bool

# Nested commands with shared args
command remote:
    ---
    Manage git remotes
    ---
    timeout int = 30
    command add:
        ---
        Add a remote
        ---
        name str
        url str
    command remove:
        ---
        Remove a remote
        ---
        name str

# Flattened alternative (could mix with nested, or enforce single style)
command deploy staging:
    ---
    Deploy to staging
    ---
    branch str = "main"

# Shared setup code (runs unconditionally before routing)
load_config(config)
if verbose:
    print("Verbose mode")

# Command routing (assuming separate boolean approach)
if default:
    process_file(file)

if edit:
    if interactive:
        $!`$EDITOR {file}`
    else:
        print(read_file(file))

if remote and add:
    // timeout, name, url all available
    add_remote(name, url, timeout)

if remote and remove:
    remove_remote(name, timeout)

if deploy and staging:
    deploy_to_staging(branch)
```

**Indentation observation:** Command implementations would become indented under `if` blocks. For small commands (2-3 lines), this might be fine. For larger ones, extracting to functions defined earlier in the file could keep routing flat and readable.

Example syntax:

```
command add:
    file str
    calls do_add

fn do_add():
    pass
```

### Conclusions

**Possible direction for v1:**
- Single-file command syntax with `command <name>:` blocks
- Global `args:` block for shared args
- `command default:` for optional default behavior when no command specified
- Nested syntax `command remote: command add:` for shared args
- Flattened syntax `command remote add:` as alternative
- Possibly allow mixing nested and flattened (or enforce single style)
- Command routing via some boolean/variable mechanism (still TBD due to concerns below)
- Shared code runs before routing

**For v2 (design later once v1 is understood):**
- Multi-file commands with `command add "./path.rsl"` or similar
- Scope sharing between root and command files
- Command file dependencies and imports
- Bundling multi-file commands into single file (see [bundling.md](./bundling.md))

**Open design questions:**
- **Command routing mechanism:** Separate booleans (`if remote and add:`) are simple but prone to logical bugs. Alternatives like `if command.remote.add:` or `if remote == "add":` might be safer but more verbose. LSP warnings could help with the boolean approach.
- **Docstring syntax:** Using `---` headers for command descriptions (like file headers) vs introducing new syntax
- **Function delegation:** Whether to add `command add: calls do_add` syntax, or just rely on manual delegation `if add: do_add()`
- **Nested/flattened mixing:** Allow both styles in same file, or enforce consistency?
- **Command file scope model:** How exactly should root and command files share scope when we implement multi-file?

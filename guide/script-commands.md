---
title: Script Commands
---

When you think of powerful CLI tools - `git`, `docker`, `kubectl` - they all share a common pattern: they're organized
around commands. You don't just run `git` with flags; you run `git commit`, `git push`, `git branch` - each a distinct
operation with its own arguments.

Rad lets you build tools like this through first-class command support. You can define multiple commands in a single
script, each with their own arguments and implementation.

## Basic Syntax

Let's start with a very simple example:

```rad linenums="1"
#!/usr/bin/env rad

command greet:
    name str
    calls greet_user

fn greet_user():
    print("Hello, {name}!")
```

This script defines a single command called `greet` that takes a `name` argument.

Invoke it by specifying the command name followed by its arguments:

```
> ./script.rad greet Alice
```

<div class="result">
```
Hello, Alice!
```
</div>

Let's break down the syntax:

1. **`command greet:`** - Defines a command named `greet`
2. **`name str`** - The command takes one required string argument called `name`
3. **`calls greet_user`** - Specifies which function to execute when this command runs
4. **`fn greet_user():`** - Defines the function that implements the command logic (defined after commands)

Command arguments (like `name`) become script-wide variables, accessible throughout your script.

!!! note "Underscores become hyphens"

    Command names with underscores are automatically converted to hyphens for CLI invocation.
    This matches standard CLI conventions (like `kubectl`, `docker-compose`).

    For example, `command deploy_staging:` is invoked as `deploy-staging`:

    ```
    > ./script.rad deploy-staging
    ```

    This is the same convention used for [argument names](./args.md).

## Multiple Commands

The power of commands emerges when you define several in one script. Let's create a simple deployment tool:

```rad linenums="1" hl_lines="3-6 8-11 13-16 18-20"
#!/usr/bin/env rad

command deploy:
    env str
    calls do_deploy

command status:
    env str
    calls do_status

fn do_deploy():
    print("Deploying to {env}...".yellow())
    print("Deployment complete!".green())

fn do_status():
    print("Checking status of {env}...".yellow())
    print("Environment {env} is healthy".green())
```

Now you can invoke either command:

```
> ./tool.rad deploy staging
```

<div class="result">
```
Deploying to staging...
Deployment complete!
```
</div>

```
> ./tool.rad status production
```

<div class="result">
```
Checking status of production...
Environment production is healthy
```
</div>

Each command has its own arguments and implementation, but they live in the same script and can share code.

## Adding Descriptions

Commands should include descriptions to make your tool self-documenting. Use the familiar `--- ... ---` header syntax:

```rad linenums="1" hl_lines="4-6 11-13"
#!/usr/bin/env rad

command deploy:
    ---
    Deploy the application to an environment
    ---
    env str
    calls do_deploy

command status:
    ---
    Check the health of an environment
    ---
    env str
    calls do_status

fn do_deploy():
    print("Deploying to {env}...".yellow())

fn do_status():
    print("Environment {env} is healthy".green())
```

These descriptions appear in the help output:

```
> ./tool.rad -h
```

<div class="result">
```
Usage:
  tool.rad [command] [OPTIONS]

Commands:
  deploy    Deploy the application to an environment
  status    Check the health of an environment
```
</div>

Notice how Rad automatically generates a usage string listing all available commands.

!!! tip "Multi-line descriptions"

    Just like script headers, command descriptions can span multiple lines:

    ```rad
    command deploy:
        ---
        Deploy the application to an environment.
        This will build, test, and deploy your application.
        ---
    ```

    **Important:** The first line appears in the script's overall help output, so keep it concise. Additional lines
    only appear when you request help for that specific command (`./tool.rad deploy -h`).

## Command Arguments

Each command can define its own arguments using the same syntax you learned in [Args](./args.md). Let's expand our
deployment tool:

```rad linenums="1" hl_lines="4-10"
#!/usr/bin/env rad

command deploy:
    ---
    Deploy the application to an environment
    ---
    env str              # Environment to deploy to
    branch str = "main"  # Branch to deploy from
    skip_tests bool      # Skip running tests before deploy
    calls do_deploy

fn do_deploy():
    if skip_tests:
        print("⚠️  Skipping tests".yellow())
    else:
        print("Running tests...".yellow())

    print("Deploying {branch} to {env}...".yellow())
    print("✅ Deployment complete!".green())
```

The arguments work exactly as they do in the `args:` block - you can use defaults, optional types, constraints, and
comments for help text.

Invoke with positional arguments:

```
> ./tool.rad deploy staging feature-branch
```

<div class="result">
```
Running tests...
Deploying feature-branch to staging...
✅ Deployment complete!
```
</div>

Or use flags (especially for booleans):

```
> ./tool.rad deploy --env=production --skip-tests
```

<div class="result">
```
⚠️  Skipping tests
Deploying main to production...
✅ Deployment complete!
```
</div>

## Shared Args

Often you want arguments that apply to *all* commands - like a `--verbose` flag or a `--config` path. Define these in
an `args:` block to share them across commands:

```rad linenums="1" hl_lines="3-5"
#!/usr/bin/env rad

args:
    verbose v bool   # Enable verbose output
    config str = "~/.config/tool.yaml"

command deploy:
    ---
    Deploy the application
    ---
    env str
    calls do_deploy

command status:
    ---
    Check environment status
    ---
    env str
    calls do_status

fn do_deploy():
    if verbose:
        print("Config: {config}".yellow())
        print("Deploying to {env}...".yellow())
    print("✅ Deployed!".green())

fn do_status():
    if verbose:
        print("Config: {config}".yellow())
        print("Checking {env}...".yellow())
    print("Environment healthy".green())
```

Shared args are available to all commands:

```
> ./tool.rad deploy staging --verbose
```

<div class="result">
```
Config: ~/.config/tool.yaml
Deploying to staging...
✅ Deployed!
```
</div>

```
> ./tool.rad status production --verbose
```

<div class="result">
```
Config: ~/.config/tool.yaml
Checking production...
Environment healthy
```
</div>

!!! note "Shared args are flag-only"

    When commands exist, shared args can only be passed as flags (like `--verbose`, `-v`, or `--config=value`), not
    positionally. This keeps the invocation clear: the first positional argument is always the command name.

    Both long form (`--verbose`) and short form (`-v`) work for shared args.

    Command-specific args can be positional or flags, just like regular script args.

## Command Callbacks

We've been using function references (`calls on_deploy`), which is the recommended approach for most commands.
However, for very short implementations, you can also use inline lambdas:

```rad linenums="1" hl_lines="3-8 10-17 19-21"
#!/usr/bin/env rad

command deploy:
    ---
    Deploy the application
    ---
    env str
    calls on_deploy

command rollback:
    ---
    Rollback a deployment
    ---
    env str
    calls fn():
        print("Rolling back {env}...".yellow())
        print("✅ Rollback complete!".green())

fn on_deploy():
    print("Deploying to {env}...".yellow())
    print("✅ Done!".green())
```

## Shared Logic

You can write code after all command blocks that runs before any callback is invoked. This is useful for setup logic that all commands need:

```rad linenums="1"
#!/usr/bin/env rad

command deploy:
    env str
    calls on_deploy

command rollback:
    env str
    calls on_rollback

// This runs before any callback
print("Initializing...".yellow())
config = read_file("config.yaml")
print("Config loaded".green())

fn on_deploy():
    // config is available here
    print("Deploying to {env} using config...")

fn on_rollback():
    // config is available here too
    print("Rolling back {env}...")
```

When you run `./script.rad deploy staging`, the flow is:

1. Parse arguments
2. Run shared logic (lines 12-14)
3. Run the callback (`on_deploy`)

This pattern is useful for loading configuration files, setting up connections, or validating preconditions that apply to all commands.

## Getting Help

Rad automatically generates help documentation for your commands. There are two levels of help:

**Script-level help** shows all available commands:

```
> ./tool.rad -h
```

<div class="result">
```
Usage:
  tool.rad [command] [OPTIONS]

Commands:
  deploy      Deploy the application
  rollback    Rollback a deployment
  status      Check environment status
```
</div>

**Command-level help** shows arguments for a specific command:

```
> ./tool.rad deploy -h
```

<div class="result">
```
Deploy the application

Usage:
  deploy <env> [branch] [OPTIONS]

Command args:
      --env str       Environment to deploy to
      --branch str    Branch to deploy from (default "main")
      --skip-tests    Skip running tests before deploy
  -v, --verbose       Enable verbose output
      --config str    (default "~/.config/tool.yaml")
```
</div>

Notice how the help includes:

- The command description
- Required and optional arguments
- Default values
- Shared args (like `--verbose` and `--config`)
- Help text from `#` comments

## Practical Example

Here's a concise, realistic example that demonstrates the "dev script" pattern - a common use case for replacing messy `Makefile`s or complex `package.json` script sections with a single, readable CLI entry point.

### Dev Script

Instead of remembering different commands for building, testing, and running your project, you can wrap them in a
single `dev.rad` script. This demonstrates shared arguments, boolean flags, and how to pass arguments down to underlying tools:

```rad linenums="1"
#!/usr/bin/env rad
---
Facilitates working with this repo's project.
---

args:
    verbose v bool   # Enable verbose output

command start:
    ---
    Start the local development server
    ---
    port int = 3000    # Port to listen on
    detach d bool      # Run in background
    calls on_start

command test:
    ---
    Run the test suite
    ---
    grep str?      # Filter tests by name
    watch w bool   # Re-run on file changes
    calls on_test

command build:
    ---
    Compile for production
    ---
    calls on_build

// Shared setup logic runs before any callback
if verbose:
    print("Checking project structure...".yellow())

if not path_exists("package.json"):
    print_err("Error: package.json not found".red())
    print_err("Run this script from the project root".yellow())
    exit(1)

fn on_start():
    print("🚀 Starting server on http://localhost:{port}...")

    cmd = "npm start -- --port {port}"

    if detach:
        $`{cmd} &`
        print("Server started in background".green())
    else:
        $`{cmd}`

fn on_test():
    opts = ""
    if watch:
        opts = "{opts} --watch"
    if grep: 
        opts = "{opts} -t '{grep}'"

    if verbose:
        print("Running: pytest {opts}".yellow())

    print("🧪 Running tests...")
    $`pytest {opts}` catch:
        print_err("Tests failed!".red())
        exit(1)

fn on_build():
    print("📦 Building for production...".yellow())

    $`rm -rf ./dist`
    $`npm run build` catch:
        print_err("Build failed".red())
        exit(1)

    print("✅ Build complete in ./dist".green())
```

**Usage:**

```
> ./dev.rad start
🚀 Starting server on http://localhost:3000...

> ./dev.rad start --port 8080 --detach
🚀 Starting server on http://localhost:8080...
Server started in background

> ./dev.rad test --grep "login_flow" --watch
🧪 Running tests...

> ./dev.rad build
📦 Building for production...
✅ Build complete in ./dist
```

Notice how this example uses:

- Shared args (`--verbose`) available to all commands
- Command-specific arguments with defaults (`port`, `detach`, `grep`, `watch`)
- Shared logic that runs before any callback
- Function references for callbacks (`calls on_start`, etc.)
- Integration with shell commands to wrap existing tools
- Clear, self-documenting help text

## Summary

- **Script commands** partition scripts into operations using `command name:` blocks
- Each command has:
    - Its own arguments (using standard `args` syntax)
    - A description block (`--- ... ---`)
    - A callback implementation (function reference or inline lambda)
- **Shared args** (from `args:` block) are available to all commands
    - Must be passed as flags when commands exist
- **Shared logic** runs before any callback. Write code after command blocks for setup that all commands need.
- **Help is automatic:**
    - `./script -h` lists available commands
    - `./script command -h` shows command-specific help
- **Callbacks:**
    - Function references: `calls function_name` (recommended)
    - Inline lambdas: `calls fn():` (for short implementations)
- **Use script commands to build CLI tools, not just scripts**

## Next

Rad provides a powerful system for looking up values from predefined resource files, which is particularly useful for
building interactive tools.

We'll explore this in the next section: [Resources](./resources.md).

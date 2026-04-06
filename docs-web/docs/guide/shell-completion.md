---
title: Shell Completion
---

Tab completion lets you press Tab in your terminal to auto-complete rad commands, flags, and argument values. You can enable it for the `rad` CLI itself and for your own Rad scripts.

## Quick Setup

Add one of these lines to your shell startup file:

=== "Bash (~/.bashrc)"

    ```shell
    eval "$(rad completion bash)"
    ```

=== "Zsh (~/.zshrc)"

    ```shell
    eval "$(rad completion zsh)"
    ```

After saving, either restart your terminal or re-source the file (e.g. `source ~/.bashrc`).

That's it - you now have tab completion for the `rad` CLI.

## What Gets Completed

With CLI completion enabled, pressing Tab after `rad` will suggest:

- **Embedded commands** - `new`, `docs`, `check`, `home`, `gen-id`, `stash`, `explain`, etc
- **Global flags** - `--help`, `--debug`, `--color`, `--quiet`, and the rest

For example, typing `rad ch` and pressing Tab completes to `rad check`. Typing `rad check --` and pressing Tab shows the flags that `check` accepts.

## Script Completions

You can generate tab completions for your own Rad scripts too - including their flags, enum values, commands, and command arguments.

To enable completion for a script, pass its path to `rad completion`:

=== "Bash"

    ```shell
    eval "$(rad completion bash ~/bin/deploy)"
    ```

=== "Zsh"

    ```shell
    eval "$(rad completion zsh ~/bin/deploy)"
    ```

### What Gets Completed

Given a script like this:

```rad title="deploy"
#!/usr/bin/env rad
---
Deploy services to a target environment.
---
args:
    service str       # Service to deploy.
    env e str = "dev" # Target environment.

    env enum ["dev", "staging", "prod"]

print("Deploying {service} to {env}...")
```

Tab completion will suggest:

- `--service` and `--env` (or `-e`) as flags
- `dev`, `staging`, `prod` as values when completing `--env`

If your script uses [commands](./script-commands.md), those are completed too - along with each command's own arguments.

!!! note "Shebang required"
    Scripts must have a `rad` shebang (e.g. `#!/usr/bin/env rad`) to be detected. Files without one are silently skipped.

## Multiple Scripts & Globs

You can pass multiple paths or glob patterns to register completions for many scripts at once:

=== "Bash"

    ```shell
    eval "$(rad completion bash ~/.rad/bin/* ~/scripts/*)"
    ```

=== "Zsh"

    ```shell
    eval "$(rad completion zsh ~/.rad/bin/* ~/scripts/*)"
    ```

Non-Rad files matched by the glob are silently skipped, so it's safe to point at directories containing a mix of file types.

!!! info "Separate lines for CLI and scripts"
    When you pass script paths, `rad completion` generates completions for those scripts only - not the rad CLI. Use a separate line without paths for rad CLI completions.

## Full Example

Here's a typical setup in `~/.bashrc` that covers everything:

=== "Bash (~/.bashrc)"

    ```shell
    # Rad tab completion
    eval "$(rad completion bash)"                    # rad CLI commands & flags
    eval "$(rad completion bash ~/.rad/bin/*)"        # all scripts in ~/.rad/bin/
    eval "$(rad completion bash ~/bin/deploy ~/bin/status)"  # specific scripts
    ```

=== "Zsh (~/.zshrc)"

    ```shell
    # Rad tab completion
    eval "$(rad completion zsh)"                    # rad CLI commands & flags
    eval "$(rad completion zsh ~/.rad/bin/*)"        # all scripts in ~/.rad/bin/
    eval "$(rad completion zsh ~/bin/deploy ~/bin/status)"  # specific scripts
    ```

!!! tip "Adding new scripts"
    When you add a new script to a directory that's already covered by a glob, re-source your shell config (or open a new terminal) to pick it up.

## Summary

- `rad completion bash` / `rad completion zsh` generates shell completion scripts.
- Without script paths, it completes the rad CLI (embedded commands, global flags).
- With script paths, it completes those scripts (flags, enum values, commands, command args).
- Scripts must have a `rad` shebang to be detected.
- Non-Rad files in glob expansions are silently skipped.
- Use separate `eval` lines for rad CLI and script completions.

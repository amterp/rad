# Rad Args Library Specification

## Overview

Rad Args is a Go library for CLI argument parsing that provides flexible positional and flag-based argument handling with support for subcommands, constraints, and various data types. It is designed to be user-friendly by default, with clear help messages and exit-on-error behavior.

## Core Architecture

### Command Structure

- **Cmd**: Central structure representing a command or subcommand. It holds flag definitions and subcommands.
- **Flags**: Named arguments that can be passed positionally or as flags.
- **Subcommands**: Nested commands with their own argument sets.
- **Global Flags**: Flags inherited by all subcommands.
- **ParseOpt**: A functional option for configuring parsing behavior.

### Flag Types

#### Non-Slice Types

- **BoolFlag**: Boolean values (true/false).
- **StringFlag**: String values with optional enum/regex constraints.
- **IntFlag**: Integer values with optional min/max constraints.
- **Int64Flag**: Int64 values with optional min/max constraints.
- **Float64Flag**: Float values with optional min/max constraints.

#### Slice Types

- **BoolSliceFlag**: Array of bools.
- **StringSliceFlag**: Array of strings.
- **IntSliceFlag**: Array of integers.
- **Int64SliceFlag**: Array of int64s.
- **Float64SliceFlag**: Array of float64s.

## Flag Configuration

### Common Properties

All flags support:
- **Name**: Primary identifier (required).
- **Short**: Single character short flag (optional).
- **Usage**: Help text description.
- **Default**: Default value when not specified.
- **Optional**: Whether the flag is required (default: false).
- **Hidden**: Whether the flag appears in any help output (default: false).
- **HiddenInLongHelp**: Whether the flag appears in the long help output (`--help`) (default: false).
- **PositionalOnly**: Flag can only be passed positionally.
- **FlagOnly**: Flag can only be passed as a named flag.

### Relational Constraints

- **Excludes**: Flags that cannot be used together with this flag. Works one-way like `Requires` - only the flag declaring the exclusion needs to specify the relationship.
- **Requires**: Flags that must be present when this flag is used.

### Value Constraints

- **EnumConstraint** (string): Restricts value to a specific set.
- **RegexConstraint** (string): Restricts value to match a regex pattern.
- **Min/Max** (numeric): Restricts value to a minimum or maximum.

### Slice Flag Options

- **Separator**: Character to split a single argument into multiple values.
- **Variadic**: Consume multiple consecutive arguments until the next flag.

## Argument Parsing

The library provides two primary methods for parsing arguments:

```go
// Parses args, printing usage and exiting on error.
func (c *Cmd) ParseOrExit(args []string, opts ...ParseOpt)

// Parses args, returning a ParseError on failure.
func (c *Cmd) ParseOrError(args []string, opts ...ParseOpt) *ParseError
```

### Parse Options

Parsing behavior can be customized using functional options (`ParseOpt`):

- **WithIgnoreUnknown(bool)**: If `true`, unknown flags and arguments are collected (retrievable via `GetUnknownArgs()`) instead of causing a parsing error.

### Positional Arguments

- All flags can be passed positionally unless marked as `FlagOnly`.
- Unless explicitly marked as `PositionalOnly`, all flags are also available as named flags (e.g., `--flag-name`).
- Positional assignment is left-to-right based on registration order.
- If a flag is set via a named flag, it's skipped in positional assignment.

### Flag Precedence

- Named flags override positional values.
- Later occurrences override earlier ones.

### Number Shorts Mode

- Activated when any IntFlag has a short name.
- Applies per-command (including inherited global flags).
- In this mode, standalone negative numbers are treated as short flags.
- To pass negative integers, use `--flag=-5` syntax.

### Slice Flag Parsing

- **Multiple occurrences**: `--flag value1 --flag value2` → `["value1", "value2"]`
- **Separator**: `--flag "value1,value2"` with separator "," → `["value1", "value2"]`
- **Variadic**: `--flag value1 value2` → `["value1", "value2"]` (stops at next flag)
- **Combined**: Variadic + separator processes both mechanisms.

### Bool Flag Clustering

- Multiple bool shorts can be clustered: `-abc` is equivalent to `-a -b -c`.
- A non-bool flag can terminate a cluster: `-abc value` (where `c` is non-bool).

## Commands and Subcommands

### Command Registration

```go
cmd := NewCmd("mycmd")
subCmd := NewCmd("subcmd")
invoked, err := cmd.RegisterCmd(subCmd)
```

### Subcommand Parsing

- Subcommands must appear immediately after the parent command.
- The first positional argument matching a subcommand name invokes that subcommand.
- If no match is found, parsing continues with the parent command's arguments.

### Global Flags

- Registered with `WithGlobal(true)` option.
- Automatically inherited by all subcommands.
- Global flags preserve their state across subcommand parsing.

## Configuration Options

### Command Options

- **SetDescription(string)**: Sets a description for the command, shown at the top of the help text.
- **SetCustomUsage(func(isLongHelp bool))**: Overrides the entire usage generation logic.
- **SetHelpEnabled(bool)**: Disables the automatic registration of `-h`/`--help` flags if set to `false`.
- **SetExcludeNameFromUsage(bool)**: Excludes the command name from usage output.

## Error Handling

The default behavior (`ParseOrExit`) is to print a user-friendly error and usage information to `stderr`, then exit with a non-zero status code. `ParseOrError` can be used to handle errors programmatically.

For testability, the library uses interface-based dependency injection for exit functionality and stderr writing, allowing for clean test mocking without race conditions.

### ParseError Structure

- Contains the error message (implementation details are private).
- The parser fails on the first error encountered.

### Error Types
- Unknown flag
- Missing required flag value
- Type conversion errors
- Constraint violations (enum, regex, min/max)

## State Management

### Parsing State

- **Configured(name)**: Returns `true` only if the user explicitly provided the flag.
- **GetUnknownArgs()**: Returns unrecognized arguments when `WithIgnoreUnknown(true)` is used.
- **used**: A per-command boolean indicating if a subcommand was invoked.

### Default Behavior

- If a flag has no default and is not marked `Optional`, parsing errors if it's not provided.
- If a default is set, `Configured(name)` returns `false` when the default value is used.
- Optional flags without defaults get the zero value of their type.

## Registration and Type Safety

### Flag Registration

```go
// Register with a returned pointer
flagPtr, err := NewString("name").Register(cmd)

// Register with an existing pointer
err := NewString("name").RegisterWithPtr(cmd, existingPtr)

// Register as a global flag
err := NewString("name").RegisterWithPtr(cmd, ptr, WithGlobal(true))
```

### Validation During Registration

- Flag names must be unique within a command.
- Default values must satisfy any defined constraints (enum, regex, min/max).
- Constraint violations return an error during registration.
- Global flag registration propagates to subcommands.

### Type Safety

- Generic `Flag[T]` and `SliceFlag[T]` structures ensure type safety.
- Registration validates that flag names are unique.

## Usage Generation

### Help Flags

When `helpEnabled` is `true` (the default), two help flags are automatically registered:
- `-h`: Triggers the "short help" output.
- `--help`: Triggers the "long help" output.

The only difference between short and long help is that flags marked `HiddenInLongHelp` are excluded from the long help output.

### Flag Ordering in Usage

Flags appear in usage output in the order they were registered, not alphabetically. This preserves the logical ordering that developers choose when defining their CLI interface.

### Default Usage Format

The generated usage string follows a structured format:

```
<description>

Usage:
  <synopsis>

Script args:
  <flags...>

Global options:
  <global flags...>
```

**Example Output:**
```
A rad-powered recreation of 'um', with the help of 'tldr'.
Allows you to check the tldr for commands, but then also
add your own notes and customize the notes in their own
entries.

Usage:
  hm <task> [OPTIONS]

Script args:
      --task str
  -e, --edit
  -l, --list        Lists stored entries. Exits after.
      --reconfigure Enable to reconfigure hm.

Global options:
  -d, --debug       Enables debug output. Intended for Rad script developers.
      --color str   Control output colorization. Valid values: [auto, always, never]. (default auto)
  -q, --quiet       Suppresses some output.
      --confirm-shell Confirm all shell commands before running them.
  -h, --help        Print usage string.
```

### Generating Usage Manually

The following methods can be used to manually generate usage strings:
- `GenerateShortUsage() string`
- `GenerateLongUsage() string`

### Custom Usage

- Override via `SetCustomUsage(func(isLongHelp bool))`. The boolean parameter indicates whether long help (`--help`) was requested.
- Inside the custom function, you can call `GenerateShortUsage`/`GenerateLongUsage` to build upon the default output. The `*Cmd` instance must be captured in a closure by the user if it's needed.

## Thread Safety and Re-parsing

### Limitations

- Not designed for concurrent access to same Cmd instance.
- Not designed for repeated parsing of same Cmd instance.

### Best Practices

- Create a new Cmd instance for each parse operation.
- Use separate Cmd instances for concurrent parsing.

## Misc

- The library must behave deterministically.

# Usage Generation Specification

This document specifies how rad-args should generate and display usage information, including colors, alignment, constraints, relationships, and all flag metadata.

## Overall Structure

```
[Description]

Usage:
  command [subcommand] <required-arg> [optional-arg] [OPTIONS]

[Commands section - if any subcommands]
[Arguments section - for all arguments (merged Arguments + Options)]
[Global options section - if applicable]
```

**Important**: Commands and Arguments are alternatives. If a subcommand is specified, it takes precedence and the main command's args become irrelevant - we dive into the subcommand's world.

**Key Paradigm**: All arguments are both flags and positional arguments unless explicitly overridden with `SetPositionalOnly(true)` or they are boolean flags (which are never positional).

**Global Flags**: Global flags registered on parent commands will also appear in subcommand usage generation.

**Customizable Headers**: Section headers ("Usage:", "Arguments:", "Commands:", "Global options:") should be customizable to allow different terminology.

## Help Modes

**Short Help vs Long Help:**
- Short help includes all sections but filters flags marked `HiddenInLongHelp`
- Long help includes all sections including flags marked `HiddenInLongHelp`
- Global options section appears in both short and long help (users should use `HiddenInLongHelp` if they want to hide global options from short help)
- Usage generation function: `GenerateUsage(isLongHelp bool)`

## Color Scheme

Using the `github.com/amterp/color` library:

- **Section Headers**: Green Bold (`GreenBoldF`) - "Usage:", "Arguments:", "Commands:", "Options:"
- **Command/Flag Names**: Bold (`BoldF`) - actual command and flag names
- **Argument Types**: Cyan (`CyanF`) - `<arg>`, `[arg]`, type indicators
- **Constraints/Metadata**: Yellow (`YellowF`) - ranges, enums, defaults
- **Descriptions**: Plain (`PlainF`) - flag and command descriptions

## Usage Line Format

```
Usage:
  command [subcommand] <required-arg> [optional-arg] [OPTIONS]
```

**Subcommands:**
- Show `[subcommand]` as generic placeholder if any subcommands exist
- Specific commands listed in Commands section below

**Arguments:**
- Required args without defaults: `<name>` in cyan
- Optional args (marked optional OR have defaults): `[name]` in cyan  
- Bool flags are never shown in usage line (not positional)
- Only positional-only args OR required flags appear in usage line
- Flags with `excludeNameFromUsage` set will not appear in synopsis but still appear in Arguments section
- Order: non-bool args first, then bools in Arguments section

**Options:**
- Always show `[OPTIONS]` at the end

## Commands Section

Only shown if there are subcommands. Always appears before Arguments section.

```
Commands:
  add                 Add a new item to the collection
  remove              Remove an item from the collection  
  list                List all items with optional filtering
```

**Format Rules:**
- Left-align command names
- Right-align descriptions with consistent spacing
- Preserve registration order
- Handle multi-line descriptions properly

## Arguments Section

Merged section showing all arguments (both positional and named flags).

```
Arguments:
  input-file str                Path to input file
  output-dir str                (optional) Output directory  
  -f, --format str              Output format. Default: json. Valid values: [json, yaml, xml]
      --timeout int             Request timeout. Default: 30. Range: [1, 300]
  -v, --verbose                 Enable verbose output. Excludes: quiet
```

**Format Rules:**
- Left-align argument names with proper indentation
- Show argument type after name (except for bools)
- Right-align descriptions with exactly 3 spaces from the longest left side (flag name + type)
- Boolean flags never show "(default false)" text
- Display order: Description. Default. Constraints. Relationships.
- Argument ordering: All arguments appear in registration order. The only exception is that non-positional arguments (flag-only args and bool flags) are moved to appear after positional arguments, but within each group (positional vs non-positional), registration order is preserved.
- Use plain argument names (no `<>` or `[]` brackets)
- Variadic optional flags show as `[files...]` format

## Flag Display Format

### Basic Flag Line Structure

```
  -s, --long-name type          Description. Default. Constraints. Relationships
      --long-only type          Description. Default. Constraints. Relationships
  -s                            Description. Default. Constraints. Relationships
```

**Display Order**: Description. Default. Constraints. Relationships.

### Alignment Rules
1. Calculate maximum width of the left side (flags + types + alignment buffer)
2. Use special alignment character (`\x00`) as placeholder
3. Replace with proper spacing to align all descriptions
4. Handle multi-line descriptions with consistent indentation

### Type Display

- **BoolFlag**: No type shown (just the flag)
- **StringFlag**: `str`  
- **IntFlag**: `int`
- **Int64Flag**: `int64`
- **Float64Flag**: `float`
- **SliceFlag[T]**: `T` for single values, `T...` for variadic (e.g., `strs`, `strs...`)

**Note**: Type names shown above match the current code implementation. The code uses abbreviated forms like `str` instead of `string`.

### Constraint Display

**Range Constraints:**
- Both bounds: `Range: [min, max]` (inclusive), `Range: (min, max)` (exclusive)
- Mixed: `Range: [min, max)`, `Range: (min, max]`
- Unbounded: `Range: [min, )`, `Range: (, max]`, `Range: (min, )`
- Examples:
  - `Range: [0, 100]`
  - `Range: (-20, )`  
  - `Range: (, 200.5]`
  - `Range: (10, 20)`

**Enum Constraints:**
- `Valid values: [option1, option2, option3]`
- Preserve input order from user

**Regex Constraints:**  
- `Must match pattern: ^[A-Z][a-z]*$`

**Slice Constraints:**
- `Separator: ","` (if custom separator specified)

**Variadic Flags:**
- All variadic flags are inherently optional (default to empty slices)
- `SetOptional()` and `SetDefault()` with empty slice have no effect on variadic flags
- Variadic flags always display as `[type...]` in synopsis (never `<type...>`)
- **Variadic Positional Precedence**: A variadic positional arg consumes all remaining positional arguments until a flag is encountered
- **Synopsis Rule**: No positional args appear in synopsis after the first variadic positional arg (they can only be set via flags)
- **Registration Error**: Registering a positional-only arg after a variadic arg is an error (impossible to set positionally)

### Default Values

- `Default: value`
- For slices: `Default: [item1, item2]`
- For bools: `Default: true/false` (only if explicitly set)

### Relationship Display

**Requires:**
- Single: `Requires: config`
- Multiple: `Requires: config, auth`

**Excludes:**
- Single: `Excludes: quiet`  
- Multiple: `Excludes: quiet, debug`

### Flag Status Markers

Flags display status markers based on their requirement and default status:

**No marker (most common)**:
- Flags with default values (the default fulfills the requirement)
- Boolean flags (implicit `false` default)
- Required positional-only flags (position indicates requirement)

**`(optional)` marker**:
- Flags explicitly marked optional with `SetOptional(true)` that have no default value
- Positional-only flags explicitly marked optional (important since users can't see flag names)
- Example: `--include strs    (optional) Include patterns`

**`(required)` marker**:
- Non-positional flags that are required and have no default value
- Example: `--name str    (required) Resource name`

**Boolean Flag Defaults**: Boolean flags with implicit or explicit `false` defaults don't show `(default false)` text. However, boolean flags explicitly set to `(default true)` will show this default value.

**Note**: The code currently defaults flags to optional, but this is incorrect behavior that needs to be fixed. Flags should default to required.

### Complete Examples

```
Arguments:
      --config str              Configuration file path. Default: ~/.app.conf
      --timeout int             Request timeout. Default: 30. Range: [1, 300]
      --retries int             (required) Retry attempts. Range: [0, )  
      --rate float              Rate limit. Default: 10.5. Range: (0, 100.5]
      --format str              Output format. Default: json. Valid values: [json, yaml, xml]
      --pattern str             (required) Name pattern. Must match pattern: ^[a-zA-Z][a-zA-Z0-9_]*$
      --tags strs...            (required) Resource tags. Separator: ",". Requires: config
      --include [strs...]       Include patterns
  -h, --help                    Print usage information
  -v, --verbose                 Enable verbose logging. Excludes: quiet
  -q, --quiet                   Suppress output. Excludes: verbose
```

## Special Cases

### Hidden Flags

- **Hidden**: Completely omitted from usage display. Specified via `SetHidden(true)`
- **HiddenInLongHelp**: Only shown in short help, omitted from long help. Specified via `SetHiddenInLongHelp(true)`

### Help Flag Behavior

The automatic help flag detection and processing should respect the `helpEnabled` setting:
- When `helpEnabled` is `true` (default): Automatic help flags are registered and help processing occurs
- When `helpEnabled` is `false`: No automatic help flag registration occurs, and custom help flags behave like normal flags without automatic usage printing and exit

### Positional-Only Flags

- Shown in Arguments section, not Options section
- Cannot have short names
- Specified via `SetPositionalOnly(true)`

### Flag-Only Flags

- Never shown as positional, always in Options
- Specified via `SetFlagOnly(true)`

### Counter/Incremental Int Flags

For int flags with short names that support counting behavior:
- Display type as `int` (no special indicator in usage)
- Can be invoked multiple times: `-vvv` sets value to 3
- Usage description should indicate counting behavior if applicable
- Example: `-v, --verbose int    Verbosity level (can be repeated)`

### Number Shorts Mode

When any int flag defines a short name, "number shorts mode" is activated:
- Affects parsing behavior: negative numbers must use `--flag=-5` syntax
- Does not change usage display format
- Users should be aware that `-1` will be interpreted as a short flag, not a negative number

## Error Handling

### Missing Required Arguments

Show which specific arguments are missing:
```
Error: Missing required arguments: <input-file>, <output-dir>

Usage:
  command <input-file> <output-dir> [OPTIONS]
...
```

### Invalid Values

Reference the constraint that was violated:
```
Error: Invalid 'timeout' value: 500 (Range: [1, 300])
Error: Invalid 'format' value: csv (Valid values: [json, yaml, xml])  
Error: Invalid 'name' value: 123abc (Must match pattern: ^[a-zA-Z][a-zA-Z0-9_]*$)
```

## Implementation Notes

### Features Requiring Implementation

The following features are specified but not yet implemented in the codebase:
- **Color integration** with `github.com/amterp/color` library
- **Constraint display** (ranges, enums, regex patterns) in usage output
- **Default value display** in usage descriptions
- **Relationship display** (requires/excludes) in usage descriptions  
- **Advanced alignment algorithm** with `\x00` placeholder characters
- **Optional flag indicators** `(optional)` markers
- **Slice separator display** in usage descriptions
- **Variadic slice formatting** `[type...]` display
- **Section header customization** functionality
- **Code fix**: Default flag behavior should be required, not optional

### Alignment Algorithm

1. First pass: calculate display width of each flag line up to description
2. Find maximum width  
3. Second pass: replace alignment character with appropriate spacing
4. Handle ANSI color codes in width calculations
5. Truncate long descriptions to prioritize full flag names

### Color Integration

- Use color functions consistently across all sections
- Provide option to disable colors (for pipes, files, etc.)
- Respect NO_COLOR environment variable

### Custom Usage Override

- `SetCustomUsage(func(isLongHelp bool))` hook to override entire usage generation
- Ability to request generated usage as string for building custom output

## Future Considerations

- Nested subcommand usage (should work automatically with Cmd-level generation)
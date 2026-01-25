---
title: epoch
---

## Preview

```rad linenums="1" hl_lines="0"
#!/usr/bin/env rad
---
Convert epoch timestamps to human-readable times across multiple timezones.
---
args:
    epoch int?  # Epoch time to convert. Defaults to current time if omitted.

if not epoch:
    epoch = now().epoch.millis
    print("Epoch millis: {epoch.yellow()}")

tz_to_flag = [
    ['Europe/London', '🇬🇧'],
    ['America/New_York', '🇺🇸'],
    ['Asia/Tokyo', '🇯🇵'],
    ['Australia/Melbourne', '🇦🇺'],
]

Timezone = tz_to_flag.map(fn(x) "{x[1]} {x[0]}")

Time = []
for tz, _ in tz_to_flag:
    time = parse_time(epoch, tz)
    Time += [time]

display:
    fields Timezone, Time

fn parse_time(epoch: int, tz: str) -> str:
    time = parse_epoch(epoch, tz=tz)
    return "{time.date} {time.time}"
```

```
> epoch -h
Convert epoch timestamps to human-readable times across multiple timezones.

Usage:
  epoch [epoch] [OPTIONS]

Script args:
      --epoch int   (optional) Epoch time to convert. Defaults to current time if omitted.
```

```
> epoch 1700000000000
Timezone                Time
🇬🇧 Europe/London         2023-11-14 22:13:20
🇺🇸 America/New_York      2023-11-14 17:13:20
🇯🇵 Asia/Tokyo            2023-11-15 07:13:20
🇦🇺 Australia/Melbourne   2023-11-15 09:13:20

> epoch
Epoch millis: 1769378597483
Timezone                Time
🇬🇧 Europe/London         2026-01-25 22:03:17
🇺🇸 America/New_York      2026-01-25 17:03:17
🇯🇵 Asia/Tokyo            2026-01-26 07:03:17
🇦🇺 Australia/Melbourne   2026-01-26 09:03:17
```

## Tutorial: Building `epoch`

### Motivation

When debugging distributed systems or reading logs, you'll often encounter epoch timestamps - large integers representing milliseconds (or seconds, or nanoseconds) since 1970. Converting these to human-readable times is tedious, and when you're coordinating across timezones, you often want to see the same moment in multiple locations at once.

Let's build a script that converts an epoch timestamp into a table showing that moment across several timezones.

### Writing the script

We can use `rad` to create the script file for us.

```sh
rad new epoch -s
```

This will set us up with an executable script named `epoch`, and the `-s` simplifies the template it's instantiated with to contain *just* a [shebang](../guide/getting-started.md#shebang).

The shebang will allow us to invoke the script as `epoch` from the CLI rather than writing out `rad ./epoch`. Open up `epoch` in your editor, and you should see something like this:

```rad linenums="1" hl_lines="0"
#!/usr/bin/env rad
```

Let's begin editing it. First, we want to quickly describe what the script is aiming to do, so we'll add a file header.

```rad linenums="1" hl_lines="2-4"
#!/usr/bin/env rad
---
Convert epoch timestamps to human-readable times across multiple timezones.
---
```

### Adding an optional argument

We want the script to accept an epoch timestamp, but it should also work without one - defaulting to the current time. We declare this with `int?`, where the `?` makes it optional (it will be `null` if not provided).

```rad linenums="1" hl_lines="5-6"
#!/usr/bin/env rad
---
Convert epoch timestamps to human-readable times across multiple timezones.
---
args:
    epoch int?  # Epoch time to convert. Defaults to current time if omitted.
```

Now we handle the case where no epoch was provided:

```rad linenums="1" hl_lines="8-10"
#!/usr/bin/env rad
---
Convert epoch timestamps to human-readable times across multiple timezones.
---
args:
    epoch int?  # Epoch time to convert. Defaults to current time if omitted.

if not epoch:
    epoch = now().epoch.millis
    print("Epoch millis: {epoch.yellow()}")
```

The [`now()`](../reference/functions.md#now) function returns a map with various time fields. We access `.epoch.millis` to get the current epoch in milliseconds, then print it so the user knows what value we're working with.

### Setting up our timezone data

We'll store our timezones alongside their flag emoji in a list of pairs. Each inner list contains `[timezone_id, flag]`:

```rad linenums="1" hl_lines="12-17"
#!/usr/bin/env rad
---
Convert epoch timestamps to human-readable times across multiple timezones.
---
args:
    epoch int?  # Epoch time to convert. Defaults to current time if omitted.

if not epoch:
    epoch = now().epoch.millis
    print("Epoch millis: {epoch.yellow()}")

tz_to_flag = [
    ['Europe/London', '🇬🇧'],
    ['America/New_York', '🇺🇸'],
    ['Asia/Tokyo', '🇯🇵'],
    ['Australia/Melbourne', '🇦🇺'],
]
```

### Building the display columns

For our table output, we need two columns: `Timezone` (for display) and `Time` (the converted times).

First, let's create the `Timezone` column. We want it to show the flag followed by the timezone name, like "🇬🇧 Europe/London". We'll use [`.map()`](../reference/functions.md#map) with a lambda to transform each pair:

```rad linenums="1" hl_lines="19"
#!/usr/bin/env rad
---
Convert epoch timestamps to human-readable times across multiple timezones.
---
args:
    epoch int?  # Epoch time to convert. Defaults to current time if omitted.

if not epoch:
    epoch = now().epoch.millis
    print("Epoch millis: {epoch.yellow()}")

tz_to_flag = [
    ['Europe/London', '🇬🇧'],
    ['America/New_York', '🇺🇸'],
    ['Asia/Tokyo', '🇯🇵'],
    ['Australia/Melbourne', '🇦🇺'],
]

Timezone = tz_to_flag.map(fn(x) "{x[1]} {x[0]}")
```

Each `x` is a `[timezone, flag]` pair, so `x[1]` is the flag and `x[0]` is the timezone name. We interpolate them into a string with the flag first.

!!! tip "List comprehension alternative"
    You could also write this as a list comprehension, which is a bit more concise in this case:
    ```rad
    Timezone = ["{x[1]} {x[0]}" for x in tz_to_flag]
    ```

### Iterating with unpacking

Now we need to build the `Time` column by converting the epoch for each timezone. When iterating over a list of lists, we can *unpack* each inner list directly into named variables:

```rad linenums="1" hl_lines="21-24"
#!/usr/bin/env rad
---
Convert epoch timestamps to human-readable times across multiple timezones.
---
args:
    epoch int?  # Epoch time to convert. Defaults to current time if omitted.

if not epoch:
    epoch = now().epoch.millis
    print("Epoch millis: {epoch.yellow()}")

tz_to_flag = [
    ['Europe/London', '🇬🇧'],
    ['America/New_York', '🇺🇸'],
    ['Asia/Tokyo', '🇯🇵'],
    ['Australia/Melbourne', '🇦🇺'],
]

Timezone = tz_to_flag.map(fn(x) "{x[1]} {x[0]}")

Time = []
for tz, _ in tz_to_flag:
    time = parse_time(epoch, tz)
    Time += [time]
```

The `for tz, _ in tz_to_flag:` line unpacks each `[timezone, flag]` pair. The first element becomes `tz`, and we use `_` for the second element to indicate we don't need the flag here - it's a convention meaning "discard this value."

This is cleaner than writing:

```rad
for pair in tz_to_flag:
    tz = pair[0]
    // ...
```

We're calling a `parse_time` function we haven't written yet - let's do that next.

### The helper function

We'll define a function that takes an epoch and timezone, then returns a formatted string:

```rad linenums="1" hl_lines="28-32"
#!/usr/bin/env rad
---
Convert epoch timestamps to human-readable times across multiple timezones.
---
args:
    epoch int?  # Epoch time to convert. Defaults to current time if omitted.

if not epoch:
    epoch = now().epoch.millis
    print("Epoch millis: {epoch.yellow()}")

tz_to_flag = [
    ['Europe/London', '🇬🇧'],
    ['America/New_York', '🇺🇸'],
    ['Asia/Tokyo', '🇯🇵'],
    ['Australia/Melbourne', '🇦🇺'],
]

Timezone = tz_to_flag.map(fn(x) "{x[1]} {x[0]}")

Time = []
for tz, _ in tz_to_flag:
    time = parse_time(epoch, tz)
    Time += [time]

// display block will go here

fn parse_time(epoch: int, tz: str) -> str:
    time = parse_epoch(epoch, tz=tz)
    return "{time.date} {time.time}"
```

The [`parse_epoch()`](../reference/functions.md#parse_epoch) function converts an epoch timestamp to a time map, accepting a `tz` parameter for the timezone. We then format the `.date` and `.time` fields into a string.

Note the type annotations `epoch: int, tz: str` and `-> str` - these are optional but help document what the function expects and returns.

### The display block

Finally, we use a [`display` block](../guide/rad-blocks.md#display-no-request) to render our two columns as a formatted table:

```rad linenums="1" hl_lines="26-27"
#!/usr/bin/env rad
---
Convert epoch timestamps to human-readable times across multiple timezones.
---
args:
    epoch int?  # Epoch time to convert. Defaults to current time if omitted.

if not epoch:
    epoch = now().epoch.millis
    print("Epoch millis: {epoch.yellow()}")

tz_to_flag = [
    ['Europe/London', '🇬🇧'],
    ['America/New_York', '🇺🇸'],
    ['Asia/Tokyo', '🇯🇵'],
    ['Australia/Melbourne', '🇦🇺'],
]

Timezone = tz_to_flag.map(fn(x) "{x[1]} {x[0]}")

Time = []
for tz, _ in tz_to_flag:
    time = parse_time(epoch, tz)
    Time += [time]

display:
    fields Timezone, Time

fn parse_time(epoch: int, tz: str) -> str:
    time = parse_epoch(epoch, tz=tz)
    return "{time.date} {time.time}"
```

The `display` block takes our pre-populated lists and renders them as aligned columns. The variable names (`Timezone`, `Time`) become the column headers.

### Try it out

```
> epoch 1700000000000
Timezone                Time
🇬🇧 Europe/London         2023-11-14 22:13:20
🇺🇸 America/New_York      2023-11-14 17:13:20
🇯🇵 Asia/Tokyo            2023-11-15 07:13:20
🇦🇺 Australia/Melbourne   2023-11-15 09:13:20
```

Or without an argument, to see the current time:

```
> epoch
Epoch millis: 1769378597483
Timezone                Time
🇬🇧 Europe/London         2026-01-25 22:03:17
🇺🇸 America/New_York      2026-01-25 17:03:17
🇯🇵 Asia/Tokyo            2026-01-26 07:03:17
🇦🇺 Australia/Melbourne   2026-01-26 09:03:17
```

## Concepts demonstrated

| Concept | Where |
|---------|-------|
| [Optional arguments](../guide/args.md) | `epoch int?` |
| Null checking | `if not epoch:` |
| [Built-in `now()`](../reference/functions.md#now) | `now().epoch.millis` |
| [List of lists](../guide/basics.md#list) | `tz_to_flag` |
| [`.map()` with lambdas](../reference/functions.md#map) | `tz_to_flag.map(fn(x) ...)` |
| [Loop unpacking](../guide/basics.md#for-loops) | `for tz, _ in tz_to_flag:` |
| [Functions with types](../guide/functions.md) | `fn parse_time(epoch: int, tz: str) -> str:` |
| [`parse_epoch()`](../reference/functions.md#parse_epoch) | Time conversion |
| [`display` blocks](../guide/rad-blocks.md#display-no-request) | Tabular output |

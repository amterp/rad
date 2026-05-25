# pick_from_resource

Loads options from a resource file and presents an interactive menu.

## Signature

`pick_from_resource(path: str, _filter: str?, *, prompt: str = "Pick an option", prefer_exact: bool = true) -> any`

## Examples

```rad
pick_from_resource("servers.json")                    // -> Menu from file
pick_from_resource("configs.json", "prod")            // -> Pre-filtered, exact match priority
pick_from_resource("data.json", prompt="Select:")     // -> Custom prompt
pick_from_resource("data.json", "x", prefer_exact=false) // -> Pure fuzzy matching
```

## Category

io

## Notes

Loads data from a JSON file and presents it as selectable options. Returns the selected item(s).

With `prefer_exact=true` (the default), exact key matches (case-insensitive) are prioritized: if exactly one entry has a
key that exactly matches the filter, it's selected immediately; if multiple match exactly, only those are shown. Set
`prefer_exact=false` to disable this and use pure fuzzy matching.

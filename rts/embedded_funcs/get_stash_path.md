# get_stash_path

Returns the full path to the script's stash directory, with the given subpath if specified.

Requires a stash ID to have been defined.

[//]: # (TODO link to stash id docs, and for below)


**Return Values**

- Without subpath defined: `<rad home>/stashes/<stash id>`
- With subpath defined: `<rad home>/stashes/<stash id>/<subpath>`

## Signature

`get_stash_path(_sub_path: str = "") -> error|str`

## Examples

```rad
root = get_stash_path()                // -> "<rad-home>/stashes/<stash-id>"
cache = get_stash_path("cache.json")   // -> "<rad-home>/stashes/<stash-id>/cache.json"
```

## Category

stash

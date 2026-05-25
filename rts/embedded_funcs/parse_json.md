# parse_json

Parses a JSON string into Rad data structures.

## Signature

`parse_json(_str: str) -> any|error`

## Examples

```rad
parse_json(r'{"name": "Alice", "age": 30}')  // -> {"name": "Alice", "age": 30}
parse_json('[1, 2, 3]')                      // -> [1, 2, 3]
parse_json('invalid json')                   // -> Error: invalid JSON
```

## Category

parsing

## Notes

Use a raw string (`r'...'`) when the JSON contains `{` or `}` - plain
single- and double-quoted strings interpolate `{expr}`, which makes
JSON literals trip the interpolator. Raw strings are also natural for
JSON pasted verbatim from a sample.

# print

Writes its arguments to stdout, separated by a space, followed by a
newline. The default workhorse for output. For error output, use
`print_err`; for structured pretty-printing, use `pprint`.

## Signature

`print(*_items: any, *, sep: str = " ", end: str = "\n") -> void`

## Parameters

- `_items` (variadic `any`): values to print. Each is converted to
  its string form via the default formatter.
- `sep` (`str`, keyword-only, default `" "`): separator inserted
  between consecutive items.
- `end` (`str`, keyword-only, default `"\n"`): trailing text after
  the last item. Set to `""` to print without a newline.

## Examples

```rad
print("hello", "world")    // -> hello world
print(1, 2, 3, sep=", ")   // -> 1, 2, 3
print("no newline", end="") // -> no newline
```

## Category

io

## See also

`print_err`, `pprint`, `debug`

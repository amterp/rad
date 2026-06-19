# print

Writes its arguments to stdout, separated by a space, followed by a
newline. The default workhorse for output. For error output, use
`print_err`; for structured pretty-printing, use `pprint`.

```rad
print(*_items: any, *, sep: str = " ", end: str = "\n") -> void
```

```rad
print("hello", "world")    // -> hello world
print(1, 2, 3, sep=", ")   // -> 1, 2, 3
print("no newline", end="") // -> no newline
```

## Parameters

- `_items` (`variadic `any`): values to print. Each is converted to
- `sep` (`str`, keyword-only, default `" "`): separator inserted
- `end` (`str`, keyword-only, default `"\n"`): trailing text after

## See also

`print_err`, `pprint`, `debug`

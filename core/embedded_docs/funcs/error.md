# error

Creates an error object with the given message.

```rad
error(_msg: str) -> error
```

```rad
fn validate(x: int):
    if x < 0:
        return error("Something went wrong")
    return x

result = validate(-1)  // -> Script will exit with this error message
```

## Notes

`return` at the top level isn't legal Rad - wrap it in a `fn` and
return the error from there, or assign it and propagate via `??` /
`catch:`.

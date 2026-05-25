# exit

Exits the script with the given exit code.

## Signature

`exit(_code: int|bool = 0) -> void`

## Examples

```rad
exit()          // -> Exits with code 0
exit(1)         // -> Exits with code 1
exit(true)      // -> Exits with code 1 (bool conversion)
exit(false)     // -> Exits with code 0 (bool conversion)
```

## Category

system

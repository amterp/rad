# confirm

Gets a boolean confirmation from the user (y/n prompt). Accepts "y", "yes", or Enter (empty input) as
confirmation.

## Signature

`confirm(prompt: str = "Confirm? [Y/n] > ") -> error|bool`

## Examples

```rad
if confirm():                        // -> Uses default "Confirm? [Y/n] > " prompt
    print("Confirmed!")

if confirm("Delete file? [Y/n] "):   // -> Custom prompt
    print("File deleted")
```

## Category

io

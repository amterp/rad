# confirm

Gets a boolean confirmation from the user (y/n prompt). Accepts "y", "yes", or Enter (empty input) as
confirmation.

```rad
confirm(prompt: str = "Confirm? [Y/n] > ") -> error|bool
```

```rad
if confirm():                        // -> Uses default "Confirm? [Y/n] > " prompt
    print("Confirmed!")

if confirm("Delete file? [Y/n] "):   // -> Custom prompt
    print("File deleted")
```

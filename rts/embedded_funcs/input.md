# input

Gets a line of text input from the user with optional prompt, default, hint, and secret mode.

## Signature

`input(prompt: str = "> ", *, hint: str = "", default: str = "", secret: bool = false) -> error|str`

## Examples

```rad
// Basic input
name = input("What's your name? ")                    // -> Prompts and waits for input

// With default value
color = input("Favorite color? ", default="blue")     // -> Returns "blue" if user presses enter

// With hint text
email = input("Email: ", hint="user@example.com")     // -> Shows placeholder text

// Hidden input for passwords
password = input("Password: ", secret=true)           // -> Hides typed characters
```

## Category

io

## Notes

**Parameters:**

| Parameter | Type           | Description                                  |
|-----------|----------------|----------------------------------------------|
| `prompt`  | `str = "> "`   | The text prompt to display to the user       |
| `hint`    | `str = ""`     | Placeholder text shown in input field        |
| `default` | `str = ""`     | Default value if user doesn't enter anything |
| `secret`  | `bool = false` | If true, hides input (useful for passwords)  |

If `secret` is true, input is hidden (useful for passwords). The `hint` parameter has no effect when `secret` is
enabled.

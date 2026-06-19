# hyperlink

Creates a clickable hyperlink in supporting terminals.

```rad
hyperlink(_val: any, _link: str) -> str
```

```rad
hyperlink("Visit Google", "https://google.com")    // -> Clickable "Visit Google" link
hyperlink("localhost", "http://localhost:3000")    // -> Clickable "localhost" link
hyperlink(42, "https://example.com")               // -> Clickable "42" link
```

## Notes

Converts text into a terminal hyperlink that can be clicked in supported terminals.

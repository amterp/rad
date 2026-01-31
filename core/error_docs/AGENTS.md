# Error Documentation Guidelines

These docs are displayed when users run `rad explain RAD<code>`. They're the second layer of help - the inline error message is first, these docs go deeper.

## Audience

Beginners learning Rad, or agents trying to fix code in an unfamiliar language. Write for someone who hit this error and wants to understand what went wrong and how to fix it.

## Tone

Match Rad's documentation voice: friendly, practical, concise. Not chatty, not sterile. Every sentence earns its place.

Good: "The parser expected a closing parenthesis but didn't find one."
Bad: "This error occurs when there is an issue with parenthesis matching in your code, which can happen for various reasons."

## Structure

No rigid template. Include what's useful, omit what isn't.

**Always include:**
- Title: `# RAD{CODE}: Brief Title`
- Opening paragraph explaining what went wrong (1-2 sentences)
- At least one example
- How to fix it

**Include when useful:**
- Common causes (if the error has multiple distinct triggers)
- Reference tables (type conversions, format specifiers, etc.)
- Philosophy or "why Rad works this way" (if it aids understanding)
- Notes about related functions or features

**Don't include just to fill space:**
- "Common Causes" for self-explanatory errors
- Multiple examples showing the same thing
- Sections that restate what the title already says

## Examples

Examples are the backbone. Show the failure, show the fix.

```rad
# Wrong
print("hello"

# Correct
print("hello")
```

One good example beats three okay ones. Add more only if they show genuinely different scenarios.

Use realistic code when possible - something someone might actually write.

## Error Handling Patterns

When showing how to handle errors, use Rad's idioms consistently:

```rad
# Simple fallback
age = parse_int(input) ?? 0

# With logging
data = read_file(path) catch:
    print_err("Could not read file: {data}")
    exit(1)

# Ignore errors explicitly
delete_path(temp) catch:
    pass
```

## Accuracy

Every function, method, or pattern you mention must actually exist in Rad. Agents will copy-paste these examples. If you're unsure whether something exists, check the reference docs or codebase.

## Length

Match depth to confusion:
- "Missing parenthesis" → A few sentences, one example
- "Type mismatch in function returns" → More explanation, maybe reference tables
- Catch-all errors (RAD10001, RAD20000) → Acknowledge they're catch-alls, suggest filing issues if the error seems like it deserves a specific code

## Catch-All Errors

Some errors are genuinely generic. Be honest about it:

> This is a catch-all error for [category]. The error message itself describes the specific problem. If you think this error should have a more specific code, consider filing an issue.

## Teaching

These docs can teach Rad concepts when it's natural. If explaining "why" helps the user understand and avoid the error in future, include it. Don't force philosophy where it doesn't fit.

Good (RAD40001 - Scientific Notation):
> Scientific notation like `1e3` (meaning 1×10³ = 1000) can represent very large or very small numbers. When used where an integer is expected, the result must be a whole number.

This explains the concept because it's not obvious. For "missing parenthesis," no such explanation is needed.

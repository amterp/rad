# Things to Revisit

- regex strings need to be more explicitly opted-into (e.g. for split func)
- `quiet` being how we suppress announcements for shell commands
- body should be a named arg in http methods?
- should the int/float etc parsing funcs be prefixed with `parse_`?
- one `http` func for all methods instead of some having their own?
- `enum name` vs. `name enum` ordering for arg constraints
- `{}` vs. `${}` for interpolation
- Ruby-style % syntax or bash heredocs/herestrings?
- Division by 0 errors or returns nan?

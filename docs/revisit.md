# Things to Revisit

- regex strings need to be more explicitly opted-into (e.g. for split func)
- `quiet` being how we suppress announcements for shell commands
- body should be a named arg in http methods?
- one `http` func for all methods instead of some having their own?
- `{}` vs. `${}` for interpolation
- Ruby-style % syntax or bash heredocs/herestrings?
- Division by 0 errors or returns nan?
- `del a[1], a[1]` on a list - currently deletes two different items, but should perhaps be atomic and "delete the same one" 

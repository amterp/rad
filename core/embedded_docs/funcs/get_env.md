# get_env

Retrieves the value of an environment variable.

```rad
get_env(_var: str) -> str
```

```rad
home_dir = get_env("HOME")                    // -> "/Users/username"
api_key = get_env("API_KEY") or "default"     // -> Uses default if not set
missing = get_env("NONEXISTENT")              // -> ""
```

## Notes

Returns the environment variable value, or empty string if not set.

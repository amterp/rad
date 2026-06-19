# get_rad_home

Returns the path to rad's home folder on the user's machine.

**Return Values**

Defaults to `$HOME/.rad`, or `$RAD_HOME` if it's defined.

```rad
get_rad_home() -> str
```

```rad
home = get_rad_home()              // -> "/Users/me/.rad" (or $RAD_HOME)
```

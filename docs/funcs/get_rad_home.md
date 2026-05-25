# get_rad_home

Returns the path to rad's home folder on the user's machine.


**Return Values**

Defaults to `$HOME/.rad`, or `$RAD_HOME` if it's defined.

## Signature

`get_rad_home() -> str`

## Examples

```rad
home = get_rad_home()              // -> "/Users/me/.rad" (or $RAD_HOME)
```

## Category

stash

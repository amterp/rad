# delete_path

Deletes a file or directory at the specified path.

## Signature

`delete_path(_path: str) -> bool`

## Examples

```rad
delete_path("temp.txt")         // -> true (if file existed and was deleted)
delete_path("missing.txt")      // -> false (file didn't exist)
delete_path("directory/")       // -> true (if directory existed and was deleted)
```

## Category

io

## Notes

Returns `true` if the path was successfully deleted, `false` if it didn't exist or couldn't be deleted.

A leading `~` in `_path` is expanded to your home directory.

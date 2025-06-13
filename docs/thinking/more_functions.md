# More Functions

## 2025-06-13

```
make_dir:
Signature idea: fn make_dir(_path: str, *, recursive: bool = true, mode: int?) -> error?
Rationale: Creating directories, especially recursively (mkdir -p). mode for permissions.
```

```
remove_dir:
Signature idea: fn remove_dir(_path: str, *, recursive: bool = false) -> error?
Rationale: Removing directories. recursive is key (rm -rf danger, so be careful with defaults or force a flag). Your delete_path might cover this if depth implies recursion on directories, but a dedicated remove_dir is often clearer.
```


```
copy_path:
Signature idea: fn copy_path(_source: str, _destination: str, *, recursive: bool = false) -> error?
Rationale: Copying files or directories (cp, cp -r).
```

```
move_path:
Signature idea: fn move_path(_source: str, _destination: str) -> error?
Rationale: Moving/renaming files or directories (mv).
```

```
current_dir:
Signature idea: fn current_dir() -> str
Rationale: Get current working directory (pwd).
```

```
change_dir:
Signature idea: fn change_dir(_path: str) -> error?
Rationale: Change current working directory (cd).
```

```
is_file, is_dir, is_link:
Signature ideas:
fn is_file(_path: str) -> bool
fn is_dir(_path: str) -> bool
fn is_link(_path: str) -> bool
Rationale: Simpler boolean checks than parsing get_path().type. Often used in scripts. get_path is good for more details, these are for quick checks.
```

```
path_join:
Signature idea: fn path_join(_base: str, *parts: str) -> str
Rationale: Platform-agnostic path joining (like Python's os.path.join). Essential for constructing paths reliably.
```

```
path_basename, path_dirname, path_splitext:
Signature ideas:
fn path_basename(_path: str) -> str
fn path_dirname(_path: str) -> str
fn path_splitext(_path: str) -> [str, str] (root, ext)
Rationale: Common path manipulations. get_path returns base_name, but dirname and splitext are also very useful.
```

```
temp_file, temp_dir:
Signature ideas:
fn temp_file(*, prefix: str?, suffix: str?, dir: str?) -> str|error (returns path to a new temporary file)
fn temp_dir(*, prefix: str?, suffix: str?, dir: str?) -> str|error (returns path to a new temporary directory)
Rationale: Scripting often requires temporary files/directories.
```

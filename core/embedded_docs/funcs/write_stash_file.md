# write_stash_file

Writes content to a file in the script's stash directory.

```rad
write_stash_file(_path: str, _content: str) -> error?
```

```rad
write_stash_file("log.txt", "Script executed at " + now().time)
write_stash_file("data/results.json", json_data)
print("Data saved to stash")
```

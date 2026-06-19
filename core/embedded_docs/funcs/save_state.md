# save_state

Saves the script's state to persistent stash storage.

```rad
save_state(_state: map) -> error?
```

```rad
state = {"counter": 42, "last_run": now().date}
save_state(state)
print("State saved")
```

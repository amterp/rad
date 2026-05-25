# load_state

Loads the script's stashed state. Creates it if it doesn't already exist.

Requires a stash ID to have been defined.


**Return Values**

1. `map` containing the saved state. Starts empty, before anything is saved to it.
2. `bool` representing if the state existed before the load, or if it was just created.

## Signature

`load_state() -> error|map`

## Examples

```rad
state = load_state()                   // -> map containing previous state
state["count"] = (state["count"] or 0) + 1
save_state(state)
```

## Category

stash

# load

Loads a value into a map using lazy evaluation. If key exists, returns cached value; otherwise runs loader function.

```rad
load(_map: map, _key: any, _loader: fn() -> any, *, reload: bool = false, override: any?) -> error|any
```

```rad
cache = {}
load(cache, "data", fn() expensive_calculation())    // -> Runs loader, caches result
load(cache, "data", fn() expensive_calculation())    // -> Returns cached value

// Force reload
load(cache, "data", fn() new_calculation(), reload=true)

// Override with specific value  
load(cache, "data", fn() ignored(), override="forced")
```

## Notes

**Parameters:**

| Parameter  | Type           | Description                              |
| ---------- | -------------- | ---------------------------------------- |
| `_map`     | `map`          | Map to store/retrieve cached values      |
| `_key`     | `any`          | Key to lookup in the map                 |
| `_loader`  | `fn() -> any`  | Function to call if key doesn't exist    |
| `reload`   | `bool = false` | Force reload even if key exists          |
| `override` | `any?`         | Use this value instead of calling loader |

If key doesn't exist, `_loader` is called and result is cached. Cannot use `reload=true` with `override` (mutually
exclusive).

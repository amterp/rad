package rl

// RemovedFuncHints maps the name of a removed or renamed builtin to a migration
// hint. Shared by the runtime (when a call to one of these names fails at
// execution) and the static checker (so `rad check` surfaces the same guidance
// instead of only a generic "did you mean" suggestion). Read from here in both
// places rather than re-hardcoding the strings, so the two never drift.
var RemovedFuncHints = map[string]string{
	"get_default":   `get_default was removed. Use: map["key"] ?? default. See: https://amterp.dev/rad/migrations/v0.8/`,
	"get_stash_dir": "get_stash_dir was renamed to get_stash_path. See: https://amterp.dev/rad/migrations/v0.9/",
}

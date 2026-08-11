package rl

// Hints shared between the static checker (rts/check) and the interpreter
// (core). Both report the same situations - one before the script runs, one
// while it runs - and a reader who sees different advice from `rad check` and
// `rad <script>` for the same line has to work out which one to trust.
//
// They lived as separate literals on both sides until the wordings drifted.
// rl is the lowest package both import, so it is where they can only be
// written once.
//
// Keep each to one imperative sentence, per core/error_docs/AGENTS.md.
const (
	// HintStrPlusMigration fires when a script adds a string to a non-string,
	// which worked before v0.9.
	HintStrPlusMigration = "Use string interpolation - + no longer coerces types since v0.9. " +
		"See https://amterp.dev/rad/migrations/v0.9/"

	// HintListAppend fires when a script adds a non-list to a list, which is
	// almost always a missed `+ [item]`.
	HintListAppend = "To append, wrap the right side in a list - e.g. `myList + [item]`."

	// HintForLoopMigration fires when a script unpacks a loop variable into an
	// index and a value, which was the pre-v0.7 for-loop shape.
	HintForLoopMigration = "Use `for item in items with loop` and read `loop.idx`. " +
		"See https://amterp.dev/rad/migrations/v0.7/"
)

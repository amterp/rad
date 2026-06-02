package testing

// Tolerance describes the diagnostics we allow (or require) for a single
// doc snippet. The default for any snippet not in docSnippetTolerances is
// "must produce zero diagnostics of any severity".
//
// Snippets are identified by "<relative-path-from-repo-root>#<8-hex-content-hash>"
// e.g. "docs-web/docs/guide/error-handling.md#a1b2c3d4". The hash changes
// whenever the snippet body changes, which orphans any matching tolerance
// entry and forces re-review on the next run.
type Tolerance struct {
	// Skip bypasses the check entirely. Use for fragments / pseudo-code
	// that can't stand alone.
	Skip bool

	// MaxSeverity allows diagnostics with severity at or below this level
	// (case-insensitive: "hint" | "warning" | "info" | "error"). Empty means
	// no severity-based tolerance.
	//
	// Severity ordering (from rts/check/structs.go): Hint < Warning < Info < Error.
	MaxSeverity string

	// ExpectedCodes is a strict set of RAD codes (e.g. "RAD30001") that
	// MUST appear in the snippet's diagnostics. Any diagnostic with a code
	// outside this set fails (subject to MaxSeverity). Missing expected
	// codes also fail.
	ExpectedCodes []string

	// Reason is mandatory: a human-readable note explaining why this
	// entry exists. Test fails if an entry has Reason == "".
	Reason string
}

// globallyToleratedCodes are advisory diagnostics accepted in ANY doc
// snippet, without a per-snippet entry. RAD30011 (unhandled fallible
// call) is a pervasive teaching-style hint: docs and examples call
// parse_int / read_file / etc. terse by design, to keep the focus on
// the feature being shown rather than on error scaffolding. Pinning it
// per-snippet would add dozens of no-signal entries. It's a Hint and
// never gates execution, so tolerating it globally hides no real error
// - the snippets still run.
var globallyToleratedCodes = map[string]bool{
	"RAD30011": true,
}

// docSnippetTolerances maps snippet IDs to their accepted diagnostic profile.
// Add entries via copy-paste from the test failure output - each failure
// prints a ready-to-paste stub.
//
// Most entries fall into a few categories:
//
//   - error_docs/Nxxxx.md "demo" snippets: the doc teaches a specific code by
//     showing a snippet that fires it. ExpectedCodes pins the code so a
//     language change that shifts what fires turns into a test failure
//     instead of silent drift.
//
//   - docs-web/docs/reference/syntax.md fragments: small illustrative bits
//     that reference placeholder helpers / variables. Pinned to current
//     diagnostic shape; language drift surfaces here too.
//
//   - Mid-rework files: Skip with a tracking note (rare).
var docSnippetTolerances = map[string]Tolerance{
	// ---- core/error_docs/* -----------------------------------------
	// Each error_docs/Nxxxx.md teaches a specific code by demonstrating
	// a snippet that fires it. The tolerance pins the code(s) that
	// legitimately fire in the demo.

	"core/error_docs/10001.md#8bc4fce1": {
		ExpectedCodes: []string{"RAD10009"},
		Reason:        "demo: `x + 1 = 6` - assigning to an expression. Parser fires generic RAD10009 (rather than the more-specific RAD10001) for this shape.",
	},
	"core/error_docs/10002.md#3b209c28": {
		ExpectedCodes: []string{"RAD10002", "RAD10009"},
		Reason:        "demo: missing-colon after if. RAD10002 is the specific diagnostic, RAD10009 is the cascading parser report on the unexpected following token.",
	},
	"core/error_docs/10008.md#075bba28": {
		ExpectedCodes: []string{"RAD10008"},
		Reason:        "demo: `args = 5` shadowing the args keyword.",
	},
	"core/error_docs/10018.md#9b45f920": {
		ExpectedCodes: []string{"RAD10018"},
		Reason:        "demo: missing indent after if-block header.",
	},
	"core/error_docs/10020.md#7ba6caed": {
		ExpectedCodes: []string{"RAD10020", "RAD20028"},
		Reason:        "demo: unterminated string literal. RAD10020 is the specific diagnostic; RAD20028 cascades on the words after the bad quote (Hello, world) which the parser re-interprets as identifiers.",
	},
	"core/error_docs/10021.md#09688585": {
		ExpectedCodes: []string{"RAD10009"},
		Reason:        "demo: `result = 5 3` - the parser fires generic RAD10009 ('unexpected 5') rather than a specific code for adjacent literals.",
	},
	"core/error_docs/10022.md#7ff5b919": {
		ExpectedCodes: []string{"RAD10009"},
		Reason:        "demo: orphan `else:` clause. Parser fires generic RAD10009.",
	},
	"core/error_docs/20028.md#6ae30c2a": {
		ExpectedCodes: []string{"RAD20028"},
		Reason:        "demo: scope leak - `config` defined inside setup() isn't visible after.",
	},
	"core/error_docs/20030.md#63c4f300": {
		ExpectedCodes: []string{"RAD20030"},
		Reason:        "demo: break outside a loop. (The old incidental RAD30002 on `i > 5` is gone now that comparisons decompose a numeric `float|int` union.)",
	},
	"core/error_docs/20031.md#34cd8497": {
		ExpectedCodes: []string{"RAD20031"},
		Reason:        "demo: continue outside a loop. (Old incidental RAD30002 gone now that numeric-union comparisons type-check.)",
	},
	"core/error_docs/30002.md#39029f35": {
		ExpectedCodes: []string{"RAD30002"},
		Reason:        "demo: `str + int` invalid operand types.",
	},
	"core/error_docs/30005.md#a96e6aa0": {
		ExpectedCodes: []string{"RAD10009"},
		Reason:        "demo: `5 = x` / `get_value() = 5`. Parser fires RAD10009 (generic) rather than a specific RAD30005 - the diagnostic name predates the parser-recovery path that would emit a more specific code.",
	},
	"core/error_docs/30006.md#4247599d": {
		ExpectedCodes: []string{"RAD30001"},
		Reason:        "demo: `len(42)` - the static checker reports the type mismatch via RAD30001 (the closest existing code for argument-type wrongness).",
	},
	"core/error_docs/30007.md#50e40b39": {
		ExpectedCodes: []string{"RAD30007"},
		Reason:        "demo: too few + too many arguments to a fn.",
	},
	"core/error_docs/40002.md#18c240a6": {
		ExpectedCodes: []string{"RAD40002"},
		Reason:        "demo: fn definition shadowing an args-block name.",
	},
	"core/error_docs/40004.md#79074c79": {
		ExpectedCodes: []string{"RAD40004"},
		Reason:        "demo: `return` at top level.",
	},
	"core/error_docs/40005.md#1ffddd58": {
		ExpectedCodes: []string{"RAD40005"},
		Reason:        "demo: `yield` at top level (outside a switch-case block).",
	},
	"core/error_docs/40006.md#da5f334e": {
		ExpectedCodes: []string{"RAD10009"},
		Reason:        "demo: `5 = x` and `a + b = 10`. Parser fires generic RAD10009 (same shape as 30005.md).",
	},
	"core/error_docs/40008.md#d08f0b15": {
		ExpectedCodes: []string{"RAD40008"},
		Reason:        "demo: deprecated `request`/`display` block keywords (removed in v0.9).",
	},
	"core/error_docs/40009.md#ad4b7b09": {
		ExpectedCodes: []string{"RAD40009"},
		Reason:        "demo: duplicate parameter name in a fn signature.",
	},
	"core/error_docs/40010.md#78dda972": {
		ExpectedCodes: []string{"RAD40010"},
		Reason:        "demo: non-exhaustive switch over a closed StrEnum type.",
	},
	"core/error_docs/40011.md#1aacd23b": {
		ExpectedCodes: []string{"RAD40011"},
		Reason:        "demo: re-declaring a name with a type annotation (same type).",
	},
	"core/error_docs/40012.md#6c159694": {
		ExpectedCodes: []string{"RAD40012"},
		Reason:        "demo: duplicate case key (unreachable case).",
	},
	"core/error_docs/40013.md#89ca9213": {
		ExpectedCodes: []string{"RAD40013"},
		Reason:        "demo: case key 'wat' not in the closed enum discriminant ['a', 'b', 'c'].",
	},

	"core/error_docs/40014.md#35566bd3": {
		ExpectedCodes: []string{"RAD40014"},
		Reason:        "demo: typo'd map key 'fjull_path' not in get_path's typed-map return shape.",
	},

	"core/error_docs/20028.md#a34c5fb8": {
		ExpectedCodes: []string{"RAD20028"},
		Reason:        "demo: typo'd identifier ('usernme' vs 'username').",
	},
	"core/error_docs/30001.md#eef6bed7": {
		ExpectedCodes: []string{"RAD30001"},
		Reason:        "demo: `double(x)` where x is `str` but the parameter is `int`. The checker reports as Hint via the general arg-assignability path; ExpectedCodes accepts at any severity.",
	},
	"core/error_docs/30010.md#4da5a819": {
		ExpectedCodes: []string{"RAD30010"},
		Reason:        "demo: assigning a 'str' to a 'int[]' element / 'int' map value. Hint severity by design at the indexed-assign site.",
	},
	"core/error_docs/40007.md#4d5acfcc": {
		ExpectedCodes: []string{"RAD40007"},
		Reason:        "demo: rad-block options ('insecure', 'quiet', 'noprint') used without a source - the warnings are exactly what the doc is teaching.",
	},
	"core/error_docs/40011.md#c8b94541": {
		ExpectedCodes: []string{"RAD40011", "RAD30001"},
		Reason:        "demo: typed re-declaration with a type change. RAD40011 is the headline diagnostic; RAD30001 cascades because 'str = \"hi\"' is also being assigned to the originally-declared 'int' slot.",
	},
	"core/error_docs/40012.md#ebde2ff5": {
		ExpectedCodes: []string{"RAD40012"},
		Reason:        "demo: exact-duplicate `case \"a\":` arms.",
	},

	// ---- docs-web/docs/examples/* ----------------------------------
	// Tutorial pages build up scripts incrementally. Intermediate
	// snippets call into functions defined later, show fragments
	// without surrounding context, or hit known checker false
	// positives in the final form.

	// epoch.md
	"docs-web/docs/examples/epoch.md#12150dae": {
		Skip:   true,
		Reason: "fragment: list-comprehension shown standalone in a !!! tip box, references `tz_to_flag` from the surrounding tutorial.",
	},
	"docs-web/docs/examples/epoch.md#35aa8cbd": {
		Skip:   true,
		Reason: "fragment: for-loop shown standalone to illustrate the verbose form; not meant to be runnable.",
	},
	"docs-web/docs/examples/epoch.md#220b2f53": {
		ExpectedCodes: []string{"RAD20028"},
		Reason:        "intermediate tutorial step: the script calls `parse_time` which the next step then defines. Demonstrates the build-up flow.",
	},
	"docs-web/docs/examples/epoch.md#bfc40cea": {
		ExpectedCodes: []string{"RAD30001"},
		Reason:        "checker false positive: narrowing on conditional reassignment + list-of-pairs tuple unpacking type inference. The script is correct at runtime; the static checker's flow inference doesn't yet track the reassignment fully (commit-3-era narrowing has gaps for `if not x: x = ...` patterns) and treats `for tz, _ in tz_to_flag:` as if tz binds to the inner list rather than its first element.",
	},
	"docs-web/docs/examples/epoch.md#5706ef9d": {
		ExpectedCodes: []string{"RAD30001"},
		Reason:        "same checker false positive as #bfc40cea - this is the duplicated final-form snippet (preview + tutorial end).",
	},

	// hm.md
	// #4eb5a0f8 (the final-form script + its preview duplicate) no longer
	// needs an entry: load_state() now strips its error arm at the call
	// (Gap 2), so `state` is `map` and the old `error|map` UFCS false
	// positive is gone. The unhandled-fallible hint it now carries is
	// globally tolerated.
	// Intermediate tutorial steps: each shows command blocks whose
	// callback functions are defined in a later step. A named callback
	// is a plain function reference, so a target not yet defined in the
	// snippet is an undefined-identifier error (RAD20028). Which targets
	// are still missing varies by step, so the reasons stay general.
	"docs-web/docs/examples/hm.md#19e0e50a": {
		ExpectedCodes: []string{"RAD20028"},
		Reason:        "intermediate tutorial step: command callbacks defined in later steps are undefined here.",
	},
	"docs-web/docs/examples/hm.md#03e13698": {
		ExpectedCodes: []string{"RAD20028"},
		Reason:        "intermediate tutorial step: command callbacks defined in later steps are undefined here.",
	},
	"docs-web/docs/examples/hm.md#0b53f387": {
		ExpectedCodes: []string{"RAD20028"},
		Reason:        "intermediate tutorial step: command callbacks defined in later steps are undefined here.",
	},
	"docs-web/docs/examples/hm.md#abb66476": {
		ExpectedCodes: []string{"RAD20028"},
		Reason:        "intermediate tutorial step: command callbacks defined in later steps are undefined here.",
	},

	// ---- docs-web/docs/reference/syntax.md -------------------------
	// Syntax reference. Snippets are small illustrative fragments
	// that reference placeholder helpers/values (e.g.
	// some_function_returning_tuple, log_admin_action) or show
	// parameter/call-syntax shapes out of function context.
	// Pinned to the current diagnostic shape so a language change
	// that shifts what these fragments produce shows up as a test
	// failure instead of silent drift.
	//
	// Most entries are RAD20028 alone (placeholder identifier
	// references). The mixed-code entries are parser-fragments where
	// the parser bails with RAD10001/10009/10021 etc.
	"docs-web/docs/reference/syntax.md#02019285": {ExpectedCodes: []string{"RAD20028"}, Reason: "syntax-reference fragment: placeholder identifier"},
	"docs-web/docs/reference/syntax.md#06decb32": {ExpectedCodes: []string{"RAD20028"}, Reason: "syntax-reference fragment: switch examples reference placeholder helpers"},
	"docs-web/docs/reference/syntax.md#0e09fee1": {ExpectedCodes: []string{"RAD10021"}, Reason: "syntax-reference fragment: named-parameter declaration shown out of function context"},
	"docs-web/docs/reference/syntax.md#1b5cdde1": {ExpectedCodes: []string{"RAD10001", "RAD10009", "RAD10021", "RAD20028"}, Reason: "syntax-reference fragment: function-parameter syntax, not a runnable call"},
	"docs-web/docs/reference/syntax.md#27f0df7a": {ExpectedCodes: []string{"RAD10001", "RAD10009", "RAD10020", "RAD10021"}, Reason: "syntax-reference fragment: function-signature shape, not runnable Rad"},
	"docs-web/docs/reference/syntax.md#2d0ec832": {ExpectedCodes: []string{"RAD20028"}, Reason: "syntax-reference fragment: references placeholder helper functions"},
	"docs-web/docs/reference/syntax.md#2e1f5722": {ExpectedCodes: []string{"RAD10001", "RAD10002", "RAD10009"}, Reason: "syntax-reference fragment: parser bails on incomplete header"},
	"docs-web/docs/reference/syntax.md#31052d6a": {ExpectedCodes: []string{"RAD20028"}, Reason: "syntax-reference fragment: references placeholder helper functions"},
	"docs-web/docs/reference/syntax.md#48c8517e": {ExpectedCodes: []string{"RAD20028"}, Reason: "syntax-reference fragment: references placeholder helper functions"},
	"docs-web/docs/reference/syntax.md#597de36b": {ExpectedCodes: []string{"RAD20028"}, Reason: "syntax-reference fragment: placeholder identifier"},
	"docs-web/docs/reference/syntax.md#5ee62c26": {ExpectedCodes: []string{"RAD20028"}, Reason: "syntax-reference fragment: placeholder identifier"},
	"docs-web/docs/reference/syntax.md#62312191": {ExpectedCodes: []string{"RAD10001"}, Reason: "syntax-reference fragment: literal-syntax shape, not a complete statement"},
	"docs-web/docs/reference/syntax.md#6aadf132": {ExpectedCodes: []string{"RAD20028"}, Reason: "syntax-reference fragment: references placeholder helper functions"},
	"docs-web/docs/reference/syntax.md#6ae643c4": {ExpectedCodes: []string{"RAD10002", "RAD20028"}, Reason: "syntax-reference fragment: incomplete block header + placeholder"},
	"docs-web/docs/reference/syntax.md#6cdc80ea": {ExpectedCodes: []string{"RAD10001"}, Reason: "syntax-reference fragment: literal-syntax shape, not a complete statement"},
	"docs-web/docs/reference/syntax.md#6ff9cb8a": {ExpectedCodes: []string{"RAD20028"}, Reason: "syntax-reference fragment: placeholder identifier"},
	"docs-web/docs/reference/syntax.md#71772390": {ExpectedCodes: []string{"RAD20028"}, Reason: "syntax-reference fragment: references placeholder helper functions"},
	"docs-web/docs/reference/syntax.md#740a28b0": {ExpectedCodes: []string{"RAD10001"}, Reason: "syntax-reference fragment: parser bails on incomplete shape"},
	"docs-web/docs/reference/syntax.md#749ba2bc": {ExpectedCodes: []string{"RAD10001", "RAD10009", "RAD10021"}, Reason: "syntax-reference fragment: type-annotation shape, not a complete statement"},
	"docs-web/docs/reference/syntax.md#80300185": {ExpectedCodes: []string{"RAD20028"}, Reason: "syntax-reference fragment: placeholder identifier"},
	"docs-web/docs/reference/syntax.md#85cb197a": {ExpectedCodes: []string{"RAD20028"}, Reason: "syntax-reference fragment: references placeholder helper functions"},
	"docs-web/docs/reference/syntax.md#8ab053dc": {ExpectedCodes: []string{"RAD20028"}, Reason: "syntax-reference fragment: references placeholder helper functions"},
	"docs-web/docs/reference/syntax.md#99de3c91": {ExpectedCodes: []string{"RAD20028"}, Reason: "syntax-reference fragment: references placeholder helper functions"},
	"docs-web/docs/reference/syntax.md#baae5c0f": {ExpectedCodes: []string{"RAD10009", "RAD20028"}, Reason: "syntax-reference fragment: parser bails + placeholder"},
	"docs-web/docs/reference/syntax.md#bc518d1c": {ExpectedCodes: []string{"RAD20028"}, Reason: "syntax-reference fragment: placeholder identifier"},
	"docs-web/docs/reference/syntax.md#be6fc738": {ExpectedCodes: []string{"RAD20028"}, Reason: "syntax-reference fragment: references placeholder helper functions"},
	"docs-web/docs/reference/syntax.md#c023cdc8": {ExpectedCodes: []string{"RAD10001"}, Reason: "syntax-reference fragment: literal-syntax shape, not a complete statement"},
	"docs-web/docs/reference/syntax.md#c8b1a340": {ExpectedCodes: []string{"RAD20028"}, Reason: "syntax-reference fragment: references placeholder helper functions"},
	"docs-web/docs/reference/syntax.md#ce0bdd06": {ExpectedCodes: []string{"RAD20028"}, Reason: "syntax-reference fragment: command callback to a placeholder helper that isn't defined in the fragment - undefined-identifier error"},
	"docs-web/docs/reference/syntax.md#d0901453": {ExpectedCodes: []string{"RAD20028"}, Reason: "syntax-reference fragment: placeholder identifier"},
	"docs-web/docs/reference/syntax.md#db1c71cb": {ExpectedCodes: []string{"RAD20028"}, Reason: "syntax-reference fragment: references placeholder helper functions"},
	"docs-web/docs/reference/syntax.md#dfc6c872": {ExpectedCodes: []string{"RAD20028"}, Reason: "syntax-reference fragment: placeholder identifier"},
	"docs-web/docs/reference/syntax.md#e2959ba4": {ExpectedCodes: []string{"RAD20028"}, Reason: "syntax-reference fragment: placeholder identifier"},
	"docs-web/docs/reference/syntax.md#e53e347e": {ExpectedCodes: []string{"RAD20028"}, Reason: "syntax-reference fragment: references placeholder helper functions"},
	"docs-web/docs/reference/syntax.md#f1256dfc": {ExpectedCodes: []string{"RAD20028"}, Reason: "syntax-reference fragment: references placeholder helper functions"},
	"docs-web/docs/reference/syntax.md#f1b88483": {ExpectedCodes: []string{"RAD20028"}, Reason: "syntax-reference fragment: references placeholder helper functions"},
	"docs-web/docs/reference/syntax.md#fcb5249a": {ExpectedCodes: []string{"RAD20028"}, Reason: "syntax-reference fragment: placeholder identifier"},
	"docs-web/docs/reference/syntax.md#fe670c01": {ExpectedCodes: []string{"RAD10001", "RAD10021"}, Reason: "syntax-reference fragment: parameter-declaration shape, not a complete statement"},

	// ---- docs-web/docs/migrations/* --------------------------------
	// Migration guides document syntax changes between versions. The
	// "Before" snippets intentionally show old syntax that no longer
	// parses; the "After" snippets reference placeholder data
	// (`names`, `items`, etc.) for illustration. The point is the
	// shape of the change, not runnable scripts. Skip is appropriate
	// here - drift in these historical guides is unlikely and the
	// guides should remain readable in their original form.

	"docs-web/docs/migrations/v0.7.md#2d7ecdf7": {Skip: true, Reason: "migration guide: before/after example with placeholder data"},
	"docs-web/docs/migrations/v0.7.md#6db1bdca": {Skip: true, Reason: "migration guide: before/after example with placeholder data"},
	"docs-web/docs/migrations/v0.7.md#70359282": {Skip: true, Reason: "migration guide: post-migration example with placeholder data"},
	"docs-web/docs/migrations/v0.7.md#972a32f8": {Skip: true, Reason: "migration guide: field-modifier fragment shown outside its enclosing rad block"},
	"docs-web/docs/migrations/v0.7.md#caeaf4c2": {Skip: true, Reason: "migration guide: pre-migration syntax demonstration"},
	"docs-web/docs/migrations/v0.7.md#e5b4dfb5": {Skip: true, Reason: "migration guide: before/after example with placeholder data"},

	"docs-web/docs/migrations/v0.8.md#03ef8490": {Skip: true, Reason: "migration guide: before/after example"},
	"docs-web/docs/migrations/v0.8.md#1b69685b": {Skip: true, Reason: "migration guide: before/after example"},
	"docs-web/docs/migrations/v0.8.md#605fdb7f": {Skip: true, Reason: "migration guide: before/after example"},
	"docs-web/docs/migrations/v0.8.md#9cb85811": {Skip: true, Reason: "migration guide: before/after example"},
	"docs-web/docs/migrations/v0.8.md#c1e69a91": {Skip: true, Reason: "migration guide: before/after example"},

	"docs-web/docs/migrations/v0.9.md#0a2e075f": {Skip: true, Reason: "migration guide: before/after example"},
	"docs-web/docs/migrations/v0.9.md#103e45ed": {Skip: true, Reason: "migration guide: before/after example"},
	"docs-web/docs/migrations/v0.9.md#1a1b40a1": {Skip: true, Reason: "migration guide: before/after example"},
	"docs-web/docs/migrations/v0.9.md#46d67a51": {Skip: true, Reason: "migration guide: before/after example"},
	"docs-web/docs/migrations/v0.9.md#65133c82": {Skip: true, Reason: "migration guide: before/after example"},
	"docs-web/docs/migrations/v0.9.md#99b3964d": {Skip: true, Reason: "migration guide: before/after example"},
	"docs-web/docs/migrations/v0.9.md#b0cf666d": {Skip: true, Reason: "migration guide: before/after example"},

	// ---- docs-web/docs/guide/basics.md ------------------------------
	"docs-web/docs/guide/basics.md#210039b1": {
		Skip:   true,
		Reason: "guide fragment: illustrating destructuring syntax with placeholder fn calls (get_coordinates/get_dimensions).",
	},
	"docs-web/docs/guide/basics.md#3cc65dff": {
		Skip:   true,
		Reason: "guide fragment: map access shown with placeholder maps from earlier in the doc.",
	},
	"docs-web/docs/guide/basics.md#42184492": {
		Skip:   true,
		Reason: "guide fragment: truthy-check example with placeholder 'my_list'.",
	},
	"docs-web/docs/guide/basics.md#630b78cc": {
		ExpectedCodes: []string{"RAD10001", "RAD10009"},
		Reason:        "intentional demo: 'a++ > 0' / 'b = a++' shown as invalid - the doc is teaching that ++ doesn't return a value.",
	},
	"docs-web/docs/guide/basics.md#79bfcb2e": {
		Skip:   true,
		Reason: "guide fragment: len-check shown with placeholder 'my_list'.",
	},
	"docs-web/docs/guide/basics.md#a4a5a25f": {
		Skip:   true,
		Reason: "guide fragment: list indexing demo with placeholder lists from earlier in the doc.",
	},
	"docs-web/docs/guide/basics.md#c8768f04": {
		Skip:   true,
		Reason: "guide fragment: map dot-access with placeholder maps.",
	},
	"docs-web/docs/guide/basics.md#d307cd5f": {
		Skip:   true,
		Reason: "guide fragment: zip iteration with placeholder lists from earlier in the doc.",
	},

	// ---- docs-web/docs/guide/error-handling.md ----------------------
	// Error-handling guide examples reference placeholder data /
	// functions (age_str, parse_int chains, response objects) to
	// keep the focus on the error-handling pattern, not on building
	// runnable scaffolding around each one.

	"docs-web/docs/guide/error-handling.md#10f68837": {
		Skip:   true,
		Reason: "guide fragment: ?? fallback patterns with placeholder inputs.",
	},
	"docs-web/docs/guide/error-handling.md#61ad72f0": {
		Skip:   true,
		Reason: "guide fragment: catch block with placeholder 'temp_file' / 'delete_path'.",
	},
	"docs-web/docs/guide/error-handling.md#69d3eb50": {
		Skip:   true,
		Reason: "guide fragment: ?? chaining with placeholder 'user' / 'config_path'.",
	},
	// #82491324 no longer needs an entry: `parse_float(...)` now strips
	// its error arm at the call (Gap 2), so `price` is `float` and the
	// old `float|error * float` RAD30002 hint is gone - the very false
	// positive that hint represented. It now carries the globally-
	// tolerated unhandled-fallible hint instead.
	"docs-web/docs/guide/error-handling.md#8ce3fff7": {
		Skip:   true,
		Reason: "guide fragment: catch-chaining with placeholder 'risky_call' / 'fallback_call'.",
	},
	// #b558bc7f no longer needs an entry: `parse_int(port_str) ?? 8080`
	// now yields `int` because the call strips its error arm before `??`
	// (Gap 2), so the residual-`int|error|int` RAD30001 hint into
	// validate_port is gone. This is exactly the narrowing the old
	// comment wished for.
	"docs-web/docs/guide/error-handling.md#e380902c": {
		Skip:   true,
		Reason: "guide fragment: nested-access with placeholder 'response'.",
	},
	"docs-web/docs/guide/error-handling.md#e534702b": {
		Skip:   true,
		Reason: "guide fragment: catch variants with placeholder inputs.",
	},
	"docs-web/docs/guide/error-handling.md#e905ce4a": {
		Skip:   true,
		Reason: "guide fragment: ?? vs catch comparison with placeholder inputs.",
	},

	// ---- docs-web/docs/guide/functions.md ---------------------------
	"docs-web/docs/guide/functions.md#426da0a8": {
		ExpectedCodes: []string{"RAD20028"},
		Reason:        "intentional demo: calling 'helper()' before its definition - the doc teaches that nested-fn definitions are NOT hoisted.",
	},
	"docs-web/docs/guide/functions.md#8bf15e1a": {
		Skip:   true,
		Reason: "guide fragment: http_post UFCS chain with placeholder 'url'.",
	},
	"docs-web/docs/guide/functions.md#d14f4dfc": {
		Skip:   true,
		Reason: "guide fragment: UFCS chaining demo with placeholder 'text'.",
	},

	// ---- docs-web/docs/guide/rad-blocks.md --------------------------
	// rad-block snippets generally need a URL source and JSON-path
	// declarations to be self-contained. The guide focuses on the
	// rad-block syntax / option shape, so most snippets use a
	// placeholder 'url' (string variable) and reference Name/Age/etc
	// from JSON paths defined in earlier snippets.

	"docs-web/docs/guide/rad-blocks.md#022c66af": {
		ExpectedCodes: []string{"RAD10001", "RAD10009", "RAD10020"},
		Reason:        "field-modifier fragment: multi-column modifier shown without enclosing rad block, parser fails.",
	},
	"docs-web/docs/guide/rad-blocks.md#0560b3e1": {
		Skip:   true,
		Reason: "guide fragment: http_get with placeholder 'url' and 'my_headers'.",
	},
	"docs-web/docs/guide/rad-blocks.md#0ad1a299": {
		Skip:   true,
		Reason: "rad-block fragment: placeholder 'url' source.",
	},
	"docs-web/docs/guide/rad-blocks.md#61689672": {
		Skip:   true,
		Reason: "rad-block fragment: args + body shown without complete script context.",
	},
	"docs-web/docs/guide/rad-blocks.md#68d6ee62": {
		Skip:   true,
		Reason: "rad-block fragment: placeholder 'url' source.",
	},
	"docs-web/docs/guide/rad-blocks.md#71667048": {
		Skip:   true,
		Reason: "rad-block fragment: placeholder 'url' source.",
	},
	"docs-web/docs/guide/rad-blocks.md#7c5675fc": {
		Skip:   true,
		Reason: "rad-block fragment: placeholder 'url' source.",
	},
	"docs-web/docs/guide/rad-blocks.md#90e9bfe7": {
		Skip:   true,
		Reason: "rad-block fragment: placeholder 'url' source + field-modifier.",
	},
	"docs-web/docs/guide/rad-blocks.md#a34f9846": {
		Skip:   true,
		Reason: "rad-block fragment: placeholder 'url' source + sort option.",
	},
	"docs-web/docs/guide/rad-blocks.md#a40787d6": {
		ExpectedCodes: []string{"RAD10021"},
		Reason:        "field-modifier fragment: 'Population: / map fn(p) ...' shown outside its enclosing rad block.",
	},
	"docs-web/docs/guide/rad-blocks.md#ad6e335e": {
		ExpectedCodes: []string{"RAD10001", "RAD10009", "RAD10021"},
		Reason:        "field-modifier fragment: filter modifier shown outside its enclosing rad block.",
	},
	"docs-web/docs/guide/rad-blocks.md#c16433f2": {
		ExpectedCodes: []string{"RAD10021"},
		Reason:        "field-modifier fragment: map with format spec shown outside its enclosing rad block.",
	},
	"docs-web/docs/guide/rad-blocks.md#ceb89e4a": {
		Skip:   true,
		Reason: "rad-block fragment: placeholder 'url' source + sort option.",
	},
	"docs-web/docs/guide/rad-blocks.md#d868d02e": {
		Skip:   true,
		Reason: "rad-block fragment: placeholder 'url' source + sort desc.",
	},
	"docs-web/docs/guide/rad-blocks.md#e01da3b0": {
		Skip:   true,
		Reason: "rad-block fragment: placeholder 'url' source.",
	},
	"docs-web/docs/guide/rad-blocks.md#e1abf714": {
		ExpectedCodes: []string{"RAD10021"},
		Reason:        "field-modifier fragment: map with ctx shown outside its enclosing rad block.",
	},
	"docs-web/docs/guide/rad-blocks.md#e2cf2fba": {
		Skip:   true,
		Reason: "rad-block fragment: placeholder 'url' source + conditional logic on fields.",
	},
	"docs-web/docs/guide/rad-blocks.md#e86fab3c": {
		Skip:   true,
		Reason: "rad-block fragment: placeholder 'url' source + field-modifier.",
	},
	"docs-web/docs/guide/rad-blocks.md#ed646d32": {
		Skip:   true,
		Reason: "rad-block fragment: placeholder 'url' source + field-modifier.",
	},
	"docs-web/docs/guide/rad-blocks.md#f0668f92": {
		Skip:   true,
		Reason: "rad-block fragment: placeholder 'url' source + multi-field modifier.",
	},

	// ---- docs-web/docs/guide/script-commands.md ---------------------
	"docs-web/docs/guide/script-commands.md#3e824696": {
		ExpectedCodes: []string{"RAD10001"},
		Reason:        "guide fragment: command block declaration shown standalone (without an enclosing script context).",
	},

	// ---- docs-web/docs/guide/stashes.md -----------------------------
	// Most load_state() snippets no longer need entries: load_state()
	// strips its error arm at the call (Gap 2), so `state` is `map` and
	// the old error|map RAD30001 hints are gone; they now carry only the
	// globally-tolerated unhandled-fallible hint.
	"docs-web/docs/guide/stashes.md#ed4e81de": {
		ExpectedCodes: []string{"RAD30002"},
		Reason:        "checker hint: `count = state[\"count\"] ?? 0` is `dynamic|int` (untyped map index), so `count++` flags `dynamic|int + int`. The script is correct at runtime. Was RAD30001 pre-Gap-2 (load_state's error|map); the error arm now strips, surfacing the underlying dynamic-map-index hint instead.",
	},

	// ---- docs-web/docs/guide/type-annotations.md --------------------
	// Type-annotation examples demonstrate function signatures against
	// typed returns. The loop-built-collection cases (`counts = {};
	// counts[k] = 1; return counts`) no longer fire: a bare `list`/`map`
	// now flows into a typed `T[]`/`{K: V}` as the gradual escape hatch.
	// The remaining pins are intentional teaching demos (bad arg type,
	// missing `T?`) or a still-imprecise vararg-into-typed-param case.

	"docs-web/docs/guide/type-annotations.md#8bcb60a4": {
		ExpectedCodes: []string{"RAD30001"},
		Reason:        "checker hint: vararg `*data_points: int|float` doesn't refine into `sum()`'s `float[]` parameter. (The old RAD30002 on the division is gone now that numeric-union operands decompose.)",
	},
	"docs-web/docs/guide/type-annotations.md#b264b53b": {
		ExpectedCodes: []string{"RAD30001"},
		Reason:        "intentional demo: returning null from fn typed as str (without ?). Doc teaches that T? is required for nullable returns.",
	},
	"docs-web/docs/guide/type-annotations.md#b9ca5c02": {
		ExpectedCodes: []string{"RAD30001"},
		Reason:        "intentional demo: calling greet(42) where greet expects str - the doc teaches arg-type checking.",
	},

	// ---- docs-web/docs/releases.md ----------------------------------
	"docs-web/docs/releases.md#10ab0aee": {
		ExpectedCodes: []string{"RAD10009"},
		Reason:        "release notes: shows pre-migration syntax samples ('$!' critical, 'unsafe', 'recover:', 'fail:') that no longer parse - intentional historical reference.",
	},
}

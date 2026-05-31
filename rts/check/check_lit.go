package check

import (
	"github.com/amterp/rad/rts/rl"
)

// checkResult is the verdict of a scoped bidirectional check (see check).
// On success, ok is true and the remaining fields describe the whole node.
// On failure, offending pinpoints the deepest literal element that doesn't
// fit, with got/want the concrete types at that point - so a call site can
// emit a localized diagnostic ("element 2: int is not assignable to str")
// instead of a whole-literal one. For a top-level (non-nested) mismatch
// offending is the node itself and got/want equal the synthesized and
// expected types, so the message reads exactly as the old
// synth + IsAssignableFrom path did.
type checkResult struct {
	ok        bool
	offending rl.Node
	got       rl.TypingT
	want      rl.TypingT
}

// check is the structural companion to synth: it decides whether a value
// node is assignable to an expected type, looking *through* the literal
// shapes that synth widens away. synth turns `[1, "2"]` into `(int|str)[]`
// and `{ "k": 1 }` into `{ str: int }`, neither of which matches a
// tuple/struct target under the invariant collection rules - even when the
// literal is a perfectly valid inhabitant. check recovers that fidelity by
// walking the literal against the target recursively: `[1, "2"]` checks
// position-wise against `[int, str]`, `{ "k": 1 }` field-wise against
// `{ "k": int }`, and a plain string against a str-enum by membership.
//
// It always calls synth(node) first so ExprTypes/hover stay coherent (the
// widened type is still what's cached and shown on hover); the structural
// verdict layers on top without writing anything new to the index. synth
// never calls check, so there is no recursion cycle.
//
// Non-literal nodes, and literals whose target isn't a structured type,
// fall through to synth + IsAssignableFrom, making check a strict superset
// of the old `expected.IsAssignableFrom(synth(node))` test.
func (tc *typeChecker) check(node rl.Node, expected rl.TypingT) checkResult {
	got := tc.synth(node)
	// Unknown / already-poisoned source types accept anything, mirroring
	// the short-circuits the call sites used before delegating here.
	if got == nil || isErrorType(got) || isDynamicLike(got) {
		return accept(node, got, expected)
	}
	if expected == nil {
		return accept(node, got, expected)
	}

	// Fast path: a synthesized type that's already assignable is accepted
	// outright. IsAssignableFrom is sound - it never accepts an invalid
	// value - and only *under*-approximates for literals flowing into
	// structured types (a tuple, struct, or str-enum target). Everything
	// below exists solely to recover those false negatives, so anything
	// IsAssignableFrom already admits (including `int?` into `int?` and
	// `int|error` into `int|error`) must short-circuit here first.
	if expected.IsAssignableFrom(got) {
		return accept(node, got, expected)
	}

	// Peel the expected type to recover the literal-into-structured cases.
	switch exp := expected.(type) {
	case *rl.TypingOptionalT:
		// The whole value didn't fit `T?` (so it isn't null - the fast
		// path would have caught that); try the inner type so a literal
		// can still match it structurally.
		return tc.check(node, exp.Inner())
	case *rl.TypingUnionT:
		// Assignable if the value fits any arm. We recurse rather than
		// stop at the fast path so a literal can match a *structured*
		// arm (tuple, struct, str-enum) that IsAssignableFrom rejects.
		for _, arm := range exp.Types() {
			if res := tc.check(node, arm); res.ok {
				return accept(node, got, expected)
			}
		}
		return reject(node, got, expected)
	}

	// Structural literal rules.
	switch lit := node.(type) {
	case *rl.LitList:
		return tc.checkListLit(lit, expected, got)
	case *rl.LitMap:
		return tc.checkMapLit(lit, expected, got)
	case *rl.LitString:
		if enum, isEnum := expected.(*rl.TypingStrEnumT); isEnum && lit.Simple && strEnumContains(enum, lit.Value) {
			return accept(node, got, expected)
		}
	}

	return reject(node, got, expected)
}

// checkListLit checks a list literal against a tuple, list, or open-list
// target. Tuple targets check position-wise (arity must match); list
// targets check every element against the declared element type. Any other
// target falls back to the synthesized list type's assignability.
func (tc *typeChecker) checkListLit(lit *rl.LitList, expected, got rl.TypingT) checkResult {
	switch exp := expected.(type) {
	case *rl.TypingTupleT:
		types := exp.Types()
		if len(types) != len(lit.Elements) {
			// Arity mismatch: nothing finer to point at than the literal.
			return reject(lit, got, expected)
		}
		for i, elem := range lit.Elements {
			if res := tc.check(elem, types[i]); !res.ok {
				return res
			}
		}
		return accept(lit, got, expected)
	case *rl.TypingListT:
		elemT := exp.Elem()
		for _, elem := range lit.Elements {
			if res := tc.check(elem, elemT); !res.ok {
				return res
			}
		}
		return accept(lit, got, expected)
	case *rl.TypingAnyListT:
		// Open `list` annotation - any list literal satisfies it.
		return accept(lit, got, expected)
	}
	// check's fast path already ruled out plain assignability before
	// dispatching here, so any non-structured target is a genuine miss.
	return reject(lit, got, expected)
}

// checkMapLit checks a map literal against a struct or map target. A struct
// target requires every entry key to be a plain string literal naming a
// declared field, each value to fit its field type, and every required
// field to be present. A map target checks all keys against K and all
// values against V. Other targets fall back to assignability.
func (tc *typeChecker) checkMapLit(lit *rl.LitMap, expected, got rl.TypingT) checkResult {
	switch exp := expected.(type) {
	case *rl.TypingStructT:
		seen := make(map[string]bool, len(lit.Entries))
		for _, entry := range lit.Entries {
			name, isStr := simpleStringKey(entry.Key)
			if !isStr {
				// A computed/interpolated key can't be matched to a named
				// field statically. Point at the key.
				return reject(entry.Key, got, expected)
			}
			fieldT, _, found := exp.Field(name)
			if !found {
				return reject(entry.Key, got, expected)
			}
			if res := tc.check(entry.Value, fieldT); !res.ok {
				return res
			}
			seen[name] = true
		}
		// Every required (non-optional) field must be supplied.
		for key := range exp.Fields() {
			if !key.IsOptional && !seen[key.Name] {
				return reject(lit, got, expected)
			}
		}
		return accept(lit, got, expected)
	case *rl.TypingMapT:
		keyT, valT := exp.KeyT(), exp.ValT()
		for _, entry := range lit.Entries {
			if res := tc.check(entry.Key, keyT); !res.ok {
				return res
			}
			if res := tc.check(entry.Value, valT); !res.ok {
				return res
			}
		}
		return accept(lit, got, expected)
	case *rl.TypingAnyMapT:
		// Open `map` annotation - any map literal satisfies it.
		return accept(lit, got, expected)
	}
	// check's fast path already ruled out plain assignability before
	// dispatching here, so any non-structured target is a genuine miss.
	return reject(lit, got, expected)
}

// simpleStringKey returns the value of a plain (non-interpolated) string
// literal key and true; anything else (interpolated string, identifier,
// computed expression) returns false - it can't be matched to a named
// struct field statically.
func simpleStringKey(key rl.Node) (string, bool) {
	lit, isStr := key.(*rl.LitString)
	if !isStr || !lit.Simple {
		return "", false
	}
	return lit.Value, true
}

func strEnumContains(enum *rl.TypingStrEnumT, value string) bool {
	for _, v := range enum.Values() {
		if v == value {
			return true
		}
	}
	return false
}

func accept(node rl.Node, got, want rl.TypingT) checkResult {
	return checkResult{ok: true, offending: node, got: got, want: want}
}

func reject(node rl.Node, got, want rl.TypingT) checkResult {
	return checkResult{ok: false, offending: node, got: got, want: want}
}

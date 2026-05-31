package check

import (
	"fmt"

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
//
// gateable records whether a *failure* is provable enough to block runtime
// (Error) rather than merely suspected (Hint). It's false when the real
// value is unknown to the checker - an open gradual container (`list` /
// `map`) whose contents we can't see, or a non-literal string flowing into
// a str-enum (the string's runtime value decides membership). Gating those
// would brick valid programs, which the severity policy forbids. Meaningless
// when ok is true.
type checkResult struct {
	ok        bool
	offending rl.Node
	got       rl.TypingT
	want      rl.TypingT
	gateable  bool
	// detail is an optional clarifying note appended to a call site's
	// diagnostic message, for failures the got/want pair alone doesn't
	// explain (e.g. which required struct field was missing).
	detail string
}

// detailSuffix renders detail as a parenthetical message suffix, or "".
func (r checkResult) detailSuffix() string {
	if r.detail == "" {
		return ""
	}
	return " (" + r.detail + ")"
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
		// A union miss is only gateable when *every* arm was provably
		// missed - if any arm's miss was uncertain, the value might fit
		// it at runtime.
		allGateable := true
		for _, arm := range exp.Types() {
			res := tc.check(node, arm)
			if res.ok {
				return accept(node, got, expected)
			}
			allGateable = allGateable && res.gateable
		}
		return rejectWith(node, got, expected, allGateable)
	}

	// Structural literal rules.
	switch lit := node.(type) {
	case *rl.LitList:
		return tc.checkListLit(lit, expected, got)
	case *rl.LitMap:
		return tc.checkMapLit(lit, expected, got)
	case *rl.LitString:
		if enum, isEnum := expected.(*rl.TypingStrEnumT); isEnum && lit.Simple {
			if enum.Contains(lit.Value) {
				return accept(node, got, expected)
			}
			// A simple literal not among the members is provably wrong.
			return reject(node, got, expected)
		}
	}

	// Leaf mismatch: gateable unless the real value is unknown to us.
	return rejectWith(node, got, expected, gateableMismatch(got, expected))
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
// target requires every declared field's value to fit its type and every
// required field to be present; extra keys are allowed (the runtime ignores
// keys beyond the declared shape, so `{ "a": 1, "b": 2 }` satisfies
// `{ "a": int }`). A map target checks all keys against K and all values
// against V. Other targets fall back to assignability.
func (tc *typeChecker) checkMapLit(lit *rl.LitMap, expected, got rl.TypingT) checkResult {
	switch exp := expected.(type) {
	case *rl.TypingStructT:
		seen := make(map[string]bool, len(lit.Entries))
		hasComputedKey := false
		for _, entry := range lit.Entries {
			name, isStr := simpleStringKey(entry.Key)
			if !isStr {
				// A computed/interpolated key might name any field - we
				// can't tell which, and it might satisfy a required one.
				// (Extra keys are valid anyway.) Defer judgement.
				hasComputedKey = true
				continue
			}
			fieldT, _, found := exp.Field(name)
			if !found {
				// Extra field beyond the declared shape - runtime ignores it.
				continue
			}
			if res := tc.check(entry.Value, fieldT); !res.ok {
				return res
			}
			seen[name] = true
		}
		// Every required (non-optional) field must be present - but only
		// gate the absence when no computed key could be supplying it.
		if !hasComputedKey {
			for key := range exp.Fields() {
				if !key.IsOptional && !seen[key.Name] {
					res := reject(lit, got, expected)
					res.detail = fmt.Sprintf("missing required field '%s'", key.Name)
					return res
				}
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

func accept(node rl.Node, got, want rl.TypingT) checkResult {
	return checkResult{ok: true, offending: node, got: got, want: want, gateable: true}
}

// reject builds a provably-wrong (gateable) failure - the default for a
// concrete structural conflict. Use rejectWith when the confidence is
// conditional.
func reject(node rl.Node, got, want rl.TypingT) checkResult {
	return rejectWith(node, got, want, true)
}

func rejectWith(node rl.Node, got, want rl.TypingT, gateable bool) checkResult {
	return checkResult{ok: false, offending: node, got: got, want: want, gateable: gateable}
}

// gateableMismatch reports whether a leaf (scalar / non-structured)
// mismatch is provable enough to gate as an Error. It is true only when
// *every* possible runtime value of got is definitely incompatible with
// want. It's false - merely a Hint - whenever some runtime value could
// fit, i.e. the checker can't see the real value:
//
//   - an open gradual container (`list` / `map`) whose contents are hidden;
//   - a non-literal string into a str-enum (a simple literal would have
//     been resolved by membership upstream, so a bare `str` here is an
//     interpolation/variable whose value decides membership at runtime);
//   - a union or optional whose arms include one that would fit (e.g.
//     `int|str` into `int`, or `str?` into `str` - the value might be the
//     compatible arm).
func gateableMismatch(got, want rl.TypingT) bool {
	if isGradualContainer(got) {
		return false
	}
	switch g := got.(type) {
	case *rl.TypingUnionT:
		return allArmsGateable(g.Types(), want)
	case *rl.TypingOptionalT:
		// T? is T | null: gateable only if neither could fit.
		return allArmsGateable([]rl.TypingT{g.Inner(), rl.NewNullType()}, want)
	}
	if _, isEnum := want.(*rl.TypingStrEnumT); isEnum {
		if _, isStr := got.(*rl.TypingStrT); isStr {
			return false
		}
	}
	return true
}

// allArmsGateable reports whether every arm of a union/optional source is
// a provable mismatch against want. If any arm is assignable (a runtime
// value of that arm would fit) or is itself an uncertain mismatch, the
// whole thing is uncertain.
func allArmsGateable(arms []rl.TypingT, want rl.TypingT) bool {
	for _, arm := range arms {
		if want.IsAssignableFrom(arm) {
			return false
		}
		if !gateableMismatch(arm, want) {
			return false
		}
	}
	return true
}

// isGradualContainer reports whether t is one of the open gradual
// collection types (`list` / `map`) - the gradual-typing escape hatch
// whose element/value contents the checker doesn't track.
func isGradualContainer(t rl.TypingT) bool {
	switch t.(type) {
	case *rl.TypingAnyListT, *rl.TypingAnyMapT:
		return true
	}
	return false
}

// typeIsGateable reports whether a synthesized type is concrete enough
// that a mismatch against it is provable - the fallback the return path
// uses when there's no value node to walk (a bare return, or a
// multi-value shape we can't decompose position-wise). It's false
// whenever the type carries a component whose runtime value the checker
// can't pin down: an open gradual container, or a union/optional whose
// arms might include the compatible one. Such a mismatch stays a Hint,
// the same provable-only rule gateableMismatch applies at the leaves.
func typeIsGateable(t rl.TypingT) bool {
	switch v := t.(type) {
	case *rl.TypingAnyListT, *rl.TypingAnyMapT:
		return false
	case *rl.TypingUnionT, *rl.TypingOptionalT:
		// An arm might be the compatible one at runtime - not provable.
		return false
	case *rl.TypingTupleT:
		for _, e := range v.Types() {
			if !typeIsGateable(e) {
				return false
			}
		}
	}
	return true
}

// mismatchSeverity maps a failed checkResult to the severity its call site
// should emit: a blocking Error when the failure is provable, a non-gating
// Hint when the real value is unknown (see checkResult.gateable).
func mismatchSeverity(res checkResult) IssueSeverity {
	if res.gateable {
		return IssueError
	}
	return IssueHint
}

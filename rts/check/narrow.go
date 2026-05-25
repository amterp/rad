package check

import "github.com/amterp/rad/rts/rl"

// Frame is an immutable overlay of flow-sensitive narrowings on top of
// the base symbol types tracked in TypeInfo.SymbolTypes. Frames chain
// to a parent so a child frame can shadow or extend without copying.
//
// Used by the narrowing pass: inside `if x != null:`, the frame in
// effect for the then-body binds x to its non-null component while
// leaving the surrounding frame untouched.
//
// Frames are immutable; "narrow this symbol" produces a new frame via
// With. That keeps branch-join trivial - each branch ends with its
// own frame, and the join walks both to produce a unioned overlay.
type Frame struct {
	parent   *Frame
	bindings map[*Symbol]rl.TypingT
}

// NewFrame returns a root frame with no narrowings. Used at the entry
// of every scope (file body, function body, lambda body) where no
// outer narrowings are in effect.
func NewFrame() *Frame {
	return &Frame{}
}

// With returns a new frame that narrows sym to t. The receiver is
// unchanged. O(1) - no map copy.
func (f *Frame) With(sym *Symbol, t rl.TypingT) *Frame {
	return &Frame{parent: f, bindings: map[*Symbol]rl.TypingT{sym: t}}
}

// WithMany returns a new frame applying every entry in m as a
// narrowing. Useful when a single predicate refines multiple symbols
// at once (e.g. a compound condition or a switch peeling several
// variants in one go). Returns the receiver if m is empty.
func (f *Frame) WithMany(m map[*Symbol]rl.TypingT) *Frame {
	if len(m) == 0 {
		return f
	}
	cp := make(map[*Symbol]rl.TypingT, len(m))
	for k, v := range m {
		cp[k] = v
	}
	return &Frame{parent: f, bindings: cp}
}

// Lookup walks the frame chain looking for a narrowing on sym. Returns
// the narrowed type and true if found in this frame or any ancestor.
// Callers fall back to TypeInfo.SymbolTypes when no narrowing is
// recorded.
func (f *Frame) Lookup(sym *Symbol) (rl.TypingT, bool) {
	for cur := f; cur != nil; cur = cur.parent {
		if t, ok := cur.bindings[sym]; ok {
			return t, true
		}
	}
	return nil, false
}

// Parent returns the parent frame, or nil at the root. Exposed so the
// flow-sensitive logic can walk back to a known point (e.g. unwind a
// branch scope on early-exit).
func (f *Frame) Parent() *Frame {
	if f == nil {
		return nil
	}
	return f.parent
}

// Refinement is the truthy/falsy decomposition of a predicate: the
// narrowings that hold when the condition evaluates truthy live in
// WhenTrue, the ones that hold on the falsy branch live in WhenFalse.
// Either map may be empty/nil if the predicate doesn't refine that
// side.
//
// Built by interpretCondition when the type checker walks an
// if/while/ternary condition. Consumers layer the chosen side onto
// the current frame via Frame.WithMany.
type Refinement struct {
	WhenTrue  map[*Symbol]rl.TypingT
	WhenFalse map[*Symbol]rl.TypingT
}

// EmptyRefinement is the no-op refinement used for conditions the
// narrower doesn't know how to interpret. Returned in preference to a
// zero-value Refinement so call sites can rely on the maps being
// non-nil if needed without re-checking.
func EmptyRefinement() Refinement {
	return Refinement{
		WhenTrue:  map[*Symbol]rl.TypingT{},
		WhenFalse: map[*Symbol]rl.TypingT{},
	}
}

// Negate swaps WhenTrue and WhenFalse. Used for `not <cond>` where
// the truthy / falsy branches of the inner predicate are inverted
// relative to the surrounding flow.
func (r Refinement) Negate() Refinement {
	return Refinement{WhenTrue: r.WhenFalse, WhenFalse: r.WhenTrue}
}

// --- Condition interpretation ----------------------------------------
//
// interpretCondition turns a boolean AST condition into a Refinement.
// The catalog is intentionally narrow - we only recognize predicates
// whose shape lets us soundly invert the comparison. Unknown shapes
// return an EmptyRefinement and the surrounding flow logic treats it
// as "no narrowing this boolean."
//
// Each call sub-commit (4b, 4c, 4d, ...) adds another predicate case.
// Phase 4b lands the dispatcher + the null cases.

// interpretCondition is the entry-point dispatcher. Lives as a method
// on typeChecker because every predicate needs access to the
// resolved view (to map identifier -> Symbol) and to the current
// SymbolTypes / frame (to read the base type being refined).
func (tc *typeChecker) interpretCondition(cond rl.Node, frame *Frame) Refinement {
	if cond == nil {
		return EmptyRefinement()
	}
	switch c := cond.(type) {
	case *rl.OpBinary:
		return tc.interpretBinaryCondition(c, frame)
	case *rl.OpUnary:
		return tc.interpretUnaryCondition(c, frame)
	case *rl.Identifier:
		return tc.narrowTruthyIdent(c, frame)
	}
	return EmptyRefinement()
}

// narrowTruthyIdent handles bare `if x:` predicates. When x has a
// nullable component (Optional<T> or a union containing null) the
// truthy branch strips it - inside the body, x is the non-null part.
//
// The falsy branch deliberately does NOT narrow x to null. Rad's
// truthiness is looser than equality-to-null: empty strings, zero,
// empty collections all evaluate falsy too. So `if x:` becoming
// falsy doesn't prove x is null, only that x is one of "various
// falsy values" - which we can't represent precisely. Empty
// WhenFalse is the conservative answer.
func (tc *typeChecker) narrowTruthyIdent(ident *rl.Identifier, frame *Frame) Refinement {
	sym, ok := tc.resolved.Uses[ident]
	if !ok || sym == nil {
		return EmptyRefinement()
	}
	baseType := tc.typeOfSymbol(sym, frame)
	nonNull := stripNullFrom(baseType)
	if nonNull == nil {
		return EmptyRefinement()
	}
	return Refinement{
		WhenTrue:  map[*Symbol]rl.TypingT{sym: nonNull},
		WhenFalse: map[*Symbol]rl.TypingT{},
	}
}

// interpretUnaryCondition handles `not <expr>`. Other unary ops
// (-, +) aren'\''t boolean and don'\''t produce refinements.
//
// Logical negation flips the refinement: what was the truthy
// narrowing becomes the falsy narrowing and vice versa. Refinement.
// Negate is exactly the swap.
func (tc *typeChecker) interpretUnaryCondition(c *rl.OpUnary, frame *Frame) Refinement {
	if c.Op != rl.OpNot {
		return EmptyRefinement()
	}
	return tc.interpretCondition(c.Operand, frame).Negate()
}

// interpretBinaryCondition handles `<expr> <op> <expr>` predicates.
// Each pattern detector returns "this predicate doesn't apply" via
// EmptyRefinement; the dispatcher tries them in order so the first
// matching shape wins. Order matters when shapes overlap - the more
// specific shape should appear first.
func (tc *typeChecker) interpretBinaryCondition(c *rl.OpBinary, frame *Frame) Refinement {
	switch c.Op {
	case rl.OpEq, rl.OpNeq:
		if id := identAgainstNullLiteral(c); id != nil {
			return tc.narrowNullEquality(c.Op, id, frame)
		}
		if ident, target, ok := typeOfPattern(c); ok {
			return tc.narrowTypeOfEquality(c.Op, ident, target, frame)
		}
		if ident, target, ok := identAgainstStringLiteral(c); ok {
			return tc.narrowStrEnumEquality(c.Op, ident, target, frame)
		}
	case rl.OpIn, rl.OpNotIn:
		if ident, values, ok := identInStringList(c); ok {
			return tc.narrowStrEnumIn(c.Op, ident, values, frame)
		}
	case rl.OpAnd:
		return tc.interpretAnd(c, frame)
	case rl.OpOr:
		return tc.interpretOr(c, frame)
	}
	return EmptyRefinement()
}

// interpretAnd handles `a and b`. Short-circuit evaluation drives
// the narrowing semantics:
//
//   - WhenTrue: both a and b evaluated truthy. The truthy narrowing
//     is a'\''s WhenTrue applied first, then b'\''s WhenTrue computed
//     against the a-truthy frame and merged on top. Latter wins on
//     conflicts since it'\''s computed in the tighter frame.
//
//   - WhenFalse: at least one of a, b is falsy. This is a disjunction
//     - either "a falsy" or "a truthy, b falsy" - and we don'\''t have
//     a way to express "OR of two refinements" without losing info.
//     Conservative: empty WhenFalse. Pyright takes the same shortcut.
//
// The right-hand side is interpreted with a'\''s truthy narrowing
// active: in `x != null and x > 5`, the `x > 5` is evaluated knowing
// x is non-null. That'\''s why we walk the right under
// frame.WithMany(leftRef.WhenTrue) before interpreting.
func (tc *typeChecker) interpretAnd(c *rl.OpBinary, frame *Frame) Refinement {
	leftRef := tc.interpretCondition(c.Left, frame)
	rightFrame := frame.WithMany(leftRef.WhenTrue)
	rightRef := tc.interpretCondition(c.Right, rightFrame)
	return Refinement{
		WhenTrue:  mergeRefinementMaps(leftRef.WhenTrue, rightRef.WhenTrue),
		WhenFalse: map[*Symbol]rl.TypingT{},
	}
}

// interpretOr handles `a or b`. Mirror image of and:
//
//   - WhenFalse: both falsy. Sequential apply: a'\''s WhenFalse first,
//     b'\''s WhenFalse computed against the a-falsy frame, then
//     merged.
//
//   - WhenTrue: at least one truthy. Disjunction, conservatively empty.
func (tc *typeChecker) interpretOr(c *rl.OpBinary, frame *Frame) Refinement {
	leftRef := tc.interpretCondition(c.Left, frame)
	rightFrame := frame.WithMany(leftRef.WhenFalse)
	rightRef := tc.interpretCondition(c.Right, rightFrame)
	return Refinement{
		WhenTrue:  map[*Symbol]rl.TypingT{},
		WhenFalse: mergeRefinementMaps(leftRef.WhenFalse, rightRef.WhenFalse),
	}
}

// mergeRefinementMaps overlays b on top of a. When both refine the
// same symbol, b'\''s narrowing wins because b was computed in the
// tighter frame (after a'\''s refinement was applied).
func mergeRefinementMaps(a, b map[*Symbol]rl.TypingT) map[*Symbol]rl.TypingT {
	out := make(map[*Symbol]rl.TypingT, len(a)+len(b))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		out[k] = v
	}
	return out
}

// narrowNullEquality is invoked after the dispatcher confirms the
// shape is `<ident> ==/!= null` (in either operand order). It looks
// up the symbol, peels the nullable component from the base type, and
// records both the non-null branch and the definitely-null branch.
// With TypingNullT in the system, the previously-no-op null side now
// narrows to TypingNullT - downstream `type_of(x) == "null"` dispatch
// and similar patterns see x as definitely null in the matching arm.
func (tc *typeChecker) narrowNullEquality(op rl.Operator, ident *rl.Identifier, frame *Frame) Refinement {
	sym, ok := tc.resolved.Uses[ident]
	if !ok || sym == nil {
		return EmptyRefinement()
	}
	baseType := tc.typeOfSymbol(sym, frame)
	if baseType == nil {
		return EmptyRefinement()
	}
	nonNull := stripNullFrom(baseType)
	if nonNull == nil {
		return EmptyRefinement()
	}
	nonNullBranch := map[*Symbol]rl.TypingT{sym: nonNull}
	nullBranch := map[*Symbol]rl.TypingT{sym: rl.NewNullType()}
	switch op {
	case rl.OpEq:
		return Refinement{WhenTrue: nullBranch, WhenFalse: nonNullBranch}
	case rl.OpNeq:
		return Refinement{WhenTrue: nonNullBranch, WhenFalse: nullBranch}
	}
	return EmptyRefinement()
}

// narrowTypeOfEquality is invoked after the dispatcher confirms the
// shape is `type_of(<ident>) ==/!= "<target>"` (in either operand
// order). target is the literal string compared against type_of's
// result.
//
// Narrowing semantics: split the base type into the arm(s) matching
// the target and the arm(s) that don't.
//   - For a union, partition the arms.
//   - For an optional, treat the inner as one component and (implicit)
//     null as another. Target "null" matches the null component (which
//     we can't represent in the truthy branch - falls back to no-op
//     there) and excludes it from the falsy branch.
//   - For a leaf type, the partition collapses: either all-truthy or
//     all-falsy.
//
// Unmatched branches narrow to Never. Phase 4e wires that into a
// "this branch is unreachable" diagnostic; we just compute the type
// here and let the consumer decide what to do with it.
func (tc *typeChecker) narrowTypeOfEquality(op rl.Operator, ident *rl.Identifier, target string, frame *Frame) Refinement {
	sym, ok := tc.resolved.Uses[ident]
	if !ok || sym == nil {
		return EmptyRefinement()
	}
	baseType := tc.typeOfSymbol(sym, frame)
	if baseType == nil {
		return EmptyRefinement()
	}
	if !validTypeOfTarget(target) {
		// type_of's runtime enum can't produce this string, so the
		// equality is statically false. Narrow truthy to Never; leave
		// falsy as base.
		return refinementForBranches(op, sym, rl.NewNeverType(), baseType)
	}
	truthy, falsy := narrowByTypeOf(baseType, target)
	if truthy == nil && falsy == nil {
		return EmptyRefinement()
	}
	return refinementForBranches(op, sym, truthy, falsy)
}

// refinementForBranches builds a Refinement from the computed truthy
// and falsy narrowed types. Either may be nil meaning "no narrowing
// on this branch"; we skip the map entry in that case so frame join
// later doesn'\''t union with a base type to produce the base type.
//
// op switches truthy/falsy: == puts the matching arm in WhenTrue; !=
// puts it in WhenFalse.
func refinementForBranches(op rl.Operator, sym *Symbol, truthy, falsy rl.TypingT) Refinement {
	whenTrue := map[*Symbol]rl.TypingT{}
	whenFalse := map[*Symbol]rl.TypingT{}
	switch op {
	case rl.OpEq:
		if truthy != nil {
			whenTrue[sym] = truthy
		}
		if falsy != nil {
			whenFalse[sym] = falsy
		}
	case rl.OpNeq:
		if truthy != nil {
			whenFalse[sym] = truthy
		}
		if falsy != nil {
			whenTrue[sym] = falsy
		}
	}
	return Refinement{WhenTrue: whenTrue, WhenFalse: whenFalse}
}

// typeOfPattern detects `type_of(<ident>) == "<target>"` and the
// swapped-operand variant. Returns the inspected identifier, the
// literal target string, and true on a match.
func typeOfPattern(c *rl.OpBinary) (*rl.Identifier, string, bool) {
	if ident, ok := typeOfCallOfIdent(c.Left); ok {
		if target, ok := simpleStringValue(c.Right); ok {
			return ident, target, true
		}
	}
	if ident, ok := typeOfCallOfIdent(c.Right); ok {
		if target, ok := simpleStringValue(c.Left); ok {
			return ident, target, true
		}
	}
	return nil, "", false
}

// typeOfCallOfIdent matches exactly `type_of(<ident>)`. Anything more
// elaborate (chained call, multiple args, named args, non-identifier
// argument) falls through - those don'\''t cleanly correspond to a
// narrowable path.
func typeOfCallOfIdent(n rl.Node) (*rl.Identifier, bool) {
	call, ok := n.(*rl.Call)
	if !ok {
		return nil, false
	}
	fn, ok := call.Func.(*rl.Identifier)
	if !ok || fn.Name != "type_of" {
		return nil, false
	}
	if len(call.Args) != 1 || len(call.NamedArgs) != 0 {
		return nil, false
	}
	ident, ok := call.Args[0].(*rl.Identifier)
	if !ok {
		return nil, false
	}
	return ident, true
}

// narrowStrEnumEquality handles `<ident> ==/!= "<value>"` when the
// identifier'\''s base type is a string-enum. Truthy keeps the matching
// value (a single-valued enum); falsy keeps the rest. Either side can
// collapse to Never if the partition leaves nothing.
//
// Plain `str` is intentionally NOT narrowed - "x is the literal 'foo'"
// has no static expression today (no singleton types), and narrowing
// to a freshly-minted StrEnum<"foo"> would surprise users who
// declared the variable str.
func (tc *typeChecker) narrowStrEnumEquality(op rl.Operator, ident *rl.Identifier, target string, frame *Frame) Refinement {
	sym, ok := tc.resolved.Uses[ident]
	if !ok || sym == nil {
		return EmptyRefinement()
	}
	baseType := tc.typeOfSymbol(sym, frame)
	enum, ok := baseType.(*rl.TypingStrEnumT)
	if !ok {
		return EmptyRefinement()
	}
	truthy, falsy := partitionStrEnum(enum, map[string]bool{target: true})
	return refinementForBranches(op, sym, truthy, falsy)
}

// narrowStrEnumIn handles `<ident> in [<vals>]` and `<ident> not in
// [<vals>]`. The list literal must contain only simple string
// literals; mixed or non-literal contents disqualify the pattern.
//
// Two base-type shapes:
//
//   - Base is already a string-enum: partition into in-set (truthy)
//     and out-of-set (falsy) values. For `in`, truthy keeps the
//     intersection; for `not in`, the sides swap.
//
//   - Base is plain `str`: there's no enum to partition, but the
//     truthy branch is more precise than `str` - x is provably one
//     of the listed values. Narrow truthy to the new enum;
//     leave falsy at the base type (a plain `str` minus a finite
//     set is still effectively `str`, no useful narrowing).
func (tc *typeChecker) narrowStrEnumIn(op rl.Operator, ident *rl.Identifier, values []string, frame *Frame) Refinement {
	sym, ok := tc.resolved.Uses[ident]
	if !ok || sym == nil {
		return EmptyRefinement()
	}
	baseType := tc.typeOfSymbol(sym, frame)
	if enum, isEnum := baseType.(*rl.TypingStrEnumT); isEnum {
		set := make(map[string]bool, len(values))
		for _, v := range values {
			set[v] = true
		}
		truthy, falsy := partitionStrEnum(enum, set)
		switch op {
		case rl.OpIn:
			return refinementForBranches(rl.OpEq, sym, truthy, falsy)
		case rl.OpNotIn:
			return refinementForBranches(rl.OpNeq, sym, truthy, falsy)
		}
		return EmptyRefinement()
	}
	if _, isStr := baseType.(*rl.TypingStrT); isStr {
		enumT := rl.NewStrEnumType(dedupeStrings(values)...)
		switch op {
		case rl.OpIn:
			return Refinement{
				WhenTrue:  map[*Symbol]rl.TypingT{sym: enumT},
				WhenFalse: map[*Symbol]rl.TypingT{},
			}
		case rl.OpNotIn:
			return Refinement{
				WhenTrue:  map[*Symbol]rl.TypingT{},
				WhenFalse: map[*Symbol]rl.TypingT{sym: enumT},
			}
		}
	}
	return EmptyRefinement()
}

// dedupeStrings returns the input slice with duplicates removed,
// preserving first-occurrence order. Used so `x in ["a", "a"]`
// narrows to ["a"] rather than ["a", "a"].
func dedupeStrings(in []string) []string {
	seen := make(map[string]bool, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}

// partitionStrEnum splits an enum into the values present in keepSet
// (truthy) and the rest (falsy). Either side collapses to Never if
// the partition leaves no values - that signals an unreachable branch
// the if/else wiring can pick up later.
func partitionStrEnum(enum *rl.TypingStrEnumT, keepSet map[string]bool) (truthy, falsy rl.TypingT) {
	var truthyV, falsyV []string
	for _, v := range enum.Values() {
		if keepSet[v] {
			truthyV = append(truthyV, v)
		} else {
			falsyV = append(falsyV, v)
		}
	}
	if len(truthyV) == 0 {
		truthy = rl.NewNeverType()
	} else {
		truthy = rl.NewStrEnumType(truthyV...)
	}
	if len(falsyV) == 0 {
		falsy = rl.NewNeverType()
	} else {
		falsy = rl.NewStrEnumType(falsyV...)
	}
	return
}

// identAgainstStringLiteral matches `<ident> == "<str>"` (or swapped).
// Returns the identifier, the literal value, and true. The string
// must be a simple (non-interpolated) literal.
func identAgainstStringLiteral(c *rl.OpBinary) (*rl.Identifier, string, bool) {
	if id, ok := c.Left.(*rl.Identifier); ok {
		if s, ok := simpleStringValue(c.Right); ok {
			return id, s, true
		}
	}
	if id, ok := c.Right.(*rl.Identifier); ok {
		if s, ok := simpleStringValue(c.Left); ok {
			return id, s, true
		}
	}
	return nil, "", false
}

// identInStringList matches `<ident> in/not in [<str>, <str>, ...]`.
// Every list element must be a simple string literal; any other
// content disqualifies the pattern. Returns the identifier, the
// list of literal values, and true on match.
func identInStringList(c *rl.OpBinary) (*rl.Identifier, []string, bool) {
	ident, ok := c.Left.(*rl.Identifier)
	if !ok {
		return nil, nil, false
	}
	list, ok := c.Right.(*rl.LitList)
	if !ok {
		return nil, nil, false
	}
	values := make([]string, 0, len(list.Elements))
	for _, e := range list.Elements {
		s, ok := simpleStringValue(e)
		if !ok {
			return nil, nil, false
		}
		values = append(values, s)
	}
	return ident, values, true
}

// simpleStringValue returns the literal value of a non-interpolated
// string literal, or false for anything else. Interpolated strings
// could in principle be constant-folded, but the static checker
// doesn'\''t do constant folding and the catalog is intentionally
// shape-driven.
func simpleStringValue(n rl.Node) (string, bool) {
	s, ok := n.(*rl.LitString)
	if !ok || !s.Simple {
		return "", false
	}
	return s.Value, true
}

// identAgainstNullLiteral returns the identifier operand of a binary
// expression whose other operand is a null literal, or nil if the
// shape doesn't match. Handles `<ident> == null` and `null == <ident>`
// equivalently.
func identAgainstNullLiteral(c *rl.OpBinary) *rl.Identifier {
	if id, ok := c.Left.(*rl.Identifier); ok {
		if _, ok := c.Right.(*rl.LitNull); ok {
			return id
		}
	}
	if id, ok := c.Right.(*rl.Identifier); ok {
		if _, ok := c.Left.(*rl.LitNull); ok {
			return id
		}
	}
	return nil
}

// typeOfSymbol returns the symbol's type at this program point: the
// frame narrowing if any, the recorded inferred/declared type
// otherwise. Returns nil only when the checker has no record at all
// (e.g. a forward reference); callers should treat that as Dynamic.
func (tc *typeChecker) typeOfSymbol(sym *Symbol, frame *Frame) rl.TypingT {
	if frame != nil {
		if t, ok := frame.Lookup(sym); ok {
			return t
		}
	}
	if t, ok := tc.info.SymbolTypes[sym]; ok {
		return t
	}
	return nil
}

// validTypeOfTarget reports whether s is one of the strings that
// type_of() can actually return. Anything else can't possibly equal
// type_of(x) at runtime, so the equality is statically false - the
// caller turns that into a Never truthy branch.
func validTypeOfTarget(s string) bool {
	switch s {
	case "int", "str", "float", "bool", "list", "map", "null", "error", "function":
		return true
	}
	return false
}

// matchesTypeOf reports whether a runtime value of static type t
// would satisfy `type_of(value) == target`. The map mirrors what
// TypeAsString in core/utils.go does at runtime.
//
// "null" never matches a non-nullable static type; nullable types are
// handled by the caller (narrowByTypeOf splits Optional<T> into the
// inner and the implicit null component before consulting this).
func matchesTypeOf(t rl.TypingT, target string) bool {
	switch target {
	case "int":
		_, ok := t.(*rl.TypingIntT)
		return ok
	case "str":
		switch t.(type) {
		case *rl.TypingStrT, *rl.TypingStrEnumT:
			return true
		}
	case "float":
		_, ok := t.(*rl.TypingFloatT)
		return ok
	case "bool":
		_, ok := t.(*rl.TypingBoolT)
		return ok
	case "list":
		switch t.(type) {
		case *rl.TypingAnyListT, *rl.TypingListT, *rl.TypingTupleT:
			return true
		}
	case "map":
		switch t.(type) {
		case *rl.TypingAnyMapT, *rl.TypingMapT, *rl.TypingStructT:
			return true
		}
	case "error":
		_, ok := t.(*rl.TypingErrorT)
		return ok
	case "function":
		_, ok := t.(*rl.TypingFnT)
		return ok
	case "null":
		_, ok := t.(*rl.TypingNullT)
		return ok
	}
	return false
}

// narrowByTypeOf splits base into the portion(s) matching the
// type_of-target string and the portion(s) that don't. Returns
// (truthy, falsy) where either may be a TypingNeverT meaning "this
// side is empty - this branch is unreachable". A nil result on
// either side means "we have no static handle on this side"; the
// caller treats it as "no narrowing applied" rather than "Never".
//
// Decision table:
//   - Any / Dynamic / ErrorType base: (nil, nil) - we can'\''t prove
//     either side against a fully-open or poisoned type.
//   - Union: partition arms by recursive call. Drop Never on each
//     side (those arms contribute nothing). Preserve nil arms by
//     passing the arm through unchanged - "no static handle" on
//     this arm means we keep the whole arm as a fallback.
//   - Optional<T>:
//       target == "null":
//         truthy: TypingNullT (definite - the null arm matched).
//         falsy:  x is non-null - return T.
//       inner matches target:
//         truthy: T (the non-null component).
//         falsy:  TypingNullT (only possibility left).
//       inner doesn'\''t match target:
//         truthy: Never (inner doesn'\''t match, null doesn'\''t
//                 match any non-null target).
//         falsy:  the original Optional<T> stays.
//   - TypingNullT (definite-null leaf):
//       target == "null":  truthy=null, falsy=Never.
//       any other target:  truthy=Never, falsy=null.
//   - Leaf (non-Optional, non-null):
//       target == "null": truthy=Never, falsy=base.
//       matches:           truthy=base,  falsy=Never.
//       doesn'\''t match:   truthy=Never, falsy=base.
func narrowByTypeOf(base rl.TypingT, target string) (truthy, falsy rl.TypingT) {
	if base == nil {
		return nil, nil
	}
	switch base.(type) {
	case *rl.TypingAnyT, *rl.TypingDynamicT, *rl.TypingErrorTypeT:
		return nil, nil
	}
	if u, ok := base.(*rl.TypingUnionT); ok {
		var truthyArms, falsyArms []rl.TypingT
		for _, arm := range u.Types() {
			tA, fA := narrowByTypeOf(arm, target)
			// Drop Never contributions; preserve actual types.
			// nil from a recursive call means "this arm has no
			// static handle for either side." Pass the arm through
			// unchanged so the union join doesn'\''t lose it.
			if tA == nil {
				truthyArms = append(truthyArms, arm)
			} else if !isNeverType(tA) {
				truthyArms = append(truthyArms, tA)
			}
			if fA == nil {
				falsyArms = append(falsyArms, arm)
			} else if !isNeverType(fA) {
				falsyArms = append(falsyArms, fA)
			}
		}
		return joinNarrowArms(truthyArms), joinNarrowArms(falsyArms)
	}
	if o, ok := base.(*rl.TypingOptionalT); ok {
		inner := o.Inner()
		if target == "null" {
			// Truthy: x IS null - return TypingNullT (definite).
			// Falsy: x is non-null = the inner type.
			return rl.NewNullType(), inner
		}
		if matchesTypeOf(inner, target) {
			// Truthy: non-null inner matched. Falsy: only possibility
			// left is null.
			return inner, rl.NewNullType()
		}
		// Inner doesn'\''t match target and null doesn'\''t match any
		// non-null target. Truthy is empty; falsy is the original.
		return rl.NewNeverType(), base
	}
	if _, ok := base.(*rl.TypingNullT); ok {
		if target == "null" {
			return base, rl.NewNeverType()
		}
		return rl.NewNeverType(), base
	}
	if target == "null" {
		return rl.NewNeverType(), base
	}
	if matchesTypeOf(base, target) {
		return base, rl.NewNeverType()
	}
	return rl.NewNeverType(), base
}

// isNeverType reports whether t is the Never (bottom) type. Used by
// the union partitioner to drop "nothing on this side" contributions
// from per-arm narrowing.
func isNeverType(t rl.TypingT) bool {
	_, ok := t.(*rl.TypingNeverT)
	return ok
}

// joinNarrowArms collapses a slice of narrowed arms into a single
// TypingT: empty -> Never (no remaining arms means the branch is
// unreachable), one arm -> that arm, more -> a union. Two-arm unions
// with a null component collapse to T? to match how users spell
// nullable types and how unionTypesForJoin canonicalises inferred
// returns.
func joinNarrowArms(arms []rl.TypingT) rl.TypingT {
	switch len(arms) {
	case 0:
		return rl.NewNeverType()
	case 1:
		return arms[0]
	case 2:
		if _, leftNull := arms[0].(*rl.TypingNullT); leftNull {
			if _, rightNull := arms[1].(*rl.TypingNullT); !rightNull {
				return rl.NewOptionalType(arms[1])
			}
		}
		if _, rightNull := arms[1].(*rl.TypingNullT); rightNull {
			return rl.NewOptionalType(arms[0])
		}
		return rl.NewUnionType(arms...)
	default:
		return rl.NewUnionType(arms...)
	}
}

// stripErrorFrom returns t with any error component removed. For a
// bare TypingErrorT, returns Never (nothing left). For a union, drops
// the error arms and returns the rest. Returns nil if t has no error
// component to subtract.
//
// Used by `expr catch fallback`: the result type is (left - error) |
// right - the catch only fires on error, and the non-error portion of
// left is what flows through when no catch is needed.
func stripErrorFrom(t rl.TypingT) rl.TypingT {
	if t == nil {
		return nil
	}
	if _, ok := t.(*rl.TypingErrorT); ok {
		return rl.NewNeverType()
	}
	if u, ok := t.(*rl.TypingUnionT); ok {
		arms := u.Types()
		out := make([]rl.TypingT, 0, len(arms))
		changed := false
		for _, arm := range arms {
			if _, isErr := arm.(*rl.TypingErrorT); isErr {
				changed = true
				continue
			}
			out = append(out, arm)
		}
		if !changed {
			return nil
		}
		switch len(out) {
		case 0:
			return rl.NewNeverType()
		case 1:
			return out[0]
		default:
			return rl.NewUnionType(out...)
		}
	}
	return nil
}

// extractErrorFrom returns the error component of t - the dual of
// stripErrorFrom. For a bare TypingErrorT, returns it. For a union,
// returns the first ErrorT arm. Returns nil if t has no error
// component; callers narrowing INTO an error branch should fall back
// to a fresh TypingErrorT (the runtime guarantees the RHS errored
// to land in the catch).
func extractErrorFrom(t rl.TypingT) rl.TypingT {
	if t == nil {
		return nil
	}
	if _, ok := t.(*rl.TypingErrorT); ok {
		return t
	}
	if u, ok := t.(*rl.TypingUnionT); ok {
		for _, arm := range u.Types() {
			if _, ok := arm.(*rl.TypingErrorT); ok {
				return arm
			}
		}
	}
	return nil
}

// stripNullFrom returns t with any null component removed. For
// Optional<T> it returns T. For a union whose arms are individually
// optional, each arm is stripped. Returns nil if t has no nullable
// component to subtract - the caller treats that as "predicate
// gives us no narrowing here."
func stripNullFrom(t rl.TypingT) rl.TypingT {
	if t == nil {
		return nil
	}
	if o, ok := t.(*rl.TypingOptionalT); ok {
		return o.Inner()
	}
	if u, ok := t.(*rl.TypingUnionT); ok {
		arms := u.Types()
		out := make([]rl.TypingT, 0, len(arms))
		changed := false
		for _, arm := range arms {
			if s := stripNullFrom(arm); s != nil {
				out = append(out, s)
				changed = true
			} else {
				out = append(out, arm)
			}
		}
		if !changed {
			return nil
		}
		if len(out) == 1 {
			return out[0]
		}
		return rl.NewUnionType(out...)
	}
	return nil
}

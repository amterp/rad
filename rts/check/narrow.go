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
	}
	return EmptyRefinement()
}

// interpretBinaryCondition handles `<expr> <op> <expr>` predicates.
// Today: equality / inequality against null. Future sub-commits add
// equality with string literals (string-enum narrowing) and `in`/
// `not in` against list literals.
func (tc *typeChecker) interpretBinaryCondition(c *rl.OpBinary, frame *Frame) Refinement {
	switch c.Op {
	case rl.OpEq, rl.OpNeq:
		return tc.interpretEqualityCondition(c, frame)
	}
	return EmptyRefinement()
}

// interpretEqualityCondition recognizes `<ident> == null`, `<ident> !=
// null`, and the swapped-operand variants. Either operand may be the
// identifier; equality is symmetric and so is the predicate.
//
// Truthy / falsy assignment:
//   - `x == null` truthy: x is null (no narrowing recorded yet - we
//     have no static "Null" type), falsy: x is non-null.
//   - `x != null` is the inverse.
//
// The "narrow x to null" side is intentionally a no-op for now. We
// don't have a TypingNullT in the static system, and the practical
// payoff of "x is null in this branch" is small (users rarely access
// members of a definitely-null value). Adding the type later is a
// pure expansion: existing scripts keep working, the null branch
// just gains a tighter static answer.
func (tc *typeChecker) interpretEqualityCondition(c *rl.OpBinary, frame *Frame) Refinement {
	ident := identAgainstNullLiteral(c)
	if ident == nil {
		return EmptyRefinement()
	}
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
		// Base type has no nullable component to subtract; the
		// predicate can't refine either side meaningfully.
		return EmptyRefinement()
	}
	nonNullBranch := map[*Symbol]rl.TypingT{sym: nonNull}
	switch c.Op {
	case rl.OpEq:
		return Refinement{WhenTrue: map[*Symbol]rl.TypingT{}, WhenFalse: nonNullBranch}
	case rl.OpNeq:
		return Refinement{WhenTrue: nonNullBranch, WhenFalse: map[*Symbol]rl.TypingT{}}
	}
	return EmptyRefinement()
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

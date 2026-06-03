// Package radfmt implements `rad fmt`, a gofmt-style canonical re-printer for
// Rad scripts. It operates on the tree-sitter CST (not the typed AST, which
// drops comments) and is structured as two cores:
//
//   - a Doc IR + render machine (doc.go, render.go) ported closely from
//     Prettier's productionized Wadler printer, and
//   - construct formatters that turn CST nodes into Docs (printer.go + friends).
//
// The Doc IR is the engine: construct formatters only ever *build* Docs, and a
// single width-aware render machine decides where lines break. See DESIGN.md
// for the full reference (Doc semantics, fits, comment attachment).
package radfmt

// Doc is the sealed document IR. Every node kind implements isDoc. Construct
// formatters assemble a Doc tree; render.go turns it into a string, choosing
// flat vs broken layout for each Group against the target width.
type Doc interface{ isDoc() }

// GroupID identifies a Group so that IfBreak / IndentIfBreak can react to
// whether that specific group broke. The zero value means "no id".
type GroupID uint32

// Text is literal output. It must not contain a newline - line breaks are only
// ever produced by Line nodes, so the render machine can track columns.
type Text struct{ S string }

// Line is a breakable position. Flat: Soft renders as "", otherwise " ". Broken:
// a newline followed by the current indent. Hard forces a break regardless of
// fit and (via propagateBreaks) breaks every enclosing Group. Literal is a hard
// break that emits no indent (used for content that must keep column 0).
type Line struct {
	Soft    bool
	Hard    bool
	Literal bool
}

// Concat prints its parts in order.
type Concat struct{ Parts []Doc }

// Indent increases the indentation of its contents by one level.
type Indent struct{ Contents Doc }

// Align increases indentation by a fixed number of spaces (for alignment that
// isn't a whole indent level).
type Align struct {
	N        int
	Contents Doc
}

// Group is the unit of layout choice: the render machine tries to print
// Contents flat, and if that doesn't fit the remaining width it prints broken.
// Break (set by propagateBreaks when Contents contains a hard break, or up front
// via the Group constructor's shouldBreak) forces broken mode without measuring.
// ExpandedStates, when non-nil, makes this a conditional group (see render.go).
//
// Group is referenced by pointer so propagateBreaks can set Break in place.
type Group struct {
	Contents       Doc
	Break          bool
	ID             GroupID
	ExpandedStates []Doc
}

// Fill is a flow layout: each separator independently breaks only if the next
// content doesn't fit. Used for word-wrap-like flows.
type Fill struct{ Parts []Doc }

// IfBreak prints BreakContents when its controlling group broke, else
// FlatContents. The controlling group is GroupID if set, otherwise the nearest
// enclosing group.
type IfBreak struct {
	BreakContents Doc
	FlatContents  Doc
	GroupID       GroupID
}

// IndentIfBreak indents Contents only when the referenced group broke.
type IndentIfBreak struct {
	Contents Doc
	GroupID  GroupID
}

// LineSuffix defers its contents to the end of the current line - it's buffered
// and flushed just before the next newline. This is how trailing comments
// attach without participating in width measurement of the code before them.
type LineSuffix struct{ Contents Doc }

// LineSuffixBoundary forces any buffered LineSuffix to flush even without a
// newline (by injecting a hard break if the buffer is non-empty).
type LineSuffixBoundary struct{}

// BreakParent forces every enclosing Group to break. hardline carries one
// implicitly. It's a no-op at print time; propagateBreaks consumes it.
type BreakParent struct{}

// Trim removes trailing whitespace already emitted on the current line.
type Trim struct{}

func (Text) isDoc()               {}
func (Line) isDoc()               {}
func (Concat) isDoc()             {}
func (Indent) isDoc()             {}
func (Align) isDoc()              {}
func (*Group) isDoc()             {}
func (Fill) isDoc()               {}
func (IfBreak) isDoc()            {}
func (IndentIfBreak) isDoc()      {}
func (LineSuffix) isDoc()         {}
func (LineSuffixBoundary) isDoc() {}
func (BreakParent) isDoc()        {}
func (Trim) isDoc()               {}

// --- Constructors -----------------------------------------------------------
//
// These keep call sites readable and centralize a few normalizations (e.g.
// flattening nested concats, dropping nils).

// text returns a literal-text Doc.
func text(s string) Doc { return Text{S: s} }

// softLine is "" when flat, a newline when broken.
func softLine() Doc { return Line{Soft: true} }

// lineOrSpace is " " when flat, a newline when broken.
func lineOrSpace() Doc { return Line{} }

// hardLine always breaks and forces enclosing groups to break.
func hardLine() Doc { return Line{Hard: true} }

// literalLine is a hard break that emits no indentation.
func literalLine() Doc { return Line{Hard: true, Literal: true} }

// concat builds a Concat, flattening nested Concats and skipping nil parts so
// callers can pass optional pieces without guarding each one.
func concat(parts ...Doc) Doc {
	flat := make([]Doc, 0, len(parts))
	for _, p := range parts {
		switch p := p.(type) {
		case nil:
			continue
		case Concat:
			flat = append(flat, p.Parts...)
		default:
			flat = append(flat, p)
		}
	}
	if len(flat) == 1 {
		return flat[0]
	}
	return Concat{Parts: flat}
}

// indent increases the indent level of d by one.
func indent(d Doc) Doc { return Indent{Contents: d} }

// group wraps d so the renderer chooses flat-or-broken to fit the width. The
// returned value holds a *Group, so propagateBreaks can set Break in place.
func group(d Doc) Doc { return &Group{Contents: d} }

// ifBreak prints brk when the enclosing group broke, else flat.
func ifBreak(brk, flat Doc) Doc { return IfBreak{BreakContents: brk, FlatContents: flat} }

// lineSuffix defers d to end-of-line.
func lineSuffix(d Doc) Doc { return LineSuffix{Contents: d} }

// join interleaves sep between docs.
func join(sep Doc, docs []Doc) Doc {
	if len(docs) == 0 {
		return Concat{}
	}
	parts := make([]Doc, 0, len(docs)*2-1)
	for i, d := range docs {
		if i > 0 {
			parts = append(parts, sep)
		}
		parts = append(parts, d)
	}
	return Concat{Parts: parts}
}

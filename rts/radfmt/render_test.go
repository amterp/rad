package radfmt

import "testing"

// callDoc builds the canonical Prettier call-argument shape: try to fit the args
// on one line, otherwise break one-per-line with a trailing comma.
func callDoc(fn string, args []Doc) Doc {
	return group(concat(
		text(fn+"("),
		indent(concat(
			softLine(),
			join(concat(text(","), lineOrSpace()), args),
			ifBreak(text(","), text("")),
		)),
		softLine(),
		text(")"),
	))
}

// Worked example A (research.md §1.7): a call-argument list fits flat when short
// and breaks one-per-line with a trailing comma when it exceeds the width.
func TestRender_CallArgs_FlatVsBroken(t *testing.T) {
	short := callDoc("foo", []Doc{text("a()"), text("b()")})
	if got, want := PrintDocToString(short, MaxWidth), "foo(a(), b())"; got != want {
		t.Errorf("short call: got %q, want %q", got, want)
	}

	long := callDoc("foo", []Doc{
		text("reallyLongArg()"),
		text("omgSoManyParameters()"),
		text("IShouldRefactorThis()"),
		text("isThereSeriouslyAnotherOne()"),
	})
	// The flat form is 96 cols, so it fits at 100 but not at 80; at 80 the
	// group breaks one-arg-per-line with a trailing comma.
	if got, want := PrintDocToString(long, MaxWidth), "foo(reallyLongArg(), omgSoManyParameters(), IShouldRefactorThis(), isThereSeriouslyAnotherOne())"; got != want {
		t.Errorf("long call at 100 should stay flat:\n got: %q\nwant: %q", got, want)
	}
	want := "foo(\n" +
		"    reallyLongArg(),\n" +
		"    omgSoManyParameters(),\n" +
		"    IShouldRefactorThis(),\n" +
		"    isThereSeriouslyAnotherOne(),\n" +
		")"
	if got := PrintDocToString(long, 80); got != want {
		t.Errorf("long call at 80:\n got: %q\nwant: %q", got, want)
	}
}

// Worked example B (research.md §1.8): a trailing comment buffered via lineSuffix
// flushes before the newline, landing at the end of the line after the code.
func TestRender_LineSuffix_TrailingComment(t *testing.T) {
	doc := concat(text("a"), lineSuffix(text(" // comment")), text(";"), hardLine())
	if got, want := PrintDocToString(doc, MaxWidth), "a; // comment\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// A hardline anywhere inside a group forces that group (and its ancestors) to
// render broken, regardless of width.
func TestRender_HardlinePropagatesBreak(t *testing.T) {
	doc := group(concat(text("{"), indent(concat(hardLine(), text("x"))), hardLine(), text("}")))
	if got, want := PrintDocToString(doc, MaxWidth), "{\n    x\n}"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// softline is nothing when flat and a newline when broken; line is a space when
// flat. This pins the core Line semantics.
func TestRender_LineSemantics(t *testing.T) {
	flat := group(concat(text("("), softLine(), text("x"), softLine(), text(")")))
	if got, want := PrintDocToString(flat, MaxWidth), "(x)"; got != want {
		t.Errorf("flat softline: got %q, want %q", got, want)
	}

	// Force a break with a width of 1 so the group can't fit flat.
	broken := group(concat(text("("), indent(concat(softLine(), text("xxxxx"))), softLine(), text(")")))
	if got, want := PrintDocToString(broken, 3), "(\n    xxxxx\n)"; got != want {
		t.Errorf("broken softline: got %q, want %q", got, want)
	}
}

// Rendering is deterministic: same doc, same width, same output every time.
func TestRender_Deterministic(t *testing.T) {
	build := func() Doc {
		return callDoc("foo", []Doc{text("alpha"), text("beta"), text("gamma")})
	}
	a := PrintDocToString(build(), 12)
	b := PrintDocToString(build(), 12)
	if a != b {
		t.Errorf("non-deterministic render:\n a: %q\n b: %q", a, b)
	}
}

// Trailing whitespace before a hard break is trimmed.
func TestRender_TrimsTrailingWhitespaceBeforeBreak(t *testing.T) {
	doc := concat(text("a"), text("  "), hardLine(), text("b"))
	if got, want := PrintDocToString(doc, MaxWidth), "a\nb"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

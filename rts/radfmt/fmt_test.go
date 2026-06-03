package radfmt

import "testing"

// Pipeline-spine guarantee: while every construct is still verbatim, Format must
// return valid source byte-for-byte unchanged (changed=false, ok=true). This is
// the safety net that lets us grow construct formatters incrementally.
func TestFormat_SpineRoundTrips(t *testing.T) {
	cases := []string{
		"a = 1\n",
		"#!/usr/bin/env rad\nprint(\"hi\")\n",
		"args:\n    name str\n    age int = 30 # An age.\n\nprint(name)\n",
		"// leading comment\nx = 1 // trailing\n",
		"a = {\"x\": 1, y: 2}\nb = [1, 2, 3]\n",
	}
	for _, src := range cases {
		out, changed, ok := Format(src)
		if !ok {
			t.Errorf("ok=false for valid source:\n%q", src)
			continue
		}
		if out != src {
			t.Errorf("spine altered source:\n in: %q\nout: %q", src, out)
		}
		if changed {
			t.Errorf("changed=true but spine should be verbatim:\n%q", src)
		}
	}
}

// A file whose only problem is CRLF line endings must be reported as changed
// (and normalized to LF), not silently passed over - changed is compared against
// the original input, not the normalized form.
func TestFormat_CRLFOnlyIsChanged(t *testing.T) {
	out, changed, ok := Format("x = 1\r\ny = 2\r\n")
	if !ok {
		t.Fatal("ok=false for valid CRLF source")
	}
	if !changed {
		t.Error("CRLF-only difference should report changed=true")
	}
	if out != "x = 1\ny = 2\n" {
		t.Errorf("CRLF not normalized to LF: %q", out)
	}
}

// Safety: source with syntax errors must be returned unchanged with ok=false -
// the formatter never touches code it couldn't fully parse.
func TestFormat_InvalidSourceIsNoOp(t *testing.T) {
	bad := []string{
		"a = = 1\n",
		"if x\n", // missing colon / body
		"print(\n",
	}
	for _, src := range bad {
		out, changed, ok := Format(src)
		if ok {
			t.Errorf("expected ok=false for invalid source: %q", src)
		}
		if out != src || changed {
			t.Errorf("invalid source should be an unchanged no-op: in=%q out=%q changed=%v", src, out, changed)
		}
	}
}

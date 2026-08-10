package check

import "testing"

// The value of RAD40016 is entirely in its precision: a warning that fires on
// ordinary literal separators trains people to ignore it, and then the real
// migration hits slide past. These cases pin both halves - what must fire, and
// what must stay quiet.
func TestRegexSignal(t *testing.T) {
	fires := []struct {
		pattern string
		signal  string
	}{
		{`\s+`, `\s`},
		{`\d`, `\d`},
		{`\bword\b`, `\b`},
		{`\.\n`, `\.`},
		{`a\+b`, `\+`},
		{`[^a-z]`, "[...]"},
		{`(.*) Brown`, "*"},
		{`(^|\s)`, "|"},
		{`^the `, "^"},
		{`\s+#.*$`, `\s`},
		{`\$5`, `\$`},
		{`.* - `, "*"},
		{`a+b`, "+"},
		{`colou?r`, "?"},
		{`cat|dog`, "|"},
		{`ends$`, "$"},
		{`x{3}`, "{3}"},
		{`x{2,}`, "{2,}"},
		{`x{2,5}`, "{2,5}"},
	}
	for _, tc := range fires {
		t.Run("fires/"+tc.pattern, func(t *testing.T) {
			signal, found := regexSignal(tc.pattern)
			if !found {
				t.Fatalf("regexSignal(%q) found nothing, want %q", tc.pattern, tc.signal)
			}
			if signal != tc.signal {
				t.Errorf("regexSignal(%q) = %q, want %q", tc.pattern, signal, tc.signal)
			}
		})
	}

	// Every entry here is something people genuinely split and replace on. If
	// one of them starts warning, the rule has gone too broad.
	quiet := []string{
		"",
		",",
		".",
		"|",
		"+",
		"*",
		"?",
		"^",
		"$",
		"/",
		"-",
		"::",
		".txt",
		"1.2.3",
		"a.b.c",
		"key=value",
		"C:\\Users",
		"=> ",
		"|start",    // trailing branch only
		"end|",      // leading branch only
		"$5",        // a price, not an anchor
		"[",         // an unclosed bracket is just a bracket
		"]",         //
		"[]",        // ...and an empty class isn't a class
		"(",         //
		")",         //
		"a}b{c",     // braces out of order
		"{}",        // no repetition count
		"{abc}",     //
		"100% x2",   //
		"# heading", //
	}
	for _, pattern := range quiet {
		t.Run("quiet/"+pattern, func(t *testing.T) {
			if signal, found := regexSignal(pattern); found {
				t.Errorf("regexSignal(%q) fired on %q, want quiet", pattern, signal)
			}
		})
	}
}

func TestRepetitionWidth(t *testing.T) {
	cases := map[string]int{
		"{3}":    3,
		"{12}":   4,
		"{2,}":   4,
		"{2,5}":  5,
		"{2,50}": 6,
		"{}":     0,
		"{a}":    0,
		"{3":     0,
		"{3,":    0,
		"{3,a}":  0,
	}
	for input, want := range cases {
		t.Run(input, func(t *testing.T) {
			if got := repetitionWidth(input); got != want {
				t.Errorf("repetitionWidth(%q) = %d, want %d", input, got, want)
			}
		})
	}
}

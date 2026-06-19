package docir

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestEmit_Callout(t *testing.T) {
	src := lines(
		`!!! info "Why 3 delimiters?"`,
		"",
		"    Some prose.",
		"",
		"    ```rad",
		`    x = 1`,
		"    ```",
	)
	want := lines(
		"**Info: Why 3 delimiters?**",
		"",
		"    Some prose.",
		"",
		"    ```rad",
		"    x = 1",
		"    ```",
		"",
	)
	if got := EmitTerminal(Parse(src)); got != want {
		t.Fatalf("callout mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestEmit_Tabs(t *testing.T) {
	src := lines(
		`=== "Bash (~/.bashrc)"`,
		"",
		"    ```shell",
		`    eval "$(rad completion bash)"`,
		"    ```",
		"",
		`=== "Zsh (~/.zshrc)"`,
		"",
		"    ```shell",
		`    eval "$(rad completion zsh)"`,
		"    ```",
	)
	want := lines(
		"**Bash (~/.bashrc)**",
		"",
		"    ```shell",
		`    eval "$(rad completion bash)"`,
		"    ```",
		"",
		"**Zsh (~/.zshrc)**",
		"",
		"    ```shell",
		`    eval "$(rad completion zsh)"`,
		"    ```",
		"",
	)
	if got := EmitTerminal(Parse(src)); got != want {
		t.Fatalf("tabs mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestEmit_ResultCodeDropsDivTags(t *testing.T) {
	src := lines(
		"```rad",
		`print("Hi")`,
		"```",
		"",
		`<div class="result">`,
		"```",
		"Hi",
		"```",
		"</div>",
	)
	got := EmitTerminal(Parse(src))
	if strings.Contains(got, "<div") || strings.Contains(got, "</div>") {
		t.Fatalf("div tags leaked:\n%s", got)
	}
	want := lines(
		"```rad",
		`print("Hi")`,
		"```",
		"",
		"```",
		"Hi",
		"```",
		"",
	)
	if got != want {
		t.Fatalf("result mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

// TestEmit_TableAligns checks alignment robustly without a brittle
// fixed golden: every rendered line is the same visual width, the
// column count is preserved, and the literal pipe inside backticks
// survives.
func TestEmit_TableAligns(t *testing.T) {
	src := lines(
		"| Parameter | Type | Description |",
		"|-----------|------|-------------|",
		"| `val`     | `int | float`      | Value to constrain |",
		"| `min`     | `int | float`      | Minimum bound      |",
	)
	got := strings.TrimRight(EmitTerminal(Parse(src)), "\n")
	rows := strings.Split(got, "\n")
	if len(rows) != 4 { // header + separator + 2 data rows
		t.Fatalf("want 4 rendered rows, got %d:\n%s", len(rows), got)
	}
	width := utf8.RuneCountInString(rows[0])
	for _, r := range rows {
		if w := utf8.RuneCountInString(r); w != width {
			t.Fatalf("row not aligned (want width %d, got %d):\n%s", width, w, got)
		}
	}
	// The header and separator rows never contain literal pipes, so
	// their pipe count must equal columns+1. (Data rows can carry a
	// literal pipe inside a backtick cell, so we don't count those.)
	for _, r := range rows[:2] {
		if strings.Count(r, "|") != 4 { // 3 columns -> 4 separators
			t.Fatalf("column count drifted in row %q", r)
		}
	}
	if !strings.Contains(got, "`int | float`") {
		t.Fatalf("pipe-in-backtick cell mangled:\n%s", got)
	}
}

func TestEmit_NoTrailingMkdocsNoise(t *testing.T) {
	// A compact stand-in for a real guide page: comment, heading,
	// prose with a relative link, admonition, result div, table.
	src := lines(
		"[//]: # (internal note)",
		"## Strings",
		"",
		"See [shell commands](../guide/shell-commands.md).",
		"",
		`!!! tip "Heads up"`,
		"",
		"    Be careful.",
		"",
		`<div class="result">`,
		"```",
		"output",
		"```",
		"</div>",
		"",
		"| A | B |",
		"|---|---|",
		"| 1 | 2 |",
	)
	got := EmitTerminal(Parse(src))
	for _, bad := range []string{"[//]:", "!!!", `=== "`, "<div", "</div>"} {
		if strings.Contains(got, bad) {
			t.Fatalf("mkdocs noise %q leaked:\n%s", bad, got)
		}
	}
	if !strings.Contains(got, "**Tip: Heads up**") {
		t.Fatalf("admonition not converted:\n%s", got)
	}
}

package docir

import (
	"strings"
	"testing"
)

func lines(ls ...string) string { return strings.Join(ls, "\n") }

func TestParse_DropsAuthoringComments(t *testing.T) {
	src := lines(
		"## Data Types",
		"",
		"[//]: # (todo for number types)",
		"",
		"Rad has 6 basic types.",
		"",
		"[//]: # (TODO what about function types?!)",
		"",
		"### str",
	)
	blocks := Parse(src)
	if len(blocks) != 1 {
		t.Fatalf("want 1 Text block, got %d: %#v", len(blocks), blocks)
	}
	text, ok := blocks[0].(Text)
	if !ok {
		t.Fatalf("want Text, got %T", blocks[0])
	}
	joined := strings.Join(text.Lines, "\n")
	if strings.Contains(joined, "[//]:") {
		t.Fatalf("comment leaked into output:\n%s", joined)
	}
	// Adjacent comment-induced blank runs collapse to a single blank.
	if strings.Contains(joined, "\n\n\n") {
		t.Fatalf("blank run not collapsed:\n%q", joined)
	}
}

func TestParse_AdmonitionNestsCodeBlock(t *testing.T) {
	src := lines(
		`!!! info "Why"`,
		"",
		"    Some prose.",
		"",
		"    ```rad",
		`    x = 1`,
		"    ```",
	)
	blocks := Parse(src)
	if len(blocks) != 1 {
		t.Fatalf("want 1 Callout, got %d: %#v", len(blocks), blocks)
	}
	c, ok := blocks[0].(Callout)
	if !ok {
		t.Fatalf("want Callout, got %T", blocks[0])
	}
	if c.Kind != "info" || c.Title != "Why" {
		t.Fatalf("unexpected callout head: kind=%q title=%q", c.Kind, c.Title)
	}
	if len(c.Body) != 2 {
		t.Fatalf("want body [Text, Code], got %#v", c.Body)
	}
	if _, ok := c.Body[0].(Text); !ok {
		t.Fatalf("body[0] want Text, got %T", c.Body[0])
	}
	code, ok := c.Body[1].(Code)
	if !ok {
		t.Fatalf("body[1] want Code, got %T", c.Body[1])
	}
	if code.Lang != "rad" || code.Body != "x = 1" {
		t.Fatalf("nested code wrong: lang=%q body=%q", code.Lang, code.Body)
	}
}

func TestParse_ResultDivBecomesResultCode(t *testing.T) {
	src := lines(
		`<div class="result">`,
		"```",
		"Hello!",
		"```",
		"</div>",
	)
	blocks := Parse(src)
	if len(blocks) != 1 {
		t.Fatalf("want 1 block, got %d: %#v", len(blocks), blocks)
	}
	code, ok := blocks[0].(Code)
	if !ok {
		t.Fatalf("want Code, got %T", blocks[0])
	}
	if !code.IsResult {
		t.Fatalf("result div code should be marked IsResult")
	}
	if code.Body != "Hello!" {
		t.Fatalf("result body wrong: %q", code.Body)
	}
}

func TestParse_FenceDropsMkdocsAttrs(t *testing.T) {
	src := lines(
		`!`, // placeholder so the fence isn't the first line edge case
		"",
		`xyz`,
		"",
		"```rad linenums=\"1\" hl_lines=\"2\"",
		"name = 1",
		"```",
	)
	blocks := Parse(src)
	var code *Code
	for i := range blocks {
		if c, ok := blocks[i].(Code); ok {
			code = &c
		}
	}
	if code == nil {
		t.Fatalf("no code block parsed: %#v", blocks)
	}
	if code.Lang != "rad" {
		t.Fatalf("want lang rad (attrs dropped), got %q", code.Lang)
	}
}

func TestParse_TableCellWithPipeInBackticks(t *testing.T) {
	src := lines(
		"| Parameter | Type | Description |",
		"|-----------|------|-------------|",
		"| `val`     | `int | float`      | Value to constrain |",
	)
	blocks := Parse(src)
	if len(blocks) != 1 {
		t.Fatalf("want 1 Table, got %d: %#v", len(blocks), blocks)
	}
	tbl, ok := blocks[0].(Table)
	if !ok {
		t.Fatalf("want Table, got %T", blocks[0])
	}
	if len(tbl.Header) != 3 {
		t.Fatalf("want 3 header cells, got %d: %#v", len(tbl.Header), tbl.Header)
	}
	if len(tbl.Rows) != 1 || len(tbl.Rows[0]) != 3 {
		t.Fatalf("want one 3-cell row, got %#v", tbl.Rows)
	}
	if tbl.Rows[0][1] != "`int | float`" {
		t.Fatalf("pipe-in-backtick cell mis-split: %q", tbl.Rows[0][1])
	}
}

func TestParse_PlainProseAndHeadingsPassThrough(t *testing.T) {
	src := lines(
		"# Title",
		"",
		"A paragraph with a [link](../guide/x.md) and `code`.",
		"",
		"- item one",
		"- item two",
	)
	got := EmitTerminal(Parse(src))
	for _, want := range []string{"# Title", "[link](../guide/x.md)", "- item one"} {
		if !strings.Contains(got, want) {
			t.Fatalf("passthrough lost %q:\n%s", want, got)
		}
	}
}

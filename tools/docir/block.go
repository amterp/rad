// Package docir is a small intermediate representation for Rad's
// documentation. The source docs are authored in Material-for-MkDocs
// flavored markdown (admonitions, content tabs, result <div>s, GFM
// tables, HTML-comment authoring notes) because the website renders
// that dialect natively. None of those constructs survive a plain
// terminal renderer, so `rad docs` used to print them as literal
// noise.
//
// docir parses a doc into a flat list of typed Blocks - dropping
// authoring comments along the way - and then emits a target-specific
// rendering. Today the only emitter is EmitTerminal, which produces
// clean markdown for the embedded `rad docs` corpus (consumed both
// rendered in a TTY and piped raw into LLM context). The block model
// is the seam: a future web or LLM-specific emitter is a new function
// over the same []Block, not a new parser.
//
// The parser is deliberately scoped to the handful of constructs that
// actually appear in our corpus rather than full CommonMark - anything
// it doesn't model passes through verbatim as Text.
package docir

// Block is one structural element of a parsed doc.
type Block interface{ isBlock() }

// Text is the passthrough block: prose, headings, lists, thematic
// breaks, blank lines, and any markup docir doesn't specially model.
// Inline markdown inside the lines is left untouched - styling it is
// the terminal renderer's job at runtime. Lines have leading/trailing
// blank lines trimmed and internal blank runs collapsed to one.
type Text struct {
	Lines []string
}

// Code is a fenced code block. Lang is the first token of the fence
// info string (mkdocs extras like `linenums="1"` are dropped).
// IsResult marks a block that was wrapped in <div class="result"> -
// i.e. example output rather than input.
type Code struct {
	Lang     string
	Body     string
	IsResult bool
}

// Callout is an admonition (`!!! info "Title"`). Body is parsed
// recursively, so a code block inside an admonition is a real Code
// block, not escaped text.
type Callout struct {
	Kind  string
	Title string
	Body  []Block
}

// Tab is one labeled pane of a Tabs group.
type Tab struct {
	Label string
	Body  []Block
}

// Tabs is a content-tab group (`=== "Bash"` ...). On the web these
// render as interactive tabs; in a terminal there's no tabbing, so we
// flatten them to labeled sections.
type Tabs struct {
	Tabs []Tab
}

// Align is a table column alignment, parsed from the separator row's
// colons.
type Align int

const (
	AlignNone Align = iota
	AlignLeft
	AlignCenter
	AlignRight
)

// Table is a GFM table. Cells are stored verbatim (backticks and all);
// the terminal emitter pads them into aligned columns so they read in
// a terminal instead of as a ragged pipe soup.
type Table struct {
	Header []string
	Align  []Align
	Rows   [][]string
}

func (Text) isBlock()    {}
func (Code) isBlock()    {}
func (Callout) isBlock() {}
func (Tabs) isBlock()    {}
func (Table) isBlock()   {}

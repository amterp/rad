package lstesting

import (
	"fmt"
	"strconv"
	"strings"

	snap "github.com/amterp/go-snap"

	"github.com/amterp/rad/radls/lsp"
)

// ActionType identifies which LSP request a snapshot step makes.
type ActionType int

const (
	ActionChange ActionType = iota
	ActionCompletion
	ActionCodeAction
	ActionHover
	ActionDefinition
	ActionDocumentSymbol
	ActionReferences
	ActionRename
	ActionSemanticTokens
)

// Action carries one snapshot action's parameters. Different ActionTypes use
// different sub-fields; the omitted fields are just left zero. We don't bother
// with a discriminated union because each Action is short-lived and the fields
// are small.
type Action struct {
	Type               ActionType
	Content            string     // For CHANGE: new document text
	Position           *lsp.Pos   // For COMPLETION / HOVER / DEFINITION / REFERENCES / RENAME
	Range              *lsp.Range // For CODE_ACTION: selected range
	IncludeDeclaration bool       // For REFERENCES: matches LSP context flag
	NewName            string     // For RENAME: the rename target name
}

// SnapshotCase is what the harness runs: a document plus the sequence of
// requests to make against it.
type SnapshotCase struct {
	Title    string
	Document string
	Actions  []Action
}

// Suite declares the LSP snapshot format.
//
// Each request shape gets its own action because the LSP methods take
// different parameters; overloading one delimiter would make the files harder
// to read than this table is. What a header's arguments mean stays here rather
// than in go-snap: the engine tokenizes and counts them, this file decides that
// "1:6" is a position.
var Suite = snap.Suite{
	Inputs: []snap.Input{{Name: "DOCUMENT"}},
	Actions: []snap.Action{
		{Name: "CHANGE", HasBody: true},
		{Name: "COMPLETION", Args: 1},
		{Name: "CODE_ACTION", Args: 2},
		{Name: "HOVER", Args: 1},
		{Name: "DEFINITION", Args: 1},
		{Name: "DOCUMENT_SYMBOL"},
		// A position plus an optional "decl" token.
		{Name: "REFERENCES", Args: -1},
		{Name: "RENAME", Args: 2},
		{Name: "SEMANTIC_TOKENS"},
	},
	Outputs:  []snap.Output{{Name: "STDOUT"}},
	Parallel: true,
}

// caseFrom converts a parsed snapshot case into the harness's input.
func caseFrom(c *snap.Case) (*SnapshotCase, error) {
	tc := &SnapshotCase{Title: c.Title(), Document: c.Text("DOCUMENT")}
	for _, step := range c.Steps() {
		action, err := actionFrom(step)
		if err != nil {
			return nil, err
		}
		tc.Actions = append(tc.Actions, action)
	}
	return tc, nil
}

func actionFrom(step *snap.Step) (Action, error) {
	pos := func(i int) (*lsp.Pos, error) {
		p, err := parsePosToken(step.Args[i])
		if err != nil {
			return nil, fmt.Errorf("%s: %w", step.Name, err)
		}
		return &p, nil
	}

	switch step.Name {
	case "CHANGE":
		return Action{Type: ActionChange, Content: step.Body}, nil
	case "DOCUMENT_SYMBOL":
		return Action{Type: ActionDocumentSymbol}, nil
	case "SEMANTIC_TOKENS":
		return Action{Type: ActionSemanticTokens}, nil
	case "COMPLETION", "HOVER", "DEFINITION":
		p, err := pos(0)
		if err != nil {
			return Action{}, err
		}
		return Action{Type: map[string]ActionType{
			"COMPLETION": ActionCompletion,
			"HOVER":      ActionHover,
			"DEFINITION": ActionDefinition,
		}[step.Name], Position: p}, nil
	case "CODE_ACTION":
		start, err := pos(0)
		if err != nil {
			return Action{}, err
		}
		end, err := pos(1)
		if err != nil {
			return Action{}, err
		}
		r := lsp.Range{Start: *start, End: *end}
		return Action{Type: ActionCodeAction, Range: &r}, nil
	case "REFERENCES":
		if len(step.Args) == 0 {
			return Action{}, fmt.Errorf("REFERENCES needs a position")
		}
		p, err := pos(0)
		if err != nil {
			return Action{}, err
		}
		includeDecl := false
		for _, arg := range step.Args[1:] {
			if arg == "decl" {
				includeDecl = true
			}
		}
		return Action{Type: ActionReferences, Position: p, IncludeDeclaration: includeDecl}, nil
	case "RENAME":
		p, err := pos(0)
		if err != nil {
			return Action{}, err
		}
		return Action{Type: ActionRename, Position: p, NewName: step.Args[1]}, nil
	}
	return Action{}, fmt.Errorf("unhandled action %s", step.Name)
}

// parsePosToken parses a "line:char" token like "0:5" into an lsp.Pos.
func parsePosToken(token string) (lsp.Pos, error) {
	parts := strings.SplitN(token, ":", 2)
	if len(parts) != 2 {
		return lsp.Pos{}, fmt.Errorf("expected position as line:char, got %q", token)
	}
	line, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return lsp.Pos{}, fmt.Errorf("invalid line number %q: %w", parts[0], err)
	}
	char, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return lsp.Pos{}, fmt.Errorf("invalid character number %q: %w", parts[1], err)
	}
	return lsp.NewPos(line, char), nil
}

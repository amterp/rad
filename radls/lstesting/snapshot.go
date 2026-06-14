package lstesting

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/amterp/rad/radls/lsp"
)

// Snapshot updating is opt-in and targeted by default so a stray `-update`
// can't silently rewrite every snapshot (masking unrelated regressions). Pass
// `-update=<substr>[,<substr>...]` to rewrite only matching files; mismatches
// elsewhere still fail. Use `-update-all` for the rare blanket rewrite. This
// mirrors core/testing's snapshot flags for a consistent contributor workflow.
var (
	updateTargets = flag.String("update", "",
		"comma-separated snapshot path substrings to rewrite; "+
			"non-matching mismatches still fail. Use -update-all to rewrite everything.")
	updateAll = flag.Bool("update-all", false,
		"rewrite ALL snapshot expected outputs - prefer -update=<substr> to scope and avoid masking regressions")
)

func updateTargetList() []string {
	if strings.TrimSpace(*updateTargets) == "" {
		return nil
	}
	var out []string
	for _, t := range strings.Split(*updateTargets, ",") {
		if t = strings.TrimSpace(t); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// shouldUpdateSnapshotFile reports whether the snapshot file at path should be
// rewritten this run: true under -update-all, or when the path contains any
// -update substring.
func shouldUpdateSnapshotFile(path string) bool {
	if *updateAll {
		return true
	}
	for _, target := range updateTargetList() {
		if strings.Contains(path, target) {
			return true
		}
	}
	return false
}

// checkSnapshotUpdateFlags fails fast on the `-update -run X` trap, where the
// flag parser swallows `-run` as -update's value. Call once per snapshot test.
func checkSnapshotUpdateFlags(t *testing.T) {
	t.Helper()
	for _, target := range updateTargetList() {
		if strings.HasPrefix(target, "-") {
			t.Fatalf("-update value %q looks like a flag - write `-update=<path-substr>`, "+
				"or -update-all to rewrite everything", target)
		}
	}
}

// Action delimiters. Each request-shaped action has its own header
// because the LSP wire methods take different parameter shapes -
// trying to overload one delimiter would make the snap files
// harder to read than the small constant table here. Headers that
// carry no position/range trail with just "###"; positioned ones
// embed "line:char [line:char]" before the closing "###".
const (
	TitleDelimiter          = "### TITLE ###"
	DocumentDelimiter       = "### DOCUMENT ###"
	ChangeDelimiter         = "### CHANGE ###"
	CompletionDelimiter     = "### COMPLETION"      // "### COMPLETION 0:0 ###"
	CodeActionDelimiter     = "### CODE_ACTION"     // "### CODE_ACTION 0:0 0:0 ###"
	HoverDelimiter          = "### HOVER"           // "### HOVER 0:0 ###"
	DefinitionDelimiter     = "### DEFINITION"      // "### DEFINITION 0:0 ###"
	DocumentSymbolDelimiter = "### DOCUMENT_SYMBOL" // "### DOCUMENT_SYMBOL ###"
	ReferencesDelimiter     = "### REFERENCES"      // "### REFERENCES 0:0 [decl] ###"
	RenameDelimiter         = "### RENAME"          // "### RENAME 0:0 newName ###"
	SemanticTokensDelimiter = "### SEMANTIC_TOKENS" // "### SEMANTIC_TOKENS ###"
	StdoutDelimiter         = "### STDOUT ###"
)

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

// Action carries one snapshot action's parameters. Different
// ActionTypes use different sub-fields; the omitted fields are
// just left zero. We don't bother with a discriminated-union
// because each Action is short-lived and the fields are small.
type Action struct {
	Type               ActionType
	Content            string     // For CHANGE: new document text
	Position           *lsp.Pos   // For COMPLETION / HOVER / DEFINITION / REFERENCES / RENAME
	Range              *lsp.Range // For CODE_ACTION: selected range
	IncludeDeclaration bool       // For REFERENCES: matches LSP context flag
	NewName            string     // For RENAME: the rename target name
}

type SnapshotCase struct {
	Title    string
	Document string
	Actions  []Action
	Stdout   string
}

// ParseSnapshotFile reads a .snap file and extracts LSP test cases.
//
// Format:
//
//	### TITLE ###
//	<test name>
//	### DOCUMENT ###
//	<rad source code>
//	### CHANGE ###              (optional, repeatable)
//	<new document text>
//	### COMPLETION 0:0 ###      (optional, repeatable, header-only)
//	### CODE_ACTION 0:0 0:0 ### (optional, repeatable, header-only)
//	### STDOUT ###
//	<expected server output as normalized JSON>
func ParseSnapshotFile(path string) ([]SnapshotCase, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	const (
		stateInit = iota
		stateTitle
		stateDocument
		stateChangeBody
		stateStdout
	)

	var (
		cases         []SnapshotCase
		scanner       = bufio.NewScanner(file)
		state         = stateInit
		lineNum       = 0
		title         string
		docBuilder    strings.Builder
		changeBuilder strings.Builder
		stdoutBuilder strings.Builder
		actions       []Action
		inChange      bool // tracks whether we've entered a CHANGE section
	)

	finishCase := func() {
		if title != "" || docBuilder.Len() > 0 {
			tc := SnapshotCase{
				Title:    title,
				Document: strings.TrimSuffix(docBuilder.String(), "\n"),
				Actions:  actions,
				Stdout:   strings.TrimSuffix(stdoutBuilder.String(), "\n"),
			}
			cases = append(cases, tc)
		}
		title = ""
		docBuilder.Reset()
		changeBuilder.Reset()
		stdoutBuilder.Reset()
		actions = nil
		inChange = false
	}

	finishChange := func() {
		if inChange {
			actions = append(actions, Action{
				Type:    ActionChange,
				Content: strings.TrimSuffix(changeBuilder.String(), "\n"),
			})
			changeBuilder.Reset()
			inChange = false
		}
	}

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		switch state {
		case stateInit:
			if trimmed == TitleDelimiter {
				state = stateTitle
			}

		case stateTitle:
			title = strings.TrimSpace(line)
			if scanner.Scan() {
				lineNum++
				nextLine := strings.TrimSpace(scanner.Text())
				if nextLine == DocumentDelimiter {
					state = stateDocument
				} else {
					return nil, fmt.Errorf("%s:%d: expected '%s' after title, got '%s'",
						path, lineNum, DocumentDelimiter, nextLine)
				}
			}

		case stateDocument:
			switch {
			case trimmed == StdoutDelimiter:
				state = stateStdout
			case trimmed == ChangeDelimiter:
				inChange = true
				state = stateChangeBody
			case trimmed == TitleDelimiter:
				finishCase()
				state = stateTitle
			case isActionHeader(trimmed):
				action, err := parseActionHeader(trimmed, path, lineNum)
				if err != nil {
					return nil, err
				}
				actions = append(actions, action)
				// Stay in stateDocument - header-only actions don't change state
			default:
				docBuilder.WriteString(line)
				docBuilder.WriteString("\n")
			}

		case stateChangeBody:
			switch {
			case trimmed == StdoutDelimiter:
				finishChange()
				state = stateStdout
			case trimmed == ChangeDelimiter:
				finishChange()
				inChange = true
				// Stay in stateChangeBody for the next change
			case trimmed == TitleDelimiter:
				finishChange()
				finishCase()
				state = stateTitle
			case isActionHeader(trimmed):
				finishChange()
				action, err := parseActionHeader(trimmed, path, lineNum)
				if err != nil {
					return nil, err
				}
				actions = append(actions, action)
				state = stateDocument
			default:
				changeBuilder.WriteString(line)
				changeBuilder.WriteString("\n")
			}

		case stateStdout:
			switch {
			case trimmed == TitleDelimiter:
				finishCase()
				state = stateTitle
			default:
				stdoutBuilder.WriteString(line)
				stdoutBuilder.WriteString("\n")
			}
		}
	}

	if state != stateInit && state != stateTitle {
		finishCase()
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return cases, nil
}

// isActionHeader returns true if the line is a header-only
// action delimiter (any of the request-shaped delimiters). CHANGE
// is excluded because it has body content and gets its own
// stateChangeBody handling.
//
// Order matters here: more-specific prefixes must be tested
// before less-specific ones. DOCUMENT_SYMBOL contains DOCUMENT
// as a substring, so a naive prefix check would route the former
// to the latter. We avoid that by testing the longer name first.
func isActionHeader(line string) bool {
	return strings.HasPrefix(line, DocumentSymbolDelimiter) ||
		strings.HasPrefix(line, SemanticTokensDelimiter) ||
		strings.HasPrefix(line, CompletionDelimiter) ||
		strings.HasPrefix(line, CodeActionDelimiter) ||
		strings.HasPrefix(line, DefinitionDelimiter) ||
		strings.HasPrefix(line, ReferencesDelimiter) ||
		strings.HasPrefix(line, RenameDelimiter) ||
		strings.HasPrefix(line, HoverDelimiter)
}

// parseActionHeader parses an action header line into an Action.
// Same order-of-tests caveat as isActionHeader: longer prefixes
// first.
func parseActionHeader(line, path string, lineNum int) (Action, error) {
	switch {
	case strings.HasPrefix(line, DocumentSymbolDelimiter):
		return Action{Type: ActionDocumentSymbol}, nil
	case strings.HasPrefix(line, SemanticTokensDelimiter):
		return Action{Type: ActionSemanticTokens}, nil
	case strings.HasPrefix(line, CompletionDelimiter):
		pos, err := parsePosFromHeader(line, CompletionDelimiter, path, lineNum)
		if err != nil {
			return Action{}, err
		}
		return Action{Type: ActionCompletion, Position: &pos}, nil
	case strings.HasPrefix(line, CodeActionDelimiter):
		r, err := parseRangeFromHeader(line, CodeActionDelimiter, path, lineNum)
		if err != nil {
			return Action{}, err
		}
		return Action{Type: ActionCodeAction, Range: &r}, nil
	case strings.HasPrefix(line, DefinitionDelimiter):
		pos, err := parsePosFromHeader(line, DefinitionDelimiter, path, lineNum)
		if err != nil {
			return Action{}, err
		}
		return Action{Type: ActionDefinition, Position: &pos}, nil
	case strings.HasPrefix(line, ReferencesDelimiter):
		// REFERENCES headers carry a position and an optional
		// "decl" suffix that maps to includeDeclaration=true.
		// "### REFERENCES 1:6 ###"      -> uses-only
		// "### REFERENCES 1:6 decl ###" -> include declaration
		inner := strings.TrimPrefix(line, ReferencesDelimiter)
		inner = strings.TrimSuffix(inner, "###")
		inner = strings.TrimSpace(inner)
		parts := strings.Fields(inner)
		if len(parts) == 0 {
			return Action{}, fmt.Errorf("%s:%d: REFERENCES header missing position", path, lineNum)
		}
		pos, err := parsePosToken(parts[0], path, lineNum)
		if err != nil {
			return Action{}, err
		}
		includeDecl := false
		for _, p := range parts[1:] {
			if p == "decl" {
				includeDecl = true
			}
		}
		return Action{Type: ActionReferences, Position: &pos, IncludeDeclaration: includeDecl}, nil
	case strings.HasPrefix(line, RenameDelimiter):
		// RENAME headers carry a position and a new name token.
		// "### RENAME 1:0 newName ###"
		inner := strings.TrimPrefix(line, RenameDelimiter)
		inner = strings.TrimSuffix(inner, "###")
		inner = strings.TrimSpace(inner)
		parts := strings.Fields(inner)
		if len(parts) < 2 {
			return Action{}, fmt.Errorf("%s:%d: RENAME header needs position and new name", path, lineNum)
		}
		pos, err := parsePosToken(parts[0], path, lineNum)
		if err != nil {
			return Action{}, err
		}
		return Action{Type: ActionRename, Position: &pos, NewName: parts[1]}, nil
	case strings.HasPrefix(line, HoverDelimiter):
		pos, err := parsePosFromHeader(line, HoverDelimiter, path, lineNum)
		if err != nil {
			return Action{}, err
		}
		return Action{Type: ActionHover, Position: &pos}, nil
	}
	return Action{}, fmt.Errorf("%s:%d: unrecognized action header: %s", path, lineNum, line)
}

// parsePosToken parses a "line:char" token like "0:5" into an
// lsp.Pos. Pulled out so REFERENCES (which has additional
// suffix tokens) can share the parser without duplicating it.
func parsePosToken(token, path string, lineNum int) (lsp.Pos, error) {
	parts := strings.SplitN(token, ":", 2)
	if len(parts) != 2 {
		return lsp.Pos{}, fmt.Errorf("%s:%d: expected position as line:char, got '%s'", path, lineNum, token)
	}
	line, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return lsp.Pos{}, fmt.Errorf("%s:%d: invalid line number '%s': %w", path, lineNum, parts[0], err)
	}
	char, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return lsp.Pos{}, fmt.Errorf("%s:%d: invalid character number '%s': %w", path, lineNum, parts[1], err)
	}
	return lsp.NewPos(line, char), nil
}

// parsePosFromHeader extracts a position (line:char) from a header
// like "### COMPLETION 0:0 ###" by stripping the envelope and
// delegating to parsePosToken. Keeps the parsing logic in one
// place so a future shape change (e.g. supporting byte offsets)
// only needs one update.
func parsePosFromHeader(header, prefix, path string, lineNum int) (lsp.Pos, error) {
	inner := strings.TrimPrefix(header, prefix)
	inner = strings.TrimSuffix(inner, "###")
	inner = strings.TrimSpace(inner)
	return parsePosToken(inner, path, lineNum)
}

// parseRangeFromHeader extracts a range (startLine:startChar endLine:endChar) from a header
// like "### CODE_ACTION 0:0 0:0 ###".
func parseRangeFromHeader(header, prefix, path string, lineNum int) (lsp.Range, error) {
	inner := strings.TrimPrefix(header, prefix)
	inner = strings.TrimSuffix(inner, "###")
	inner = strings.TrimSpace(inner)

	positions := strings.Fields(inner)
	if len(positions) != 2 {
		return lsp.Range{}, fmt.Errorf("%s:%d: expected two positions (start end), got '%s'", path, lineNum, inner)
	}

	startParts := strings.SplitN(positions[0], ":", 2)
	endParts := strings.SplitN(positions[1], ":", 2)

	if len(startParts) != 2 || len(endParts) != 2 {
		return lsp.Range{}, fmt.Errorf("%s:%d: expected positions as line:char, got '%s'", path, lineNum, inner)
	}

	startLine, err := strconv.Atoi(startParts[0])
	if err != nil {
		return lsp.Range{}, fmt.Errorf("%s:%d: invalid start line '%s': %w", path, lineNum, startParts[0], err)
	}
	startChar, err := strconv.Atoi(startParts[1])
	if err != nil {
		return lsp.Range{}, fmt.Errorf("%s:%d: invalid start char '%s': %w", path, lineNum, startParts[1], err)
	}
	endLine, err := strconv.Atoi(endParts[0])
	if err != nil {
		return lsp.Range{}, fmt.Errorf("%s:%d: invalid end line '%s': %w", path, lineNum, endParts[0], err)
	}
	endChar, err := strconv.Atoi(endParts[1])
	if err != nil {
		return lsp.Range{}, fmt.Errorf("%s:%d: invalid end char '%s': %w", path, lineNum, endParts[1], err)
	}

	return lsp.NewRange(startLine, startChar, endLine, endChar), nil
}

// WriteSnapshotFile writes test cases back to a .snap file.
func WriteSnapshotFile(path string, cases []SnapshotCase) error {
	var builder strings.Builder

	for i, tc := range cases {
		if i > 0 {
			str := builder.String()
			if len(str) > 0 && str[len(str)-1] != '\n' {
				builder.WriteString("\n")
			}
		}

		builder.WriteString(TitleDelimiter)
		builder.WriteString("\n")
		builder.WriteString(tc.Title)
		builder.WriteString("\n")

		builder.WriteString(DocumentDelimiter)
		builder.WriteString("\n")
		if tc.Document != "" {
			builder.WriteString(tc.Document)
			builder.WriteString("\n")
		}

		for _, action := range tc.Actions {
			switch action.Type {
			case ActionChange:
				builder.WriteString(ChangeDelimiter)
				builder.WriteString("\n")
				if action.Content != "" {
					builder.WriteString(action.Content)
					builder.WriteString("\n")
				}
			case ActionCompletion:
				fmt.Fprintf(&builder, "%s %d:%d ###\n", CompletionDelimiter, action.Position.Line, action.Position.Character)
			case ActionCodeAction:
				fmt.Fprintf(&builder, "%s %d:%d %d:%d ###\n", CodeActionDelimiter,
					action.Range.Start.Line, action.Range.Start.Character,
					action.Range.End.Line, action.Range.End.Character)
			case ActionHover:
				fmt.Fprintf(&builder, "%s %d:%d ###\n", HoverDelimiter, action.Position.Line, action.Position.Character)
			case ActionDefinition:
				fmt.Fprintf(&builder, "%s %d:%d ###\n", DefinitionDelimiter, action.Position.Line, action.Position.Character)
			case ActionDocumentSymbol:
				fmt.Fprintf(&builder, "%s ###\n", DocumentSymbolDelimiter)
			case ActionReferences:
				if action.IncludeDeclaration {
					fmt.Fprintf(&builder, "%s %d:%d decl ###\n", ReferencesDelimiter, action.Position.Line, action.Position.Character)
				} else {
					fmt.Fprintf(&builder, "%s %d:%d ###\n", ReferencesDelimiter, action.Position.Line, action.Position.Character)
				}
			case ActionRename:
				fmt.Fprintf(&builder, "%s %d:%d %s ###\n", RenameDelimiter,
					action.Position.Line, action.Position.Character, action.NewName)
			case ActionSemanticTokens:
				fmt.Fprintf(&builder, "%s ###\n", SemanticTokensDelimiter)
			}
		}

		builder.WriteString(StdoutDelimiter)
		builder.WriteString("\n")
		if tc.Stdout != "" {
			builder.WriteString(tc.Stdout)
			builder.WriteString("\n")
		}
	}

	return os.WriteFile(path, []byte(builder.String()), 0644)
}

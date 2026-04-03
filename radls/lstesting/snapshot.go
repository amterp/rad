package lstesting

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/amterp/rad/radls/lsp"
)

var UpdateSnapshots = flag.Bool("update", false, "update snapshot expected outputs")

const (
	TitleDelimiter      = "### TITLE ###"
	DocumentDelimiter   = "### DOCUMENT ###"
	ChangeDelimiter     = "### CHANGE ###"
	CompletionDelimiter = "### COMPLETION" // e.g. "### COMPLETION 0:0 ###"
	CodeActionDelimiter = "### CODE_ACTION" // e.g. "### CODE_ACTION 0:0 0:0 ###"
	StdoutDelimiter     = "### STDOUT ###"
)

type ActionType int

const (
	ActionChange ActionType = iota
	ActionCompletion
	ActionCodeAction
)

type Action struct {
	Type     ActionType
	Content  string     // For CHANGE: new document text
	Position *lsp.Pos   // For COMPLETION: cursor position
	Range    *lsp.Range // For CODE_ACTION: selected range
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
		cases          []SnapshotCase
		scanner        = bufio.NewScanner(file)
		state          = stateInit
		lineNum        = 0
		title          string
		docBuilder     strings.Builder
		changeBuilder  strings.Builder
		stdoutBuilder  strings.Builder
		actions        []Action
		inChange       bool // tracks whether we've entered a CHANGE section
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

// isActionHeader returns true if the line is a header-only action delimiter
// (COMPLETION or CODE_ACTION). CHANGE is not included because it has body content.
func isActionHeader(line string) bool {
	return strings.HasPrefix(line, CompletionDelimiter) || strings.HasPrefix(line, CodeActionDelimiter)
}

// parseActionHeader parses an action header line into an Action.
func parseActionHeader(line, path string, lineNum int) (Action, error) {
	if strings.HasPrefix(line, CompletionDelimiter) {
		pos, err := parsePosFromHeader(line, CompletionDelimiter, path, lineNum)
		if err != nil {
			return Action{}, err
		}
		return Action{Type: ActionCompletion, Position: &pos}, nil
	}

	if strings.HasPrefix(line, CodeActionDelimiter) {
		r, err := parseRangeFromHeader(line, CodeActionDelimiter, path, lineNum)
		if err != nil {
			return Action{}, err
		}
		return Action{Type: ActionCodeAction, Range: &r}, nil
	}

	return Action{}, fmt.Errorf("%s:%d: unrecognized action header: %s", path, lineNum, line)
}

// parsePosFromHeader extracts a position (line:char) from a header like "### COMPLETION 0:0 ###".
func parsePosFromHeader(header, prefix, path string, lineNum int) (lsp.Pos, error) {
	// Strip prefix and trailing " ###"
	inner := strings.TrimPrefix(header, prefix)
	inner = strings.TrimSuffix(inner, "###")
	inner = strings.TrimSpace(inner)

	parts := strings.SplitN(inner, ":", 2)
	if len(parts) != 2 {
		return lsp.Pos{}, fmt.Errorf("%s:%d: expected position as line:char, got '%s'", path, lineNum, inner)
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

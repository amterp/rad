package core

import (
	"bufio"
	"errors"
	"io"
	"strings"

	"github.com/amterp/radish"
)

const (
	replPrompt     = "> "
	replContPrompt = ". "
)

// replRead says how a turn's input ended. It is not a bool because the three
// answers lead three different places: run it, throw it away, or stop.
type replRead int

const (
	replSubmitted replRead = iota
	replDiscarded          // Ctrl+C: abandon the buffer, prompt again
	replEOF                // Ctrl+D on an empty buffer, or the input ran out
)

// replReader gathers one turn's source. Two implementations share the
// continuation rule and nothing else: an editor when there is a terminal to
// draw on, and a plain line reader when there is not.
type replReader interface {
	Read(history []string) (string, replRead)
	Close() error
}

func newReplReader() replReader {
	if TerminalAvailable() {
		return &editorReader{}
	}
	return &lineReader{scanner: bufio.NewScanner(RIo.StdIn.Unwrap())}
}

// editorReader runs radish's multi-line editor for each turn.
//
// A fresh driver per turn is deliberate, not wasteful: radish restores the
// terminal when a prompt ends, so the terminal is in raw mode only while you
// are typing and back in cooked mode while your input runs. That is what lets
// Ctrl+C reach a running statement as a real signal, and what stops it being
// one while you are still editing.
type editorReader struct{}

func (r *editorReader) Read(history []string) (string, replRead) {
	m := radish.NewEditor().
		Prompt(replPrompt).
		ContPrompt(replContPrompt).
		Width(GetTermWidth()).
		MaxHeight(GetTermHeight()).
		History(history).
		IsComplete(func(buf string) bool { return !ReplNeedsMore(buf) }).
		Indent(ReplAutoIndent)

	_, _, err := RInteractive.Run(m)
	if err != nil {
		if !errors.Is(err, radish.ErrNotInteractive) {
			RP.RadDebugf("REPL editor error: %v", err)
		}
		return "", replEOF
	}

	switch m.Outcome() {
	case radish.EditorSubmitted:
		return m.Text(), replSubmitted
	case radish.EditorDiscarded:
		return "", replDiscarded
	default:
		return "", replEOF
	}
}

func (r *editorReader) Close() error { return nil }

// lineReader is the no-terminal path: a piped script, CI, an agent. It applies
// the same continuation rule as the editor, including the blank-line
// terminator, so a buffer that runs one way runs the other.
type lineReader struct {
	scanner *bufio.Scanner
}

func (r *lineReader) Read([]string) (string, replRead) {
	var buf []string
	for {
		if !r.scanner.Scan() {
			if err := r.scanner.Err(); err != nil && !errors.Is(err, io.EOF) {
				RP.RadDebugf("REPL read error: %v", err)
			}
			// Input ended mid-buffer: run what we have rather than discard it.
			if len(buf) > 0 {
				return strings.Join(buf, "\n"), replSubmitted
			}
			return "", replEOF
		}

		// No trimming: Rad's blocks are indentation-scoped, so leading
		// whitespace is syntax, not noise.
		line := r.scanner.Text()
		blank := strings.TrimSpace(line) == ""
		if blank && len(buf) == 0 {
			return "", replSubmitted // an empty turn; the loop skips it
		}
		buf = append(buf, line)

		joined := strings.Join(buf, "\n")
		if blank || !ReplNeedsMore(joined) {
			return joined, replSubmitted
		}
	}
}

func (r *lineReader) Close() error { return nil }

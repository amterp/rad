package rl

// Span represents a location range in source code.
// All positions are 0-indexed.
type Span struct {
	File      string
	StartByte int
	EndByte   int
	StartRow  int
	StartCol  int
	EndRow    int
	EndCol    int
}

// StartLine returns the 1-indexed start line number for display.
func (s Span) StartLine() int {
	return s.StartRow + 1
}

// StartColumn returns the 1-indexed start column number for display.
func (s Span) StartColumn() int {
	return s.StartCol + 1
}

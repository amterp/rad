package ra

import "os"

// ExitFunc is the interface for exiting the program
type ExitFunc func(int)

// StderrWriter is the interface for writing to stderr
type StderrWriter interface {
	Write([]byte) (int, error)
}

var osExit ExitFunc = os.Exit
var stderrWriter StderrWriter = os.Stderr

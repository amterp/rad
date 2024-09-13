package core

import (
	"io"
)

type RadIo struct {
	StdIn  io.Reader
	StdOut io.Writer
	StdErr io.Writer
}

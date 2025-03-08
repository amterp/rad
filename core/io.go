package core

import (
	"bytes"
	"io"
	"os"
)

type RadIo struct {
	StdIn  CheckableReader
	StdOut io.Writer
	StdErr io.Writer
}

type CheckableReader interface {
	Read(p []byte) (n int, err error)
	HasContent() bool
	Unwrap() io.Reader
}

type FileReader struct {
	file *os.File
}

func (fr *FileReader) Read(p []byte) (n int, err error) {
	return fr.file.Read(p)
}

func (fr *FileReader) HasContent() bool {
	stat, err := fr.file.Stat()
	if err != nil {
		return false
	}
	// Check if data is available (not a terminal)
	return stat.Size() > 0 || (stat.Mode()&os.ModeCharDevice) == 0
}

func (fr *FileReader) Unwrap() io.Reader {
	return fr.file
}

func NewFileReader(file *os.File) CheckableReader {
	return &FileReader{file: file}
}

type BufferReader struct {
	buffer *bytes.Buffer
}

func (br *BufferReader) Read(p []byte) (n int, err error) {
	return br.buffer.Read(p)
}

func (br *BufferReader) HasContent() bool {
	return br.buffer.Len() > 0
}

func (br *BufferReader) Unwrap() io.Reader {
	return br.buffer
}

func NewBufferReader(buffer *bytes.Buffer) CheckableReader {
	return &BufferReader{buffer: buffer}
}

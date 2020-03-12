package types

import (
	"os"
)

// PipePair is a pair of pipe
type PipePair struct {
	Reader *os.File
	Writer *os.File
}

// Read from pipe pair's reader
func (pp *PipePair) Read(p []byte) (int, error) {
	return pp.Reader.Read(p)
}

// Write to pipe pair's writer
func (pp *PipePair) Write(p []byte) (int, error) {
	return pp.Writer.Write(p)
}

// Close the pipe pair
func (pp *PipePair) Close() error {
	pp.Writer.Close()
	return pp.Reader.Close()
}

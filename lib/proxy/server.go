package proxy

import (
	"fmt"
	"net/rpc"
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

// Proxy struct of to rpc server
type Proxy struct {
}

// RegisterAndServe to register rpc and serve
func RegisterAndServe(pipes PipePair) {
	err := rpc.Register(new(Proxy))
	if err != nil {
		fmt.Println(err)
	}
	rpc.ServeConn(&pipes)
}

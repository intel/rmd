package proxyserver

import (
	"fmt"
	"net/rpc"

	"github.com/intel/rmd/internal/proxy/types"
)

// Proxy struct of to rpc server
type Proxy struct {
}

// RegisterAndServe to register rpc and serve
func RegisterAndServe(pipes types.PipePair) {
	err := rpc.Register(new(Proxy))
	if err != nil {
		fmt.Println(err)
	}
	rpc.ServeConn(&pipes)
}

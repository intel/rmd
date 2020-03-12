package proxyclient

import (
	"fmt"
	"net/rpc"

	"github.com/intel/rmd/internal/proxy/types"
)

// Client is the connection to rpc server
var Client *rpc.Client

// ConnectRPCServer by a pipe pair
// Be care about this method usage, it can only be called once while
// we start RMD API server, sync.once could be one choice, developer
// should control it.
func ConnectRPCServer(in types.PipePair) error {
	Client = rpc.NewClient(&in)
	if Client == nil {
		return fmt.Errorf("Failed to connect rpc server")
	}
	return nil
}

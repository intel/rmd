package proxyclient

import (
	"github.com/intel/rmd/internal/proxy/types"
)

// PAMAuthenticate leverage PAM to do authentication
func PAMAuthenticate(user string, pass string) error {

	req := types.PAMRequest{
		User: user,
		Pass: pass,
	}
	return Client.Call("Proxy.PAMAuthenticate", req, nil)
}

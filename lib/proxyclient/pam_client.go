package proxyclient

import (
	"github.com/intel/rmd/lib/proxy"
)

// PAMAuthenticate leverage PAM to do authentication
func PAMAuthenticate(user string, pass string) error {

	req := proxy.PAMRequest{
		User: user,
		Pass: pass,
	}
	return proxy.Client.Call("Proxy.PAMAuthenticate", req, nil)
}

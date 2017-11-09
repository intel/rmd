package proxy

import (
	"github.com/intel/rmd/lib/pam"
)

// PAMRequest is request from rpc client
type PAMRequest struct {
	User string
	Pass string
}

// PAMAuthenticate calls PAM authenticate
func (*Proxy) PAMAuthenticate(request PAMRequest, dummy *int) error {

	c := pam.Credential{
		Username: request.User,
		Password: request.Pass,
	}

	return c.PAMAuthenticate()

}

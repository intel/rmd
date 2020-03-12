package proxyserver

import (
	"github.com/intel/rmd/internal/pam"
	"github.com/intel/rmd/internal/proxy/types"
)

// PAMAuthenticate calls PAM authenticate
func (*Proxy) PAMAuthenticate(request types.PAMRequest, dummy *int) error {

	c := pam.Credential{
		Username: request.User,
		Password: request.Pass,
	}

	return c.PAMAuthenticate()

}

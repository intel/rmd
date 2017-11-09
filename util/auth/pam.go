package auth

import (
	"bytes"
	"github.com/emicklei/go-restful"
	"github.com/intel/rmd/lib/proxyclient"
	"net/http"
)

// PAMAuthenticate does PAM authenticate with PAM DB or PAM Unix
func PAMAuthenticate(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {

	// PAM enabled only for HTTPS requests
	if req.Request.TLS == nil {
		chain.ProcessFilter(req, resp)
		return
	}

	// Get user credentials
	u, p, ok := req.Request.BasicAuth()

	if !ok {
		resp.WriteErrorString(http.StatusBadRequest, "Malformed credentials\n")
		return
	}

	// PAM authenticate
	err := proxyclient.PAMAuthenticate(u, p)

	if err != nil {
		resp.AddHeader("WWW-Authenticate", "Basic realm=RMD")
		var buffer bytes.Buffer
		buffer.WriteString(err.Error())
		buffer.WriteString("\n")
		resp.WriteErrorString(http.StatusUnauthorized, buffer.String())
		return
	}

	chain.ProcessFilter(req, resp)
}

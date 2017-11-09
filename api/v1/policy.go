package v1

import (
	"net/http"

	"github.com/emicklei/go-restful"
	"github.com/intel/rmd/model/policy"
)

// PolicyResource represents policy API resource
type PolicyResource struct {
}

// Register handlers
func (c PolicyResource) Register(container *restful.Container) {
	ws := new(restful.WebService)
	ws.
		Path("/v1/policy").
		Doc("Show the policy defined on the host").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/").To(c.PolicyGet).
		Doc("Get the policy on the host").
		Operation("PolicyGet"))

	container.Add(ws)
}

// PolicyGet is handler to for GET
func (c PolicyResource) PolicyGet(request *restful.Request, response *restful.Response) {
	p, err := policy.GetDefaultPlatformPolicy()
	if err != nil {
		response.WriteError(http.StatusNotFound, err)
		return
	}
	response.WriteEntity(p)
}

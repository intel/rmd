package policy

import (
	"net/http"

	"github.com/emicklei/go-restful"
)

// Register handlers
func Register(prefix string, container *restful.Container) {
	ws := new(restful.WebService)
	ws.
		Path(prefix + "policy").
		Doc("Show the policy defined on the host").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/").To(Get).
		Doc("Get the policy on the host").
		Operation("PolicyGet"))

	container.Add(ws)
}

// Get is handler to for GET
func Get(request *restful.Request, response *restful.Response) {

	result, err := GetDefaultPlatformPolicy()
	if err != nil {
		response.WriteError(http.StatusNotFound, err)
		return
	}

	response.WriteEntity(result)
}

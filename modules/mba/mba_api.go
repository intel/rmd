package mba

import (
	"net/http"

	"github.com/emicklei/go-restful"
)

// Register add handlers for /v1/mba endpoint
func Register(prefix string, container *restful.Container) {
	ws := new(restful.WebService)
	ws.
		Path(prefix + "mba").
		Doc("Show the mba defined on the host").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/").To(Get).
		Doc("Get the mba info on the host").
		Operation("MbaGet"))

	container.Add(ws)
}

// Get is handler to for GET
func Get(request *restful.Request, response *restful.Response) {
	m := &Info{}
	err := m.Get()
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}
	response.WriteEntity(m)
}

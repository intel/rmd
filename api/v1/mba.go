package v1

import (
	"net/http"

	"github.com/emicklei/go-restful"
	m_mba "github.com/intel/rmd/model/mba"
)

// MbaResource represents Mba API
type MbaResource struct {
}

// Register handlers
func (c *MbaResource) Register(container *restful.Container) {
	ws := new(restful.WebService)
	ws.
		Path("/v1/mba").
		Doc("Show the mba defined on the host").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/").To(c.MbaGet).
		Doc("Get the mba info on the host").
		Operation("MbaGet"))

	container.Add(ws)
}

// MbaGet is handler to for GET
func (c *MbaResource) MbaGet(request *restful.Request, response *restful.Response) {
	m := &m_mba.MbaInfo{}
	err := m.Get()
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}
	response.WriteEntity(m)
}

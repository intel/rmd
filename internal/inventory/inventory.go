package inventory

import (
	"net/http"

	"github.com/emicklei/go-restful"
)

const endpointPath = "/v1/inventory"

// Inventory invenotry struct
type Inventory struct {
}

// Capability represents capability and its availability
type Capability struct {
	Name      string `json:"name"`
	Available bool   `json:"available"`
}

// Register registers REST endpoint
func (i *Inventory) Register(container *restful.Container) {
	ws := new(restful.WebService)
	ws.
		Path(endpointPath).
		Doc("Show the system hardware inventory and capabilities").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	// register handlers for supported GET actions
	ws.Route(ws.GET("/").To(i.GetCapabilities).
		Doc("Get the pstates for all CPU cores").
		Operation("PstateGetAll"))

	container.Add(ws)
}

// GetCapabilities return inventory capabilites
func (i *Inventory) GetCapabilities(request *restful.Request, response *restful.Response) {
	// response.WriteError(http.StatusNoContent, nil)
	result := []Capability{}

	// call each inventory checking function and add returned data to result
	result = append(result, CheckRDT(), CheckScaling())

	response.WriteHeaderAndEntity(http.StatusOK, result)
}

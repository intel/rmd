package v1

import (
	"net/http"

	"github.com/emicklei/go-restful"
	log "github.com/sirupsen/logrus"

	rmderror "github.com/intel/rmd/api/error"
	hospitality "github.com/intel/rmd/model/hospitality"
)

// HospitalityResource is the API resource
type HospitalityResource struct{}

// Register routers
func (h HospitalityResource) Register(container *restful.Container) {
	ws := new(restful.WebService)
	ws.
		Path("/v1/hospitality").
		Doc("Show the hospitality information of a host").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.POST("/").To(h.HospitalityGetByRequest).
		Doc("Get the hospitality information per request.").
		Operation("HospitalityGetByRequest").
		Writes(HospitalityResource{}))

	container.Add(ws)
}

// HospitalityGetByRequest returns hospitality score by request
func (h HospitalityResource) HospitalityGetByRequest(request *restful.Request, response *restful.Response) {
	hr := &hospitality.Request{}
	err := request.ReadEntity(&hr)

	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}

	log.Infof("Try to get hospitality score by %v", hr)
	score := &hospitality.Hospitality{}
	e := score.GetByRequest(hr)
	if e != nil {
		err := e.(*rmderror.AppError)
		response.WriteErrorString(err.Code, err.Error())
		return
	}
	response.WriteEntity(score)
}

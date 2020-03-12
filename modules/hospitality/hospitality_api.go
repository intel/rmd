package hospitality

import (
	"net/http"

	"github.com/emicklei/go-restful"
	log "github.com/sirupsen/logrus"

	rmderror "github.com/intel/rmd/internal/error"
)

// Register add handlers for /v1/hospitality endpoint
func Register(prefix string, container *restful.Container) {
	ws := new(restful.WebService)
	ws.
		Path(prefix + "hospitality").
		Doc("Show the hospitality information of a host").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.POST("/").To(GetByRequest).
		Doc("Get the hospitality information per request.").
		Operation("GetByRequest"))

	container.Add(ws)
}

// GetByRequest returns hospitality score by request
func GetByRequest(request *restful.Request, response *restful.Response) {
	hr := &Request{}
	err := request.ReadEntity(&hr)

	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}

	log.Infof("Try to get hospitality score by %v", hr)
	score := &Hospitality{}
	err = score.GetByRequest(hr)
	if err != nil {
		apperr, ok := err.(rmderror.AppError)
		if ok {
			response.WriteErrorString(apperr.Code, apperr.Error())
		} else {
			response.WriteErrorString(http.StatusInternalServerError, err.Error())
		}
		return
	}
	response.WriteEntity(score)
}

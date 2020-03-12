package workload

import (
	"net/http"

	"github.com/emicklei/go-restful"
	rmderror "github.com/intel/rmd/internal/error"
	wltypes "github.com/intel/rmd/modules/workload/types"
	log "github.com/sirupsen/logrus"
)

// Register add handlers for /v1/workloads endpoint
func Register(prefix string, container *restful.Container) {

	err := Init()
	if err != nil {
		// just inform user about DB creation failing
		log.Errorf("Failed to create data base. Reason: %s", err.Error())
		return
	}
	//no error case - perform registration
	ws := new(restful.WebService)
	ws.
		Path(prefix + "workloads").
		Doc("Show work loads").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/").To(Get).
		Doc("Get all work loads").
		Operation("WorkLoadGet"))

	ws.Route(ws.POST("/").To(NewWorkload).
		Doc("Create new work load").
		Operation("WorkLoadNew"))

	ws.Route(ws.GET("/{id:[0-9]*}").To(GetByID).
		Doc("Get workload by id").
		Param(ws.PathParameter("id", "id").DataType("string")).
		Operation("WorkLoadGetById"))

	ws.Route(ws.PATCH("/{id:[0-9]*}").To(Patch).
		Doc("Patch workload by id").
		Param(ws.PathParameter("id", "id").DataType("string")).
		Operation("WorkLoadPatch"))

	ws.Route(ws.DELETE("/{id:[0-9]*}").To(DeleteByID).
		Doc("Delete workload by id").
		Param(ws.PathParameter("id", "id").DataType("string")).
		Operation("WorkLoadDeleteByID"))

	container.Add(ws)
}

// Get handles GET /v1/workloads
func Get(request *restful.Request, response *restful.Response) {
	ws, err := GetAll()
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}
	response.WriteEntity(ws)
}

// GetByID handle GET /v1/workloads/{id}
func GetByID(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("id")
	log.Infof("Try to get workload by %s", id)
	wl, err := GetWorkloadByID(id)
	if len(wl.ID) == 0 {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusNotFound, "404: Could not found workload")
		return
	}
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}
	response.WriteEntity(wl)
}

// NewWorkload handle POST /v1/workloads
// sample POST request data
// body : '{ "core_ids" : ["1","2"], "policy": "gold" }'
// body : '{ "task_ids" : ["123"], "policy" : "silver" }'
// body : '{ "core_ids" : ["123"], "cache" : { "max" : 4, "min": 2 }, pstate : { "ratio": 3.0, "monitoring" : true } }
func NewWorkload(request *restful.Request, response *restful.Response) {
	wl := new(wltypes.RDTWorkLoad)
	err := request.ReadEntity(&wl)
	// set owner/origin for workload
	wl.Origin = "REST"

	log.Infof("Try to create workload %v", wl)
	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}

	if err := Validate(wl); err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusBadRequest,
			"Failed to validate workload. Reason: "+err.Error())
		return
	}

	e := Enforce(wl)
	if e != nil {
		response.AddHeader("Content-Type", "text/plain")
		httpStatus := http.StatusInternalServerError
		appErr, ok := e.(rmderror.AppError)
		// Some thing wrong in user's request parameters. Delete the DB.
		if ok && appErr.Code == http.StatusBadRequest {
			err = Delete(wl)
			if err != nil {
				// just log because we already are here in error case
				log.Errorf("Failed to delete workload from data base")
			}
			httpStatus = http.StatusBadRequest
		}
		response.WriteErrorString(httpStatus, e.Error())
		return
	}

	err = Create(wl)
	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}
	response.WriteHeaderAndEntity(http.StatusCreated, wl)
}

// Patch handles PATCH /v1/workloads/{id}
func Patch(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("id")
	wl, err := GetWorkloadByID(id)
	if len(wl.ID) == 0 || err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusNotFound, "404: Could not found workload")
		return
	}

	// workloads created by REST should be handled only by REST
	if wl.Origin == "REST" {
		log.Debug("Origin set as REST - Trying to modify workload...")
		newwl := new(wltypes.RDTWorkLoad)
		request.ReadEntity(&newwl)
		newwl.ID = id
		log.Infof("Try to patch a workload %v", newwl)

		if err = Update(&wl, newwl); err != nil {
			httpStatus := http.StatusInternalServerError
			apperr, ok := err.(rmderror.AppError)
			if ok && apperr.Code == rmderror.BadRequest {
				httpStatus = http.StatusBadRequest
			}
			response.WriteErrorString(httpStatus, err.Error())
			return
		}
		response.WriteEntity(wl)
	} else {
		log.Error("REST origin cannot modify non-REST workload")
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusOK, "200: You only have permission to modify REST origin workloads")
		return
	}
}

// DeleteByID handles DELETE /v1/workloads/{id}
func DeleteByID(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("id")
	log.Infof("Try to delete workload with id: %s", id)
	wl, err := GetWorkloadByID(id)

	if len(wl.ID) == 0 {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusNotFound, "404: Could not found workload")
		return
	}

	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}

	// workloads created by REST should be handled only by REST
	if wl.Origin == "REST" {
		log.Debug("Origin set as REST - Trying to delete workload...")

		if err = Release(&wl); err != nil {
			log.Error("Failed to release workload")
			response.WriteErrorString(http.StatusInternalServerError, err.Error())
			return
		}

		if err = Delete(&wl); err != nil {
			log.Error("Failed to delete workload")
			response.WriteErrorString(http.StatusInternalServerError, err.Error())
			return
		}

	} else {
		log.Error("REST origin cannot delete non-REST workload")
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusOK, "200: You only have permission to delete REST origin workloads")
		return
	}
}

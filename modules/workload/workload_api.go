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

	// create general table for workloads with no backend data for User
	userws := []wltypes.UserRDTWorkLoad{}

	for _, singleWorkload := range ws {
		wl := wltypes.UserRDTWorkLoad{}
		wl.ID = singleWorkload.ID
		wl.CoreIDs = singleWorkload.CoreIDs
		wl.TaskIDs = singleWorkload.TaskIDs
		wl.Policy = singleWorkload.Policy
		wl.Status = singleWorkload.Status
		wl.CosName = singleWorkload.CosName
		wl.Rdt = singleWorkload.Rdt
		wl.Plugins = singleWorkload.Plugins
		wl.UUID = singleWorkload.UUID
		wl.Origin = singleWorkload.Origin
		userws = append(userws, wl)
	}

	response.WriteEntity(userws)
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

	// create workload structure with no backend data for User
	userwl := wltypes.UserRDTWorkLoad{}
	userwl.ID = wl.ID
	userwl.CoreIDs = wl.CoreIDs
	userwl.TaskIDs = wl.TaskIDs
	userwl.Policy = wl.Policy
	userwl.Status = wl.Status
	userwl.CosName = wl.CosName
	userwl.Rdt = wl.Rdt
	userwl.Plugins = wl.Plugins
	userwl.UUID = wl.UUID
	userwl.Origin = wl.Origin

	response.WriteEntity(userwl)
}

// NewWorkload handle POST /v1/workloads
// sample POST request data
// body : '{ "core_ids" : ["1","2"], "policy": "gold" }'
// body : '{ "task_ids" : ["123"], "policy" : "silver" }'
// body : '{ "core_ids" : ["123"], "rdt" : { "cache" : { "max" : 4, "min": 2 } } }
func NewWorkload(request *restful.Request, response *restful.Response) {
	// workload only to return to the user with no backend params
	userWl := new(wltypes.UserRDTWorkLoad)
	err := request.ReadEntity(&userWl)
	// set owner/origin for workload
	userWl.Origin = "REST"

	log.Infof("Try to create workload %v", userWl)
	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusInternalServerError, "Failed to read request correctly. Please check request syntax and data")
		log.Errorf("Failed to read request due to: %v", err.Error())
		return
	}

	// create inner workload structure for all operations
	wl := new(wltypes.RDTWorkLoad)
	wl.ID = userWl.ID
	wl.CoreIDs = userWl.CoreIDs
	wl.TaskIDs = userWl.TaskIDs
	wl.Policy = userWl.Policy
	wl.Status = userWl.Status
	wl.CosName = userWl.CosName
	wl.Rdt = userWl.Rdt
	wl.Plugins = userWl.Plugins
	wl.UUID = userWl.UUID
	wl.Origin = userWl.Origin

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

	//Need to update data after all operations to display them for User
	userWl.ID = wl.ID
	userWl.Status = wl.Status
	userWl.CosName = wl.CosName
	userWl.UUID = wl.UUID
	// params below could change due to policy/manual params overwritting
	userWl.Policy = wl.Policy
	userWl.Rdt = wl.Rdt
	userWl.Plugins = wl.Plugins

	response.WriteHeaderAndEntity(http.StatusCreated, userWl)
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

		userWl := wltypes.UserRDTWorkLoad{}
		userWl.ID = wl.ID
		userWl.CoreIDs = wl.CoreIDs
		userWl.TaskIDs = wl.TaskIDs
		userWl.Policy = wl.Policy
		userWl.Status = wl.Status
		userWl.CosName = wl.CosName
		userWl.Rdt = wl.Rdt
		userWl.Plugins = wl.Plugins
		userWl.UUID = wl.UUID
		userWl.Origin = wl.Origin

		response.WriteEntity(userWl)
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

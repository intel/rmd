package v1

import (
	"net/http"

	"github.com/emicklei/go-restful"
	"github.com/intel/rmd/db"
	tw "github.com/intel/rmd/model/types/workload"
	"github.com/intel/rmd/model/workload"
	log "github.com/sirupsen/logrus"
)

// WorkLoadResource is workload api resource
type WorkLoadResource struct {
	Db db.DB
}

// Register handlers
func (w WorkLoadResource) Register(container *restful.Container) {
	ws := new(restful.WebService)
	ws.
		Path("/v1/workloads").
		Doc("Show work loads").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/").To(w.WorkLoadGet).
		Doc("Get all work loads").
		Operation("WorkLoadGet"))

	ws.Route(ws.POST("/").To(w.WorkLoadNew).
		Doc("Create new work load").
		Operation("WorkLoadNew"))

	ws.Route(ws.GET("/{id:[0-9]*}").To(w.WorkLoadGetByID).
		Doc("Get workload by id").
		Param(ws.PathParameter("id", "id").DataType("string")).
		Operation("WorkLoadGetById"))

	ws.Route(ws.PATCH("/{id:[0-9]*}").To(w.WorkLoadPatch).
		Doc("Patch workload by id").
		Param(ws.PathParameter("id", "id").DataType("string")).
		Operation("WorkLoadPatch"))

	ws.Route(ws.DELETE("/{id:[0-9]*}").To(w.WorkLoadDeleteByID).
		Doc("Delete workload by id").
		Param(ws.PathParameter("id", "id").DataType("string")).
		Operation("WorkLoadDeleteByID"))

	container.Add(ws)
}

// WorkLoadGet handles GET /v1/workloads
func (w WorkLoadResource) WorkLoadGet(request *restful.Request, response *restful.Response) {
	ws, err := w.Db.GetAllWorkload()
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}
	response.WriteEntity(ws)
}

// WorkLoadGetByID handle GET /v1/workloads/{id}
func (w WorkLoadResource) WorkLoadGetByID(request *restful.Request, response *restful.Response) {

	id := request.PathParameter("id")
	log.Infof("Try to get workload by %s", id)
	wl, err := w.Db.GetWorkloadByID(id)
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

// WorkLoadNew handle POST /v1/workloads
// body : '{"core_ids":["1","2"], "task_ids":["123","456"], "policys": ["foo"], "algorithms": ["bar"], "group": ["infra"]}'
func (w *WorkLoadResource) WorkLoadNew(request *restful.Request, response *restful.Response) {
	wl := new(tw.RDTWorkLoad)
	err := request.ReadEntity(&wl)
	log.Infof("Try to create workload %v", wl)
	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}

	if err := workload.Validate(wl); err != nil {
		response.WriteErrorString(http.StatusBadRequest,
			"Failed to validate workload. Reason: "+err.Error())
		return
	}

	err = w.Db.ValidateWorkload(wl)
	if err != nil {
		response.WriteErrorString(http.StatusBadRequest,
			"Failed to validate workload. Reason: "+err.Error())
		return
	}

	e := workload.Enforce(wl)
	if e != nil {
		response.WriteErrorString(e.Code, e.Error())
		// Some thing wrong in user's request parameters. Delete the DB.
		if e.Code == http.StatusBadRequest {
			err = w.Db.DeleteWorkload(wl)
		}
		return
	}

	err = w.Db.CreateWorkload(wl)
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}
	response.WriteHeaderAndEntity(http.StatusCreated, wl)
}

// WorkLoadPatch handles PATCH /v1/workloads/{id}
func (w WorkLoadResource) WorkLoadPatch(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("id")
	wl, err := w.Db.GetWorkloadByID(id)
	if len(wl.ID) == 0 || err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusNotFound, "404: Could not found workload")
		return
	}

	newwl := new(tw.RDTWorkLoad)
	request.ReadEntity(&newwl)
	newwl.ID = id

	log.Infof("Try to patch a workload %v", newwl)
	e := workload.Update(&wl, newwl)
	if e != nil {
		response.WriteErrorString(e.Code, e.Error())
		return
	}
	if err = w.Db.UpdateWorkload(&wl); err != nil {
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}

	response.WriteEntity(wl)
}

// WorkLoadDeleteByID handles DELETE /v1/workloads/{id}
func (w WorkLoadResource) WorkLoadDeleteByID(request *restful.Request, response *restful.Response) {

	id := request.PathParameter("id")
	log.Infof("Try to delete workload by %s", id)
	wl, err := w.Db.GetWorkloadByID(id)

	if len(wl.ID) == 0 {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusNotFound, "404: Could not found workload")
		return
	}

	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}

	if err = workload.Release(&wl); err != nil {
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}

	err = w.Db.DeleteWorkload(&wl)
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
	}
}

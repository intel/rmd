package cache

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/emicklei/go-restful"
	rmderror "github.com/intel/rmd/internal/error"
	log "github.com/sirupsen/logrus"
)

func getCacheLevelFromURL(request *restful.Request) uint32 {
	var ilev uint32
	slev := strings.TrimLeft(request.PathParameter("cache-level"), "l")
	if slev == "c" {
		ilev = GetLLC()
	} else {
		// go-resfull re is a gate for badrequest. so no err here.
		l, _ := strconv.Atoi(slev)
		ilev = uint32(l)
	}
	return ilev
}

// Register handlers
func Register(prefix string, container *restful.Container) {
	ws := new(restful.WebService)
	ws.
		Path(prefix + "cache").
		Doc("Show the cache information of a host").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/").To(Get).
		Doc("Get the cache information, summary.").
		Operation("CachesGet"))

	ws.Route(ws.GET("/{cache-level:^l([2-3]|lc)$}").To(LevelGet).
		Doc("Get the info of a specified level cache.").
		Param(ws.PathParameter("cache-level", "cache level").DataType("string")).
		Operation("CachesLevelGet"))

	// FIXME : should use pattern, \d\{1,3\}
	ws.Route(ws.GET("/{cache-level:^l([2-3]|lc)$}/{id:^[0-9]{1,9}$").To(GetSpecifiedCache).
		Doc("Get the info of a specified cache.").
		Param(ws.PathParameter("cache-level", "cache level").DataType("string")).
		Param(ws.PathParameter("id", "cache id").DataType("uint")).
		Operation("CacheGet"))
	// NOTE : seems DataType("uint") just for check?

	container.Add(ws)
}

// Get handles GET /v1/cache
func Get(request *restful.Request, response *restful.Response) {
	c := &CachesSummary{}
	err := c.Get()
	// FIXME : We should classify the error.
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}
	response.WriteEntity(c)
}

// LevelGet handles GET /v1/cache/l[2|3|lc]
func LevelGet(request *restful.Request, response *restful.Response) {
	c := &Infos{}

	ilev := getCacheLevelFromURL(request)
	log.Printf("Request Level %d", ilev)
	err := c.GetByLevel(ilev)
	if err != nil {
		appErr, ok := err.(rmderror.AppError)
		if ok {
			response.WriteErrorString(appErr.Code, appErr.Error())
		} else {
			response.WriteErrorString(http.StatusBadRequest, appErr.Error())
		}
		return
	}
	response.WriteEntity(c)
}

// GetSpecifiedCache handles GET /v1/cache/l[2, 3]/{id}
func GetSpecifiedCache(request *restful.Request, response *restful.Response) {
	c := &Infos{}

	ilev := getCacheLevelFromURL(request)
	// FIXME : should use pattern, \d\{1,3\}
	id, err := strconv.Atoi(request.PathParameter("id"))
	if err != nil {
		err := fmt.Errorf("Please input the correct id, it should be digital")
		response.WriteError(http.StatusBadRequest, err)
		return
	}
	log.Printf("Request Level%d, id: %d\n", ilev, id)

	e := c.GetByLevel(ilev)
	if e != nil {
		appErr, ok := e.(rmderror.AppError)
		// FIXME : We should classify the error.
		if ok {
			response.WriteErrorString(appErr.Code, appErr.Error())
		} else {
			response.WriteErrorString(http.StatusInternalServerError, appErr.Error())
		}
		return
	}
	ci, ok := c.Caches[uint32(id)]
	if !ok {
		err := fmt.Errorf("Cache id %d for level %d is not found", id, ilev)
		response.WriteError(http.StatusNotFound, err)
		return
	}
	response.WriteEntity(ci)
}

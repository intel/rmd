package v1

import (
	"github.com/emicklei/go-restful"
	"github.com/intel/rmd/util/rdtpool/config"
)

// CachePoolResource represents CachePool API resource
type CachePoolResource struct {
}

// Register handlers
func (c CachePoolResource) Register(container *restful.Container) {
	ws := new(restful.WebService)
	ws.
		Path("/v1/cachepool").
		Doc("Show the cachepool defined on the host").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/").To(c.CacheGet).
		Doc("Get the cachepool on the host").
		Operation("CacheGet"))

	container.Add(ws)
}

// CacheGet is handler to for GET
func (c CachePoolResource) CacheGet(request *restful.Request, response *restful.Response) {
	poolConf := config.NewCachePoolConfig()
	cpSettings := make(map[string]uint)
	cpSettings["guarantee"] = poolConf.Guarantee
	cpSettings["besteffort"] = poolConf.Besteffort
	cpSettings["shared"] = poolConf.Shared
	response.WriteEntity(cpSettings)
}

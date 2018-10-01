// Copyright 2018 QCT (Quanta Cloud Technology). All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package v1

import (
	"github.com/emicklei/go-restful"
	"github.com/intel/rmd/lib/mba"
	"github.com/intel/rmd/lib/proc"
)

// MbaInfo represents Mba API
type MbaInfo struct {
	Mba     bool `json:"mba"`
	MbaOn   bool `json:"mba_enable,omitempty"`
	MbaStep int  `json:"mba_step,omitempty"`
	MbaMin  int  `json:"mba_min,omitempty"`
}

// Register handlers
func (c *MbaInfo) Register(container *restful.Container) {
	flag, err := proc.IsMbaAvailiable()
	if err == nil {
		c.Mba = flag
		c.MbaOn = proc.IsEnableMba()
		if c.MbaOn {
			mbaStep, mbaMin, err := mba.GetMbaInfo()
			if err == nil {
				c.MbaStep = mbaStep
				c.MbaMin = mbaMin
			}
		}
	}

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
func (c *MbaInfo) MbaGet(request *restful.Request, response *restful.Response) {
	response.WriteEntity(c)
}

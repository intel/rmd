// This file contains implementation of REST handlers invoked by HandleRequest() method
// of ModuleInterface implemented in interface.go

package main

import (
	"github.com/emicklei/go-restful"
)

// ResultID is a struct used in REST handler for /sampleplugin
type ResultID struct {
	Name       string `json:"name"`
	Identifier int64  `json:"id"`
}

// ResultStatus is a struct used in REST handler for /sampleplugin/status
type ResultStatus struct {
	Status bool `json:"status"`
}

// Two functions below are not part of module interface. Also they don't have to be exported.
// They are added to "module" struct for easier access of module's internal values

func (m *module) handleMain(request *restful.Request, response *restful.Response) {
	stat := ResultID{Name: "sampleplugin", Identifier: m.identifier}
	response.WriteEntity(stat)
}

func (m *module) handleStatus(request *restful.Request, response *restful.Response) {
	response.WriteEntity(ResultStatus{Status: m.initialized})
}

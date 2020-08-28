// This file contains implementation of ModuleInterface and necessary Handle variable.
// It is needed for RMD to load plugin and access it's implementation.

package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/emicklei/go-restful"
)

type module struct {
	initialized bool
	counter     int64
	identifier  int64
	lastValue   float64
}

// Handle is an entry point for module
var Handle module

// Implementation of necessary ModuleInterface functions below

// Initialize is an initialization function
func (m *module) Initialize(config map[string]interface{}) error {
	fmt.Println("Initializing sampleplugin")

	// check if mandatory parameter exists
	modIDIface, ok := config["identifier"]
	if !ok {
		return errors.New("sampleplugin::Initialize() Missing 'identifier' initialization param")
	}
	// check type of mandatory param
	modIDValue, ok := modIDIface.(int64)
	if !ok {
		return errors.New("sampleplugin::Initialize() Invalid type of 'identifier' initialization param")
	}
	// check value of mandatory param
	if modIDValue <= 0 {
		return errors.New("sampleplugin::Initialize() Invalid value of 'identifier' initialization param")
	}

	m.identifier = modIDValue
	m.counter = 0
	m.initialized = true
	return nil
}

// Enforce sends params to be set
func (m *module) Enforce(params map[string]interface{}) (string, error) {
	if !m.initialized {
		return "", errors.New("sampleplugin::Enforce() Plugin not initialized")
	}
	data, err := convertParamsToData(params)
	if err != nil {
		// this case should never happen as parameters are validated on user-process
		// before sending through proxy to root process
		return "", errors.New("sampleplugin::Enforce() Failed to get params")
	}

	// just use value to show that it has been processed properly
	m.lastValue = data.value
	m.counter++
	return strconv.Itoa(int(m.counter)), nil
}

// Release removes setting (turns off monitoring) for given list of cpu cores
func (m *module) Release(params map[string]interface{}) error {
	if !m.initialized {
		return errors.New("sampleplugin::Release() Plugin not initialized")
	}

	// first check for ENFORCEID - param generated during Enforce() execution
	enforceIface, ok := params["ENFORCEID"]
	if !ok {
		return errors.New("sampleplugin::Release() No necessary input params")
	}
	// ENFORCEID value should be of type string
	enforceString, ok := enforceIface.(string)
	if !ok {
		return errors.New("sampleplugin::Release() Invalid type of input param")
	}
	// ENFORCEID cannot be smaller than 1
	enforceInt, err := strconv.Atoi(enforceString)
	if err != nil || enforceInt < 1 {
		return errors.New("sampleplugin::Release() Invalid value of input param")
	}

	return nil
}

// Validate validates provided params for future Enforce
func (m *module) Validate(params map[string]interface{}) error {
	if !m.initialized {
		return errors.New("sampleplugin::Validate() Plugin not initialized")
	}
	_, err := convertParamsToData(params)
	if err != nil {
		return fmt.Errorf("sampleplugin::Validate() %v", err.Error())
	}
	return nil
}

// GetEndpointPrefixes returns REST endpoints handled by module in go-restful compatible format
func (m *module) GetEndpointPrefixes() []string {
	if !m.initialized {
		return []string{}
	}
	return []string{"/sampleplugin", "/sampleplugin/status"}
}

// HandleRequest is called by HTTP request routing mechanism
func (m *module) HandleRequest(request *restful.Request, response *restful.Response) {
	if !m.initialized {
		response.WriteError(http.StatusInternalServerError, fmt.Errorf("Data requested from not initialized plugin"))
		return
	}

	if strings.HasSuffix(request.Request.RequestURI, "/sampleplugin") {
		// handle main REST path
		m.handleMain(request, response)
	} else if strings.HasSuffix(request.Request.RequestURI, "/sampleplugin/status") {
		// handle 'status' sub-path
		m.handleStatus(request, response)
	} else {
		// should never happen but better handle this
		response.WriteError(http.StatusInternalServerError, fmt.Errorf("Invalid request uri to /sampleplugin"))
	}
}

// GetCapabilities returns comma separated list of platform resources used by plugin
//
// Currently this function is not implemented as RMD is not supporting capabilities at the moment
func (m *module) GetCapabilities() string {
	return ""
}

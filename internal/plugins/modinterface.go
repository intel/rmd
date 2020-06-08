package plugins

import (
	"github.com/emicklei/go-restful"
)

// ModuleInterface defines interface of loadable RMD modules (dynamic loading with use of "plugin" package)
type ModuleInterface interface {

	// Initialize is a module initialization function
	// config param contains all information needed to initialize plugin
	// (ex. path to config file)
	Initialize(params map[string]interface{}) error

	// GetEndpointPrefixes returns declaration of REST endpoints handled by this module
	// If function's implementation for specific module returns:
	// { "/endpoint1" and "/endpoint2/" }
	// then RMD will expose and forward to this module requests for URI's:
	// - http://ip:port/v1/endpoint1
	// - http://ip:port/v1/endpoint2/
	// - all http://ip:port/v1/endpoint2/{something}
	GetEndpointPrefixes() []string

	// HandleRequest is called by HTTP request routing mechanism
	//
	// NOTE: currently "emicklei/go-restfull" package is used for HTTP requests routing
	// There's also a plan for standard HTTP package usage with following handling function
	// HandleRequest(wrt http.ResponseWriter, req *http.Request) error
	HandleRequest(request *restful.Request, response *restful.Response)

	// Validate allows workload module to check parameters before trying to enforce them
	Validate(params map[string]interface{}) error

	// Enforce allocates resources or set platform params according to data in 'params' map
	// Returned string should contain identifier for allocated resource.
	// If plugin does not need to store any identifier for future use in Release() then string should be empty
	Enforce(params map[string]interface{}) (string, error)

	// Release removes setting for given params
	// (in case of pstate it will be just disabling of monitoring for specified cores)
	Release(params map[string]interface{}) error

	// GetCapabilities returns comma separated list of platform resources used by plugin
	GetCapabilities() string
}

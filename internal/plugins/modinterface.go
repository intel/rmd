package plugins

import (
	"github.com/emicklei/go-restful"
)

// ModuleInterface defines interface of loadable RMD modules (dynamic loading with use of "plugin" package)
type ModuleInterface interface {
	// Initialize is a module initialization function
	// config param contains all information needed to initialize plugin (ex. path to config file)
	Initialize(config map[string]interface{}) error
	// Register is a REST endpoint registration function. If plugin is not registering any endpoint than function should be empty
	Register(prefix string, container *restful.Container) error

	// Enforce sends params to be set, as some plugins can need it then list of cpus and/or process ids has to be passed
	Enforce(cpus, processes []int, params map[string]interface{}) error
	// Release removes setting for given params (in case of pstate it will be just disabling of monitoring for specified cores)
	Release(cpus, processes []int, params map[string]interface{}) error

	// Get returns configuration/state for given cpus/processes if applicable.
	// Returns result as json string
	Get(cpus, processes []int) (string, error)
	// GetAll returns all configuration for this module.
	// Returns result as json string
	GetAll() (string, error)

	// ProxyCall is a generic function to be called by RMD proxy server implementation
	// Thanks to enclosing all internal functions in one generic it is possible to make
	// RMD proxy implementation unaware of module specific functions
	ProxyCall(function string, params map[string]interface{}) error
}

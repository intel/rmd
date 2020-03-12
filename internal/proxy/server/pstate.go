package proxyserver

import (
	"errors"

	"github.com/intel/rmd/modules/pstate"
)

// GenericCall is single generic method to be used over proxy
// In future proxy can be refactored to use generic approach in all modules
//
// WARNING: At the moment only following P-State related functions names are valid:
// - "DeleteConfig"
// - "PatchConfigs"
func (*Proxy) GenericCall(params map[string]interface{}, unused *int) error {
	if pstate.Instance == nil {
		return errors.New("Can't call P-State Instance - not loaded")
	}
	fnameRaw, ok := params["PROXYFUNCTION"]
	if !ok {
		return errors.New("Function name not defined in GenericCall")
	}
	fnameString, ok := fnameRaw.(string)
	if !ok {
		return errors.New("Function name type invalid")
	}
	mnameRaw, ok := params["PROXYMODULE"]
	if !ok {
		return errors.New("Module name not defined in GenericCall")
	}
	// currently module name not needed - prepared for future multi-plugin handling
	_, ok = mnameRaw.(string)
	if !ok {
		return errors.New("Module name type invalid")
	}

	// remove function and module names from map - not needed anymore
	delete(params, "PROXYFUNCTION")
	delete(params, "PROXYMODULE")

	// no error till now - can call generic method on attached module
	return pstate.Instance.ProxyCall(fnameString, params)
}

package proxyserver

import (
	"errors"

	"github.com/intel/rmd/internal/plugins"
)

// Enforce simplifies enforcing data in a more generic way
func (*Proxy) Enforce(params map[string]interface{}, id *string) error {

	rmdModuleAsInterface, ok := params["RMDMODULE"]
	if !ok {
		return errors.New("No RMD module specified")
	}

	rmdModuleAsString, ok := rmdModuleAsInterface.(string)
	if !ok {
		return errors.New("Failed to convert type for RMDMODULE")
	}

	// remove RMDMODULE element from map because
	// we don't want it to be passed to plugin
	delete(params, "RMDMODULE")

	var err error
	*id, err = plugins.Enforce(rmdModuleAsString, params)

	return err
}

// Release simplifies releasing data in a more generic way
func (*Proxy) Release(params map[string]interface{}, unused *int) error {

	rmdModuleAsInterface, ok := params["RMDMODULE"]
	if !ok {
		return errors.New("No RMD module specified")
	}

	rmdModuleAsString, ok := rmdModuleAsInterface.(string)
	if !ok {
		return errors.New("Failed to convert type for RMDMODULE")
	}

	// remove RMDMODULE element from map because
	// we don't want it to be passed to plugin
	delete(params, "RMDMODULE")

	return plugins.Release(rmdModuleAsString, params)
}

package plugins

import (
	"errors"

	logger "github.com/sirupsen/logrus"
)

// Interfaces stores module-name to ModuleInterface mapping for all loaded plugins
var Interfaces = make(map[string]ModuleInterface)

// Enforce simplifies enforcing data using specified plugin
//
// Function checks if plugin is loaded, verifies if stored interface is not null and then calls Enforce() method of stored interface.
// Returns error if any of steps above fails
func Enforce(moduleName string, params map[string]interface{}) (string, error) {
	logger.Debugf("Enforce() requested for %v plugin", moduleName)
	iface, ok := Interfaces[moduleName]
	if !ok {
		return "", errors.New("Selected plugin is not loaded")
	}
	if iface == nil {
		return "", errors.New("Internal error: nil pointer to plugin")
	}
	return iface.Enforce(params)
}

// Release simplifies releasing data using specified plugin
//
// Function checks if plugin is loaded, verifies if stored interface is not null and then calls Release() method of stored interface.
// Returns error if any of steps above fails
func Release(moduleName string, params map[string]interface{}) error {
	logger.Debugf("Release() requested for %v plugin", moduleName)
	iface, ok := Interfaces[moduleName]
	if !ok {
		return errors.New("Selected plugin is not loaded")
	}
	if iface == nil {
		return errors.New("Internal error: nil pointer to plugin")
	}
	return iface.Release(params)
}

// Validate simplifies data validation for selected plugin
//
// Function checks if plugin is loaded, verifies if stored interface is not null and then calls Validate() method of stored interface.
// Returns error if any of steps above fails
func Validate(moduleName string, params map[string]interface{}) error {
	logger.Debugf("Validate() requested for %v plugin", moduleName)
	iface, ok := Interfaces[moduleName]
	if !ok {
		return errors.New("Selected plugin is not loaded")
	}
	if iface == nil {
		return errors.New("Internal error: nil pointer to plugin")
	}
	return iface.Validate(params)
}

// Store stores provided module interface under provided name
//
// If module already exists or one of params is not correct (ex. nil interface)
// error is returned
func Store(name string, iface ModuleInterface) error {
	if name == "" || iface == nil {
		return errors.New("Invalid parameter")
	}

	_, ok := Interfaces[name]
	if ok {
		// module with this name already exists
		return errors.New("Module already registered under this name")
	}

	logger.Debugf("Plugin %v saved in loaded plugins", name)
	Interfaces[name] = iface
	return nil
}

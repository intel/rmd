package pstate

import (
	"github.com/intel/rmd/internal/plugins"
)

// Instance is a pointer to loaded implementation of ModuleInterface
var Instance plugins.ModuleInterface

// Load loads dynamic library (.so file) with P-State plugin
func Load(path string) error {
	result, err := plugins.Load(path)
	if err == nil {
		Instance = result
	}

	return err
}

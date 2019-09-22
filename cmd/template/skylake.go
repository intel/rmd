package template

// Skylake options
var Skylake = map[string]interface{}{
	"os_cacheways":               1,
	"infra_cacheways":            10,
	"max_shared":                 10,
	"guarantee":                  6,
	"besteffort":                 3,
	"shared":                     1,
	"mba_percentage_osgroup":     100,
	"mba_percentage_infragroup":  100,
	"mba_percentage_shared_pool": 100}

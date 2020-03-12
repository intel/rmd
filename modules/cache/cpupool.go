package cache

import (
	"sync"

	"github.com/intel/rmd/modules/cache/config"
	util "github.com/intel/rmd/utils/bitmap"
	"github.com/intel/rmd/utils/cpu"
)

// Workload can only get CPUs from this pool.
var cpuPoolPerCache = map[string]map[string]*util.Bitmap{
	"all":      map[string]*util.Bitmap{},
	"isolated": map[string]*util.Bitmap{}}
var cpuPoolOnce sync.Once

// GetCPUPools is helper function to get Reserved CPUs
func GetCPUPools() (map[string]map[string]*util.Bitmap, error) {
	var returnErr error

	cpuPoolOnce.Do(func() {

		var osCPUbm, infraCPUbm, isolatedCPUbm *util.Bitmap

		osconf := config.NewOSConfig()
		osCPUbm, err := BitmapsCPUWrapper([]string{osconf.CPUSet})
		if err != nil {
			returnErr = err
			return
		}
		infraconf := config.NewInfraConfig()
		if infraconf != nil && len(infraconf.CPUSet) > 0 {
			infraCPUbm, err = BitmapsCPUWrapper([]string{infraconf.CPUSet})
			if err != nil {
				returnErr = err
				return
			}
		} else {
			infraCPUbm, _ = BitmapsCPUWrapper("Ox0")
		}

		level := GetLLC()
		syscaches, err := GetSysCaches(int(level))
		if err != nil {
			returnErr = err
			return
		}

		isocpu := cpu.IsolatedCPUs()

		if isocpu != "" {
			isolatedCPUbm, _ = BitmapsCPUWrapper([]string{isocpu})
		} else {
			isolatedCPUbm, _ = BitmapsCPUWrapper("Ox0")
		}

		for _, sc := range syscaches {
			bm, _ := BitmapsCPUWrapper([]string{sc.SharedCPUList})
			cpuPoolPerCache["all"][sc.ID] = bm.Axor(osCPUbm).Axor(infraCPUbm)
			cpuPoolPerCache["isolated"][sc.ID] = bm.Axor(osCPUbm).Axor(infraCPUbm).And(isolatedCPUbm)
		}
	})
	return cpuPoolPerCache, returnErr
}

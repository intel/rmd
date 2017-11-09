package rdtpool

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/intel/rmd/lib/cache"
	"github.com/intel/rmd/lib/proxyclient"
	util "github.com/intel/rmd/lib/util"
	"github.com/intel/rmd/util/rdtpool/base"
	"github.com/intel/rmd/util/rdtpool/config"
)

var osGroupReserve = &base.Reserved{}
var osOnce sync.Once

// GetOSGroupReserve returns os reserved resource group
func GetOSGroupReserve() (base.Reserved, error) {
	var returnErr error
	osOnce.Do(func() {
		conf := config.NewOSConfig()
		osCPUbm, err := base.CPUBitmaps([]string{conf.CPUSet})
		if err != nil {
			returnErr = err
			return
		}
		osGroupReserve.AllCPUs = osCPUbm

		level := syscache.GetLLC()
		syscaches, err := syscache.GetSysCaches(int(level))
		if err != nil {
			returnErr = err
			return
		}

		// We though the ways number are same on all caches ID
		// FIXME if exception, fix it.
		ways, _ := strconv.Atoi(syscaches["0"].WaysOfAssociativity)
		if conf.CacheWays > uint(ways) {
			returnErr = fmt.Errorf("The request OSGroup cache ways %d is larger than available %d",
				conf.CacheWays, ways)
			return
		}

		schemata := map[string]*util.Bitmap{}
		osCPUs := map[string]*util.Bitmap{}

		for _, sc := range syscaches {
			bm, _ := base.CPUBitmaps([]string{sc.SharedCPUList})
			osCPUs[sc.ID] = osCPUbm.And(bm)
			if osCPUs[sc.ID].IsEmpty() {
				schemata[sc.ID], returnErr = base.CacheBitmaps("0")
				if returnErr != nil {
					return
				}
			} else {
				mask := strconv.FormatUint(1<<conf.CacheWays-1, 16)
				//FIXME  check RMD for the bootcheck.
				schemata[sc.ID], returnErr = base.CacheBitmaps(mask)
				if returnErr != nil {
					return
				}
			}
		}
		osGroupReserve.CPUsPerNode = osCPUs
		osGroupReserve.Schemata = schemata
	})

	return *osGroupReserve, returnErr
}

// SetOSGroup sets os group
func SetOSGroup() error {
	reserve, err := GetOSGroupReserve()
	if err != nil {
		return err
	}

	allres := proxyclient.GetResAssociation()
	osGroup := allres["."]
	originBM, err := base.CPUBitmaps(osGroup.CPUs)
	if err != nil {
		return err
	}

	// NOTE , simpleness, brutal. Stolen CPUs from other groups.
	newBM := originBM.Or(reserve.AllCPUs)
	osGroup.CPUs = newBM.ToString()

	level := syscache.GetLLC()
	cacheLevel := "L" + strconv.FormatUint(uint64(level), 10)
	schemata, _ := GetAvailableCacheSchemata(allres, []string{"infra", "."}, "none", cacheLevel)

	for i, v := range osGroup.Schemata[cacheLevel] {
		cacheID := strconv.Itoa(int(v.ID))
		if !reserve.CPUsPerNode[cacheID].IsEmpty() {
			// OSGroup is the first Group, use the edge cache ways.
			// FIXME , left or right cache ways, need to be check.
			conf := config.NewOSConfig()
			request, _ := base.CacheBitmaps(strconv.FormatUint(1<<conf.CacheWays-1, 16))
			// NOTE , simpleness, brutal. Reset Cache for OS Group,
			// even the cache is occupied by other group.
			availableWays := schemata[cacheID].Or(request)
			expectWays := availableWays.ToBinStrings()[0]

			osGroup.Schemata[cacheLevel][i].Mask = strconv.FormatUint(1<<uint(len(expectWays))-1, 16)
		} else {
			osGroup.Schemata[cacheLevel][i].Mask = base.GetCosInfo().CbmMask
		}
	}
	if err := proxyclient.Commit(osGroup, "."); err != nil {
		return err
	}
	return nil
}

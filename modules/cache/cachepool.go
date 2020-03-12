package cache

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/intel/rmd/modules/cache/config"
	util "github.com/intel/rmd/utils/bitmap"
)

var cachePoolReserved = make(map[string]*Reserved, 0)
var cachePoolOnce sync.Once

// helper function to get Reserved resource
func getReservedCache(
	wayCandidate int,
	wayOffset, osCacheWays uint,
	osCPUbm *util.Bitmap,
	sysc map[string]SysCache) (*Reserved, error) {

	r := &Reserved{}

	schemata := map[string]*util.Bitmap{}
	osCPUs := map[string]*util.Bitmap{}
	var err error

	for _, sc := range sysc {
		wc := wayCandidate
		bm, _ := BitmapsCPUWrapper([]string{sc.SharedCPUList})
		osCPUs[sc.ID] = osCPUbm.And(bm)
		// no os group on this cache id
		if !osCPUs[sc.ID].IsEmpty() {
			wc = wc << osCacheWays
		}
		wc = wc << wayOffset
		mask := strconv.FormatUint(uint64(wc), 16)
		schemata[sc.ID], err = BitmapsCacheWrapper(mask)
		if err != nil {
			return r, err
		}
	}

	r.Schemata = schemata
	return r, nil
}

// GetCachePoolLayout returns cache pool layout based on configuration
func GetCachePoolLayout() (map[string]*Reserved, error) {
	var returnErr error
	cachePoolOnce.Do(func() {
		poolConf := config.NewCachePoolConfig()
		osConf := config.NewOSConfig()
		ways := GetCosInfo().CbmMaskLen

		if osConf.CacheWays+poolConf.Guarantee+poolConf.Besteffort+poolConf.Shared > uint(ways) {
			returnErr = fmt.Errorf(
				"Error config: Guarantee + Besteffort + Shared + OS reserved ways should be less or equal to %d", ways)
			return
		}

		// set layout for cache pool
		level := GetLLC()
		syscaches, err := GetSysCaches(int(level))
		osCPUbm, err := BitmapsCPUWrapper([]string{osConf.CPUSet})

		if err != nil {
			returnErr = err
			return
		}

		if poolConf.Guarantee > 0 {
			wc := 1<<poolConf.Guarantee - 1
			resev, err := getReservedCache(wc,
				0,
				osConf.CacheWays,
				osCPUbm,
				syscaches)
			if err != nil {
				returnErr = err
				return
			}
			cachePoolReserved[Guarantee] = resev
		}

		if poolConf.Besteffort > 0 {
			wc := 1<<poolConf.Besteffort - 1
			resev, err := getReservedCache(wc,
				poolConf.Guarantee,
				osConf.CacheWays,
				osCPUbm,
				syscaches)

			if err != nil {
				returnErr = err
				return
			}
			cachePoolReserved[Besteffort] = resev
			cachePoolReserved[Besteffort].Shrink = poolConf.Shrink
		}

		if poolConf.Shared > 0 {
			wc := 1<<poolConf.Shared - 1
			resev, err := getReservedCache(wc,
				poolConf.Guarantee+poolConf.Besteffort,
				osConf.CacheWays,
				osCPUbm,
				syscaches)

			if err != nil {
				returnErr = err
				return
			}
			cachePoolReserved[Shared] = resev
			cachePoolReserved[Shared].Name = Shared
			cachePoolReserved[Shared].Quota = poolConf.MaxAllowedShared
		}
	})

	return cachePoolReserved, returnErr
}

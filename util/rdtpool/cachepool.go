package rdtpool

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/intel/rmd/lib/cache"
	util "github.com/intel/rmd/lib/util"
	"github.com/intel/rmd/util/rdtpool/base"
	"github.com/intel/rmd/util/rdtpool/config"
)

var cachePoolReserved = make(map[string]*base.Reserved, 0)
var cachePoolOnce sync.Once

// helper function to get Reserved resource
func getReservedCache(
	wayCandidate int,
	wayOffset, osCacheWays uint,
	osCPUbm *util.Bitmap,
	sysc map[string]syscache.SysCache) (*base.Reserved, error) {

	r := &base.Reserved{}

	schemata := map[string]*util.Bitmap{}
	osCPUs := map[string]*util.Bitmap{}
	var err error

	for _, sc := range sysc {
		wc := wayCandidate
		bm, _ := base.CPUBitmaps([]string{sc.SharedCPUList})
		osCPUs[sc.ID] = osCPUbm.And(bm)
		// no os group on this cache id
		if !osCPUs[sc.ID].IsEmpty() {
			wc = wc << osCacheWays
		}
		wc = wc << wayOffset
		mask := strconv.FormatUint(uint64(wc), 16)
		schemata[sc.ID], err = base.CacheBitmaps(mask)
		if err != nil {
			return r, err
		}
	}

	r.Schemata = schemata
	return r, nil
}

// GetCachePoolLayout returns cache pool layout based on configuration
func GetCachePoolLayout() (map[string]*base.Reserved, error) {
	var returnErr error
	cachePoolOnce.Do(func() {
		poolConf := config.NewCachePoolConfig()
		osConf := config.NewOSConfig()
		ways := base.GetCosInfo().CbmMaskLen

		if osConf.CacheWays+poolConf.Guarantee+poolConf.Besteffort+poolConf.Shared > uint(ways) {
			returnErr = fmt.Errorf(
				"Error config: Guarantee + Besteffort + Shared + OS reserved ways should be less or equal to %d", ways)
			return
		}

		// set layout for cache pool
		level := syscache.GetLLC()
		syscaches, err := syscache.GetSysCaches(int(level))
		osCPUbm, err := base.CPUBitmaps([]string{osConf.CPUSet})

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

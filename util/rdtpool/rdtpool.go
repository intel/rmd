package rdtpool

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/intel/rmd/lib/resctrl"
	libutil "github.com/intel/rmd/lib/util"
	"github.com/intel/rmd/util"
	"github.com/intel/rmd/util/rdtpool/base"
)

// A map that contains all reserved resource information
// Use resource key as index, the key could be as following:
//
// OS: os group information, it contains reserved cache information.
//     Required.
// INFRA: infra group information.
//        Optional.
// GUARANTEE: guarantee cache pool information, it's a pool instead of a
//            resource group. When try to allocate max_cache = min_cache,
//            use the mask in guarantee pool.
//            Optional.
// BESTEFFORT: besteffort pool information, it's a pool instead of a resource
//             group. When try to allocate max_cache > min_cache, allocate
//             from this pool.
//             Optional.
// SHARED: shared group, it's a resource group instead of a pool. When try
//         to allocate max_cache == min_cache == 0, just add cpus, tasks IDs
//         to this resource group. Need to count how many workload in this
//         resource group when calculating hosptility score.
//         Optional

const (
	// OS is os resource group name
	OS = "os"
	// Infra is infra resource group name
	Infra = "infra"
	// Guarantee is guarantee resource group name
	Guarantee = "guarantee"
	// Besteffort is besteffort resource group name
	Besteffort = "besteffort"
	// Shared is shared resource group name
	Shared = "shared"
)

// ReservedInfo is all reserved resource group inforamtaion
var ReservedInfo map[string]*base.Reserved
var revinfoOnce sync.Once

// GetReservedInfo returns all reserved information
func GetReservedInfo() map[string]*base.Reserved {

	revinfoOnce.Do(func() {
		ReservedInfo = make(map[string]*base.Reserved, 10)

		r, err := GetOSGroupReserve()
		if err == nil {
			ReservedInfo[OS] = &r
		}

		fr, err := GetInfraGroupReserve()
		if err == nil {
			ReservedInfo[Infra] = &fr
		}

		poolinfo, err := GetCachePoolLayout()
		if err == nil {
			for k, v := range poolinfo {
				ReservedInfo[k] = v
			}
		}
	})

	return ReservedInfo
}

// GetAvailableCacheSchemata returns available schemata of caches from
// specific pool: guarantee, besteffort, shared or just none
func GetAvailableCacheSchemata(allres map[string]*resctrl.ResAssociation,
	ignoreGroups []string,
	pool string,
	cacheLevel string) (map[string]*libutil.Bitmap, error) {

	GetReservedInfo()
	// FIXME  A central util to generate schemata Bitmap
	schemata := map[string]*libutil.Bitmap{}

	if len(allres) >= base.GetCosInfo().NumClosids {
		return nil, fmt.Errorf("error, not enough CLOS on host, %d used", len(allres))
	}

	if pool == "none" {
		for k := range ReservedInfo[OS].Schemata {
			schemata[k], _ = base.CacheBitmaps(base.GetCosInfo().CbmMask)
		}
	} else {
		resv, ok := ReservedInfo[pool]
		if !ok {
			return nil, fmt.Errorf("error doesn't support pool %s", pool)
		}

		for k, v := range resv.Schemata {
			schemata[k] = v
		}
	}

	for k, v := range allres {
		if util.HasElem(ignoreGroups, k) {
			continue
		}
		if sv, ok := v.Schemata[cacheLevel]; ok {
			for _, cv := range sv {
				k := strconv.Itoa(int(cv.ID))
				bm, _ := base.CacheBitmaps(cv.Mask)
				// And check cpu list is empty
				if cv.Mask == base.GetCosInfo().CbmMask {
					continue
				}
				schemata[k] = schemata[k].Axor(bm)
			}
		}
	}
	return schemata, nil
}

// GetCachePoolName will return pool name based on MaxCache, MinCache
func GetCachePoolName(MaxWays, MinWays uint32) (string, error) {
	if MaxWays == 0 {
		return Shared, nil
	} else if MaxWays > MinWays && MinWays != 0 {
		return Besteffort, nil
	} else if MaxWays == MinWays {
		return Guarantee, nil
	}

	return "", fmt.Errorf("max_cache=%d, min_cache=%d, doens't support", MaxWays, MinWays)
}

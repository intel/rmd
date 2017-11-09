package cache

// This model is just for cache info
// We can ref k8s

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	rmderror "github.com/intel/rmd/api/error"
	"github.com/intel/rmd/lib/cache"
	"github.com/intel/rmd/lib/cpu"
	"github.com/intel/rmd/lib/proc"
	"github.com/intel/rmd/lib/resctrl"
	"github.com/intel/rmd/model/policy"
	"github.com/intel/rmd/util/rdtpool"
	"github.com/intel/rmd/util/rdtpool/base"
)

// SizeMap is the map to bits of unit
var SizeMap = map[string]uint32{
	"K": 1024,
	"M": 1024 * 1024,
}

// Info is details of cache
type Info struct {
	ID               uint32 `json:"cache_id"`
	NumWays          uint32
	NumSets          uint32
	NumPartitions    uint32
	LineSize         uint32
	TotalSize        uint32 `json:"total_size"`
	WaySize          uint32
	NumClasses       uint32
	WayContention    uint64
	CacheLevel       uint32
	Location         string            `json:"location_on_socket"`
	Node             string            `json:"location_on_node"`
	ShareCPUList     string            `json:"share_cpu_list"`
	AvaliableWays    string            `json:"avaliable_ways"`
	AvaliableCPUs    string            `json:"avaliable_cpus"`
	AvaliableIsoCPUs string            `json:"avaliable_isolated_cpus"`
	AvaliablePolicy  map[string]uint32 `json:"avaliable_policy"` // should move out here
}

// Infos is group of cache info
type Infos struct {
	Num    uint32          `json:"number"`
	Caches map[uint32]Info `json:"Caches"`
}

// Summary is summary of cache
type Summary struct {
	Num int      `json:"number"`
	IDs []string `json:"caches_id"`
}

// CachesSummary Cat, Cqm seems CPU's feature.
// Should be better
// type Rdt struct {
// 	Cat   bool
// 	CatOn bool
// 	Cdp   bool
// 	CdpOn bool
// }
type CachesSummary struct {
	Rdt    bool               `json:"rdt"`
	Cqm    bool               `json:"cqm"`
	Cdp    bool               `json:"cdp"`
	CdpOn  bool               `json:"cdp_enable"`
	Cat    bool               `json:"cat"`
	CatOn  bool               `json:"cat_enable"`
	Caches map[string]Summary `json:"caches"`
}

func (c *CachesSummary) getCaches() error {
	levs := syscache.AvailableCacheLevel()

	c.Caches = make(map[string]Summary)
	for _, l := range levs {
		summary := &Summary{}
		il, err := strconv.Atoi(l)
		if err != nil {
			return err
		}
		caches, err := syscache.GetSysCaches(il)
		if err != nil {
			return err
		}
		for _, v := range caches {
			summary.IDs = append(summary.IDs, v.ID)

		}
		summary.Num = len(caches)
		c.Caches["l"+l] = *summary
	}
	return nil
}

// Get returns summary of cache
func (c *CachesSummary) Get() error {
	var err error
	var flag bool
	flag, err = proc.IsRdtAvailiable()
	if err != nil {
		return nil
	}
	c.Rdt = flag
	c.Cat = flag

	flag, err = proc.IsCqmAvailiable()
	if err != nil {
		return nil
	}
	c.Cqm = flag

	flag, err = proc.IsCdpAvailiable()
	if err != nil {
		return nil
	}
	c.Cdp = flag

	flag = proc.IsEnableCat()
	c.CatOn = flag

	flag = proc.IsEnableCdp()
	c.CdpOn = flag

	err = c.getCaches()
	if err != nil {
		return nil
	}

	return nil
}

// Convert a string cache size to uint32 in B
// eg: 1K = 1024
func convertCacheSize(size string) uint32 {
	unit := size[len(size)-1:]

	s := strings.TrimRight(size, unit)

	isize, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return uint32(isize) * SizeMap[unit]
}

// GetByLevel returns cache info by cache level
func (c *Infos) GetByLevel(level uint32) *rmderror.AppError {

	llc := syscache.GetLLC()

	if llc != level {
		err := fmt.Errorf("Don't support cache level %d, Only expose last level cache %d", level, llc)
		return rmderror.NewAppError(http.StatusBadRequest,
			"Error to get available cache", err)
	}

	// syscache.AvailableCacheLevel return []string
	levs := syscache.AvailableCacheLevel()
	sort.Strings(levs)

	syscaches, err := syscache.GetSysCaches(int(level))
	if err != nil {
		return rmderror.NewAppError(http.StatusInternalServerError,
			"Error to get available cache", err)
	}

	cacheLevel := "L" + strconv.FormatUint(uint64(level), 10)

	allres := resctrl.GetResAssociation()
	av, _ := rdtpool.GetAvailableCacheSchemata(allres, []string{"infra"}, "none", cacheLevel)

	c.Caches = make(map[uint32]Info)

	for _, sc := range syscaches {
		id, _ := strconv.Atoi(sc.ID)
		_, ok := c.Caches[uint32(id)]
		if ok {
			// syscache.GetSysCaches returns caches per each CPU, there maybe
			// multiple cpus chares on same cache.
			continue
		} else {
			// TODO: NumPartitions uint32,  NumClasses    uint32
			//       WayContention uint64,  Location string
			newCachdinfo := Info{}

			ui32, _ := strconv.Atoi(sc.CoherencyLineSize)
			newCachdinfo.LineSize = uint32(ui32)

			ui32, _ = strconv.Atoi(sc.NumberOfSets)
			newCachdinfo.NumSets = uint32(ui32)

			// FIXME the relation between NumWays and sc.PhysicalLinePartition
			ui32, _ = strconv.Atoi(sc.WaysOfAssociativity)
			newCachdinfo.NumWays = uint32(ui32)
			newCachdinfo.WaySize = newCachdinfo.LineSize * newCachdinfo.NumSets

			newCachdinfo.ID = uint32(id)
			newCachdinfo.TotalSize = convertCacheSize(sc.Size)
			newCachdinfo.ShareCPUList = sc.SharedCPUList
			newCachdinfo.CacheLevel = level

			cpuid := strings.SplitN(sc.SharedCPUList, "-", 2)[0]
			newCachdinfo.Location, _ = cpu.LocateOnSocket(cpuid)
			newCachdinfo.Node = cpu.LocateOnNode(cpuid)

			newCachdinfo.AvaliableWays = av[sc.ID].ToString()

			cpuPools, _ := rdtpool.GetCPUPools()
			defaultCpus, _ := base.CPUBitmaps(resctrl.GetResAssociation()["."].CPUs)
			newCachdinfo.AvaliableCPUs = cpuPools["all"][sc.ID].And(defaultCpus).ToHumanString()
			newCachdinfo.AvaliableIsoCPUs = cpuPools["isolated"][sc.ID].And(defaultCpus).ToHumanString()

			p, err := policy.GetDefaultPlatformPolicy()
			if err != nil {
				return rmderror.NewAppError(http.StatusInternalServerError,
					"Error to get policy", err)
			}
			ap := make(map[string]uint32)
			//ap_counter := make(map[string]int)
			for _, pv := range p {
				// pv is policy.CATConfig.Catpolicy
				for t := range pv {
					// t is the policy tier name
					tier, err := policy.GetDefaultPolicy(t)
					if err != nil {
						return rmderror.NewAppError(http.StatusInternalServerError,
							"Error to get policy", err)
					}

					iMax, err := strconv.Atoi(tier["MaxCache"])
					if err != nil {
						return rmderror.NewAppError(http.StatusInternalServerError,
							"Error to get max cache", err)
					}
					iMin, err := strconv.Atoi(tier["MinCache"])
					if err != nil {
						return rmderror.NewAppError(http.StatusInternalServerError,
							"Error to get min cache", err)
					}

					getAvailablePolicyCount(ap, iMax, iMin, allres, t, cacheLevel, sc.ID)

				}

			}
			newCachdinfo.AvaliablePolicy = ap

			c.Caches[uint32(id)] = newCachdinfo
			c.Num = c.Num + 1
		}
	}

	return nil
}

func getAvailablePolicyCount(ap map[string]uint32,
	iMax, iMin int,
	allres map[string]*resctrl.ResAssociation,
	tier, cacheLevel, cID string) error {

	var ways int

	reserved := rdtpool.GetReservedInfo()

	pool, _ := rdtpool.GetCachePoolName(uint32(iMax), uint32(iMin))

	switch pool {
	case rdtpool.Guarantee:
		ways = iMax
	case rdtpool.Besteffort:
		ways = iMin
	case rdtpool.Shared:
		// TODO get live count ?
		ap[tier] = uint32(reserved[rdtpool.Shared].Quota)
		return nil
	}

	pav, _ := rdtpool.GetAvailableCacheSchemata(allres, []string{"infra", "."}, pool, cacheLevel)
	ap[tier] = 0
	freeBitmapStrs := pav[cID].ToBinStrings()

	for _, val := range freeBitmapStrs {
		if val[0] == '1' {
			valLen := len(val)
			ap[tier] += uint32(valLen / ways)
		}
	}

	return nil
}

package cache

// This model is just for cache info
// We can ref k8s

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	rmderror "github.com/intel/rmd/internal/error"
	proxyclient "github.com/intel/rmd/internal/proxy/client"
	"github.com/intel/rmd/modules/policy"
	util "github.com/intel/rmd/utils/bitmap"
	"github.com/intel/rmd/utils/cpu"
	"github.com/intel/rmd/utils/proc"
	"github.com/intel/rmd/utils/resctrl"
	log "github.com/sirupsen/logrus"
)

const (
	// SysCPUPath is patch of cpu device
	SysCPUPath = "/sys/devices/system/cpu/"
)

// SysCache is struct of cache of host
type SysCache struct {
	CoherencyLineSize     string
	ID                    string
	Level                 string
	NumberOfSets          string
	PhysicalLinePartition string
	SharedCPUList         string
	SharedCPUMap          string
	Size                  string
	Type                  string
	WaysOfAssociativity   string
}

// SizeMap is the map to bits of unit
var SizeMap = map[string]uint32{
	"K": 1024,
	"M": 1024 * 1024,
}

// Info is details of cache
type Info struct {
	ID                uint32 `json:"cache_id"`
	NumWays           uint32
	NumSets           uint32
	NumPartitions     uint32
	LineSize          uint32
	TotalSize         uint32 `json:"total_size"`
	WaySize           uint32
	NumClasses        uint32
	WayContention     uint64
	CacheLevel        uint32
	Location          string            `json:"location_on_socket"`
	Node              string            `json:"location_on_node"`
	ShareCPUList      string            `json:"share_cpu_list"`
	AvailableWays     string            `json:"available_ways"`
	AvailableCPUs     string            `json:"available_cpus"`
	AvailableIsoCPUs  string            `json:"available_isolated_cpus"`
	AvailablePolicy   map[string]uint32 `json:"available_policy"` // should move out here
	AvailableWaysPool map[string]string `json:"available_ways_pool"`
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

// CosInfo contains info about Class Of Service
type CosInfo struct {
	CbmMaskLen int
	MinCbmBits int
	NumClosids int
	CbmMask    string
}

var catCosInfo = &CosInfo{0, 0, 0, ""}
var infoOnce sync.Once

// Reserved schemata information
type Reserved struct {
	AllCPUs     *util.Bitmap            //cpu bit mask
	SchemaNum   int                     // Numbers of schema
	Name        string                  // Resource group name if it is a resource group instead of pool
	Schemata    map[string]*util.Bitmap // Schema list
	CPUsPerNode map[string]*util.Bitmap // CPU bitmap
	Quota       uint                    // Max allowed usage for this resource
	Shrink      bool                    // Wether shrink in BE pool
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
	levs, err := AvailableCacheLevel()
	if err != nil {
		return err
	}
	c.Caches = make(map[string]Summary)
	for _, l := range levs {
		summary := &Summary{}
		il, err := strconv.Atoi(l)
		if err != nil {
			return err
		}
		caches, err := GetSysCaches(il)
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
	flag, err = proc.IsRdtAvailable()
	if err != nil {
		return nil
	}
	c.Rdt = flag
	c.Cat = flag

	flag, err = proc.IsCqmAvailable()
	if err != nil {
		return nil
	}
	c.Cqm = flag

	flag, err = proc.IsCdpAvailable()
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
func (c *Infos) GetByLevel(level uint32) error {

	llc := GetLLC()

	if llc != level {
		err := fmt.Errorf("Don't support cache level %d, Only expose last level cache %d", level, llc)
		return rmderror.NewAppError(http.StatusBadRequest,
			"Error to get available cache", err)
	}

	// syscache.AvailableCacheLevel return []string
	levs, err := AvailableCacheLevel()
	if err != nil {
		return err
	}
	sort.Strings(levs)

	syscaches, err := GetSysCaches(int(level))
	if err != nil {
		return rmderror.NewAppError(http.StatusInternalServerError,
			"Error to get available cache", err)
	}

	cacheLevel := "L" + strconv.FormatUint(uint64(level), 10)

	allres := resctrl.GetResAssociation()
	av, err := GetAvailableCacheSchemata(allres, []string{"infra"}, "none", cacheLevel)
	if err != nil {
		return rmderror.AppErrorf(http.StatusInternalServerError,
			"Unable to read cache schemata; %s", err.Error())
	}

	avGuarantee, err := GetAvailableCacheSchemata(allres, []string{"infra", "."}, "guarantee", cacheLevel)
	if err != nil && !strings.Contains(err.Error(), "error doesn't support pool") {
		return rmderror.AppErrorf(http.StatusInternalServerError,
			"Unable to read cache schemata; %s", err.Error())
	}

	avBesteffort, err := GetAvailableCacheSchemata(allres, []string{"infra", "."}, "besteffort", cacheLevel)
	if err != nil && !strings.Contains(err.Error(), "error doesn't support pool") {
		return rmderror.AppErrorf(http.StatusInternalServerError,
			"Unable to read cache schemata; %s", err.Error())
	}

	avShared, err := GetAvailableCacheSchemata(allres, []string{"infra", "."}, "shared", cacheLevel)
	if err != nil && !strings.Contains(err.Error(), "error doesn't support pool") {
		return rmderror.AppErrorf(http.StatusInternalServerError,
			"Unable to read cache schemata; %s", err.Error())
	}

	avInfra, err := GetAvailableCacheSchemata(allres, []string{"infra", "."}, "infra", cacheLevel)
	if err != nil && !strings.Contains(err.Error(), "error doesn't support pool") {
		return rmderror.AppErrorf(http.StatusInternalServerError,
			"Unable to read cache schemata; %s", err.Error())
	}

	avOs, err := GetAvailableCacheSchemata(allres, []string{"infra", "."}, "os", cacheLevel)
	if err != nil && !strings.Contains(err.Error(), "error doesn't support pool") {
		return rmderror.AppErrorf(http.StatusInternalServerError,
			"Unable to read cache schemata; %s", err.Error())
	}

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

			newCachdinfo.AvailableWays = av[sc.ID].ToString()
			availWaysPool := make(map[string]string)
			if avGuarantee != nil {
				if elem, ok := avGuarantee[sc.ID]; ok {
					availWaysPool["guaranteed"] = elem.ToHumanString()
				}
			}
			if avBesteffort != nil {
				if elem, ok := avBesteffort[sc.ID]; ok {
					availWaysPool["besteffort"] = elem.ToHumanString()
				}
			}
			if avShared != nil {
				if elem, ok := avShared[sc.ID]; ok {
					availWaysPool["shared"] = elem.ToHumanString()
				}
			}
			if avInfra != nil {
				if elem, ok := avInfra[sc.ID]; ok {
					availWaysPool["infra"] = elem.ToHumanString()
				}
			}
			if avOs != nil {
				availWaysPool["os"] = avOs[sc.ID].ToHumanString()
			}
			newCachdinfo.AvailableWaysPool = availWaysPool

			cpuPools, _ := GetCPUPools()
			var defaultCpus *util.Bitmap
			if resAssoc, ok := resctrl.GetResAssociation()["."]; ok == true && resAssoc != nil {
				defaultCpus, _ = BitmapsCPUWrapper(resAssoc.CPUs)
			}
			if item, ok := cpuPools["all"][sc.ID]; ok {
				newCachdinfo.AvailableCPUs = item.And(defaultCpus).ToHumanString()
			}
			if item, ok := cpuPools["isolated"][sc.ID]; ok {
				newCachdinfo.AvailableIsoCPUs = item.And(defaultCpus).ToHumanString()
			}

			defPolicy, policyErr := policy.GetDefaultPlatformPolicy()
			if policyErr != nil {
				log.Errorf("Failed to get default platform policy. Reason: %s", policyErr.Error())
			}

			availPolicy := make(map[string]uint32)
			for policyName, modules := range defPolicy {
				//get max cache
				iMax, err := strconv.Atoi(modules["cache"]["max"])
				if err != nil {
					log.Errorf("Error to get max cache. Reason: %s", err.Error())
					return rmderror.NewAppError(http.StatusInternalServerError,
						"Error to get max cache", err)
				}

				//get min cache
				iMin, err := strconv.Atoi(modules["cache"]["min"])
				if err != nil {
					log.Errorf("Error to get min cache. Reason: %s", err.Error())
					return rmderror.NewAppError(http.StatusInternalServerError,
						"Error to get min cache", err)
				}

				err = getAvailablePolicyCount(availPolicy, iMax, iMin, allres, policyName, cacheLevel, sc.ID)
				if err != nil {
					log.Errorf("Failed to get available policy count. Reason: %s", err.Error())
					return rmderror.AppErrorf(http.StatusInternalServerError,
						"Failed to get available policy count. Reason: %s", err.Error())
				}
			}
			newCachdinfo.AvailablePolicy = availPolicy

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

	reserved := GetReservedInfo()

	pool, _ := GetCachePoolName(uint32(iMax), uint32(iMin))

	switch pool {
	case Guarantee:
		ways = iMax
	case Besteffort:
		ways = iMin
	case Shared:
		// TODO get live count ?
		if r, ok := reserved[Shared]; !ok {
			// no Shared group was set
			ap[tier] = 0
		} else {
			ap[tier] = uint32(r.Quota)
		}
		return nil
	}

	ap[tier] = 0
	pav, err := GetAvailableCacheSchemata(allres, []string{"infra", "."}, pool, cacheLevel)

	if err != nil && !strings.Contains(err.Error(), "error doesn't support pool") {
		return err
	}

	if len(pav) == 0 {
		return nil
	}

	freeBitmapStrs := pav[cID].ToBinStrings()

	for _, val := range freeBitmapStrs {
		if val[0] == '1' {
			valLen := len(val)
			ap[tier] += uint32(valLen / ways)
		}
	}

	return nil
}

// GetCosInfo is Concurrency-safe.
func GetCosInfo() CosInfo {
	infoOnce.Do(func() {
		rcinfo := proxyclient.GetRdtCosInfo()
		level := GetLLC()
		targetLev := strconv.FormatUint(uint64(level), 10)
		cacheLevel := "l" + targetLev

		catCosInfo.CbmMaskLen = util.CbmLen(rcinfo[cacheLevel].CbmMask)
		catCosInfo.MinCbmBits = rcinfo[cacheLevel].MinCbmBits
		catCosInfo.NumClosids = rcinfo[cacheLevel].NumClosids
		catCosInfo.CbmMask = rcinfo[cacheLevel].CbmMask
	})
	return *catCosInfo
}

// BitmapsCPUWrapper is a wrapper for Bitmap
func BitmapsCPUWrapper(cpuids interface{}) (*util.Bitmap, error) {
	// FIXME need a wrap for CPU bitmap.
	cpunum := cpu.HostCPUNum()
	if cpunum == 0 {
		// return nil or an empty Bitmap?
		var bm *util.Bitmap
		return bm, fmt.Errorf("Unable to get Total CPU numbers on Host")
	}
	return util.NewBitmap(cpunum, cpuids)
}

// BitmapsCacheWrapper is a wrapper for Cache bitmap
func BitmapsCacheWrapper(bitmask interface{}) (*util.Bitmap, error) {
	// FIXME need a wrap for CPU bitmap.
	len := GetCosInfo().CbmMaskLen
	if len == 0 {
		// return nil or an empty Bitmap?
		var bm *util.Bitmap
		return bm, fmt.Errorf("Unable to get Total cache ways on Host")
	}
	return util.NewBitmap(len, bitmask)
}

// /sys/devices/system/cpu/cpu*/cache/index*/*
// pass var caches map[string]SysCache
/*
usage:
    ignore := []string{"uevent"}
    syscache := &SysCache{}
	filepath.Walk(dir, getSysCache(ignore, syscache))
*/
func getSysCache(ignore []string, cache *SysCache) filepath.WalkFunc {

	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// add log
			return nil
		}

		// ignore dir.
		f := filepath.Base(path)
		if info.IsDir() {
			for _, d := range ignore {
				if d == f {
					return filepath.SkipDir
				}
			}
			return nil
		}
		for _, d := range ignore {
			if d == f {
				return nil
			}
		}

		name := strings.Replace(strings.Title(strings.Replace(f, "_", " ", -1)), " ", "", -1)
		data, err := ioutil.ReadFile(path)
		if err != nil {
			// add log
			return err
		}
		// golint does allow us to define a struct name like Cpu, Id
		// so we need to deal these cases:
		// handle Cpu -> CPU case
		// handle Id -> ID
		if strings.Contains(name, "Cpu") {
			name = strings.Replace(name, "Cpu", "CPU", -1)
		} else if name == "Id" {
			name = "ID"
		}
		return util.SetField(cache, name, strings.TrimSpace(string(data)))
	}
}

// GetSysCaches traverse all sys cache file for a specify level
func GetSysCaches(level int) (map[string]SysCache, error) {
	ignore := []string{"uevent", "power"}
	caches := make(map[string]SysCache)
	files, err := filepath.Glob(SysCPUPath + "cpu*/cache/index" + strconv.Itoa(level))
	if err != nil {
		return caches, err
	}

	for _, f := range files {
		cache := &SysCache{}
		err := filepath.Walk(f, getSysCache(ignore, cache))
		if err != nil {
			return caches, err
		}
		if _, ok := caches[cache.ID]; !ok {
			caches[cache.ID] = *cache
		}
	}
	return caches, nil
}

// AvailableCacheLevel returns the L2 and L3 level cache, strip L1 cache.
// By default, get the info from cpu0 path, any issue?
// The type of return should be string or int?
func AvailableCacheLevel() ([]string, error) {
	var levels []string
	files, err := filepath.Glob(SysCPUPath + "cpu0/cache/index*/level")
	if err != nil {
		return levels, err
	}
	for _, f := range files {
		// NOTE: ReadFile() function does not cause any DOS Resource Exhaustion here
		// since the file being read is a known system file (/sys/devices/system/cpu/*)
		dat, err := ioutil.ReadFile(f)
		if err != nil {
			return levels, err
		}
		sdat := strings.TrimRight(string(dat), "\n")
		if 0 != strings.Compare("1", sdat) {
			levels = append(levels, sdat)
		}
	}
	return levels, nil
}

// GetLLC return the last level of the cache on the host
func GetLLC() uint32 {
	avl, err := AvailableCacheLevel()
	if err != nil {
		return 0
	}
	sort.Sort(sort.Reverse(sort.StringSlice(avl)))
	l, err := strconv.Atoi(avl[0])
	if err != nil {
		return 0
	}
	return uint32(l)
}

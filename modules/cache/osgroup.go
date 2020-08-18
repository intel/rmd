package cache

import (
	"errors"
	"fmt"
	"strconv"
	"sync"

	proxyclient "github.com/intel/rmd/internal/proxy/client"
	"github.com/intel/rmd/modules/cache/config"
	util "github.com/intel/rmd/utils/bitmap"
	"github.com/intel/rmd/utils/pqos"
	"github.com/intel/rmd/utils/proc"
)

var osGroupReserve = &Reserved{}
var osOnce sync.Once

// GetOSGroupReserve returns os reserved resource group
func GetOSGroupReserve() (Reserved, error) {
	var returnErr error
	osOnce.Do(func() {
		conf := config.NewOSConfig()
		osCPUbm, err := BitmapsCPUWrapper([]string{conf.CPUSet})
		if err != nil {
			returnErr = err
			return
		}
		osGroupReserve.AllCPUs = osCPUbm

		level := GetLLC()
		syscaches, err := GetSysCaches(int(level))
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
			bm, _ := BitmapsCPUWrapper([]string{sc.SharedCPUList})
			osCPUs[sc.ID] = osCPUbm.And(bm)
			if osCPUs[sc.ID].IsEmpty() {
				schemata[sc.ID], returnErr = BitmapsCacheWrapper("0")
				if returnErr != nil {
					return
				}
			} else {
				mask := strconv.FormatUint(1<<conf.CacheWays-1, 16)
				//FIXME  check RMD for the bootcheck.
				schemata[sc.ID], returnErr = BitmapsCacheWrapper(mask)
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

	allres := proxyclient.GetResAssociation(pqos.GetAvailableCLOSes())
	osGroup, ok := allres[pqos.OSGroupCOS]
	if !ok {
		return errors.New("OSGroup not found")
	}
	// Removing "MB" from the Cache Schemata because it causes error while writing Mbps value
	// Resctrl bug: approximates(takes the ceil) the given value. When MBA mbps max value given
	// then it takes the ceil of the value and it goes off range. Hence deleting default MBA values.
	_, ok = osGroup.CacheSchemata["MB"]
	if ok {
		// PQOS TODO MBA mode is now set in rmd.toml and verified by server
		// so code below can be replaced with config flag checking
		for _, v := range osGroup.CacheSchemata["MB"] {
			if v.Mask > strconv.Itoa(MaxMBAPercentage) {
				proc.SetMbaMbpsMode(true)
			} else {
				proc.SetMbaMbpsMode(false)
			}
		}
		delete(osGroup.CacheSchemata, "MB")
	} else {
		proc.SetMbaMbpsMode(false)
	}
	originBM, err := BitmapsCPUWrapper(osGroup.CPUs)
	if err != nil {
		return err
	}

	// NOTE , simpleness, brutal. Stolen CPUs from other groups.
	newBM := originBM.Or(reserve.AllCPUs)
	osGroup.CPUs = newBM.ToString()

	level := GetLLC()
	cacheLevel := "L" + strconv.FormatUint(uint64(level), 10)
	schemata, err := GetAvailableCacheSchemata(allres, []string{pqos.InfraGoupCOS, pqos.OSGroupCOS}, "none", cacheLevel)

	if err != nil {
		return err
	}

	for i, v := range osGroup.CacheSchemata[cacheLevel] {
		cacheID := strconv.Itoa(int(v.ID))
		// OSGroup is the first Group, use the edge cache ways.
		// FIXME , left or right cache ways, need to be check.
		conf := config.NewOSConfig()
		request, _ := BitmapsCacheWrapper(strconv.FormatUint(1<<conf.CacheWays-1, 16))
		// NOTE , simpleness, brutal. Reset Cache for OS Group,
		// even the cache is occupied by other group.
		availableWays := schemata[cacheID].Or(request)
		expectWays := availableWays.ToBinStrings()[0]

		osGroup.CacheSchemata[cacheLevel][i].Mask = strconv.FormatUint(1<<uint(len(expectWays))-1, 16)
	}

	return proxyclient.Commit(osGroup, pqos.OSGroupCOS)
}

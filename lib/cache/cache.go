// +build linux

package syscache

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	libutil "github.com/intel/rmd/lib/util"
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
		return libutil.SetField(cache, name, strings.TrimSpace(string(data)))
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
func AvailableCacheLevel() []string {
	var levels []string
	files, _ := filepath.Glob(SysCPUPath + "cpu0/cache/index*/level")
	for _, f := range files {
		dat, _ := ioutil.ReadFile(f)
		sdat := strings.TrimRight(string(dat), "\n")
		if 0 != strings.Compare("1", sdat) {
			levels = append(levels, sdat)
		}
	}
	return levels
}

// GetLLC return the last level of the cache on the host
func GetLLC() uint32 {
	avl := AvailableCacheLevel()
	sort.Sort(sort.Reverse(sort.StringSlice(avl)))
	l, err := strconv.Atoi(avl[0])
	if err != nil {
		return 0
	}
	return uint32(l)
}

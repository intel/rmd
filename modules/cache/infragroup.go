package cache

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/gobwas/glob"
	log "github.com/sirupsen/logrus"

	proxyclient "github.com/intel/rmd/internal/proxy/client"
	"github.com/intel/rmd/modules/cache/config"
	util "github.com/intel/rmd/utils/bitmap"
	"github.com/intel/rmd/utils/pqos"
	"github.com/intel/rmd/utils/proc"
	"github.com/intel/rmd/utils/resctrl"
)

var infraGroupReserve = &Reserved{}
var infraOnce sync.Once

func getGlobTasks() []glob.Glob {
	conf := config.NewInfraConfig()
	l := len(conf.Tasks)
	gs := make([]glob.Glob, l, l)
	for i, v := range conf.Tasks {
		g := glob.MustCompile(v)
		gs[i] = g
	}
	return gs
}

// GetInfraGroupReserve returns reserved infra group
// NOTE  This group can be merged into GetOSGroupReserve
func GetInfraGroupReserve() (Reserved, error) {
	var returnErr error
	infraOnce.Do(func() {
		conf := config.NewInfraConfig()
		if conf == nil || conf.CacheWays == 0 {
			return
		}
		infraCPUbm, err := BitmapsCPUWrapper([]string{conf.CPUSet})
		if err != nil {
			returnErr = err
			return
		}
		infraGroupReserve.AllCPUs = infraCPUbm

		level := GetLLC()
		syscaches, err := GetSysCaches(int(level))
		if err != nil {
			returnErr = err
			return
		}

		// NOTE  here we do not guarantee OS and Infra Group will avoid overlap.
		// We can FIX it on bootcheek.
		// We though the ways number are same on all caches ID
		// FIXME if exception, fix it.
		ways, _ := strconv.Atoi(syscaches["0"].WaysOfAssociativity)
		if conf.CacheWays > uint(ways) {
			returnErr = fmt.Errorf("The request InfraGroup cache ways %d is larger than available %d",
				conf.CacheWays, ways)
			return
		}

		schemata := map[string]*util.Bitmap{}
		infraCPUs := map[string]*util.Bitmap{}

		for _, sc := range syscaches {
			bm, _ := BitmapsCPUWrapper([]string{sc.SharedCPUList})
			infraCPUs[sc.ID] = infraCPUbm.And(bm)
			if infraCPUs[sc.ID].IsEmpty() {
				schemata[sc.ID], returnErr = BitmapsCacheWrapper("0")
				if returnErr != nil {
					return
				}
			} else {
				// FIXME  We need to confirm the location of DDIO caches.
				// We Put on the left ways, opposite position of OS group cache ways.
				ways := uint(GetCosInfo().CbmMaskLen)
				mask := strconv.FormatUint((1<<conf.CacheWays-1)<<(ways-conf.CacheWays), 16)
				//FIXME  check RMD for the bootcheck.
				schemata[sc.ID], returnErr = BitmapsCacheWrapper(mask)
				if returnErr != nil {
					return
				}
			}
		}

		infraGroupReserve.CPUsPerNode = infraCPUs
		infraGroupReserve.Schemata = schemata
	})

	return *infraGroupReserve, returnErr

}

// SetInfraGroup sets infra resource group based on configuration
func SetInfraGroup() error {
	conf := config.NewInfraConfig()
	if conf == nil || conf.CacheWays == 0 {
		return nil
	}

	reserve, err := GetInfraGroupReserve()
	if err != nil {
		return err
	}

	level := GetLLC()
	cacheLevel := "L" + strconv.FormatUint(uint64(level), 10)
	ways := GetCosInfo().CbmMaskLen
	// pqos.GetAvailableCLOSes() returns list of CLOSes still available for use
	allres := proxyclient.GetResAssociation(pqos.GetAvailableCLOSes())
	infraGroup, ok := allres[pqos.InfraGoupCOS]
	if !ok {
		infraGroup = resctrl.NewResAssociation()
		l := len(reserve.Schemata)
		infraGroup.CacheSchemata[cacheLevel] = make([]resctrl.CacheCos, l, l)
	}
	// Removing "MB" from the Cache Schemata because it causes error while writing Mbps value
	// Resctrl bug: approximates(takes the ceil) the given value. When MBA mbps max value given
	// then it takes the ceil of the value and it goes off range. Hence deleting default MBA values.
	_, ok = infraGroup.CacheSchemata["MB"]
	if ok {
		delete(infraGroup.CacheSchemata, "MB")
	}
	infraGroup.CPUs = reserve.AllCPUs.ToString()

	for k, v := range reserve.Schemata {
		id, _ := strconv.Atoi(k)
		var mask string
		if !reserve.CPUsPerNode[k].IsEmpty() {
			mask = v.ToString()
		} else {
			mask = strconv.FormatUint(1<<uint(ways)-1, 16)
		}
		cc := resctrl.CacheCos{
			ID:   uint8(id),
			Mask: mask,
		}
		infraGroup.CacheSchemata[cacheLevel][id] = cc
	}

	gt := getGlobTasks()
	tasks := []string{}
	ps := proc.ListProcesses()
	for k, v := range ps {
		for _, g := range gt {
			if g.Match(v.CmdLine) {
				tasks = append(tasks, k)
				log.Infof("Add task: %d to infra group. Command line: %s",
					v.Pid, strings.TrimSpace(v.CmdLine))
			}
		}
	}

	infraGroup.Tasks = append(infraGroup.Tasks, tasks...)

	return proxyclient.Commit(infraGroup, pqos.InfraGoupCOS)
}

package rdtpool

import (
	"fmt"
	"github.com/gobwas/glob"
	"github.com/intel/rmd/lib/cache"
	"github.com/intel/rmd/lib/proc"
	"github.com/intel/rmd/lib/resctrl"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"sync"

	"github.com/intel/rmd/lib/proxyclient"
	util "github.com/intel/rmd/lib/util"
	"github.com/intel/rmd/util/rdtpool/base"
	"github.com/intel/rmd/util/rdtpool/config"
)

var groupName = "infra"

var infraGroupReserve = &base.Reserved{}
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
func GetInfraGroupReserve() (base.Reserved, error) {
	var returnErr error
	infraOnce.Do(func() {
		conf := config.NewInfraConfig()
		if conf == nil {
			return
		}
		infraCPUbm, err := base.CPUBitmaps([]string{conf.CPUSet})
		if err != nil {
			returnErr = err
			return
		}
		infraGroupReserve.AllCPUs = infraCPUbm

		level := syscache.GetLLC()
		syscaches, err := syscache.GetSysCaches(int(level))
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
			bm, _ := base.CPUBitmaps([]string{sc.SharedCPUList})
			infraCPUs[sc.ID] = infraCPUbm.And(bm)
			if infraCPUs[sc.ID].IsEmpty() {
				schemata[sc.ID], returnErr = base.CacheBitmaps("0")
				if returnErr != nil {
					return
				}
			} else {
				// FIXME  We need to confirm the location of DDIO caches.
				// We Put on the left ways, opposite position of OS group cache ways.
				ways := uint(base.GetCosInfo().CbmMaskLen)
				mask := strconv.FormatUint((1<<conf.CacheWays-1)<<(ways-conf.CacheWays), 16)
				//FIXME  check RMD for the bootcheck.
				schemata[sc.ID], returnErr = base.CacheBitmaps(mask)
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
	if conf == nil {
		return nil
	}

	reserve, err := GetInfraGroupReserve()
	if err != nil {
		return err
	}

	level := syscache.GetLLC()
	cacheLevel := "L" + strconv.FormatUint(uint64(level), 10)
	ways := base.GetCosInfo().CbmMaskLen

	allres := proxyclient.GetResAssociation()
	infraGroup, ok := allres[groupName]
	if !ok {
		infraGroup = resctrl.NewResAssociation()
		l := len(reserve.Schemata)
		infraGroup.Schemata[cacheLevel] = make([]resctrl.CacheCos, l, l)
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
		infraGroup.Schemata[cacheLevel][id] = cc
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

	if err := proxyclient.Commit(infraGroup, groupName); err != nil {
		return err
	}

	return nil
}

package workload

// workload api objects to represent resources in RMD

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"

	libcache "github.com/intel/rmd/lib/cache"
	"github.com/intel/rmd/lib/cpu"
	"github.com/intel/rmd/lib/proc"
	"github.com/intel/rmd/lib/proxyclient"
	"github.com/intel/rmd/lib/resctrl"
	libutil "github.com/intel/rmd/lib/util"

	rmderror "github.com/intel/rmd/api/error"
	"github.com/intel/rmd/db"
	"github.com/intel/rmd/model/cache"
	"github.com/intel/rmd/model/policy"
	tw "github.com/intel/rmd/model/types/workload"
	"github.com/intel/rmd/util"
	"github.com/intel/rmd/util/rdtpool"
	rmdbase "github.com/intel/rmd/util/rdtpool/base"
)

var l sync.Mutex

// Validate the request workload object is validated.
func Validate(w *tw.RDTWorkLoad) error {
	if len(w.TaskIDs) <= 0 && len(w.CoreIDs) <= 0 {
		return fmt.Errorf("No task or core id specified")
	}

	// Firstly verify the task.
	ps := proc.ListProcesses()
	for _, task := range w.TaskIDs {
		if _, ok := ps[task]; !ok {
			return fmt.Errorf("The task: %s does not exist", task)
		}
	}

	if w.Policy == "" {
		if w.MaxCache == nil || w.MinCache == nil {
			return fmt.Errorf("Need to provide max_cache and min_cache if no policy specified")
		}
	}

	return nil
}

// Enforce a user request workload based on defined policy
func Enforce(w *tw.RDTWorkLoad) *rmderror.AppError {
	w.Status = tw.Failed

	l.Lock()
	defer l.Unlock()
	resaall := proxyclient.GetResAssociation()

	er := &tw.EnforceRequest{}
	if err := populateEnforceRequest(er, w); err != nil {
		return err
	}

	targetLev := strconv.FormatUint(uint64(libcache.GetLLC()), 10)
	av, err := rdtpool.GetAvailableCacheSchemata(resaall, []string{"infra", "."}, er.Type, "L"+targetLev)
	if err != nil {
		return rmderror.NewAppError(http.StatusInternalServerError,
			"Error to get available cache", err)
	}

	reserved := rdtpool.GetReservedInfo()
	changedRes := make(map[string]*resctrl.ResAssociation, 0)
	candidate := make(map[string]*libutil.Bitmap, 0)

	for k, v := range av {
		cacheID, _ := strconv.Atoi(k)
		if !inCacheList(uint32(cacheID), er.CacheIDs) && er.Type != rdtpool.Shared {
			candidate[k], _ = libutil.NewBitmap(
				rmdbase.GetCosInfo().CbmMaskLen,
				rmdbase.GetCosInfo().CbmMask)
			continue
		}
		switch er.Type {
		case rdtpool.Guarantee:
			// TODO
			// candidate[k] = v.GetBestMatchConnectiveBits(er.MaxWays, 0, true)
			candidate[k] = v.GetConnectiveBits(er.MaxWays, 0, false)
		case rdtpool.Besteffort:
			// Always to try to allocate max cache ways, if fail try to
			// get the most available ones

			freeBitmaps := v.ToBinStrings()
			var maxWays uint32
			maxWays = 0
			for _, val := range freeBitmaps {
				if val[0] == '1' {
					valLen := len(val)
					if (valLen/int(er.MinWays) > 0) && maxWays < uint32(valLen) {
						maxWays = uint32(valLen)
					}
				}
			}
			if maxWays <= 0 {
				if !reserved[rdtpool.Besteffort].Shrink {
					return rmderror.AppErrorf(http.StatusBadRequest,
						"Not enough cache left on cache_id %s", k)
				}
				// Try to Shrink workload in besteffort pool
				cand, changed, err := shrinkBEPool(resaall, reserved[rdtpool.Besteffort].Schemata[k], cacheID, er.MinWays)
				if err != nil {
					return rmderror.AppErrorf(http.StatusInternalServerError,
						"Errors while try to shrink cache ways on cache_id %s", k)
				}
				log.Printf("Shriking cache ways in besteffort pool, candidate schemata for cache id  %d is %s", cacheID, cand.ToString())
				candidate[k] = cand
				// Merge changed association to a map, we will commit this map
				// later
				for k, v := range changed {
					if _, ok := changedRes[k]; !ok {
						changedRes[k] = v
					}
				}
			} else {
				if maxWays > er.MaxWays {
					maxWays = er.MaxWays
				}
				candidate[k] = v.GetConnectiveBits(maxWays, 0, false)
			}

		case rdtpool.Shared:
			candidate[k] = reserved[rdtpool.Shared].Schemata[k]
		}

		if candidate[k].IsEmpty() {
			return rmderror.AppErrorf(http.StatusBadRequest,
				"Not enough cache left on cache_id %s", k)
		}
	}

	var resAss *resctrl.ResAssociation
	var grpName string

	if er.Type == rdtpool.Shared {
		grpName = reserved[rdtpool.Shared].Name
		if res, ok := resaall[grpName]; !ok {
			resAss = newResAss(candidate, targetLev)
		} else {
			resAss = res
		}
	} else {
		resAss = newResAss(candidate, targetLev)
		if len(w.TaskIDs) > 0 {
			grpName = w.TaskIDs[0] + "-" + er.Type
		} else if len(w.CoreIDs) > 0 {
			grpName = w.CoreIDs[0] + "-" + er.Type
		}
	}

	if len(w.CoreIDs) >= 0 {
		bm, _ := rmdbase.CPUBitmaps(w.CoreIDs)
		oldbm, _ := rmdbase.CPUBitmaps(resAss.CPUs)
		bm = bm.Or(oldbm)
		resAss.CPUs = bm.ToString()
	} else {
		if len(resAss.CPUs) == 0 {
			resAss.CPUs = ""
		}
	}
	resAss.Tasks = append(resAss.Tasks, w.TaskIDs...)

	if err = proxyclient.Commit(resAss, grpName); err != nil {
		log.Errorf("Error while try to commit resource group for workload %s, group name %s", w.ID, grpName)
		return rmderror.NewAppError(http.StatusInternalServerError,
			"Error to commit resource group for workload.", err)
	}

	// loop to change shrinked resource
	// TODO: there's corners if there are multiple changed resource groups,
	// but we failed to commit one of them (worest case is the last group),
	// there's no rollback.
	// possible fix is to adding this into a task flow
	for name, res := range changedRes {
		log.Debugf("Shink %s group", name)
		if err = proxyclient.Commit(res, name); err != nil {
			log.Errorf("Error while try to commit shrinked resource group, name: %s", name)
			proxyclient.DestroyResAssociation(grpName)
			return rmderror.NewAppError(http.StatusInternalServerError,
				"Error to shrink resource group", err)
		}
	}

	// reset os group
	if err = rdtpool.SetOSGroup(); err != nil {
		log.Errorf("Error while try to commit resource group for default group")
		proxyclient.DestroyResAssociation(grpName)
		return rmderror.NewAppError(http.StatusInternalServerError,
			"Error while try to commit resource group for default group.", err)
	}

	w.CosName = grpName
	w.Status = tw.Successful
	return nil
}

// Release Cos of the workload
func Release(w *tw.RDTWorkLoad) error {
	l.Lock()
	defer l.Unlock()

	resaall := proxyclient.GetResAssociation()

	r, ok := resaall[w.CosName]

	if !ok {
		log.Warningf("Could not find COS %s.", w.CosName)
		return nil
	}

	r.Tasks = util.SubtractStringSlice(r.Tasks, w.TaskIDs)
	cpubm, _ := rmdbase.CPUBitmaps(r.CPUs)

	if len(w.CoreIDs) > 0 {
		wcpubm, _ := rmdbase.CPUBitmaps(w.CoreIDs)
		cpubm = cpubm.Axor(wcpubm)
	}

	// safely remove resource group if no tasks and cpu bit map is empty
	if len(r.Tasks) < 1 && cpubm.IsEmpty() {
		log.Printf("Remove resource group: %s", w.CosName)
		if err := proxyclient.DestroyResAssociation(w.CosName); err != nil {
			return err
		}
		if err := rdtpool.SetOSGroup(); err != nil {
			return err
		}
		return nil
	}
	// remove workload task ids from resource group
	if len(w.TaskIDs) > 0 {
		if err := proxyclient.RemoveTasks(w.TaskIDs); err != nil {
			log.Printf("Ignore Error while remove tasks %s", err)
			return nil
		}
	}
	if len(w.CoreIDs) > 0 {
		r.CPUs = cpubm.ToString()
		return proxyclient.Commit(r, w.CosName)
	}
	return nil
}

// Update a workload
func Update(w, patched *tw.RDTWorkLoad) *rmderror.AppError {

	// if we change policy/max_cache/min_cache, release current resource group
	// and re-enforce it.
	reEnforce := false
	if patched.MaxCache != nil {
		if w.MaxCache == nil {
			w.MaxCache = patched.MaxCache
			reEnforce = true
		}
		if w.MaxCache != nil && *w.MaxCache != *patched.MaxCache {
			*w.MaxCache = *patched.MaxCache
			reEnforce = true
		}
	}

	if patched.MinCache != nil {
		if w.MinCache == nil {
			w.MinCache = patched.MinCache
			reEnforce = true
		}
		if w.MinCache != nil && *w.MinCache != *patched.MinCache {
			*w.MinCache = *patched.MinCache
			reEnforce = true
		}
	}

	if patched.Policy != w.Policy {
		w.Policy = patched.Policy
		reEnforce = true
	}

	if reEnforce == true {
		if err := Release(w); err != nil {
			return rmderror.NewAppError(http.StatusInternalServerError, "Faild to release workload",
				fmt.Errorf(""))
		}

		if len(patched.TaskIDs) > 0 {
			w.TaskIDs = patched.TaskIDs
		}
		if len(patched.CoreIDs) > 0 {
			w.CoreIDs = patched.CoreIDs
		}
		return Enforce(w)
	}

	l.Lock()
	defer l.Unlock()
	resaall := proxyclient.GetResAssociation()

	if !reflect.DeepEqual(patched.CoreIDs, w.CoreIDs) ||
		!reflect.DeepEqual(patched.TaskIDs, w.TaskIDs) {
		err := Validate(patched)
		if err != nil {
			return rmderror.NewAppError(http.StatusBadRequest, "Failed to validate workload", err)
		}

		targetResAss, ok := resaall[w.CosName]
		if !ok {
			return rmderror.NewAppError(http.StatusInternalServerError, "Can not find resource group name",
				fmt.Errorf(""))
		}

		if len(patched.TaskIDs) > 0 {
			// FIXME  Is this a bug? Seems the targetResAss.Tasks is inconsistent with w.TaskIDs
			targetResAss.Tasks = append(targetResAss.Tasks, patched.TaskIDs...)
			w.TaskIDs = patched.TaskIDs
		}
		if len(patched.CoreIDs) > 0 {
			bm, err := rmdbase.CPUBitmaps(patched.CoreIDs)
			if err != nil {
				return rmderror.NewAppError(http.StatusBadRequest,
					"Failed to Pareser workload coreIDs.", err)
			}
			// TODO: check if this new CoreIDs overwrite other resource group
			targetResAss.CPUs = bm.ToString()
			w.CoreIDs = patched.CoreIDs
		}
		// commit changes
		if err = proxyclient.Commit(targetResAss, w.CosName); err != nil {
			log.Errorf("Error while try to commit resource group for workload %s, group name %s", w.ID, w.CosName)
			return rmderror.NewAppError(http.StatusInternalServerError,
				"Error to commit resource group for workload.", err)
		}
	}
	return nil
}

func getCacheIDs(taskids []string, cpubitmap string, cacheinfos *cache.Infos, cpunum int) []uint32 {
	var CacheIDs []uint32
	cpubm, _ := libutil.NewBitmap(cpunum, cpubitmap)

	for _, t := range taskids {
		af, err := proc.GetCPUAffinity(t)
		if err != nil {
			log.Warningf("Failed to get cpu affinity for task %s", t)
			// FIXME get default affinity instead of a hard code 400 cpus
			af, _ = libutil.NewBitmap(cpunum, strings.Repeat("f", 100))
		}
		cpubm = cpubm.Or(af)
	}

	// No warry, cpubitmap is empty if taskids is None
	for _, c := range cacheinfos.Caches {
		// Okay, NewBitmap only support string list if we using human style
		bm, _ := libutil.NewBitmap(cpunum, strings.Split(c.ShareCPUList, "\n"))
		if !cpubm.And(bm).IsEmpty() {
			CacheIDs = append(CacheIDs, c.ID)
		}
	}
	return CacheIDs
}

func inCacheList(cache uint32, cacheList []uint32) bool {
	// TODO: if this case, workload has taskids.
	// Later we need to have abilitity to discover if has taskset
	// to pin this taskids on a cpuset or not, for now we allocate
	// cache on all cache.
	// FIXME: this shouldn't happen here actually
	if len(cacheList) == 0 {
		return true
	}

	for _, c := range cacheList {
		if cache == c {
			return true
		}
	}
	return false
}

func populateEnforceRequest(req *tw.EnforceRequest, w *tw.RDTWorkLoad) *rmderror.AppError {

	w.Status = tw.None
	cpubitstr := ""
	if len(w.CoreIDs) >= 0 {
		bm, err := rmdbase.CPUBitmaps(w.CoreIDs)
		if err != nil {
			return rmderror.NewAppError(http.StatusBadRequest,
				"Failed to Parese workload coreIDs.", err)
		}
		cpubitstr = bm.ToString()
	}

	cacheinfo := &cache.Infos{}
	cacheinfo.GetByLevel(libcache.GetLLC())

	cpunum := cpu.HostCPUNum()
	if cpunum == 0 {
		return rmderror.AppErrorf(http.StatusInternalServerError,
			"Unable to get Total CPU numbers on Host")
	}

	req.CacheIDs = getCacheIDs(w.TaskIDs, cpubitstr, cacheinfo, cpunum)

	populatePolicy := true

	if w.MinCache != nil {
		req.MinWays = *w.MinCache
	}
	if w.MaxCache != nil {
		req.MaxWays = *w.MaxCache
		populatePolicy = false
	}
	// else get max/min from policy
	if populatePolicy {
		p, err := policy.GetDefaultPolicy(w.Policy)
		if err != nil {
			return rmderror.NewAppError(http.StatusInternalServerError,
				"Could not find the Polciy.", err)
		}

		maxWays, err := strconv.Atoi(p["MaxCache"])
		if err != nil {
			return rmderror.NewAppError(http.StatusInternalServerError,
				"Error define MaxCache in Polciy.", err)
		}
		req.MaxWays = uint32(maxWays)

		minWays, err := strconv.Atoi(p["MinCache"])
		if err != nil {
			return rmderror.NewAppError(http.StatusInternalServerError,
				"Error define MinCache in Polciy.", err)
		}
		req.MinWays = uint32(minWays)
	}

	var err error
	req.Type, err = rdtpool.GetCachePoolName(req.MaxWays, req.MinWays)
	if err != nil {
		return rmderror.NewAppError(http.StatusBadRequest,
			"Bad cache ways request",
			err)
	}
	return nil
}

func newResAss(r map[string]*libutil.Bitmap, level string) *resctrl.ResAssociation {
	newResAss := resctrl.ResAssociation{}
	newResAss.Schemata = make(map[string][]resctrl.CacheCos)

	targetLev := "L" + level

	for k, v := range r {
		cacheID, _ := strconv.Atoi(k)
		newcos := resctrl.CacheCos{ID: uint8(cacheID), Mask: v.ToString()}
		newResAss.Schemata[targetLev] = append(newResAss.Schemata[targetLev], newcos)

		log.Debugf("Newly created Mask for Cache %s is %s", k, newcos.Mask)
	}
	return &newResAss
}

// shrinkBEPool requres to provide cacheid of the request, MinCache ways (
// because we lack cache now if we need to shrink), of cause resassociations
// besteffort pool reserved cache way bitmap.
// returns: bitmap we allocated for the new request
// returns: a map[string]*resctrl.ResAssociation as we changed other workloads'
// cache ways, need to reflect them into resctrl fs.
// returns: error if internal error happens.
func shrinkBEPool(resaall map[string]*resctrl.ResAssociation,
	reservedSchemata *libutil.Bitmap,
	cacheID int,
	reqways uint32) (*libutil.Bitmap, map[string]*resctrl.ResAssociation, error) {

	besteffortRes := make(map[string]*resctrl.ResAssociation)
	dbc, _ := db.NewDB()
	// do a copy
	availableSchemata := &(*reservedSchemata)
	targetLev := strconv.FormatUint(uint64(libcache.GetLLC()), 10)
	for name, v := range resaall {
		if strings.HasSuffix(name, "-"+rdtpool.Besteffort) {
			besteffortRes[name] = v
			ws, _ := dbc.QueryWorkload(map[string]interface{}{
				"CosName": name})
			if len(ws) == 0 {
				return nil, besteffortRes, fmt.Errorf(
					"Internal error, can not find exsting workload for resource group name %s", name)
			}
			cosSchemata, _ := rmdbase.CacheBitmaps(v.Schemata["L"+targetLev][cacheID].Mask)
			// TODO: need find a better way to reduce the cache way fragments
			// as currently we are using map to keep resctrl group, it's non-order
			// so it's little hard to get which resctrl group next to which.
			// just using max - min slot to shrink the cache. Hence, the result
			// would only shrink one of the resource group to min one
			minSchemata := cosSchemata.GetConnectiveBits(*ws[0].MinCache, 0, false)
			availableSchemata = availableSchemata.Axor(minSchemata)
		}
	}
	// I would like to allocate cache from low to high, this will help to
	// reduce cos
	candidateSchemata := availableSchemata.GetConnectiveBits(reqways, 0, true)

	// loop besteffortRes to find which assocation need to be changed.
	changedRes := make(map[string]*resctrl.ResAssociation)
	for name, v := range besteffortRes {
		cosSchemata, _ := rmdbase.CacheBitmaps(v.Schemata["L"+targetLev][cacheID].Mask)
		tmpSchemataStr := cosSchemata.Axor(candidateSchemata).ToString()
		if tmpSchemataStr != cosSchemata.ToString() {
			// Changing pointers, the change will be reflact to the origin one
			v.Schemata["L"+targetLev][cacheID].Mask = tmpSchemataStr
			changedRes[name] = v
		}
	}

	return candidateSchemata, changedRes, nil
}

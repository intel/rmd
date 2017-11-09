package hospitality

// This model is just for cache info
// We can ref k8s

import (
	"net/http"
	"strconv"

	rmderror "github.com/intel/rmd/api/error"
	"github.com/intel/rmd/db"
	libcache "github.com/intel/rmd/lib/cache"
	"github.com/intel/rmd/lib/proxyclient"
	"github.com/intel/rmd/model/policy"
	"github.com/intel/rmd/util/rdtpool"
	log "github.com/sirupsen/logrus"
)

// Request represents the hospitality request
type Request struct {
	MaxCache uint32  `json:"max_cache,omitempty"`
	MinCache uint32  `json:"min_cache,omitempty"`
	Policy   string  `json:"policy,omitempty"`
	CacheID  *uint32 `json:"cache_id,omitempty"`
}

// CacheScore represents the score on specific cache id
type CacheScore map[string]uint32

// Hospitality represents the score of the host
/*
{
	"score": {
		"l3": {
			"0": 30
			"1": 30
		}
	}
}
*/
type Hospitality struct {
	SC map[string]CacheScore `json:"score"`
}

// GetByRequest returns hospitality score by request
func (h *Hospitality) GetByRequest(req *Request) error {
	level := libcache.GetLLC()
	targetLev := strconv.FormatUint(uint64(level), 10)
	cacheLevel := "l" + targetLev
	cacheS := make(map[string]uint32)
	h.SC = map[string]CacheScore{cacheLevel: cacheS}

	max := req.MaxCache
	min := req.MinCache

	if req.Policy != "" {
		tier, err := policy.GetDefaultPolicy(req.Policy)
		if err != nil {
			return rmderror.NewAppError(http.StatusInternalServerError,
				"Can not find Policy", err)
		}
		m, _ := strconv.Atoi(tier["MaxCache"])
		n, _ := strconv.Atoi(tier["MinCache"])
		max = uint32(m)
		min = uint32(n)
	}
	return h.GetByRequestMaxMin(max, min, req.CacheID, targetLev)
}

// GetByRequestMaxMin constructs Hospitality struct by max and min cache ways
func (h *Hospitality) GetByRequestMaxMin(max, min uint32, cacheIDuint *uint32, targetLev string) error {

	var reqType string

	if max == 0 && min == 0 {
		reqType = rdtpool.Shared
	} else if max > min && min != 0 {
		reqType = rdtpool.Besteffort
	} else if max == min {
		reqType = rdtpool.Guarantee
	} else {
		return rmderror.AppErrorf(http.StatusBadRequest,
			"Bad request, max_cache=%d, min_cache=%d", max, min)
	}

	resaall := proxyclient.GetResAssociation()

	av, _ := rdtpool.GetAvailableCacheSchemata(resaall, []string{"infra", "."}, reqType, "L"+targetLev)

	cacheS := make(map[string]uint32)
	h.SC = map[string]CacheScore{"l" + targetLev: cacheS}

	reserved := rdtpool.GetReservedInfo()

	if reqType == rdtpool.Shared {
		dbc, _ := db.NewDB()
		ws, _ := dbc.QueryWorkload(map[string]interface{}{
			"CosName": reserved[rdtpool.Shared].Name,
			"Status":  "Successful"})
		totalCount := reserved[rdtpool.Shared].Quota
		for k := range av {
			if uint(len(ws)) < totalCount {
				cacheS[k] = 100
			} else {
				cacheS[k] = 0
			}
			retrimCache(k, cacheIDuint, &cacheS)
		}
		return nil
	}

	for k, v := range av {
		var fbs []string
		cacheS[k] = 0

		fbs = v.ToBinStrings()

		log.Debugf("Free bitmask on cache [%s] is [%s]", k, fbs)
		// Calculate total supported
		for _, val := range fbs {
			if val[0] == '1' {
				valLen := len(val)
				if (valLen/int(min) > 0) && cacheS[k] < uint32(valLen) {
					cacheS[k] = uint32(valLen)
				}
			}
		}
		if cacheS[k] > 0 {
			// (NOTES): Gurantee will return 0|100
			// Besteffort will return (max continious ways) / max
			cacheS[k] = (cacheS[k] * 100) / max
			if cacheS[k] > 100 {
				cacheS[k] = 100
			}
		} else {
			cacheS[k] = 0
		}

		retrimCache(k, cacheIDuint, &cacheS)
	}
	return nil
}

func retrimCache(cacheID string, cacheIDuint *uint32, cacheS *map[string]uint32) {

	icacheID, _ := strconv.Atoi(cacheID)
	if cacheIDuint != nil {
		// We only care about specific cache_id
		if *cacheIDuint != uint32(icacheID) {
			delete(*cacheS, cacheID)
		}
	}
}

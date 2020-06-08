package hospitality

// This model is just for cache info
// We can ref k8s

import (
	"net/http"
	"strconv"

	"strings"

	"github.com/intel/rmd/internal/db"
	rmderror "github.com/intel/rmd/internal/error"
	proxyclient "github.com/intel/rmd/internal/proxy/client"
	"github.com/intel/rmd/modules/cache"
	"github.com/intel/rmd/modules/policy"
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
	level := cache.GetLLC()
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
		//get max cache
		mAsInterface, ok := tier["cache"]["max"]
		if !ok {
			return rmderror.NewAppError(http.StatusInternalServerError,
				"Error to get max cache", err)
		}

		m, ok := mAsInterface.(int)
		if !ok {
			return rmderror.NewAppError(http.StatusInternalServerError,
				"Error to get max cache", err)
		}

		//get min cache
		nAsInterface, ok := tier["cache"]["min"]
		if !ok {
			return rmderror.NewAppError(http.StatusInternalServerError,
				"Error to get min cache", err)
		}

		n, ok := nAsInterface.(int)
		if !ok {
			return rmderror.NewAppError(http.StatusInternalServerError,
				"Error to get min cache", err)
		}

		max = uint32(m)
		min = uint32(n)
	}
	return h.GetByRequestMaxMin(max, min, req.CacheID, targetLev)
}

// GetByRequestMaxMin constructs Hospitality struct by max and min cache ways
func (h *Hospitality) GetByRequestMaxMin(max, min uint32, cacheIDuint *uint32, targetLev string) error {

	var reqType string

	if max == 0 && min == 0 {
		reqType = cache.Shared
	} else if max > min && min != 0 {
		reqType = cache.Besteffort
	} else if max == min {
		reqType = cache.Guarantee
	} else {
		return rmderror.AppErrorf(http.StatusBadRequest,
			"Bad request, max_cache=%d, min_cache=%d", max, min)
	}

	resaall := proxyclient.GetResAssociation()

	av, err := cache.GetAvailableCacheSchemata(resaall, []string{"infra", "."}, reqType, "L"+targetLev)
	if err != nil && !strings.Contains(err.Error(), "error doesn't support pool") {
		return rmderror.AppErrorf(http.StatusInternalServerError,
			"Unable to read cache schemata; %s", err.Error())
	}

	cacheS := make(map[string]uint32)
	h.SC = map[string]CacheScore{"l" + targetLev: cacheS}

	reserved := cache.GetReservedInfo()

	if reqType == cache.Shared {
		dbc, err := db.NewDB()
		if err != nil {
			return err
		}
		ws, err := dbc.QueryWorkload(map[string]interface{}{
			"CosName": reserved[cache.Shared].Name,
			"Status":  "Successful"})
		if err != nil {
			return err
		}
		totalCount := reserved[cache.Shared].Quota
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
			// (NOTES): Guarantee will return 0|100
			// Besteffort will return (max continuous ways) / max
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

	//in this case we don't need to handle error because we are deleting only during "specific situation"
	//error is not that "specific situation" so it will not generate impact on code logic
	icacheID, _ := strconv.Atoi(cacheID)
	if cacheIDuint != nil {
		// We only care about specific cache_id
		if *cacheIDuint != uint32(icacheID) {
			delete(*cacheS, cacheID)
		}
	}
}

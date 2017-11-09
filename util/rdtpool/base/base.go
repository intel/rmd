package base

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/intel/rmd/lib/cache"
	"github.com/intel/rmd/lib/cpu"
	"github.com/intel/rmd/lib/proxyclient"
	"github.com/intel/rmd/lib/util"
)

// CosInfo is class of service infor
// FIXME should find a good accommodation for file
type CosInfo struct {
	CbmMaskLen int
	MinCbmBits int
	NumClosids int
	CbmMask    string
}

var catCosInfo = &CosInfo{0, 0, 0, ""}
var infoOnce sync.Once

// Reserved schemata inforamtion
type Reserved struct {
	AllCPUs     *util.Bitmap            //cpu bit masp
	SchemaNum   int                     // Numbers of schema
	Name        string                  // Resource group name if it is a resource group instead of pool
	Schemata    map[string]*util.Bitmap // Schema list
	CPUsPerNode map[string]*util.Bitmap // CPU bitmap
	Quota       uint                    // Max allowed usage for this resource
	Shrink      bool                    // Wether shrink in BE pool
}

// GetCosInfo is Concurrency-safe.
func GetCosInfo() CosInfo {
	infoOnce.Do(func() {
		rcinfo := proxyclient.GetRdtCosInfo()
		level := syscache.GetLLC()
		targetLev := strconv.FormatUint(uint64(level), 10)
		cacheLevel := "l" + targetLev

		catCosInfo.CbmMaskLen = util.CbmLen(rcinfo[cacheLevel].CbmMask)
		catCosInfo.MinCbmBits = rcinfo[cacheLevel].MinCbmBits
		catCosInfo.NumClosids = rcinfo[cacheLevel].NumClosids
		catCosInfo.CbmMask = rcinfo[cacheLevel].CbmMask
	})
	return *catCosInfo
}

// CPUBitmaps is a wrapper for Bitmap
func CPUBitmaps(cpuids interface{}) (*util.Bitmap, error) {
	// FIXME need a wrap for CPU bitmap.
	cpunum := cpu.HostCPUNum()
	if cpunum == 0 {
		// return nil or an empty Bitmap?
		var bm *util.Bitmap
		return bm, fmt.Errorf("Unable to get Total CPU numbers on Host")
	}
	return util.NewBitmap(cpunum, cpuids)
}

// CacheBitmaps is a wrapper for Cache bitmap
func CacheBitmaps(bitmask interface{}) (*util.Bitmap, error) {
	// FIXME need a wrap for CPU bitmap.
	len := GetCosInfo().CbmMaskLen
	if len == 0 {
		// return nil or an empty Bitmap?
		var bm *util.Bitmap
		return bm, fmt.Errorf("Unable to get Total cache ways on Host")
	}
	return util.NewBitmap(len, bitmask)
}

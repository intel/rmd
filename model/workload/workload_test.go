package workload

import (
	"testing"

	. "github.com/prashantv/gostub"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/intel/rmd/lib/proc"
	"github.com/intel/rmd/model/cache"

	tw "github.com/intel/rmd/model/types/workload"
)

func TestGetCacheIDs(t *testing.T) {
	cacheinfos := &cache.Infos{Num: 2,
		Caches: map[uint32]cache.Info{
			0: cache.Info{ID: 0, ShareCPUList: "0-3"},
			1: cache.Info{ID: 1, ShareCPUList: "4-7"},
		}}

	cpubitmap := "3"

	cache_ids := getCacheIDs([]string{}, cpubitmap, cacheinfos, 8)
	if len(cache_ids) != 1 && cache_ids[0] != 0 {
		t.Errorf("cache_ids should be [0], but we get %v", cache_ids)
	}

	cpubitmap = "1f"
	cache_ids = getCacheIDs([]string{}, cpubitmap, cacheinfos, 8)
	if len(cache_ids) != 2 {
		t.Errorf("cache_ids should be [0, 1], but we get %v", cache_ids)
	}

	cpubitmap = "10"
	cache_ids = getCacheIDs([]string{}, cpubitmap, cacheinfos, 8)
	if len(cache_ids) != 1 && cache_ids[0] != 1 {
		t.Errorf("cache_ids should be [1], but we get %v", cache_ids)
	}

	cpubitmap = "f00"
	cache_ids = getCacheIDs([]string{}, cpubitmap, cacheinfos, 8)
	if len(cache_ids) != 0 {
		t.Errorf("cache_ids should be [], but we get %v", cache_ids)
	}

}

func TestValidateWorkLoad(t *testing.T) {
	Convey("Test Validate workload", t, func() {
		Convey("Validate with empty workload", func() {
			subs := StubFunc(&proc.ListProcesses, map[string]proc.Process{"1": proc.Process{1, "cmdline"}})
			defer subs.Reset()
			var cache uint32 = 1
			wl := &tw.RDTWorkLoad{}
			err := Validate(wl)
			So(err, ShouldNotBeNil)

			wl.TaskIDs = []string{"1"}
			Convey("Validate with task ids", func() {
				err := Validate(wl)
				So(err, ShouldNotBeNil)

				wl.Policy = "gold"
				Convey("Validate with task ids and Policy", func() {
					err := Validate(wl)
					So(err, ShouldBeNil)
				})
			})

			wl.MaxCache = &cache
			Convey("Validate with MaxCache is not nil but MinCache is nil", func() {
				err := Validate(wl)
				So(err, ShouldNotBeNil)

				wl.MinCache = &cache
				Convey("Validate with MaxCache & MinCache are not nil", func() {

					err := Validate(wl)
					So(err, ShouldBeNil)
				})
			})

			wl.TaskIDs = []string{"2"}
			Convey("Validate with task ids does not existed", func() {
				err := Validate(wl)
				So(err, ShouldNotBeNil)
			})
		})
	})
}

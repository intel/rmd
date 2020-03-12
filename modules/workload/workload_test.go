package workload

import (
	"reflect"
	"testing"

	"github.com/intel/rmd/modules/cache"
	tw "github.com/intel/rmd/modules/workload/types"
	"github.com/intel/rmd/utils/proc"
	. "github.com/prashantv/gostub"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func TestGetCacheIDs(t *testing.T) {
	cacheinfos := &cache.Infos{Num: 2,
		Caches: map[uint32]cache.Info{
			0: cache.Info{ID: 0, ShareCPUList: "0-3"},
			1: cache.Info{ID: 1, ShareCPUList: "4-7"},
		}}

	cpubitmap := "3"

	cacheIDs := getCacheIDs([]string{}, cpubitmap, cacheinfos, 8)
	if len(cacheIDs) != 1 && cacheIDs[0] != 0 {
		t.Errorf("cache_ids should be [0], but we get %v", cacheIDs)
	}

	cpubitmap = "1f"
	cacheIDs = getCacheIDs([]string{}, cpubitmap, cacheinfos, 8)
	if len(cacheIDs) != 2 {
		t.Errorf("cache_ids should be [0, 1], but we get %v", cacheIDs)
	}

	cpubitmap = "10"
	cacheIDs = getCacheIDs([]string{}, cpubitmap, cacheinfos, 8)
	if len(cacheIDs) != 1 && cacheIDs[0] != 1 {
		t.Errorf("cache_ids should be [1], but we get %v", cacheIDs)
	}

	cpubitmap = "f00"
	cacheIDs = getCacheIDs([]string{}, cpubitmap, cacheinfos, 8)
	if len(cacheIDs) != 0 {
		t.Errorf("cache_ids should be [], but we get %v", cacheIDs)
	}

}

func TestValidateWorkLoad(t *testing.T) {

	pflag.String("address", "", "Listen address")
	pflag.Int("tlsport", 0, "TLS listen port")
	pflag.BoolP("debug", "d", false, "Enable debug")
	pflag.String("unixsock", "", "Unix sock file path")
	pflag.Int("debugport", 0, "Debug listen port")
	pflag.String("conf-dir", "", "Directly of config file")
	pflag.String("clientauth", "challenge", "The policy the server will follow for TLS Client Authentication")
	// set database details
	viper.Set("database.backend", "bolt")
	viper.Set("database.dbname", "rmd")
	viper.Set("database.transport", "/tmp/rmd_test.db")

	pflag.Parse()
	//prepare DB for test
	err := Init()
	if err != nil {
		t.Log("Database initialization failure!")
		t.FailNow()
	}

	Convey("Test Validate workload", t, func(c C) {
		c.Convey("Validate with empty workload", func(c C) {
			subs := StubFunc(&proc.ListProcesses, map[string]proc.Process{"1": proc.Process{Pid: 1, CmdLine: "cmdline"}})
			defer subs.Reset()
			var cache uint32 = 1
			wl := &tw.RDTWorkLoad{}
			err := Validate(wl)
			c.So(err, ShouldNotBeNil)

			wl.TaskIDs = []string{"1"}
			c.Convey("Validate with task ids", func(c C) {
				err := Validate(wl)
				c.So(err, ShouldNotBeNil)

				wl.Policy = "gold"
				c.Convey("Validate with task ids and Policy", func(c C) {
					err := Validate(wl)
					c.So(err, ShouldBeNil)
				})
			})

			// Test for Workload with Cache params provided
			wl.Cache.Max = &cache
			c.Convey("Validate with MaxCache is not nil but MinCache is nil", func(c C) {
				err := Validate(wl)
				c.So(err, ShouldNotBeNil)

				wl.Cache.Min = &cache
				c.Convey("Validate with MaxCache & MinCache are not nil", func(c C) {

					err := Validate(wl)
					c.So(err, ShouldBeNil)
				})
			})

			wl.TaskIDs = []string{"2"}
			c.Convey("Validate with task ids does not existed", func(c C) {
				err := Validate(wl)
				c.So(err, ShouldNotBeNil)
			})
		})
	})
}

func Test_prepareCoreIDs(t *testing.T) {
	type args struct {
		w []string
	}
	tests := []struct {
		name    string
		args    args
		want    []int
		wantErr bool
	}{
		{"4 case", args{[]string{"4"}}, []int{4}, false},
		{"4 5 case", args{[]string{"4", "5"}}, []int{4, 5}, false},
		{"4-9 case", args{[]string{"4-9"}}, []int{4, 5, 6, 7, 8, 9}, false},
		{"9-4 case", args{[]string{"9-4"}}, []int{}, true},
		{"4 5 7-9 case", args{[]string{"4", "5", "7-9"}}, []int{4, 5, 7, 8, 9}, false},
		{"\n4-9 case", args{[]string{"\n4-9"}}, []int{}, true},
		{"4-\n9 case", args{[]string{"4-\n9"}}, []int{}, true},
		{"\n4 case", args{[]string{"\n4"}}, []int{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := prepareCoreIDs(tt.args.w)
			if (err != nil) != tt.wantErr {
				t.Errorf("prepareCoreIDs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("prepareCoreIDs() = %v, want %v", got, tt.want)
			}
		})
	}
}

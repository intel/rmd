package workload

import (
	"os"
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

func init() {
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
}

func TestGetCacheIDs(t *testing.T) {
	cacheinfos := &cache.Infos{Num: 2,
		Caches: map[uint32]cache.Info{
			0: cache.Info{ID: 0, ShareCPUList: "0-3"},
			1: cache.Info{ID: 1, ShareCPUList: "4-7"},
		}}

	cpubitmap := "3"

	socketIDs := getSocketIDs([]string{}, cpubitmap, cacheinfos, 8)
	if len(socketIDs) != 1 && socketIDs[0] != 0 {
		t.Errorf("cache_ids should be [0], but we get %v", socketIDs)
	}

	cpubitmap = "1f"
	socketIDs = getSocketIDs([]string{}, cpubitmap, cacheinfos, 8)
	if len(socketIDs) != 2 {
		t.Errorf("cache_ids should be [0, 1], but we get %v", socketIDs)
	}

	cpubitmap = "10"
	socketIDs = getSocketIDs([]string{}, cpubitmap, cacheinfos, 8)
	if len(socketIDs) != 1 && socketIDs[0] != 1 {
		t.Errorf("cache_ids should be [1], but we get %v", socketIDs)
	}

	cpubitmap = "f00"
	socketIDs = getSocketIDs([]string{}, cpubitmap, cacheinfos, 8)
	if len(socketIDs) != 0 {
		t.Errorf("cache_ids should be [], but we get %v", socketIDs)
	}

}

func TestValidateWorkLoad(t *testing.T) {
	//prepare DB for test
	err := Init()
	if err != nil {
		t.Errorf("Cannot create database - tests results can be corrupted\n")
	}

	Convey("Test Validate workload", t, func(c C) {
		c.Convey("Validate with empty workload", func(c C) {
			subs := StubFunc(&proc.ListProcesses, map[string]proc.Process{"1": proc.Process{Pid: 1, CmdLine: "cmdline"}})
			defer subs.Reset()
			var cache uint32 = 1
			var mba uint32 = 50
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
			wl.Rdt.Cache.Max = &cache
			c.Convey("Validate with MaxCache is not nil but MinCache is nil", func(c C) {
				err := Validate(wl)
				c.So(err, ShouldNotBeNil)

				wl.Rdt.Cache.Min = &cache
				wl.Rdt.Mba.Percentage = &mba
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
		{"\n4-9 case", args{[]string{"\n4-9"}}, []int{4, 5, 6, 7, 8, 9}, false},
		{"4-\n9 case", args{[]string{"4-\n9"}}, []int{4, 5, 6, 7, 8, 9}, false},
		{"\n4 case", args{[]string{"\n4"}}, []int{4}, false},
		{"abc case", args{[]string{"abc"}}, []int{}, true},
		{"abc-\n9 case", args{[]string{"abc-\n9"}}, []int{}, true},
		{"\n4-abc case", args{[]string{"\n4-abc"}}, []int{}, true},
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

func TestInit(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"correct case", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Init(); (err != nil) != tt.wantErr {
				t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_fillWorkloadByPolicy(t *testing.T) {

	// the same value for min and max cache
	var origCache uint32 = 2
	origpStateRatio := 3.0
	pStateMonitoring := "on"

	correctWorkload := tw.RDTWorkLoad{}
	correctWorkload.CoreIDs = []string{"3"}
	correctWorkload.Origin = "REST"
	correctWorkload.Status = "Successful"
	correctWorkload.CosName = "3-guarantee"
	correctWorkload.Policy = "gold"
	correctWorkload.Rdt.Cache.Max = &origCache
	correctWorkload.Rdt.Cache.Min = &origCache
	correctWorkload.PState.Ratio = &origpStateRatio
	correctWorkload.PState.Monitoring = &pStateMonitoring

	noPolicyWorkload := tw.RDTWorkLoad{}
	noPolicyWorkload.CoreIDs = []string{"3"}
	noPolicyWorkload.Origin = "REST"
	noPolicyWorkload.Status = "Successful"
	noPolicyWorkload.CosName = "3-guarantee"
	noPolicyWorkload.Rdt.Cache.Max = &origCache
	noPolicyWorkload.Rdt.Cache.Min = &origCache
	noPolicyWorkload.PState.Ratio = &origpStateRatio
	noPolicyWorkload.PState.Monitoring = &pStateMonitoring

	type args struct {
		wrkld *tw.RDTWorkLoad
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"Nil as workload", args{wrkld: nil}, true},
		{"Lack of specified policy case", args{wrkld: &noPolicyWorkload}, true},
		{"Correct case", args{wrkld: &correctWorkload}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := fillWorkloadByPolicy(tt.args.wrkld); (err != nil) != tt.wantErr {
				t.Errorf("fillWorkloadByPolicy() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_inCacheList(t *testing.T) {

	type args struct {
		cache     uint32
		cacheList []uint32
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"Correct case", args{cache: (uint32)(3), cacheList: []uint32{3}}, true},
		{"Empty list case", args{cache: (uint32)(3), cacheList: []uint32{}}, true},
		{"Cache not in list case", args{cache: (uint32)(3), cacheList: []uint32{5}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := inCacheList(tt.args.cache, tt.args.cacheList); got != tt.want {
				t.Errorf("inCacheList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetByUUID(t *testing.T) {
	type args struct {
		uuid string
	}
	tests := []struct {
		name          string
		args          args
		wantResult    tw.RDTWorkLoad
		wantErr       bool
		wantWorkingDB bool
	}{
		{"DB not initialized case", args{uuid: "sdfsdfs-sdfsdfsd-sdafsdf"}, tw.RDTWorkLoad{}, true, false},
		{"Not existing UUID case", args{uuid: "fdgdfg-ghjghjgjgh-sdafsdf"}, tw.RDTWorkLoad{}, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.wantWorkingDB == true {
				//prepare DB for test
				err := Init()
				if err != nil {
					t.Errorf("Cannot create database - tests results can be corrupted\n")
				}
			} else {
				// enforce "Service database not initialized"
				workloadDatabase = nil
			}
			gotResult, err := GetByUUID(tt.args.uuid)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetByUUID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("GetByUUID() = %v, want %v", gotResult, tt.wantResult)
			}
		})

	}

}

func TestGetAll(t *testing.T) {

	myTable := []tw.RDTWorkLoad{}

	tests := []struct {
		name          string
		want          []tw.RDTWorkLoad
		wantErr       bool
		wantWorkingDB bool
	}{
		{"DB not initialized case", myTable, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.wantWorkingDB == true {
				//prepare DB for test
				err := Init()
				if err != nil {
					t.Errorf("Cannot create database - tests results can be corrupted\n")
				}
			} else {
				// enforce "Service database not initialized"
				workloadDatabase = nil
			}

			got, err := GetAll()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAll() = %v, want %v", got, tt.want)
			}
		})

	}
}

func TestGetWorkloadByID(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name          string
		args          args
		wantResult    tw.RDTWorkLoad
		wantErr       bool
		wantWorkingDB bool
	}{
		{"DB not initialized case", args{id: "131313131313131313131"}, tw.RDTWorkLoad{}, true, false},
		{"Workload ID not exists in DB case", args{id: "12121212121212"}, tw.RDTWorkLoad{}, true, true},
		// case for positive case is made in TestCreate testcase where GetWorkloadByID is used to get already created workload
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantWorkingDB == true {
				//prepare DB for test
				err := Init()
				if err != nil {
					t.Errorf("Cannot create database - tests results can be corrupted\n")
				}
			} else {
				// enforce "Service database not initialized"
				workloadDatabase = nil
			}
			gotResult, err := GetWorkloadByID(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetWorkloadByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("GetWorkloadByID() = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

func TestCreate(t *testing.T) {

	workloadID := "5632"
	myWorkload := tw.RDTWorkLoad{ID: workloadID, Policy: "bronze"}

	type args struct {
		wl *tw.RDTWorkLoad
	}
	tests := []struct {
		name          string
		args          args
		wantErr       bool
		wantWorkingDB bool
	}{
		{"DB not initialized case", args{wl: &myWorkload}, true, false},
		{"Nil case", args{wl: nil}, true, true},
		{"Correct case", args{wl: &myWorkload}, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// remove db if exists to have a clean test
			pathToDB := "/tmp/rmd_test.db"
			if _, err := os.Stat(pathToDB); err == nil {
				err := os.Remove(pathToDB)
				if err != nil {
					t.Errorf("Failed to remove DB for clean test - other tests results can be corrupted")
				}
			}

			if tt.wantWorkingDB == true {
				//prepare DB for test
				err := Init()
				if err != nil {
					t.Errorf("Cannot create database - tests results can be corrupted\n")
				}
			} else {
				// enforce "Service database not initialized"
				workloadDatabase = nil
			}
			if err := Create(tt.args.wl); (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
			//check existence in db only for correct case (perform whole workload lifecycle)
			if tt.wantErr == false {
				// check if workload is in db
				_, err := GetWorkloadByID(workloadID)
				if err != nil {
					t.Errorf("TestCreate - Created workload was not found in db due to: %s", err)
				}
				// delete workload
				err = Delete(tt.args.wl)
				if err != nil {
					t.Errorf("TestCreate - Failed to delete created workload due to: %s", err)
				}

			}

		})
	}
}

func TestDelete(t *testing.T) {

	myWorkload := tw.RDTWorkLoad{ID: "13141516"}
	type args struct {
		wl *tw.RDTWorkLoad
	}
	tests := []struct {
		name          string
		args          args
		wantErr       bool
		wantWorkingDB bool
	}{
		{"DB not initialized case", args{wl: &myWorkload}, true, false},
		{"Nil case", args{wl: nil}, true, true},
		{"Not existing workload case", args{wl: &myWorkload}, false, true},
		// case for positive case is made in TestCreate testcase where Delete is used to remove already created workload
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantWorkingDB == true {
				//prepare DB for test
				err := Init()
				if err != nil {
					t.Errorf("Cannot create database - tests results can be corrupted\n")
				}
			} else {
				// enforce "Service database not initialized"
				workloadDatabase = nil
			}
			if err := Delete(tt.args.wl); (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_validateInDB(t *testing.T) {

	myWorkload := tw.RDTWorkLoad{ID: "151515151515151"}
	type args struct {
		wl *tw.RDTWorkLoad
	}
	tests := []struct {
		name          string
		args          args
		wantErr       bool
		wantWorkingDB bool
	}{
		{"DB not initialized case", args{wl: &myWorkload}, true, false},
		{"Nil case", args{wl: nil}, true, true},
		{"Correct case", args{wl: &myWorkload}, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantWorkingDB == true {
				//prepare DB for test
				err := Init()
				if err != nil {
					t.Errorf("Cannot create database - tests results can be corrupted\n")
				}
			} else {
				// enforce "Service database not initialized"
				workloadDatabase = nil
			}
			if err := validateInDB(tt.args.wl); (err != nil) != tt.wantErr {
				t.Errorf("validateInDB() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_updateInDB(t *testing.T) {
	myWorkload := tw.RDTWorkLoad{ID: "2626262626262626262"}
	type args struct {
		w *tw.RDTWorkLoad
	}
	tests := []struct {
		name          string
		args          args
		wantErr       bool
		wantWorkingDB bool
	}{
		{"DB not initialized case", args{w: &myWorkload}, true, false},
		{"Nil case", args{w: nil}, true, true},
		{"Correct case", args{w: &myWorkload}, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantWorkingDB == true {
				//prepare DB for test
				err := Init()
				if err != nil {
					t.Errorf("Cannot create database - tests results can be corrupted\n")
				}
			} else {
				// enforce "Service database not initialized"
				workloadDatabase = nil
			}
			if err := updateInDB(tt.args.w); (err != nil) != tt.wantErr {
				t.Errorf("updateInDB() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRelease(t *testing.T) {

	myWorkload := tw.RDTWorkLoad{ID: "362620909908098262"}
	type args struct {
		w *tw.RDTWorkLoad
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"Lack of COS name", args{w: &myWorkload}, false}, //this case is warning not error
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Release(tt.args.w); (err != nil) != tt.wantErr {
				t.Errorf("Release() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_populateEnforceRequest(t *testing.T) {
	// the same value for min and max cache
	var origCache uint32 = 2
	origpStateRatio := 3.0
	pStateMonitoringOn := "on"
	pStateMonitoringOff := "off"

	wMonitoringOn := tw.RDTWorkLoad{}
	wMonitoringOn.CoreIDs = []string{"3"}
	wMonitoringOn.Origin = "REST"
	wMonitoringOn.Status = "None"
	wMonitoringOn.Rdt.Cache.Max = &origCache
	wMonitoringOn.Rdt.Cache.Min = &origCache
	wMonitoringOn.PState.Ratio = &origpStateRatio
	wMonitoringOn.PState.Monitoring = &pStateMonitoringOn

	wMonitoringOff := tw.RDTWorkLoad{}
	wMonitoringOff.CoreIDs = []string{"3"}
	wMonitoringOff.Origin = "REST"
	wMonitoringOff.Status = "None"
	wMonitoringOff.Rdt.Cache.Max = &origCache
	wMonitoringOff.Rdt.Cache.Min = &origCache
	wMonitoringOff.PState.Ratio = &origpStateRatio
	wMonitoringOff.PState.Monitoring = &pStateMonitoringOff

	wPolicyExists := tw.RDTWorkLoad{}
	wPolicyExists.CoreIDs = []string{"3"}
	wPolicyExists.Origin = "REST"
	wPolicyExists.Status = "None"
	wPolicyExists.Policy = "silver"
	wPolicyExists.Rdt.Cache.Max = &origCache
	wPolicyExists.Rdt.Cache.Min = &origCache
	wPolicyExists.PState.Ratio = &origpStateRatio
	wPolicyExists.PState.Monitoring = &pStateMonitoringOn

	wWrongPolicyExists := tw.RDTWorkLoad{}
	wWrongPolicyExists.CoreIDs = []string{"3"}
	wWrongPolicyExists.Origin = "REST"
	wWrongPolicyExists.Status = "None"
	wWrongPolicyExists.Policy = "fakePolicyName"
	wWrongPolicyExists.Rdt.Cache.Max = &origCache
	wWrongPolicyExists.Rdt.Cache.Min = &origCache
	wWrongPolicyExists.PState.Ratio = &origpStateRatio
	wWrongPolicyExists.PState.Monitoring = &pStateMonitoringOn

	req := &tw.EnforceRequest{}

	type args struct {
		req *tw.EnforceRequest
		w   *tw.RDTWorkLoad
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"Correct monitoring ON lack of policy case", args{w: &wMonitoringOn, req: req}, false},
		{"Correct monitoring OFF lack of policy case", args{w: &wMonitoringOff, req: req}, false},
		{"Correct policy silver case", args{w: &wPolicyExists, req: req}, false},
		{"Fake policy fakePolicyName case", args{w: &wWrongPolicyExists, req: req}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := populateEnforceRequest(tt.args.req, tt.args.w); (err != nil) != tt.wantErr {
				t.Errorf("populateEnforceRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_validate(t *testing.T) {
	// the same value for min and max cache
	var origCache uint32 = 2
	var origMba uint32 = 50
	var wrongMbaValue uint32 = 130
	// Check if MBA is supported in the host. Error check not required
	isMbaSupported, _ = proc.IsMbaAvailable()

	w := tw.RDTWorkLoad{}
	w.CoreIDs = []string{"3"}
	w.Origin = "REST"
	w.Status = "None"
	w.Rdt.Cache.Max = &origCache
	w.Rdt.Cache.Min = &origCache
	w.Rdt.Mba.Percentage = &origMba

	wWrongMbaValue := tw.RDTWorkLoad{}
	wWrongMbaValue.CoreIDs = []string{"4"}
	wWrongMbaValue.Origin = "REST"
	wWrongMbaValue.Status = "None"
	wWrongMbaValue.Rdt.Cache.Max = &origCache
	wWrongMbaValue.Rdt.Cache.Min = &origCache
	wWrongMbaValue.Rdt.Mba.Percentage = &wrongMbaValue

	wLackOfCacheID := tw.RDTWorkLoad{}
	wLackOfCacheID.Origin = "REST"
	wLackOfCacheID.Status = "None"
	wLackOfCacheID.Rdt.Cache.Max = &origCache
	wLackOfCacheID.Rdt.Cache.Min = &origCache

	wPolicyExists := tw.RDTWorkLoad{}
	wPolicyExists.CoreIDs = []string{"3"}
	wPolicyExists.Origin = "REST"
	wPolicyExists.Status = "None"
	wPolicyExists.Policy = "silver"
	wPolicyExists.Rdt.Cache.Max = &origCache
	wPolicyExists.Rdt.Cache.Min = &origCache

	wNilAsMxCache := tw.RDTWorkLoad{}
	w.CoreIDs = []string{"3"}
	wNilAsMxCache.Origin = "REST"
	wNilAsMxCache.Status = "None"
	wNilAsMxCache.Rdt.Cache.Max = nil
	wNilAsMxCache.Rdt.Cache.Min = &origCache

	type args struct {
		w *tw.RDTWorkLoad
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"Correct case no PState", args{w: &w}, !isMbaSupported},
		{"Incorrect Mba Value provided", args{w: &wWrongMbaValue}, true},
		{"Lack of cache ID disabled", args{w: &wLackOfCacheID}, true},
		{"Policy silver case", args{w: &wPolicyExists}, false},
		{"Nil max cache", args{w: &wNilAsMxCache}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validate(tt.args.w); (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

package db

import (
	"os"
	"testing"

	workload "github.com/intel/rmd/modules/workload/types"
	"github.com/spf13/viper"
)

const (
	dbfilename = "./testdb"
	testuuid1  = "aaaaaaaa-bbbb-cccc-dddd-000000000001"
	testuuid2  = "aaaaaaaa-bbbb-cccc-dddd-000000000002"
	testuuidf  = "aaaaaaaa-bbbb-cccc-dddd-00000000000f"
)

func init() {
	// prepare necessary config options
	viper.Set("database.backend", "bolt")
	viper.Set("database.transport", dbfilename)
	viper.Set("database.dbname", "rmd")
}

func Setup(t *testing.T) DB {
	// remove old DB file if exists
	if _, err := os.Stat(dbfilename); os.IsNotExist(err) == false {
		os.Remove(dbfilename)
	}

	testdb, err := NewDB()
	if err != nil {
		// failed to create DB so test also failed
		t.Fatal("Unable to create database - exiting test")
		return nil
	}

	return testdb
}

func TestBoltDB_Initialize(t *testing.T) {

	db := Setup(t)

	type args struct {
		transport string
		dbname    string
	}
	tests := []struct {
		name    string
		b       DB
		args    args
		wantErr bool
	}{
		{"Proper DB initialization", db, args{"unused", "unused"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.b.Initialize(tt.args.transport, tt.args.dbname); (err != nil) != tt.wantErr {
				t.Errorf("BoltDB.Initialize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBoltDB_ValidateAndCreate(t *testing.T) {
	db := Setup(t)
	err := db.Initialize("unused", "unused")
	if err != nil {
		t.Fatal("DB initialization failure - exiting test")
	}

	tempMax := uint32(4)
	tempMin := uint32(2)

	type args struct {
		w *workload.RDTWorkLoad
	}
	tests := []struct {
		name        string
		b           DB
		args        args
		validateErr bool
		createErr   bool
	}{
		{"NIL workload testcase", db, args{nil}, true, true},
		{"Empty workload", db, args{w: &workload.RDTWorkLoad{}}, true, false},
		{"Proper workload with policy", db, args{w: &workload.RDTWorkLoad{
			ID:      "",
			UUID:    testuuid1,
			CoreIDs: []string{"5"},
			Policy:  "gold"},
		}, false, false},
		{"Proper workload with data", db, args{w: &workload.RDTWorkLoad{
			ID:      "",
			UUID:    testuuid2,
			CoreIDs: []string{"6"},
			Cache: struct {
				Max *uint32 `json:"max,omitempty"`
				Min *uint32 `json:"min,omitempty"`
			}{
				Max: &tempMax,
				Min: &tempMin,
			},
		},
		}, false, false},
		{"Duplicated core ids", db, args{w: &workload.RDTWorkLoad{
			ID:      "",
			UUID:    testuuidf,
			CoreIDs: []string{"6"},
			Policy:  "gold"},
		}, true, false},
		{"Duplicated uuid", db, args{w: &workload.RDTWorkLoad{
			ID:      "",
			UUID:    testuuid2,
			CoreIDs: []string{"1"},
			Policy:  "gold"},
		}, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if verr := tt.b.ValidateWorkload(tt.args.w); (verr != nil) != tt.validateErr {
				t.Errorf("BoltDB.ValidateWorkload() error = %v, validateErr %v", verr, tt.validateErr)
			} else {
				if verr == nil {
					if cerr := tt.b.CreateWorkload(tt.args.w); (cerr != nil) != tt.createErr {
						t.Errorf("BoltDB.CreateWorkload() error = %v, createErr %v", verr, tt.createErr)
					}
				}
			}
		})
	}
}

func TestBoltDB_CreateWorkload(t *testing.T) {

	db := Setup(t)
	err := db.Initialize("unused", "unused")
	if err != nil {
		t.Fatal("DB initialization failure - exiting test")
	}

	type args struct {
		w *workload.RDTWorkLoad
	}
	tests := []struct {
		name    string
		b       DB
		args    args
		wantErr bool
	}{
		{"NIL workload testcase", db, args{nil}, true},
		{"Proper workload", db, args{&workload.RDTWorkLoad{
			ID:      "1",
			CoreIDs: []string{"5"},
			Policy:  "gold"},
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.b.CreateWorkload(tt.args.w); (err != nil) != tt.wantErr {
				t.Errorf("BoltDB.CreateWorkload() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBoltDB_CreateGetDeleteWorkload(t *testing.T) {
	db := Setup(t)
	err := db.Initialize("unused", "unused")
	if err != nil {
		t.Fatal("DB initialization failure - exiting test")
	}

	type args struct {
		w *workload.RDTWorkLoad
	}

	wrkld1 := workload.RDTWorkLoad{}
	wrkld1.UUID = testuuid1
	wrkld1.CoreIDs = []string{"1"}
	wrkld1.Policy = "silver"
	wrkld2 := workload.RDTWorkLoad{}
	wrkld2.UUID = testuuid2
	wrkld2.CoreIDs = []string{"2"}
	cmax := uint32(4)
	cmin := uint32(2)
	wrkld2.Cache.Max = &cmax
	wrkld2.Cache.Min = &cmin

	err = db.CreateWorkload(&wrkld1)
	if err != nil {
		t.Fatalf("Failed to add 1st workload: %v", err.Error())
	}

	err = db.CreateWorkload(&wrkld2)
	if err != nil {
		t.Fatalf("Failed to add 2nd workload: %v", err.Error())
	}

	// Get first by ID
	reswrkld, err := db.GetWorkloadByID(wrkld1.ID)
	if err != nil {
		t.Fatalf("Failed to get 1st workload by ID: %v", err.Error())
	}

	if reswrkld.ID != wrkld1.ID ||
		reswrkld.UUID != wrkld1.UUID ||
		reswrkld.Policy != wrkld1.Policy ||
		reswrkld.CoreIDs[0] != wrkld1.CoreIDs[0] {
		t.Errorf("Original (%v) and received (%v) workload differ", wrkld1, reswrkld)
	}

	// Get second by UUID
	reswrkld, err = db.GetWorkloadByUUID(wrkld2.UUID)
	if err != nil {
		t.Fatalf("Failed to get 2nd workload by UUID: %v", err.Error())
	}

	if reswrkld.ID != wrkld2.ID ||
		reswrkld.UUID != wrkld2.UUID ||
		reswrkld.Policy != wrkld2.Policy ||
		reswrkld.CoreIDs[0] != wrkld2.CoreIDs[0] {
		t.Errorf("Original (%v) and received (%v) workload differ", wrkld2, reswrkld)
	}

	// Remove workload1
	err = db.DeleteWorkload(&wrkld1)
	if err != nil {
		t.Errorf("Failed to remove 1st workload")
	}

	// Remove workload2
	err = db.DeleteWorkload(&wrkld2)
	if err != nil {
		t.Errorf("Failed to remove 2nd workload")
	}
}

package db

import (
	"fmt"

	// from app import an config is really not a good idea.
	// uncouple it from APP. Or we can add it in a rmd/config
	"github.com/intel/rmd/db/config"
	libutil "github.com/intel/rmd/lib/util"
	"github.com/intel/rmd/model/types/workload"
	"github.com/intel/rmd/util"
)

// WorkloadTableName is the table name for workload
const WorkloadTableName = "workload"

// DB is the interface for a db engine
type DB interface {
	Initialize(transport, dbname string) error
	CreateWorkload(w *workload.RDTWorkLoad) error
	DeleteWorkload(w *workload.RDTWorkLoad) error
	UpdateWorkload(w *workload.RDTWorkLoad) error
	GetAllWorkload() ([]workload.RDTWorkLoad, error)
	GetWorkloadByID(id string) (workload.RDTWorkLoad, error)
	ValidateWorkload(w *workload.RDTWorkLoad) error
	QueryWorkload(query map[string]interface{}) ([]workload.RDTWorkLoad, error)
}

// NewDB return DB connection
func NewDB() (DB, error) {
	dbcon := config.NewConfig()
	if dbcon.Backend == "bolt" {
		return newBoltDB()
	} else if dbcon.Backend == "mgo" {
		return newMgoDB()
	} else {
		return nil, fmt.Errorf("Unsupported DB backend %s", dbcon.Backend)
	}
}

// this function does 3 things to valicate a user request workload is
// validate at data base layer
func validateWorkload(w workload.RDTWorkLoad, ws []workload.RDTWorkLoad) error {

	if len(w.ID) < 1 && len(w.TaskIDs) < 1 && len(w.CoreIDs) < 1 {
		return nil
	}

	// User post a workload id/uuid in it's request
	if w.ID != "" {
		for _, i := range ws {
			if w.ID == i.ID {
				return fmt.Errorf("Workload id %s has existed", w.ID)
			}
		}
	}

	// Validate if the task id of workload has existed.
	for _, t := range w.TaskIDs {
		for _, wi := range ws {
			if util.HasElem(wi.TaskIDs, t) {
				return fmt.Errorf("Taskid %s has existed in workload %s", t, wi.ID)
			}
		}
	}

	if len(w.CoreIDs) == 0 {
		return nil
	}

	// Validate if the core id of workload has overlap with crrent ones.
	bm, _ := libutil.NewBitmap(w.CoreIDs)
	bmsum, _ := libutil.NewBitmap("")

	for _, c := range ws {
		if len(c.CoreIDs) > 0 {
			tmpbm, _ := libutil.NewBitmap(c.CoreIDs)
			bmsum = bmsum.Or(tmpbm)
		}
	}

	bminter := bm.And(bmsum)

	if !bminter.IsEmpty() {
		return fmt.Errorf("CPU list %s has been assigned", bminter.ToString())
	}

	return nil
}

package db

import (
	"errors"
	"fmt"

	// from app import an config is really not a good idea.
	// uncouple it from APP. Or we can add it in a rmd/config
	"github.com/intel/rmd/internal/db/config"
	wltypes "github.com/intel/rmd/modules/workload/types"
	util "github.com/intel/rmd/utils"
	libutil "github.com/intel/rmd/utils/bitmap"
)

// WorkloadTableName is the table name for workload
const WorkloadTableName = "workload"

// MappingTableName contains mapping between UUID and WorkloadID
const MappingTableName = "mapping"

// DB is the interface for a db engine
type DB interface {
	Initialize(transport, dbname string) error
	CreateWorkload(w *wltypes.RDTWorkLoad) error
	DeleteWorkload(w *wltypes.RDTWorkLoad) error
	UpdateWorkload(w *wltypes.RDTWorkLoad) error
	GetAllWorkload() ([]wltypes.RDTWorkLoad, error)
	GetWorkloadByID(id string) (wltypes.RDTWorkLoad, error)
	GetWorkloadByUUID(id string) (wltypes.RDTWorkLoad, error)
	ValidateWorkload(w *wltypes.RDTWorkLoad) error
	QueryWorkload(query map[string]interface{}) ([]wltypes.RDTWorkLoad, error)
}

// NewDB return DB connection
func NewDB() (DB, error) {
	dbcon := config.NewConfig()
	if dbcon.Backend == "bolt" {
		return newBoltDB()
	} else if dbcon.Backend == "mgo" {
		//return newMgoDB() commented out now.
		return nil, fmt.Errorf("Mongo DB is not currently supported in RMD: Select bolt backend db")
	} else {
		return nil, fmt.Errorf("Unsupported DB backend %s", dbcon.Backend)
	}
}

// this function does 3 things to validate a user request workload is
// validate at data base layer
func validateWorkload(w wltypes.RDTWorkLoad, ws []wltypes.RDTWorkLoad) error {

	if len(w.ID) == 0 && len(w.TaskIDs) == 0 && len(w.CoreIDs) == 0 {
		return errors.New("Incomplete Workload definition")
	}

	// User post a workload id/uuid in it's request
	if w.ID != "" {
		for _, i := range ws {
			if w.ID == i.ID {
				return fmt.Errorf("Workload with id %s already exists", w.ID)
			}
		}
	}

	// User post a workload id/uuid in it's request
	if w.UUID != "" {
		for _, i := range ws {
			if w.UUID == i.UUID {
				return fmt.Errorf("UUID %s already exists in workload %s", w.UUID, i.ID)
			}
		}
	}

	// Validate if the task id of workload has existed.
	for _, t := range w.TaskIDs {
		for _, wi := range ws {
			if util.HasElem(wi.TaskIDs, t) {
				return fmt.Errorf("TaskID %s already exists in workload %s", t, wi.ID)
			}
		}
	}

	if len(w.CoreIDs) == 0 {
		return nil
	}

	// Validate if the core id of workload has overlap with current ones.
	bm, err := libutil.NewBitmap(w.CoreIDs)
	if err != nil {
		return err
	}
	bmsum, err := libutil.NewBitmap("")
	if err != nil {
		return err
	}
	for _, c := range ws {
		if len(c.CoreIDs) > 0 {
			tmpbm, err := libutil.NewBitmap(c.CoreIDs)
			if err != nil {
				return err
			}
			bmsum = bmsum.Or(tmpbm)
		}
	}

	bminter := bm.And(bmsum)

	var coreBits uint32
	if bminter.Len > 0 {
		coreBits = uint32(bminter.Bits[0])
	}
	var cores = []string{}
	if bminter.Len > 0 && coreBits > 0 {
		var bit uint32
		for bit = 0; bit < uint32(bminter.Len); bit++ {
			if (coreBits % 2) == 1 {
				cores = append(cores, fmt.Sprint(bit))
			}
			coreBits = coreBits / 2
		}
	}

	if !bminter.IsEmpty() {
		return fmt.Errorf("CPU list %s has been assigned", cores)
	}

	return nil
}

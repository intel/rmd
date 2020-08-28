package db

import (
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"strings"
	"sync"

	bolt "github.com/etcd-io/bbolt"

	"github.com/intel/rmd/internal/db/config"
	wltypes "github.com/intel/rmd/modules/workload/types"
	util "github.com/intel/rmd/utils"
)

var boltSession *bolt.DB

var boltSessionOnce sync.Once

// BoltDB connection
type BoltDB struct {
	session *bolt.DB
}

// We thought, open a file, means open a session.
// Unity Concept with mongodb
func getSession() error {
	var err error
	boltSessionOnce.Do(func() {
		conf := config.NewConfig()
		var isfile bool
		isfile, err = util.IsRegularFile(conf.Transport)
		if err != nil {
			// if error is "no such file or directory" then it is a recoverable situation
			// otherwise error should be forwarded
			if !strings.Contains(err.Error(), "no such file") {
				boltSession = nil
				err = errors.New("Provided database path is not a regular file")
				return
			}
		} else if !isfile {
			boltSession = nil
			err = errors.New("Provided database path is not a regular file")
			return
		}
		// no error till now - open/recreate database file
		boltSession, err = bolt.Open(conf.Transport, 0600, nil)
	})
	return err
}

func closeSession() {
}

func newBoltDB() (DB, error) {
	var db BoltDB
	if err := getSession(); err != nil {
		return &db, err
	}
	db.session = boltSession
	if err := db.Initialize("", ""); err != nil {
		return &db, err
	}
	return &db, nil
}

// Initialize creates two buckets: for storing workloads and UUID-ID mapping
func (b *BoltDB) Initialize(transport, dbname string) error {
	return b.session.Update(func(tx *bolt.Tx) error {
		// First touch a Bucket for workloads ...
		_, err := tx.CreateBucketIfNotExists([]byte(WorkloadTableName))
		if err != nil {
			return err
		}
		// ... and for UUID-ID mapping
		_, err = tx.CreateBucketIfNotExists([]byte(MappingTableName))
		if err != nil {
			return err
		}
		return nil
	})

}

// ValidateWorkload from data base view
func (b *BoltDB) ValidateWorkload(w *wltypes.RDTWorkLoad) error {
	if w == nil {
		return errors.New("NIL workload given")
	}
	/* When create a new workload we need to verify that the new PIDs
	   we the workload specified should not existed */
	ws, err := b.GetAllWorkload()
	if err != nil {
		return err
	}
	return validateWorkload(*w, ws)
}

// CreateWorkload creates workload in db
func (b *BoltDB) CreateWorkload(w *wltypes.RDTWorkLoad) error {
	if w == nil {
		return errors.New("NIL workload given")
	}

	return b.session.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(WorkloadTableName))
		if bucket == nil {
			return errors.New("Bucket fetching failed")
		}
		if (w != nil) && (w.ID == "") {
			// Generate ID for the workload.
			id, err := bucket.NextSequence()
			if err != nil {
				return err
			}
			w.ID = strconv.Itoa(int(id))
		}
		// Marshal  data into bytes.
		buf, err := json.Marshal(w)
		if err != nil {
			return err
		}

		// add entry to mapping bucket if UUID exists
		if len(w.UUID) > 0 {
			mb := tx.Bucket([]byte(MappingTableName))
			err = mb.Put([]byte(w.UUID), []byte(w.ID))
			if err != nil {
				return errors.New("Failed to add workload mapping")
			}
		}

		// Persist bytes to users bucket.
		return bucket.Put([]byte(w.ID), buf)
	})
}

// DeleteWorkload removes workload from db
func (b *BoltDB) DeleteWorkload(w *wltypes.RDTWorkLoad) error {
	if w == nil {
		return errors.New("NIL workload given")
	}

	return b.session.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(WorkloadTableName))
		if bucket == nil {
			return errors.New("Bucket 'bucket' creation failed")
		}

		if len(w.UUID) > 0 {
			mb := tx.Bucket([]byte(MappingTableName))
			if mb == nil {
				return errors.New("Bucket 'mba' mcreation failed")
			}

			err := mb.Delete([]byte(w.UUID))
			if err != nil {
				return errors.New("Failed to delete mapping for given UUID")
			}
		}

		return bucket.Delete([]byte(w.ID))
	})
}

// UpdateWorkload updates
func (b *BoltDB) UpdateWorkload(w *wltypes.RDTWorkLoad) error {
	if w == nil {
		return errors.New("NIL workload given")
	}
	if len(w.ID) == 0 {
		// unable to update workload without ID
		return errors.New("Cannot update worload without ID")
	}
	// TODO The UUID field should never be changed in Workload
	// but ensure there's no inconsistency if somehow UUID has been changed!
	return b.session.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(WorkloadTableName))
		if bucket == nil {
			return errors.New("Bucket fetching failed")
		}
		buf, err := json.Marshal(w)
		if err != nil {
			return err
		}

		return bucket.Put([]byte(w.ID), buf)
	})
}

// GetAllWorkload returns all workloads in db
func (b *BoltDB) GetAllWorkload() ([]wltypes.RDTWorkLoad, error) {
	ws := []wltypes.RDTWorkLoad{}
	err := b.session.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(WorkloadTableName))
		if bucket == nil {
			return errors.New("Bucket fetching failed")
		}
		cursor := bucket.Cursor()
		if cursor == nil {
			return errors.New("Cursor creation failed")
		}
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			w := wltypes.RDTWorkLoad{}
			err := json.Unmarshal(v, &w)
			if err != nil {
				return err
			}
			ws = append(ws, w)
		}
		return nil
	})
	return ws, err
}

// GetWorkloadByID by ID
func (b *BoltDB) GetWorkloadByID(id string) (wltypes.RDTWorkLoad, error) {
	w := wltypes.RDTWorkLoad{}
	err := b.session.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(WorkloadTableName))
		if bucket == nil {
			return errors.New("Bucket fetching failed")
		}
		v := bucket.Get([]byte(id))
		if v == nil {
			return errors.New("No workload found for given ID")
		}

		return json.Unmarshal(v, &w)
	})
	return w, err
}

// QueryWorkload with given params
func (b *BoltDB) QueryWorkload(query map[string]interface{}) ([]wltypes.RDTWorkLoad, error) {
	ws, err := b.GetAllWorkload()
	if err != nil {
		return []wltypes.RDTWorkLoad{}, err
	}

	rws := []wltypes.RDTWorkLoad{}

	for _, w := range ws {
		find := true
		for k, v := range query {
			if _, ok := reflect.TypeOf(w).FieldByName(k); ok {
				if !reflect.DeepEqual(reflect.ValueOf(w).FieldByName(k).Interface(), v) {
					find = false
					break
				}
			} else {
				find = false
			}
		}
		if find {
			rws = append(rws, w)
		}
	}
	return rws, nil
}

// GetWorkloadByUUID Returns workload specified by UUID (if such exists in DB)
func (b *BoltDB) GetWorkloadByUUID(id string) (wltypes.RDTWorkLoad, error) {
	w := wltypes.RDTWorkLoad{}
	err := b.session.View(func(tx *bolt.Tx) error {
		// get workload and mapping buckets
		wb := tx.Bucket([]byte(WorkloadTableName))
		if wb == nil {
			return errors.New("Bucket fetching failed")
		}
		mb := tx.Bucket([]byte(MappingTableName))
		if mb == nil {
			return errors.New("Bucket fetching failed")
		}
		// first get mapping entry
		mapentry := mb.Get([]byte(id))
		if mapentry == nil {
			return errors.New("No workload found for given UUID")
		}

		// if no error get workload from DB
		wlentry := wb.Get([]byte(mapentry))
		if wlentry == nil {
			return errors.New("No workload found for given ID")
		}
		return json.Unmarshal(wlentry, &w)
	})
	return w, err
}

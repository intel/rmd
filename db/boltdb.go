package db

import (
	"encoding/json"
	"reflect"
	"strconv"
	"sync"

	bolt "github.com/coreos/bbolt"

	"github.com/intel/rmd/db/config"
	"github.com/intel/rmd/model/types/workload"
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
	return &db, nil
}

// Initialize does initialize
func (b *BoltDB) Initialize(transport, dbname string) error {
	b.session.Update(func(tx *bolt.Tx) error {
		// First touch a Bucket
		_, err := tx.CreateBucketIfNotExists([]byte(WorkloadTableName))
		if err != nil {
			return err
		}
		return nil
	})

	return nil
}

// ValidateWorkload from data base view
func (b *BoltDB) ValidateWorkload(w *workload.RDTWorkLoad) error {
	/* When create a new workload we need to verify that the new PIDs
	   we the workload specified should not existed */
	ws, err := b.GetAllWorkload()
	if err != nil {
		return err
	}
	if err = validateWorkload(*w, ws); err != nil {
		return err
	}
	return nil
}

// CreateWorkload creates workload in db
func (b *BoltDB) CreateWorkload(w *workload.RDTWorkLoad) error {
	return b.session.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(WorkloadTableName))

		if w.ID == "" {
			// Generate ID for the workload.
			id, _ := b.NextSequence()
			w.ID = strconv.Itoa(int(id))
		}
		// Marshal  data into bytes.
		buf, err := json.Marshal(w)
		if err != nil {
			return err
		}
		// Persist bytes to users bucket.
		return b.Put([]byte(w.ID), buf)
	})
}

// DeleteWorkload removes workload from db
func (b *BoltDB) DeleteWorkload(w *workload.RDTWorkLoad) error {
	return b.session.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(WorkloadTableName))
		return b.Delete([]byte(w.ID))
	})
}

// UpdateWorkload updates
func (b *BoltDB) UpdateWorkload(w *workload.RDTWorkLoad) error {

	return b.session.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(WorkloadTableName))

		buf, err := json.Marshal(w)
		if err != nil {
			return err
		}

		return b.Put([]byte(w.ID), buf)
	})
}

// GetAllWorkload returns all workloads in db
func (b *BoltDB) GetAllWorkload() ([]workload.RDTWorkLoad, error) {
	ws := []workload.RDTWorkLoad{}
	err := b.session.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(WorkloadTableName))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			w := workload.RDTWorkLoad{}
			json.Unmarshal(v, &w)
			ws = append(ws, w)
		}
		return nil
	})
	return ws, err
}

// GetWorkloadByID by ID
func (b *BoltDB) GetWorkloadByID(id string) (workload.RDTWorkLoad, error) {
	w := workload.RDTWorkLoad{}
	err := b.session.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(WorkloadTableName))
		v := b.Get([]byte(id))
		return json.Unmarshal(v, &w)
	})
	return w, err
}

// QueryWorkload with given params
func (b *BoltDB) QueryWorkload(query map[string]interface{}) ([]workload.RDTWorkLoad, error) {
	ws, err := b.GetAllWorkload()
	if err != nil {
		return []workload.RDTWorkLoad{}, err
	}

	rws := []workload.RDTWorkLoad{}

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

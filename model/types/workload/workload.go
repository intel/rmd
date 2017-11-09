package workload

import "github.com/intel/rmd/lib/resctrl"

const (
	// Successful status
	Successful = "Successful"
	// Failed enfoced
	Failed = "Failed"
	// Invalid workload
	Invalid = "Invalid"
	// None status
	None = "None"
)

//RDTWorkLoad is the workload struct of RMD
type RDTWorkLoad struct {
	// ID
	ID string `json:"id,omitempty"`
	// core ids, the work load run on top of cores/cpus
	CoreIDs []string `json:"core_ids,omitempty"`
	// task ids, the work load's task ids
	TaskIDs []string `json:"task_ids,omitempty"`
	// policy the workload want to apply
	Policy string `json:"policy,omitempty"`
	// Status
	Status string `json:"status"`
	// CosNamej
	CosName string `json:"cos_name"`
	// Max Cache ways, use pointer to distinguish 0 value and empty value
	MaxCache *uint32 `json:"max_cache,omitempty"`
	// Min Cache ways, use pointer to distinguish 0 value and empty value
	MinCache *uint32 `json:"min_cache,omitempty"`
}

// EnforceRequest build this struct when create Resasscciation
type EnforceRequest struct {
	// all resassociations on the host
	Resall map[string]*resctrl.ResAssociation
	// max cache ways
	MaxWays uint32
	// min cache ways, not used yet
	MinWays uint32
	// enforce on which cache ids
	CacheIDs []uint32
	// consume from base group or not
	Consume bool
	// request type
	Type string
}

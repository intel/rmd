package types

import "github.com/intel/rmd/utils/resctrl"

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
	// Cache module (RDT) related settings
	Cache struct {
		// Max Cache ways, use pointer to distinguish 0 value and empty value
		Max *uint32 `json:"max,omitempty"`
		// Min Cache ways, use pointer to distinguish 0 value and empty value
		Min *uint32 `json:"min,omitempty"`
	} `json:"cache,omitempty"`
	// P-State module related settings
	PState struct {
		// PstateBR, optional field, refer to Bran Ratio in P-State plugin
		Ratio *float64 `json:"ratio,omitempty"`
		// Monitoring, optional field, used together with PstateBR, declared as pointer to distinguish on, off and not given
		Monitoring *string `json:"monitoring,omitempty"`
	} `json:"pstate,omitempty"`
	// UUID field is for storing OpenStack instance UUID
	UUID string `json:"uuid,omitempty"`
	// Origin, mandatory field, is for distinction who is responsible for current workload (REST API / Notification)
	// possible values: REST and OPENSTACK
	Origin string `json:"origin"`
}

// EnforceRequest build this struct when create Resasscciation
type EnforceRequest struct {
	// all resassociations on the host
	Resall map[string]*resctrl.ResAssociation
	// max cache ways
	MaxWays uint32
	// min cache ways, not used yet
	MinWays uint32
	// cache specification is not mandatory, this flag marks if cache values are used
	UseCache bool
	// enforce on which cache ids
	CacheIDs []uint32
	// consume from base group or not
	Consume bool
	// request type
	Type string
	// Power (P-State) related params are optional
	PState bool
	// Branch ratio (used only if PState flag above is true)
	PStateBR float64
	// Monitoring (used only if PState flag above is true)
	PStateMonitoring bool
}

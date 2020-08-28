package types

import (
	"github.com/intel/rmd/modules/cache"
	libutil "github.com/intel/rmd/utils/bitmap"
	"github.com/intel/rmd/utils/resctrl"
)

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

//UserRDTWorkLoad is the workload struct of RMD used by User
type UserRDTWorkLoad struct {
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
	// RDT module (RDT) related settings (Cache, MBA)
	Rdt struct {
		// Cache Settings
		Cache struct {
			// Max Cache ways, use pointer to distinguish 0 value and empty value
			Max *uint32 `json:"max,omitempty"`
			// Min Cache ways, use pointer to distinguish 0 value and empty value
			Min *uint32 `json:"min,omitempty"`
		} `json:"cache,omitempty"`
		// MBA settings
		Mba struct {
			// MBA values to be specified in Percentage
			Percentage *uint32 `json:"percentage,omitempty"`
			// MBA values to be specified in MB per sec
			Mbps *uint32 `json:"mbps,omitempty"`
		} `json:"mba,omitempty"`
	} `json:"rdt,omitempty"`
	// Plugins contains information about RMD plugins and theirs settings
	Plugins map[string]map[string]interface{} `json:"plugins,omitempty"`
	// UUID field is for storing OpenStack instance UUID
	UUID string `json:"uuid,omitempty"`
	// Origin, mandatory field, is for distinction who is responsible for current workload (REST API / Notification)
	// possible values: REST and OPENSTACK
	Origin string `json:"origin"`
}

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
	// RDT module (RDT) related settings (Cache, MBA)
	Rdt struct {
		// Cache Settings
		Cache struct {
			// Max Cache ways, use pointer to distinguish 0 value and empty value
			Max *uint32 `json:"max,omitempty"`
			// Min Cache ways, use pointer to distinguish 0 value and empty value
			Min *uint32 `json:"min,omitempty"`
		} `json:"cache,omitempty"`
		// MBA settings
		Mba struct {
			Percentage *uint32 `json:"percentage,omitempty"`
			Mbps       *uint32 `json:"mbps,omitempty"`
		} `json:"mba,omitempty"`
	} `json:"rdt,omitempty"`
	// Plugins contains information about RMD plugins and theirs settings
	Plugins map[string]map[string]interface{} `json:"plugins,omitempty"`
	// UUID field is for storing OpenStack instance UUID
	UUID string `json:"uuid,omitempty"`
	// Origin, mandatory field, is for distinction who is responsible for current workload (REST API / Notification)
	// possible values: REST and OPENSTACK
	Origin string `json:"origin"`
	// BackendPluginInfo contains backend related information to handle RMD plugins in code
	// There is no reason to return those info to User
	BackendPluginInfo map[string]string `json:"backend_plugin_info,omitempty"`
}

// EnforceRequest build this struct when create ResAssociation
type EnforceRequest struct {
	// all resassociations on the host
	Resall map[string]*resctrl.ResAssociation
	// max cache ways
	MaxWays uint32
	// min cache ways, not used yet
	MinWays uint32
	// cache specification is not mandatory, this flag marks if cache values are used
	UseCache bool
	// enforce RDT request on these socket ID's
	SocketIDs []uint32
	// Mba
	UseMba bool
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

// RDTEnforce cotains all Cache results and MBA params together
type RDTEnforce struct {
	// all resassociations on the host
	Resall map[string]*resctrl.ResAssociation
	// target level of cache
	TargetLev string
	// target for MBA
	TargetMba string
	// cache calculations in all sockets
	CandidateCache map[string]*libutil.Bitmap
	// mba calculations in all sockets
	CandidateMba map[string]*uint32

	ChangedRes map[string]*resctrl.ResAssociation
	// reserved cache values in all sockets
	Reserved map[string]*cache.Reserved
	// available cache schemata
	AvailableSchemata map[string]*libutil.Bitmap
}

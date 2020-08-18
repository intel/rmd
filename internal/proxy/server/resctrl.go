package proxyserver

import (
	"github.com/intel/rmd/internal/proxy/types"
	"github.com/intel/rmd/utils/pqos"
	"github.com/intel/rmd/utils/resctrl"
)

// Commit resource group
func (*Proxy) Commit(r types.ResctrlRequest, dummy *int) error {
	// return resctrl.Commit(&r.Res, r.Name)
	// Call PQOS Wrapper
	pqos.AllocateCLOS(&r.Res, r.Name)
	return nil
}

// DestroyResAssociation remove resource association
func (*Proxy) DestroyResAssociation(grpName string, dummy *int) error {
	return resctrl.DestroyResAssociation(grpName)
}

// RemoveTasks move tasks to default group
func (*Proxy) RemoveTasks(tasks []string, dummy *int) error {
	// return resctrl.RemoveTasks(tasks)
	// Call PQOS Wrapper
	pqos.DeallocateCLOS(tasks)
	return nil
}

// EnableCat mounts resctrl
func (*Proxy) EnableCat(dummy *int, result *bool) error {
	*result = resctrl.EnableCat()
	return nil
}

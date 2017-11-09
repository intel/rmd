package proxy

import (
	"github.com/intel/rmd/lib/resctrl"
)

// Request struct of to rpc server
type Request struct {
	Name string
	Res  resctrl.ResAssociation
}

// Commit resource group
func (*Proxy) Commit(r Request, dummy *int) error {
	return resctrl.Commit(&r.Res, r.Name)
}

// DestroyResAssociation remove resource association
func (*Proxy) DestroyResAssociation(grpName string, dummy *int) error {
	return resctrl.DestroyResAssociation(grpName)
}

// RemoveTasks move tasks to default group
func (*Proxy) RemoveTasks(tasks []string, dummy *int) error {
	return resctrl.RemoveTasks(tasks)
}

// EnableCat mounts resctrl
func (*Proxy) EnableCat(dummy *int, result *bool) error {
	*result = resctrl.EnableCat()
	return nil
}

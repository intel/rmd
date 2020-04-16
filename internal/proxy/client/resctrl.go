package proxyclient

import (
	"fmt"

	"github.com/intel/rmd/internal/proxy/types"
	"github.com/intel/rmd/utils/resctrl"
)

// GetResAssociation returns all resource group association
func GetResAssociation() map[string]*resctrl.ResAssociation {
	return resctrl.GetResAssociation()
}

// GetRdtCosInfo returns RDT information
func GetRdtCosInfo() map[string]*resctrl.RdtCosInfo {
	return resctrl.GetRdtCosInfo()
}

// IsIntelRdtMounted will check if resctrl mounted or not
func IsIntelRdtMounted() bool {
	return resctrl.IsIntelRdtMounted()
}

// Commit resctrl.ResAssociation with given name
func Commit(r *resctrl.ResAssociation, name string) error {
	// TODO how to get error reason
	req := types.ResctrlRequest{
		Name: name,
		Res:  *r,
	}
	return Client.Call("Proxy.Commit", req, nil)
}

// DestroyResAssociation by resource group name
func DestroyResAssociation(name string) error {
	// TODO how to get error reason
	// Add checking before using client and do reconnect
	return Client.Call("Proxy.DestroyResAssociation", name, nil)
}

// RemoveTasks moves tasks to default resource group
func RemoveTasks(tasks []string) error {
	// TODO how to get error reason
	return Client.Call("Proxy.RemoveTasks", tasks, nil)
}

// EnableCat enable cat feature on host
func EnableCat() error {
	var result bool
	if err := Client.Call("Proxy.EnableCat", 0, &result); err != nil {
		return err
	}
	if result {
		return nil
	}
	return fmt.Errorf("Can not enable cat")
}

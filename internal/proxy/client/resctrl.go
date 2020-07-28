package proxyclient

import (
	"fmt"

	"github.com/intel/rmd/internal/proxy/types"
	"github.com/intel/rmd/utils/resctrl"
)

// GetResAssociation returns all resource group association
func GetResAssociation(availableCLOS []string) map[string]*resctrl.ResAssociation {
	return resctrl.GetResAssociation(availableCLOS)
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
	// fmt.Println("CLient Side Commit : ", name, r)
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
	return Client.Call("Proxy.RemoveTasks", tasks, nil)
}

// RemoveCores moves cores to default resource group
func RemoveCores(cores []string) error {
	return Client.Call("Proxy.RemoveCores", cores, nil)
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

// ResetCOSParamsToDefaults resets L3 cache and MBA to default values for common COS#
func ResetCOSParamsToDefaults(cosName string) error {
	// Call PQOS Wrapper
	return Client.Call("Proxy.ResetCOSParamsToDefaults", cosName, nil)
}

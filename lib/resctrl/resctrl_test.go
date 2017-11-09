// +build linux

package resctrl

import (
	"fmt"
	"testing"
)

func TestGetResAssociation(t *testing.T) {
	ress := GetResAssociation()
	for name, res := range ress {
		if name == "CG1" {
			fmt.Println(name)
			fmt.Println(res)
			fmt.Println(res.Schemata["L3CODE"])
		}
	}
}

func TestGetRdtCosInfo(t *testing.T) {

	infos := GetRdtCosInfo()
	for name, info := range infos {
		fmt.Println(name)
		fmt.Println(info)
	}
}

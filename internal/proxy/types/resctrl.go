package types

import (
	"github.com/intel/rmd/utils/resctrl"
)

// ResctrlRequest struct of to rpc server
type ResctrlRequest struct {
	Name string
	Res  resctrl.ResAssociation
}

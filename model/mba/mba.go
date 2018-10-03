package mba

import (
	"github.com/intel/rmd/lib/mba"
	"github.com/intel/rmd/lib/proc"
)

// MbaInfo is the mba information
type MbaInfo struct {
	Mba     bool `json:"mba"`
	MbaOn   bool `json:"mba_enable,omitempty"`
	MbaStep int  `json:"mba_step,omitempty"`
	MbaMin  int  `json:"mba_min,omitempty"`
}

// Get returns mba status
func (c *MbaInfo) Get() error {
	flag, err := proc.IsMbaAvailable()
	if err == nil {
		c.Mba = flag
		if c.Mba {
			flag, err = proc.IsEnableMba()
			if err != nil {
				return err
			}
			c.MbaOn = flag
			if c.MbaOn {
				mbaStep, mbaMin, err := mba.GetMbaInfo()
				if err == nil {
					c.MbaStep = mbaStep
					c.MbaMin = mbaMin
				} else {
					return err
				}
			}
		}
	} else {
		return err
	}
	return nil
}

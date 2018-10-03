package mba

import (
	"errors"
	"testing"

	. "github.com/prashantv/gostub"

	"github.com/intel/rmd/lib/mba"
	"github.com/intel/rmd/lib/proc"
)

func TestMbaInfoGet(t *testing.T) {
	subs := StubFunc(&proc.IsMbaAvailable, true, nil)
	subs.StubFunc(&proc.IsEnableMba, true, nil)
	subs.StubFunc(&mba.GetMbaInfo, 10, 20, nil)
	defer subs.Reset()

	// Check for normal status
	m := MbaInfo{}
	err := m.Get()
	if m.MbaStep != 10 || m.MbaMin != 20 || m.MbaOn != true || m.Mba != true || err != nil {
		t.Error("Get Mba info error: normal")
	}

	// Check if CPU flag "mba" not exist
	subs.StubFunc(&proc.IsMbaAvailable, false, nil)
	m = MbaInfo{}
	err = m.Get()
	if m.MbaStep != 0 || m.MbaMin != 0 || m.MbaOn != false || m.Mba != false || err != nil {
		t.Error("Get Mba info error: CPU flag \"mba\" not exist")
	}

	// Check if user didn't mount resctrl
	subs.StubFunc(&proc.IsMbaAvailable, true, nil)
	subs.StubFunc(&proc.IsEnableMba, false, nil)
	m = MbaInfo{}
	err = m.Get()
	if m.MbaStep != 0 || m.MbaMin != 0 || m.MbaOn != false || m.Mba != true || err != nil {
		t.Error("Get Mba info error: user didn't mount resctrl")
	}

	// Check if system error
	subs.StubFunc(&proc.IsEnableMba, true, nil)
	subs.StubFunc(&mba.GetMbaInfo, 0, 0, errors.New("test"))
	m = MbaInfo{}
	err = m.Get()
	if err == nil {
		t.Error("Get Mba info error: system error")
	}
}

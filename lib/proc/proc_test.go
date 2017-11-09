package proc

import (
	"fmt"
	"strconv"
	"testing"

	//"github.com/intel/rmd/lib/cpu"
	"github.com/intel/rmd/lib/util"
	// Better to make that package a util package
	"github.com/intel/rmd/test/test_helpers"
)

func TestListProcesses(t *testing.T) {
	ps := ListProcesses()
	if len(ps) == 0 {
		t.Errorf("Faild to list all process\n")
	}

}

func TestGetCPUAffinity(t *testing.T) {

	ospid, _ := testhelpers.CreateNewProcess("sleep 100")

	pid := strconv.Itoa(ospid.Pid)

	oldaf, err := GetCPUAffinity(pid)
	if err != nil {
		t.Errorf("Failed to get CPU affinity for process id %s", pid)
	}

	// should verify the default cpu affinity
	fmt.Println(oldaf.ToHumanString())

	af, _ := util.NewBitmap("f")

	err = SetCPUAffinity(pid, af)
	if err != nil {
		t.Errorf("Failed to set CPU affinity for process id %s", pid)
	}

	afset, err := GetCPUAffinity(pid)
	if err != nil {
		t.Errorf("Failed to get CPU affinity for process id %s", pid)
	}

	if af.ToHumanString() != afset.ToHumanString() {
		t.Errorf("Error to set CPU affinity for process id %s", pid)
	}

	testhelpers.CleanupProcess(ospid)
}

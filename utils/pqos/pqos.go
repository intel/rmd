package pqos

//#cgo CFLAGS: -W -Wall -Wextra -Wstrict-prototypes -Wmissing-prototypes -Wmissing-declarations -Wold-style-definition -Wpointer-arith -Wcast-qual -Wundef -Wwrite-strings -Wformat -Wformat-security -fstack-protector -fPIE -D_FORTIFY_SOURCE=2 -Wunreachable-code -Wsign-compare -Wno-endif-labels -g -O2
//#cgo LDFLAGS: -lpqos -lpthread
//int pqos_wrapper_main(int clos, int pid, int core, int s0, int s1, int num_of_pid, int num_of_cores);
//int pqos_wrapper_init();
import "C"
import (
	"errors"
	"strconv"
	"strings"

	"github.com/intel/rmd/utils/resctrl"
	log "github.com/sirupsen/logrus"
)

var availableCLOS []string
var occupiedCLOS []string

// StartCLOSPool ...
func StartCLOSPool() {
	var numOfclos int
	numOfclos, err := resctrl.GetNumOfCLOS()
	if err != nil {
		log.Error("Error in resctrl code CLOS: ", err)
	} else {
		availableCLOS = make([]string, numOfclos-2)
		occupiedCLOS = append(occupiedCLOS, []string{"COS0", "COS1"}...)
		for index := range availableCLOS {
			availableCLOS[index] = "COS" + strconv.Itoa(index+2)
		}
		log.Println(availableCLOS)
	}
}

// OccupyCLOS ...
func OccupyCLOS() error {
	if len(availableCLOS) > 0 {
		occupiedCLOS = append(occupiedCLOS, availableCLOS[0])
	} else {
		log.Error("No Available CLOS\n")
	}
	return nil
}

// UpdateAvailableCLOS ...
func UpdateAvailableCLOS() error {
	if len(availableCLOS) > 1 {
		availableCLOS = availableCLOS[1:]
	} else {
		availableCLOS = availableCLOS[:0]
	}
	return nil
}

// ErrInCLOS ...
func ErrInCLOS() {
	clos := occupiedCLOS[len(occupiedCLOS)-1]
	occupiedCLOS = occupiedCLOS[:len(occupiedCLOS)-1]
	availableCLOS = append(availableCLOS, clos)
}

// GetAvailableCLOS ...
func GetAvailableCLOS() []string {
	return availableCLOS
}

// GetOccupiedCLOS ...
func GetOccupiedCLOS() []string {
	return occupiedCLOS
}

// RecentlyOccupiedCLOS ...
func RecentlyOccupiedCLOS() string {
	return occupiedCLOS[len(occupiedCLOS)-1]
}

// AllocateCLOS ...
func AllocateCLOS(res *resctrl.ResAssociation, name string) {
	var clos int
	// Supported names are only COSx where x is a non-negative value
	if strings.HasPrefix(name, "COS") && len(name) > 3 {
		id, err := strconv.Atoi(name[3:])
		if err != nil || clos < 0 {
			log.Errorf("Invalid CLOS name given: %v Using COS0 instead.", name)
			clos = 0
		}
		clos = id
	} else {
		clos = 0
	}

	var pid, numPid int
	numPid = 0
	if len(res.Tasks) > 0 {
		pid, _ = strconv.Atoi(res.Tasks[0])
		numPid = 1
	}

	core, _ := strconv.ParseUint(strings.Split(res.CPUs, ",")[1], 16, 64)

	var s = []uint64{0, 0}
	s[res.CacheSchemata["L3"][0].ID], _ = strconv.ParseUint(res.CacheSchemata["L3"][0].Mask, 16, 64)
	s[res.CacheSchemata["L3"][1].ID], _ = strconv.ParseUint(res.CacheSchemata["L3"][1].Mask, 16, 64)
	log.Debugf("AllocateCLOS: %x %x", s[0], s[1])

	C.pqos_wrapper_main(C.int(clos), C.int(pid), C.int(core), C.int(s[0]), C.int(s[1]), C.int(numPid), C.int(1))
}

// DeallocateCLOS ...
func DeallocateCLOS(Tasks []string) {

}

// Init ...
func Init() error {
	result := C.pqos_wrapper_init()
	if result != 0 {
		// pqos_wrapper_init() returns non-zero value in case of initialization failure
		return errors.New("Failed to initialize PQOS driver")
	}
	return nil
}

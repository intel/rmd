package pqos

//#cgo CFLAGS: -W -Wall -Wextra -Wstrict-prototypes -Wmissing-prototypes -Wmissing-declarations -Wold-style-definition -Wpointer-arith -Wcast-qual -Wundef -Wwrite-strings -Wformat -Wformat-security -fstack-protector -fPIE -D_FORTIFY_SOURCE=2 -Wunreachable-code -Wsign-compare -Wno-endif-labels -g -O2
//#cgo LDFLAGS: -lpqos -lpthread
//int pqos_wrapper_init();
//int pqos_wrapper_check_mba_support(int *mbaMode);
//int pqos_wrapper_finish();
//int pqos_wrapper_reset_api();
//int pqos_wrapper_alloc_release(const unsigned *core_array, unsigned int core_amount_in_array);
//int pqos_wrapper_alloc_assign(const unsigned *core_array, unsigned int core_amount_in_array, unsigned *class_id);
//int pqos_wrapper_set_mba_for_common_cos(unsigned classID, int mbaMode, const unsigned *mbaMax, const unsigned *socketsToSetArray, int numOfSockets);
//int pqos_wrapper_alloc_l3cache(unsigned classID, const unsigned *waysMask, const unsigned *socketsToSet, int numOfSockets);
//int pqos_wrapper_assoc_core(const unsigned *classIDs, const unsigned *cores, int numOfCores);
//int pqos_wrapper_assoc_pid(const unsigned *classIDs, const unsigned *tasks, int numOfTasks);
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

// MbaStruct contains values needed to set MBA values
// amount of elements in MbaMaxes must be equal amount of SocketsToSet
type MbaStruct struct {
	ClassID      int   //class of service (COS#)
	MbaMode      int   // 0 for percentage mode or 1 for MBps mode
	MbaMaxes     []int // mba values to set on a specified sockets
	SocketsToSet []int // sockets to set for common ClassID (COS#)
}

// L3CacheStruct contains values needed to set L3 values
// amount of elements in WaysMask must be equal amount of SocketsToSet
type L3CacheStruct struct {
	ClassID      int      // class of service (COS#)
	WaysMask     []uint64 // bit mask for L3 cache ways for all specified sockets
	SocketsToSet []int    // sockets to set
}

// AssocCoresStruct contains values needed to assoc cores
// amount of elements in Cores must be equal elements amount in ClassIDs
type AssocCoresStruct struct {
	ClassIDs []int // classes of service (COS#)
	Cores    []int // cores to associate
}

// AssocTasksStruct contains values needed to assoc pid/task
// amount of elements in Tasks must be equal elements amount in ClassIDs
type AssocTasksStruct struct {
	ClassIDs []int // classes of service (COS#)
	Tasks    []int // tasks to associate
}

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
	log.Debugf("AllocateCLOS: %x %x    pid: %v  numPid: %v  core: %v", s[0], s[1], pid, numPid, core)

	// set MBA if specified in res
	if len(res.MbaSchemata["MB"]) > 0 {
		log.Debugf("MBA params detected - need to set MBA")

		mbaMode, err := CheckMBA()
		if err != nil {
			log.Debugf("Failed to check MBA mode")
			return
		}

		var mbaToSet MbaStruct
		mbaToSet.ClassID = clos
		mbaToSet.MbaMode = mbaMode
		mbaToSet.MbaMaxes = []int{}
		mbaToSet.SocketsToSet = []int{}

		for _, elem := range res.MbaSchemata["MB"] {
			log.Debugf("Want MBA = %v on a socket %v", elem.Mba, elem.ID)
			mbaToSet.SocketsToSet = append(mbaToSet.SocketsToSet, int(elem.ID))
			mbaToSet.MbaMaxes = append(mbaToSet.MbaMaxes, int(elem.Mba))
		}

		log.Debugf("MBA struct to set: %v", mbaToSet)
		err = SetMbaForSingleCos(mbaToSet)
		if err != nil {
			log.Errorf("Failed to set MBA for cos%v. Reason: %v", clos, err)
		}

	}
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

// CheckMBA checks if MBA is supported
// PQOS must be initialized before using this function to avoid error
// mbaMode values:
//      -1 means that MBA is not supported
//       0 means that MBA is enabled in percentage mode
//       1 means that MBA is enabled in MBps mode
func CheckMBA() (mbaMode int, err error) {
	var mbaModeAsCInt C.int
	result := C.pqos_wrapper_check_mba_support((*C.int)(&mbaModeAsCInt))
	mbaMode = int(mbaModeAsCInt)

	if result != 0 {
		return mbaMode, errors.New("Failed to check MBA mode")
	}

	switch mbaMode {
	case -1:
		log.Debugf("MBA mode not supported\n")
	case 0:
		log.Debugf("MBA percentage mode enabled\n")
	case 1:
		log.Debugf("MBA in MBps mode enabled\n")
	}

	return mbaMode, nil
}

/*Finish shuts down PQoS module*/
func Finish() error {
	result := C.pqos_wrapper_finish()

	if result != 0 {
		return errors.New("Failed to shut down PQoS module")
	}
	return nil
}

/*ResetAPI resets configuration of allocation technologies*/
func ResetAPI() error {
	result := C.pqos_wrapper_reset_api()
	if result != 0 {
		return errors.New("Failed to reset PQoS configuration of allocation technologies")
	}
	return nil
}

/*ReleaseAllocatedCores reassign cores in coreArray to default COS#0
  please be aware that function will not reset COS params to default values
  because releasing core from COS is enough
  * [in] coreArray    list of core ids
  * [in] numOfCores   number of core ids in the core_array
*/
func ReleaseAllocatedCores(coreArray []int) error {

	numOfCores := len(coreArray)
	coreArrayAsUInts := make([]C.uint, 0, numOfCores)
	for _, s := range coreArray {
		coreArrayAsUInts = append(coreArrayAsUInts, C.uint(s))
	}

	result := C.pqos_wrapper_alloc_release(&(coreArrayAsUInts[0]), (C.uint)(numOfCores))
	if result != 0 {
		return errors.New("Failed to reassign tasks in task_array to default COS#0")
	}

	return nil
}

/*AllocAssign assigns first available COS to cores in core_array
  coreArray   list of core ids
  numOfCores  number of core ids in the core_array
  classID     place to store reserved COS id
  err         operation status (nil or error)
*/
func AllocAssign(coreArray []int) (classID int, err error) {

	numOfCores := len(coreArray)
	coreArrayAsUInts := make([]C.uint, 0, numOfCores)
	for _, s := range coreArray {
		coreArrayAsUInts = append(coreArrayAsUInts, C.uint(s))
	}
	var classIDAsCInt C.uint

	result := C.pqos_wrapper_alloc_assign(&(coreArrayAsUInts[0]), (C.uint)(numOfCores), &classIDAsCInt)
	classID = int(classIDAsCInt)

	if result != 0 {
		return classID, errors.New("Failed to assigns first available COS to specified cores")
	}
	return classID, nil
}

// SetMbaForSingleCos sets classes of service defined by mba on mba id for common COS#
// mbaValuesToSet - contains values needed to set MBA values
// Function returns operation status (nil or error)
func SetMbaForSingleCos(mbaValuesToSet MbaStruct) error {

	if len(mbaValuesToSet.MbaMaxes) != len(mbaValuesToSet.SocketsToSet) {
		return errors.New("Amount of elements in MbaMaxes must be equal amount of sockets")
	}

	numOfElements := len(mbaValuesToSet.SocketsToSet)
	mbaMaxesAsUInts := make([]C.uint, 0, numOfElements)
	for _, s := range mbaValuesToSet.MbaMaxes {
		mbaMaxesAsUInts = append(mbaMaxesAsUInts, C.uint(s))
	}

	socketsToSetAsUInts := make([]C.uint, 0, numOfElements)
	for _, s := range mbaValuesToSet.SocketsToSet {
		socketsToSetAsUInts = append(socketsToSetAsUInts, C.uint(s))
	}
	result := C.pqos_wrapper_set_mba_for_common_cos(C.uint(mbaValuesToSet.ClassID), C.int(mbaValuesToSet.MbaMode), &(mbaMaxesAsUInts[0]), &(socketsToSetAsUInts[0]), C.int(numOfElements))

	if result != 0 {
		return errors.New("Failed to set MBA for common COS")
	}
	return nil
}

// AllocL3Cache allocates L3 Cache for common COS#
// l3ValuesToSet - contains values needed to set L3 values
// Function returns operation status (nil or error)
func AllocL3Cache(l3ValuesToSet L3CacheStruct) error {

	if len(l3ValuesToSet.WaysMask) != len(l3ValuesToSet.SocketsToSet) {
		return errors.New("Amount of elements in WaysMask must be equal amount of sockets")
	}
	numOfElements := len(l3ValuesToSet.SocketsToSet)

	waysMaskAsUInts := make([]C.uint, 0, numOfElements)
	for _, s := range l3ValuesToSet.WaysMask {
		waysMaskAsUInts = append(waysMaskAsUInts, C.uint(s))
	}

	socketsToSetAsUInts := make([]C.uint, 0, numOfElements)
	for _, s := range l3ValuesToSet.SocketsToSet {
		socketsToSetAsUInts = append(socketsToSetAsUInts, C.uint(s))
	}

	result := C.pqos_wrapper_alloc_l3cache(C.uint(l3ValuesToSet.ClassID), &(waysMaskAsUInts[0]), &(socketsToSetAsUInts[0]), C.int(numOfElements))

	if result != 0 {
		return errors.New("Failed to set L3 for common COS")
	}
	return nil
}

// AssocCore associates core
// coresStruct - contains values needed to associate core
// Function returns operation status (nil or error)
func AssocCore(coresStruct AssocCoresStruct) error {

	if len(coresStruct.ClassIDs) != len(coresStruct.Cores) {
		return errors.New("Amount of elements in Cores must be equal amount of ClassIDs")
	}
	numOfElements := len(coresStruct.Cores)

	classIDsAsUInts := make([]C.uint, 0, numOfElements)
	for _, s := range coresStruct.ClassIDs {
		classIDsAsUInts = append(classIDsAsUInts, C.uint(s))
	}

	coresAsUInts := make([]C.uint, 0, numOfElements)
	for _, s := range coresStruct.Cores {
		coresAsUInts = append(coresAsUInts, C.uint(s))
	}

	result := C.pqos_wrapper_assoc_core(&(classIDsAsUInts[0]), &(coresAsUInts[0]), C.int(numOfElements))

	if result != 0 {
		return errors.New("Failed to associate specified cores with given class IDs")
	}
	return nil
}

// AssocTask associates pid/task
// tasksStruct - contains values needed to associate task/pid
// Function returns operation status (nil or error)
func AssocTask(tasksStruct AssocTasksStruct) error {

	if len(tasksStruct.ClassIDs) != len(tasksStruct.Tasks) {
		return errors.New("Amount of elements in Tasks must be equal amount of ClassIDs")
	}
	numOfElements := len(tasksStruct.Tasks)

	classIDsAsUInts := make([]C.uint, 0)
	for _, s := range tasksStruct.ClassIDs {
		classIDsAsUInts = append(classIDsAsUInts, C.uint(s))
	}

	tasksAsUInts := make([]C.uint, 0)
	for _, s := range tasksStruct.Tasks {
		tasksAsUInts = append(tasksAsUInts, C.uint(s))
	}

	result := C.pqos_wrapper_assoc_pid(&(classIDsAsUInts[0]), &(tasksAsUInts[0]), C.int(numOfElements))

	if result != 0 {
		return errors.New("Failed to associate specified cores with given class IDs")
	}
	return nil
}

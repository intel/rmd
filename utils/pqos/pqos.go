package pqos

//#cgo CFLAGS: -W -Wall -Wextra -Wstrict-prototypes -Wmissing-prototypes -Wmissing-declarations -Wold-style-definition -Wpointer-arith -Wcast-qual -Wundef -Wwrite-strings -Wformat -Wformat-security -fstack-protector -fPIE -D_FORTIFY_SOURCE=2 -Wunreachable-code -Wsign-compare -Wno-endif-labels -g -O2
//#cgo LDFLAGS: -lpqos -lpthread
//int pqos_wrapper_init(void);
//int pqos_wrapper_check_mba_support(int *mbaMode);
//int pqos_wrapper_finish(void);
//int pqos_wrapper_reset_api(int mba_mode);
//int pqos_wrapper_alloc_release(const unsigned *core_array, unsigned int core_amount_in_array);
//int pqos_wrapper_alloc_assign(const unsigned *core_array, unsigned int core_amount_in_array, unsigned *class_id);
//int pqos_wrapper_set_mba_for_common_cos(unsigned classID, int mbaMode, const unsigned *mbaMax, const unsigned *socketsToSetArray, int numOfSockets);
//int pqos_wrapper_alloc_l3cache(unsigned classID, const unsigned *waysMask, const unsigned *socketsToSet, int numOfSockets);
//int pqos_wrapper_assoc_core(unsigned classID, const unsigned *cores, int numOfCores);
//int pqos_wrapper_assoc_pid(unsigned classID, const unsigned *tasks, int numOfTasks);
//int pqos_wrapper_get_clos_num(int *l3ca_clos_num, int *mba_clos_num);
//int pqos_wrapper_get_num_of_sockets(int *numOfSockets);
//int pqos_wrapper_get_num_of_cacheways(int *numOfCacheways);
import "C"
import (
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/intel/rmd/utils/resctrl"
	log "github.com/sirupsen/logrus"
)

//constants for setting MBA to default values
const (
	// from documentation math.MaxUint32 BUT there is rounding bug in PQoS
	// so we must subtract 10 (default step for MBA) to avoid overflow
	defaultMBpsMBAValue       = math.MaxUint32 - 10
	defaultPercentageMBAValue = 100
)

var (
	availableCLOSes []string
	usedCLOSes      []string
	reservedCLOSes  []string
	sharedCLOS      string
	numOfSockets    int // amount of current machine's sockets
	numOfCacheways  int // amount of current machine's cache ways
)

// MBAMode describes selected MBA (Memory Bandwidth Allocation) mode
type MBAMode int

const (
	// MBANone means MBA not used by RMD so not needed
	MBANone MBAMode = iota
	// MBAPercentage means MBA in percentage (default) mode is used
	MBAPercentage
	// MBAMbps means MBA in Mbps mode is used
	MBAMbps
)

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

// AssocCoresStruct contains values needed to assoc cores for common ClassID
type AssocCoresStruct struct {
	ClassID int   // common class of service (COS#)
	Cores   []int // cores to associate
}

// AssocTasksStruct contains values needed to assoc pid/task for common ClassID
type AssocTasksStruct struct {
	ClassID int   // common class of service (COS#)
	Tasks   []int // tasks to associate
}

// names of COSes reserved for special purposes
// OSGroup is defined as "." instead of "COS0" as some functions are based on filesystem path
const (
	// OSGroupCOS
	OSGroupCOS = "."
	// InfraGoupCOS
	InfraGoupCOS = "COS1"
)

// InitCLOSPool initializes pool of available CLOSes based on number of CLOSes supported by the platform
func InitCLOSPool() error {
	// WARNING: Cannot use PQOS function here as this code is launched in user process (in workload initialization)
	// TODO In future get only necessary number of CLOSes (ex. only L3 or only MBA)
	numOfClos, err := resctrl.GetNumOfCLOS(true, true)
	if err != nil {
		return fmt.Errorf("Error when fetching number of CLOSes: %v", err.Error())
	}
	reservedCLOSes = []string{OSGroupCOS, InfraGoupCOS}
	// lists of available and used CLOSes will never have all platform CLOSes as COS0 and COS1 are reserved
	availableCLOSes = make([]string, 0, numOfClos-2)
	usedCLOSes = make([]string, 0, numOfClos-2)
	for index := 2; index < numOfClos; index++ {
		availableCLOSes = append(availableCLOSes, "COS"+strconv.Itoa(index))
	}
	log.Debugf("Available CLOSes %v", availableCLOSes)
	return nil
}

// GetAvailableCLOSes returns list of CLOSes (copy of original) available for use
func GetAvailableCLOSes() []string {
	result := make([]string, len(availableCLOSes))
	copy(result, availableCLOSes)
	return result
}

// GetUsedCLOSes returns list of CLOSes already in use (except reserved CLOSes)
func GetUsedCLOSes() []string {
	result := make([]string, len(usedCLOSes))
	copy(result, usedCLOSes)
	return result
}

// AllocateCLOS ...
func AllocateCLOS(res *resctrl.ResAssociation, name string) {
	var clos int
	// Supported names are only COSx where x is a non-negative value
	if strings.HasPrefix(name, "COS") && len(name) > 3 {
		id, err := strconv.Atoi(name[3:])
		if err != nil || id < 0 {
			log.Errorf("Invalid CLOS number (%v) AllocateCLOS()", name)
			return
		}
		clos = id
	} else {
		// OSGroupCOS is internally marked by path (so by ".")
		if name == OSGroupCOS {
			clos = 0
		} else {
			log.Errorf("Invalid CLOS name (%v) given to AllocateCLOS()", name)
			return
		}
	}

	var s = []uint64{0, 0}
	s[res.CacheSchemata["L3"][0].ID], _ = strconv.ParseUint(res.CacheSchemata["L3"][0].Mask, 16, 64)
	s[res.CacheSchemata["L3"][1].ID], _ = strconv.ParseUint(res.CacheSchemata["L3"][1].Mask, 16, 64)
	log.Debugf("AllocateCLOS: %x %x   clos: %v", s[0], s[1], clos)

	var cacheToSet L3CacheStruct
	cacheToSet.ClassID = clos
	cacheToSet.WaysMask = s

	socketsToSet := []int{}
	for i := 0; i < GetNumOfSockets(); i++ {
		// TODO PQOS Add here handling for WaysMask
		socketsToSet = append(socketsToSet, i)
	}
	cacheToSet.SocketsToSet = socketsToSet

	err := AllocL3Cache(cacheToSet)
	if err != nil {
		log.Errorf("Failed to allocate L3 cache. Reason: %v", err)
		return
	}

	// don't need to invoke AssocTask/AssocCore code for COS#0
	if clos == 0 {
		return
	}

	tasksAmount := len(res.Tasks)

	if tasksAmount > 0 {
		log.Debugf("Tasks association will be performed")

		tasksAsInts := make([]int, 0, tasksAmount)
		for _, singlePidAsString := range res.Tasks {

			pidAsInt, err := strconv.Atoi(singlePidAsString)
			if err != nil {
				log.Errorf("Failed convert pid string format into integer. Reason: %v", err)
				return
			}
			tasksAsInts = append(tasksAsInts, pidAsInt)
		}

		log.Debugf("Tasks as integers: %v", tasksAsInts)

		var tasksToAssoc AssocTasksStruct
		tasksToAssoc.ClassID = clos
		tasksToAssoc.Tasks = tasksAsInts
		err = AssocTask(tasksToAssoc)
		if err != nil {
			log.Errorf("Failed to associate tasks. Reason: %v", err)
			return
		}
	} else {
		log.Debugf("Cores association will be performed")

		// need to convert core bitmask into core integer array
		cores, err := CoreMaskToSlice(res.CPUs)
		if err != nil {
			log.Errorf("Failed to parse core bitmask string into core integer array.Reason: %v", err)
			return
		}

		log.Debugf("Cores in array form: %v", cores)

		var coresToAssoc AssocCoresStruct
		coresToAssoc.ClassID = clos
		coresToAssoc.Cores = cores
		err = AssocCore(coresToAssoc)
		if err != nil {
			log.Errorf("Failed to associate cores. Reason: %v", err)
			return
		}
	}

	// set MBA if specified in res
	if len(res.MbaSchemata["MB"]) > 0 {
		log.Debugf("MBA params detected - need to set MBA")

		mbaMode, err := CheckMBA()
		if err != nil {
			log.Errorf("Failed to check MBA mode")
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

// DeallocateTasks ...
func DeallocateTasks(tasks []string) error {
	if len(tasks) == 0 {
		return errors.New("Empty task list given to DeallocateTasks")
	}
	// avoid printing too much
	if len(tasks) > 20 {
		log.Debugf("DeallocateTasks called for: %v (... and more)", tasks[:20])
	} else {
		log.Debugf("DeallocateTasks called for: %v", tasks)
	}

	taskCount := len(tasks)

	cTasks := make([]C.uint, 0, taskCount)
	for _, tidStr := range tasks {
		// assuming that task id strings have already been validated (not validating again)
		tidNum, _ := strconv.Atoi(tidStr)
		cTasks = append(cTasks, C.uint(tidNum))
	}

	result := C.pqos_wrapper_assoc_pid(C.uint(0), &(cTasks[0]), C.int(taskCount))

	if result != 0 {
		return errors.New("Failed to re-associate given tasks to CLOS 0")
	}

	return nil
}

// DeallocateCores ...
func DeallocateCores(cores []string) error {
	if len(cores) == 0 {
		return errors.New("Empty task list given to DeallocateCores")
	}
	// avoid printing too much
	if len(cores) > 20 {
		log.Debugf("DeallocateCores called for: %v (... and more)", cores[:20])
	} else {
		log.Debugf("DeallocateCores called for: %v", cores)
	}

	coreCount := len(cores)

	cCores := make([]C.uint, 0, coreCount)
	for _, tidStr := range cores {
		// assuming that task id strings have already been validated (not validating again)
		tidNum, _ := strconv.Atoi(tidStr)
		cCores = append(cCores, C.uint(tidNum))
	}

	result := C.pqos_wrapper_assoc_core(C.uint(0), &(cCores[0]), C.int(coreCount))

	if result != 0 {
		return errors.New("Failed to re-associate given cores to CLOS 0")
	}

	return nil
}

// Init ...
func Init() error {
	result := C.pqos_wrapper_init()
	if result != 0 {
		// pqos_wrapper_init() returns non-zero value in case of initialization failure
		return errors.New("Failed to initialize PQOS driver")
	}

	// get number of sockets and save as global variable
	var currentNumOfSockets C.int
	result = C.pqos_wrapper_get_num_of_sockets(&currentNumOfSockets)
	if result != 0 {
		return errors.New("Failed to get sockets amount for current machine")
	}
	numOfSockets = int(currentNumOfSockets)

	// get number of cache ways and save as global variable
	var currentNumOfCacheways C.int
	result = C.pqos_wrapper_get_num_of_cacheways(&currentNumOfCacheways)
	if result != 0 {
		return errors.New("Failed to get cache ways amount for current machine")
	}
	numOfCacheways = int(currentNumOfCacheways)

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
	result := C.pqos_wrapper_check_mba_support(&mbaModeAsCInt)
	mbaMode = int(mbaModeAsCInt)

	if result != 0 {
		return mbaMode, errors.New("Failed to check MBA mode")
	}

	switch mbaMode {
	case -1:
		log.Debugf("MBA mode not supported")
	case 0:
		log.Debugf("MBA percentage mode enabled")
	case 1:
		log.Debugf("MBA in MBps mode enabled")
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
func ResetAPI(mode MBAMode) error {
	result := C.pqos_wrapper_reset_api(C.int(mode))
	if result != 0 {
		return fmt.Errorf("Failed to reset PQoS library with error: %v", result)
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
	result := C.pqos_wrapper_set_mba_for_common_cos(C.uint(mbaValuesToSet.ClassID),
		C.int(mbaValuesToSet.MbaMode), &(mbaMaxesAsUInts[0]), &(socketsToSetAsUInts[0]), C.int(numOfElements))

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
	log.Debugf("Want to set wayMask: %x on sockets: %v", waysMaskAsUInts, socketsToSetAsUInts)
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

	numOfElements := len(coresStruct.Cores)

	coresAsUInts := make([]C.uint, 0, numOfElements)
	for _, s := range coresStruct.Cores {
		coresAsUInts = append(coresAsUInts, C.uint(s))
	}

	result := C.pqos_wrapper_assoc_core(C.uint(coresStruct.ClassID), &(coresAsUInts[0]), C.int(numOfElements))

	if result != 0 {
		return errors.New("Failed to associate specified cores with given class IDs")
	}
	return nil
}

// AssocTask associates pid/task
// tasksStruct - contains values needed to associate task/pid
// Function returns operation status (nil or error)
func AssocTask(tasksStruct AssocTasksStruct) error {

	numOfElements := len(tasksStruct.Tasks)

	tasksAsUInts := make([]C.uint, 0)
	for _, s := range tasksStruct.Tasks {
		tasksAsUInts = append(tasksAsUInts, C.uint(s))
	}

	result := C.pqos_wrapper_assoc_pid(C.uint(tasksStruct.ClassID), &(tasksAsUInts[0]), C.int(numOfElements))

	if result != 0 {
		return errors.New("Failed to associate specified cores with given class IDs")
	}
	return nil
}

// GetNumOfCLOSes returns number of Class Of Services supported by platform
// Function params (cache and mba) allows to select whether function shall check
// number of CLOSes for L3 Cache Allocation, Memory Bandwidth Allocation or both
// (minimum of two values are then returned)
func GetNumOfCLOSes(cache, mba bool) (int, error) {
	var l3ClosNum, mbaClosNum C.int

	if !cache && !mba {
		return 0, errors.New("Invalid selection of capabilities for fetching number of CLOSes")
	}

	result := C.pqos_wrapper_get_clos_num(&l3ClosNum, &mbaClosNum)

	userID := os.Geteuid()
	if result != 0 {
		return 0, fmt.Errorf("Error when checking number of CLOSes in platform as user: %v", userID)
	}

	if cache && !mba {
		return int(l3ClosNum), nil
	}

	if !cache && mba {
		return int(mbaClosNum), nil
	}

	// both flags defined so return lower value
	if l3ClosNum < mbaClosNum {
		return int(l3ClosNum), nil
	}
	return int(mbaClosNum), nil
}

// UseAvailableCLOS takes one CLOS name from list of available ones, moves it to used CLOSes and returns it's name
// If no available CLOS found then empty name and non-nil error is returned
func UseAvailableCLOS() (string, error) {
	// TODO In future add locks for thread safety (in some rare situtations race condition can appear)
	if len(availableCLOSes) < 1 {
		return "", errors.New("No free CLOS available")
	}
	result := availableCLOSes[0]
	availableCLOSes = availableCLOSes[1:] // remove 1st element
	log.Debugf("Used CLOS %v Available CLOS: %v", result, availableCLOSes)
	usedCLOSes = append(usedCLOSes, result)
	return result, nil
}

// ReturnClos moves CLOS with given name from used list into available list
func ReturnClos(name string) error {
	// TODO In future add locks for thread safety (in some rare situtations race condition can appear)
	// perform some basic checks for easier debugging
	if name == "" || !strings.HasPrefix(name, "COS") {
		return fmt.Errorf("Invalid clos name given: '%v'", name)
	}
	index := 0
	found := false
	for index < len(usedCLOSes) {
		if usedCLOSes[index] == name {
			found = true
			break
		}
		index++
	}
	if !found {
		return fmt.Errorf("CLOS '%v' not found on list of used CLOSes", name)
	}
	// return CLOS name to available list ...
	availableCLOSes = append(availableCLOSes, name)
	// ... and remove it from used list
	switch index {
	case 0:
		// remove first item
		usedCLOSes = usedCLOSes[1:]
	case len(usedCLOSes) - 1:
		// remove last item
		usedCLOSes = usedCLOSes[:len(usedCLOSes)-1]
	default:
		// remove some middle item
		usedCLOSes = append(usedCLOSes[:index], usedCLOSes[index+1:]...)
	}
	log.Debugf("CLOS '%v' removed from list of used CLOSes", name)
	return nil
}

// GetNumberOfFreeCLOSes returns number of CLOS that are available to use
func GetNumberOfFreeCLOSes() int {
	return len(availableCLOSes)
}

// Shared COS (shared group) is one for all workloads with "shared" type
// This group should be created when needed first time and re-used after that

// GetSharedCLOS returns name assigned to shared workloads COS.
// If shared COS have not been reserved yet function tries to reserve it
func GetSharedCLOS() (string, error) {
	// TODO In future add locks for thread safety (in some rare situtations race condition can appear)
	if sharedCLOS == "" {
		// have to reserve new COS
		newShared, err := UseAvailableCLOS()
		if err != nil {
			return "", errors.New("No free CLOS left to create shared group")
		}
		sharedCLOS = newShared
	}
	return sharedCLOS, nil
}

// MarkCLOSasUsed moves given CLOS name from available CLOSes to used CLOSes
// Used if for some reason (ex. check workload initialization) specific CLOS
// is in use but not fetched using pqos package API
func MarkCLOSasUsed(name string) error {
	if !strings.HasPrefix(name, "COS") {
		return errors.New("Invalid CLOS name given ")
	}

	index := -1
	for idx, val := range availableCLOSes {
		if val == name {
			index = idx
			break
		}
	}

	if index == -1 {
		// CLOS not found
		return fmt.Errorf("CLOS %v not found on available list", name)
	}

	// return CLOS name to available list ...
	usedCLOSes = append(usedCLOSes, name)
	// ... and remove it from used list
	switch index {
	case 0:
		// remove first item
		availableCLOSes = availableCLOSes[1:]
	case len(availableCLOSes) - 1:
		// remove last item
		availableCLOSes = availableCLOSes[:len(availableCLOSes)-1]
	default:
		// remove some middle item
		availableCLOSes = append(availableCLOSes[:index], availableCLOSes[index+1:]...)
	}
	log.Debugf("%v moved to used CLOSes", name)

	return nil
}

// CoreMaskToSlice converts bitmasks into cores array
func CoreMaskToSlice(mask string) ([]int, error) {
	if len(mask) == 0 {
		return []int{}, errors.New("Empty mask")
	}
	maskParts := strings.Split(mask, ",")

	uintLen := 32
	maskPartBase := 0
	if len(maskParts[len(maskParts)-1]) > 8 {
		uintLen = 64
	}

	coreList := make([]int, 0, 16) // for most cases there will be no more than 16 cores in workload
	// process mask parts from last one so from lowest core numbers
	for index := len(maskParts) - 1; index >= 0; index-- {
		maskPartAsUint, err := strconv.ParseUint(maskParts[index], 16, uintLen)
		if err != nil {
			return []int{}, fmt.Errorf("String parsing error: %v", err.Error())
		}
		for shiftNum := 0; shiftNum < uintLen; shiftNum++ {
			if (maskPartAsUint & 0x0001) == 1 {
				coreList = append(coreList, shiftNum+maskPartBase)
			}
			maskPartAsUint >>= 1
		}
		maskPartBase += uintLen

	}
	return coreList, nil
}

// resetMBAToDefaults resets MBA to default values for specified COS#
func resetMBAToDefaults(cos int) error {

	log.Debugf("Setting MBA to default values for COS%v", cos)

	mbaMode, err := CheckMBA()
	if err != nil {
		log.Error("Failed to check MBA mode.")
		return err
	}

	var mbaToSet MbaStruct
	mbaToSet.ClassID = cos
	mbaToSet.MbaMode = mbaMode

	valuesForMBA := []int{}
	socketsForMBA := []int{}
	defaultModeValueToSet := defaultPercentageMBAValue
	if mbaMode != 0 {
		defaultModeValueToSet = defaultMBpsMBAValue
	}

	for i := 0; i < GetNumOfSockets(); i++ {
		valuesForMBA = append(valuesForMBA, defaultModeValueToSet)
		socketsForMBA = append(socketsForMBA, i)
	}

	mbaToSet.MbaMaxes = valuesForMBA
	mbaToSet.SocketsToSet = socketsForMBA

	err = SetMbaForSingleCos(mbaToSet)
	if err != nil {
		log.Errorf("Failed to set reset MBA for cos%v. Reason: %v", cos, err)
	}

	return nil
}

// GetNumOfSockets returns amount of sockets for current machine
// which is a global variable for PQoS package
func GetNumOfSockets() int {
	return numOfSockets
}

// GetNumOfCacheways returns amount of cacheways for current machine
// which is a global variable for PQoS package
func GetNumOfCacheways() int {
	return numOfCacheways
}

// resetL3CacheToDefaults resets L3 cache to default values for specified COS#
func resetL3CacheToDefaults(cos int) error {

	log.Debugf("Setting L3 cache to default values for COS%d", cos)

	var cacheToSet L3CacheStruct
	cacheToSet.ClassID = cos

	valuesForCache := []uint64{}
	socketsForCache := []int{}
	defaultCacheValue := uint64(1<<GetNumOfCacheways() - 1) //prepare default mask to set where -1 to avoid 00000001 when 0 provided

	for i := 0; i < GetNumOfSockets(); i++ {
		valuesForCache = append(valuesForCache, defaultCacheValue)
		socketsForCache = append(socketsForCache, i)
	}
	cacheToSet.WaysMask = valuesForCache
	cacheToSet.SocketsToSet = socketsForCache

	err := AllocL3Cache(cacheToSet)
	if err != nil {
		log.Errorf("Failed to reset L3 cache to defaults. Reason: %v", err)
		return err
	}

	return nil
}

// ResetCOSParamsToDefaults resets L3 cache and MBA to default values for specified COS#
func ResetCOSParamsToDefaults(cosName string) error {
	log.Debugf("Setting L3 cache to default values for %v", cosName)
	//splits "COS#"" into "COS" and "#"
	cosSlice := strings.SplitAfter(cosName, "COS")
	cosAsInt, err := strconv.Atoi(cosSlice[1])
	if err != nil {
		log.Errorf("Failed to convert COS from string to int")
		return err
	}

	err = resetL3CacheToDefaults(cosAsInt)
	if err != nil {
		log.Errorf("%v", err)
		return err
	}

	err = resetMBAToDefaults(cosAsInt)
	if err != nil {
		log.Errorf("%v", err)
		return err
	}

	return nil
}

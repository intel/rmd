// +build linux

package resctrl

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	libutil "github.com/intel/rmd/utils/bitmap"
	appConf "github.com/intel/rmd/utils/config"
)

// internal package variables
var (
	intelRdtRootLock sync.Mutex
	intelRdtRoot     string
)

// variables that can be accessed from the outside of package
var (
	// RdtInfo is global immutable variable
	RdtInfo *map[string]*RdtCosInfo
	// SysResctrl is the absolute path to the root of the Intel RDT "resource control" filesystem
	SysResctrl string
)

// Init does config initial
func Init() error {
	appconf := appConf.NewConfig()
	SysResctrl = appconf.Def.SysResctrl
	return nil
}

// NotFoundError represents not found error
type NotFoundError struct {
	ResourceControl string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("mountpoint for %s not found", e.ResourceControl)
}

// NewNotFoundError returns new error of NotFoundError
func NewNotFoundError(res string) error {
	return &NotFoundError{
		ResourceControl: res,
	}
}

// IsNotFound returns if notfound error happened
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*NotFoundError)
	return ok
}

func writeFile(dir, file, data string) error {
	if dir == "" {
		return fmt.Errorf("no such directory for %s", file)
	}
	if !strings.HasSuffix(data, "\n") {
		data = data + "\n"
	}
	if err := ioutil.WriteFile(filepath.Join(dir, file), []byte(data), 0644); err != nil {
		return fmt.Errorf("failed to write %v to %v: %v", data, file, err)
	}
	return nil
}

// DestroyResAssociation removes resource group
func DestroyResAssociation(group string) error {
	return os.RemoveAll(filepath.Join(SysResctrl, group))
}

// CacheCos is the COS of a cache
type CacheCos struct {
	ID   uint8
	Mask string
}

// MbaCos is the COS of a cache
type MbaCos struct {
	ID  uint8
	Mba uint32
}

// ResAssociation is the resource group in resctrl
// TODO  need to paser the tag for field setting.
type ResAssociation struct {
	Tasks         []string
	CPUs          string
	CacheSchemata map[string][]CacheCos
	MbaSchemata   map[string][]MbaCos
}

// NewResAssociation gives new empty ResAssociation
func NewResAssociation() *ResAssociation {
	ra := &ResAssociation{}
	ra.Tasks = []string{}
	ra.CacheSchemata = make(map[string][]CacheCos)
	ra.MbaSchemata = make(map[string][]MbaCos)
	return ra
}

// parserResAssociation does the parsing.
// Usage:
//    ress := make(map[string]*ResAssociation)
//	  filepath.Walk(SysResctrl, parserResAssociation(SysResctrl, ignore, ress))
func parserResAssociation(basepath string, ignore []string, ps map[string]*ResAssociation) filepath.WalkFunc {
	parser := func(res *ResAssociation, name string, val []byte) error {
		switch name {
		case "Cpus":
			str := strings.TrimSpace(string(val))
			libutil.SetField(res, "CPUs", str)
			return nil
		case "Schemata":
			strs := strings.Split(string(val), "\n")
			if len(strs) > 1 {
				res.CacheSchemata = make(map[string][]CacheCos)
			}
			for _, data := range strs {
				datas := strings.SplitN(data, ":", 2)
				key := datas[0]

				key = strings.Replace(key, " ", "", -1)

				if key == "" {
					return nil
				}
				if _, ok := res.CacheSchemata[key]; !ok {
					res.CacheSchemata[key] = make([]CacheCos, 0, 10)
				}

				coses := strings.Split(datas[1], ";")
				for _, cos := range coses {
					infos := strings.SplitN(cos, "=", 2)
					id, _ := strconv.ParseUint(infos[0], 10, 8)
					cacheCos := &CacheCos{uint8(id), infos[1]}
					res.CacheSchemata[key] = append(res.CacheSchemata[key], *cacheCos)
				}

			}
			return nil
		default:
			strs := strings.Split(string(val), "\n")
			// Notes, remove the last element, it's a empty string
			// It will cause error while write back tasks to resctrl
			libutil.SetField(res, name, strs[:len(strs)-1])
			return nil
		}
	}

	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// add log
			return nil
		}
		f := filepath.Base(path)
		rel, err := filepath.Rel(basepath, path)
		pkey := rel
		if info.IsDir() {
			// ignore dir.
			for _, d := range ignore {
				if d == f {
					return filepath.SkipDir
				}
			}
			ps[pkey] = &ResAssociation{}
			return nil
		}
		for _, d := range ignore {
			if d == f {
				return nil
			}
		}

		dir := filepath.Dir(path)
		rel, err = filepath.Rel(basepath, dir)
		pkey = rel

		name := strings.Replace(strings.Title(strings.Replace(f, "_", " ", -1)), " ", "", -1)
		data, err := ioutil.ReadFile(path)
		pl := ps[pkey]
		parser(pl, name, data)
		return nil
	}
}

// GetResAssociation returns all resource groups
// access the resctrl need flock to avoid race with other agent.
// Go does not support flock lib.
// That need cgo, please ref:
// https://gist.github.com/ericchiang/ce0fdcac5659d0a80b38
// now we can use lib/flock/flock.go
func GetResAssociation(availableCLOS []string) map[string]*ResAssociation {
	ignore := []string{"info", "mon_data", "mon_groups"}
	if availableCLOS != nil {
		ignore = append(ignore, availableCLOS...)
	}
	ress := make(map[string]*ResAssociation)
	filepath.Walk(SysResctrl, parserResAssociation(SysResctrl, ignore, ress))
	return ress
}

// RdtCosInfo is from /sys/fs/resctrl/info
type RdtCosInfo struct {
	CbmMask    string
	MinCbmBits int
	NumClosids int
}

/*
Usage:
    ignore := []string{"info"}  // ignore the toppath
	info := make(map[string]*RdtCosInfo)
    basepath := SysResctrl+"/info"
	filepath.Walk(basepath, ParserRdtCosInfo(basepath, ignore, info))
	fmt.Println(info["l3data"])  //for RDT, we can get info["l3data"]

*/
func parserRdtCosInfo(basepath string, ignore []string, mres map[string]*RdtCosInfo) filepath.WalkFunc {

	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// add log
			return nil
		}
		f := filepath.Base(path)
		rel, err := filepath.Rel(basepath, path)
		pkey := rel
		if info.IsDir() {
			for _, d := range ignore {
				if d == f {
					return nil
				}
			}
			mres[strings.ToLower(pkey)] = &RdtCosInfo{}
			return nil
		}
		for _, d := range ignore {
			if d == f {
				return nil
			}
		}

		dir := filepath.Dir(path)
		rel, err = filepath.Rel(basepath, dir)
		pkey = strings.ToLower(rel)

		name := strings.Replace(strings.Title(strings.Replace(f, "_", " ", -1)), " ", "", -1)
		data, err := ioutil.ReadFile(path)
		strs := strings.TrimSpace(string(data))
		res := mres[pkey]
		return libutil.SetField(res, name, strs)
	}
}

// GetRdtCosInfo gives RDT info
// access the resctrl need flock to avoid race with other agent.
// Go does not support flock lib.
// That need cgo, please ref:
// https://gist.github.com/ericchiang/ce0fdcac5659d0a80b38
// now we can use lib/flock/flock.go
func GetRdtCosInfo() map[string]*RdtCosInfo {
	if RdtInfo != nil {
		return *RdtInfo
	}
	ignore := []string{"info", "bit_usage", "shareable_bits"} // ignore the toppath
	info := make(map[string]*RdtCosInfo)
	basepath := SysResctrl + "/info"
	filepath.Walk(basepath, parserRdtCosInfo(basepath, ignore, info))
	RdtInfo = &info
	return *RdtInfo
}

// Return the mount point path of Intel RDT "resource control" filesysem
func findIntelRdtMountpointDir() (string, error) {
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return "", err
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		text := s.Text()
		fields := strings.Split(text, " ")
		// Safe as mountinfo encodes mountpoints with spaces as \040.
		index := strings.Index(text, " - ")
		postSeparatorFields := strings.Fields(text[index+3:])
		numPostFields := len(postSeparatorFields)

		// This is an error as we can't detect if the mount is for "Intel RDT"
		if numPostFields == 0 {
			return "", fmt.Errorf("Found no fields post '-' in %q", text)
		}

		if postSeparatorFields[0] == "resctrl" {
			// Check that the mount is properly formated.
			if numPostFields < 3 {
				return "", fmt.Errorf("Error found less than 3 fields post '-' in %q", text)
			}

			return fields[4], nil
		}
	}
	if err := s.Err(); err != nil {
		return "", err
	}

	return "", NewNotFoundError("Intel RDT")
}

// Gets the root path of Intel RDT "resource control" filesystem
func getIntelRdtRoot() (string, error) {
	intelRdtRootLock.Lock()
	defer intelRdtRootLock.Unlock()

	if intelRdtRoot != "" {
		return intelRdtRoot, nil
	}

	root, err := findIntelRdtMountpointDir()
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(root); err != nil {
		return "", err
	}

	intelRdtRoot = root
	return intelRdtRoot, nil
}

// IsIntelRdtMounted give true/false of RDT mounted or not
func IsIntelRdtMounted() bool {
	_, err := getIntelRdtRoot()
	if err != nil {
		if IsNotFound(err) {
			return false
		}
	}
	return true
}

// DisableRdt unmount resctrl
func DisableRdt() bool {
	if IsIntelRdtMounted() {
		if err := exec.Command("umount", "/sys/fs/resctrl").Run(); err != nil {
			return false
		}
	}
	return true
}

// EnableCat mounts resctrl
func EnableCat() bool {
	// mount -t resctrl resctrl /sys/fs/resctrl
	if err := os.MkdirAll("/sys/fs/resctrl", 0755); err != nil {
		return false
	}
	if err := exec.Command("mount", "-t", "resctrl", "resctrl", "/sys/fs/resctrl").Run(); err != nil {
		return false
	}
	return true
}

// EnableCdp mounts resctrl with option -o
func EnableCdp() bool {
	// mount -t resctrl -o cdp resctrl /sys/fs/resctrl
	if err := os.MkdirAll("/sys/fs/resctrl", 0755); err != nil {
		return false
	}
	if err := exec.Command("mount", "-t", "resctrl", "-o", "cdp", "resctrl", "/sys/fs/resctrl").Run(); err != nil {
		return false
	}
	return true
}

// RemoveTasks move tasks to default resctrl group
// Resctrl doesn't support remove tasks from sysfs, the way to remove tasks from
// resource group is to move them to default group
func RemoveTasks(tasks []string) error {
	var err error
	for _, v := range tasks {
		err = writeFile(SysResctrl, "tasks", v)
	}
	return err
}

// GetNumOfCLOS returns number COSes for L3 Cache, MBA or both depending on input params
// If both L3 Cache and MBA selected function returns lower of two values
func GetNumOfCLOS(getL3Clos, getMbaClos bool) (int, error) {

	if !getL3Clos && !getMbaClos {
		return 0, errors.New("Invalid input flags for GetNumOfCLOS")
	}

	var numL3Clos, numMbaClos int
	if getL3Clos {
		l3NumFile, err := os.Open(SysResctrl + "/info/L3/num_closids")
		if err != nil {
			return 0, err
		}
		defer l3NumFile.Close()

		s := bufio.NewScanner(l3NumFile)
		var text string
		for s.Scan() {
			text = s.Text()
		}

		numL3Clos, err = strconv.Atoi(text)
		if err != nil {
			return 0, err
		}
	}

	if getMbaClos {
		mbaNumFile, err := os.Open(SysResctrl + "/info/MB/num_closids")
		if err != nil {
			return 0, err
		}
		defer mbaNumFile.Close()

		s := bufio.NewScanner(mbaNumFile)
		var text string
		for s.Scan() {
			text = s.Text()
		}

		numMbaClos, err = strconv.Atoi(text)
		if err != nil {
			return 0, err
		}
	}

	// return only number of L3 CLOS
	if getL3Clos && !getMbaClos {
		return numL3Clos, nil
	}

	// return only number of MBA CLOS
	if getL3Clos && !getMbaClos {
		return numMbaClos, nil
	}

	// return lower of L3 CLOS/MBA CLOS
	if numL3Clos < numMbaClos {
		return numL3Clos, nil
	}

	return numMbaClos, nil
}

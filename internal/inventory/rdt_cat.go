package inventory

import (
	"bufio"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
)

const (
	kernelCmdPath    = "/proc/cmdline"
	kernelCmdPattern = "rdt=l3cat"
	cpuInfoPath      = "/proc/cpuinfo"
	cpuInfoPattern1  = "rdt_a"
	cpuInfoPattern2  = "cat_l3"
	fsInfoPath       = "/proc/filesystems"
	fsInfoPattern    = "resctrl"
)

var rdtOneGuard sync.Once
var rdtResult Capability

// CheckRDT checks if CAT_RDT support exist on host
func CheckRDT() Capability {
	rdtOneGuard.Do(func() {
		rdtResult = Capability{Name: "pstate", Available: false}
		// check platform
		rdtResult.Available = checkKernelCmd() && checkCPUFlag() && checkResctrlFs()
	})

	return rdtResult
}

func checkKernelCmd() bool {
	lineBuffer := make([]byte, 1024)
	// check kernel flag
	kernFile, err := os.Open(kernelCmdPath)
	if err != nil {
		log.Error("Failed to check kernel params: ", err.Error())
		return false
	}
	defer kernFile.Close()

	cnt, err := kernFile.Read(lineBuffer)
	if err != nil {
		log.Error("Failed to check kernel params: ", err.Error())
		return false
	}
	kernelCmdLine := string(lineBuffer[:cnt])
	if strings.Contains(kernelCmdLine, kernelCmdPattern) != true {
		log.Debug("Kernel command line does not contain: ", kernelCmdPattern)
		return false
	}

	return true
}

func checkCPUFlag() bool {
	// check CPU flag
	cpuFile, err := os.Open(cpuInfoPath)
	if err != nil {
		log.Error("Failed to check cpu flags: ", err.Error())
		return false
	}
	defer cpuFile.Close()

	scanner := bufio.NewScanner(cpuFile)
	for scanner.Scan() != false {
		line := scanner.Text()
		if strings.HasPrefix(line, "flags") {
			flags := strings.Split(line, ":")
			// should never happen but better check it
			if len(flags) != 2 {
				log.Error("Broken CPU flags")
				return false
			}
			if strings.Contains(line, cpuInfoPattern1) != true {
				log.Debug("Missing CPU flag: ", cpuInfoPattern1)
				return false
			}
			if strings.Contains(line, cpuInfoPattern2) != true {
				log.Debug("Missing CPU flag: ", cpuInfoPattern2)
				return false
			}
			// there can be multpiple cores and so multiple flags lines
			// but there's no need to analyze all of them
			break
		}
		// skip line processing if it's not the CPU flags line
	}

	return true
}

func checkResctrlFs() bool {
	// check resctrl filesystem
	fsData, err := ioutil.ReadFile(fsInfoPath)
	if err != nil {
		log.Error("Failed to read filesystems: ", err.Error())
		return false
	}
	fsString := string(fsData)

	if strings.Contains(fsString, fsInfoPattern) != true {
		log.Error("Filesystem not found: ", fsInfoPattern)
		return false
	}

	// no error till now -> all necessary items/flags found
	return true
}

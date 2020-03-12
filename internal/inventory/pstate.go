package inventory

import (
	"os"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
)

const scalingDriverPath = "/sys/devices/system/cpu/cpu0/cpufreq/scaling_driver"
const bufSize = 128

var supportedDrivers = []string{"acpi", "intel_pstate"}

var scalingOneGuard sync.Once
var scalingResult Capability

// CheckScaling checks if proper CPU freq. scaling driver is available
func CheckScaling() Capability {
	scalingOneGuard.Do(func() {
		scalingResult = Capability{Name: "pstate", Available: checkPstateAvailability()}
	})

	return scalingResult
}

func checkPstateAvailability() bool {
	if _, err := os.Stat(scalingDriverPath); os.IsNotExist(err) {
		// file does not exists - cannot check scaling driver
		log.Error("Unable to check scaling driver!")
		return false
	}

	input, err := os.Open(scalingDriverPath)
	if err != nil {
		log.Error("Unable to open scaling driver file!")
		return false
	}
	defer input.Close()

	buffer := make([]byte, bufSize)
	num, err := input.Read(buffer)
	if err != nil {
		log.Error("Unable to read scaling driver file!")
		return false
	}
	driver := strings.Trim(string(buffer[:num]), " \n")

	for _, supported := range supportedDrivers {
		if driver == supported {
			log.Debug("Supported driver detected: ", supported)
			return true
		}
	}

	log.Debug("Unsupported scaling driver: ", driver)
	return false
}

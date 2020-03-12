// +build openstack

package openstack

import (
	"errors"
	"io/ioutil"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/intel/rmd/internal/inventory"
	rdtconf "github.com/intel/rmd/modules/cache/config"
	yaml "gopkg.in/yaml.v2"
)

const (
	schemaVersion      = "1.0"
	llcSectionTitle    = "CUSTOM_LLC"
	pstateEnabledFlag  = "CUSTOM_P_STATE_ENABLED"
	pstateDisabledFlag = "CUSTOM_P_STATE_DISABLED"
	cacheWaysInfoPath  = "/sys/devices/system/cpu/cpu0/cache/index3/ways_of_associativity"
)

type meta struct {
	Version string `yaml:"schema_version"`
}

type customResourceDetails struct {
	Total           uint `yaml:"total"`
	Reserved        uint `yaml:"reserved"`
	MinUnit         uint `yaml:"min_unit"`
	MaxUnit         uint `yaml:"max_unit"`
	StepSize        uint `yaml:"step_size"`
	AllocationRatio uint `yaml:"allocation_ratio"`
}

type llcResource struct {
	CustomLLC customResourceDetails `yaml:"CUSTOM_LLC"`
}

type providerEntity struct {
	Identification struct {
		UUID string `yaml:"uuid"`
	} `yaml:"identification"`
	Inventories struct {
		Additional []llcResource
	} `yaml:"inventories"`
	Traits struct {
		Additional []string `yaml:"additional"`
	} `yaml:"traits"`
}

// GenerateFile Provider Config File generation according to scheme presented here:
// https://github.com/openstack/nova-specs/blob/master/specs/train/approved/provider-config-file.rst
func GenerateFile(filePath string) error {
	log.Debug("Generating Provider Config File for Nova")
	if len(filePath) == 0 {
		log.Error("Invalid (empty) file path")
		return errors.New("Invalid (empty) file path")
	}

	cacheAvailable := inventory.CheckRDT()
	pstateAvailabie := inventory.CheckScaling()

	providerData := make(map[string]interface{})

	// add schema information
	mt := meta{Version: schemaVersion}
	providerData["meta"] = mt

	// add providers
	var entity providerEntity
	var llcRes llcResource

	entity.Identification.UUID = "$COMPUTE_NODE"

	if cacheAvailable.Available {
		// read total number of cache ways
		buffer, err := ioutil.ReadFile(cacheWaysInfoPath)
		if err != nil {
			log.Error("Failed to read cache ways info: ", err.Error())
			return err
		}
		total, err := strconv.ParseInt(strings.Trim(string(buffer), " \n"), 10, 32)
		if err != nil {
			log.Error("Invalid number of cache ways in file: ", err.Error())
			return err
		}
		llcRes.CustomLLC.Total = uint(total) // number cache ways read from system

		// fetch RDTpool OS (reserved) config
		cfg := rdtconf.NewOSConfig()
		llcRes.CustomLLC.Reserved = cfg.CacheWays // OS group

		llcRes.CustomLLC.MinUnit = 1                                                  // const
		llcRes.CustomLLC.MaxUnit = llcRes.CustomLLC.Total - llcRes.CustomLLC.Reserved // FIXME, should be max of one socket (minimum of all available sockets)
		llcRes.CustomLLC.StepSize = 1                                                 // const
		llcRes.CustomLLC.AllocationRatio = 1                                          // const
	}
	entity.Inventories.Additional = []llcResource{llcRes}

	if pstateAvailabie.Available {
		// add P-State enabled flag
		entity.Traits.Additional = []string{pstateEnabledFlag}
	} else {
		// add P-State disabled flag
		entity.Traits.Additional = []string{pstateDisabledFlag}
	}

	providerData["providers"] = []providerEntity{entity}

	// convert providerData to yaml format
	buffer, err := yaml.Marshal(providerData)
	if err != nil {
		log.Error("Failed to prepare content for Provider Config File: ", err.Error())
	} else {
		err = ioutil.WriteFile(filePath, buffer, 0644)
		if err != nil {
			log.Error("Failed to write Provider Config File: ", err.Error())
		}
	}
	return err
}

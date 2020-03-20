package policy

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"

	"github.com/intel/rmd/utils"
	appConf "github.com/intel/rmd/utils/config"
	"github.com/intel/rmd/utils/cpu"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var supportedConfigType = map[string]int{
	"yaml": 1,
	"toml": 1,
}

// Param represents single policy attribute
// example: "MaxCache: 4"
//
type Param map[string]string

// Module represents attributes for single module
// example:
//     Cache:
//          Min: 1
//          Max: 1
//
type Module map[string]Param

// Policy represents policy type
// example:
//       gold:
//           Cache:
//                Min: 1
//                Max: 1
//           Pstate:
//                Ratio: 0.1
//
type Policy map[string]Module

// CPUArchitecture represents cpu model
// example:
//    broadwell:
//       gold:
//           Cache:
//                Min: 1
//                Max: 1
//           Pstate:
//                Ratio: 0.1
//
type CPUArchitecture map[string]Policy

var lock sync.Mutex

// GetDefaultPlatformPolicy gets policy for default platform
func GetDefaultPlatformPolicy() (Policy, error) {
	platform, err := LoadPolicyInfo()
	if err != nil {
		log.Errorf("Failed to load policy. Reason: %s", err.Error())
		return Policy{}, fmt.Errorf("Failed to load policy")
	}

	return platform, err
}

// LoadPolicyInfo returns policy for cpu on current machine (example: broadwell)
func LoadPolicyInfo() (Policy, error) {

	// 1. Read file with all defined policies
	rmdpolicy, err := loadPolicy()
	if err != nil {
		log.Errorf("Failed to load policies. Reason: %s", err.Error())
		return Policy{}, fmt.Errorf("Failed to load policies. Reason: %s", err.Error())
	}

	// 2. check architecure (need only stuff related with used CPU)
	cpu := strings.ToLower(cpu.GetMicroArch(cpu.GetSignature()))
	if cpu == "" {
		return Policy{}, fmt.Errorf("Unknown platform, please update the cpu_map.toml conf file")
	}

	// 3. Get main element of policy for your cpu
	platform, ok := rmdpolicy[cpu]
	if !ok {
		return Policy{}, fmt.Errorf("Error while get platform policy: %s", cpu)
	}

	return platform, nil
}

// loadPolicy loads pre-defined policies from configure file
func loadPolicy() (CPUArchitecture, error) {

	appconf := appConf.NewConfig()
	configFileExt := filepath.Ext(appconf.Def.PolicyPath)

	if !strings.HasPrefix(configFileExt, ".") {
		err := fmt.Errorf("Unknown policy file type extension %s", configFileExt)
		log.Errorf("error: %v", err)
		return nil, err
	}

	configType := strings.TrimPrefix(configFileExt, ".")

	if _, ok := supportedConfigType[configType]; !ok {
		err := fmt.Errorf("Unsupported policy file type extension %s", configType)
		log.Errorf("error: %v", err)
		return nil, err
	}

	isfile, err := util.IsRegularFile(appconf.Def.PolicyPath)
	if err != nil || !isfile {
		return nil, fmt.Errorf("Invalid policy file path %s", appconf.Def.PolicyPath)
	}

	text, err := ioutil.ReadFile(appconf.Def.PolicyPath)
	if err != nil { // Handle errors reading the config file
		err := fmt.Errorf("Error during config file reading: %s", err)
		log.Errorf("error: %v", err)
		return nil, err
	}

	parserViper := viper.New()
	if parserViper == nil {
		err := fmt.Errorf("Failed to create Viper parser - cannot continue")
		log.Errorf("error: %v", err)
		return nil, err
	}
	parserViper.SetConfigType(configType)
	err = parserViper.ReadConfig(bytes.NewBuffer(text)) // Find and read the config file
	if err != nil {
		err := fmt.Errorf("Fatal error config file: %s", err)
		log.Errorf("error: %v", err)
		return nil, err
	}

	result := CPUArchitecture{}
	err = parserViper.Unmarshal(&result)
	if err != nil {
		log.Errorf("error: %v", err)
	}

	return result, err
}

// GetDefaultPolicy returns default policy
func GetDefaultPolicy(policyName string) (Module, error) {
	lock.Lock()
	defer lock.Unlock()

	// we can't load policy only once and stored it due to project
	// requirements so each time we are loading policy info "on demand"
	platform, err := LoadPolicyInfo()

	if err != nil {
		log.Errorf("Cannot get default policy: %s", err.Error())
		return Module{}, err
	}

	// below example how "platform" can look
	// [map[
	//		gold:	[map[Cache:[map[Max:10 Min:10]] ]]
	//		silver: [map[Cache:[map[Max:8 Min:6]] ]]
	//		]
	//	]
	//

	// check if requested policy exists (for example: "gold")
	defaultPolicyParams, ok := platform[policyName]
	if !ok {
		return Module{}, fmt.Errorf("Can not find specified policy %s", policyName)
	}

	return defaultPolicyParams, nil
}

package policy

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	appConf "github.com/intel/rmd/app/config"
	"github.com/intel/rmd/lib/cpu"
)

// Attr is Policy attribute
type Attr map[string]string

// Policy is the pre-defiend policies
type Policy map[string][]Attr

// CATConfig represents for all policy config
type CATConfig struct {
	Catpolicy map[string][]Policy `yaml:"catpolicy"`
}

var supportedConfigType = map[string]int{
	"yaml": 1,
	"toml": 1,
}

var config *CATConfig
var lock sync.Mutex

// LoadPolicy loads pre-defined polies from configure file
func LoadPolicy() (*CATConfig, error) {
	appconf := appConf.NewConfig()
	configFileExt := filepath.Ext(appconf.Def.PolicyPath)

	if !strings.HasPrefix(configFileExt, ".") {
		err := fmt.Errorf("Unknow policy file type extension %s", configFileExt)
		log.Fatalf("error: %v", err)
		return nil, err
	}

	configType := strings.TrimPrefix(configFileExt, ".")

	if _, ok := supportedConfigType[configType]; !ok {
		err := fmt.Errorf("Unsupported policy file type extension %s", configType)
		log.Fatalf("error: %v", err)
		return nil, err
	}

	r, err := ioutil.ReadFile(appconf.Def.PolicyPath)
	if err != nil { // Handle errors reading the config file
		err := fmt.Errorf("Fatal error config file: %s", err)
		log.Fatalf("error: %v", err)
		return nil, err
	}
	runtimeViper := viper.New()
	runtimeViper.SetConfigType(configType)
	err = runtimeViper.ReadConfig(bytes.NewBuffer(r)) // Find and read the config file
	if err != nil {                                   // Handle errors reading the config file
		err := fmt.Errorf("Fatal error config file: %s", err)
		log.Fatalf("error: %v", err)
		return nil, err

	}

	c := CATConfig{}
	err = runtimeViper.Unmarshal(&c)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	return &c, err
}

// GetPlatformPolicy returns specified platform policies
func GetPlatformPolicy(cpu string) ([]Policy, error) {

	lock.Lock()
	defer lock.Unlock()
	var err error

	if config == nil {
		config, err = LoadPolicy()
		if err != nil {
			return []Policy{}, err
		}
	}

	p, ok := config.Catpolicy[cpu]

	if !ok {
		return []Policy{}, fmt.Errorf("Error while get platform policy: %s", cpu)
	}

	return p, nil
}

// GetDefaultPlatformPolicy is wrapper for GetPlatformPolicy
func GetDefaultPlatformPolicy() ([]Policy, error) {
	cpu := cpu.GetMicroArch(cpu.GetSignature())
	if cpu == "" {
		return []Policy{}, fmt.Errorf("Unknow platform, please update the cpu_map.toml conf file")
	}

	return GetPlatformPolicy(strings.ToLower(cpu))
}

// GetPolicy return a map of the policy of the host
func GetPolicy(cpu, policy string) (map[string]string, error) {
	m := make(map[string]string)

	platform, err := GetPlatformPolicy(cpu)

	if err != nil {
		return m, fmt.Errorf("Can not find specified platform policy")
	}

	var policyCandidate []Policy

	for _, p := range platform {
		_, ok := p[policy]
		if ok {
			policyCandidate = append(policyCandidate, p)
		}
	}
	if len(policyCandidate) == 1 {
		for _, item := range policyCandidate[0][policy] {
			// merge to one map
			for k, v := range item {
				m[k] = v
			}
		}
		return m, nil
	}
	return m, fmt.Errorf("Can not find specified policy %s", policy)
}

// GetDefaultPolicy return a map of the default policy of the host
func GetDefaultPolicy(policy string) (map[string]string, error) {
	cpu := cpu.GetMicroArch(cpu.GetSignature())
	if cpu == "" {
		return map[string]string{}, fmt.Errorf("Unknow platform, please update the cpu_map.toml conf file")
	}
	return GetPolicy(cpu, policy)
}

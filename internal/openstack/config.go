// +build openstack

package openstack

import (
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Config represenent configuration for Openstack features
type Config struct {
	ProviderConfigPath string `toml:"providerConfigPath"`
	AmqpURI            string `toml:"amqpuri"`
	BindingKey         string `toml:"bindingKey"`
	KeystoneURL        string `toml:"keystoneUrl"`
	KeystoneLogin      string `toml:"keystoneLogin"`
	KeystonePassword   string `toml:"keystonePassword"`
}

var oscfg = &Config{}
var runOnce sync.Once

// GetConfig reads P-State plugin configuration from configuration file
func GetConfig() *Config {
	runOnce.Do(func() {
		err := viper.UnmarshalKey("openstack", oscfg)
		if err != nil {
			log.Error("openstack.GetConfig() error:", err)
		}
	})
	return oscfg
}

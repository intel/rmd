// +build openstack

package openstack

import (
	"errors"
	"os"

	log "github.com/sirupsen/logrus"
)

var hostname string

// Init function initializes openstack integration features
func Init() error {
	//TODO maybe sync DoOnce?
	var err error
	hostname, err = os.Hostname()
	if err != nil {
		log.Errorf("Could not get host name, RMD will not be able to filter notifications for current host! %v", err)
		return errors.New("Could not get host name")
	}

	oscfg := GetConfig()
	// root process and rmd user process have to perform different operations
	if os.Geteuid() == 0 {
		// read OpenStack related configuration
		if len(oscfg.ProviderConfigPath) == 0 {
			log.Error("No Provider Configuration file path available")
			return errors.New("Missing Provider Configuration file path")
		}

		// launch Provider Config file generator
		if err := GenerateFile(oscfg.ProviderConfigPath); err != nil {
			log.Error("Failed to generate provider config file")
			return err
		}
	} else {
		// token is needed for authorization and access to an OpenStack environment
		getToken()

		// enable OpenStack events listener
		if err := NovaListenerStart(oscfg.AmqpURI, oscfg.BindingKey); err != nil {
			log.Error("Failed to launch Nova event listener")
			return err
		}
	}
	return nil
}

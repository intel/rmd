// +build !openstack

package openstack

import (
	"errors"

	log "github.com/sirupsen/logrus"
)

// Init function initializes openstack integration features
func Init() error {
	log.Error("Calling Openstack initialization while this RMD build does not support OpenStack integration")

	return errors.New("OpenStack enabled but not supported - check logs and config file")
}

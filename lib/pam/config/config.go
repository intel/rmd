package config

import (
	"github.com/spf13/viper"
	"sync"
)

// PAM is the configuration option
type PAM struct {
	Service string `toml:"service"`
}

var once sync.Once
var pam = &PAM{"rmd"}

// GetPAMConfig reads config from config file
func GetPAMConfig() *PAM {
	once.Do(func() {
		viper.UnmarshalKey("pam", pam)
	})
	return pam
}

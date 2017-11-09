package config

import (
	"strconv"
	"sync"

	// "github.com/heirko/go-contrib/logrusHelper"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Log is the log config struct
type Log struct {
	Path   string `toml:"path"`
	Env    string `toml:"env"`
	Level  string `toml:"level"`
	Stdout bool   `toml:"stdout"`
}

var configOnce sync.Once
var log = &Log{"var/log/rmd.log", "dev", "debug", true}

// NewConfig loads log config
func NewConfig() Log {
	configOnce.Do(func() {
		// FIXME , we are planing to use logrusHelper. Seems we still
		// need missing some initialization for logrus. But it repors error as
		// follow:
		// # github.com/heirko/go-contrib/logrusHelper
		// undefined: logrus_mate.LoggerConfig
		// var c = logrusHelper.UnmarshalConfiguration(viper) // Unmarshal configuration from Viper
		// logrusHelper.SetConfig(logrus.StandardLogger(), c) // for e.g. apply it to logrus default instance

		viper.UnmarshalKey("log", log)

		logDir := pflag.Lookup("log-dir").Value.String()
		if logDir != "" {
			log.Path = logDir
		}

		// FIXME , we should get the value of logtostderr by reflect
		// or flag directly, instead of strconv.ParseBool
		tostd, _ := strconv.ParseBool(pflag.Lookup("logtostderr").Value.String())
		if tostd == true {
			log.Stdout = true
		}
	})

	return *log
}

package config

import (
	"github.com/spf13/viper"
	"sync"
)

// Database struct
type Database struct {
	Backend   string `toml:"backend"`
	Transport string `toml:"transport"`
	DBName    string `toml:"dbname"`
}

var configOnce sync.Once

// FIXME , the DBName should not be "bolt", "rmd" is more better.
var db = &Database{"bolt", "var/run/rmd.db", "bolt"}

// NewConfig is Concurrency safe.
func NewConfig() Database {
	configOnce.Do(func() {
		// FIXME , we are planing to use logrusHelper. Seems we still
		// need missing some initialization for logrus. But it repors error as
		// follow:
		// # github.com/heirko/go-contrib/logrusHelper
		// undefined: logrus_mate.LoggerConfig
		// var c = logrusHelper.UnmarshalConfiguration(viper) // Unmarshal configuration from Viper
		// logrusHelper.SetConfig(logrus.StandardLogger(), c) // for e.g. apply it to logrus default instance

		viper.UnmarshalKey("database", db)
	})

	return *db
}

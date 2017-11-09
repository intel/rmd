package config

import (
	"crypto/tls"
	"sync"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// ClientAuth is a string to tls clientAuthType map
var ClientAuth = map[string]tls.ClientAuthType{
	"no":              tls.NoClientCert,
	"require":         tls.RequestClientCert,
	"require_any":     tls.RequireAnyClientCert,
	"challenge_given": tls.VerifyClientCertIfGiven,
	"challenge":       tls.RequireAndVerifyClientCert,
}

const (
	// CAFile is the certificate authority file
	CAFile = "ca.pem"
	// CertFile is the certificate file
	CertFile = "rmd-cert.pem"
	// KeyFile is the rmd private key file
	KeyFile = "rmd-key.pem"
	// ClientCAFile certificate authority file of client side
	ClientCAFile = "ca.pem"
)

// Default is the configuration in default section of config file
// TODO consider create a new struct for TLSConfig
type Default struct {
	Address      string `toml:"address"`
	TLSPort      uint   `toml:"tlsport"`
	CertPath     string `toml:"certpath"`
	ClientCAPath string `toml:"clientcapath"`
	ClientAuth   string `toml:"clientauth"`
	UnixSock     string `toml:"unixsock"`
	PolicyPath   string `toml:"policypath"`
}

// Database represents data base configuration
type Database struct {
	Backend   string `toml:"backend"`
	Transport string `toml:"transport"`
	DBName    string `toml:"dbname"`
}

// Debug configurations
type Debug struct {
	Enabled   bool `toml:"enabled"`
	Debugport uint `toml:"debugport"`
}

// Config represent the configuration struct
type Config struct {
	Def Default  `mapstructure:"default"`
	Db  Database `mapstructure:"database"`
	Dbg Debug    `mapstructure:"debug"`
}

var configOnce sync.Once
var def = Default{
	"localhost",
	8443,
	"etc/rmd/cert/server",
	"etc/rmd/cert/client",
	"challenge",
	"",
	"etc/rmd/policy.yaml",
}

var db = Database{}
var dbg = Debug{}
var config = &Config{def, db, dbg}

// NewConfig loads configurations from config file and pflag
func NewConfig() Config {

	configOnce.Do(func() {

		// Take the value from pflag which was defined in flag.go
		// The default value of the struct as not taken if we
		// bind pflag to viper
		viper.BindPFlag("default.address", pflag.Lookup("address"))
		viper.BindPFlag("default.tlsport", pflag.Lookup("tlsport"))
		viper.BindPFlag("default.unixsock", pflag.Lookup("unixsock"))
		viper.BindPFlag("default.clientauth", pflag.Lookup("clientauth"))
		viper.BindPFlag("debug.enabled", pflag.Lookup("debug"))
		viper.BindPFlag("debug.debugport", pflag.Lookup("debugport"))

		err := viper.Unmarshal(config)
		if err != nil {
			panic(err)
		}
	})

	return *config
}

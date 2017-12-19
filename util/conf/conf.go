package conf

import (
	// Do init flag
	_ "flag"
	"fmt"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var defaultConfigPath = []string{
	"/usr/local/etc/rmd/",
	"/etc/rmd/",
	"./etc/rmd",
}

// Init does config initial
func Init() error {
	viper.SetConfigName("rmd") // no need to include file extension
	// TODO  consider to introduce Cobra. let Viper work with Cobra.
	confDir := pflag.Lookup("conf-dir").Value.String()
	if confDir != "" {
		viper.AddConfigPath(confDir)
	}
	for _, p := range defaultConfigPath {
		viper.AddConfigPath(p)
	}
	err := viper.ReadInConfig()
	if err != nil {
		// NOTE (ShaoHe Feng): only can use fmt.Println, can not use log.
		// For log is not init at this point.
		fmt.Printf("No config file found from %v, fall back to using default setting\n", defaultConfigPath)
	}
	return nil
}

package conf

import (
	// Do init flag
	_ "flag"
	"fmt"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Init does config initial
func Init() error {
	viper.SetConfigName("rmd") // no need to include file extension
	// TODO  consider to introduce Cobra. let Viper work with Cobra.
	confDir := pflag.Lookup("conf-dir").Value.String()
	if confDir != "" {
		viper.AddConfigPath(confDir)
	}
	viper.AddConfigPath("/etc/rmd/")  // path to look for the config file in
	viper.AddConfigPath("$HOME/rmd")  // call multiple times to add many search paths
	viper.AddConfigPath("./etc/rmd/") // set the path of your config file
	err := viper.ReadInConfig()
	if err != nil {
		// NOTE (ShaoHe Feng): only can use fmt.Println, can not use log.
		// For log is not init at this point.
		fmt.Println(err)
		return err
	}
	return nil
}

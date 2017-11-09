package cpu

import (
	"github.com/spf13/viper"
	"sync"
)

// ignore stepping
var m = map[uint32]string{
	0x406e0: "Skylake",
	0x506e0: "Skylake",
	0x50650: "Skylake",
	0x806e0: "Kaby Lake",
	0x906e0: "Kaby Lake",
	0x306d0: "Broadwell",
	0x406f0: "Broadwell",
}

type microArch struct {
	Family int `toml:"famliy"`
	Model  int `toml:"model"`
}

var cpumapOnce sync.Once

// NewCPUMap init internal cpu map
// Concurrency safe.
func NewCPUMap() map[uint32]string {
	cpumapOnce.Do(func() {
		var rtViper = viper.New()
		var maps = map[string][]microArch{}

		// supported extensions are "json", "toml", "yaml", "yml", "properties", "props", "prop"
		rtViper.SetConfigType("toml")
		rtViper.SetConfigName("cpu_map")    // no need to include file extension
		rtViper.AddConfigPath("/etc/rmd/")  // path to look for the config file in
		rtViper.AddConfigPath("$HOME/rmd")  // call multiple times to add many search paths
		rtViper.AddConfigPath("./etc/rmd/") // set the path of your config file
		err := rtViper.ReadInConfig()

		if err != nil {
			panic("Failed to Read from CPU config")
		}

		rtViper.Unmarshal(&maps)
		for k, mv := range maps {
			for _, v := range mv {
				sig := (v.Family>>4)<<20 + (v.Family&0xf)<<8 + (v.Model>>4)<<16 + (v.Model&0xf)<<4
				m[uint32(sig)] = k
			}
		}

	})
	return m
}

// GetMicroArch returns micro arch
func GetMicroArch(sig uint32) string {
	s := sig & 0xFFFF0FF0
	NewCPUMap()
	if v, ok := m[s]; ok {
		return v
	}
	return ""
}

package cpu

import (
	"github.com/spf13/viper"

	"fmt"
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
var defaultConfigPath = []string{
	"/usr/local/etc/rmd/",
	"/etc/rmd/",
	"./etc/rmd",
}

// NewCPUMap init internal cpu map
// Concurrency safe.
func NewCPUMap() map[uint32]string {
	cpumapOnce.Do(func() {
		var rtViper = viper.New()
		var maps = map[string][]microArch{}

		rtViper.SetConfigType("toml")
		rtViper.SetConfigName("cpu_map")
		for _, p := range defaultConfigPath {
			viper.AddConfigPath(p)
		}
		err := rtViper.ReadInConfig()

		if err != nil {
			// TODO using log
			fmt.Printf("No cpu map found from %v, fall back to using default cpu map\n", defaultConfigPath)
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

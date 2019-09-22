package config

import (
	"sync"

	"github.com/spf13/viper"
)

// OSGroup represents os group configuration
type OSGroup struct {
	CacheWays     uint   `toml:"cacheways"`
	CPUSet        string `toml:"cpuset"`
	MbaPercentage int
}

// InfraGroup represents infra group configuration
type InfraGroup struct {
	CacheWays     uint     `toml:"cacheways"`
	CPUSet        string   `toml:"cpuset"`
	Tasks         []string `toml:"tasks"`
	MbaPercentage int
}

// CachePool represents cache pool layout configuration
type CachePool struct {
	MaxAllowedShared    uint `toml:"max_allowed_shared"`
	Guarantee           uint `toml:"guarantee"`
	Besteffort          uint `toml:"besteffort"`
	Shared              uint `toml:"shared"`
	Shrink              bool `toml:"shrink"`
	MbaPercentageShared int
}

var infraConfigOnce sync.Once
var osConfigOnce sync.Once
var cachePoolConfigOnce sync.Once

var infragroup = &InfraGroup{1, "0", nil, -1}
var osgroup = &OSGroup{1, "0", -1}

// FIXME: the default may not work on some platform
var cachepool = &CachePool{10, 10, 7, 2, false, -1}

// NewInfraConfig reads InfraGroup configuration
func NewInfraConfig() *InfraGroup {
	infraConfigOnce.Do(func() {
		key := "InfraGroup"
		if !viper.IsSet(key) {
			infragroup = nil
			return
		}
		viper.UnmarshalKey(key, infragroup)
	})
	return infragroup
}

// NewOSConfig reads OSGroup configuration
func NewOSConfig() *OSGroup {
	osConfigOnce.Do(func() {
		viper.UnmarshalKey("OSGroup", osgroup)
	})
	return osgroup
}

// NewCachePoolConfig reads cache pool layout configuration
func NewCachePoolConfig() *CachePool {
	cachePoolConfigOnce.Do(func() {
		viper.UnmarshalKey("CachePool", cachepool)
	})
	return cachepool
}

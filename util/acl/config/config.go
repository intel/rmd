package config

import (
	"strings"
	"sync"

	"github.com/spf13/viper"
)

// Author describes a set of authorization.
type Author uint32

// These are the generalized file operations that can trigger a notification.
const (
	Signature Author = 1 << iota
	OU
	CN
)

const (
	CertClientUserRole = "Cert Client User"
)

// ACL what?
type ACL struct {
	Path          string `toml:"path"`
	Filter        string `toml:"filter"`
	AdminCert     string `toml:"admincert"`
	UserCert      string `toml:"usercert"`
	Authorization Author
}

var authmap = map[string]Author{
	"signature": Signature,
	"role":      OU,
	"username":  CN,
}

var once sync.Once
var acl = &ACL{"/etc/rmd/acl/", "url", "", "", Signature}

// NewACLConfig create new ACL config
func NewACLConfig() *ACL {
	once.Do(func() {
		viper.UnmarshalKey("acl", acl)
		a := strings.ToLower(viper.GetString("acl.authorization"))
		if v, ok := authmap[a]; ok {
			acl.Authorization = v
		}
	})
	return acl
}

package acl

import (
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/casbin/casbin"
	"github.com/emicklei/go-restful"
	log "github.com/sirupsen/logrus"

	"github.com/intel/rmd/util/acl/config"
)

// Enforcer does enforce
type Enforcer struct {
	url      *casbin.Enforcer
	ip       *casbin.Enforcer
	protocol *casbin.Enforcer
}

// VersionTrim is ...
var VersionTrim = regexp.MustCompile(`^/v\d+/`)
var enforcer = &Enforcer{}
var once sync.Once

// NewEnforcer creates enforcer
func NewEnforcer() (*Enforcer, error) {
	var returnErr error
	defer func() {
		if r := recover(); r != nil {
			returnErr = fmt.Errorf("init Enforcer error: %s", r)
		}
	}()

	once.Do(func() {
		aclconf := config.NewACLConfig()
		for _, filter := range strings.Split(aclconf.Filter, ",") {
			model := path.Join(aclconf.Path, filter, "model.conf")
			policy := path.Join(aclconf.Path, filter, "policy.csv")
			switch filter {
			case "url":
				e := casbin.NewEnforcer(model, policy)
				// NOTE, the policy file should define a role named user.
				e.AddRoleForUser(config.CertClientUserRole, "user")
				enforcer.url = e
				log.Infof("succssfully set %s acl", filter)
			case "ip":
				log.Infof("Do not support %s acl at present", filter)
			case "proto":
				log.Infof("Do not support %s acl at present", filter)
			default:
				log.Errorf("Unknow acl type %s", filter)
			}
		}
	})
	return enforcer, returnErr
}

// Enforce does enforce based on request
func (e *Enforcer) Enforce(request *restful.Request, sub string) bool {
	allow := false
	obj := VersionTrim.ReplaceAllString(path.Clean(request.Request.RequestURI), "/")
	act := request.Request.Method
	if e.url != nil {
		allow = e.url.Enforce(sub, obj, act)
	} else {
		allow = true
	}
	// TODO support ip and proto Enforce.
	return allow
}

// GetAdminCerts Get all Admin certification files from a given path.
func GetAdminCerts() ([]string, error) {
	aclconf := config.NewACLConfig()
	if aclconf.AdminCert == "" {
		return []string{}, nil
	}

	return filepath.Glob(aclconf.AdminCert + "/*.pem")
}

// GetUserCerts Get all user certification files from a given path.
func GetUserCerts() ([]string, error) {
	aclconf := config.NewACLConfig()
	if aclconf.UserCert == "" {
		return []string{}, nil
	}

	return filepath.Glob(aclconf.UserCert + "/*.pem")
}

// GetCertsPath Get all certification files include admin's and common user's from a given path.
func GetCertsPath() []string {
	aclconf := config.NewACLConfig()
	return []string{aclconf.AdminCert, aclconf.UserCert}
}

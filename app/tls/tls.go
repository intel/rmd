package tls

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"

	"github.com/emicklei/go-restful"
	log "github.com/sirupsen/logrus"

	appConf "github.com/intel/rmd/app/config"
	acl "github.com/intel/rmd/util/acl"
	aclConf "github.com/intel/rmd/util/acl/config"
	"net/http"
)

var signatureRWM sync.RWMutex
var adminCertSignature []string
var userCertSignature []string

// GenTLSConfig  generate TLS configure.
func GenTLSConfig() (*tls.Config, error) {
	var roots *x509.CertPool
	var clientPool *x509.CertPool
	tlsfiles := map[string]string{}
	appconf := appConf.NewConfig()
	files, err := filepath.Glob(appconf.Def.CertPath + "/*.pem")
	if err != nil {
		return nil, err
	}
	// avoid to check whether files exist.
	for _, f := range files {
		switch filepath.Base(f) {
		case appConf.CAFile:
			tlsfiles["ca"] = f
			roots, err = GetCertPool(f)
			if err != nil {
				return nil, err
			}
		case appConf.CertFile:
			tlsfiles["cert"] = f
		case appConf.KeyFile:
			tlsfiles["key"] = f
		}
	}
	if len(tlsfiles) < 3 {
		missing := []string{}
		for _, k := range []string{"cert", "ca", "key"} {
			_, ok := tlsfiles[k]
			if !ok {
				missing = append(missing, k)
			}
		}
		return nil, fmt.Errorf("Missing enough files for tls config: %s", strings.Join(missing, ", "))
	}

	// In product env, ClientAuth should >= challenge_given
	clientauth, ok := appConf.ClientAuth[appconf.Def.ClientAuth]
	if !ok {
		return nil, errors.New(
			"Unknow ClientAuth config setting: " + appconf.Def.ClientAuth)
	}
	if clientauth >= appConf.ClientAuth["challenge_given"] {
		clientPool, err = GetCertPool(filepath.Join(appconf.Def.ClientCAPath, appConf.ClientCAFile))
		if err != nil {
			return nil, err
		}
	}

	tlsCert, err := tls.LoadX509KeyPair(tlsfiles["cert"], tlsfiles["key"])
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		RootCAs:      roots,
		ClientAuth:   clientauth,
		Certificates: []tls.Certificate{tlsCert},
		ClientCAs:    clientPool,
		MinVersion:   tls.VersionTLS11,
	}, nil
}

// ACL is handler for api server to pass acl of a request
func ACL(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	check := func(user string) {
		e, _ := acl.NewEnforcer()
		if e.Enforce(req, user) != true {
			log.Errorf("User " + user + " is not authorized to access this resource")
			resp.WriteErrorString(http.StatusForbidden, "User \""+user+"\" is not authorized to access this resource\n")
			return
		}
		chain.ProcessFilter(req, resp)
	}

	appconf := appConf.NewConfig()
	clientauth, ok := appConf.ClientAuth[appconf.Def.ClientAuth]

	if !ok {
		log.Errorf("Bad client auth option configuration:" + appconf.Def.ClientAuth)
		resp.WriteErrorString(http.StatusInternalServerError, "Bad client auth option configuration\n")
		return
	}

	if req.Request.TLS == nil || (clientauth > tls.NoClientCert && clientauth < tls.RequireAndVerifyClientCert) {
		chain.ProcessFilter(req, resp)
		return
	}

	if clientauth == tls.NoClientCert {
		// Validate user obtained from basic authentication.
		// The credentials have passed the PAM authentication test.

		// Get user credentials
		u, _, ok := req.Request.BasicAuth()

		if !ok {
			resp.WriteErrorString(http.StatusBadRequest, "Malformed credentials\n")
			return
		}

		// Check user against ACL rules
		check(u)
		return
	}

	cn := req.Request.TLS.PeerCertificates[0].Subject.CommonName
	author := aclConf.NewACLConfig().Authorization
	if author == aclConf.Signature {
		sig := string(req.Request.TLS.PeerCertificates[0].Signature)
		for _, s := range GetAdminCertSignatures() {
			if strings.Compare(string(sig), s) == 0 {
				chain.ProcessFilter(req, resp)
				return
			}
		}
		for _, s := range GetUserCertSignatures() {
			if strings.Compare(string(sig), s) == 0 {
				check(aclConf.CertClientUserRole)
				return
			}
		}
		log.Errorf(cn + "is not allow to access this resource, with its signature : " + sig)
		resp.WriteErrorString(401, cn+" is not Authorized.")
		return
	}

	if author == aclConf.OU {
		e, _ := acl.NewEnforcer()
		OU := req.Request.TLS.PeerCertificates[0].Subject.OrganizationalUnit
		for _, v := range OU {
			if e.Enforce(req, strings.ToLower(v)) == true {
				chain.ProcessFilter(req, resp)
				return
			}
		}
		log.Errorf("User is not allow to access this resource")
		resp.WriteErrorString(401, "OU: "+strings.Join(OU, ", ")+" is not Authorized")
		return
	}

	if author == aclConf.CN {
		check(cn)
	}
}

// GetCertPool Get Certification pool
func GetCertPool(cafile string) (*x509.CertPool, error) {
	pool := x509.NewCertPool()
	// Should we get SystemCertPool ?
	data, err := ioutil.ReadFile(cafile)
	if err != nil {
		return nil, err
	}
	ok := pool.AppendCertsFromPEM(data)
	if !ok {
		return nil, errors.New("failed to parse root certificate")
	}
	return pool, nil
}

// NewCertSignatures Generate a list of Certification Signature
func NewCertSignatures(admin bool) (signatures []string, err error) {
	var files []string
	if admin {
		files, err = acl.GetAdminCerts()
	} else {
		files, err = acl.GetUserCerts()
	}
	if err != nil {
		return signatures, err
	}

	for _, f := range files {
		dat, err := ioutil.ReadFile(f)
		if err != nil {
			log.Errorf("Unable to read signatures file: %s. Error: %s", f, err)
		}
		block, _ := pem.Decode(dat)
		if block == nil || block.Type != "CERTIFICATE" {
			log.Errorf("Failed to decode client certificate %s. Certificate type: %s", f, block.Type)
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			log.Errorf("Failed to parse client certificate %s. Error: %s", f, err)
		} else {
			signatures = append(signatures, string(cert.Signature))
		}
	}

	return signatures, nil
}

// InitCertSignatures Initialize the list of Certification Signature
// Should be called once.
func InitCertSignatures() (err error) {
	var watcher *fsnotify.Watcher
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	// No choice to close watcher. V2 will support goroutine gracefully exit.
	// defer watcher.Close()

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&(fsnotify.Create+fsnotify.Write+fsnotify.Remove) > 0 {
					log.Infof("Client cert files are changed, reload. Event: %s", event)
					paths := acl.GetCertsPath()
					if filepath.HasPrefix(event.Name, paths[0]) {
						cs, err := NewCertSignatures(true)
						if err != nil {
							log.Errorf("Error to get admin client signatures list. %s", err)
						}
						signatureRWM.Lock()
						adminCertSignature = cs
						signatureRWM.Unlock()
						log.Infof("Load %d valid admin certificate signatures.", len(adminCertSignature))
					} else if filepath.HasPrefix(event.Name, paths[1]) {
						cs, err := NewCertSignatures(false)
						if err != nil {
							log.Errorf("Error to get common user client signatures list. %s", err)
						}
						signatureRWM.Lock()
						userCertSignature = cs
						signatureRWM.Unlock()
						log.Infof("Load %d valid common user certificate signatures.", len(userCertSignature))
					}
				}
			case err := <-watcher.Errors:
				log.Errorf("Error to watch client certificate path. Error: %s", err)
			}
		}
	}()

	for _, p := range acl.GetCertsPath() {
		err = watcher.Add(p)
		if err != nil {
			return err
		}
	}
	adminCertSignature, err = NewCertSignatures(true)
	if err != nil {
		return err
	}
	userCertSignature, err = NewCertSignatures(false)
	return
}

// GetAdminCertSignatures Get the list of Certification Signature
func GetAdminCertSignatures() []string {
	signatureRWM.RLock()
	defer signatureRWM.RUnlock()
	return adminCertSignature
}

// GetUserCertSignatures Get the list of Certification Signature
func GetUserCertSignatures() []string {
	signatureRWM.RLock()
	defer signatureRWM.RUnlock()
	return userCertSignature
}

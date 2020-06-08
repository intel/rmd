package main

import (
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	appConf "github.com/intel/rmd/utils/config"

	"github.com/emicklei/go-restful"
	"github.com/intel/rmd/internal/openstack"
	"github.com/intel/rmd/internal/plugins"
	"github.com/intel/rmd/modules/cache"
	"github.com/intel/rmd/modules/hospitality"
	"github.com/intel/rmd/modules/mba"
	"github.com/intel/rmd/modules/policy"
	"github.com/intel/rmd/modules/workload"
	"github.com/intel/rmd/utils/auth"
	apptls "github.com/intel/rmd/utils/tls"
	log "github.com/sirupsen/logrus"
)

const (
	prefix string = "/v1/"
)

// GenericConfig is the generic config for the application
type GenericConfig struct {
	Address                string
	Port                   string
	TLSPort                string
	UnixSock               string
	EnableUISupport        bool
	DBBackend              string
	Transport              string
	DBName                 string
	IsClientCertAuthOption bool
	Debug                  bool
	OpenStackEnable        bool
}

// Config is application configuration
type Config struct {
	Generic *GenericConfig
}

func buildServerConfig() *Config {

	appconfig := appConf.NewConfig()
	// Default is client cert option for tls
	isClientCertAuthOption := true

	clientAuthOption, ok := appConf.ClientAuth[appconfig.Def.ClientAuth]

	// Default to cert option if invalid client auth option is specified
	if ok && (clientAuthOption == tls.NoClientCert) {
		isClientCertAuthOption = false
	}

	genericconfig := GenericConfig{
		Address:                appconfig.Def.Address,
		TLSPort:                strconv.FormatUint(uint64(appconfig.Def.TLSPort), 10),
		Port:                   strconv.FormatUint(uint64(appconfig.Dbg.Debugport), 10),
		UnixSock:               appconfig.Def.UnixSock,
		DBBackend:              appconfig.Db.Backend,
		Transport:              appconfig.Db.Transport,
		DBName:                 appconfig.Db.DBName,
		IsClientCertAuthOption: isClientCertAuthOption,
		Debug:                  appconfig.Dbg.Enabled,
		OpenStackEnable:        appconfig.Def.OpenStackEnable,
	}

	return &Config{
		Generic: &genericconfig,
	}
}

// Initialize server from config
func Initialize(c *Config) (*restful.Container, error) {

	// By default, no admin/user cert path is configured, so don't initialize
	// certification signature
	if !c.Generic.Debug {
		if err := apptls.InitCertSignatures(); err != nil {
			return nil, err
		}
	}

	wsContainer := restful.NewContainer()

	// Enable PAM authentication when "no" client cert auth option is provided
	if !c.Generic.IsClientCertAuthOption {
		wsContainer.Filter(auth.PAMAuthenticate)
	}

	wsContainer.Filter(apptls.ACL)
	wsContainer.Router(restful.CurlyRouter{})

	// Register controller to container
	cache.Register(prefix, wsContainer)
	policy.Register(prefix, wsContainer)
	hospitality.Register(prefix, wsContainer)
	workload.Register(prefix, wsContainer)
	mba.Register(prefix, wsContainer)

	// iterate over list of modules and register enpoints for them
	for pluginName, pluginInterface := range plugins.Interfaces {
		log.Debugf("Registering REST endpoint for plugin: %v", pluginName)
		if pluginInterface == nil {
			// this should never happen but better handle this situation to avoid segfaults
			log.Errorf("Nil plugin interface found in registered plugins list")
			return wsContainer, errors.New("Internal error: nil interface")
		}
		endpoints := pluginInterface.GetEndpointPrefixes()
		if len(endpoints) == 0 {
			// no REST endpoints provided by this module
			continue
		}

		ws := new(restful.WebService)
		ws.
			Path(prefix).
			Consumes(restful.MIME_JSON).
			Produces(restful.MIME_JSON)

		for _, ep := range endpoints {
			log.Debugf("- adding %v", ep)
			ws.Route(ws.GET(ep).To(pluginInterface.HandleRequest))
		}
		wsContainer.Add(ws)
	}

	return wsContainer, nil
}

// RunServer to run the apiserver.
func RunServer() {

	var server *http.Server
	config := buildServerConfig()

	container, err := Initialize(config)
	if err != nil {
		log.Fatal(err)
	}

	// Notification listener should run in rmd user process
	if config.Generic.OpenStackEnable {
		if err := openstack.Init(); err != nil {
			log.Error("OpenStack initialization failed: ", err.Error())
		}
		log.Debug("OpenStack initialized properly")
	}

	// TODO cleanup this mass logic, especially for unixsock

	if config.Generic.Debug {
		server = &http.Server{
			Addr:    config.Generic.Address + ":" + config.Generic.Port,
			Handler: container}
	} else {
		// TODO We need to config server.TLSConfig
		// TODO Support self-sign CA. self-sign CA can be in development evn.
		tlsconf, err := apptls.GenTLSConfig()
		if err != nil {
			log.Fatal(err)
		}

		server = &http.Server{
			Addr:         config.Generic.Address + ":" + config.Generic.TLSPort,
			Handler:      container,
			TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
			TLSConfig:    tlsconf}
	}

	serverStart := func() {
		if config.Generic.Debug {
			log.Fatal(server.ListenAndServe())
		} else {
			log.Fatal(server.ListenAndServeTLS("", "")) // Use certs from TLSConfig.
		}
	}

	if config.Generic.UnixSock == "" {
		serverStart()
	} else {
		go func() {
			serverStart()
		}()
	}

	// Unix Socket.
	container, err = Initialize(config)
	if err != nil {
		log.Fatal(err)
	}

	userver := &http.Server{
		Handler: container}

	unixListener, err := net.Listen("unix", config.Generic.UnixSock)
	if err != nil {
		log.Info(err, unixListener)
		return
	}
	// TODO need to check, should defer unixListener.Close()
	defer func() {
		if unixListener != nil {
			log.Infof("Close Unix socket listener. RMD exits!")
			unixListener.Close()
		}
	}()

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt, syscall.SIGTERM)
	go func(l net.Listener, c chan os.Signal) {
		sig := <-c
		if l != nil {
			log.Infof("Close Unix socket listener.")
			l.Close()
		}
		log.Infof("Caught signal %s: RMD exits!", sig)
		os.Exit(0)
	}(unixListener, sigchan)

	//REMOVE these 2 line codes, if we want to support Unix Socket!
	unixListener.Close()
	log.Fatal("Sorry, do not support Unix listener at present!")

	err = userver.Serve(unixListener)
	if err != nil {
		log.Fatal(err)
	}
}

package app

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	appConf "github.com/intel/rmd/app/config"

	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful-swagger12"
	"github.com/intel/rmd/api/v1"
	apptls "github.com/intel/rmd/app/tls"
	"github.com/intel/rmd/db"
	"github.com/intel/rmd/util/auth"
	log "github.com/sirupsen/logrus"
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
}

// Config is application configuration
type Config struct {
	Generic *GenericConfig
	Swagger *swagger.Config
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
		Debug: appconfig.Dbg.Enabled,
	}

	swaggerconfig := swagger.Config{
		WebServicesUrl: fmt.Sprintf("http://%s:%d", appconfig.Def.Address, appconfig.Dbg.Debugport),
		ApiPath:        "/apidocs.json",
		SwaggerPath:    "/apidocs/", // Optionally, specifiy where the UI is located
		// FIXME this depends on https://github.com/swagger-api/swagger-ui.git need to copy dist from it
		SwaggerFilePath: "/usr/local/share/go/src/github.com/wordnik/swagger-ui/dist",
		ApiVersion:      "1.0",
	}

	return &Config{
		Generic: &genericconfig,
		Swagger: &swaggerconfig,
	}
}

// Initialize server from config
func Initialize(c *Config) (*restful.Container, error) {
	db, err := db.NewDB()
	if err != nil {
		return nil, err
	}

	if err := apptls.InitCertSignatures(); err != nil {
		return nil, err
	}

	wsContainer := restful.NewContainer()

	// Enable PAM authentication when "no" client cert auth option is provided
	if !c.Generic.IsClientCertAuthOption {
		wsContainer.Filter(auth.PAMAuthenticate)
	}

	wsContainer.Filter(apptls.ACL)
	wsContainer.Router(restful.CurlyRouter{})

	caches := v1.CachesResource{}
	policy := v1.PolicyResource{}
	hospitality := v1.HospitalityResource{}
	wls := v1.WorkLoadResource{Db: db}

	// Register controller to container
	caches.Register(wsContainer)
	policy.Register(wsContainer)
	hospitality.Register(wsContainer)
	wls.Register(wsContainer)

	// Install adds the SgaggerUI webservices
	c.Swagger.WebServices = wsContainer.RegisteredWebServices()
	swagger.RegisterSwaggerService(*(c.Swagger), wsContainer)

	// TODO error handle
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
			Addr:      config.Generic.Address + ":" + config.Generic.TLSPort,
			Handler:   container,
			TLSConfig: tlsconf}
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

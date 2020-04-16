package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	dbconf "github.com/intel/rmd/internal/db/config"
	"github.com/intel/rmd/internal/openstack"
	"github.com/intel/rmd/internal/plugins"
	proxyclient "github.com/intel/rmd/internal/proxy/client"
	proxyserver "github.com/intel/rmd/internal/proxy/server"
	proxytypes "github.com/intel/rmd/internal/proxy/types"
	"github.com/intel/rmd/modules/pstate"
	util "github.com/intel/rmd/utils"
	"github.com/intel/rmd/utils/bootcheck"
	"github.com/intel/rmd/utils/config"
	appconf "github.com/intel/rmd/utils/config"
	"github.com/intel/rmd/utils/flag"
	"github.com/intel/rmd/utils/log"
	logconf "github.com/intel/rmd/utils/log/config"
	"github.com/intel/rmd/utils/pidfile"
	"github.com/intel/rmd/utils/proc"
	"github.com/intel/rmd/utils/resctrl"
	"github.com/intel/rmd/version"
	loginfo "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

var rmduser = "rmd"

func main() {
	// use pipe pair to communicate between root and normal process
	var in, out proxytypes.PipePair
	flag.InitFlags()

	if pflag.Lookup("version").Value.String() == "true" {
		fmt.Printf("RMD version: %s (%s)\n", version.Info["version"], version.Info["revision"])
		os.Exit(0)
	}

	if err := config.Init(); err != nil {
		fmt.Println("Init config failed:", err)
		os.Exit(1)
	}
	if err := log.Init(); err != nil {
		fmt.Println("Init log failed:", err)
		os.Exit(1)
	}

	// check if pstate configuration exists and module is enabled
	pcfg, err := plugins.GetConfig("pstate")
	if err == nil && len(pcfg) > 0 {
		cfgflag, ok1 := pcfg["enabled"]
		cfgpath, ok2 := pcfg["path"]
		if !ok1 || !ok2 {
			// incomplete configuration
			fmt.Println("Configuration for 'pstate' is not complete - exiting")
			os.Exit(1)
		}

		enabled, ok1 := cfgflag.(bool)
		path, ok2 := cfgpath.(string)
		if !ok1 || !ok2 {
			// incorrect types in configuration
			fmt.Println("Configuration for 'pstate' has incorrect value types - exiting")
			os.Exit(1)
		}

		if enabled {
			err := pstate.Load(path)
			if err != nil {
				fmt.Println("Failed to load pstate plugin library:", err.Error())
				os.Exit(1)
			}
			// add information about proxy to configuration
			pcfg["RMDPROXY"] = &proxyclient.GenericCaller
			// pstate.Instance cannot be nil if no error from Load() returned
			err = pstate.Instance.Initialize(pcfg)
			if err != nil {
				fmt.Println("Failed to initialize 'pstate' instance:", err.Error())
				os.Exit(1)
			}
		}
	} else {
		// logger should work here so only log this message
		loginfo.Println("No proper config for 'pstate' - skipping P-State module")
	}

	if err := proc.Init(); err != nil {
		fmt.Println("proc init failed:", err)
		os.Exit(1)
	}

	if err := resctrl.Init(); err != nil {
		fmt.Println("resctrl init failed:", err)
		os.Exit(1)
	}

	cleanupFunc := func() {
		pidfile.ClosePID()
		in.Reader.Close()
		out.Writer.Close()
		out.Reader.Close()
		in.Writer.Close()
	}

	if os.Getuid() == 0 {
		if !util.IsUserExist(rmduser) {
			if err := util.CreateUser(rmduser); err != nil {
				fmt.Printf("Failed to create %s user", rmduser)
				os.Exit(1)
			}
		}

		if err := pidfile.CreatePID(); err != nil {
			fmt.Println("Create PID file fail. Reason: " + err.Error())
			os.Exit(1)
		}

		in.Reader, out.Writer, _ = os.Pipe()
		out.Reader, in.Writer, _ = os.Pipe()

		// FIXME: This is a quickly fix. Will improve later.
		file := logconf.NewConfig().Path
		if err := util.Chown(file, rmduser); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		ts := dbconf.NewConfig().Transport
		bd := dbconf.NewConfig().Backend
		if bd == "bolt" {
			if err := util.Chown(ts, rmduser); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
		child, err := util.DropRunAs(rmduser, logconf.NewConfig().Stdout, in.Writer, in.Reader)

		if err != nil {
			fmt.Println("Failed to drop root privilege")
			os.Exit(1)
		}

		// wait for child status
		go func(p *os.Process) {
			processState, _ := p.Wait()
			if !processState.Success() {
				fmt.Println("Failed to start rmd API server, check log for details")
				cleanupFunc()
				os.Exit(1)
			}
		}(child)

		// Part of OpenStack initialization has to be done as root
		cfg := appconf.NewConfig()
		if cfg.Def.OpenStackEnable {
			if err := openstack.Init(); err != nil {
				fmt.Println("openstack.Init() failed:", err)
				os.Exit(1)
			}
		}

		fmt.Printf("RMD server started, REST API server serving on process %d\n", child.Pid)
		proxyserver.RegisterAndServe(out)
	}

	// Below are executed in child process
	flg := syscall.SIGHUP
	if _, _, err := syscall.RawSyscall(syscall.SYS_PRCTL, syscall.PR_SET_PDEATHSIG, uintptr(flg), 0); err != 0 {
		loginfo.Println(err)
		os.Exit(1)
	}
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGHUP)
	go func() {
		sig := <-sigc
		//NOTE, should we add some cleanup?
		cleanupFunc()
		loginfo.Printf("Received %s, shutdown RMD for root process exit.", sig.String())
		// Do not Exit(0), for there are some thing wrong with supper RMD.
		os.Exit(1)
	}()

	//in.Writer
	in.Writer = os.NewFile(3, "")
	//in.Reader
	in.Reader = os.NewFile(4, "")
	err = proxyclient.ConnectRPCServer(in)
	if err != nil {
		loginfo.Println(err)
		os.Exit(1)
	}
	// should go after connect rpc server
	bootcheck.SanityCheck()
	// should tell root process we are fail or success for the santify check
	RunServer()
}

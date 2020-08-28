package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	dbconf "github.com/intel/rmd/internal/db/config"
	"github.com/intel/rmd/internal/openstack"
	"github.com/intel/rmd/internal/plugins"
	proxyclient "github.com/intel/rmd/internal/proxy/client"
	proxyserver "github.com/intel/rmd/internal/proxy/server"
	proxytypes "github.com/intel/rmd/internal/proxy/types"
	cacheconf "github.com/intel/rmd/modules/cache/config"
	util "github.com/intel/rmd/utils"
	"github.com/intel/rmd/utils/bootcheck"
	appconf "github.com/intel/rmd/utils/config"
	"github.com/intel/rmd/utils/flag"
	"github.com/intel/rmd/utils/log"
	logconf "github.com/intel/rmd/utils/log/config"
	"github.com/intel/rmd/utils/pidfile"
	"github.com/intel/rmd/utils/pqos"
	"github.com/intel/rmd/utils/proc"
	"github.com/intel/rmd/utils/resctrl"
	"github.com/intel/rmd/version"
	loginfo "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
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

	if os.Geteuid() == 0 {
		// PQOS initialization should be performed only in root process
		if err := pqos.Init(); err != nil {
			fmt.Println("PQOS initialization failed:", err)
			os.Exit(1)
		}
	}

	if err := appconf.Init(); err != nil {
		fmt.Println("Init config failed:", err)
		os.Exit(1)
	}
	if err := log.Init(); err != nil {
		fmt.Println("Init log failed:", err)
		os.Exit(1)
	}

	cfg := appconf.NewConfig()
	pluginsList := strings.Split(cfg.Def.Plugins, ",")
	for _, pluginName := range pluginsList {
		pluginName = strings.Trim(pluginName, " \t")
		if len(pluginName) == 0 {
			continue
		}
		fmt.Println("loading RMD plugin:", pluginName)
		if strings.Trim(pluginName, " ") == "cache" {
			// cache is currently hardcoded - no need to do anything
			continue
		}
		pluginCfg, err := plugins.GetConfig(pluginName)
		if err != nil {
			loginfo.Errorf("Failed to load config for plugin %v with error: %v", pluginName, err.Error())
			exitOnInitError()
		}
		pluginPath, ok := pluginCfg["path"]
		if !ok {
			// should fail here as without path it's not possible to load plugin
			loginfo.Errorf("Unable to load %v plugin. Please check plugin path", pluginName)
			exitOnInitError()
		}
		pluginPathString, ok := pluginPath.(string)
		if !ok {
			loginfo.Errorf("Unable to load %v plugin. Please check plugin path", pluginName)
			exitOnInitError()
		}
		modIface, err := plugins.Load(pluginPathString)
		if err != nil {
			loginfo.Errorf("Failed to load plugin file %v with error: %v", pluginPathString, err.Error())
			exitOnInitError()
		}
		err = modIface.Initialize(pluginCfg)
		if err != nil {
			loginfo.Errorf("Failed to load %v plugin with error: %v", pluginName, err.Error())
			exitOnInitError()
		}

		err = plugins.Store(pluginName, modIface)
		if err != nil {
			loginfo.Errorf("Failed to load %v plugin with error: %v", pluginName, err.Error())
			exitOnInitError()
		}
	}

	if err := proc.Init(); err != nil {
		fmt.Println("proc init failed:", err)
		os.Exit(1)
	}

	// WARNING: resctrl module has to be used as sometimes user process tries to read something fro resctrl fs.
	//          This cannot be replaced with PQOS as PQOS is initialized only in root process
	if err := resctrl.Init(); err != nil {
		fmt.Println("resctrl init failed:", err)
		os.Exit(1)
	}

	if os.Geteuid() == 0 { // root process - compare platform and rmd.toml
		// - check which mode is set in config
		rdtc := cacheconf.RDTConfig{MBAMode: "percentage"} // default value used if not set in config file
		err := viper.UnmarshalKey("rdt", &rdtc)
		if err != nil {
			loginfo.Errorf("Failed to check RDT config in rmd.toml")
			exitOnInitError()
		} else {
			loginfo.Debugf("Configured MBA mode: %v", rdtc.MBAMode)
		}
		mbaInt, err := cacheconf.MBAModeToInt(rdtc.MBAMode)
		if err != nil {
			loginfo.Errorf("Invalid MBA mode in rmd.toml")
			exitOnInitError()
		}
		forceFlagString := pflag.Lookup("force-config").Value.String()
		var forceFlag bool
		if strings.ToLower(forceFlagString) == "true" {
			forceFlag = true
		}
		// Check if MBA is supported on platform and in which mode ONLY WHEN REQUESTED in rmd.toml
		if mbaInt != -1 {
			platformMba, err := pqos.CheckMBA()
			forceReload := false
			if err != nil {
				if !forceFlag {
					loginfo.Errorf("Failed to check MBA mode. Re-launch RMD or launch with '--force-config' to force platform settings")
					exitOnInitError()
				}
				forceReload = true
			}
			if platformMba != mbaInt {
				if !forceFlag {
					loginfo.Errorf("Requested and currently set MBA mode differ." +
						"Change settings in rmd.toml, re-mount resctrl filesystem or use '--force-config' flag to force platform settings")
					exitOnInitError()
				}
				forceReload = true
			}
			if forceReload {
				// reset PQoS (re-mount resctrl) to assure proper MBA mode
				loginfo.Warningf("Resetting PQOS and re-mounting resctrl filesystem")
				if mbaInt == 0 {
					err = pqos.ResetAPI(pqos.MBAPercentage)
				} else {
					err = pqos.ResetAPI(pqos.MBAMbps)
				}
				if err != nil {
					loginfo.Errorf("Failed to reset PQOS with correct MBA mode: %v", err.Error())
					exitOnInitError()
				}
				// check MBA mode after reset
				platformMba, err = pqos.CheckMBA()
				if err != nil {
					loginfo.Errorf("Failed to check MBA mode in platform: %v", err.Error())
					exitOnInitError()
				}
				// compare again platform MBA mode with configured one
				if platformMba != mbaInt {
					loginfo.Errorf("Requested and available MBA mode differ")
					exitOnInitError()
				}
			}
		}
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

		// need to create dir to store db if not exists
		tsDirPath := filepath.Dir(ts)
		if _, err := os.Stat(tsDirPath); os.IsNotExist(err) {
			os.Mkdir(tsDirPath, 0755) //rwxr-xr-x
		}
		// need to set correct ownership to that dir
		if err := util.Chown(tsDirPath, rmduser); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

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

	if err := proxyclient.ConnectRPCServer(in); err != nil {
		loginfo.Println(err)
		os.Exit(1)
	}
	// should go after connect rpc server
	bootcheck.SanityCheck()
	// should tell root process we are fail or success for the santify check
	RunServer()
}

// helper function for cleaner code - prints message about initialization error on stdout and exits application
func exitOnInitError() {
	fmt.Println("RMD initialization error. Check logs for details")
	os.Exit(1)
}

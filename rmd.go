package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/intel/rmd/app"
	dbconf "github.com/intel/rmd/db/config"
	"github.com/intel/rmd/lib/proxy"
	"github.com/intel/rmd/util"
	"github.com/intel/rmd/util/bootcheck"
	"github.com/intel/rmd/util/conf"
	"github.com/intel/rmd/util/flag"
	"github.com/intel/rmd/util/log"
	logconf "github.com/intel/rmd/util/log/config"
	"github.com/intel/rmd/util/pidfile"
)

var rmduser = "rmd"

func main() {
	// use pipe pair to communicate between root and normal process
	var in, out proxy.PipePair
	flag.InitFlags()
	if err := conf.Init(); err != nil {
		fmt.Println("Init config failed:", err)
		os.Exit(1)
	}
	if err := log.Init(); err != nil {
		fmt.Println("Init log failed:", err)
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
			fmt.Println("Failed to drop root priviledge")
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

		fmt.Printf("RMD server started, REST API server serving on process %d\n", child.Pid)
		proxy.RegisterAndServe(out)
	}

	// Below are executed in child process
	flg := syscall.SIGHUP
	if _, _, err := syscall.RawSyscall(syscall.SYS_PRCTL, syscall.PR_SET_PDEATHSIG, uintptr(flg), 0); err != 0 {
		fmt.Println(err)
		os.Exit(1)
	}
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGHUP)
	go func() {
		sig := <-sigc
		//NOTE, should we add some cleanup?
		cleanupFunc()
		fmt.Printf("Received %s, shutdown RMD for root process exit.", sig.String())
		// Do not Exit(0), for there are some thing wrong with supper RMD.
		os.Exit(1)
	}()

	//in.Writer
	in.Writer = os.NewFile(3, "")
	//in.Reader
	in.Reader = os.NewFile(4, "")
	err := proxy.ConnectRPCServer(in)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// should go after connect rpc server
	bootcheck.SanityCheck()
	// should tell root process we are fail or success for the santify check
	app.RunServer()
}

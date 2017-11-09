package pidfile

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// PIDFile Pid file of rmd process
const PIDFile = "/var/run/rmd.pid"

// CreatePID Create pid file and write my pid into that file
func CreatePID() error {
	// FIXME one edge case does not handle, that is the process exist,
	// but pidfile is removed. Beside call shell command "lsof", I have no idea.
	// TODO should we consider owner of pid file?

	// check pid file exist.
	dat, err := ioutil.ReadFile(PIDFile)
	if err == nil {
		pidstr := strings.TrimSpace(string(dat))
		files, _ := filepath.Glob("/proc/" + pidstr)
		if len(files) > 0 {
			return fmt.Errorf("RMD %s is already running, exit", pidstr)
		}
		// From golang doc, os.FindProcess always return successful.
		// err == nil and process.Pid > 0
	}

	flag := os.O_RDWR | os.O_CREATE | os.O_TRUNC
	f, err := os.OpenFile(PIDFile, flag, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	pid := strconv.Itoa(os.Getpid())
	_, err = f.Write([]byte(pid))
	if err != nil {
		return err
	}
	return nil
	// use Flock to check the whether the process exists? Which is better.
	// return Flock(f, 0, true)
}

// ClosePID Force remove pid file
func ClosePID() {
	dat, err := ioutil.ReadFile(PIDFile)
	if err == nil {
		p := strings.TrimSpace(string(dat))
		pid := strconv.Itoa(os.Getpid())
		if p == pid {
			os.Remove(PIDFile)
		}
	}
}

package mba

import (
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/intel/rmd/lib/proc"
)

// GetMbaInfo traverse resctrl MB to get info
var GetMbaInfo = func() (int, int, error) {
	step, min := 0, 0
	dat, err := ioutil.ReadFile(proc.MbaInfoPath + "/bandwidth_gran")
	if err != nil {
		return step, min, err
	}
	step, err = strconv.Atoi(strings.TrimSpace(string(dat)))
	if err != nil {
		return step, min, err
	}

	dat, err = ioutil.ReadFile(proc.MbaInfoPath + "/min_bandwidth")
	if err != nil {
		return step, min, err
	}
	min, err = strconv.Atoi(strings.TrimSpace(string(dat)))
	if err != nil {
		return step, min, err
	}
	return step, min, nil
}

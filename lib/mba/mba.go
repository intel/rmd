package mba

import (
	"bufio"
	"io/ioutil"
	"os"
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

// GetCellNumber return the number of CPU cells
var GetCellNumber = func() (int, error) {
	file, err := os.Open("/sys/fs/resctrl/schemata")
	if err != nil {
		return 0, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	c := 0
	for schemata, err := reader.ReadString('\n'); err == nil; schemata, err = reader.ReadString('\n') {
		if strings.Contains(schemata, "MB") {
			slist := strings.Split(schemata, ";")
			c = len(slist)
			break
		}
	}
	return c, nil
}

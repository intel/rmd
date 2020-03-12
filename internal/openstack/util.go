// + build linux,openstack

package openstack

import (
	"os/exec"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

// getPIDByName finds PID based on part of process cmdline
func getPIDByName(name string) (string, error) {
	out, err := exec.Command("pgrep", "-f", name).Output()
	if err != nil {
		log.Error(err)
		return "", err
	}
	result := strings.Trim(string(out), " \n")
	return result, nil
}

// getPIDTaskSet returns list of pinned cores for given PID
func getPIDTaskSet(pid string) ([]string, error) {
	out, err := exec.Command("taskset", "-p", pid).Output()
	if err != nil {
		log.Error(err)
		return nil, err
	}
	words := strings.Fields(string(out))
	mask := words[len(words)-1]
	mask = strings.ToLower(mask)

	//do not parse the mask if no affinity is set (all letters are f)
	if strings.Count(mask, "f") != len(mask) {
		return parseTasksetMask(mask), nil
	}

	return []string{}, nil
}

// parseTasksetMask returns list of set bit indices for given mask
func parseTasksetMask(mask string) []string {
	ret := []string{}
	l := uint(len(mask)) - 1
	for i := uint(0); i <= l; i++ {
		//start from the end of mask, and parse feach 'hex letter' separately
		val, err := strconv.ParseInt(string(mask[l-i]), 16, 16)
		if err != nil {
			log.Println("Failed to parse mask")
			return []string{}
		}
		for j := uint(0); j < 4; j++ {
			if ((1 << j) & val) != 0 {
				ret = append(ret, strconv.FormatUint(uint64(i*4+j), 10))
			}
		}
	}
	return ret
}

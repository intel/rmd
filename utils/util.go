package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"reflect"
	"strconv"
	"syscall"

	appconf "github.com/intel/rmd/utils/config"
)

// HasElem Helper function to find if a elem in a slice
func HasElem(s interface{}, elem interface{}) bool {
	arrv := reflect.ValueOf(s)
	if arrv.Kind() == reflect.Slice {
		for i := 0; i < arrv.Len(); i++ {
			if arrv.Index(i).Interface() == elem {
				return true
			}
		}
	}
	return false
}

// SubtractStringSlice remove string from slice
func SubtractStringSlice(slice, s []string) []string {
	for _, i := range s {
		for pos, j := range slice {
			if i == j {
				slice = append(slice[:pos], slice[pos+1:]...)
				break
			}
		}
	}
	return slice
}

// IsUserExist check if user exist on host
func IsUserExist(name string) bool {
	_, err := user.Lookup(name)
	if err != nil {
		return false
	}
	return true
}

// CreateUser will create a normal user by name
func CreateUser(name string) error {
	path, err := exec.LookPath("useradd")
	if err != nil {
		return err
	}
	cmd := exec.Command(path, name)
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

// GetUserGUID give user's GUID
func GetUserGUID(name string) (int, int, error) {
	User, err := user.Lookup(name)
	if err != nil {
		return 0, 0, err
	}

	uid, _ := strconv.Atoi(User.Uid)
	gid, _ := strconv.Atoi(User.Gid)
	return uid, gid, nil
}

// Chown changes owner
func Chown(file, user string) error {
	if _, err := os.Stat(file); err == nil {
		uid, gid, err := GetUserGUID(user)
		if err != nil {
			return err
		}
		if err := os.Chown(file, uid, gid); err != nil {
			fmt.Println("Failed to change owner of file:", file)
			return err
		}
	}
	return nil
}

//DropRunAs will drop root previlidge and run as a normal user
func DropRunAs(name string, debug bool, files ...*os.File) (*os.Process, error) {

	if os.Getuid() != 0 {
		return nil, fmt.Errorf("Need to run as root user")
	}
	uid, gid, err := GetUserGUID(name)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(os.Args[0], os.Args[1:]...)
	// ExtraFiles specifies additional open files to be inherited by the
	// new process. It does not include standard input, standard output, or
	// standard error. If non-nil, entry i becomes file descriptor 3+i.
	cmd.ExtraFiles = files
	if debug {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{
			Uid: uint32(uid),
			Gid: uint32(gid),
		},
		Setsid: true,
	}
	if err = cmd.Start(); err != nil {
		return nil, err
	}
	//cmd.Process.Release()
	return cmd.Process, nil
}

// IsRegularFile checks if given path point to a file - not to symlink nor directory
func IsRegularFile(path string) (bool, error) {
	// try to call 'stat' - if failed then probably file does not exists
	finfo, err := os.Lstat(path)
	if err != nil {
		return false, err
	}

	// check if is dir
	if finfo.IsDir() {
		return false, nil
	}

	// check if is regular file (will return true also for hardlink)
	fmode := finfo.Mode()
	if !fmode.IsRegular() {
		return false, nil
	}

	// no problem found with path - return true
	return true, nil
}

// UnifyMapParamsTypes responsible to unify types of parameters in map
// to avoid misleading types when passing values to plugins
// Example: 3.0 can be json.Number but should be float
// RMD plugins expects following type of params:
// => int64 (for integers)
// => float64 (for real numbers)
// => bool
// => string
// => nil
// => []float64
// => []int64
// => interface{}
func UnifyMapParamsTypes(pluginParams map[string]interface{}) (map[string]interface{}, error) {

	unifiedPluginParams := make(map[string]interface{})

	// error for all when at least one conversion from loop was failed
	for paramNameAsString, paramValueAsInterface := range pluginParams {

		switch v := paramValueAsInterface.(type) {

		case json.Number: // Number in JSON is used for any numeric type, either integers or floating point numbers
			paramValueInt64, err := v.Int64()
			if err == nil {
				unifiedPluginParams[paramNameAsString] = paramValueInt64
				break
			}

			paramValueFloat64, err := v.Float64()
			if err == nil {
				unifiedPluginParams[paramNameAsString] = paramValueFloat64
				break
			}

			return map[string]interface{}{}, errors.New("Failed to convert from json.Number")

		case float32:
			paramValueFloat64 := float64(v)
			unifiedPluginParams[paramNameAsString] = paramValueFloat64
			break

		case int:
		case int32:
			paramValueInt64 := int64(v)
			unifiedPluginParams[paramNameAsString] = paramValueInt64
			break

		case []int:
		case []int32:
			tempTable := []int64{}
			for _, value := range v {
				singleElem := int64(value)
				tempTable = append(tempTable, singleElem)
			}

			unifiedPluginParams[paramNameAsString] = tempTable
			break

		case []float32:
			tempTable := []float64{}
			for _, value := range v {
				singleElem := float64(value)
				tempTable = append(tempTable, singleElem)
			}
			unifiedPluginParams[paramNameAsString] = tempTable
			break

		case nil:
			unifiedPluginParams[paramNameAsString] = paramValueAsInterface
			break
		case bool:
			unifiedPluginParams[paramNameAsString] = paramValueAsInterface
			break
		case string:
			unifiedPluginParams[paramNameAsString] = paramValueAsInterface
			break
		case []int64:
			unifiedPluginParams[paramNameAsString] = paramValueAsInterface
			break
		case []float64:
			unifiedPluginParams[paramNameAsString] = paramValueAsInterface
			break
		case interface{}:
			unifiedPluginParams[paramNameAsString] = paramValueAsInterface
			break
		case []interface{}:
			unifiedPluginParams[paramNameAsString] = paramValueAsInterface
			break

		default:
			return map[string]interface{}{}, errors.New("Failed to unify plugin params - unknown param type")
		}
	}
	return unifiedPluginParams, nil
}

//GetDbValidatorInterval getter for db validator interval
func GetDbValidatorInterval() int {
	currentCfg := appconf.NewConfig()
	return int(currentCfg.Def.DbValidatorInterval)
}

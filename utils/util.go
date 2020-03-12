package util

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"reflect"
	"strconv"
	"syscall"
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

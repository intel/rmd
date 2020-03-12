package flock

import (
	"fmt"
	"os"
	"syscall"
	"time"
)

// Flock acquires an advisory lock on a file descriptor.
func Flock(file *os.File, timeout time.Duration, exclusive ...bool) error {

	lockStates := map[bool]int{true: syscall.LOCK_EX, false: syscall.LOCK_SH}
	flag := syscall.LOCK_SH
	if len(exclusive) > 0 {
		flag = lockStates[exclusive[0]]
	}

	s := time.Now()
	t := s
	// timeout < 0 means loop forever. timeout = 0 means just once.
	if timeout > 0 {
		t = s.Add(time.Duration(timeout))
	}

	// A Duration represents the elapsed time between two instants as an int64 nanosecond count.
	// The representation limits the largest representable duration to approximately 290 years.
	// So here we use time Before/After
	for time.Duration(timeout) <= 0 || s.Before(t) {
		// Otherwise attempt to obtain an exclusive lock.
		err := syscall.Flock(int(file.Fd()), flag|syscall.LOCK_NB)
		if timeout == 0 {
			return err
		}
		if err == syscall.EWOULDBLOCK {
			// Wait for a bit and try again.
			time.Sleep(time.Millisecond * 50)
			s = time.Now()
		} else {
			return err
		}
	}

	// FIXME uniform error.
	return fmt.Errorf("Timeout to get flock")
}

// Funlock releases an advisory lock on a file descriptor.
func Funlock(file *os.File) error {
	return syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
}

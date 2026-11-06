//go:build !windows

package save

import (
	"golang.org/x/sys/unix"
	"os"
)

func acquirePlatformLock(file *os.File) (any, bool, error) {
	err := unix.Flock(int(file.Fd()), unix.LOCK_EX|unix.LOCK_NB)
	if err == unix.EWOULDBLOCK || err == unix.EAGAIN {
		return nil, true, nil
	}
	return nil, false, err
}

func releasePlatformLock(file *os.File, _ any) error {
	return unix.Flock(int(file.Fd()), unix.LOCK_UN)
}

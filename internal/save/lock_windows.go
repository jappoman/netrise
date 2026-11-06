//go:build windows

package save

import (
	"golang.org/x/sys/windows"
	"os"
)

func acquirePlatformLock(file *os.File) (any, bool, error) {
	overlapped := &windows.Overlapped{}
	err := windows.LockFileEx(windows.Handle(file.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, 0, 1, 0, overlapped)
	if err == windows.ERROR_LOCK_VIOLATION {
		return nil, true, nil
	}
	return overlapped, false, err
}

func releasePlatformLock(file *os.File, state any) error {
	overlapped, ok := state.(*windows.Overlapped)
	if !ok {
		return nil
	}
	return windows.UnlockFileEx(windows.Handle(file.Fd()), 0, 1, 0, overlapped)
}

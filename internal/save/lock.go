package save

import (
	"errors"
	"fmt"
	"github.com/jappoman/netrise/internal/game"
	"os"
	"path/filepath"
)

var ErrSessionLocked = errors.New("campaign is already running")

type SessionLock struct {
	file  *os.File
	state any
}

func AcquireSessionLock(dir string, faction game.Faction) (*SessionLock, error) {
	if faction != game.Runner && faction != game.Corporation {
		return nil, fmt.Errorf("unknown faction %q", faction)
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	path := filepath.Join(dir, string(faction)+".lock")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	state, busy, err := acquirePlatformLock(file)
	if err != nil {
		_ = file.Close()
		return nil, err
	}
	if busy {
		_ = file.Close()
		return nil, ErrSessionLocked
	}
	return &SessionLock{file: file, state: state}, nil
}

func (l *SessionLock) Close() error {
	if l == nil || l.file == nil {
		return nil
	}
	unlockErr := releasePlatformLock(l.file, l.state)
	closeErr := l.file.Close()
	l.file = nil
	if unlockErr != nil {
		return unlockErr
	}
	return closeErr
}

package save

import (
	"errors"
	"github.com/jappoman/netrise/internal/game"
	"testing"
)

func TestSessionLockRejectsDuplicateAndReleases(t *testing.T) {
	dir := t.TempDir()
	first, err := AcquireSessionLock(dir, game.Runner)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	if _, err = AcquireSessionLock(dir, game.Runner); !errors.Is(err, ErrSessionLocked) {
		t.Fatalf("duplicate Runner session was not rejected: %v", err)
	}
	corporation, err := AcquireSessionLock(dir, game.Corporation)
	if err != nil {
		t.Fatalf("independent Corporation session was blocked: %v", err)
	}
	defer corporation.Close()
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	second, err := AcquireSessionLock(dir, game.Runner)
	if err != nil {
		t.Fatalf("Runner lock was not released: %v", err)
	}
	if err := second.Close(); err != nil {
		t.Fatal(err)
	}
}

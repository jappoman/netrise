package save

import (
	"bufio"
	"errors"
	"github.com/jappoman/netrise/internal/game"
	"io"
	"os"
	"os/exec"
	"testing"
)

func TestSessionLockHelper(t *testing.T) {
	if os.Getenv("NETRISE_LOCK_HELPER") != "1" {
		return
	}
	lock, err := AcquireSessionLock(os.Getenv("NETRISE_LOCK_DIR"), game.Runner)
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Close()
	_, _ = os.Stdout.WriteString("locked\n")
	_, _ = io.Copy(io.Discard, os.Stdin)
}

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

func TestSessionLockRejectsSeparateProcess(t *testing.T) {
	dir := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=TestSessionLockHelper", "--")
	cmd.Env = append(os.Environ(), "NETRISE_LOCK_HELPER=1", "NETRISE_LOCK_DIR="+dir)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	locked := bufio.NewScanner(stdout)
	if !locked.Scan() || locked.Text() != "locked" {
		_ = stdin.Close()
		_ = cmd.Wait()
		t.Fatalf("lock helper did not acquire session lock: %q", locked.Text())
	}
	if second, err := AcquireSessionLock(dir, game.Runner); !errors.Is(err, ErrSessionLocked) || second != nil {
		_ = stdin.Close()
		_ = cmd.Wait()
		t.Fatalf("separate process lock was not enforced: lock=%v err=%v", second, err)
	}
	if err := stdin.Close(); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Wait(); err != nil {
		t.Fatalf("lock helper did not exit cleanly: %v", err)
	}
}

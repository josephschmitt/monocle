package adapters

import (
	"bufio"
	"net"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// fakeServer pretends to be a monocle serve — just listens on a unix socket.
// We use it to exercise the "socket already alive" path.
func fakeServer(t *testing.T, socketPath string) (stop func()) {
	t.Helper()
	l, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("fake listen: %v", err)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			// Echo newline so Dial completes cleanly.
			go func() {
				defer conn.Close()
				_ = bufio.NewScanner(conn)
			}()
		}
	}()
	return func() {
		_ = l.Close()
		wg.Wait()
		_ = os.Remove(socketPath)
	}
}

func TestEnsureServe_ReusesExisting(t *testing.T) {
	dir := t.TempDir()
	sockPath := filepath.Join(dir, "sock")
	stop := fakeServer(t, sockPath)
	defer stop()

	got, spawned, err := EnsureServe(AutoSpawnOptions{Socket: sockPath})
	if err != nil {
		t.Fatalf("ensure serve: %v", err)
	}
	if got != sockPath {
		t.Errorf("socket = %q, want %q", got, sockPath)
	}
	if spawned {
		t.Error("should NOT have spawned when socket alive")
	}
}

func TestSocketAlive(t *testing.T) {
	dir := t.TempDir()
	sockPath := filepath.Join(dir, "sock")

	// Nothing listening → not alive
	if socketAlive(sockPath) {
		t.Error("expected not alive for missing socket")
	}

	// Start a listener → alive
	l, err := net.Listen("unix", sockPath)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer l.Close()
	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				return
			}
			c.Close()
		}
	}()
	// Give the listener a beat to become ready
	time.Sleep(10 * time.Millisecond)

	if !socketAlive(sockPath) {
		t.Error("expected alive for bound socket")
	}

	// Close → stale socket file remains but dial fails → not alive
	l.Close()
	if socketAlive(sockPath) {
		t.Error("expected not alive after listener closed")
	}
}

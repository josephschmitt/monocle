package core

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/josephschmitt/monocle/internal/protocol"
)

// TestIdleShutdown_DisabledByZero verifies that idle shutdown is a no-op when
// the timeout is zero — the monitor goroutine never starts.
func TestIdleShutdown_DisabledByZero(t *testing.T) {
	engine, _ := setupTestEngine(t)
	engine.SetIdleTimeout(0)

	select {
	case <-engine.IdleShutdownCh():
		t.Fatal("idle shutdown fired with IdleTimeout=0")
	case <-time.After(100 * time.Millisecond):
		// expected
	}
}

// TestIdleShutdown_FiresAfterGrace drives a short tick interval + short
// grace so we can assert the shutdown fires within the test budget.
func TestIdleShutdown_FiresAfterGrace(t *testing.T) {
	engine := newTestEngine(t)
	// Configure test-only tick/grace BEFORE StartServer; monitor reads them once.
	engine.server.idleTickInterval = 20 * time.Millisecond
	engine.server.idleGrace = 20 * time.Millisecond
	engine.SetIdleTimeout(20 * time.Millisecond)

	socketPath := testSocketPath(t)
	if err := engine.StartServer(socketPath); err != nil {
		t.Fatalf("start server: %v", err)
	}
	t.Cleanup(func() { engine.Shutdown() })

	// Nothing ever connects → lastDisconnectAt stays zero → monitor must
	// NOT fire. Verify that first.
	select {
	case <-engine.IdleShutdownCh():
		t.Fatal("idle shutdown fired before any client connected (last=zero)")
	case <-time.After(150 * time.Millisecond):
		// expected — no connection ever → no idle start
	}

	// One-shot request → active briefly, then zero → monitor should fire.
	dialRequestAndClose(t, socketPath)

	select {
	case <-engine.IdleShutdownCh():
		// success
	case <-time.After(2 * time.Second):
		t.Fatal("idle shutdown did not fire after client disconnected")
	}
}

// TestIdleShutdown_ResetsOnReconnect verifies that a new client clears
// lastDisconnectAt so the idle countdown restarts on the next full drain.
func TestIdleShutdown_ResetsOnReconnect(t *testing.T) {
	engine := newTestEngine(t)
	engine.server.idleTickInterval = 20 * time.Millisecond
	engine.server.idleGrace = 20 * time.Millisecond
	engine.SetIdleTimeout(500 * time.Millisecond) // longer than our reconnect window

	socketPath := testSocketPath(t)
	if err := engine.StartServer(socketPath); err != nil {
		t.Fatalf("start: %v", err)
	}
	t.Cleanup(func() { engine.Shutdown() })

	// First disconnect → starts countdown.
	dialRequestAndClose(t, socketPath)
	time.Sleep(40 * time.Millisecond)

	engine.server.subscriberMu.Lock()
	firstDisc := engine.server.lastDisconnectAt
	engine.server.subscriberMu.Unlock()
	if firstDisc.IsZero() {
		t.Fatal("expected lastDisconnectAt set after first disconnect")
	}

	// Reconnect mid-countdown — clears lastDisconnectAt.
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	time.Sleep(40 * time.Millisecond) // let onConnect run
	engine.server.subscriberMu.Lock()
	cleared := engine.server.lastDisconnectAt.IsZero()
	engine.server.subscriberMu.Unlock()
	conn.Close()

	if !cleared {
		t.Error("expected lastDisconnectAt cleared after reconnect")
	}
}

// --- helpers ---

// newTestEngine mirrors setupTestEngine but without starting the server.
// Tests that need to configure idle knobs before StartServer call this.
func newTestEngine(t *testing.T) *Engine {
	t.Helper()
	// Reuse setupTestEngine but immediately shut down the server it started
	// so we can reconfigure and restart with custom idle knobs.
	engine, _ := setupTestEngine(t)
	_ = engine.server.Shutdown()
	// Clear out the idleStop channel — Shutdown closed it. Replace with a
	// fresh pair so StartServer can restart the monitor loop cleanly.
	engine.server.idleStop = make(chan struct{})
	engine.server.shutdownCh = make(chan struct{})
	engine.server.listener = nil
	engine.server.socketPath = ""
	return engine
}

func testSocketPath(t *testing.T) string {
	t.Helper()
	// macOS caps unix socket paths at ~104 bytes; t.TempDir() paths are
	// usually too long. Use /tmp + a short hash of the test name.
	hash := sha256.Sum256([]byte(t.Name()))
	return fmt.Sprintf("/tmp/monocle-idle-%s.sock", hex.EncodeToString(hash[:])[:10])
}

func dialRequestAndClose(t *testing.T, socketPath string) {
	t.Helper()
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	msg := protocol.GetReviewStatusMsg{Type: protocol.TypeGetReviewStatus}
	data, _ := protocol.Encode(&msg)
	if _, err := conn.Write(data); err != nil {
		t.Fatalf("write: %v", err)
	}
	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 1<<20), 1<<20)
	_ = scanner.Scan()
	conn.Close()
}

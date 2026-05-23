package adapters

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"time"
)

// AutoSpawnOptions controls how EnsureServe locates and spawns monocle serve.
type AutoSpawnOptions struct {
	// RepoRoot is the repository the engine should attach to. Autospawn
	// passes this as `-C <repoRoot>` to the child serve process and uses
	// DefaultSocketPath(repoRoot) when Socket is empty.
	RepoRoot string

	// Socket, when non-empty, overrides DefaultSocketPath(RepoRoot).
	Socket string

	// Binary is the monocle binary to exec. Defaults to the currently
	// running executable (os.Executable), which keeps `go run ./cmd/...`
	// and installed binaries behaving correctly.
	Binary string

	// ReadyTimeout bounds how long EnsureServe waits for the spawned
	// serve to start listening. Defaults to 2s.
	ReadyTimeout time.Duration

	// ProbeInterval is how often EnsureServe retries Dial between spawn
	// and readiness. Defaults to 50ms.
	ProbeInterval time.Duration
}

// EnsureServe probes the socket for the given repo. If a serve is already
// listening, it returns immediately. Otherwise it spawns `monocle serve -C
// <repoRoot>` detached, polls for readiness, and returns the socket path
// once the engine is accepting connections.
//
// The child process is detached from the launching terminal (via the
// platform-specific helper detachChildProcess) so it outlives the TUI —
// closing the frontend doesn't kill the engine, and another frontend can
// attach next time.
func EnsureServe(opts AutoSpawnOptions) (socketPath string, spawned bool, err error) {
	socketPath = opts.Socket
	if socketPath == "" {
		if opts.RepoRoot == "" {
			return "", false, errors.New("autospawn: RepoRoot or Socket required")
		}
		socketPath = DefaultSocketPath(opts.RepoRoot)
	}

	if socketAlive(socketPath) {
		return socketPath, false, nil
	}

	// Stale socket file (leftover from a crashed serve) — remove it so the
	// child process can bind cleanly. `monocle serve` does the same on
	// start, but racing on this is harmless.
	_ = os.Remove(socketPath)

	binary := opts.Binary
	if binary == "" {
		exe, err := os.Executable()
		if err != nil {
			return "", false, fmt.Errorf("autospawn: resolve binary: %w", err)
		}
		binary = exe
	}

	args := []string{"serve"}
	if opts.RepoRoot != "" {
		args = append(args, "-C", opts.RepoRoot)
	}
	if opts.Socket != "" {
		args = append(args, "--socket", opts.Socket)
	}

	cmd := exec.Command(binary, args...)
	detachChildProcess(cmd)
	// Detach stdio so the child doesn't hold the parent's terminal.
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return "", false, fmt.Errorf("autospawn: start serve: %w", err)
	}
	// Release so we don't leave a zombie if the child exits while we're
	// polling. The serve is expected to long-outlive us.
	if err := cmd.Process.Release(); err != nil {
		return "", true, fmt.Errorf("autospawn: release child: %w", err)
	}

	timeout := opts.ReadyTimeout
	if timeout == 0 {
		timeout = 2 * time.Second
	}
	interval := opts.ProbeInterval
	if interval == 0 {
		interval = 50 * time.Millisecond
	}

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if socketAlive(socketPath) {
			return socketPath, true, nil
		}
		time.Sleep(interval)
	}
	return socketPath, true, fmt.Errorf("autospawn: serve did not become ready within %s", timeout)
}

// socketAlive reports whether socketPath is currently accepting connections.
// A stale socket file left by a crashed serve returns false because Dial
// fails against it.
func socketAlive(socketPath string) bool {
	if _, err := os.Stat(socketPath); errors.Is(err, os.ErrNotExist) {
		return false
	}
	conn, err := net.DialTimeout("unix", socketPath, 250*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

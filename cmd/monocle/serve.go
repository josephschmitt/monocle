package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/josephschmitt/monocle/internal/adapters"
	"github.com/josephschmitt/monocle/internal/core"
	"github.com/josephschmitt/monocle/internal/db"
)

// ServeCmd runs a headless engine + socket server. Frontends (TUI, Desktop,
// future plugins) connect as thin socket clients instead of embedding their
// own Engine.
type ServeCmd struct {
	WorkDirFlag
	Socket string `help:"Override socket path" env:"MONOCLE_SOCKET" default:""`
}

// StopCmd sends SIGTERM to a running `monocle serve` process for the target
// repo, if any, and waits for it to exit.
type StopCmd struct {
	WorkDirFlag
	Socket  string        `help:"Override socket path" env:"MONOCLE_SOCKET" default:""`
	Timeout time.Duration `help:"Maximum time to wait for the server to exit" default:"5s"`
}

// pidFilePath returns the PID file path that pairs with a given socket path.
// The socket at /tmp/monocle-<hash>.sock pairs with /tmp/monocle-<hash>.pid.
func pidFilePath(socketPath string) string {
	if strings.HasSuffix(socketPath, ".sock") {
		return strings.TrimSuffix(socketPath, ".sock") + ".pid"
	}
	return socketPath + ".pid"
}

func writePIDFile(path string) error {
	return os.WriteFile(path, []byte(strconv.Itoa(os.Getpid())), 0o644)
}

func readPIDFile(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, fmt.Errorf("parse pid: %w", err)
	}
	return pid, nil
}

func removePIDFile(path string) {
	_ = os.Remove(path)
}

// Run launches the headless engine and blocks on SIGINT/SIGTERM.
func (c *ServeCmd) Run() error {
	cfg, err := core.LoadConfig()
	if err != nil {
		cfg = core.DefaultConfig()
	}

	database, err := db.Open(db.DBPath())
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer database.Close()

	repoRoot, nonGitMode, err := resolveRepoRoot(c.WorkDir)
	if err != nil {
		return err
	}

	socketPath := c.Socket
	if socketPath == "" {
		socketPath = adapters.DefaultSocketPath(repoRoot)
	}

	// Refuse to start if another serve already holds the socket. The
	// SocketServer.Start path does this too, but probing here gives a
	// cleaner error before we allocate an engine.
	if conn, err := net.Dial("unix", socketPath); err == nil {
		conn.Close()
		return fmt.Errorf("monocle serve already running for this repo (socket %s in use)", socketPath)
	}

	engine, err := core.NewEngine(cfg, database, repoRoot, nonGitMode)
	if err != nil {
		return fmt.Errorf("create engine: %w", err)
	}

	// Resolve an initial session the same way runTUI does today: continue
	// the latest session if any, otherwise start fresh. `monocle serve`
	// has no picker UI, so `--resume` and `--session` variants stay with
	// the `monocle` launcher.
	if nonGitMode {
		if err := startNewSession(engine, repoRoot); err != nil {
			return err
		}
	} else if err := resolveSession(engine, repoRoot, true /* continue */, false, ""); err != nil {
		return err
	}
	engine.ReloadPendingFeedback()

	if err := engine.StartServer(socketPath); err != nil {
		return fmt.Errorf("start server: %w", err)
	}

	pidPath := pidFilePath(socketPath)
	if err := writePIDFile(pidPath); err != nil {
		engine.Shutdown()
		return fmt.Errorf("write pid file: %w", err)
	}
	defer removePIDFile(pidPath)

	fmt.Fprintf(os.Stdout, "monocle serve: listening on %s (pid %d)\n", socketPath, os.Getpid())

	// Block on SIGINT/SIGTERM.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	engine.Shutdown()
	return nil
}

// Run signals a running serve process to exit and waits for it to close the
// PID file. No-op when no PID file exists.
func (c *StopCmd) Run() error {
	socketPath, err := resolveSocketForWorkDir(c.Socket, c.WorkDir)
	if err != nil {
		return err
	}
	pidPath := pidFilePath(socketPath)

	pid, err := readPIDFile(pidPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Fprintln(os.Stderr, "monocle stop: no server running")
			return nil
		}
		return err
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("find process %d: %w", pid, err)
	}
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		if errors.Is(err, os.ErrProcessDone) || strings.Contains(err.Error(), "process already finished") {
			removePIDFile(pidPath)
			return nil
		}
		return fmt.Errorf("signal %d: %w", pid, err)
	}

	// Poll until the PID file disappears (serve removes it on exit) or we
	// exceed the caller's timeout.
	deadline := time.Now().Add(c.Timeout)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(pidPath); errors.Is(err, os.ErrNotExist) {
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return fmt.Errorf("timed out waiting for monocle serve (pid %d) to exit", pid)
}


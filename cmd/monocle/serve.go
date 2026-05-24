package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"runtime"
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
	Socket      string        `help:"Override socket path" env:"MONOCLE_SOCKET" default:""`
	IdleTimeout time.Duration `help:"Exit after this idle interval past the 60s grace window (0 disables)" name:"idle-timeout"`
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

// processBasename returns the basename of the first argv element of pid,
// or "" if it can't determine it. Splits into the platform-specific helper
// (pidProcessBasename) to keep Windows working without /proc or `ps`.
func processBasename(pid int) string {
	return pidProcessBasename(pid)
}

// pidIsAlive reports whether a process with the given pid currently
// exists. Used by StopCmd to distinguish "pid is gone, pid file is
// stale" from "pid is alive but we can't identify it".
func pidIsAlive(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// On Unix os.FindProcess always succeeds; the canonical liveness
	// probe is signal 0, which performs the kernel's permission/existence
	// check without actually sending a signal. On Windows FindProcess
	// returns an error for dead pids, so reaching this point already
	// implies alive.
	if runtime.GOOS == "windows" {
		return true
	}
	return proc.Signal(syscall.Signal(0)) == nil
}

// pidLooksLikeMonocle reports whether pid's process image is plausibly a
// monocle binary, so StopCmd doesn't SIGTERM an unrelated process after
// a crashed serve leaves a stale .pid file.
//
// Matching policy: the basename of the executable must begin with
// "monocle" (case-insensitive on Windows). That covers stock installs
// (`monocle`), versioned/renamed binaries (`monocle-v0.46.1`,
// `monocle-dev`), `go test ./cmd/monocle` flows (binary `monocle.test`),
// and `dlv exec ./monocle` (basename `monocle`). A bare prefix check
// would still false-positive on `monocle-unrelated-tool`, but matching
// monocle's own variants is more important than excluding that very
// unlikely case — and StopCmd will only ever signal a pid the user
// previously asked monocle serve to record.
//
// We deliberately do NOT match on a substring scan of the full cmdline:
// that would falsely include `vim monocle.go`, `sudo monocle ...`,
// `bash -c 'monocle ...'`, `grep monocle`, and any other process whose
// argv merely mentions the string.
func pidLooksLikeMonocle(pid int) bool {
	base := processBasename(pid)
	if base == "" {
		return false
	}
	// Trim a common .exe suffix on Windows.
	trimmed := strings.TrimSuffix(base, ".exe")
	if runtime.GOOS == "windows" {
		trimmed = strings.ToLower(strings.TrimSuffix(strings.ToLower(base), ".exe"))
	}
	return strings.HasPrefix(trimmed, "monocle")
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

	// Resolve idle timeout precedence: explicit flag > config file > default.
	// A negative flag value disables idle shutdown entirely.
	idle := core.DefaultIdleTimeout
	if cfg.IdleTimeout != 0 {
		idle = time.Duration(cfg.IdleTimeout)
	}
	if c.IdleTimeout != 0 {
		idle = c.IdleTimeout
	}
	if idle > 0 {
		engine.SetIdleTimeout(idle)
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

	// Block on SIGINT/SIGTERM or the idle-shutdown signal, whichever
	// fires first.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-sig:
	case <-engine.IdleShutdownCh():
		fmt.Fprintln(os.Stdout, "monocle serve: idle timeout reached, exiting")
	}

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

	// Verify the PID actually belongs to a monocle process before
	// signalling. A crashed serve leaves the .pid file behind, and the
	// kernel may have reassigned that PID to an unrelated process
	// (editor, ssh, system daemon) — SIGTERM'ing it would silently kill
	// something innocent.
	//
	// IMPORTANT: only remove the pid file when we can prove the
	// recorded process is dead (no such pid OR the pid belongs to a
	// clearly different binary). If the basename guard just returns
	// "I'm not sure" — empty basename, /proc read denied, ps timed out
	// — we leave the pid file alone so a retry can still find the
	// running daemon. Removing it on every uncertain match was the bug:
	// it orphaned the live daemon for any non-stock binary name (renames,
	// `monocle.test`, `dlv exec`, etc.) and the user had no recourse
	// short of pkill.
	base := processBasename(pid)
	if base == "" {
		// Couldn't determine. Assume the pid is gone (most likely cause)
		// AND that the pid file is stale enough to clear; if the daemon
		// is actually alive the OS will refuse our subsequent signal
		// with ESRCH and we'll clean up there.
		if !pidIsAlive(pid) {
			removePIDFile(pidPath)
			return fmt.Errorf("pid %d in %s is not running; cleaned up stale pid file", pid, pidPath)
		}
		// Pid exists but we can't identify it. Refuse to signal AND
		// leave the pid file alone — better to ask the user to confirm
		// than to silently kill a stranger or silently orphan our own
		// daemon.
		return fmt.Errorf(
			"pid %d in %s is running but its image could not be identified; "+
				"if this is monocle serve, remove %s and retry, otherwise leave it alone",
			pid, pidPath, pidPath,
		)
	}
	if !pidLooksLikeMonocle(pid) {
		// We DID identify the image and it's clearly not us. The pid
		// got reassigned; safe to clean up the stale pid file.
		removePIDFile(pidPath)
		return fmt.Errorf(
			"pid %d in %s is %q, not monocle serve; cleaned up stale pid file",
			pid, pidPath, base,
		)
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


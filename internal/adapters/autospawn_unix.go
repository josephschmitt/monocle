//go:build !windows

package adapters

import (
	"os"
	"os/exec"
	"syscall"
)

// detachChildProcess places the spawned monocle serve in a new session so a
// SIGHUP / Ctrl-C in the launching terminal doesn't kill it.
func detachChildProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}

// openLogFile creates the autospawn stderr log file safely:
//   - O_EXCL means the open fails if anything exists at logPath (so a
//     hostile symlink restored after our Remove can't redirect us).
//   - O_NOFOLLOW means even a still-present symlink at the leaf is
//     refused rather than traversed.
//
// Together they shut the TOCTOU window in shared /tmp scenarios where
// the socket path is predictable.
func openLogFile(logPath string) (*os.File, error) {
	return os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_EXCL|syscall.O_NOFOLLOW, 0o600)
}

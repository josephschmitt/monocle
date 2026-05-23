//go:build !windows

package adapters

import (
	"os/exec"
	"syscall"
)

// detachChildProcess places the spawned monocle serve in a new session so a
// SIGHUP / Ctrl-C in the launching terminal doesn't kill it.
func detachChildProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}

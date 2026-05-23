//go:build windows

package adapters

import (
	"os/exec"
	"syscall"
)

// detachChildProcess starts the child in its own process group so closing the
// console window (Ctrl-C, parent exit) doesn't propagate a CTRL_C_EVENT to it.
// Windows has no equivalent of setsid; the new process group is the closest
// analogue and is what most long-lived "daemon-style" child processes use.
func detachChildProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
}

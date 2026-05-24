//go:build windows

package adapters

import (
	"os"
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

// openLogFile creates the autospawn stderr log file. Windows has no
// O_NOFOLLOW (NTFS reparse-point traversal works differently), so we
// rely on O_EXCL to refuse opening if anything exists at logPath — the
// Remove earlier in the caller cleared any stale file, so a fresh open
// can only succeed against a fresh inode we just made room for.
func openLogFile(logPath string) (*os.File, error) {
	return os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o600)
}

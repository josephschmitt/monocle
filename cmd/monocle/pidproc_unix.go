//go:build !windows

package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// pidProcessBasename returns the basename of the executable backing pid.
// Linux: reads /proc/<pid>/exe (or argv[0] from /proc/<pid>/cmdline) so
// the StopCmd guard matches on the actual binary name, not a substring.
// macOS/BSD: shells out to `ps -p <pid> -o comm=` with a hard 1s timeout
// so a hung ps can't wedge `monocle stop`.
func pidProcessBasename(pid int) string {
	// Linux fast path: symlink-resolve /proc/<pid>/exe.
	if target, err := os.Readlink(fmt.Sprintf("/proc/%d/exe", pid)); err == nil {
		return filepath.Base(target)
	}
	// Linux fallback: argv[0] from cmdline.
	if data, err := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid)); err == nil && len(data) > 0 {
		// NUL-separated fields; argv[0] is the first.
		argv0 := string(data)
		if idx := strings.IndexByte(argv0, '\x00'); idx >= 0 {
			argv0 = argv0[:idx]
		}
		if argv0 != "" {
			return filepath.Base(argv0)
		}
	}
	// macOS / BSD fallback: ps -p <pid> -o comm=, bounded.
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "ps", "-p", strconv.Itoa(pid), "-o", "comm=")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	comm := strings.TrimSpace(string(out))
	if comm == "" {
		return ""
	}
	return filepath.Base(comm)
}

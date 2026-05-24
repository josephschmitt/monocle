//go:build windows

package main

import (
	"context"
	"encoding/csv"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// pidProcessBasename returns the executable name for pid on Windows.
// Windows has no /proc and bare `ps` is not on the default PATH, so we
// invoke tasklist.exe in CSV mode and parse the image name. Bounded by a
// 1s context so a hung tasklist can't wedge `monocle stop`.
func pidProcessBasename(pid int) string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	// /FI "PID eq N" filters; /FO CSV /NH gives parseable output with no header.
	cmd := exec.CommandContext(ctx, "tasklist.exe",
		"/FI", "PID eq "+strconv.Itoa(pid),
		"/FO", "CSV", "/NH")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	r := csv.NewReader(strings.NewReader(strings.TrimSpace(string(out))))
	record, err := r.Read()
	if err != nil || len(record) == 0 {
		return ""
	}
	return strings.TrimSpace(record[0])
}

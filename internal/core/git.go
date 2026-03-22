package core

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/anthropics/monocle/internal/types"
)

// GitClient wraps git operations for a repository.
type GitClient struct {
	repoRoot string
}

// NewGitClient creates a GitClient for the given repo root.
func NewGitClient(repoRoot string) *GitClient {
	return &GitClient{repoRoot: repoRoot}
}

// RepoRoot returns the repository root path.
func (g *GitClient) RepoRoot() string {
	return g.repoRoot
}

// Diff returns the list of changed files between baseRef and the working tree.
func (g *GitClient) Diff(baseRef string) ([]types.ChangedFile, error) {
	out, err := g.run("diff", "--name-status", baseRef)
	if err != nil {
		return nil, fmt.Errorf("git diff --name-status: %w", err)
	}

	var files []types.ChangedFile
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) < 2 {
			continue
		}

		status := parseFileStatus(parts[0])
		path := parts[1]
		// Handle renames: R100\told\tnew
		if strings.HasPrefix(parts[0], "R") && len(parts) > 1 {
			tabParts := strings.SplitN(line, "\t", 3)
			if len(tabParts) == 3 {
				path = tabParts[2]
			}
		}

		files = append(files, types.ChangedFile{
			Path:   path,
			Status: status,
		})
	}
	return files, nil
}

// FileDiff returns the parsed diff for a single file.
// contextLines controls the number of unchanged lines around each hunk (-U flag).
// A value of 0 or less uses git's default (3).
func (g *GitClient) FileDiff(baseRef, path string, contextLines int) (*types.DiffResult, error) {
	args := []string{"diff"}
	if contextLines > 0 {
		args = append(args, fmt.Sprintf("-U%d", contextLines))
	}
	args = append(args, baseRef, "--", path)
	out, err := g.run(args...)
	if err != nil {
		// diff returns exit 1 when there are differences, which is expected
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			// Use the output even though exit code was 1
		} else {
			return nil, fmt.Errorf("git diff %s: %w", path, err)
		}
	}

	return &types.DiffResult{
		Path:  path,
		Hunks: parseDiff(out),
	}, nil
}

// FileContent returns file content at a given ref, or from the working tree if ref is empty.
func (g *GitClient) FileContent(ref, path string) (string, error) {
	if ref == "" || ref == "WORKING" {
		absPath := filepath.Join(g.repoRoot, path)
		out, err := exec.Command("cat", absPath).Output()
		if err != nil {
			return "", fmt.Errorf("read file %s: %w", path, err)
		}
		return string(out), nil
	}
	out, err := g.run("show", ref+":"+path)
	if err != nil {
		return "", fmt.Errorf("git show %s:%s: %w", ref, path, err)
	}
	return out, nil
}

// CurrentRef returns the current HEAD commit hash.
func (g *GitClient) CurrentRef() (string, error) {
	out, err := g.run("rev-parse", "HEAD")
	if err != nil {
		return "", fmt.Errorf("git rev-parse HEAD: %w", err)
	}
	return strings.TrimSpace(out), nil
}

// LogEntry represents a single commit in the log.
type LogEntry struct {
	Hash    string
	Subject string
}

// RecentCommits returns the last n commits as short hash + subject.
func (g *GitClient) RecentCommits(n int) ([]LogEntry, error) {
	out, err := g.run("log", "--oneline", fmt.Sprintf("-n%d", n), "--format=%h %s")
	if err != nil {
		return nil, fmt.Errorf("git log: %w", err)
	}

	var entries []LogEntry
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		entry := LogEntry{Hash: parts[0]}
		if len(parts) > 1 {
			entry.Subject = parts[1]
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func (g *GitClient) run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = g.repoRoot
	out, err := cmd.Output()
	if err != nil {
		return string(out), err
	}
	return string(out), nil
}

func parseFileStatus(s string) types.FileChangeStatus {
	switch {
	case s == "A":
		return types.FileAdded
	case s == "D":
		return types.FileDeleted
	case s == "M":
		return types.FileModified
	case strings.HasPrefix(s, "R"):
		return types.FileRenamed
	default:
		return types.FileModified
	}
}

// parseDiff parses unified diff output into structured hunks.
func parseDiff(raw string) []types.DiffHunk {
	var hunks []types.DiffHunk
	lines := strings.Split(raw, "\n")

	var current *types.DiffHunk
	for _, line := range lines {
		if strings.HasPrefix(line, "@@") {
			if current != nil {
				hunks = append(hunks, *current)
			}
			current = parseHunkHeader(line)
			continue
		}
		if current == nil {
			continue
		}

		dl := types.DiffLine{Content: line}
		switch {
		case strings.HasPrefix(line, "+"):
			dl.Kind = types.DiffLineAdded
			dl.NewLineNum = current.NewStart + countLines(current.Lines, types.DiffLineAdded, types.DiffLineContext)
			dl.Content = line[1:]
		case strings.HasPrefix(line, "-"):
			dl.Kind = types.DiffLineRemoved
			dl.OldLineNum = current.OldStart + countLines(current.Lines, types.DiffLineRemoved, types.DiffLineContext)
			dl.Content = line[1:]
		default:
			if len(line) > 0 && line[0] == ' ' {
				dl.Content = line[1:]
			}
			dl.Kind = types.DiffLineContext
			dl.OldLineNum = current.OldStart + countLines(current.Lines, types.DiffLineRemoved, types.DiffLineContext)
			dl.NewLineNum = current.NewStart + countLines(current.Lines, types.DiffLineAdded, types.DiffLineContext)
		}
		current.Lines = append(current.Lines, dl)
	}
	if current != nil {
		hunks = append(hunks, *current)
	}

	return hunks
}

func parseHunkHeader(line string) *types.DiffHunk {
	// Format: @@ -old_start,old_count +new_start,new_count @@ optional header
	h := &types.DiffHunk{}

	// Find the ranges between @@ markers
	parts := strings.SplitN(line, "@@", 3)
	if len(parts) < 2 {
		return h
	}
	if len(parts) == 3 {
		h.Header = strings.TrimSpace(parts[2])
	}

	ranges := strings.TrimSpace(parts[1])
	rangeParts := strings.Fields(ranges)

	for _, rp := range rangeParts {
		if strings.HasPrefix(rp, "-") {
			nums := strings.SplitN(rp[1:], ",", 2)
			h.OldStart, _ = strconv.Atoi(nums[0])
			if len(nums) > 1 {
				h.OldCount, _ = strconv.Atoi(nums[1])
			} else {
				h.OldCount = 1
			}
		} else if strings.HasPrefix(rp, "+") {
			nums := strings.SplitN(rp[1:], ",", 2)
			h.NewStart, _ = strconv.Atoi(nums[0])
			if len(nums) > 1 {
				h.NewCount, _ = strconv.Atoi(nums[1])
			} else {
				h.NewCount = 1
			}
		}
	}

	return h
}

func countLines(lines []types.DiffLine, kinds ...types.DiffLineKind) int {
	n := 0
	for _, l := range lines {
		for _, k := range kinds {
			if l.Kind == k {
				n++
				break
			}
		}
	}
	return n
}

package adapters

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
)

// FindRepoRoot walks up from startDir looking for a .git entry (dir or file,
// to handle worktrees/submodules). Returns the directory containing .git, or
// the absolute startDir if none found.
func FindRepoRoot(startDir string) string {
	abs, err := filepath.Abs(startDir)
	if err != nil {
		return startDir
	}

	dir := abs
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root without finding .git
			return abs
		}
		dir = parent
	}
}

// DefaultSocketPath computes a deterministic socket path from a directory.
// Returns /tmp/monocle-<sha256_first12>.sock (30 chars, well within the
// ~104-byte macOS socket path limit).
func DefaultSocketPath(dir string) string {
	abs, err := filepath.Abs(dir)
	if err != nil {
		abs = dir
	}
	sum := sha256.Sum256([]byte(abs))
	hash := hex.EncodeToString(sum[:])[:12]
	return "/tmp/monocle-" + hash + ".sock"
}

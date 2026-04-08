package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/josephschmitt/monocle/internal/types"
)

// DirClient implements GitAPI for non-git directories.
// It treats the directory as a flat collection of files with no version history.
type DirClient struct {
	repoRoot       string
	ignorePatterns []string
}

// NewDirClient creates a DirClient for the given directory.
func NewDirClient(root string, ignorePatterns []string) *DirClient {
	return &DirClient{repoRoot: root, ignorePatterns: ignorePatterns}
}

func (d *DirClient) RepoRoot() string {
	return d.repoRoot
}

// CurrentRef returns a sentinel value since there is no git history.
func (d *DirClient) CurrentRef() (string, error) {
	return "WORKING", nil
}

// Diff walks the directory and returns all regular files as "added".
func (d *DirClient) Diff(_ string) ([]types.ChangedFile, error) {
	var files []types.ChangedFile

	err := filepath.WalkDir(d.repoRoot, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}

		name := entry.Name()

		if entry.IsDir() {
			if strings.HasPrefix(name, ".") || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}

		if !entry.Type().IsRegular() {
			return nil
		}
		if strings.HasPrefix(name, ".") {
			return nil
		}

		rel, err := filepath.Rel(d.repoRoot, path)
		if err != nil {
			return nil
		}

		if d.isIgnored(rel) {
			return nil
		}

		if d.hasNullBytes(path) {
			return nil
		}

		files = append(files, types.ChangedFile{
			Path:   rel,
			Status: types.FileNone,
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk directory: %w", err)
	}

	return files, nil
}

// FileDiff returns a synthetic diff showing the entire file as added.
func (d *DirClient) FileDiff(_, path string, _ int) (*types.DiffResult, error) {
	content, err := d.FileContent("", path)
	if err != nil {
		return nil, err
	}

	hunks := buildSyntheticDiff(content)
	return &types.DiffResult{
		Path:  path,
		Hunks: hunks,
	}, nil
}

// FileContent reads a file from the directory.
// Returns an error for binary/non-text files.
func (d *DirClient) FileContent(_, path string) (string, error) {
	absPath := filepath.Join(d.repoRoot, path)
	if isNonText(absPath) || d.hasNullBytes(absPath) {
		return "", fmt.Errorf("binary file — cannot preview %s", path)
	}
	data, err := os.ReadFile(absPath)
	if err != nil {
		return "", fmt.Errorf("read file %s: %w", path, err)
	}
	return string(data), nil
}

// RecentCommits returns nothing since there is no git history.
func (d *DirClient) RecentCommits(_ int) ([]LogEntry, error) {
	return nil, nil
}

// ResolveRef returns an error since there are no git refs.
func (d *DirClient) ResolveRef(ref string) (string, error) {
	return "", fmt.Errorf("no git repository: cannot resolve ref %q", ref)
}

// HashObject is not supported in non-git mode. Returns empty string.
func (d *DirClient) HashObject(_ string) (string, error) {
	return "", fmt.Errorf("no git repository: cannot hash object")
}

// CatFile is not supported in non-git mode. Returns empty string.
func (d *DirClient) CatFile(_ string) (string, error) {
	return "", fmt.Errorf("no git repository: cannot cat file")
}

// isIgnored checks if a relative path matches any ignore pattern.
func (d *DirClient) isIgnored(rel string) bool {
	for _, pattern := range d.ignorePatterns {
		// Match against the full relative path and each path component
		if matched, _ := filepath.Match(pattern, rel); matched {
			return true
		}
		if matched, _ := filepath.Match(pattern, filepath.Base(rel)); matched {
			return true
		}
		// Check if any directory component matches (e.g., "vendor/" matches "vendor/foo.go")
		trimmed := strings.TrimSuffix(pattern, "/")
		if trimmed != pattern {
			for _, part := range strings.Split(filepath.Dir(rel), string(filepath.Separator)) {
				if part == trimmed {
					return true
				}
			}
		}
	}
	return false
}

// nonTextExtensions lists file extensions that are not meaningfully previewable as text.
var nonTextExtensions = map[string]bool{
	".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
	".ppt": true, ".pptx": true, ".odt": true, ".ods": true, ".odp": true,
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".bmp": true,
	".ico": true, ".webp": true, ".svg": true, ".tiff": true, ".tif": true,
	".mp3": true, ".mp4": true, ".wav": true, ".avi": true, ".mov": true,
	".mkv": true, ".flac": true, ".ogg": true, ".webm": true,
	".zip": true, ".tar": true, ".gz": true, ".bz2": true, ".xz": true,
	".7z": true, ".rar": true, ".zst": true,
	".exe": true, ".dll": true, ".so": true, ".dylib": true, ".a": true,
	".o": true, ".obj": true, ".class": true, ".pyc": true,
	".woff": true, ".woff2": true, ".ttf": true, ".otf": true, ".eot": true,
	".sqlite": true, ".db": true,
}

// isNonText checks if a file is non-text based on its extension.
func isNonText(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return nonTextExtensions[ext]
}

// hasNullBytes checks if a file contains null bytes in the first 512 bytes,
// indicating it is a binary file.
func (d *DirClient) hasNullBytes(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, _ := f.Read(buf)
	if n == 0 {
		return false
	}

	for _, b := range buf[:n] {
		if b == 0 {
			return true
		}
	}
	return false
}

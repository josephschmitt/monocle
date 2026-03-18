package adapters

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindRepoRoot_FromSubdirectory(t *testing.T) {
	dir := t.TempDir()
	// Create .git dir at root
	os.MkdirAll(filepath.Join(dir, ".git"), 0755)
	// Create a subdirectory
	sub := filepath.Join(dir, "src", "pkg")
	os.MkdirAll(sub, 0755)

	got := FindRepoRoot(sub)
	if got != dir {
		t.Fatalf("FindRepoRoot(%s) = %s, want %s", sub, got, dir)
	}
}

func TestFindRepoRoot_NoGitDir(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "some", "path")
	os.MkdirAll(sub, 0755)

	got := FindRepoRoot(sub)
	// Should return the absolute path of the start dir
	abs, _ := filepath.Abs(sub)
	if got != abs {
		t.Fatalf("FindRepoRoot(%s) = %s, want %s", sub, got, abs)
	}
}

func TestFindRepoRoot_AtRepoRoot(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".git"), 0755)

	got := FindRepoRoot(dir)
	if got != dir {
		t.Fatalf("FindRepoRoot(%s) = %s, want %s", dir, got, dir)
	}
}

func TestFindRepoRoot_GitFile(t *testing.T) {
	// Worktrees use a .git file instead of a .git directory
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".git"), []byte("gitdir: /somewhere/else"), 0644)
	sub := filepath.Join(dir, "sub")
	os.MkdirAll(sub, 0755)

	got := FindRepoRoot(sub)
	if got != dir {
		t.Fatalf("FindRepoRoot(%s) = %s, want %s", sub, got, dir)
	}
}

func TestDefaultSocketPath_Deterministic(t *testing.T) {
	p1 := DefaultSocketPath("/some/repo")
	p2 := DefaultSocketPath("/some/repo")
	if p1 != p2 {
		t.Fatalf("not deterministic: %s != %s", p1, p2)
	}
}

func TestDefaultSocketPath_DifferentRoots(t *testing.T) {
	p1 := DefaultSocketPath("/repo/one")
	p2 := DefaultSocketPath("/repo/two")
	if p1 == p2 {
		t.Fatalf("different roots should produce different paths: both %s", p1)
	}
}

func TestDefaultSocketPath_Format(t *testing.T) {
	p := DefaultSocketPath("/some/repo")
	if !strings.HasPrefix(p, "/tmp/monocle-") {
		t.Fatalf("expected /tmp/monocle- prefix, got %s", p)
	}
	if !strings.HasSuffix(p, ".sock") {
		t.Fatalf("expected .sock suffix, got %s", p)
	}
	// /tmp/monocle- (13) + 12 hex chars + .sock (5) = 30
	if len(p) != 30 {
		t.Fatalf("expected length 30, got %d (%s)", len(p), p)
	}
}

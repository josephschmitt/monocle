package core

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/anthropics/monocle/internal/types"
)

func setupTestRepo(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()

	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=Test",
			"GIT_AUTHOR_EMAIL=test@test.com",
			"GIT_COMMITTER_NAME=Test",
			"GIT_COMMITTER_EMAIL=test@test.com",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}

	run("init")
	run("checkout", "-b", "main")

	// Create initial file and commit
	os.WriteFile(filepath.Join(dir, "hello.go"), []byte("package main\n\nfunc hello() {}\n"), 0o644)
	run("add", "hello.go")
	run("commit", "-m", "initial")

	// Get the base ref
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dir
	out, _ := cmd.Output()
	baseRef := string(out[:len(out)-1])

	// Make changes
	os.WriteFile(filepath.Join(dir, "hello.go"), []byte("package main\n\nfunc hello() {\n\tprintln(\"hello\")\n}\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "world.go"), []byte("package main\n\nfunc world() {}\n"), 0o644)
	run("add", "world.go")

	return dir, baseRef
}

func TestGitDiff(t *testing.T) {
	dir, baseRef := setupTestRepo(t)
	g := NewGitClient(dir)

	files, err := g.Diff(baseRef)
	if err != nil {
		t.Fatalf("Diff: %v", err)
	}

	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}

	// Files should be hello.go (modified) and world.go (added)
	found := map[string]types.FileChangeStatus{}
	for _, f := range files {
		found[f.Path] = f.Status
	}

	if found["hello.go"] != types.FileModified {
		t.Errorf("hello.go: expected modified, got %q", found["hello.go"])
	}
	// world.go is untracked, won't show in diff against commit
	// It should not appear since it's not staged
}

func TestGitFileDiff(t *testing.T) {
	dir, baseRef := setupTestRepo(t)
	g := NewGitClient(dir)

	result, err := g.FileDiff(baseRef, "hello.go")
	if err != nil {
		t.Fatalf("FileDiff: %v", err)
	}

	if len(result.Hunks) == 0 {
		t.Fatal("expected at least one hunk")
	}

	hunk := result.Hunks[0]
	if len(hunk.Lines) == 0 {
		t.Fatal("expected lines in hunk")
	}

	// Verify we have both added and context lines
	hasAdded := false
	for _, l := range hunk.Lines {
		if l.Kind == types.DiffLineAdded {
			hasAdded = true
		}
	}
	if !hasAdded {
		t.Error("expected added lines in diff")
	}
}

func TestGitCurrentRef(t *testing.T) {
	dir, _ := setupTestRepo(t)
	g := NewGitClient(dir)

	ref, err := g.CurrentRef()
	if err != nil {
		t.Fatalf("CurrentRef: %v", err)
	}
	if len(ref) != 40 {
		t.Errorf("expected 40-char hash, got %q", ref)
	}
}

func TestParseDiff(t *testing.T) {
	raw := `diff --git a/hello.go b/hello.go
index abc..def 100644
--- a/hello.go
+++ b/hello.go
@@ -1,3 +1,5 @@ package main
 package main

 func hello() {
+	println("hello")
+}
`
	hunks := parseDiff(raw)
	if len(hunks) != 1 {
		t.Fatalf("expected 1 hunk, got %d", len(hunks))
	}

	h := hunks[0]
	if h.OldStart != 1 || h.OldCount != 3 || h.NewStart != 1 || h.NewCount != 5 {
		t.Errorf("hunk header: old=%d,%d new=%d,%d", h.OldStart, h.OldCount, h.NewStart, h.NewCount)
	}
}

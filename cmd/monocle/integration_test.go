package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/josephschmitt/monocle/internal/adapters"
	"github.com/josephschmitt/monocle/internal/client"
	"github.com/josephschmitt/monocle/internal/core"
	"github.com/josephschmitt/monocle/internal/types"
)

// TestServeToClientE2E builds the monocle binary, spawns `monocle serve` via
// the auto-spawn helper, connects as an EngineClient, exercises a handful of
// read/write methods, then stops the server. This is the closest we can get
// to "user runs monocle" without a TTY.
func TestServeToClientE2E(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix sockets unavailable on windows")
	}

	// Build the binary to a temp path.
	buildDir := t.TempDir()
	binary := filepath.Join(buildDir, "monocle")
	cmd := exec.Command("go", "build", "-o", binary, "./")
	cmd.Dir = "."
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}

	// Fresh "repo" with one file.
	repoRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(repoRoot, "a.go"), []byte("package a\n"), 0o644); err != nil {
		t.Fatalf("seed file: %v", err)
	}

	// Override the socket path so we don't collide with a user's live serve.
	hash := sha256.Sum256([]byte(t.Name() + repoRoot))
	socketPath := fmt.Sprintf("/tmp/monocle-e2e-%s.sock", hex.EncodeToString(hash[:])[:10])
	_ = os.Remove(socketPath)

	got, _, err := adapters.EnsureServe(adapters.AutoSpawnOptions{
		RepoRoot:     repoRoot,
		Socket:       socketPath,
		Binary:       binary,
		ReadyTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("ensure serve: %v", err)
	}
	if got != socketPath {
		t.Errorf("socket = %q, want %q", got, socketPath)
	}
	defer func() {
		// Best-effort stop.
		stop := exec.Command(binary, "stop", "-C", repoRoot, "--socket", socketPath)
		_ = stop.Run()
	}()

	ec, err := client.NewEngineClient(socketPath)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer ec.Close()

	// Server already has a default session from `serve` startup.
	sess := ec.GetSession()
	if sess == nil {
		t.Fatal("expected a default session from monocle serve")
	}

	// Exercise a write path: add a comment.
	c, err := ec.AddComment(core.CommentTarget{
		TargetType: types.TargetFile,
		TargetRef:  "a.go",
		LineStart:  1,
		LineEnd:    1,
	}, types.CommentNote, "hi")
	if err != nil {
		t.Fatalf("add comment: %v", err)
	}
	if c == nil || c.Body != "hi" {
		t.Errorf("unexpected comment: %+v", c)
	}
}

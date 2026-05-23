package client

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/josephschmitt/monocle/internal/core"
	"github.com/josephschmitt/monocle/internal/db"
	"github.com/josephschmitt/monocle/internal/types"
)

// setupEngine spins up a non-git-mode engine against a temp directory with
// one file, starts its socket server, and returns the live engine + socket
// path. Cleanup is registered via t.Cleanup.
func setupEngine(t *testing.T) (*core.Engine, string) {
	t.Helper()

	repoRoot := t.TempDir()
	// Seed one file so GetChangedFiles returns something deterministic.
	if err := os.WriteFile(filepath.Join(repoRoot, "a.go"), []byte("package a\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { database.Close() })

	engine, err := core.NewEngine(core.DefaultConfig(), database, repoRoot, true /* nonGitMode */)
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}

	if _, err := engine.StartSession(core.SessionOptions{
		Agent:    "test",
		RepoRoot: repoRoot,
	}); err != nil {
		t.Fatalf("start session: %v", err)
	}

	hash := sha256.Sum256([]byte(t.Name()))
	socketPath := fmt.Sprintf("/tmp/monocle-client-test-%s.sock", hex.EncodeToString(hash[:])[:10])
	if err := engine.StartServer(socketPath); err != nil {
		t.Fatalf("start server: %v", err)
	}
	t.Cleanup(func() { engine.Shutdown() })

	return engine, socketPath
}

func TestEngineClient_RoundTrip(t *testing.T) {
	engine, socketPath := setupEngine(t)

	ec, err := NewEngineClient(socketPath)
	if err != nil {
		t.Fatalf("new engine client: %v", err)
	}
	t.Cleanup(func() { ec.Close() })

	// GetSession — session started in setupEngine.
	if sess := ec.GetSession(); sess == nil {
		t.Error("GetSession returned nil")
	}

	// RefreshChangedFiles — should find the seeded a.go.
	files, err := ec.RefreshChangedFiles()
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if len(files) == 0 {
		t.Error("expected at least one changed file")
	}

	// Comment round-trip.
	c, err := ec.AddComment(core.CommentTarget{
		TargetType: types.TargetFile,
		TargetRef:  "a.go",
		LineStart:  1,
		LineEnd:    1,
	}, types.CommentIssue, "needs a docstring")
	if err != nil {
		t.Fatalf("add comment: %v", err)
	}
	if c == nil || c.Body != "needs a docstring" {
		t.Errorf("unexpected comment: %+v", c)
	}

	edited, err := ec.EditComment(c.ID, types.CommentSuggestion, "add a docstring")
	if err != nil {
		t.Fatalf("edit comment: %v", err)
	}
	if edited.Body != "add a docstring" {
		t.Errorf("edit did not apply: %q", edited.Body)
	}

	if err := ec.ResolveComment(c.ID); err != nil {
		t.Fatalf("resolve: %v", err)
	}

	// Server-side state via engine matches the wire result.
	sess := engine.GetSession()
	if sess == nil || len(sess.Comments) != 1 {
		t.Errorf("expected 1 comment on engine session, got %d", len(sess.Comments))
	}

	// Marking flows.
	if err := ec.MarkReviewed("a.go"); err != nil {
		t.Fatalf("mark reviewed: %v", err)
	}
	if err := ec.UnmarkReviewed("a.go"); err != nil {
		t.Fatalf("unmark: %v", err)
	}

	// Config round-trip. GetConfig returns a pointer cached on the client.
	cfg := ec.GetConfig()
	if cfg == nil {
		t.Fatal("GetConfig returned nil")
	}
	origWrap := cfg.Wrap
	cfg.Wrap = !origWrap
	if err := ec.SaveConfig(); err != nil {
		t.Fatalf("save config: %v", err)
	}
	// Server engine saw the mutated value.
	if engine.GetConfig().Wrap == origWrap {
		t.Errorf("SaveConfig did not propagate mutation: got Wrap=%v, want %v", engine.GetConfig().Wrap, !origWrap)
	}

	// Status queries.
	if path := ec.GetSocketPath(); path != socketPath {
		t.Errorf("GetSocketPath = %q, want %q", path, socketPath)
	}
	// EngineClient subscribes passively (TUI is a viewer, not an agent),
	// so it must NOT bump the agent-facing subscriber count — otherwise
	// Submit() would flip into push mode and the real agent would lose
	// the feedback.
	if count := ec.GetSubscriberCount(); count != 0 {
		t.Errorf("GetSubscriberCount = %d, want 0 (passive subscribe must not count)", count)
	}
}

func TestEngineClient_EventBus(t *testing.T) {
	engine, socketPath := setupEngine(t)
	_ = engine

	// Seed a second file to add as an additional-path (triggers
	// EventAdditionalFileAdded from the server).
	repoRoot := engine.GetSession().RepoRoot
	extra := filepath.Join(repoRoot, "extra.go")
	if err := os.WriteFile(extra, []byte("package a\n"), 0o644); err != nil {
		t.Fatalf("seed extra: %v", err)
	}

	ec, err := NewEngineClient(socketPath)
	if err != nil {
		t.Fatalf("new engine client: %v", err)
	}
	t.Cleanup(func() { ec.Close() })

	events := make(chan core.EventPayload, 4)
	unsub := ec.On(core.EventAdditionalFileAdded, func(p core.EventPayload) {
		events <- p
	})

	if _, err := ec.AddAdditionalPaths([]string{extra}); err != nil {
		t.Fatalf("add additional paths: %v", err)
	}

	select {
	case p := <-events:
		if p.Kind != core.EventAdditionalFileAdded {
			t.Errorf("wrong event kind: %v", p.Kind)
		}
		if p.Path != extra {
			t.Errorf("event path = %q, want %q", p.Path, extra)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("event not dispatched to client")
	}

	// Unsubscribing stops delivery for further events.
	unsub()
	extra2 := filepath.Join(repoRoot, "extra2.go")
	if err := os.WriteFile(extra2, []byte("package a\n"), 0o644); err != nil {
		t.Fatalf("seed extra2: %v", err)
	}
	if _, err := ec.AddAdditionalPaths([]string{extra2}); err != nil {
		t.Fatalf("add additional paths 2: %v", err)
	}
	select {
	case p := <-events:
		t.Errorf("received event after unsub: %+v", p)
	case <-time.After(200 * time.Millisecond):
		// expected
	}
}

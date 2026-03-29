package client_test

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/josephschmitt/monocle/internal/client"
	"github.com/josephschmitt/monocle/internal/core"
	"github.com/josephschmitt/monocle/internal/db"
	"github.com/josephschmitt/monocle/internal/protocol"
)

func setupTestEngine(t *testing.T) (*core.Engine, string) {
	t.Helper()

	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { database.Close() })

	tmpDir := t.TempDir()
	cfg := core.DefaultConfig()
	engine, err := core.NewEngine(cfg, database, tmpDir, true /* nonGitMode */)
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}

	_, err = engine.StartSession(core.SessionOptions{
		Agent:    "test",
		RepoRoot: tmpDir,
	})
	if err != nil {
		t.Fatalf("start session: %v", err)
	}

	hash := sha256.Sum256([]byte(t.Name()))
	socketPath := fmt.Sprintf("/tmp/monocle-test-%s.sock", hex.EncodeToString(hash[:])[:8])
	if err := engine.StartServer(socketPath); err != nil {
		t.Fatalf("start server: %v", err)
	}
	t.Cleanup(func() { engine.Shutdown() })

	return engine, socketPath
}

func TestClient_ReviewStatus(t *testing.T) {
	_, socketPath := setupTestEngine(t)

	c, err := client.Connect(socketPath)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer c.Close()

	msg := &protocol.GetReviewStatusMsg{Type: protocol.TypeGetReviewStatus}
	resp, err := c.Request(msg, client.DefaultTimeout)
	if err != nil {
		t.Fatalf("request: %v", err)
	}

	status, ok := resp.(*protocol.GetReviewStatusResponse)
	if !ok {
		t.Fatalf("expected *GetReviewStatusResponse, got %T", resp)
	}
	if status.Status != "no_feedback" {
		t.Errorf("status = %q, want %q", status.Status, "no_feedback")
	}
}

func TestClient_PollFeedback_NoWait(t *testing.T) {
	_, socketPath := setupTestEngine(t)

	c, err := client.Connect(socketPath)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer c.Close()

	msg := &protocol.PollFeedbackMsg{Type: protocol.TypePollFeedback, Wait: false}
	resp, err := c.Request(msg, client.DefaultTimeout)
	if err != nil {
		t.Fatalf("request: %v", err)
	}

	feedback, ok := resp.(*protocol.PollFeedbackResponse)
	if !ok {
		t.Fatalf("expected *PollFeedbackResponse, got %T", resp)
	}
	if feedback.HasFeedback {
		t.Error("expected no feedback")
	}
}

func TestClient_SubmitContent(t *testing.T) {
	_, socketPath := setupTestEngine(t)

	c, err := client.Connect(socketPath)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer c.Close()

	msg := &protocol.SubmitContentMsg{
		Type:        protocol.TypeSubmitContent,
		ID:          "test-plan",
		Title:       "Test Plan",
		Content:     "# My Plan\n\nDo the thing.",
		ContentType: "md",
		IsPlan:      true,
	}
	resp, err := c.Request(msg, client.DefaultTimeout)
	if err != nil {
		t.Fatalf("request: %v", err)
	}

	submit, ok := resp.(*protocol.SubmitContentResponse)
	if !ok {
		t.Fatalf("expected *SubmitContentResponse, got %T", resp)
	}
	if !submit.Success {
		t.Errorf("expected success, got message: %s", submit.Message)
	}
}

func TestClient_AddFiles(t *testing.T) {
	_, socketPath := setupTestEngine(t)

	c, err := client.Connect(socketPath)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer c.Close()

	msg := &protocol.AddAdditionalFilesMsg{
		Type:  protocol.TypeAddAdditionalFiles,
		Paths: []string{t.TempDir()},
	}
	resp, err := c.Request(msg, client.DefaultTimeout)
	if err != nil {
		t.Fatalf("request: %v", err)
	}

	add, ok := resp.(*protocol.AddAdditionalFilesResponse)
	if !ok {
		t.Fatalf("expected *AddAdditionalFilesResponse, got %T", resp)
	}
	if !add.Success {
		t.Errorf("expected success, got message: %s", add.Message)
	}
}

func TestClient_ErrNotRunning(t *testing.T) {
	_, err := client.Connect("/tmp/monocle-does-not-exist.sock")
	if err != client.ErrNotRunning {
		t.Errorf("expected ErrNotRunning, got %v", err)
	}
}

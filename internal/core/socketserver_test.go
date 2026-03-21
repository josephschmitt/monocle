package core

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/anthropics/monocle/internal/db"
	"github.com/anthropics/monocle/internal/protocol"
	"github.com/anthropics/monocle/internal/types"
)

func setupTestEngine(t *testing.T) (*Engine, string) {
	t.Helper()

	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { database.Close() })

	repoRoot, _ := setupTestRepo(t)
	cfg := DefaultConfig()
	engine, err := NewEngine(cfg, database, repoRoot)
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}

	// Start a session so the engine is usable
	_, err = engine.StartSession(SessionOptions{
		Agent:    "test",
		RepoRoot: repoRoot,
	})
	if err != nil {
		t.Fatalf("start session: %v", err)
	}

	// Start socket server (use /tmp with short hash to stay within macOS 104-byte socket path limit)
	hash := sha256.Sum256([]byte(t.Name()))
	socketPath := fmt.Sprintf("/tmp/monocle-test-%s.sock", hex.EncodeToString(hash[:])[:8])
	if err := engine.StartServer(socketPath); err != nil {
		t.Fatalf("start server: %v", err)
	}
	t.Cleanup(func() { engine.Shutdown() })

	return engine, socketPath
}

func TestSocketServer_OneShot(t *testing.T) {
	_, socketPath := setupTestEngine(t)

	// Connect and send a one-shot review-status request
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	msg := protocol.GetReviewStatusMsg{Type: protocol.TypeGetReviewStatus}
	data, _ := protocol.Encode(&msg)
	conn.Write(data)

	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	if !scanner.Scan() {
		t.Fatal("no response")
	}

	decoded, err := protocol.Decode(scanner.Bytes())
	if err != nil {
		t.Fatalf("decode: %v", err)
	}

	resp, ok := decoded.(*protocol.GetReviewStatusResponse)
	if !ok {
		t.Fatalf("expected *GetReviewStatusResponse, got %T", decoded)
	}
	if resp.Status != "no_feedback" {
		t.Errorf("status = %q, want %q", resp.Status, "no_feedback")
	}
}

func TestSocketServer_Subscription(t *testing.T) {
	engine, socketPath := setupTestEngine(t)

	// Connect and subscribe
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	// Send subscribe message
	sub := protocol.SubscribeMsg{
		Type:   protocol.TypeSubscribe,
		Events: []string{string(EventFeedbackSubmitted)},
	}
	data, _ := protocol.Encode(&sub)
	conn.Write(data)

	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	// Read ack
	if !scanner.Scan() {
		t.Fatal("no ack")
	}
	decoded, err := protocol.Decode(scanner.Bytes())
	if err != nil {
		t.Fatalf("decode ack: %v", err)
	}
	ack, ok := decoded.(*protocol.SubscribeResponse)
	if !ok {
		t.Fatalf("expected *SubscribeResponse, got %T", decoded)
	}
	if !ack.Success {
		t.Fatal("expected success ack")
	}

	// Emit an event from the engine
	engine.emit(EventFeedbackSubmitted, EventPayload{
		Kind:    EventFeedbackSubmitted,
		Message: "## Code Review — Changes Requested",
		Status:  "request_changes",
	})

	// Read the event notification
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	if !scanner.Scan() {
		t.Fatal("no event notification received")
	}

	decoded, err = protocol.Decode(scanner.Bytes())
	if err != nil {
		t.Fatalf("decode notification: %v", err)
	}
	notif, ok := decoded.(*protocol.EventNotification)
	if !ok {
		t.Fatalf("expected *EventNotification, got %T", decoded)
	}
	if notif.Event != string(EventFeedbackSubmitted) {
		t.Errorf("event = %q, want %q", notif.Event, EventFeedbackSubmitted)
	}
	if notif.Payload["message"] != "## Code Review — Changes Requested" {
		t.Errorf("payload.message = %q", notif.Payload["message"])
	}
}

func TestSocketServer_SubscriptionWithRequests(t *testing.T) {
	_, socketPath := setupTestEngine(t)

	// Connect and subscribe
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	sub := protocol.SubscribeMsg{
		Type:   protocol.TypeSubscribe,
		Events: []string{string(EventFeedbackSubmitted)},
	}
	data, _ := protocol.Encode(&sub)
	conn.Write(data)

	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	// Read ack
	if !scanner.Scan() {
		t.Fatal("no ack")
	}

	// Send a request/response message on the same connection
	reqMsg := protocol.GetReviewStatusMsg{Type: protocol.TypeGetReviewStatus}
	data, _ = protocol.Encode(&reqMsg)
	conn.Write(data)

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	if !scanner.Scan() {
		t.Fatal("no response to request")
	}

	decoded, err := protocol.Decode(scanner.Bytes())
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	resp, ok := decoded.(*protocol.GetReviewStatusResponse)
	if !ok {
		t.Fatalf("expected *GetReviewStatusResponse, got %T", decoded)
	}
	if resp.Status != "no_feedback" {
		t.Errorf("status = %q, want %q", resp.Status, "no_feedback")
	}
}

func TestSocketServer_SubmitEmitsFeedbackEvent(t *testing.T) {
	engine, socketPath := setupTestEngine(t)

	// Add a comment so Submit generates actual feedback
	_, err := engine.AddComment(CommentTarget{
		TargetType: types.TargetFile,
		TargetRef:  "test.go",
		LineStart:  1,
		LineEnd:    1,
	}, types.CommentIssue, "Fix this")
	if err != nil {
		t.Fatalf("add comment: %v", err)
	}

	// Connect and subscribe to feedback events
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	sub := protocol.SubscribeMsg{
		Type:   protocol.TypeSubscribe,
		Events: []string{string(EventFeedbackSubmitted)},
	}
	data, _ := protocol.Encode(&sub)
	conn.Write(data)

	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	// Read ack
	if !scanner.Scan() {
		t.Fatal("no ack")
	}

	// Submit the review
	_, err = engine.Submit(types.ActionRequestChanges, "")
	if err != nil {
		t.Fatalf("submit: %v", err)
	}

	// Should receive an EventNotification
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))

	// May get EventFeedbackStatusChanged first, then EventFeedbackSubmitted
	// Read until we find the feedback_submitted event
	found := false
	for i := 0; i < 5; i++ {
		if !scanner.Scan() {
			break
		}
		var raw map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &raw); err != nil {
			continue
		}
		if raw["type"] == protocol.TypeEventNotification && raw["event"] == string(EventFeedbackSubmitted) {
			found = true
			payload, _ := raw["payload"].(map[string]any)
			msg, _ := payload["message"].(string)
			if msg == "" {
				t.Error("expected non-empty feedback message in event payload")
			}
			break
		}
	}
	if !found {
		t.Error("did not receive feedback_submitted event notification")
	}
}

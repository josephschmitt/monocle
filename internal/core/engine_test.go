package core

import (
	"testing"
	"time"

	"github.com/anthropics/monocle/internal/db"
	"github.com/anthropics/monocle/internal/protocol"
	"github.com/anthropics/monocle/internal/types"
)

func TestShouldAutoApprove(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(e *Engine)
		expected bool
	}{
		{
			name:     "nil session",
			setup:    func(e *Engine) { e.current = nil },
			expected: true,
		},
		{
			name: "empty session",
			setup: func(e *Engine) {
				e.current = &types.ReviewSession{}
			},
			expected: true,
		},
		{
			name: "session with changed files",
			setup: func(e *Engine) {
				e.current = &types.ReviewSession{
					ChangedFiles: []types.ChangedFile{{Path: "foo.go"}},
				}
			},
			expected: false,
		},
		{
			name: "session with content items",
			setup: func(e *Engine) {
				e.current = &types.ReviewSession{
					ContentItems: []types.ContentItem{{ID: "item-1"}},
				}
			},
			expected: false,
		},
		{
			name: "session with active comment",
			setup: func(e *Engine) {
				e.current = &types.ReviewSession{
					Comments: []types.ReviewComment{
						{ID: "c1", Outdated: false},
					},
				}
			},
			expected: false,
		},
		{
			name: "session with only outdated comments",
			setup: func(e *Engine) {
				e.current = &types.ReviewSession{
					Comments: []types.ReviewComment{
						{ID: "c1", Outdated: true},
						{ID: "c2", Outdated: true},
					},
				}
			},
			expected: true,
		},
		{
			name: "queued feedback",
			setup: func(e *Engine) {
				e.current = &types.ReviewSession{}
				e.feedback.Submit(&FormattedReview{
					Formatted:    "review",
					CommentCount: 1,
					Action:       "request_changes",
				})
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Engine{
				feedback: NewFeedbackQueue(),
			}
			tt.setup(e)

			e.mu.Lock()
			got := e.shouldAutoApprove()
			e.mu.Unlock()

			if got != tt.expected {
				t.Errorf("shouldAutoApprove() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestHandleStopWithPlanContent(t *testing.T) {
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	e := &Engine{
		feedback:    NewFeedbackQueue(),
		database:    database,
		subscribers: make(map[EventKind]map[int]EventCallback),
		sessions:    &SessionManager{db: database, git: NewGitClient(t.TempDir())},
	}
	e.current = &types.ReviewSession{
		ID:           "sess-1",
		FileStatuses: make(map[string]bool),
	}

	done := make(chan *protocol.StopResponse)
	go func() {
		done <- e.handleStop(&protocol.StopMsg{
			Type:               protocol.TypeStop,
			RequestID:          "r1",
			ReviewContent:      "# Plan\n1. Step one\n2. Step two",
			ReviewContentTitle: "Plan",
			ReviewContentType:  "markdown",
		})
	}()

	// Give handleStop time to process content and block
	time.Sleep(50 * time.Millisecond)

	// Verify the plan was injected as a content item
	e.mu.RLock()
	items := e.current.ContentItems
	e.mu.RUnlock()
	if len(items) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(items))
	}
	if items[0].ID != "plan" {
		t.Errorf("expected content item ID 'plan', got %q", items[0].ID)
	}
	if items[0].Content != "# Plan\n1. Step one\n2. Step two" {
		t.Errorf("unexpected content: %q", items[0].Content)
	}

	// handleStop should be blocking (not auto-approved)
	select {
	case <-done:
		t.Fatal("handleStop returned early — should be blocking because of plan content")
	case <-time.After(100 * time.Millisecond):
		// expected: still blocking
	}

	// Approve to unblock
	e.feedback.Approve()

	select {
	case resp := <-done:
		if resp.Continue {
			t.Error("expected continue=false for approve")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("handleStop did not return after approve")
	}

	// Verify content items were cleared after stop completed
	e.mu.RLock()
	itemsAfter := e.current.ContentItems
	e.mu.RUnlock()
	if len(itemsAfter) != 0 {
		t.Errorf("expected content items cleared after stop, got %d", len(itemsAfter))
	}
}

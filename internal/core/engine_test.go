package core

import (
	"testing"

	"github.com/anthropics/monocle/internal/db"
	"github.com/anthropics/monocle/internal/types"
)

func TestGetReviewStatusInfo_NoFeedback(t *testing.T) {
	e := &Engine{
		feedback:    NewFeedbackQueue(),
		subscribers: make(map[EventKind]map[int]EventCallback),
	}
	e.current = &types.ReviewSession{}

	info := e.GetReviewStatusInfo()
	if info.Status != "no_feedback" {
		t.Errorf("expected no_feedback, got %q", info.Status)
	}
}

func TestGetReviewStatusInfo_Pending(t *testing.T) {
	e := &Engine{
		feedback:    NewFeedbackQueue(),
		subscribers: make(map[EventKind]map[int]EventCallback),
	}
	e.current = &types.ReviewSession{
		Comments: []types.ReviewComment{
			{ID: "c1", Outdated: false},
			{ID: "c2", Outdated: false},
		},
	}

	e.feedback.Submit(&FormattedReview{
		Formatted:    "review",
		CommentCount: 2,
		Action:       "request_changes",
	})

	info := e.GetReviewStatusInfo()
	if info.Status != "pending" {
		t.Errorf("expected pending, got %q", info.Status)
	}
	if info.CommentCount != 2 {
		t.Errorf("expected 2 comments, got %d", info.CommentCount)
	}
}

func TestGetReviewStatusInfo_PauseRequested(t *testing.T) {
	e := &Engine{
		feedback:    NewFeedbackQueue(),
		subscribers: make(map[EventKind]map[int]EventCallback),
	}
	e.current = &types.ReviewSession{}

	e.feedback.SetPauseRequested(true)

	info := e.GetReviewStatusInfo()
	if info.Status != "pause_requested" {
		t.Errorf("expected pause_requested, got %q", info.Status)
	}
}

func TestSubmitContentForReview(t *testing.T) {
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	e := &Engine{
		feedback:    NewFeedbackQueue(),
		database:    database,
		subscribers: make(map[EventKind]map[int]EventCallback),
	}
	e.current = &types.ReviewSession{
		ID:           "sess-1",
		FileStatuses: make(map[string]bool),
	}

	// Submit content
	err = e.SubmitContentForReview("plan", "Implementation Plan", "# Plan\n1. Step one", "markdown")
	if err != nil {
		t.Fatalf("SubmitContentForReview: %v", err)
	}

	// Verify content item was added
	e.mu.RLock()
	items := e.current.ContentItems
	e.mu.RUnlock()
	if len(items) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(items))
	}
	if items[0].ID != "plan" {
		t.Errorf("expected content item ID 'plan', got %q", items[0].ID)
	}

	// Update same content item
	err = e.SubmitContentForReview("plan", "Updated Plan", "# Updated Plan\n1. New step", "markdown")
	if err != nil {
		t.Fatalf("update SubmitContentForReview: %v", err)
	}

	e.mu.RLock()
	items = e.current.ContentItems
	e.mu.RUnlock()
	if len(items) != 1 {
		t.Fatalf("expected 1 content item after update, got %d", len(items))
	}
	if items[0].Title != "Updated Plan" {
		t.Errorf("expected updated title, got %q", items[0].Title)
	}
}

func TestRequestPauseAndCancel(t *testing.T) {
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	e := &Engine{
		feedback:    NewFeedbackQueue(),
		database:    database,
		subscribers: make(map[EventKind]map[int]EventCallback),
	}
	e.current = &types.ReviewSession{
		ID:           "sess-1",
		AgentStatus:  types.AgentStatusWorking,
		FileStatuses: make(map[string]bool),
	}

	// Request pause
	e.RequestPause()

	if !e.feedback.IsPauseRequested() {
		t.Error("expected pause requested")
	}
	if e.current.AgentStatus != types.AgentStatusPaused {
		t.Errorf("expected Paused status, got %q", e.current.AgentStatus)
	}

	// Cancel pause
	e.CancelPause()

	if e.feedback.IsPauseRequested() {
		t.Error("expected pause cancelled")
	}
	if e.current.AgentStatus != types.AgentStatusWorking {
		t.Errorf("expected Working status, got %q", e.current.AgentStatus)
	}
}

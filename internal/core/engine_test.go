package core

import (
	"sync"
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

func TestAddComment(t *testing.T) {
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
	if err := database.CreateSession(e.current); err != nil {
		t.Fatalf("create session: %v", err)
	}

	target := CommentTarget{
		TargetType: types.TargetFile,
		TargetRef:  "main.go",
		LineStart:  10,
		LineEnd:    12,
	}
	comment, err := e.AddComment(target, types.CommentIssue, "This needs fixing")
	if err != nil {
		t.Fatalf("AddComment: %v", err)
	}

	if comment.Body != "This needs fixing" {
		t.Errorf("expected body 'This needs fixing', got %q", comment.Body)
	}
	if comment.TargetRef != "main.go" {
		t.Errorf("expected target ref 'main.go', got %q", comment.TargetRef)
	}
	if comment.Type != types.CommentIssue {
		t.Errorf("expected type 'issue', got %q", comment.Type)
	}

	// Verify in-memory
	if len(e.current.Comments) != 1 {
		t.Fatalf("expected 1 comment in memory, got %d", len(e.current.Comments))
	}
	if e.current.Comments[0].ID != comment.ID {
		t.Errorf("in-memory comment ID mismatch")
	}

	// Verify in DB
	dbComments, err := database.GetComments("sess-1")
	if err != nil {
		t.Fatalf("GetComments: %v", err)
	}
	if len(dbComments) != 1 {
		t.Fatalf("expected 1 comment in DB, got %d", len(dbComments))
	}
	if dbComments[0].Body != "This needs fixing" {
		t.Errorf("DB comment body mismatch: %q", dbComments[0].Body)
	}
}

func TestEditComment(t *testing.T) {
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
	if err := database.CreateSession(e.current); err != nil {
		t.Fatalf("create session: %v", err)
	}

	// Add a comment first
	target := CommentTarget{
		TargetType: types.TargetFile,
		TargetRef:  "main.go",
		LineStart:  5,
		LineEnd:    5,
	}
	comment, err := e.AddComment(target, types.CommentSuggestion, "Original body")
	if err != nil {
		t.Fatalf("AddComment: %v", err)
	}

	// Edit the comment
	edited, err := e.EditComment(comment.ID, "Updated body")
	if err != nil {
		t.Fatalf("EditComment: %v", err)
	}
	if edited.Body != "Updated body" {
		t.Errorf("expected edited body 'Updated body', got %q", edited.Body)
	}

	// Verify in-memory
	if e.current.Comments[0].Body != "Updated body" {
		t.Errorf("in-memory body not updated: %q", e.current.Comments[0].Body)
	}

	// Verify in DB
	dbComments, err := database.GetComments("sess-1")
	if err != nil {
		t.Fatalf("GetComments: %v", err)
	}
	if dbComments[0].Body != "Updated body" {
		t.Errorf("DB body not updated: %q", dbComments[0].Body)
	}
}

func TestDeleteComment(t *testing.T) {
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
	if err := database.CreateSession(e.current); err != nil {
		t.Fatalf("create session: %v", err)
	}

	// Add a comment
	target := CommentTarget{
		TargetType: types.TargetFile,
		TargetRef:  "main.go",
		LineStart:  1,
		LineEnd:    1,
	}
	comment, err := e.AddComment(target, types.CommentNote, "A note")
	if err != nil {
		t.Fatalf("AddComment: %v", err)
	}

	// Delete the comment
	if err := e.DeleteComment(comment.ID); err != nil {
		t.Fatalf("DeleteComment: %v", err)
	}

	// Verify in-memory
	if len(e.current.Comments) != 0 {
		t.Errorf("expected 0 comments in memory, got %d", len(e.current.Comments))
	}

	// Verify in DB
	dbComments, err := database.GetComments("sess-1")
	if err != nil {
		t.Fatalf("GetComments: %v", err)
	}
	if len(dbComments) != 0 {
		t.Errorf("expected 0 comments in DB, got %d", len(dbComments))
	}
}

func TestDismissOutdated(t *testing.T) {
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
	if err := database.CreateSession(e.current); err != nil {
		t.Fatalf("create session: %v", err)
	}

	// Add two comments
	target := CommentTarget{
		TargetType: types.TargetFile,
		TargetRef:  "main.go",
		LineStart:  1,
		LineEnd:    1,
	}
	_, err = e.AddComment(target, types.CommentIssue, "Comment 1")
	if err != nil {
		t.Fatalf("AddComment c1: %v", err)
	}
	_, err = e.AddComment(target, types.CommentNote, "Comment 2")
	if err != nil {
		t.Fatalf("AddComment c2: %v", err)
	}

	// Mark all as outdated in DB, then reload into e.current
	if err := database.MarkOutdated("sess-1"); err != nil {
		t.Fatalf("MarkOutdated: %v", err)
	}

	// Reload comments from DB into memory to get the outdated flag
	dbComments, err := database.GetComments("sess-1")
	if err != nil {
		t.Fatalf("GetComments: %v", err)
	}
	e.current.Comments = dbComments

	// Verify both are outdated
	for _, c := range e.current.Comments {
		if !c.Outdated {
			t.Fatalf("expected comment %s to be outdated", c.ID)
		}
	}

	// Now add a fresh (non-outdated) comment
	c3, err := e.AddComment(target, types.CommentPraise, "Comment 3 - fresh")
	if err != nil {
		t.Fatalf("AddComment c3: %v", err)
	}

	// Dismiss outdated
	if err := e.DismissOutdated(); err != nil {
		t.Fatalf("DismissOutdated: %v", err)
	}

	// Verify in-memory: only c3 remains
	if len(e.current.Comments) != 1 {
		t.Fatalf("expected 1 comment in memory after dismiss, got %d", len(e.current.Comments))
	}
	if e.current.Comments[0].ID != c3.ID {
		t.Errorf("expected remaining comment to be c3 (%s), got %s", c3.ID, e.current.Comments[0].ID)
	}

	// Verify in DB: only c3 remains
	dbComments, err = database.GetComments("sess-1")
	if err != nil {
		t.Fatalf("GetComments after dismiss: %v", err)
	}
	if len(dbComments) != 1 {
		t.Fatalf("expected 1 comment in DB after dismiss, got %d", len(dbComments))
	}
	if dbComments[0].ID != c3.ID {
		t.Errorf("expected DB comment to be c3, got %s", dbComments[0].ID)
	}
}

func TestClearComments(t *testing.T) {
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
	if err := database.CreateSession(e.current); err != nil {
		t.Fatalf("create session: %v", err)
	}

	// Add two comments (both will initially be non-outdated)
	target := CommentTarget{
		TargetType: types.TargetFile,
		TargetRef:  "main.go",
		LineStart:  1,
		LineEnd:    1,
	}
	_, err = e.AddComment(target, types.CommentIssue, "Active comment")
	if err != nil {
		t.Fatalf("AddComment active: %v", err)
	}
	_, err = e.AddComment(target, types.CommentNote, "Will be outdated")
	if err != nil {
		t.Fatalf("AddComment outdated: %v", err)
	}

	// Mark all as outdated in DB
	if err := database.MarkOutdated("sess-1"); err != nil {
		t.Fatalf("MarkOutdated: %v", err)
	}
	// Reload from DB to get correct outdated flags
	dbComments, err := database.GetComments("sess-1")
	if err != nil {
		t.Fatalf("GetComments: %v", err)
	}
	e.current.Comments = dbComments

	// Now add a fresh active comment (after marking outdated, so it stays active)
	_, err = e.AddComment(target, types.CommentSuggestion, "Fresh active comment")
	if err != nil {
		t.Fatalf("AddComment fresh: %v", err)
	}

	// Clear active (non-outdated) comments
	if err := e.ClearComments(); err != nil {
		t.Fatalf("ClearComments: %v", err)
	}

	// Verify in-memory: only outdated comments remain
	if len(e.current.Comments) != 2 {
		t.Fatalf("expected 2 outdated comments in memory, got %d", len(e.current.Comments))
	}
	for _, c := range e.current.Comments {
		if !c.Outdated {
			t.Errorf("expected remaining comment %s to be outdated", c.ID)
		}
	}

	// Verify in DB: only outdated comments remain
	dbComments, err = database.GetComments("sess-1")
	if err != nil {
		t.Fatalf("GetComments after clear: %v", err)
	}
	if len(dbComments) != 2 {
		t.Fatalf("expected 2 comments in DB after clear, got %d", len(dbComments))
	}
	for _, c := range dbComments {
		if !c.Outdated {
			t.Errorf("expected DB comment %s to be outdated", c.ID)
		}
	}
}

func TestMarkReviewedAndUnmark(t *testing.T) {
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
		ChangedFiles: []types.ChangedFile{
			{Path: "main.go", Status: types.FileModified, Reviewed: false},
			{Path: "util.go", Status: types.FileAdded, Reviewed: false},
		},
	}
	if err := database.CreateSession(e.current); err != nil {
		t.Fatalf("create session: %v", err)
	}
	// Insert the changed files into DB so MarkFileReviewed has rows to update
	for i := range e.current.ChangedFiles {
		if err := database.UpsertChangedFile("sess-1", &e.current.ChangedFiles[i]); err != nil {
			t.Fatalf("upsert changed file: %v", err)
		}
	}

	// Mark main.go as reviewed
	if err := e.MarkReviewed("main.go"); err != nil {
		t.Fatalf("MarkReviewed: %v", err)
	}

	// Verify FileStatuses
	if !e.current.FileStatuses["main.go"] {
		t.Error("expected FileStatuses['main.go'] to be true")
	}

	// Verify ChangedFiles
	for _, f := range e.current.ChangedFiles {
		if f.Path == "main.go" && !f.Reviewed {
			t.Error("expected main.go ChangedFile.Reviewed to be true")
		}
	}

	// Unmark main.go
	if err := e.UnmarkReviewed("main.go"); err != nil {
		t.Fatalf("UnmarkReviewed: %v", err)
	}

	// Verify FileStatuses is false
	if e.current.FileStatuses["main.go"] {
		t.Error("expected FileStatuses['main.go'] to be false after unmark")
	}

	// Verify ChangedFiles
	for _, f := range e.current.ChangedFiles {
		if f.Path == "main.go" && f.Reviewed {
			t.Error("expected main.go ChangedFile.Reviewed to be false after unmark")
		}
	}
}

func TestGetReviewSummary(t *testing.T) {
	e := &Engine{
		feedback:    NewFeedbackQueue(),
		subscribers: make(map[EventKind]map[int]EventCallback),
	}
	e.current = &types.ReviewSession{
		ID:           "sess-1",
		FileStatuses: make(map[string]bool),
		Comments: []types.ReviewComment{
			{ID: "c1", TargetType: types.TargetFile, TargetRef: "main.go", Type: types.CommentIssue, Body: "Bug here"},
			{ID: "c2", TargetType: types.TargetFile, TargetRef: "main.go", Type: types.CommentSuggestion, Body: "Consider this"},
			{ID: "c3", TargetType: types.TargetFile, TargetRef: "util.go", Type: types.CommentNote, Body: "FYI"},
			{ID: "c4", TargetType: types.TargetContent, TargetRef: "plan-1", Type: types.CommentPraise, Body: "Nice plan"},
			{ID: "c5", TargetType: types.TargetFile, TargetRef: "main.go", Type: types.CommentIssue, Body: "Outdated issue", Outdated: true},
		},
	}

	summary, err := e.GetReviewSummary()
	if err != nil {
		t.Fatalf("GetReviewSummary: %v", err)
	}

	// Verify counts (outdated comments are excluded from counts)
	if summary.IssueCt != 1 {
		t.Errorf("expected 1 issue, got %d", summary.IssueCt)
	}
	if summary.SuggestionCt != 1 {
		t.Errorf("expected 1 suggestion, got %d", summary.SuggestionCt)
	}
	if summary.NoteCt != 1 {
		t.Errorf("expected 1 note, got %d", summary.NoteCt)
	}
	if summary.PraiseCt != 1 {
		t.Errorf("expected 1 praise, got %d", summary.PraiseCt)
	}

	// Verify file groupings (outdated excluded)
	mainComments := summary.FileComments["main.go"]
	if len(mainComments) != 2 {
		t.Errorf("expected 2 comments on main.go, got %d", len(mainComments))
	}
	utilComments := summary.FileComments["util.go"]
	if len(utilComments) != 1 {
		t.Errorf("expected 1 comment on util.go, got %d", len(utilComments))
	}

	// Verify content groupings
	planComments := summary.ContentComments["plan-1"]
	if len(planComments) != 1 {
		t.Errorf("expected 1 comment on plan-1, got %d", len(planComments))
	}
}

func TestSubmit(t *testing.T) {
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	formatter := NewReviewFormatter(func(path string, start, end int) string {
		return ""
	})

	e := &Engine{
		feedback:    NewFeedbackQueue(),
		database:    database,
		formatter:   formatter,
		subscribers: make(map[EventKind]map[int]EventCallback),
	}
	e.current = &types.ReviewSession{
		ID:           "sess-1",
		FileStatuses: make(map[string]bool),
		ReviewRound:  1,
		Comments: []types.ReviewComment{
			{
				ID:         "c1",
				TargetType: types.TargetFile,
				TargetRef:  "main.go",
				LineStart:  10,
				LineEnd:    10,
				Type:       types.CommentIssue,
				Body:       "Fix this bug",
			},
		},
	}
	if err := database.CreateSession(e.current); err != nil {
		t.Fatalf("create session: %v", err)
	}

	// Subscribe to events before submitting
	var receivedEvents []EventPayload
	var eventMu sync.Mutex
	e.On(EventFeedbackSubmitted, func(payload EventPayload) {
		eventMu.Lock()
		receivedEvents = append(receivedEvents, payload)
		eventMu.Unlock()
	})

	// Submit the review
	if err := e.Submit(types.ActionRequestChanges, "Please fix the issues"); err != nil {
		t.Fatalf("Submit: %v", err)
	}

	// Verify feedback is pending
	if !e.feedback.HasPending() {
		t.Error("expected feedback to be pending after submit")
	}

	// Poll the feedback and verify it's non-empty
	review := e.feedback.Poll()
	if review == nil {
		t.Fatal("expected non-nil review from poll")
	}
	if review.Formatted == "" {
		t.Error("expected non-empty formatted review")
	}
	if review.Action != string(types.ActionRequestChanges) {
		t.Errorf("expected action 'request_changes', got %q", review.Action)
	}

	// Verify event was emitted
	eventMu.Lock()
	eventCount := len(receivedEvents)
	eventMu.Unlock()
	if eventCount != 1 {
		t.Errorf("expected 1 EventFeedbackSubmitted event, got %d", eventCount)
	}
}

func TestSetBaseRef(t *testing.T) {
	dir, baseRef := setupTestRepo(t)

	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	e := &Engine{
		feedback:       NewFeedbackQueue(),
		database:       database,
		git:            NewGitClient(dir),
		autoAdvanceRef: true,
		subscribers:    make(map[EventKind]map[int]EventCallback),
	}
	e.current = &types.ReviewSession{
		ID:           "sess-1",
		BaseRef:      baseRef,
		RepoRoot:     dir,
		FileStatuses: make(map[string]bool),
	}
	if err := database.CreateSession(e.current); err != nil {
		t.Fatalf("create session: %v", err)
	}

	// SetBaseRef with the baseRef (which is HEAD, the only commit)
	if err := e.SetBaseRef(baseRef); err != nil {
		t.Fatalf("SetBaseRef: %v", err)
	}

	// The base should be set (either to parent or to ref itself for root commit)
	if e.current.BaseRef == "" {
		t.Error("expected BaseRef to be set")
	}

	// Auto-advance should be disabled
	if e.autoAdvanceRef {
		t.Error("expected autoAdvanceRef to be false after SetBaseRef")
	}
}

func TestSetAutoAdvanceRef(t *testing.T) {
	e := &Engine{
		feedback:       NewFeedbackQueue(),
		autoAdvanceRef: false,
		subscribers:    make(map[EventKind]map[int]EventCallback),
	}

	// Verify initial state
	if e.IsAutoAdvanceRef() {
		t.Error("expected IsAutoAdvanceRef to be false initially")
	}

	// Enable auto-advance
	e.SetAutoAdvanceRef(true)
	if !e.IsAutoAdvanceRef() {
		t.Error("expected IsAutoAdvanceRef to be true after SetAutoAdvanceRef(true)")
	}

	// Disable auto-advance
	e.SetAutoAdvanceRef(false)
	if e.IsAutoAdvanceRef() {
		t.Error("expected IsAutoAdvanceRef to be false after SetAutoAdvanceRef(false)")
	}
}

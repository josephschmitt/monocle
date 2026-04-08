package core

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/josephschmitt/monocle/internal/db"
	"github.com/josephschmitt/monocle/internal/types"
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
			{ID: "c1"},
			{ID: "c2"},
		},
	}

	e.feedback.Submit(&FormattedReview{
		Formatted:    "review",
		CommentCount: 2,
		Action:       "request_changes",
	}, false)

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
	err = e.SubmitContentForReview("plan", "Implementation Plan", "# Plan\n1. Step one", "markdown", true)
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
	err = e.SubmitContentForReview("plan", "Updated Plan", "# Updated Plan\n1. New step", "markdown", true)
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
		FileStatuses: make(map[string]bool),
	}

	// Request pause
	e.RequestPause()

	if !e.feedback.IsPauseRequested() {
		t.Error("expected pause requested")
	}

	// Cancel pause
	e.CancelPause()

	if e.feedback.IsPauseRequested() {
		t.Error("expected pause cancelled")
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
	edited, err := e.EditComment(comment.ID, types.CommentIssue, "Updated body")
	if err != nil {
		t.Fatalf("EditComment: %v", err)
	}
	if edited.Body != "Updated body" {
		t.Errorf("expected edited body 'Updated body', got %q", edited.Body)
	}
	if edited.Type != types.CommentIssue {
		t.Errorf("expected edited type %q, got %q", types.CommentIssue, edited.Type)
	}

	// Verify in-memory
	if e.current.Comments[0].Body != "Updated body" {
		t.Errorf("in-memory body not updated: %q", e.current.Comments[0].Body)
	}
	if e.current.Comments[0].Type != types.CommentIssue {
		t.Errorf("in-memory type not updated: %q", e.current.Comments[0].Type)
	}

	// Verify in DB
	dbComments, err := database.GetComments("sess-1")
	if err != nil {
		t.Fatalf("GetComments: %v", err)
	}
	if dbComments[0].Body != "Updated body" {
		t.Errorf("DB body not updated: %q", dbComments[0].Body)
	}
	if dbComments[0].Type != types.CommentIssue {
		t.Errorf("DB type not updated: %q", dbComments[0].Type)
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

	target := CommentTarget{
		TargetType: types.TargetFile,
		TargetRef:  "main.go",
		LineStart:  1,
		LineEnd:    1,
	}
	_, err = e.AddComment(target, types.CommentIssue, "Comment 1")
	if err != nil {
		t.Fatalf("AddComment: %v", err)
	}
	_, err = e.AddComment(target, types.CommentNote, "Comment 2")
	if err != nil {
		t.Fatalf("AddComment: %v", err)
	}

	if err := e.ClearComments(); err != nil {
		t.Fatalf("ClearComments: %v", err)
	}

	// Verify in-memory: no comments remain
	if len(e.current.Comments) != 0 {
		t.Fatalf("expected 0 comments in memory, got %d", len(e.current.Comments))
	}

	// Verify in DB: no comments remain
	dbComments, err := database.GetComments("sess-1")
	if err != nil {
		t.Fatalf("GetComments after clear: %v", err)
	}
	if len(dbComments) != 0 {
		t.Fatalf("expected 0 comments in DB, got %d", len(dbComments))
	}
}

func TestClearReview(t *testing.T) {
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
		ReviewRound:  1,
		ChangedFiles: []types.ChangedFile{
			{Path: "main.go", Status: types.FileModified, Reviewed: true},
		},
	}
	e.current.FileStatuses["main.go"] = true

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
	if _, err := e.AddComment(target, types.CommentIssue, "fix this"); err != nil {
		t.Fatalf("AddComment: %v", err)
	}

	// Add a content item (plan)
	item := &types.ContentItem{
		ID:      "plan-1",
		Title:   "Plan",
		Content: "some plan",
		IsPlan:  true,
	}
	if err := database.UpsertContentItem("sess-1", item); err != nil {
		t.Fatalf("UpsertContentItem: %v", err)
	}
	e.current.ContentItems = []types.ContentItem{*item}

	// Mark file as reviewed in DB
	if err := database.MarkFileReviewed("sess-1", "main.go", true); err != nil {
		t.Fatalf("MarkFileReviewed: %v", err)
	}

	// Clear the review
	if err := e.ClearReview(); err != nil {
		t.Fatalf("ClearReview: %v", err)
	}

	// Verify comments cleared
	if len(e.current.Comments) != 0 {
		t.Errorf("expected 0 comments in memory, got %d", len(e.current.Comments))
	}
	dbComments, _ := database.GetComments("sess-1")
	if len(dbComments) != 0 {
		t.Errorf("expected 0 comments in DB, got %d", len(dbComments))
	}

	// Verify content items cleared
	if len(e.current.ContentItems) != 0 {
		t.Errorf("expected 0 content items in memory, got %d", len(e.current.ContentItems))
	}
	dbItems, _ := database.GetContentItems("sess-1")
	if len(dbItems) != 0 {
		t.Errorf("expected 0 content items in DB, got %d", len(dbItems))
	}

	// Verify reviewed states reset
	if e.current.ChangedFiles[0].Reviewed {
		t.Error("expected file reviewed=false")
	}
	if e.current.FileStatuses["main.go"] {
		t.Error("expected file status=false")
	}

	// Verify round NOT advanced
	if e.current.ReviewRound != 1 {
		t.Errorf("expected round to stay at 1, got %d", e.current.ReviewRound)
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
		},
	}

	summary, err := e.GetReviewSummary()
	if err != nil {
		t.Fatalf("GetReviewSummary: %v", err)
	}

	// Verify counts
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

	// Verify file groupings
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
	}, types.ReviewFormatConfig{IncludeSnippets: true, MaxSnippetLines: 10, IncludeSummary: true})

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
	if _, err := e.Submit(types.ActionRequestChanges, "Please fix the issues"); err != nil {
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

func TestSubmitRequestChangesRequiresContent(t *testing.T) {
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	formatter := NewReviewFormatter(func(path string, start, end int) string {
		return ""
	}, types.ReviewFormatConfig{IncludeSnippets: true, MaxSnippetLines: 10, IncludeSummary: true})

	e := &Engine{
		feedback:    NewFeedbackQueue(),
		database:    database,
		formatter:   formatter,
		subscribers: make(map[EventKind]map[int]EventCallback),
	}

	setup := func(comments []types.ReviewComment) {
		e.current = &types.ReviewSession{
			ID:           "sess-1",
			FileStatuses: make(map[string]bool),
			ReviewRound:  1,
			Comments:     comments,
		}
		_ = database.CreateSession(e.current)
	}

	t.Run("reject empty request_changes", func(t *testing.T) {
		setup(nil)
		_, err := e.Submit(types.ActionRequestChanges, "")
		if err == nil {
			t.Error("expected error for empty request_changes")
		}
	})

	t.Run("reject request_changes with only resolved comments", func(t *testing.T) {
		setup([]types.ReviewComment{
			{ID: "c1", TargetType: types.TargetFile, TargetRef: "main.go", Type: types.CommentIssue, Body: "Bug", Resolved: true},
		})
		_, err := e.Submit(types.ActionRequestChanges, "")
		if err == nil {
			t.Error("expected error for request_changes with only resolved comments")
		}
	})

	t.Run("accept request_changes with body", func(t *testing.T) {
		setup(nil)
		_, err := e.Submit(types.ActionRequestChanges, "Please fix")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("accept request_changes with unresolved comment", func(t *testing.T) {
		setup([]types.ReviewComment{
			{ID: "c1", TargetType: types.TargetFile, TargetRef: "main.go", Type: types.CommentIssue, Body: "Bug"},
		})
		_, err := e.Submit(types.ActionRequestChanges, "")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("accept empty approve", func(t *testing.T) {
		setup(nil)
		_, err := e.Submit(types.ActionApprove, "")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestGetReviewSummaryExcludesResolved(t *testing.T) {
	e := &Engine{
		feedback:    NewFeedbackQueue(),
		subscribers: make(map[EventKind]map[int]EventCallback),
	}
	e.current = &types.ReviewSession{
		ID:           "sess-1",
		FileStatuses: make(map[string]bool),
		Comments: []types.ReviewComment{
			{ID: "c1", TargetType: types.TargetFile, TargetRef: "main.go", Type: types.CommentIssue, Body: "Bug"},
			{ID: "c2", TargetType: types.TargetFile, TargetRef: "main.go", Type: types.CommentIssue, Body: "Resolved bug", Resolved: true},
			{ID: "c3", TargetType: types.TargetFile, TargetRef: "util.go", Type: types.CommentSuggestion, Body: "Resolved suggestion", Resolved: true},
			{ID: "c4", TargetType: types.TargetContent, TargetRef: "plan-1", Type: types.CommentNote, Body: "Note"},
		},
	}

	summary, err := e.GetReviewSummary()
	if err != nil {
		t.Fatalf("GetReviewSummary: %v", err)
	}

	if summary.IssueCt != 1 {
		t.Errorf("expected 1 issue, got %d", summary.IssueCt)
	}
	if summary.SuggestionCt != 0 {
		t.Errorf("expected 0 suggestions, got %d", summary.SuggestionCt)
	}
	if summary.NoteCt != 1 {
		t.Errorf("expected 1 note, got %d", summary.NoteCt)
	}

	// Resolved comments should not appear in file groupings
	mainComments := summary.FileComments["main.go"]
	if len(mainComments) != 1 {
		t.Errorf("expected 1 comment on main.go (resolved excluded), got %d", len(mainComments))
	}
	if _, ok := summary.FileComments["util.go"]; ok {
		t.Error("expected no comments on util.go (all resolved)")
	}
}

func TestSetBaseRef(t *testing.T) {
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	resolvedRef := "resolved123resolved123resolved123resolved"
	stub := &gitStub{
		repoRoot:   "/tmp/repo",
		currentRef: resolvedRef,
	}

	now := time.Now()
	session := &types.ReviewSession{
		ID: "sess-1", Agent: "claude",
		RepoRoot: "/tmp/repo", BaseRef: "old-base", ReviewRound: 1,
		FileStatuses: make(map[string]bool), CreatedAt: now, UpdatedAt: now,
	}
	database.CreateSession(session)

	e := &Engine{
		feedback:       NewFeedbackQueue(),
		database:       database,
		git:            stub,
		autoAdvanceRef: true,
		subscribers:    make(map[EventKind]map[int]EventCallback),
	}
	e.current = session

	if err := e.SetBaseRef("some-ref"); err != nil {
		t.Fatalf("SetBaseRef: %v", err)
	}

	if e.current.BaseRef != resolvedRef {
		t.Errorf("expected BaseRef %q, got %q", resolvedRef, e.current.BaseRef)
	}

	if e.autoAdvanceRef {
		t.Error("expected autoAdvanceRef to be false after SetBaseRef")
	}
}

func TestSetAutoAdvanceRef(t *testing.T) {
	e := &Engine{
		feedback:       NewFeedbackQueue(),
		autoAdvanceRef: false,
		lastKnownHead:  "abc123",
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
	if e.lastKnownHead != "" {
		t.Errorf("expected lastKnownHead to be reset to empty, got %q", e.lastKnownHead)
	}

	// Disable auto-advance
	e.SetAutoAdvanceRef(false)
	if e.IsAutoAdvanceRef() {
		t.Error("expected IsAutoAdvanceRef to be false after SetAutoAdvanceRef(false)")
	}
}

func TestAutoAdvanceRefAfterManualSelect(t *testing.T) {
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	headRef := "head123head123head123head123head123head12"
	manualRef := "manual456manual456manual456manual456manual"

	stub := &gitStub{
		repoRoot:   "/tmp/repo",
		currentRef: headRef,
	}

	now := time.Now()
	session := &types.ReviewSession{
		ID: "sess-1", Agent: "claude",
		RepoRoot: "/tmp/repo", BaseRef: headRef, ReviewRound: 1,
		FileStatuses: make(map[string]bool), CreatedAt: now, UpdatedAt: now,
	}
	database.CreateSession(session)

	e := &Engine{
		feedback:       NewFeedbackQueue(),
		database:       database,
		git:            stub,
		sessions:       NewSessionManager(database, stub),
		autoAdvanceRef: true,
		subscribers:    make(map[EventKind]map[int]EventCallback),
	}
	e.current = session

	// Step 1: Initial refresh in auto mode — tracks HEAD
	if _, err := e.RefreshChangedFiles(); err != nil {
		t.Fatalf("initial RefreshChangedFiles: %v", err)
	}
	if e.current.BaseRef != headRef {
		t.Errorf("step 1: expected BaseRef %q, got %q", headRef, e.current.BaseRef)
	}

	// Step 2: Manually select a different ref
	stub.currentRef = manualRef
	if err := e.SetBaseRef("some-commit"); err != nil {
		t.Fatalf("SetBaseRef: %v", err)
	}
	if e.current.BaseRef != manualRef {
		t.Errorf("step 2: expected BaseRef %q, got %q", manualRef, e.current.BaseRef)
	}
	if e.autoAdvanceRef {
		t.Error("step 2: expected autoAdvanceRef to be false")
	}

	// Step 3: Switch back to auto mode
	stub.currentRef = headRef
	e.SetAutoAdvanceRef(true)

	// Step 4: Refresh — should update BaseRef back to HEAD
	if _, err := e.RefreshChangedFiles(); err != nil {
		t.Fatalf("RefreshChangedFiles after re-enable: %v", err)
	}
	if e.current.BaseRef != headRef {
		t.Errorf("step 4: expected BaseRef to return to %q, got %q", headRef, e.current.BaseRef)
	}
}

func TestResolveComment(t *testing.T) {
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	now := time.Now()
	session := &types.ReviewSession{
		ID:           "sess-1",
		Agent:        "claude",
		RepoRoot:     "/tmp",
		BaseRef:      "abc",
		ReviewRound:  1,
		FileStatuses: make(map[string]bool),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	database.CreateSession(session)

	e := &Engine{
		feedback:    NewFeedbackQueue(),
		database:    database,
		subscribers: make(map[EventKind]map[int]EventCallback),
	}
	e.current = session

	// Add a comment
	comment, err := e.AddComment(CommentTarget{
		TargetType: types.TargetFile,
		TargetRef:  "main.go",
		LineStart:  10,
	}, types.CommentIssue, "Fix this bug")
	if err != nil {
		t.Fatalf("AddComment: %v", err)
	}

	// Initially not resolved
	if comment.Resolved {
		t.Error("comment should not be resolved initially")
	}

	// Resolve it
	if err := e.ResolveComment(comment.ID); err != nil {
		t.Fatalf("ResolveComment: %v", err)
	}

	// Verify in-memory
	e.mu.RLock()
	resolved := false
	for _, c := range e.current.Comments {
		if c.ID == comment.ID {
			resolved = c.Resolved
			break
		}
	}
	e.mu.RUnlock()
	if !resolved {
		t.Error("expected comment to be resolved in memory")
	}

	// Verify in DB
	dbComments, _ := database.GetComments("sess-1")
	if len(dbComments) != 1 || !dbComments[0].Resolved {
		t.Error("expected comment to be resolved in DB")
	}

	// Toggle back to unresolved
	if err := e.ResolveComment(comment.ID); err != nil {
		t.Fatalf("unresolve: %v", err)
	}
	dbComments, _ = database.GetComments("sess-1")
	if len(dbComments) != 1 || dbComments[0].Resolved {
		t.Error("expected comment to be unresolved after toggle")
	}
}

func TestFormatSkipsResolvedComments(t *testing.T) {
	f := NewReviewFormatter(nil, types.ReviewFormatConfig{IncludeSnippets: true, MaxSnippetLines: 10, IncludeSummary: true})
	comments := []types.ReviewComment{
		{
			ID:         "c1",
			TargetType: types.TargetFile,
			TargetRef:  "main.go",
			Type:       types.CommentIssue,
			Body:       "Active issue",
		},
		{
			ID:         "c2",
			TargetType: types.TargetFile,
			TargetRef:  "main.go",
			Type:       types.CommentIssue,
			Body:       "Resolved issue",
			Resolved:   true,
		},
	}

	result := f.Format(&types.ReviewSession{}, comments, types.ActionRequestChanges, "")

	if !strings.Contains(result.Formatted, "Active issue") {
		t.Error("active comment should be included")
	}
	if strings.Contains(result.Formatted, "Resolved issue") {
		t.Error("resolved comment should be excluded from formatted output")
	}
	if !strings.Contains(result.Formatted, "1 issue(s)") {
		t.Error("summary should only count active (non-resolved) comments")
	}
}

func TestSnapshotCreatedOnRequestChanges(t *testing.T) {
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer database.Close()
	stub := &gitStub{
		repoRoot:   "/tmp/repo",
		currentRef: "head123",
	}

	formatter := NewReviewFormatter(func(path string, start, end int) string {
		return ""
	}, types.ReviewFormatConfig{})

	now := time.Now()
	e := &Engine{
		feedback:    NewFeedbackQueue(),
		database:    database,
		git:         stub,
		formatter:   formatter,
		cfg:         &types.Config{},
		subscribers: make(map[EventKind]map[int]EventCallback),
	}
	e.current = &types.ReviewSession{
		ID:           "sess-1",
		Agent:        "claude",
		RepoRoot:     "/tmp/repo",
		BaseRef:      "base123",
		ReviewRound:  1,
		FileStatuses: make(map[string]bool),
		ChangedFiles: []types.ChangedFile{
			{Path: "main.go", Status: types.FileModified},
		},
		Comments: []types.ReviewComment{
			{ID: "c1", TargetType: types.TargetFile, TargetRef: "main.go", Type: types.CommentIssue, Body: "fix"},
		},
		CreatedAt: now, UpdatedAt: now,
	}
	database.CreateSession(e.current)

	// Submit request_changes
	if _, err := e.Submit(types.ActionRequestChanges, "please fix"); err != nil {
		t.Fatalf("Submit: %v", err)
	}

	// Verify snapshot was created
	snapshots, err := database.GetSnapshots("sess-1")
	if err != nil {
		t.Fatalf("GetSnapshots: %v", err)
	}
	if len(snapshots) != 1 {
		t.Fatalf("expected 1 snapshot, got %d", len(snapshots))
	}
	if snapshots[0].ReviewRound != 1 {
		t.Errorf("expected round 1, got %d", snapshots[0].ReviewRound)
	}
	if snapshots[0].HeadRef != "head123" {
		t.Errorf("expected head ref head123, got %q", snapshots[0].HeadRef)
	}

	// Verify snapshot files
	snap, _ := database.GetSnapshot(snapshots[0].ID)
	if len(snap.Files) != 1 {
		t.Fatalf("expected 1 file in snapshot, got %d", len(snap.Files))
	}
	if snap.Files[0].Path != "main.go" {
		t.Errorf("expected main.go, got %q", snap.Files[0].Path)
	}
}

func TestSnapshotWipedOnApprove(t *testing.T) {
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer database.Close()
	stub := &gitStub{
		repoRoot:   "/tmp/repo",
		currentRef: "head123",
	}

	formatter := NewReviewFormatter(func(path string, start, end int) string {
		return ""
	}, types.ReviewFormatConfig{})

	now := time.Now()
	e := &Engine{
		feedback:    NewFeedbackQueue(),
		database:    database,
		git:         stub,
		formatter:   formatter,
		cfg:         &types.Config{},
		subscribers: make(map[EventKind]map[int]EventCallback),
	}
	e.current = &types.ReviewSession{
		ID:           "sess-1",
		Agent:        "claude",
		RepoRoot:     "/tmp/repo",
		BaseRef:      "base123",
		ReviewRound:  1,
		FileStatuses: make(map[string]bool),
		ChangedFiles: []types.ChangedFile{
			{Path: "main.go", Status: types.FileModified},
		},
		Comments: []types.ReviewComment{
			{ID: "c1", TargetType: types.TargetFile, TargetRef: "main.go", Type: types.CommentIssue, Body: "fix"},
		},
		CreatedAt: now, UpdatedAt: now,
	}
	database.CreateSession(e.current)

	// First, submit request_changes to create a snapshot
	e.Submit(types.ActionRequestChanges, "fix please")

	// Verify snapshot exists
	has, _ := database.HasSnapshots("sess-1")
	if !has {
		t.Fatal("expected snapshot after request_changes")
	}

	// Now submit approve
	e.current.Comments = nil // clear comments for approve
	e.Submit(types.ActionApprove, "")

	// Verify snapshots were wiped
	has, _ = database.HasSnapshots("sess-1")
	if has {
		t.Error("expected snapshots to be wiped after approve")
	}
}

func TestSnapshotWipedOnSetBaseRef(t *testing.T) {
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer database.Close()
	stub := &gitStub{
		repoRoot:   "/tmp/repo",
		currentRef: "resolved123",
	}

	now := time.Now()
	e := &Engine{
		feedback:       NewFeedbackQueue(),
		database:       database,
		git:            stub,
		autoAdvanceRef: true,
		subscribers:    make(map[EventKind]map[int]EventCallback),
	}
	e.current = &types.ReviewSession{
		ID: "sess-1", Agent: "claude",
		RepoRoot: "/tmp/repo", BaseRef: "old-base", ReviewRound: 1,
		FileStatuses: make(map[string]bool), CreatedAt: now, UpdatedAt: now,
	}
	database.CreateSession(e.current)

	// Manually insert a snapshot
	database.CreateSubmission("sess-1", &types.ReviewSubmission{
		ID: "sub-1", SessionID: "sess-1", Action: types.ActionRequestChanges,
		FormattedReview: "review", ReviewRound: 1, SubmittedAt: now,
	})
	database.CreateSnapshot("sess-1", "sub-1", 1, "head1", "base1", []types.SnapshotFile{
		{Path: "main.go", Status: types.FileModified, BlobSHA: "sha1"},
	})

	has, _ := database.HasSnapshots("sess-1")
	if !has {
		t.Fatal("expected snapshot to exist")
	}

	// Change base ref
	if err := e.SetBaseRef("some-ref"); err != nil {
		t.Fatalf("SetBaseRef: %v", err)
	}

	// Verify snapshots were wiped
	has, _ = database.HasSnapshots("sess-1")
	if has {
		t.Error("expected snapshots to be wiped after SetBaseRef")
	}
}

func TestMarkReviewedOnSubmitConfig(t *testing.T) {
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer database.Close()
	stub := &gitStub{
		repoRoot:   "/tmp/repo",
		currentRef: "head123",
	}

	formatter := NewReviewFormatter(func(path string, start, end int) string {
		return ""
	}, types.ReviewFormatConfig{})

	now := time.Now()

	// Test "commented" mode: only files with comments get marked
	e := &Engine{
		feedback:    NewFeedbackQueue(),
		database:    database,
		git:         stub,
		formatter:   formatter,
		cfg:         &types.Config{MarkReviewedOnSubmit: "commented"},
		subscribers: make(map[EventKind]map[int]EventCallback),
	}
	e.current = &types.ReviewSession{
		ID: "sess-1", Agent: "claude", RepoRoot: "/tmp/repo",
		BaseRef: "base", ReviewRound: 1,
		FileStatuses: make(map[string]bool),
		ChangedFiles: []types.ChangedFile{
			{Path: "main.go", Status: types.FileModified},
			{Path: "utils.go", Status: types.FileModified},
		},
		Comments: []types.ReviewComment{
			{ID: "c1", TargetType: types.TargetFile, TargetRef: "main.go", Type: types.CommentIssue, Body: "fix"},
		},
		CreatedAt: now, UpdatedAt: now,
	}
	database.CreateSession(e.current)
	database.UpsertChangedFile("sess-1", &e.current.ChangedFiles[0])
	database.UpsertChangedFile("sess-1", &e.current.ChangedFiles[1])

	e.Submit(types.ActionRequestChanges, "fix")

	// Check the snapshot — only main.go should be marked as reviewed
	snapshots, _ := database.GetSnapshots("sess-1")
	if len(snapshots) == 0 {
		t.Fatal("expected a snapshot")
	}
	snap, _ := database.GetSnapshot(snapshots[0].ID)
	for _, f := range snap.Files {
		if f.Path == "main.go" && !f.Reviewed {
			t.Error("expected main.go to be reviewed (has comment)")
		}
		if f.Path == "utils.go" && f.Reviewed {
			t.Error("expected utils.go to NOT be reviewed (no comment)")
		}
	}
}

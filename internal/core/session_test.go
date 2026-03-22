package core

import (
	"testing"
	"time"

	"github.com/anthropics/monocle/internal/db"
	"github.com/anthropics/monocle/internal/types"
)

func newTestSessionManager(t *testing.T) (*SessionManager, *gitStub) {
	t.Helper()
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	stub := &gitStub{
		repoRoot:   "/tmp/test-repo",
		currentRef: "abc123def456abc123def456abc123def456abc1",
		files: []types.ChangedFile{
			{Path: "hello.go", Status: types.FileModified},
			{Path: "world.go", Status: types.FileAdded},
		},
	}
	return NewSessionManager(database, stub), stub
}

func TestCreateSession(t *testing.T) {
	sm, stub := newTestSessionManager(t)

	session, err := sm.CreateSession(SessionOptions{})
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	if session.ID == "" {
		t.Error("expected non-empty session ID")
	}

	if session.BaseRef != stub.currentRef {
		t.Errorf("expected BaseRef %q (HEAD), got %q", stub.currentRef, session.BaseRef)
	}

	if session.Agent != "claude" {
		t.Errorf("expected default agent 'claude', got %q", session.Agent)
	}

	if session.RepoRoot != stub.repoRoot {
		t.Errorf("expected RepoRoot %q, got %q", stub.repoRoot, session.RepoRoot)
	}

	if session.ReviewRound != 1 {
		t.Errorf("expected ReviewRound 1, got %d", session.ReviewRound)
	}

	// Verify session exists in DB
	dbSession, err := sm.db.GetSession(session.ID)
	if err != nil {
		t.Fatalf("GetSession from DB: %v", err)
	}
	if dbSession.ID != session.ID {
		t.Errorf("DB session ID mismatch: %q vs %q", dbSession.ID, session.ID)
	}
}

func TestCreateSession_WithBaseRef(t *testing.T) {
	sm, _ := newTestSessionManager(t)

	customRef := "custom-ref-hash-1234567890123456789012"
	session, err := sm.CreateSession(SessionOptions{
		BaseRef: customRef,
	})
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	if session.BaseRef != customRef {
		t.Errorf("expected BaseRef %q, got %q", customRef, session.BaseRef)
	}
}

func TestResumeSession(t *testing.T) {
	sm, _ := newTestSessionManager(t)

	// Create a session
	session, err := sm.CreateSession(SessionOptions{BaseRef: "abc123"})
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	// Add changed files to DB
	file1 := &types.ChangedFile{Path: "hello.go", Status: types.FileModified, Reviewed: true}
	file2 := &types.ChangedFile{Path: "world.go", Status: types.FileAdded, Reviewed: false}
	sm.db.UpsertChangedFile(session.ID, file1)
	sm.db.UpsertChangedFile(session.ID, file2)

	// Add a comment to DB
	now := time.Now()
	comment := &types.ReviewComment{
		ID: "comment-1", TargetType: types.TargetFile, TargetRef: "hello.go",
		LineStart: 3, LineEnd: 3, Type: types.CommentIssue, Body: "This needs fixing",
		ReviewRound: 1, CreatedAt: now, UpdatedAt: now,
	}
	sm.db.CreateComment(session.ID, comment)

	// Resume the session
	resumed, err := sm.ResumeSession(session.ID)
	if err != nil {
		t.Fatalf("ResumeSession: %v", err)
	}

	if resumed.ID != session.ID {
		t.Errorf("resumed ID mismatch")
	}
	if len(resumed.ChangedFiles) != 2 {
		t.Fatalf("expected 2 changed files, got %d", len(resumed.ChangedFiles))
	}
	if len(resumed.Comments) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(resumed.Comments))
	}
	if resumed.Comments[0].Body != "This needs fixing" {
		t.Errorf("unexpected comment body: %q", resumed.Comments[0].Body)
	}
	if !resumed.FileStatuses["hello.go"] {
		t.Error("expected hello.go to be marked as reviewed")
	}
	if resumed.FileStatuses["world.go"] {
		t.Error("expected world.go to not be marked as reviewed")
	}
}

func TestRefreshChangedFiles(t *testing.T) {
	sm, _ := newTestSessionManager(t)

	session, err := sm.CreateSession(SessionOptions{BaseRef: "abc123"})
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	// First refresh
	files, err := sm.RefreshChangedFiles(session)
	if err != nil {
		t.Fatalf("RefreshChangedFiles: %v", err)
	}

	if len(files) != 2 {
		t.Fatalf("expected 2 changed files, got %d", len(files))
	}

	// Verify session's ChangedFiles are populated
	if len(session.ChangedFiles) != 2 {
		t.Fatal("expected session.ChangedFiles to be populated")
	}

	// Verify files are in the DB
	dbFiles, _ := sm.db.GetChangedFiles(session.ID)
	if len(dbFiles) != 2 {
		t.Fatalf("expected 2 files in DB, got %d", len(dbFiles))
	}

	// Mark a file as reviewed and refresh again
	sm.db.MarkFileReviewed(session.ID, files[0].Path, true)
	session.ChangedFiles[0].Reviewed = true

	files2, err := sm.RefreshChangedFiles(session)
	if err != nil {
		t.Fatalf("RefreshChangedFiles (second): %v", err)
	}

	if !files2[0].Reviewed {
		t.Errorf("expected %s to remain reviewed after refresh", files2[0].Path)
	}
}

func TestAdvanceRound(t *testing.T) {
	sm, stub := newTestSessionManager(t)

	session, err := sm.CreateSession(SessionOptions{BaseRef: "old-base-ref"})
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	// Add a changed file
	file := &types.ChangedFile{Path: "hello.go", Status: types.FileModified}
	sm.db.UpsertChangedFile(session.ID, file)
	session.ChangedFiles = []types.ChangedFile{*file}
	session.FileStatuses = map[string]bool{"hello.go": true}

	// Add a comment
	now := time.Now()
	comment := &types.ReviewComment{
		ID: "comment-adv-1", TargetType: types.TargetFile, TargetRef: "hello.go",
		LineStart: 1, LineEnd: 1, Type: types.CommentIssue, Body: "round 1 comment",
		ReviewRound: 1, CreatedAt: now, UpdatedAt: now,
	}
	sm.db.CreateComment(session.ID, comment)

	// Advance round
	if err := sm.AdvanceRound(session); err != nil {
		t.Fatalf("AdvanceRound: %v", err)
	}

	if session.ReviewRound != 2 {
		t.Errorf("expected ReviewRound 2, got %d", session.ReviewRound)
	}
	if session.BaseRef != stub.currentRef {
		t.Errorf("expected BaseRef updated to %q, got %q", stub.currentRef, session.BaseRef)
	}
	if session.ChangedFiles != nil {
		t.Errorf("expected ChangedFiles to be nil")
	}
	if len(session.FileStatuses) != 0 {
		t.Errorf("expected FileStatuses to be empty")
	}

	// Comments in DB should be marked as outdated
	comments, _ := sm.db.GetComments(session.ID)
	if len(comments) != 1 || !comments[0].Outdated {
		t.Error("expected comment to be marked as outdated")
	}

	// Changed files should be deleted from DB
	dbFiles, _ := sm.db.GetChangedFiles(session.ID)
	if len(dbFiles) != 0 {
		t.Errorf("expected 0 changed files in DB after advance, got %d", len(dbFiles))
	}
}

func TestListSessions(t *testing.T) {
	sm, _ := newTestSessionManager(t)

	sm.CreateSession(SessionOptions{BaseRef: "abc", RepoRoot: "/repo1"})
	sm.CreateSession(SessionOptions{BaseRef: "abc", RepoRoot: "/repo1"})
	sm.CreateSession(SessionOptions{BaseRef: "abc", RepoRoot: "/repo2"})

	all, _ := sm.ListSessions(ListSessionsOptions{})
	if len(all) != 3 {
		t.Errorf("expected 3 sessions, got %d", len(all))
	}

	filtered, _ := sm.ListSessions(ListSessionsOptions{RepoRoot: "/repo1"})
	if len(filtered) != 2 {
		t.Errorf("expected 2 sessions for /repo1, got %d", len(filtered))
	}

	other, _ := sm.ListSessions(ListSessionsOptions{RepoRoot: "/repo2"})
	if len(other) != 1 {
		t.Errorf("expected 1 session for /repo2, got %d", len(other))
	}
}

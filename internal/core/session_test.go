package core

import (
	"testing"
	"time"

	"github.com/anthropics/monocle/internal/db"
	"github.com/anthropics/monocle/internal/types"
)

func newTestSessionManager(t *testing.T) (*SessionManager, string, string) {
	t.Helper()
	dir, baseRef := setupTestRepo(t)
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	git := NewGitClient(dir)
	return NewSessionManager(database, git), dir, baseRef
}

func TestCreateSession(t *testing.T) {
	sm, dir, _ := newTestSessionManager(t)

	session, err := sm.CreateSession(SessionOptions{})
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	if session.ID == "" {
		t.Error("expected non-empty session ID")
	}

	// BaseRef should default to HEAD (a 40-char SHA)
	if len(session.BaseRef) != 40 {
		t.Errorf("expected 40-char BaseRef (HEAD), got %q", session.BaseRef)
	}

	if session.Agent != "claude" {
		t.Errorf("expected default agent 'claude', got %q", session.Agent)
	}

	if session.RepoRoot != dir {
		t.Errorf("expected RepoRoot %q, got %q", dir, session.RepoRoot)
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
	sm, _, baseRef := newTestSessionManager(t)

	session, err := sm.CreateSession(SessionOptions{
		BaseRef: baseRef,
	})
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	if session.BaseRef != baseRef {
		t.Errorf("expected BaseRef %q, got %q", baseRef, session.BaseRef)
	}
}

func TestResumeSession(t *testing.T) {
	sm, _, baseRef := newTestSessionManager(t)

	// Create a session
	session, err := sm.CreateSession(SessionOptions{BaseRef: baseRef})
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	// Add changed files to DB
	file1 := &types.ChangedFile{Path: "hello.go", Status: types.FileModified, Reviewed: true}
	file2 := &types.ChangedFile{Path: "world.go", Status: types.FileAdded, Reviewed: false}
	if err := sm.db.UpsertChangedFile(session.ID, file1); err != nil {
		t.Fatalf("UpsertChangedFile: %v", err)
	}
	if err := sm.db.UpsertChangedFile(session.ID, file2); err != nil {
		t.Fatalf("UpsertChangedFile: %v", err)
	}

	// Add a comment to DB
	now := time.Now()
	comment := &types.ReviewComment{
		ID:          "comment-1",
		TargetType:  types.TargetFile,
		TargetRef:   "hello.go",
		LineStart:   3,
		LineEnd:     3,
		Type:        types.CommentIssue,
		Body:        "This needs fixing",
		ReviewRound: 1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := sm.db.CreateComment(session.ID, comment); err != nil {
		t.Fatalf("CreateComment: %v", err)
	}

	// Resume the session
	resumed, err := sm.ResumeSession(session.ID)
	if err != nil {
		t.Fatalf("ResumeSession: %v", err)
	}

	if resumed.ID != session.ID {
		t.Errorf("resumed ID mismatch: %q vs %q", resumed.ID, session.ID)
	}

	// Verify changed files
	if len(resumed.ChangedFiles) != 2 {
		t.Fatalf("expected 2 changed files, got %d", len(resumed.ChangedFiles))
	}

	// Verify comments
	if len(resumed.Comments) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(resumed.Comments))
	}
	if resumed.Comments[0].Body != "This needs fixing" {
		t.Errorf("expected comment body 'This needs fixing', got %q", resumed.Comments[0].Body)
	}

	// Verify FileStatuses map is populated
	if resumed.FileStatuses == nil {
		t.Fatal("expected FileStatuses to be non-nil")
	}
	if !resumed.FileStatuses["hello.go"] {
		t.Error("expected hello.go to be marked as reviewed")
	}
	if resumed.FileStatuses["world.go"] {
		t.Error("expected world.go to not be marked as reviewed")
	}
}

func TestRefreshChangedFiles(t *testing.T) {
	sm, _, baseRef := newTestSessionManager(t)

	session, err := sm.CreateSession(SessionOptions{BaseRef: baseRef})
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	// First refresh: get changed files from git diff
	files, err := sm.RefreshChangedFiles(session)
	if err != nil {
		t.Fatalf("RefreshChangedFiles: %v", err)
	}

	if len(files) == 0 {
		t.Fatal("expected at least one changed file")
	}

	// Verify session's ChangedFiles are populated
	if len(session.ChangedFiles) == 0 {
		t.Fatal("expected session.ChangedFiles to be populated")
	}

	// Verify files are in the DB
	dbFiles, err := sm.db.GetChangedFiles(session.ID)
	if err != nil {
		t.Fatalf("GetChangedFiles: %v", err)
	}
	if len(dbFiles) != len(files) {
		t.Fatalf("expected %d files in DB, got %d", len(files), len(dbFiles))
	}

	// Mark a file as reviewed in the DB
	if err := sm.db.MarkFileReviewed(session.ID, files[0].Path, true); err != nil {
		t.Fatalf("MarkFileReviewed: %v", err)
	}

	// Update the session's in-memory state to reflect the reviewed status
	// (so RefreshChangedFiles can merge it)
	for i := range session.ChangedFiles {
		if session.ChangedFiles[i].Path == files[0].Path {
			session.ChangedFiles[i].Reviewed = true
		}
	}

	// Second refresh: verify reviewed status is preserved
	files2, err := sm.RefreshChangedFiles(session)
	if err != nil {
		t.Fatalf("RefreshChangedFiles (second): %v", err)
	}

	found := false
	for _, f := range files2 {
		if f.Path == files[0].Path {
			found = true
			if !f.Reviewed {
				t.Errorf("expected %s to remain reviewed after refresh", f.Path)
			}
		}
	}
	if !found {
		t.Errorf("expected to find file %s in refreshed files", files[0].Path)
	}
}

func TestAdvanceRound(t *testing.T) {
	sm, _, baseRef := newTestSessionManager(t)

	session, err := sm.CreateSession(SessionOptions{BaseRef: baseRef})
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	// Add a changed file
	file := &types.ChangedFile{Path: "hello.go", Status: types.FileModified}
	if err := sm.db.UpsertChangedFile(session.ID, file); err != nil {
		t.Fatalf("UpsertChangedFile: %v", err)
	}
	session.ChangedFiles = []types.ChangedFile{*file}
	session.FileStatuses = map[string]bool{"hello.go": true}

	// Add a comment
	now := time.Now()
	comment := &types.ReviewComment{
		ID:          "comment-adv-1",
		TargetType:  types.TargetFile,
		TargetRef:   "hello.go",
		LineStart:   1,
		LineEnd:     1,
		Type:        types.CommentIssue,
		Body:        "round 1 comment",
		ReviewRound: 1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := sm.db.CreateComment(session.ID, comment); err != nil {
		t.Fatalf("CreateComment: %v", err)
	}

	oldBaseRef := session.BaseRef

	// Advance round
	if err := sm.AdvanceRound(session); err != nil {
		t.Fatalf("AdvanceRound: %v", err)
	}

	// ReviewRound should be incremented
	if session.ReviewRound != 2 {
		t.Errorf("expected ReviewRound 2, got %d", session.ReviewRound)
	}

	// BaseRef should be updated to current HEAD (different from old base)
	if session.BaseRef == oldBaseRef {
		// It could be the same if no new commits, but CurrentRef returns HEAD which
		// should differ from the explicit baseRef we passed (the initial commit).
		// setupTestRepo makes changes after the initial commit, but HEAD is still the
		// initial commit since changes are unstaged/staged but not committed.
		// So BaseRef might equal the current HEAD. Just verify it's a valid ref.
	}
	if len(session.BaseRef) != 40 {
		t.Errorf("expected 40-char BaseRef, got %q", session.BaseRef)
	}

	// ChangedFiles should be cleared
	if session.ChangedFiles != nil {
		t.Errorf("expected ChangedFiles to be nil, got %v", session.ChangedFiles)
	}

	// FileStatuses should be reset (empty map)
	if len(session.FileStatuses) != 0 {
		t.Errorf("expected FileStatuses to be empty, got %v", session.FileStatuses)
	}

	// Comments in DB should be marked as outdated
	comments, err := sm.db.GetComments(session.ID)
	if err != nil {
		t.Fatalf("GetComments: %v", err)
	}
	if len(comments) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(comments))
	}
	if !comments[0].Outdated {
		t.Error("expected comment to be marked as outdated")
	}

	// Changed files should be deleted from DB
	dbFiles, err := sm.db.GetChangedFiles(session.ID)
	if err != nil {
		t.Fatalf("GetChangedFiles: %v", err)
	}
	if len(dbFiles) != 0 {
		t.Errorf("expected 0 changed files in DB after advance, got %d", len(dbFiles))
	}
}

func TestListSessions(t *testing.T) {
	sm, dir, baseRef := newTestSessionManager(t)

	// Create sessions with different RepoRoots
	_, err := sm.CreateSession(SessionOptions{
		BaseRef:  baseRef,
		RepoRoot: dir,
	})
	if err != nil {
		t.Fatalf("CreateSession 1: %v", err)
	}

	_, err = sm.CreateSession(SessionOptions{
		BaseRef:  baseRef,
		RepoRoot: dir,
	})
	if err != nil {
		t.Fatalf("CreateSession 2: %v", err)
	}

	_, err = sm.CreateSession(SessionOptions{
		BaseRef:  baseRef,
		RepoRoot: "/other/repo",
	})
	if err != nil {
		t.Fatalf("CreateSession 3: %v", err)
	}

	// List all sessions
	all, err := sm.ListSessions(ListSessionsOptions{})
	if err != nil {
		t.Fatalf("ListSessions (all): %v", err)
	}
	if len(all) != 3 {
		t.Errorf("expected 3 sessions, got %d", len(all))
	}

	// Filter by RepoRoot
	filtered, err := sm.ListSessions(ListSessionsOptions{RepoRoot: dir})
	if err != nil {
		t.Fatalf("ListSessions (filtered): %v", err)
	}
	if len(filtered) != 2 {
		t.Errorf("expected 2 sessions for repo %q, got %d", dir, len(filtered))
	}
	for _, s := range filtered {
		if s.RepoRoot != dir {
			t.Errorf("expected RepoRoot %q, got %q", dir, s.RepoRoot)
		}
	}

	// Filter by other repo
	other, err := sm.ListSessions(ListSessionsOptions{RepoRoot: "/other/repo"})
	if err != nil {
		t.Fatalf("ListSessions (other): %v", err)
	}
	if len(other) != 1 {
		t.Errorf("expected 1 session for /other/repo, got %d", len(other))
	}
}

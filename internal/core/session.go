package core

import (
	"fmt"
	"time"

	"github.com/anthropics/monocle/internal/db"
	"github.com/anthropics/monocle/internal/types"
	"github.com/google/uuid"
)

// SessionManager handles session lifecycle operations.
type SessionManager struct {
	db  *db.DB
	git *GitClient
}

// NewSessionManager creates a new SessionManager.
func NewSessionManager(database *db.DB, git *GitClient) *SessionManager {
	return &SessionManager{db: database, git: git}
}

// CreateSession starts a new review session.
func (sm *SessionManager) CreateSession(opts SessionOptions) (*types.ReviewSession, error) {
	baseRef := opts.BaseRef
	if baseRef == "" {
		ref, err := sm.git.CurrentRef()
		if err != nil {
			return nil, fmt.Errorf("get current ref: %w", err)
		}
		baseRef = ref
	}

	now := time.Now()
	session := &types.ReviewSession{
		ID:             uuid.New().String(),
		Agent:          opts.Agent,
		AgentStatus:    types.AgentStatusIdle,
		RepoRoot:       opts.RepoRoot,
		BaseRef:        baseRef,
		IgnorePatterns: opts.IgnorePatterns,
		ReviewRound:    1,
		FileStatuses:   make(map[string]bool),
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if session.Agent == "" {
		session.Agent = "claude"
	}
	if session.RepoRoot == "" {
		session.RepoRoot = sm.git.RepoRoot()
	}

	if err := sm.db.CreateSession(session); err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	return session, nil
}

// ResumeSession loads an existing session from the database.
func (sm *SessionManager) ResumeSession(sessionID string) (*types.ReviewSession, error) {
	session, err := sm.db.GetSession(sessionID)
	if err != nil {
		return nil, fmt.Errorf("get session %s: %w", sessionID, err)
	}

	// Load related data
	files, err := sm.db.GetChangedFiles(session.ID)
	if err != nil {
		return nil, fmt.Errorf("get changed files: %w", err)
	}
	session.ChangedFiles = files

	items, err := sm.db.GetContentItems(session.ID)
	if err != nil {
		return nil, fmt.Errorf("get content items: %w", err)
	}
	session.ContentItems = items

	comments, err := sm.db.GetComments(session.ID)
	if err != nil {
		return nil, fmt.Errorf("get comments: %w", err)
	}
	session.Comments = comments

	// Build file statuses map
	session.FileStatuses = make(map[string]bool)
	for _, f := range files {
		session.FileStatuses[f.Path] = f.Reviewed
	}

	return session, nil
}

// RefreshChangedFiles re-runs git diff and updates the session's file list.
func (sm *SessionManager) RefreshChangedFiles(session *types.ReviewSession) ([]types.ChangedFile, error) {
	files, err := sm.git.Diff(session.BaseRef)
	if err != nil {
		return nil, fmt.Errorf("git diff: %w", err)
	}

	// Merge with existing review status
	existingStatus := make(map[string]bool)
	for _, f := range session.ChangedFiles {
		existingStatus[f.Path] = f.Reviewed
	}

	for i := range files {
		files[i].Reviewed = existingStatus[files[i].Path]
		if err := sm.db.UpsertChangedFile(session.ID, &files[i]); err != nil {
			return nil, fmt.Errorf("upsert file %s: %w", files[i].Path, err)
		}
	}

	session.ChangedFiles = files
	return files, nil
}

// AdvanceRound increments the review round and marks existing comments as outdated.
func (sm *SessionManager) AdvanceRound(session *types.ReviewSession) error {
	session.ReviewRound++
	session.UpdatedAt = time.Now()

	if err := sm.db.UpdateSession(session); err != nil {
		return fmt.Errorf("update session round: %w", err)
	}

	if err := sm.db.MarkOutdated(session.ID); err != nil {
		return fmt.Errorf("mark outdated: %w", err)
	}

	return nil
}

// ListSessions returns session summaries.
func (sm *SessionManager) ListSessions(opts ListSessionsOptions) ([]types.SessionSummary, error) {
	return sm.db.ListSessions(opts.RepoRoot, opts.Limit)
}

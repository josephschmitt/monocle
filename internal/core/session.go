package core

import (
	"fmt"
	"time"

	"github.com/josephschmitt/monocle/internal/db"
	"github.com/josephschmitt/monocle/internal/types"
	"github.com/google/uuid"
)

// SessionManager handles session lifecycle operations.
type SessionManager struct {
	db  *db.DB
	git GitAPI
}

// NewSessionManager creates a new SessionManager.
func NewSessionManager(database *db.DB, git GitAPI) *SessionManager {
	return &SessionManager{db: database, git: git}
}

// CreateSession starts a new review session with the given options.
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
		RepoRoot:       opts.RepoRoot,
		BaseRef:        baseRef,
		IgnorePatterns: opts.IgnorePatterns,
		ReviewRound:    1,
		FileStatuses:   make(map[string]bool),
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if session.RepoRoot == "" {
		session.RepoRoot = sm.git.RepoRoot()
	}

	if err := sm.db.CreateSession(session); err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	return session, nil
}

// ResumeSession loads an existing session from the database and restores all related state.
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

	additionalFiles, err := sm.db.GetAdditionalFiles(session.ID)
	if err != nil {
		return nil, fmt.Errorf("get additional files: %w", err)
	}
	session.AdditionalFiles = additionalFiles

	// Build file statuses map
	session.FileStatuses = make(map[string]bool)
	for _, f := range files {
		session.FileStatuses[f.Path] = f.Reviewed
	}

	return session, nil
}

// RefreshChangedFiles re-runs git diff and updates the session's file list,
// merging with existing review status. It also returns the reviewed-state map
// captured from the DB *before* the transactional replace deleted rows, keyed
// by path. Callers that re-add files which the replace pruned (e.g. reverted
// snapshot files in filesRelativeToSnapshot) use this map to preserve the prior
// reviewed flag, which would otherwise be lost when the row is re-inserted.
func (sm *SessionManager) RefreshChangedFiles(session *types.ReviewSession) ([]types.ChangedFile, map[string]bool, error) {
	files, err := sm.git.Diff(session.BaseRef)
	if err != nil {
		return nil, nil, fmt.Errorf("git diff: %w", err)
	}

	// Read reviewed state from DB (source of truth) rather than in-memory
	// session.ChangedFiles, which can be stale during concurrent submit. This is
	// captured before ReplaceChangedFiles deletes rows, so it still includes
	// files that are about to be pruned (e.g. reverted snapshot-only files).
	dbFiles, _ := sm.db.GetChangedFiles(session.ID)
	existingStatus := make(map[string]bool, len(dbFiles))
	for _, f := range dbFiles {
		existingStatus[f.Path] = f.Reviewed
	}

	// Merge reviewed state onto the current set, then persist the whole set in a
	// single transactional replace. This prunes rows for files that are no longer
	// in the diff (deleted or newly gitignored untracked files) and avoids a
	// per-file write storm under repos with many changes.
	ptrs := make([]*types.ChangedFile, len(files))
	for i := range files {
		files[i].Reviewed = existingStatus[files[i].Path]
		ptrs[i] = &files[i]
	}
	if err := sm.db.ReplaceChangedFiles(session.ID, ptrs); err != nil {
		return nil, nil, fmt.Errorf("replace changed files: %w", err)
	}

	session.ChangedFiles = files
	return files, existingStatus, nil
}

// AdvanceRound increments the review round. Files, artifacts, and base ref are
// untouched — reviewed state is managed by the submit-time mark/reset logic,
// and the periodic refresh handles downstream updates.
func (sm *SessionManager) AdvanceRound(session *types.ReviewSession) error {
	session.ReviewRound++
	session.UpdatedAt = time.Now()

	if err := sm.db.UpdateSession(session); err != nil {
		return fmt.Errorf("update session round: %w", err)
	}

	return nil
}

// ListSessions returns session summaries.
func (sm *SessionManager) ListSessions(opts ListSessionsOptions) ([]types.SessionSummary, error) {
	return sm.db.ListSessions(opts.RepoRoot, opts.Limit)
}

package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/josephschmitt/monocle/internal/db"
	"github.com/josephschmitt/monocle/internal/protocol"
	"github.com/josephschmitt/monocle/internal/types"
	"github.com/google/uuid"
)

// Engine implements EngineAPI and coordinates all Monocle subsystems.
type Engine struct {
	mu sync.RWMutex

	cfg       *types.Config
	database  *db.DB
	git       GitAPI
	server    *SocketServer
	feedback  *FeedbackQueue
	formatter *ReviewFormatter
	sessions  *SessionManager

	current *types.ReviewSession

	// autoAdvanceRef: when true, baseRef advances to HEAD on each refresh
	autoAdvanceRef bool
	lastKnownHead  string
	selectedRef    string // the commit the user picked (BaseRef stores its parent for diffing)

	// activeSnapshot: when non-nil, diffs are computed against this snapshot's stored state
	activeSnapshot *types.ReviewSnapshot

	// event subscribers: EventKind -> subscriber ID -> callback
	subscribers map[EventKind]map[int]EventCallback
	nextSubID   int
}

// NewEngine constructs an Engine with all subsystems wired together.
// When nonGitMode is true, a DirClient is used instead of GitClient,
// allowing Monocle to browse non-git directories.
func NewEngine(cfg *types.Config, database *db.DB, repoRoot string, nonGitMode bool) (*Engine, error) {
	var git GitAPI
	if nonGitMode {
		git = NewDirClient(repoRoot, cfg.IgnorePatterns)
	} else {
		git = NewGitClient(repoRoot)
	}
	server := NewSocketServer()
	feedback := NewFeedbackQueue()

	e := &Engine{
		cfg:            cfg,
		database:       database,
		git:            git,
		server:         server,
		feedback:       feedback,
		sessions:       NewSessionManager(database, git),
		autoAdvanceRef: !nonGitMode,
		subscribers:    make(map[EventKind]map[int]EventCallback),
	}

	e.formatter = NewReviewFormatter(func(path string, start, end int) string {
		content, err := git.FileContent("", path)
		if err != nil {
			return ""
		}
		return extractLines(content, start, end)
	}, cfg.ReviewFormat)

	e.formatter.SetContentItemProvider(func(id string) string {
		e.mu.RLock()
		session := e.current
		e.mu.RUnlock()
		if session == nil {
			return ""
		}
		item, err := database.GetContentItem(session.ID, id)
		if err != nil || item == nil {
			return ""
		}
		return item.Content
	})

	server.SetEngine(e)

	return e, nil
}

// extractLines returns the requested line range (1-based, inclusive) from content.
func extractLines(content string, start, end int) string {
	if start <= 0 {
		return ""
	}
	var lines []byte
	lineNum := 1
	lineStart := 0
	for i := 0; i <= len(content); i++ {
		if i == len(content) || content[i] == '\n' {
			if lineNum >= start && lineNum <= end {
				line := content[lineStart:i]
				lines = append(lines, []byte(line)...)
				lines = append(lines, '\n')
			}
			if lineNum > end {
				break
			}
			lineNum++
			lineStart = i + 1
		}
	}
	return string(lines)
}

// -- Session lifecycle --

func (e *Engine) StartSession(opts SessionOptions) (*types.ReviewSession, error) {
	session, err := e.sessions.CreateSession(opts)
	if err != nil {
		return nil, err
	}

	if _, err := e.sessions.RefreshChangedFiles(session); err != nil {
		return nil, fmt.Errorf("refresh changed files: %w", err)
	}

	e.mu.Lock()
	e.current = session
	e.mu.Unlock()

	return session, nil
}

func (e *Engine) ResumeSession(sessionID string) (*types.ReviewSession, error) {
	session, err := e.sessions.ResumeSession(sessionID)
	if err != nil {
		return nil, err
	}

	if _, err := e.sessions.RefreshChangedFiles(session); err != nil {
		return nil, fmt.Errorf("refresh changed files: %w", err)
	}

	e.mu.Lock()
	e.current = session
	e.mu.Unlock()

	return session, nil
}

func (e *Engine) GetSession() *types.ReviewSession {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.current
}

func (e *Engine) ListSessions(opts ListSessionsOptions) ([]types.SessionSummary, error) {
	return e.sessions.ListSessions(opts)
}

// -- Browsing --

func (e *Engine) RefreshChangedFiles() ([]types.ChangedFile, error) {
	e.mu.RLock()
	session := e.current
	e.mu.RUnlock()
	if session == nil {
		return nil, fmt.Errorf("no active session")
	}

	// Auto-advance baseRef to HEAD when commits happen
	if e.autoAdvanceRef {
		head, err := e.git.CurrentRef()
		if err == nil && head != e.lastKnownHead {
			e.lastKnownHead = head
			if head != session.BaseRef {
				e.mu.Lock()
				if e.current != nil {
					e.current.BaseRef = head
					e.current.UpdatedAt = time.Now()
					_ = e.database.UpdateSession(e.current)
				}
				e.mu.Unlock()
			}
		}
	}

	files, err := e.sessions.RefreshChangedFiles(session)
	if err != nil {
		return nil, err
	}

	// Auto-unmark reviewed files that changed since the latest snapshot
	snapshots, snapErr := e.database.GetSnapshots(session.ID)
	if snapErr == nil && len(snapshots) > 0 {
		latest, loadErr := e.database.GetSnapshot(snapshots[0].ID) // most recent (sorted desc)
		if loadErr == nil {
			e.autoUnmarkChangedFiles(session, files, latest)
		}
	}

	e.mu.Lock()
	if e.current != nil && e.current.ID == session.ID {
		e.current.ChangedFiles = files
	}
	e.mu.Unlock()

	return files, nil
}

func (e *Engine) GetChangedFiles() []types.ChangedFile {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if e.current == nil {
		return nil
	}
	return e.current.ChangedFiles
}

func (e *Engine) GetContentItems() []types.ContentItem {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if e.current == nil {
		return nil
	}
	return e.current.ContentItems
}

func (e *Engine) GetFileDiff(path string) (*types.DiffResult, error) {
	e.mu.RLock()
	session := e.current
	snapshot := e.activeSnapshot
	e.mu.RUnlock()
	if session == nil {
		return nil, fmt.Errorf("no active session")
	}

	// Snapshot mode: diff stored content against current working tree
	if snapshot != nil {
		return e.snapshotFileDiff(snapshot, path)
	}

	return e.git.FileDiff(session.BaseRef, path, e.cfg.ContextLines)
}

func (e *Engine) GetFileContent(path string) (string, error) {
	return e.git.FileContent("", path)
}

func (e *Engine) GetContentItem(id string) (*types.ContentItem, error) {
	e.mu.RLock()
	session := e.current
	e.mu.RUnlock()
	if session == nil {
		return nil, fmt.Errorf("no active session")
	}
	return e.database.GetContentItem(session.ID, id)
}

// GetContentDiff computes a diff between the previous and current version of a content item.
// Returns nil if no previous version exists.
func (e *Engine) GetContentDiff(id string) (*types.DiffResult, error) {
	e.mu.RLock()
	session := e.current
	e.mu.RUnlock()
	if session == nil {
		return nil, fmt.Errorf("no active session")
	}
	versions, err := e.database.GetContentVersions(session.ID, id)
	if err != nil {
		return nil, err
	}
	if len(versions) < 2 {
		return nil, nil
	}
	prev := versions[len(versions)-2]
	curr := versions[len(versions)-1]
	hunks, err := TextDiff(prev.Content, curr.Content)
	if err != nil {
		return nil, fmt.Errorf("compute content diff: %w", err)
	}
	return &types.DiffResult{Path: id, Hunks: hunks}, nil
}

// GetContentVersions returns all versions of a content item.
func (e *Engine) GetContentVersions(id string) ([]types.ContentVersion, error) {
	e.mu.RLock()
	session := e.current
	e.mu.RUnlock()
	if session == nil {
		return nil, fmt.Errorf("no active session")
	}
	return e.database.GetContentVersions(session.ID, id)
}

// GetContentDiffBetweenVersions computes a diff between two specific versions of a content item.
func (e *Engine) GetContentDiffBetweenVersions(id string, fromVersion, toVersion int) (*types.DiffResult, error) {
	e.mu.RLock()
	session := e.current
	e.mu.RUnlock()
	if session == nil {
		return nil, fmt.Errorf("no active session")
	}
	from, err := e.database.GetContentVersion(session.ID, id, fromVersion)
	if err != nil {
		return nil, fmt.Errorf("get version %d: %w", fromVersion, err)
	}
	to, err := e.database.GetContentVersion(session.ID, id, toVersion)
	if err != nil {
		return nil, fmt.Errorf("get version %d: %w", toVersion, err)
	}
	hunks, err := TextDiff(from.Content, to.Content)
	if err != nil {
		return nil, fmt.Errorf("compute version diff: %w", err)
	}
	return &types.DiffResult{Path: id, Hunks: hunks}, nil
}

// -- Additional files --

func (e *Engine) GetAdditionalFiles() []types.AdditionalFile {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if e.current == nil {
		return nil
	}
	return e.current.AdditionalFiles
}

func (e *Engine) AddAdditionalPaths(paths []string) ([]types.AdditionalFile, error) {
	e.mu.Lock()
	if e.current == nil {
		e.mu.Unlock()
		return nil, fmt.Errorf("no active session")
	}
	session := e.current

	// Build set of existing paths for dedup
	existing := make(map[string]bool, len(session.AdditionalFiles))
	for _, af := range session.AdditionalFiles {
		existing[af.Path] = true
	}

	var added []types.AdditionalFile
	for _, p := range paths {
		absPath, err := filepath.Abs(p)
		if err != nil {
			continue
		}

		info, err := os.Stat(absPath)
		if err != nil {
			continue
		}

		if info.IsDir() {
			_ = filepath.WalkDir(absPath, func(path string, d os.DirEntry, err error) error {
				if err != nil {
					return nil
				}
				name := d.Name()
				if d.IsDir() {
					// Skip hidden and noisy directories
					if strings.HasPrefix(name, ".") || name == "node_modules" {
						return filepath.SkipDir
					}
					return nil
				}
				// Skip hidden files and .DS_Store
				if strings.HasPrefix(name, ".") {
					return nil
				}
				if !d.Type().IsRegular() {
					return nil
				}
				if existing[path] {
					return nil
				}
				relName, _ := filepath.Rel(absPath, path)
				if relName == "" {
					relName = filepath.Base(path)
				}
				af := types.AdditionalFile{
					Path: path,
					Name: relName,
				}
				session.AdditionalFiles = append(session.AdditionalFiles, af)
				existing[path] = true
				added = append(added, af)
				return nil
			})
		} else {
			if existing[absPath] {
				continue
			}
			af := types.AdditionalFile{
				Path: absPath,
				Name: filepath.Base(absPath),
			}
			session.AdditionalFiles = append(session.AdditionalFiles, af)
			existing[absPath] = true
			added = append(added, af)
		}
	}
	// Persist to DB
	for i := range added {
		if err := e.database.UpsertAdditionalFile(session.ID, &added[i]); err != nil {
			e.mu.Unlock()
			return nil, fmt.Errorf("persist additional file %s: %w", added[i].Path, err)
		}
	}
	e.mu.Unlock()

	for _, af := range added {
		e.emit(EventAdditionalFileAdded, EventPayload{
			Kind: EventAdditionalFileAdded,
			Path: af.Path,
		})
	}

	return added, nil
}

func (e *Engine) GetAdditionalFileContent(absPath string) (string, error) {
	e.mu.RLock()
	found := false
	if e.current != nil {
		for _, af := range e.current.AdditionalFiles {
			if af.Path == absPath {
				found = true
				break
			}
		}
	}
	e.mu.RUnlock()

	if !found {
		return "", fmt.Errorf("path not in additional files: %s", absPath)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return "", fmt.Errorf("read additional file: %w", err)
	}
	return string(data), nil
}

func (e *Engine) handleAddAdditionalFiles(msg *protocol.AddAdditionalFilesMsg) *protocol.AddAdditionalFilesResponse {
	added, err := e.AddAdditionalPaths(msg.Paths)
	if err != nil {
		return &protocol.AddAdditionalFilesResponse{
			Type:    protocol.TypeAddAdditionalFilesResponse,
			Success: false,
			Message: err.Error(),
		}
	}

	return &protocol.AddAdditionalFilesResponse{
		Type:    protocol.TypeAddAdditionalFilesResponse,
		Success: true,
		Message: fmt.Sprintf("Added %d file(s) for review", len(added)),
		Count:   len(added),
	}
}

// -- Commenting --

func (e *Engine) AddComment(target CommentTarget, commentType types.CommentType, body string) (*types.ReviewComment, error) {
	e.mu.RLock()
	session := e.current
	e.mu.RUnlock()
	if session == nil {
		return nil, fmt.Errorf("no active session")
	}

	now := time.Now()
	comment := &types.ReviewComment{
		ID:          uuid.New().String(),
		TargetType:  target.TargetType,
		TargetRef:   target.TargetRef,
		LineStart:   target.LineStart,
		LineEnd:     target.LineEnd,
		Type:        commentType,
		Body:        body,
		ReviewRound: session.ReviewRound,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := e.database.CreateComment(session.ID, comment); err != nil {
		return nil, fmt.Errorf("create comment: %w", err)
	}

	e.mu.Lock()
	session.Comments = append(session.Comments, *comment)
	e.mu.Unlock()

	return comment, nil
}

func (e *Engine) EditComment(commentID string, commentType types.CommentType, body string) (*types.ReviewComment, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.current == nil {
		return nil, fmt.Errorf("no active session")
	}

	var found *types.ReviewComment
	for i := range e.current.Comments {
		if e.current.Comments[i].ID == commentID {
			found = &e.current.Comments[i]
			break
		}
	}
	if found == nil {
		return nil, fmt.Errorf("comment %s not found", commentID)
	}

	found.Type = commentType
	found.Body = body
	found.UpdatedAt = time.Now()

	if err := e.database.UpdateComment(found); err != nil {
		return nil, fmt.Errorf("update comment: %w", err)
	}

	result := *found
	return &result, nil
}

func (e *Engine) DeleteComment(commentID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.current == nil {
		return fmt.Errorf("no active session")
	}

	if err := e.database.DeleteComment(commentID); err != nil {
		return fmt.Errorf("delete comment: %w", err)
	}

	comments := e.current.Comments[:0]
	for _, c := range e.current.Comments {
		if c.ID != commentID {
			comments = append(comments, c)
		}
	}
	e.current.Comments = comments

	return nil
}

func (e *Engine) ResolveComment(commentID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.current == nil {
		return fmt.Errorf("no active session")
	}

	var found *types.ReviewComment
	for i := range e.current.Comments {
		if e.current.Comments[i].ID == commentID {
			found = &e.current.Comments[i]
			break
		}
	}
	if found == nil {
		return fmt.Errorf("comment %s not found", commentID)
	}

	found.Resolved = !found.Resolved
	found.UpdatedAt = time.Now()

	if err := e.database.ResolveComment(commentID, found.Resolved); err != nil {
		return fmt.Errorf("resolve comment: %w", err)
	}

	return nil
}

func (e *Engine) ClearComments() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.current == nil {
		return fmt.Errorf("no active session")
	}

	if err := e.database.ClearComments(e.current.ID); err != nil {
		return fmt.Errorf("clear comments: %w", err)
	}

	e.current.Comments = nil

	return nil
}

// ClearReview resets the current review to a blank slate: clears all comments,
// content items/plans, and reviewed states. Does not advance the round or
// create a submission record.
func (e *Engine) ClearReview() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.current == nil {
		return fmt.Errorf("no active session")
	}

	sessionID := e.current.ID

	if err := e.database.ClearComments(sessionID); err != nil {
		return fmt.Errorf("clear comments: %w", err)
	}
	e.current.Comments = nil

	if err := e.database.DeleteContentItems(sessionID); err != nil {
		return fmt.Errorf("clear content items: %w", err)
	}
	e.current.ContentItems = nil

	if err := e.database.ResetAllReviewed(sessionID); err != nil {
		return fmt.Errorf("reset reviewed: %w", err)
	}
	for i := range e.current.ChangedFiles {
		e.current.ChangedFiles[i].Reviewed = false
	}
	for i := range e.current.AdditionalFiles {
		e.current.AdditionalFiles[i].Reviewed = false
	}
	for k := range e.current.FileStatuses {
		e.current.FileStatuses[k] = false
	}

	return nil
}

// -- Review status --

func (e *Engine) MarkReviewed(path string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.current == nil {
		return fmt.Errorf("no active session")
	}

	// Check additional files first
	for i := range e.current.AdditionalFiles {
		if e.current.AdditionalFiles[i].Path == path {
			e.current.AdditionalFiles[i].Reviewed = true
			return e.database.MarkAdditionalFileReviewed(e.current.ID, path, true)
		}
	}

	if err := e.database.MarkFileReviewed(e.current.ID, path, true); err != nil {
		return fmt.Errorf("mark reviewed: %w", err)
	}

	e.current.FileStatuses[path] = true
	for i := range e.current.ChangedFiles {
		if e.current.ChangedFiles[i].Path == path {
			e.current.ChangedFiles[i].Reviewed = true
			break
		}
	}

	return nil
}

func (e *Engine) UnmarkReviewed(path string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.current == nil {
		return fmt.Errorf("no active session")
	}

	// Check additional files first
	for i := range e.current.AdditionalFiles {
		if e.current.AdditionalFiles[i].Path == path {
			e.current.AdditionalFiles[i].Reviewed = false
			return e.database.MarkAdditionalFileReviewed(e.current.ID, path, false)
		}
	}

	if err := e.database.MarkFileReviewed(e.current.ID, path, false); err != nil {
		return fmt.Errorf("unmark reviewed: %w", err)
	}

	e.current.FileStatuses[path] = false
	for i := range e.current.ChangedFiles {
		if e.current.ChangedFiles[i].Path == path {
			e.current.ChangedFiles[i].Reviewed = false
			break
		}
	}

	return nil
}

func (e *Engine) MarkContentReviewed(id string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.current == nil {
		return fmt.Errorf("no active session")
	}

	if err := e.database.MarkContentItemReviewed(e.current.ID, id, true); err != nil {
		return fmt.Errorf("mark content reviewed: %w", err)
	}

	for i := range e.current.ContentItems {
		if e.current.ContentItems[i].ID == id {
			e.current.ContentItems[i].Reviewed = true
			break
		}
	}

	return nil
}

func (e *Engine) UnmarkContentReviewed(id string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.current == nil {
		return fmt.Errorf("no active session")
	}

	if err := e.database.MarkContentItemReviewed(e.current.ID, id, false); err != nil {
		return fmt.Errorf("unmark content reviewed: %w", err)
	}

	for i := range e.current.ContentItems {
		if e.current.ContentItems[i].ID == id {
			e.current.ContentItems[i].Reviewed = false
			break
		}
	}

	return nil
}

func (e *Engine) ResetAllReviewed() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.current == nil {
		return fmt.Errorf("no active session")
	}

	if err := e.database.ResetAllReviewed(e.current.ID); err != nil {
		return fmt.Errorf("reset all reviewed: %w", err)
	}

	for i := range e.current.ChangedFiles {
		e.current.ChangedFiles[i].Reviewed = false
	}
	for i := range e.current.AdditionalFiles {
		e.current.AdditionalFiles[i].Reviewed = false
	}
	for i := range e.current.ContentItems {
		e.current.ContentItems[i].Reviewed = false
	}
	for k := range e.current.FileStatuses {
		e.current.FileStatuses[k] = false
	}

	return nil
}

func (e *Engine) MarkAllReviewed() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.current == nil {
		return fmt.Errorf("no active session")
	}

	if err := e.database.MarkAllReviewed(e.current.ID); err != nil {
		return fmt.Errorf("mark all reviewed: %w", err)
	}

	for i := range e.current.ChangedFiles {
		e.current.ChangedFiles[i].Reviewed = true
	}
	for i := range e.current.AdditionalFiles {
		e.current.AdditionalFiles[i].Reviewed = true
	}
	for i := range e.current.ContentItems {
		e.current.ContentItems[i].Reviewed = true
	}
	for k := range e.current.FileStatuses {
		e.current.FileStatuses[k] = true
	}

	return nil
}

// -- Submission --

func (e *Engine) GetReviewSummary() (*types.ReviewSummary, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.current == nil {
		return nil, fmt.Errorf("no active session")
	}

	summary := &types.ReviewSummary{
		Session:                e.current,
		FileComments:           make(map[string][]types.ReviewComment),
		ContentComments:        make(map[string][]types.ReviewComment),
		AdditionalFileComments: make(map[string][]types.ReviewComment),
	}

	for _, c := range e.current.Comments {
		if c.Resolved {
			continue
		}
		switch c.TargetType {
		case types.TargetFile:
			summary.FileComments[c.TargetRef] = append(summary.FileComments[c.TargetRef], c)
		case types.TargetContent:
			summary.ContentComments[c.TargetRef] = append(summary.ContentComments[c.TargetRef], c)
		case types.TargetAdditionalFile:
			summary.AdditionalFileComments[c.TargetRef] = append(summary.AdditionalFileComments[c.TargetRef], c)
		}
		switch c.Type {
		case types.CommentIssue:
			summary.IssueCt++
		case types.CommentSuggestion:
			summary.SuggestionCt++
		case types.CommentNote:
			summary.NoteCt++
		case types.CommentPraise:
			summary.PraiseCt++
		}
	}

	return summary, nil
}

func (e *Engine) Submit(action types.SubmitAction, body string) (*SubmitResult, error) {
	e.mu.RLock()
	session := e.current
	e.mu.RUnlock()

	if session == nil {
		return nil, fmt.Errorf("no active session")
	}

	// Validate: request_changes must include at least one unresolved comment or body text
	if action == types.ActionRequestChanges {
		hasBody := strings.TrimSpace(body) != ""
		hasUnresolved := false
		for _, c := range session.Comments {
			if !c.Resolved {
				hasUnresolved = true
				break
			}
		}
		if !hasBody && !hasUnresolved {
			return nil, fmt.Errorf("request_changes requires at least one comment or review body")
		}
	}

	formatted := e.formatter.Format(session, session.Comments, action, body)

	// Check if a push-mode agent is connected (subscriber = channel mode)
	agentConnected := e.server != nil && e.server.SubscriberCount() > 0

	e.feedback.Submit(formatted, agentConnected)

	// Save submission record
	now := time.Now()
	sub := &types.ReviewSubmission{
		ID:              uuid.New().String(),
		SessionID:       session.ID,
		Action:          types.SubmitAction(formatted.Action),
		FormattedReview: formatted.Formatted,
		CommentCount:    formatted.CommentCount,
		ReviewRound:     session.ReviewRound,
		SubmittedAt:     now,
	}
	if agentConnected {
		sub.DeliveredAt = &now // Channel delivers immediately
	}
	_ = e.database.CreateSubmission(session.ID, sub)

	// Snapshot lifecycle: request_changes creates a snapshot, approve wipes all
	if action == types.ActionRequestChanges {
		e.markReviewedOnSubmit(session)
		_ = e.createSnapshot(session, sub.ID)
	} else {
		// Approve: wipe all snapshots (review is done)
		e.deleteSnapshots()
	}

	// Reset all reviewed states after submitting
	_ = e.ResetAllReviewed()

	e.emit(EventFeedbackStatusChanged, EventPayload{
		Kind:   EventFeedbackStatusChanged,
		Status: e.feedback.GetStatus(),
	})

	e.emit(EventFeedbackSubmitted, EventPayload{
		Kind:    EventFeedbackSubmitted,
		Message: buildFeedbackSummary(formatted.Action, session.Comments),
		Status:  formatted.Action,
	})

	if agentConnected {
		// Push mode: advance round for a clean slate (channel delivers immediately)
		e.mu.Lock()
		_ = e.sessions.AdvanceRound(session)
		e.mu.Unlock()

		e.feedback.ClearStatus()

		e.emit(EventFeedbackStatusChanged, EventPayload{
			Kind:   EventFeedbackStatusChanged,
			Status: "none",
		})
		e.emit(EventFileChanged, EventPayload{
			Kind: EventFileChanged,
		})
	}

	return &SubmitResult{AgentConnected: agentConnected}, nil
}

// buildFeedbackSummary creates a human-readable one-liner for channel notifications.
func buildFeedbackSummary(action string, comments []types.ReviewComment) string {
	issues, suggestions, notes, _ := countByType(comments)

	// Build counts portion (skip praise — not actionable)
	var parts []string
	if issues > 0 {
		if issues == 1 {
			parts = append(parts, "1 issue")
		} else {
			parts = append(parts, fmt.Sprintf("%d issues", issues))
		}
	}
	if suggestions > 0 {
		if suggestions == 1 {
			parts = append(parts, "1 suggestion")
		} else {
			parts = append(parts, fmt.Sprintf("%d suggestions", suggestions))
		}
	}
	if notes > 0 {
		if notes == 1 {
			parts = append(parts, "1 note")
		} else {
			parts = append(parts, fmt.Sprintf("%d notes", notes))
		}
	}
	counts := strings.Join(parts, ", ")

	if action == string(types.ActionRequestChanges) {
		if counts != "" {
			return fmt.Sprintf("Your reviewer requested changes (%s). Call get_feedback to retrieve the full review and address their comments.", counts)
		}
		return "Your reviewer requested changes. Call get_feedback to retrieve the full review and address their comments."
	}

	// Approved
	if counts != "" {
		return fmt.Sprintf("Your reviewer approved your changes with %s. Call get_feedback to read the review.", counts)
	}
	return "Your reviewer approved your changes. Call get_feedback to read the review."
}

func (e *Engine) FormatReview(action types.SubmitAction, body string) (string, error) {
	e.mu.RLock()
	session := e.current
	e.mu.RUnlock()

	if session == nil {
		return "", fmt.Errorf("no active session")
	}

	formatted := e.formatter.Format(session, session.Comments, action, body)
	return formatted.Formatted, nil
}

func (e *Engine) GetSubmissions() ([]types.ReviewSubmission, error) {
	e.mu.RLock()
	session := e.current
	e.mu.RUnlock()
	if session == nil {
		return nil, fmt.Errorf("no active session")
	}
	return e.database.GetSubmissions(session.ID)
}

// -- Base ref management --

// SetBaseRef manually sets the diff baseline and disables auto-advance.
// The base is set to the parent of the given ref so the diff includes that
// commit's changes (the user selects a commit to review, not to exclude).
func (e *Engine) SetBaseRef(ref string) error {
	// Resolve to the parent so the selected commit's changes are included
	resolved, err := e.git.ResolveRef(ref + "~1")
	if err != nil {
		// Fall back to the ref itself if it has no parent (root commit)
		resolved, err = e.git.ResolveRef(ref)
		if err != nil {
			return fmt.Errorf("resolve ref %q: %w", ref, err)
		}
	}

	// Remember the user's actual selection for display purposes
	selected, err := e.git.ResolveRef(ref)
	if err != nil {
		selected = resolved
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if e.current == nil {
		return fmt.Errorf("no active session")
	}

	e.current.BaseRef = resolved
	e.current.UpdatedAt = time.Now()
	e.autoAdvanceRef = false
	e.selectedRef = selected
	_ = e.database.UpdateSession(e.current)

	// Wipe snapshots — changing base ref invalidates old review state
	e.deleteSnapshots()

	return nil
}

// SetAutoAdvanceRef enables or disables auto-advancing baseRef to HEAD on each refresh.
func (e *Engine) SetAutoAdvanceRef(enabled bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.autoAdvanceRef = enabled
	if enabled {
		e.lastKnownHead = "" // Force HEAD re-detection on next refresh
		e.selectedRef = ""
		// Wipe snapshots — switching to auto mode invalidates review state
		e.deleteSnapshots()
	}
}

// SelectedBaseRef returns the commit the user selected in the ref picker.
// This differs from session.BaseRef which stores the parent for diffing.
// Returns empty string when in auto-advance mode.
func (e *Engine) SelectedBaseRef() string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.selectedRef
}

// IsAutoAdvanceRef returns whether auto-advance is enabled.
func (e *Engine) IsAutoAdvanceRef() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.autoAdvanceRef
}

// RecentCommits returns recent commits for the ref picker.
func (e *Engine) RecentCommits(n int) ([]LogEntry, error) {
	return e.git.RecentCommits(n)
}

// -- Review snapshots --

// GetSnapshots returns all review snapshots for the current session.
func (e *Engine) GetSnapshots() ([]types.ReviewSnapshot, error) {
	e.mu.RLock()
	session := e.current
	e.mu.RUnlock()
	if session == nil {
		return nil, fmt.Errorf("no active session")
	}
	return e.database.GetSnapshots(session.ID)
}

// SetSnapshotBase activates snapshot-based diffing against the given snapshot.
func (e *Engine) SetSnapshotBase(snapshotID int) error {
	snap, err := e.database.GetSnapshot(snapshotID)
	if err != nil {
		return fmt.Errorf("get snapshot: %w", err)
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	e.activeSnapshot = snap
	return nil
}

// ClearSnapshotBase deactivates snapshot-based diffing.
func (e *Engine) ClearSnapshotBase() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.activeSnapshot = nil
}

// GetActiveSnapshot returns the currently active snapshot, or nil if not in snapshot mode.
func (e *Engine) GetActiveSnapshot() *types.ReviewSnapshot {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.activeSnapshot
}

// HasSnapshots returns true if any snapshots exist for the current session.
func (e *Engine) HasSnapshots() (bool, error) {
	e.mu.RLock()
	session := e.current
	e.mu.RUnlock()
	if session == nil {
		return false, fmt.Errorf("no active session")
	}
	return e.database.HasSnapshots(session.ID)
}

// createSnapshot captures the current state of all changed files into a snapshot.
func (e *Engine) createSnapshot(session *types.ReviewSession, submissionID string) error {
	if e.git == nil {
		return nil
	}

	headRef, err := e.git.CurrentRef()
	if err != nil {
		headRef = "unknown"
	}

	var snapshotFiles []types.SnapshotFile
	for _, f := range session.ChangedFiles {
		sf := types.SnapshotFile{
			Path:     f.Path,
			Status:   f.Status,
			Reviewed: f.Reviewed,
		}

		// Hash the working tree content into git's object store
		if f.Status != types.FileDeleted {
			sha, err := e.git.HashObject(f.Path)
			if err != nil {
				// Fallback: try reading content directly (non-git mode)
				content, readErr := e.git.FileContent("", f.Path)
				if readErr == nil {
					sf.Content = content
				}
			} else {
				sf.BlobSHA = sha
			}
		}

		snapshotFiles = append(snapshotFiles, sf)
	}

	_, err = e.database.CreateSnapshot(
		session.ID, submissionID, session.ReviewRound,
		headRef, session.BaseRef, snapshotFiles,
	)
	return err
}

// deleteSnapshots removes all snapshots for the current session and clears active snapshot.
func (e *Engine) deleteSnapshots() {
	if e.current != nil {
		_ = e.database.DeleteSnapshots(e.current.ID)
	}
	e.activeSnapshot = nil
}

// markReviewedOnSubmit marks files as reviewed based on the config setting.
func (e *Engine) markReviewedOnSubmit(session *types.ReviewSession) {
	var mode string
	if e.cfg != nil {
		mode = e.cfg.MarkReviewedOnSubmit
	}
	if mode == "" {
		mode = "all"
	}

	switch mode {
	case "all":
		_ = e.MarkAllReviewed()
	case "commented":
		// Mark only files that have unresolved comments
		commentedPaths := make(map[string]bool)
		for _, c := range session.Comments {
			if !c.Resolved && c.TargetType == types.TargetFile {
				commentedPaths[c.TargetRef] = true
			}
		}
		for _, f := range session.ChangedFiles {
			if commentedPaths[f.Path] {
				_ = e.MarkReviewed(f.Path)
			}
		}
	case "manual":
		// Do nothing — respect only explicit r toggles
	}
}

// autoUnmarkChangedFiles compares current files against the latest snapshot and
// marks files as unreviewed if their content changed since the snapshot.
func (e *Engine) autoUnmarkChangedFiles(session *types.ReviewSession, files []types.ChangedFile, snapshot *types.ReviewSnapshot) {
	// Build lookup from snapshot files
	snapshotBySHA := make(map[string]string) // path -> blob_sha
	for _, sf := range snapshot.Files {
		if sf.BlobSHA != "" {
			snapshotBySHA[sf.Path] = sf.BlobSHA
		}
	}

	for i := range files {
		f := &files[i]
		snapshotSHA, inSnapshot := snapshotBySHA[f.Path]

		if !inSnapshot {
			// New file since snapshot — should be unreviewed
			if f.Reviewed {
				f.Reviewed = false
				session.FileStatuses[f.Path] = false
				_ = e.database.MarkFileReviewed(session.ID, f.Path, false)
			}
			continue
		}

		if f.Status == types.FileDeleted {
			// File deleted — mark unreviewed
			if f.Reviewed {
				f.Reviewed = false
				session.FileStatuses[f.Path] = false
				_ = e.database.MarkFileReviewed(session.ID, f.Path, false)
			}
			continue
		}

		// Compare current content hash against snapshot
		currentSHA, err := e.git.HashObject(f.Path)
		if err != nil {
			continue // can't hash, skip
		}

		if currentSHA != snapshotSHA {
			// Content changed since snapshot — unmark reviewed
			if f.Reviewed {
				f.Reviewed = false
				session.FileStatuses[f.Path] = false
				_ = e.database.MarkFileReviewed(session.ID, f.Path, false)
			}
		} else if !f.Reviewed {
			// Content unchanged since snapshot — mark as reviewed
			f.Reviewed = true
			session.FileStatuses[f.Path] = true
			_ = e.database.MarkFileReviewed(session.ID, f.Path, true)
		}
	}
}

// snapshotFileDiff computes a diff between the snapshot's stored content and the current working tree.
func (e *Engine) snapshotFileDiff(snapshot *types.ReviewSnapshot, path string) (*types.DiffResult, error) {
	// Find the file in the snapshot
	var snapshotFile *types.SnapshotFile
	for i := range snapshot.Files {
		if snapshot.Files[i].Path == path {
			snapshotFile = &snapshot.Files[i]
			break
		}
	}

	// Get current working tree content
	currentContent, currentErr := e.git.FileContent("", path)

	if snapshotFile == nil {
		// File is new since snapshot — show as all-added
		if currentErr != nil {
			return &types.DiffResult{Path: path}, nil
		}
		hunks := buildSyntheticDiff(currentContent)
		return &types.DiffResult{Path: path, Hunks: hunks}, nil
	}

	// Get snapshot content
	var oldContent string
	if snapshotFile.BlobSHA != "" {
		var err error
		oldContent, err = e.git.CatFile(snapshotFile.BlobSHA)
		if err != nil {
			return nil, fmt.Errorf("retrieve snapshot content for %s: %w", path, err)
		}
	} else {
		oldContent = snapshotFile.Content
	}

	if currentErr != nil {
		// File was deleted since snapshot — show as all-removed
		hunks := buildSyntheticDeleteDiff(oldContent)
		return &types.DiffResult{Path: path, Hunks: hunks}, nil
	}

	// Both exist — compute text diff
	if oldContent == currentContent {
		return &types.DiffResult{Path: path}, nil // no changes
	}

	hunks, err := TextDiff(oldContent, currentContent)
	if err != nil {
		return nil, fmt.Errorf("compute snapshot diff for %s: %w", path, err)
	}
	return &types.DiffResult{Path: path, Hunks: hunks}, nil
}

// -- Server --

// StartServer starts the Unix domain socket server at the given path.
func (e *Engine) StartServer(socketPath string) error {
	return e.server.Start(socketPath)
}

// -- Feedback (MCP channel) --

// PollFeedback returns pending feedback without blocking.
func (e *Engine) PollFeedback() *FormattedReview {
	return e.feedback.Poll()
}

// WaitForFeedback blocks until the user submits feedback (pause flow).
func (e *Engine) WaitForFeedback() *FormattedReview {
	return e.feedback.WaitForFeedback()
}

// GetReviewStatusInfo returns the current review status for CLI queries.
func (e *Engine) GetReviewStatusInfo() *ReviewStatusInfo {
	if e.feedback.IsPauseRequested() {
		return &ReviewStatusInfo{
			Status:  "pause_requested",
			Summary: "Your reviewer has requested a pause. Use the get_feedback tool with wait=true to receive feedback.",
		}
	}

	if e.feedback.HasPending() {
		e.mu.RLock()
		commentCount := 0
		if e.current != nil {
			commentCount = len(e.current.Comments)
		}
		e.mu.RUnlock()

		return &ReviewStatusInfo{
			Status:       "pending",
			CommentCount: commentCount,
			Summary:      fmt.Sprintf("%d comment(s) pending review.", commentCount),
		}
	}

	return &ReviewStatusInfo{
		Status:  "no_feedback",
		Summary: "No feedback pending.",
	}
}

// SubmitContentForReview adds or updates a content item (plan, doc) for review.
func (e *Engine) SubmitContentForReview(id, title, content, contentType string, isPlan bool) error {
	e.mu.Lock()
	if e.current == nil {
		e.mu.Unlock()
		return fmt.Errorf("no active session")
	}
	session := e.current

	now := time.Now()
	item := types.ContentItem{
		ID:          id,
		Title:       title,
		Content:     content,
		ContentType: contentType,
		IsPlan:      isPlan,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Upsert into session's content items
	found := false
	for i := range session.ContentItems {
		if session.ContentItems[i].ID == id {
			session.ContentItems[i].Title = title
			session.ContentItems[i].Content = content
			session.ContentItems[i].ContentType = contentType
			session.ContentItems[i].IsPlan = isPlan
			session.ContentItems[i].UpdatedAt = now
			item = session.ContentItems[i]
			found = true
			break
		}
	}
	if !found {
		session.ContentItems = append(session.ContentItems, item)
	}
	e.mu.Unlock()

	// Persist to DB (also creates a new version row and updates item.VersionCount)
	_ = e.database.UpsertContentItem(session.ID, &item)

	// Update in-memory version count
	e.mu.Lock()
	for i := range session.ContentItems {
		if session.ContentItems[i].ID == id {
			session.ContentItems[i].VersionCount = item.VersionCount
			break
		}
	}
	e.mu.Unlock()

	e.emit(EventContentItemAdded, EventPayload{
		Kind:   EventContentItemAdded,
		ItemID: id,
	})

	return nil
}

// RequestPause sets the pause flag so the agent sees "pause_requested" on next status check.
func (e *Engine) RequestPause() {
	e.feedback.SetPauseRequested(true)

	e.emit(EventPauseChanged, EventPayload{
		Kind:   EventPauseChanged,
		Status: "pause_requested",
	})
}

// CancelPause clears the pause flag.
func (e *Engine) CancelPause() {
	e.feedback.SetPauseRequested(false)

	e.emit(EventPauseChanged, EventPayload{
		Kind:   EventPauseChanged,
		Status: "cancelled",
	})
}

func (e *Engine) GetFeedbackStatus() string {
	return e.feedback.GetStatus()
}

func (e *Engine) GetQueuedCount() int {
	return e.feedback.QueuedCount()
}

// ReloadPendingFeedback checks the DB for undelivered submissions from the
// current session and reloads them into the in-memory FeedbackQueue.
// Called on session resume so queued feedback survives restarts.
func (e *Engine) ReloadPendingFeedback() {
	e.mu.RLock()
	session := e.current
	e.mu.RUnlock()
	if session == nil {
		return
	}

	subs, err := e.database.GetUndeliveredSubmissions(session.ID)
	if err != nil || len(subs) == 0 {
		return
	}

	for _, sub := range subs {
		review := &FormattedReview{
			Formatted:    sub.FormattedReview,
			CommentCount: sub.CommentCount,
			Action:       string(sub.Action),
		}
		e.feedback.Submit(review, false)
	}
}

func (e *Engine) GetSubscriberCount() int {
	return e.server.SubscriberCount()
}

func (e *Engine) GetSocketPath() string {
	return e.server.SocketPath()
}

// -- Events --

func (e *Engine) On(event EventKind, callback EventCallback) UnsubscribeFunc {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.subscribers[event] == nil {
		e.subscribers[event] = make(map[int]EventCallback)
	}
	id := e.nextSubID
	e.nextSubID++
	e.subscribers[event][id] = callback

	return func() {
		e.mu.Lock()
		defer e.mu.Unlock()
		delete(e.subscribers[event], id)
	}
}

// emit notifies all subscribers for the given event. Must not be called with e.mu held.
func (e *Engine) emit(event EventKind, payload EventPayload) {
	e.mu.RLock()
	subs := make([]EventCallback, 0, len(e.subscribers[event]))
	for _, cb := range e.subscribers[event] {
		subs = append(subs, cb)
	}
	e.mu.RUnlock()

	for _, cb := range subs {
		cb(payload)
	}
}

// -- Lifecycle --

// Shutdown stops the socket server and cleans up resources.
// GetConfig returns the current configuration.
func (e *Engine) GetConfig() *types.Config {
	return e.cfg
}

// SaveConfig persists the current configuration to disk.
func (e *Engine) SaveConfig() error {
	return SaveConfig(e.cfg)
}

func (e *Engine) Shutdown() {
	_ = e.server.Shutdown()
}

// -- Socket message handlers (called by SocketServer) --

// handleGetReviewStatus returns the current review state.
func (e *Engine) handleGetReviewStatus(_ *protocol.GetReviewStatusMsg) *protocol.GetReviewStatusResponse {
	info := e.GetReviewStatusInfo()
	return &protocol.GetReviewStatusResponse{
		Type:         protocol.TypeGetReviewStatusResponse,
		Status:       info.Status,
		CommentCount: info.CommentCount,
		Summary:      info.Summary,
	}
}

// handlePollFeedback returns pending feedback, optionally blocking until available.
// In push (channel) mode, round advancement happens in Submit().
// In queue mode, round advancement happens here when feedback is picked up.
func (e *Engine) handlePollFeedback(msg *protocol.PollFeedbackMsg) *protocol.PollFeedbackResponse {
	var result *PollResult

	if msg.Wait {
		e.emit(EventWaitStatusChanged, EventPayload{
			Kind:   EventWaitStatusChanged,
			Status: "waiting",
		})
		result = e.feedback.WaitForFeedbackWithInfo()
		e.emit(EventWaitStatusChanged, EventPayload{
			Kind:   EventWaitStatusChanged,
			Status: "",
		})
	} else {
		result = e.feedback.PollWithInfo()
	}

	if result == nil || len(result.Reviews) == 0 {
		return &protocol.PollFeedbackResponse{
			Type:        protocol.TypePollFeedbackResponse,
			HasFeedback: false,
		}
	}

	// If this feedback was NOT already channel-delivered, perform queue delivery
	// side effects: advance round, mark delivered, clear comments, emit events.
	if !result.ChannelDelivered {
		e.completeQueuedDelivery()
	}

	feedback, commentCount, action := result.CombinedFeedback()

	return &protocol.PollFeedbackResponse{
		Type:         protocol.TypePollFeedbackResponse,
		HasFeedback:  true,
		Feedback:     feedback,
		CommentCount: commentCount,
		Action:       action,
	}
}

// completeQueuedDelivery performs the side effects of delivering queued feedback:
// advancing the round, marking DB submissions as delivered, clearing comments,
// and emitting events so the TUI can update.
func (e *Engine) completeQueuedDelivery() {
	e.mu.Lock()
	session := e.current
	if session != nil {
		_ = e.sessions.AdvanceRound(session)
	}
	e.mu.Unlock()

	if session != nil {
		_ = e.database.MarkSubmissionsDelivered(session.ID)
	}

	_ = e.ClearComments()

	e.feedback.ClearStatus()

	e.emit(EventFeedbackPickedUp, EventPayload{
		Kind: EventFeedbackPickedUp,
	})
	e.emit(EventFeedbackStatusChanged, EventPayload{
		Kind:   EventFeedbackStatusChanged,
		Status: "none",
	})
	e.emit(EventFileChanged, EventPayload{
		Kind: EventFileChanged,
	})
}

// handleSubmitContent receives reviewable content (plans, docs) from the agent.
func (e *Engine) handleSubmitContent(msg *protocol.SubmitContentMsg) *protocol.SubmitContentResponse {
	id := msg.ID
	if id == "" {
		id = uuid.New().String()
	}

	err := e.SubmitContentForReview(id, msg.Title, msg.Content, msg.ContentType, msg.IsPlan)
	if err != nil {
		return &protocol.SubmitContentResponse{
			Type:    protocol.TypeSubmitContentResponse,
			Success: false,
			Message: err.Error(),
		}
	}

	return &protocol.SubmitContentResponse{
		Type:    protocol.TypeSubmitContentResponse,
		Success: true,
		Message: fmt.Sprintf("Content submitted for review: %s", msg.Title),
	}
}

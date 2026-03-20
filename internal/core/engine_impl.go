package core

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/anthropics/monocle/internal/db"
	"github.com/anthropics/monocle/internal/protocol"
	"github.com/anthropics/monocle/internal/types"
	"github.com/google/uuid"
)

// Engine implements EngineAPI and coordinates all Monocle subsystems.
type Engine struct {
	mu sync.RWMutex

	cfg       *types.Config
	database  *db.DB
	git       *GitClient
	server    *SocketServer
	feedback  *FeedbackQueue
	formatter *ReviewFormatter
	sessions  *SessionManager

	current *types.ReviewSession

	// autoAdvanceRef: when true, baseRef advances to HEAD on each refresh
	autoAdvanceRef bool
	lastKnownHead  string

	// event subscribers: EventKind -> subscriber ID -> callback
	subscribers map[EventKind]map[int]EventCallback
	nextSubID   int
}

// NewEngine constructs an Engine with all subsystems wired together.
func NewEngine(cfg *types.Config, database *db.DB, repoRoot string) (*Engine, error) {
	git := NewGitClient(repoRoot)
	server := NewSocketServer()
	feedback := NewFeedbackQueue()

	e := &Engine{
		cfg:            cfg,
		database:       database,
		git:            git,
		server:         server,
		feedback:       feedback,
		sessions:       NewSessionManager(database, git),
		autoAdvanceRef: true,
		subscribers:    make(map[EventKind]map[int]EventCallback),
	}

	e.formatter = NewReviewFormatter(func(path string, start, end int) string {
		content, err := git.FileContent("", path)
		if err != nil {
			return ""
		}
		return extractLines(content, start, end)
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
	e.mu.RUnlock()
	if session == nil {
		return nil, fmt.Errorf("no active session")
	}
	return e.git.FileDiff(session.BaseRef, path)
}

func (e *Engine) GetFileContent(path string) (string, error) {
	return e.git.FileContent("", path)
}

func (e *Engine) GetContentItem(id string) (*types.ContentItem, error) {
	return e.database.GetContentItem(id)
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

func (e *Engine) EditComment(commentID string, body string) (*types.ReviewComment, error) {
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

func (e *Engine) DismissOutdated() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.current == nil {
		return fmt.Errorf("no active session")
	}

	if err := e.database.DismissOutdated(e.current.ID); err != nil {
		return fmt.Errorf("dismiss outdated: %w", err)
	}

	active := e.current.Comments[:0]
	for _, c := range e.current.Comments {
		if !c.Outdated {
			active = append(active, c)
		}
	}
	e.current.Comments = active

	return nil
}

// -- Review status --

func (e *Engine) MarkReviewed(path string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.current == nil {
		return fmt.Errorf("no active session")
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

// -- Submission --

func (e *Engine) GetReviewSummary() (*types.ReviewSummary, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.current == nil {
		return nil, fmt.Errorf("no active session")
	}

	summary := &types.ReviewSummary{
		Session:         e.current,
		FileComments:    make(map[string][]types.ReviewComment),
		ContentComments: make(map[string][]types.ReviewComment),
		DeliveryStatus:  e.feedback.GetStatus(),
	}

	for _, c := range e.current.Comments {
		if c.Outdated {
			continue
		}
		switch c.TargetType {
		case types.TargetFile:
			summary.FileComments[c.TargetRef] = append(summary.FileComments[c.TargetRef], c)
		case types.TargetContent:
			summary.ContentComments[c.TargetRef] = append(summary.ContentComments[c.TargetRef], c)
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

func (e *Engine) Submit() (*types.SubmitResult, error) {
	e.mu.RLock()
	session := e.current
	e.mu.RUnlock()

	if session == nil {
		return nil, fmt.Errorf("no active session")
	}

	formatted := e.formatter.Format(session, session.Comments)

	e.feedback.Submit(formatted)

	// Save submission record
	sub := &types.ReviewSubmission{
		ID:              uuid.New().String(),
		SessionID:       session.ID,
		Action:          types.SubmitAction(formatted.Action),
		FormattedReview: formatted.Formatted,
		CommentCount:    formatted.CommentCount,
		ReviewRound:     session.ReviewRound,
		SubmittedAt:     time.Now(),
	}
	_ = e.database.CreateSubmission(session.ID, sub)

	e.emit(EventFeedbackStatusChanged, EventPayload{
		Kind:   EventFeedbackStatusChanged,
		Status: e.feedback.GetStatus(),
	})

	e.emit(EventFeedbackSubmitted, EventPayload{
		Kind:    EventFeedbackSubmitted,
		Message: formatted.Formatted,
		Status:  formatted.Action,
	})

	return &types.SubmitResult{
		Delivered: false,
		Queued:    true,
	}, nil
}

func (e *Engine) Approve() (*types.SubmitResult, error) {
	e.mu.RLock()
	session := e.current
	e.mu.RUnlock()

	if session == nil {
		return nil, fmt.Errorf("no active session")
	}

	e.feedback.Approve()

	e.emit(EventFeedbackStatusChanged, EventPayload{
		Kind:   EventFeedbackStatusChanged,
		Status: e.feedback.GetStatus(),
	})

	e.emit(EventFeedbackSubmitted, EventPayload{
		Kind:    EventFeedbackSubmitted,
		Message: "## Code Review — Approved\n\nNo issues found. Code looks good!",
		Status:  string(types.ActionApprove),
	})

	return &types.SubmitResult{
		Delivered: false,
		Queued:    false,
	}, nil
}

// -- Base ref management --

// SetBaseRef manually sets the diff baseline and disables auto-advance.
func (e *Engine) SetBaseRef(ref string) error {
	// Resolve the ref to a full hash
	resolved, err := e.git.run("rev-parse", ref)
	if err != nil {
		return fmt.Errorf("resolve ref %q: %w", ref, err)
	}
	resolved = strings.TrimSpace(resolved)

	e.mu.Lock()
	defer e.mu.Unlock()

	if e.current == nil {
		return fmt.Errorf("no active session")
	}

	e.current.BaseRef = resolved
	e.current.UpdatedAt = time.Now()
	e.autoAdvanceRef = false
	_ = e.database.UpdateSession(e.current)

	return nil
}

// SetAutoAdvanceRef enables or disables auto-advancing baseRef to HEAD on each refresh.
func (e *Engine) SetAutoAdvanceRef(enabled bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.autoAdvanceRef = enabled
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
			for _, c := range e.current.Comments {
				if !c.Outdated {
					commentCount++
				}
			}
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
func (e *Engine) SubmitContentForReview(id, title, content, contentType string) error {
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

	// Persist to DB
	_ = e.database.UpsertContentItem(session.ID, &item)

	e.emit(EventContentItemAdded, EventPayload{
		Kind:   EventContentItemAdded,
		ItemID: id,
	})

	return nil
}

// RequestPause sets the pause flag so the agent sees "pause_requested" on next status check.
func (e *Engine) RequestPause() {
	e.feedback.SetPauseRequested(true)

	e.mu.Lock()
	if e.current != nil {
		e.current.AgentStatus = types.AgentStatusPaused
		_ = e.database.UpdateSession(e.current)
	}
	e.mu.Unlock()

	e.emit(EventAgentStatusChanged, EventPayload{
		Kind:   EventAgentStatusChanged,
		Status: string(types.AgentStatusPaused),
	})
	e.emit(EventPauseChanged, EventPayload{
		Kind:   EventPauseChanged,
		Status: "pause_requested",
	})
}

// CancelPause clears the pause flag.
func (e *Engine) CancelPause() {
	e.feedback.SetPauseRequested(false)

	e.mu.Lock()
	if e.current != nil {
		e.current.AgentStatus = types.AgentStatusWorking
		_ = e.database.UpdateSession(e.current)
	}
	e.mu.Unlock()

	e.emit(EventAgentStatusChanged, EventPayload{
		Kind:   EventAgentStatusChanged,
		Status: string(types.AgentStatusWorking),
	})
	e.emit(EventPauseChanged, EventPayload{
		Kind:   EventPauseChanged,
		Status: "cancelled",
	})
}

// -- Agent status --

func (e *Engine) GetAgentStatus() types.AgentStatus {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if e.current == nil {
		return types.AgentStatusIdle
	}
	return e.current.AgentStatus
}

func (e *Engine) GetFeedbackStatus() string {
	return e.feedback.GetStatus()
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
func (e *Engine) handlePollFeedback(msg *protocol.PollFeedbackMsg) *protocol.PollFeedbackResponse {
	if msg.Wait {
		review := e.WaitForFeedback()

		// After feedback is delivered, advance the review round
		e.mu.Lock()
		if e.current != nil {
			e.current.AgentStatus = types.AgentStatusWorking
			_ = e.sessions.AdvanceRound(e.current)
			_ = e.database.UpdateSession(e.current)
		}
		e.mu.Unlock()

		e.emit(EventAgentStatusChanged, EventPayload{
			Kind:   EventAgentStatusChanged,
			Status: string(types.AgentStatusWorking),
		})
		e.emit(EventFileChanged, EventPayload{
			Kind: EventFileChanged,
		})

		return &protocol.PollFeedbackResponse{
			Type:         protocol.TypePollFeedbackResponse,
			HasFeedback:  true,
			Feedback:     review.Formatted,
			CommentCount: review.CommentCount,
		}
	}

	// Non-blocking poll
	review := e.PollFeedback()
	if review == nil {
		return &protocol.PollFeedbackResponse{
			Type:        protocol.TypePollFeedbackResponse,
			HasFeedback: false,
		}
	}

	// After feedback is delivered, advance the review round
	e.mu.Lock()
	if e.current != nil {
		_ = e.sessions.AdvanceRound(e.current)
	}
	e.mu.Unlock()

	e.emit(EventFileChanged, EventPayload{
		Kind: EventFileChanged,
	})

	return &protocol.PollFeedbackResponse{
		Type:         protocol.TypePollFeedbackResponse,
		HasFeedback:  true,
		Feedback:     review.Formatted,
		CommentCount: review.CommentCount,
	}
}

// handleSubmitContent receives reviewable content (plans, docs) from the agent.
func (e *Engine) handleSubmitContent(msg *protocol.SubmitContentMsg) *protocol.SubmitContentResponse {
	id := msg.ID
	if id == "" {
		id = uuid.New().String()
	}

	err := e.SubmitContentForReview(id, msg.Title, msg.Content, msg.ContentType)
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

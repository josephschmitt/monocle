package core

import (
	"github.com/josephschmitt/monocle/internal/types"
)

// SubmitResult contains information about the outcome of a review submission.
type SubmitResult struct {
	// AgentConnected indicates whether an agent was connected at submit time.
	// When false, the review was saved but may not have been delivered.
	AgentConnected bool
}

// EventKind represents the type of engine event.
type EventKind string

const (
	EventFileChanged           EventKind = "file_changed"
	EventFeedbackStatusChanged EventKind = "feedback_status_changed"
	EventContentItemAdded      EventKind = "content_item_added"
	EventPauseChanged          EventKind = "pause_changed"
	EventFeedbackSubmitted     EventKind = "feedback_submitted"
	EventConnectionChanged     EventKind = "connection_changed"
	EventAdditionalFileAdded   EventKind = "additional_file_added"
	EventFeedbackPickedUp      EventKind = "feedback_picked_up"
	EventWaitStatusChanged     EventKind = "wait_status_changed"
)

// EventPayload carries data for an engine event.
type EventPayload struct {
	Kind    EventKind
	Path    string // for file events
	ItemID  string // for content item events
	Status  string // for status events
	Message string // optional context
}

// EventCallback is the signature for event subscribers.
type EventCallback func(EventPayload)

// UnsubscribeFunc removes an event subscription when called.
type UnsubscribeFunc func()

// CommentTarget identifies where a comment is attached.
type CommentTarget struct {
	TargetType types.TargetType
	TargetRef  string // file path or content item ID
	LineStart  int
	LineEnd    int
}

// SessionOptions configures a new session.
type SessionOptions struct {
	Agent          string
	RepoRoot       string
	BaseRef        string
	IgnorePatterns []string
}

// ListSessionsOptions filters session listings.
type ListSessionsOptions struct {
	RepoRoot string
	Limit    int
}

// EngineAPI defines the interface between the TUI and the engine.
// The TUI only depends on this interface — never on engine internals.
type EngineAPI interface {
	// Session lifecycle
	StartSession(opts SessionOptions) (*types.ReviewSession, error)
	ResumeSession(sessionID string) (*types.ReviewSession, error)
	GetSession() *types.ReviewSession
	ListSessions(opts ListSessionsOptions) ([]types.SessionSummary, error)

	// Browsing
	RefreshChangedFiles() ([]types.ChangedFile, error)
	GetChangedFiles() []types.ChangedFile
	GetContentItems() []types.ContentItem
	GetFileDiff(path string) (*types.DiffResult, error)
	GetFileContent(path string) (string, error)
	GetContentItem(id string) (*types.ContentItem, error)
	GetContentDiff(id string) (*types.DiffResult, error)
	GetContentVersions(id string) ([]types.ContentVersion, error)
	GetContentDiffBetweenVersions(id string, fromVersion, toVersion int) (*types.DiffResult, error)

	// Additional files
	GetAdditionalFiles() []types.AdditionalFile
	AddAdditionalPaths(paths []string) ([]types.AdditionalFile, error)
	GetAdditionalFileContent(absPath string) (string, error)

	// Commenting
	AddComment(target CommentTarget, commentType types.CommentType, body string) (*types.ReviewComment, error)
	EditComment(commentID string, commentType types.CommentType, body string) (*types.ReviewComment, error)
	DeleteComment(commentID string) error
	ResolveComment(commentID string) error
	ClearComments() error
	ClearReview() error

	// Review status
	MarkReviewed(path string) error
	UnmarkReviewed(path string) error
	MarkContentReviewed(id string) error
	UnmarkContentReviewed(id string) error
	ResetAllReviewed() error
	MarkAllReviewed() error

	// Submission
	GetReviewSummary() (*types.ReviewSummary, error)
	Submit(action types.SubmitAction, body string) (*SubmitResult, error)
	FormatReview(action types.SubmitAction, body string) (string, error)
	GetSubmissions() ([]types.ReviewSubmission, error)

	// Base ref management
	SetBaseRef(ref string) error
	SetAutoAdvanceRef(enabled bool)
	IsAutoAdvanceRef() bool
	SelectedBaseRef() string
	RecentCommits(n int) ([]LogEntry, error)

	// Review snapshots
	GetSnapshots() ([]types.ReviewSnapshot, error)
	SetSnapshotBase(snapshotID int) error
	ClearSnapshotBase()
	GetActiveSnapshot() *types.ReviewSnapshot
	HasSnapshots() (bool, error)

	// Server (socket for MCP channel)
	StartServer(socketPath string) error

	// Feedback (MCP channel)
	PollFeedback() *FormattedReview
	WaitForFeedback() *FormattedReview
	GetReviewStatusInfo() *ReviewStatusInfo
	SubmitContentForReview(id, title, content, contentType string, isPlan bool) error
	RequestPause()
	CancelPause()

	// Feedback status
	GetFeedbackStatus() string
	GetQueuedCount() int
	ReloadPendingFeedback()

	// Connection status
	GetSubscriberCount() int
	GetSocketPath() string

	// Events
	On(event EventKind, callback EventCallback) UnsubscribeFunc

	// Config
	GetConfig() *types.Config
	SaveConfig() error

	// Lifecycle
	Shutdown()
}

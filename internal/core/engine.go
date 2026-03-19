package core

import (
	"github.com/anthropics/monocle/internal/types"
)

// EventKind represents the type of engine event.
type EventKind string

const (
	EventFileChanged           EventKind = "file_changed"
	EventAgentStatusChanged    EventKind = "agent_status_changed"
	EventFeedbackStatusChanged EventKind = "feedback_status_changed"
	EventContentItemAdded      EventKind = "content_item_added"
	EventCommentsOutdated      EventKind = "comments_outdated"
	EventPauseChanged          EventKind = "pause_changed"
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

	// Commenting
	AddComment(target CommentTarget, commentType types.CommentType, body string) (*types.ReviewComment, error)
	EditComment(commentID string, body string) (*types.ReviewComment, error)
	DeleteComment(commentID string) error
	DismissOutdated() error

	// Review status
	MarkReviewed(path string) error
	UnmarkReviewed(path string) error

	// Submission
	GetReviewSummary() (*types.ReviewSummary, error)
	Submit() (*types.SubmitResult, error)
	Approve() (*types.SubmitResult, error)

	// Server (socket for CLI subcommands)
	StartServer(socketPath string) error

	// Feedback (skills-based model)
	PollFeedback() *FormattedReview
	WaitForFeedback() *FormattedReview
	GetReviewStatusInfo() *ReviewStatusInfo
	SubmitContentForReview(id, title, content, contentType string) error
	RequestPause()
	CancelPause()

	// Agent status
	GetAgentStatus() types.AgentStatus
	GetFeedbackStatus() string

	// Events
	On(event EventKind, callback EventCallback) UnsubscribeFunc

	// Lifecycle
	Shutdown()
}

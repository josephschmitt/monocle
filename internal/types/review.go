package types

import "time"

type AgentStatus string

const (
	AgentStatusIdle    AgentStatus = "idle"
	AgentStatusWorking AgentStatus = "working"
	AgentStatusPaused  AgentStatus = "paused"
)

type FileChangeStatus string

const (
	FileAdded    FileChangeStatus = "added"
	FileModified FileChangeStatus = "modified"
	FileDeleted  FileChangeStatus = "deleted"
	FileRenamed  FileChangeStatus = "renamed"
)

type CommentType string

const (
	CommentIssue      CommentType = "issue"
	CommentSuggestion CommentType = "suggestion"
	CommentNote       CommentType = "note"
	CommentPraise     CommentType = "praise"
)

type TargetType string

const (
	TargetFile    TargetType = "file"
	TargetContent TargetType = "content"
)

type SubmitAction string

const (
	ActionRequestChanges SubmitAction = "request_changes"
	ActionApprove        SubmitAction = "approve"
)

type ReviewSession struct {
	ID             string
	Agent          string
	AgentStatus    AgentStatus
	RepoRoot       string
	BaseRef        string
	ChangedFiles   []ChangedFile
	ContentItems   []ContentItem
	Comments       []ReviewComment
	FileStatuses   map[string]bool // path -> reviewed
	IgnorePatterns []string
	ReviewRound    int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type ChangedFile struct {
	Path     string
	Status   FileChangeStatus
	Hunks    []DiffHunk
	Reviewed bool
	Comments []ReviewComment
}

type ContentItem struct {
	ID          string
	Title       string
	Content     string
	ContentType string
	Reviewed    bool
	Comments    []ReviewComment
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ReviewComment struct {
	ID          string
	TargetType  TargetType
	TargetRef   string // file path or content item ID
	LineStart   int
	LineEnd     int
	Type        CommentType
	Body        string
	CodeSnippet string
	Resolved    bool
	Outdated    bool
	ReviewRound int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ReviewSubmission struct {
	ID              string
	SessionID       string
	Action          SubmitAction
	FormattedReview string
	CommentCount    int
	ReviewRound     int
	SubmittedAt     time.Time
}

type DiffHunk struct {
	OldStart int
	OldCount int
	NewStart int
	NewCount int
	Header   string
	Lines    []DiffLine
}

type DiffLineKind string

const (
	DiffLineContext DiffLineKind = "context"
	DiffLineAdded  DiffLineKind = "added"
	DiffLineRemoved DiffLineKind = "removed"
)

type DiffLine struct {
	Kind       DiffLineKind
	OldLineNum int
	NewLineNum int
	Content    string
}

type DiffResult struct {
	Path  string
	Hunks []DiffHunk
	Raw   string
}

type ReviewSummary struct {
	Session         *ReviewSession
	FileComments    map[string][]ReviewComment // path -> comments
	ContentComments map[string][]ReviewComment // item id -> comments
	IssueCt         int
	SuggestionCt    int
	NoteCt          int
	PraiseCt        int
	DeliveryStatus  string
}

type SubmitResult struct {
	Delivered bool
	Queued    bool
}

type SessionSummary struct {
	ID           string
	Agent        string
	RepoRoot     string
	FileCount    int
	CommentCount int
	ReviewRound  int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

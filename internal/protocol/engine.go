package protocol

import "github.com/josephschmitt/monocle/internal/types"

// Message types for the EngineAPI surface — the TUI and other frontends call
// these instead of holding a direct *core.Engine reference.

// Inbound (frontend → engine)
const (
	// Sessions
	TypeStartSession  = "start_session"
	TypeResumeSession = "resume_session"
	TypeGetSession    = "get_session"
	TypeListSessions  = "list_sessions"

	// Files
	TypeRefreshChangedFiles = "refresh_changed_files"
	TypeGetChangedFiles     = "get_changed_files"
	TypeGetFileDiff         = "get_file_diff"
	TypeGetFileContent      = "get_file_content"

	// Content
	TypeGetContentItems              = "get_content_items"
	TypeGetContentItem               = "get_content_item"
	TypeGetContentDiff               = "get_content_diff"
	TypeGetContentVersions           = "get_content_versions"
	TypeGetContentDiffBetweenVersion = "get_content_diff_between_versions"
	TypeDismissArtifact              = "dismiss_artifact"

	// Additional files
	TypeGetAdditionalFiles       = "get_additional_files"
	TypeGetAdditionalFileContent = "get_additional_file_content"

	// Comments
	TypeAddComment     = "add_comment"
	TypeEditComment    = "edit_comment"
	TypeDeleteComment  = "delete_comment"
	TypeResolveComment = "resolve_comment"
	TypeClearComments  = "clear_comments"
	TypeClearReview    = "clear_review"

	// Marking
	TypeMarkReviewed          = "mark_reviewed"
	TypeUnmarkReviewed        = "unmark_reviewed"
	TypeMarkContentReviewed   = "mark_content_reviewed"
	TypeUnmarkContentReviewed = "unmark_content_reviewed"
	TypeResetAllReviewed      = "reset_all_reviewed"
	TypeMarkAllReviewed       = "mark_all_reviewed"

	// Submission
	TypeGetReviewSummary = "get_review_summary"
	TypeSubmit           = "submit"
	TypeFormatReview     = "format_review"
	TypeGetSubmissions   = "get_submissions"

	// Base ref
	TypeSetBaseRef         = "set_base_ref"
	TypeSetAutoAdvanceRef  = "set_auto_advance_ref"
	TypeIsAutoAdvanceRef   = "is_auto_advance_ref"
	TypeSelectedBaseRef    = "selected_base_ref"
	TypeRecentCommits      = "recent_commits"

	// Snapshots
	TypeGetSnapshots       = "get_snapshots"
	TypeSetSnapshotBase    = "set_snapshot_base"
	TypeClearSnapshotBase  = "clear_snapshot_base"
	TypeGetActiveSnapshot  = "get_active_snapshot"
	TypeHasSnapshots       = "has_snapshots"

	// Config
	TypeGetConfig              = "get_config"
	TypeSaveConfig             = "save_config"
	TypeIsReviewTrackingEnabled = "is_review_tracking_enabled"

	// Status
	TypeGetFeedbackStatus     = "get_feedback_status"
	TypeGetQueuedCount        = "get_queued_count"
	TypeReloadPendingFeedback = "reload_pending_feedback"
	TypeGetSubscriberCount    = "get_subscriber_count"
	TypeGetSocketPath         = "get_socket_path"

	// Pause flow
	TypeSetPause = "set_pause"
)

// Outbound (engine → frontend)
const (
	// Sessions
	TypeStartSessionResponse  = "start_session_response"
	TypeResumeSessionResponse = "resume_session_response"
	TypeGetSessionResponse    = "get_session_response"
	TypeListSessionsResponse  = "list_sessions_response"

	// Files
	TypeRefreshChangedFilesResponse = "refresh_changed_files_response"
	TypeGetChangedFilesResponse     = "get_changed_files_response"
	TypeGetFileDiffResponse         = "get_file_diff_response"
	TypeGetFileContentResponse      = "get_file_content_response"

	// Content
	TypeGetContentItemsResponse              = "get_content_items_response"
	TypeGetContentItemResponse               = "get_content_item_response"
	TypeGetContentDiffResponse               = "get_content_diff_response"
	TypeGetContentVersionsResponse           = "get_content_versions_response"
	TypeGetContentDiffBetweenVersionResponse = "get_content_diff_between_versions_response"
	TypeDismissArtifactResponse              = "dismiss_artifact_response"

	// Additional files
	TypeGetAdditionalFilesResponse       = "get_additional_files_response"
	TypeGetAdditionalFileContentResponse = "get_additional_file_content_response"

	// Comments
	TypeAddCommentResponse     = "add_comment_response"
	TypeEditCommentResponse    = "edit_comment_response"
	TypeDeleteCommentResponse  = "delete_comment_response"
	TypeResolveCommentResponse = "resolve_comment_response"
	TypeClearCommentsResponse  = "clear_comments_response"
	TypeClearReviewResponse    = "clear_review_response"

	// Marking
	TypeMarkReviewedResponse          = "mark_reviewed_response"
	TypeUnmarkReviewedResponse        = "unmark_reviewed_response"
	TypeMarkContentReviewedResponse   = "mark_content_reviewed_response"
	TypeUnmarkContentReviewedResponse = "unmark_content_reviewed_response"
	TypeResetAllReviewedResponse      = "reset_all_reviewed_response"
	TypeMarkAllReviewedResponse       = "mark_all_reviewed_response"

	// Submission
	TypeGetReviewSummaryResponse = "get_review_summary_response"
	TypeSubmitResponse           = "submit_response"
	TypeFormatReviewResponse     = "format_review_response"
	TypeGetSubmissionsResponse   = "get_submissions_response"

	// Base ref
	TypeSetBaseRefResponse        = "set_base_ref_response"
	TypeSetAutoAdvanceRefResponse = "set_auto_advance_ref_response"
	TypeIsAutoAdvanceRefResponse  = "is_auto_advance_ref_response"
	TypeSelectedBaseRefResponse   = "selected_base_ref_response"
	TypeRecentCommitsResponse     = "recent_commits_response"

	// Snapshots
	TypeGetSnapshotsResponse      = "get_snapshots_response"
	TypeSetSnapshotBaseResponse   = "set_snapshot_base_response"
	TypeClearSnapshotBaseResponse = "clear_snapshot_base_response"
	TypeGetActiveSnapshotResponse = "get_active_snapshot_response"
	TypeHasSnapshotsResponse      = "has_snapshots_response"

	// Config
	TypeGetConfigResponse               = "get_config_response"
	TypeSaveConfigResponse              = "save_config_response"
	TypeIsReviewTrackingEnabledResponse = "is_review_tracking_enabled_response"

	// Status
	TypeGetFeedbackStatusResponse     = "get_feedback_status_response"
	TypeGetQueuedCountResponse        = "get_queued_count_response"
	TypeReloadPendingFeedbackResponse = "reload_pending_feedback_response"
	TypeGetSubscriberCountResponse    = "get_subscriber_count_response"
	TypeGetSocketPathResponse         = "get_socket_path_response"

	// Pause flow
	TypeSetPauseResponse = "set_pause_response"
)

// LogEntry mirrors core.LogEntry for wire transmission (protocol must not
// import core). Trivially convertible on both sides.
type LogEntry struct {
	Hash    string `json:"hash"`
	Subject string `json:"subject"`
}

// --- Sessions ---

type StartSessionMsg struct {
	Type           string   `json:"type"`
	Agent          string   `json:"agent"`
	RepoRoot       string   `json:"repo_root"`
	BaseRef        string   `json:"base_ref,omitempty"`
	IgnorePatterns []string `json:"ignore_patterns,omitempty"`
}

type StartSessionResponse struct {
	Type    string                `json:"type"`
	Session *types.ReviewSession  `json:"session,omitempty"`
	Error   string                `json:"error,omitempty"`
}

type ResumeSessionMsg struct {
	Type      string `json:"type"`
	SessionID string `json:"session_id"`
}

type ResumeSessionResponse struct {
	Type    string                `json:"type"`
	Session *types.ReviewSession  `json:"session,omitempty"`
	Error   string                `json:"error,omitempty"`
}

type GetSessionMsg struct {
	Type string `json:"type"`
}

type GetSessionResponse struct {
	Type    string                `json:"type"`
	Session *types.ReviewSession  `json:"session,omitempty"`
}

type ListSessionsMsg struct {
	Type     string `json:"type"`
	RepoRoot string `json:"repo_root,omitempty"`
	Limit    int    `json:"limit,omitempty"`
}

type ListSessionsResponse struct {
	Type     string                 `json:"type"`
	Sessions []types.SessionSummary `json:"sessions,omitempty"`
	Error    string                 `json:"error,omitempty"`
}

// --- Files ---

type RefreshChangedFilesMsg struct {
	Type string `json:"type"`
}

type RefreshChangedFilesResponse struct {
	Type  string              `json:"type"`
	Files []types.ChangedFile `json:"files,omitempty"`
	Error string              `json:"error,omitempty"`
}

type GetChangedFilesMsg struct {
	Type string `json:"type"`
}

type GetChangedFilesResponse struct {
	Type  string              `json:"type"`
	Files []types.ChangedFile `json:"files,omitempty"`
}

type GetFileDiffMsg struct {
	Type string `json:"type"`
	Path string `json:"path"`
}

type GetFileDiffResponse struct {
	Type  string            `json:"type"`
	Diff  *types.DiffResult `json:"diff,omitempty"`
	Error string            `json:"error,omitempty"`
}

type GetFileContentMsg struct {
	Type string `json:"type"`
	Path string `json:"path"`
}

type GetFileContentResponse struct {
	Type    string `json:"type"`
	Content string `json:"content,omitempty"`
	Error   string `json:"error,omitempty"`
}

// --- Content (artifacts) ---

type GetContentItemsMsg struct {
	Type string `json:"type"`
}

type GetContentItemsResponse struct {
	Type  string              `json:"type"`
	Items []types.ContentItem `json:"items,omitempty"`
}

type GetContentItemMsg struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type GetContentItemResponse struct {
	Type  string             `json:"type"`
	Item  *types.ContentItem `json:"item,omitempty"`
	Error string             `json:"error,omitempty"`
}

type GetContentDiffMsg struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type GetContentDiffResponse struct {
	Type  string            `json:"type"`
	Diff  *types.DiffResult `json:"diff,omitempty"`
	Error string            `json:"error,omitempty"`
}

type GetContentVersionsMsg struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type GetContentVersionsResponse struct {
	Type     string                 `json:"type"`
	Versions []types.ContentVersion `json:"versions,omitempty"`
	Error    string                 `json:"error,omitempty"`
}

type GetContentDiffBetweenVersionsMsg struct {
	Type        string `json:"type"`
	ID          string `json:"id"`
	FromVersion int    `json:"from_version"`
	ToVersion   int    `json:"to_version"`
}

type GetContentDiffBetweenVersionsResponse struct {
	Type  string            `json:"type"`
	Diff  *types.DiffResult `json:"diff,omitempty"`
	Error string            `json:"error,omitempty"`
}

type DismissArtifactMsg struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type DismissArtifactResponse struct {
	Type  string `json:"type"`
	Error string `json:"error,omitempty"`
}

// --- Additional files ---

type GetAdditionalFilesMsg struct {
	Type string `json:"type"`
}

type GetAdditionalFilesResponse struct {
	Type  string                 `json:"type"`
	Files []types.AdditionalFile `json:"files,omitempty"`
}

type GetAdditionalFileContentMsg struct {
	Type    string `json:"type"`
	AbsPath string `json:"abs_path"`
}

type GetAdditionalFileContentResponse struct {
	Type    string `json:"type"`
	Content string `json:"content,omitempty"`
	Error   string `json:"error,omitempty"`
}

// --- Comments ---

type AddCommentMsg struct {
	Type        string            `json:"type"`
	TargetType  types.TargetType  `json:"target_type"`
	TargetRef   string            `json:"target_ref"`
	LineStart   int               `json:"line_start"`
	LineEnd     int               `json:"line_end"`
	CommentType types.CommentType `json:"comment_type"`
	Body        string            `json:"body"`
}

type AddCommentResponse struct {
	Type    string               `json:"type"`
	Comment *types.ReviewComment `json:"comment,omitempty"`
	Error   string               `json:"error,omitempty"`
}

type EditCommentMsg struct {
	Type        string            `json:"type"`
	CommentID   string            `json:"comment_id"`
	CommentType types.CommentType `json:"comment_type"`
	Body        string            `json:"body"`
}

type EditCommentResponse struct {
	Type    string               `json:"type"`
	Comment *types.ReviewComment `json:"comment,omitempty"`
	Error   string               `json:"error,omitempty"`
}

type DeleteCommentMsg struct {
	Type      string `json:"type"`
	CommentID string `json:"comment_id"`
}

type DeleteCommentResponse struct {
	Type  string `json:"type"`
	Error string `json:"error,omitempty"`
}

type ResolveCommentMsg struct {
	Type      string `json:"type"`
	CommentID string `json:"comment_id"`
}

type ResolveCommentResponse struct {
	Type  string `json:"type"`
	Error string `json:"error,omitempty"`
}

type ClearCommentsMsg struct {
	Type string `json:"type"`
}

type ClearCommentsResponse struct {
	Type  string `json:"type"`
	Error string `json:"error,omitempty"`
}

type ClearReviewMsg struct {
	Type string `json:"type"`
}

type ClearReviewResponse struct {
	Type  string `json:"type"`
	Error string `json:"error,omitempty"`
}

// --- Review marking ---

type MarkReviewedMsg struct {
	Type string `json:"type"`
	Path string `json:"path"`
}

type MarkReviewedResponse struct {
	Type  string `json:"type"`
	Error string `json:"error,omitempty"`
}

type UnmarkReviewedMsg struct {
	Type string `json:"type"`
	Path string `json:"path"`
}

type UnmarkReviewedResponse struct {
	Type  string `json:"type"`
	Error string `json:"error,omitempty"`
}

type MarkContentReviewedMsg struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type MarkContentReviewedResponse struct {
	Type  string `json:"type"`
	Error string `json:"error,omitempty"`
}

type UnmarkContentReviewedMsg struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type UnmarkContentReviewedResponse struct {
	Type  string `json:"type"`
	Error string `json:"error,omitempty"`
}

type ResetAllReviewedMsg struct {
	Type string `json:"type"`
}

type ResetAllReviewedResponse struct {
	Type  string `json:"type"`
	Error string `json:"error,omitempty"`
}

type MarkAllReviewedMsg struct {
	Type string `json:"type"`
}

type MarkAllReviewedResponse struct {
	Type  string `json:"type"`
	Error string `json:"error,omitempty"`
}

// --- Submission ---

type GetReviewSummaryMsg struct {
	Type string `json:"type"`
}

type GetReviewSummaryResponse struct {
	Type    string               `json:"type"`
	Summary *types.ReviewSummary `json:"summary,omitempty"`
	Error   string               `json:"error,omitempty"`
}

type SubmitMsg struct {
	Type   string             `json:"type"`
	Action types.SubmitAction `json:"action"`
	Body   string             `json:"body"`
}

type SubmitResponse struct {
	Type           string `json:"type"`
	AgentConnected bool   `json:"agent_connected"`
	Error          string `json:"error,omitempty"`
}

type FormatReviewMsg struct {
	Type   string             `json:"type"`
	Action types.SubmitAction `json:"action"`
	Body   string             `json:"body"`
}

type FormatReviewResponse struct {
	Type      string `json:"type"`
	Formatted string `json:"formatted,omitempty"`
	Error     string `json:"error,omitempty"`
}

type GetSubmissionsMsg struct {
	Type string `json:"type"`
}

type GetSubmissionsResponse struct {
	Type        string                   `json:"type"`
	Submissions []types.ReviewSubmission `json:"submissions,omitempty"`
	Error       string                   `json:"error,omitempty"`
}

// --- Base ref management ---

type SetBaseRefMsg struct {
	Type string `json:"type"`
	Ref  string `json:"ref"`
}

type SetBaseRefResponse struct {
	Type  string `json:"type"`
	Error string `json:"error,omitempty"`
}

type SetAutoAdvanceRefMsg struct {
	Type    string `json:"type"`
	Enabled bool   `json:"enabled"`
}

type SetAutoAdvanceRefResponse struct {
	Type string `json:"type"`
}

type IsAutoAdvanceRefMsg struct {
	Type string `json:"type"`
}

type IsAutoAdvanceRefResponse struct {
	Type    string `json:"type"`
	Enabled bool   `json:"enabled"`
}

type SelectedBaseRefMsg struct {
	Type string `json:"type"`
}

type SelectedBaseRefResponse struct {
	Type string `json:"type"`
	Ref  string `json:"ref,omitempty"`
}

type RecentCommitsMsg struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

type RecentCommitsResponse struct {
	Type    string     `json:"type"`
	Commits []LogEntry `json:"commits,omitempty"`
	Error   string     `json:"error,omitempty"`
}

// --- Snapshots ---

type GetSnapshotsMsg struct {
	Type string `json:"type"`
}

type GetSnapshotsResponse struct {
	Type      string                 `json:"type"`
	Snapshots []types.ReviewSnapshot `json:"snapshots,omitempty"`
	Error     string                 `json:"error,omitempty"`
}

type SetSnapshotBaseMsg struct {
	Type       string `json:"type"`
	SnapshotID int    `json:"snapshot_id"`
}

type SetSnapshotBaseResponse struct {
	Type  string `json:"type"`
	Error string `json:"error,omitempty"`
}

type ClearSnapshotBaseMsg struct {
	Type string `json:"type"`
}

type ClearSnapshotBaseResponse struct {
	Type string `json:"type"`
}

type GetActiveSnapshotMsg struct {
	Type string `json:"type"`
}

type GetActiveSnapshotResponse struct {
	Type     string                `json:"type"`
	Snapshot *types.ReviewSnapshot `json:"snapshot,omitempty"`
}

type HasSnapshotsMsg struct {
	Type string `json:"type"`
}

type HasSnapshotsResponse struct {
	Type  string `json:"type"`
	Has   bool   `json:"has"`
	Error string `json:"error,omitempty"`
}

// --- Config ---

type GetConfigMsg struct {
	Type string `json:"type"`
}

type GetConfigResponse struct {
	Type   string        `json:"type"`
	Config *types.Config `json:"config,omitempty"`
}

type SaveConfigMsg struct {
	Type   string       `json:"type"`
	Config types.Config `json:"config"`
}

type SaveConfigResponse struct {
	Type  string `json:"type"`
	Error string `json:"error,omitempty"`
}

type IsReviewTrackingEnabledMsg struct {
	Type string `json:"type"`
}

type IsReviewTrackingEnabledResponse struct {
	Type    string `json:"type"`
	Enabled bool   `json:"enabled"`
}

// --- Status ---

type GetFeedbackStatusMsg struct {
	Type string `json:"type"`
}

type GetFeedbackStatusResponse struct {
	Type   string `json:"type"`
	Status string `json:"status"`
}

type GetQueuedCountMsg struct {
	Type string `json:"type"`
}

type GetQueuedCountResponse struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

type ReloadPendingFeedbackMsg struct {
	Type string `json:"type"`
}

type ReloadPendingFeedbackResponse struct {
	Type string `json:"type"`
}

type GetSubscriberCountMsg struct {
	Type string `json:"type"`
}

type GetSubscriberCountResponse struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

type GetSocketPathMsg struct {
	Type string `json:"type"`
}

type GetSocketPathResponse struct {
	Type string `json:"type"`
	Path string `json:"path,omitempty"`
}

// --- Pause flow ---

// SetPauseMsg toggles the engine's pause-requested flag. Requested=true is
// the equivalent of Engine.RequestPause; false maps to CancelPause. Carries
// both in one message so a single protocol path covers both directions.
type SetPauseMsg struct {
	Type      string `json:"type"`
	Requested bool   `json:"requested"`
}

type SetPauseResponse struct {
	Type string `json:"type"`
}

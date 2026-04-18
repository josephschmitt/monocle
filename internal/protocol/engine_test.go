package protocol

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/josephschmitt/monocle/internal/types"
)

// TestEngineMessagesRoundTrip exercises Encode → Decode on every new engine
// surface message pair. It catches missing Decode registrations and bad JSON
// tags without needing a running engine.
func TestEngineMessagesRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		msg  any
	}{
		// Sessions
		{"StartSession", &StartSessionMsg{Type: TypeStartSession, Agent: "claude", RepoRoot: "/r", BaseRef: "main"}},
		{"StartSessionResponse", &StartSessionResponse{Type: TypeStartSessionResponse, Session: &types.ReviewSession{ID: "s1"}}},
		{"ResumeSession", &ResumeSessionMsg{Type: TypeResumeSession, SessionID: "s1"}},
		{"ResumeSessionResponse", &ResumeSessionResponse{Type: TypeResumeSessionResponse, Session: &types.ReviewSession{ID: "s1"}}},
		{"GetSession", &GetSessionMsg{Type: TypeGetSession}},
		{"GetSessionResponse", &GetSessionResponse{Type: TypeGetSessionResponse, Session: &types.ReviewSession{ID: "s1"}}},
		{"ListSessions", &ListSessionsMsg{Type: TypeListSessions, Limit: 10}},
		{"ListSessionsResponse", &ListSessionsResponse{Type: TypeListSessionsResponse, Sessions: []types.SessionSummary{{ID: "s1"}}}},

		// Files
		{"RefreshChangedFiles", &RefreshChangedFilesMsg{Type: TypeRefreshChangedFiles}},
		{"RefreshChangedFilesResponse", &RefreshChangedFilesResponse{Type: TypeRefreshChangedFilesResponse, Files: []types.ChangedFile{{Path: "a.go"}}}},
		{"GetChangedFiles", &GetChangedFilesMsg{Type: TypeGetChangedFiles}},
		{"GetChangedFilesResponse", &GetChangedFilesResponse{Type: TypeGetChangedFilesResponse}},
		{"GetFileDiff", &GetFileDiffMsg{Type: TypeGetFileDiff, Path: "a.go"}},
		{"GetFileDiffResponse", &GetFileDiffResponse{Type: TypeGetFileDiffResponse, Diff: &types.DiffResult{Path: "a.go"}}},
		{"GetFileContent", &GetFileContentMsg{Type: TypeGetFileContent, Path: "a.go"}},
		{"GetFileContentResponse", &GetFileContentResponse{Type: TypeGetFileContentResponse, Content: "hello"}},

		// Content
		{"GetContentItems", &GetContentItemsMsg{Type: TypeGetContentItems}},
		{"GetContentItemsResponse", &GetContentItemsResponse{Type: TypeGetContentItemsResponse}},
		{"GetContentItem", &GetContentItemMsg{Type: TypeGetContentItem, ID: "c1"}},
		{"GetContentItemResponse", &GetContentItemResponse{Type: TypeGetContentItemResponse, Item: &types.ContentItem{ID: "c1"}}},
		{"GetContentDiff", &GetContentDiffMsg{Type: TypeGetContentDiff, ID: "c1"}},
		{"GetContentDiffResponse", &GetContentDiffResponse{Type: TypeGetContentDiffResponse}},
		{"GetContentVersions", &GetContentVersionsMsg{Type: TypeGetContentVersions, ID: "c1"}},
		{"GetContentVersionsResponse", &GetContentVersionsResponse{Type: TypeGetContentVersionsResponse}},
		{"GetContentDiffBetweenVersions", &GetContentDiffBetweenVersionsMsg{Type: TypeGetContentDiffBetweenVersion, ID: "c1", FromVersion: 1, ToVersion: 2}},
		{"GetContentDiffBetweenVersionsResponse", &GetContentDiffBetweenVersionsResponse{Type: TypeGetContentDiffBetweenVersionResponse}},
		{"DismissArtifact", &DismissArtifactMsg{Type: TypeDismissArtifact, ID: "c1"}},
		{"DismissArtifactResponse", &DismissArtifactResponse{Type: TypeDismissArtifactResponse}},

		// Additional files
		{"GetAdditionalFiles", &GetAdditionalFilesMsg{Type: TypeGetAdditionalFiles}},
		{"GetAdditionalFilesResponse", &GetAdditionalFilesResponse{Type: TypeGetAdditionalFilesResponse}},
		{"GetAdditionalFileContent", &GetAdditionalFileContentMsg{Type: TypeGetAdditionalFileContent, AbsPath: "/tmp/x"}},
		{"GetAdditionalFileContentResponse", &GetAdditionalFileContentResponse{Type: TypeGetAdditionalFileContentResponse, Content: "x"}},

		// Comments
		{"AddComment", &AddCommentMsg{Type: TypeAddComment, TargetType: types.TargetFile, TargetRef: "a.go", LineStart: 1, LineEnd: 2, CommentType: types.CommentNote, Body: "b"}},
		{"AddCommentResponse", &AddCommentResponse{Type: TypeAddCommentResponse, Comment: &types.ReviewComment{ID: "x"}}},
		{"EditComment", &EditCommentMsg{Type: TypeEditComment, CommentID: "x", CommentType: types.CommentIssue, Body: "b"}},
		{"EditCommentResponse", &EditCommentResponse{Type: TypeEditCommentResponse}},
		{"DeleteComment", &DeleteCommentMsg{Type: TypeDeleteComment, CommentID: "x"}},
		{"DeleteCommentResponse", &DeleteCommentResponse{Type: TypeDeleteCommentResponse}},
		{"ResolveComment", &ResolveCommentMsg{Type: TypeResolveComment, CommentID: "x"}},
		{"ResolveCommentResponse", &ResolveCommentResponse{Type: TypeResolveCommentResponse}},
		{"ClearComments", &ClearCommentsMsg{Type: TypeClearComments}},
		{"ClearCommentsResponse", &ClearCommentsResponse{Type: TypeClearCommentsResponse}},
		{"ClearReview", &ClearReviewMsg{Type: TypeClearReview}},
		{"ClearReviewResponse", &ClearReviewResponse{Type: TypeClearReviewResponse}},

		// Marking
		{"MarkReviewed", &MarkReviewedMsg{Type: TypeMarkReviewed, Path: "a.go"}},
		{"MarkReviewedResponse", &MarkReviewedResponse{Type: TypeMarkReviewedResponse}},
		{"UnmarkReviewed", &UnmarkReviewedMsg{Type: TypeUnmarkReviewed, Path: "a.go"}},
		{"UnmarkReviewedResponse", &UnmarkReviewedResponse{Type: TypeUnmarkReviewedResponse}},
		{"MarkContentReviewed", &MarkContentReviewedMsg{Type: TypeMarkContentReviewed, ID: "c1"}},
		{"MarkContentReviewedResponse", &MarkContentReviewedResponse{Type: TypeMarkContentReviewedResponse}},
		{"UnmarkContentReviewed", &UnmarkContentReviewedMsg{Type: TypeUnmarkContentReviewed, ID: "c1"}},
		{"UnmarkContentReviewedResponse", &UnmarkContentReviewedResponse{Type: TypeUnmarkContentReviewedResponse}},
		{"ResetAllReviewed", &ResetAllReviewedMsg{Type: TypeResetAllReviewed}},
		{"ResetAllReviewedResponse", &ResetAllReviewedResponse{Type: TypeResetAllReviewedResponse}},
		{"MarkAllReviewed", &MarkAllReviewedMsg{Type: TypeMarkAllReviewed}},
		{"MarkAllReviewedResponse", &MarkAllReviewedResponse{Type: TypeMarkAllReviewedResponse}},

		// Submission
		{"GetReviewSummary", &GetReviewSummaryMsg{Type: TypeGetReviewSummary}},
		{"GetReviewSummaryResponse", &GetReviewSummaryResponse{Type: TypeGetReviewSummaryResponse}},
		{"Submit", &SubmitMsg{Type: TypeSubmit, Action: types.ActionApprove, Body: "ok"}},
		{"SubmitResponse", &SubmitResponse{Type: TypeSubmitResponse, AgentConnected: true}},
		{"FormatReview", &FormatReviewMsg{Type: TypeFormatReview, Action: types.ActionRequestChanges, Body: "b"}},
		{"FormatReviewResponse", &FormatReviewResponse{Type: TypeFormatReviewResponse, Formatted: "f"}},
		{"GetSubmissions", &GetSubmissionsMsg{Type: TypeGetSubmissions}},
		{"GetSubmissionsResponse", &GetSubmissionsResponse{Type: TypeGetSubmissionsResponse}},

		// Base ref
		{"SetBaseRef", &SetBaseRefMsg{Type: TypeSetBaseRef, Ref: "main"}},
		{"SetBaseRefResponse", &SetBaseRefResponse{Type: TypeSetBaseRefResponse}},
		{"SetAutoAdvanceRef", &SetAutoAdvanceRefMsg{Type: TypeSetAutoAdvanceRef, Enabled: true}},
		{"SetAutoAdvanceRefResponse", &SetAutoAdvanceRefResponse{Type: TypeSetAutoAdvanceRefResponse}},
		{"IsAutoAdvanceRef", &IsAutoAdvanceRefMsg{Type: TypeIsAutoAdvanceRef}},
		{"IsAutoAdvanceRefResponse", &IsAutoAdvanceRefResponse{Type: TypeIsAutoAdvanceRefResponse, Enabled: true}},
		{"SelectedBaseRef", &SelectedBaseRefMsg{Type: TypeSelectedBaseRef}},
		{"SelectedBaseRefResponse", &SelectedBaseRefResponse{Type: TypeSelectedBaseRefResponse, Ref: "main"}},
		{"RecentCommits", &RecentCommitsMsg{Type: TypeRecentCommits, Count: 5}},
		{"RecentCommitsResponse", &RecentCommitsResponse{Type: TypeRecentCommitsResponse, Commits: []LogEntry{{Hash: "abc", Subject: "msg"}}}},

		// Snapshots
		{"GetSnapshots", &GetSnapshotsMsg{Type: TypeGetSnapshots}},
		{"GetSnapshotsResponse", &GetSnapshotsResponse{Type: TypeGetSnapshotsResponse}},
		{"SetSnapshotBase", &SetSnapshotBaseMsg{Type: TypeSetSnapshotBase, SnapshotID: 1}},
		{"SetSnapshotBaseResponse", &SetSnapshotBaseResponse{Type: TypeSetSnapshotBaseResponse}},
		{"ClearSnapshotBase", &ClearSnapshotBaseMsg{Type: TypeClearSnapshotBase}},
		{"ClearSnapshotBaseResponse", &ClearSnapshotBaseResponse{Type: TypeClearSnapshotBaseResponse}},
		{"GetActiveSnapshot", &GetActiveSnapshotMsg{Type: TypeGetActiveSnapshot}},
		{"GetActiveSnapshotResponse", &GetActiveSnapshotResponse{Type: TypeGetActiveSnapshotResponse}},
		{"HasSnapshots", &HasSnapshotsMsg{Type: TypeHasSnapshots}},
		{"HasSnapshotsResponse", &HasSnapshotsResponse{Type: TypeHasSnapshotsResponse, Has: true}},

		// Config
		{"GetConfig", &GetConfigMsg{Type: TypeGetConfig}},
		{"GetConfigResponse", &GetConfigResponse{Type: TypeGetConfigResponse, Config: &types.Config{}}},
		{"SaveConfig", &SaveConfigMsg{Type: TypeSaveConfig, Config: types.Config{Wrap: true}}},
		{"SaveConfigResponse", &SaveConfigResponse{Type: TypeSaveConfigResponse}},
		{"IsReviewTrackingEnabled", &IsReviewTrackingEnabledMsg{Type: TypeIsReviewTrackingEnabled}},
		{"IsReviewTrackingEnabledResponse", &IsReviewTrackingEnabledResponse{Type: TypeIsReviewTrackingEnabledResponse, Enabled: true}},

		// Status
		{"GetFeedbackStatus", &GetFeedbackStatusMsg{Type: TypeGetFeedbackStatus}},
		{"GetFeedbackStatusResponse", &GetFeedbackStatusResponse{Type: TypeGetFeedbackStatusResponse, Status: "pending"}},
		{"GetQueuedCount", &GetQueuedCountMsg{Type: TypeGetQueuedCount}},
		{"GetQueuedCountResponse", &GetQueuedCountResponse{Type: TypeGetQueuedCountResponse, Count: 3}},
		{"ReloadPendingFeedback", &ReloadPendingFeedbackMsg{Type: TypeReloadPendingFeedback}},
		{"ReloadPendingFeedbackResponse", &ReloadPendingFeedbackResponse{Type: TypeReloadPendingFeedbackResponse}},
		{"GetSubscriberCount", &GetSubscriberCountMsg{Type: TypeGetSubscriberCount}},
		{"GetSubscriberCountResponse", &GetSubscriberCountResponse{Type: TypeGetSubscriberCountResponse, Count: 1}},
		{"GetSocketPath", &GetSocketPathMsg{Type: TypeGetSocketPath}},
		{"GetSocketPathResponse", &GetSocketPathResponse{Type: TypeGetSocketPathResponse, Path: "/tmp/sock"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := Encode(tc.msg)
			if err != nil {
				t.Fatalf("encode: %v", err)
			}
			// Strip trailing newline appended by Encode
			decoded, err := Decode(data[:len(data)-1])
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			// Decoded must be pointer to same concrete type as input
			gotType := sprintType(decoded)
			wantType := sprintType(tc.msg)
			if gotType != wantType {
				t.Errorf("round-trip type mismatch: got %s, want %s", gotType, wantType)
			}
		})
	}
}

func sprintType(v any) string {
	return fmt.Sprintf("%s", reflect.TypeOf(v))
}

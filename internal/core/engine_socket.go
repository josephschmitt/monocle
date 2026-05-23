package core

import (
	"github.com/josephschmitt/monocle/internal/protocol"
)

// This file holds socket-message handlers for the EngineAPI surface.
// Each handler is a thin wrapper: it unpacks the protocol message,
// calls the existing engine method, and packs the response. The
// engine's business logic is untouched.

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// --- Sessions ---

func (e *Engine) handleStartSession(msg *protocol.StartSessionMsg) *protocol.StartSessionResponse {
	session, err := e.StartSession(SessionOptions{
		Agent:          msg.Agent,
		RepoRoot:       msg.RepoRoot,
		BaseRef:        msg.BaseRef,
		IgnorePatterns: msg.IgnorePatterns,
	})
	return &protocol.StartSessionResponse{
		Type:    protocol.TypeStartSessionResponse,
		Session: session,
		Error:   errString(err),
	}
}

func (e *Engine) handleResumeSession(msg *protocol.ResumeSessionMsg) *protocol.ResumeSessionResponse {
	session, err := e.ResumeSession(msg.SessionID)
	return &protocol.ResumeSessionResponse{
		Type:    protocol.TypeResumeSessionResponse,
		Session: session,
		Error:   errString(err),
	}
}

func (e *Engine) handleGetSession(_ *protocol.GetSessionMsg) *protocol.GetSessionResponse {
	return &protocol.GetSessionResponse{
		Type:    protocol.TypeGetSessionResponse,
		Session: e.GetSession(),
	}
}

func (e *Engine) handleListSessions(msg *protocol.ListSessionsMsg) *protocol.ListSessionsResponse {
	sessions, err := e.ListSessions(ListSessionsOptions{
		RepoRoot: msg.RepoRoot,
		Limit:    msg.Limit,
	})
	return &protocol.ListSessionsResponse{
		Type:     protocol.TypeListSessionsResponse,
		Sessions: sessions,
		Error:    errString(err),
	}
}

// --- Files ---

func (e *Engine) handleRefreshChangedFiles(_ *protocol.RefreshChangedFilesMsg) *protocol.RefreshChangedFilesResponse {
	files, err := e.RefreshChangedFiles()
	return &protocol.RefreshChangedFilesResponse{
		Type:  protocol.TypeRefreshChangedFilesResponse,
		Files: files,
		Error: errString(err),
	}
}

func (e *Engine) handleGetChangedFiles(_ *protocol.GetChangedFilesMsg) *protocol.GetChangedFilesResponse {
	return &protocol.GetChangedFilesResponse{
		Type:  protocol.TypeGetChangedFilesResponse,
		Files: e.GetChangedFiles(),
	}
}

func (e *Engine) handleGetFileDiff(msg *protocol.GetFileDiffMsg) *protocol.GetFileDiffResponse {
	diff, err := e.GetFileDiff(msg.Path)
	return &protocol.GetFileDiffResponse{
		Type:  protocol.TypeGetFileDiffResponse,
		Diff:  diff,
		Error: errString(err),
	}
}

func (e *Engine) handleGetFileContent(msg *protocol.GetFileContentMsg) *protocol.GetFileContentResponse {
	content, err := e.GetFileContent(msg.Path)
	return &protocol.GetFileContentResponse{
		Type:    protocol.TypeGetFileContentResponse,
		Content: content,
		Error:   errString(err),
	}
}

// --- Content ---

func (e *Engine) handleGetContentItems(_ *protocol.GetContentItemsMsg) *protocol.GetContentItemsResponse {
	return &protocol.GetContentItemsResponse{
		Type:  protocol.TypeGetContentItemsResponse,
		Items: e.GetContentItems(),
	}
}

func (e *Engine) handleGetContentItem(msg *protocol.GetContentItemMsg) *protocol.GetContentItemResponse {
	item, err := e.GetContentItem(msg.ID)
	return &protocol.GetContentItemResponse{
		Type:  protocol.TypeGetContentItemResponse,
		Item:  item,
		Error: errString(err),
	}
}

func (e *Engine) handleGetContentDiff(msg *protocol.GetContentDiffMsg) *protocol.GetContentDiffResponse {
	diff, err := e.GetContentDiff(msg.ID)
	return &protocol.GetContentDiffResponse{
		Type:  protocol.TypeGetContentDiffResponse,
		Diff:  diff,
		Error: errString(err),
	}
}

func (e *Engine) handleGetContentVersions(msg *protocol.GetContentVersionsMsg) *protocol.GetContentVersionsResponse {
	versions, err := e.GetContentVersions(msg.ID)
	return &protocol.GetContentVersionsResponse{
		Type:     protocol.TypeGetContentVersionsResponse,
		Versions: versions,
		Error:    errString(err),
	}
}

func (e *Engine) handleGetContentDiffBetweenVersions(msg *protocol.GetContentDiffBetweenVersionsMsg) *protocol.GetContentDiffBetweenVersionsResponse {
	diff, err := e.GetContentDiffBetweenVersions(msg.ID, msg.FromVersion, msg.ToVersion)
	return &protocol.GetContentDiffBetweenVersionsResponse{
		Type:  protocol.TypeGetContentDiffBetweenVersionResponse,
		Diff:  diff,
		Error: errString(err),
	}
}

func (e *Engine) handleDismissArtifact(msg *protocol.DismissArtifactMsg) *protocol.DismissArtifactResponse {
	err := e.DismissArtifact(msg.ID)
	return &protocol.DismissArtifactResponse{
		Type:  protocol.TypeDismissArtifactResponse,
		Error: errString(err),
	}
}

// --- Additional files ---

func (e *Engine) handleGetAdditionalFiles(_ *protocol.GetAdditionalFilesMsg) *protocol.GetAdditionalFilesResponse {
	return &protocol.GetAdditionalFilesResponse{
		Type:  protocol.TypeGetAdditionalFilesResponse,
		Files: e.GetAdditionalFiles(),
	}
}

func (e *Engine) handleGetAdditionalFileContent(msg *protocol.GetAdditionalFileContentMsg) *protocol.GetAdditionalFileContentResponse {
	content, err := e.GetAdditionalFileContent(msg.AbsPath)
	return &protocol.GetAdditionalFileContentResponse{
		Type:    protocol.TypeGetAdditionalFileContentResponse,
		Content: content,
		Error:   errString(err),
	}
}

// --- Comments ---

func (e *Engine) handleAddComment(msg *protocol.AddCommentMsg) *protocol.AddCommentResponse {
	comment, err := e.AddComment(
		CommentTarget{
			TargetType: msg.TargetType,
			TargetRef:  msg.TargetRef,
			LineStart:  msg.LineStart,
			LineEnd:    msg.LineEnd,
		},
		msg.CommentType,
		msg.Body,
	)
	return &protocol.AddCommentResponse{
		Type:    protocol.TypeAddCommentResponse,
		Comment: comment,
		Error:   errString(err),
	}
}

func (e *Engine) handleEditComment(msg *protocol.EditCommentMsg) *protocol.EditCommentResponse {
	comment, err := e.EditComment(msg.CommentID, msg.CommentType, msg.Body)
	return &protocol.EditCommentResponse{
		Type:    protocol.TypeEditCommentResponse,
		Comment: comment,
		Error:   errString(err),
	}
}

func (e *Engine) handleDeleteComment(msg *protocol.DeleteCommentMsg) *protocol.DeleteCommentResponse {
	err := e.DeleteComment(msg.CommentID)
	return &protocol.DeleteCommentResponse{
		Type:  protocol.TypeDeleteCommentResponse,
		Error: errString(err),
	}
}

func (e *Engine) handleResolveComment(msg *protocol.ResolveCommentMsg) *protocol.ResolveCommentResponse {
	err := e.ResolveComment(msg.CommentID)
	return &protocol.ResolveCommentResponse{
		Type:  protocol.TypeResolveCommentResponse,
		Error: errString(err),
	}
}

func (e *Engine) handleClearComments(_ *protocol.ClearCommentsMsg) *protocol.ClearCommentsResponse {
	err := e.ClearComments()
	return &protocol.ClearCommentsResponse{
		Type:  protocol.TypeClearCommentsResponse,
		Error: errString(err),
	}
}

func (e *Engine) handleClearReview(_ *protocol.ClearReviewMsg) *protocol.ClearReviewResponse {
	err := e.ClearReview()
	return &protocol.ClearReviewResponse{
		Type:  protocol.TypeClearReviewResponse,
		Error: errString(err),
	}
}

// --- Marking ---

func (e *Engine) handleMarkReviewed(msg *protocol.MarkReviewedMsg) *protocol.MarkReviewedResponse {
	err := e.MarkReviewed(msg.Path)
	return &protocol.MarkReviewedResponse{
		Type:  protocol.TypeMarkReviewedResponse,
		Error: errString(err),
	}
}

func (e *Engine) handleUnmarkReviewed(msg *protocol.UnmarkReviewedMsg) *protocol.UnmarkReviewedResponse {
	err := e.UnmarkReviewed(msg.Path)
	return &protocol.UnmarkReviewedResponse{
		Type:  protocol.TypeUnmarkReviewedResponse,
		Error: errString(err),
	}
}

func (e *Engine) handleMarkContentReviewed(msg *protocol.MarkContentReviewedMsg) *protocol.MarkContentReviewedResponse {
	err := e.MarkContentReviewed(msg.ID)
	return &protocol.MarkContentReviewedResponse{
		Type:  protocol.TypeMarkContentReviewedResponse,
		Error: errString(err),
	}
}

func (e *Engine) handleUnmarkContentReviewed(msg *protocol.UnmarkContentReviewedMsg) *protocol.UnmarkContentReviewedResponse {
	err := e.UnmarkContentReviewed(msg.ID)
	return &protocol.UnmarkContentReviewedResponse{
		Type:  protocol.TypeUnmarkContentReviewedResponse,
		Error: errString(err),
	}
}

func (e *Engine) handleResetAllReviewed(_ *protocol.ResetAllReviewedMsg) *protocol.ResetAllReviewedResponse {
	err := e.ResetAllReviewed()
	return &protocol.ResetAllReviewedResponse{
		Type:  protocol.TypeResetAllReviewedResponse,
		Error: errString(err),
	}
}

func (e *Engine) handleMarkAllReviewed(_ *protocol.MarkAllReviewedMsg) *protocol.MarkAllReviewedResponse {
	err := e.MarkAllReviewed()
	return &protocol.MarkAllReviewedResponse{
		Type:  protocol.TypeMarkAllReviewedResponse,
		Error: errString(err),
	}
}

// --- Submission ---

func (e *Engine) handleGetReviewSummary(_ *protocol.GetReviewSummaryMsg) *protocol.GetReviewSummaryResponse {
	summary, err := e.GetReviewSummary()
	return &protocol.GetReviewSummaryResponse{
		Type:    protocol.TypeGetReviewSummaryResponse,
		Summary: summary,
		Error:   errString(err),
	}
}

func (e *Engine) handleSubmit(msg *protocol.SubmitMsg) *protocol.SubmitResponse {
	result, err := e.Submit(msg.Action, msg.Body)
	resp := &protocol.SubmitResponse{
		Type:  protocol.TypeSubmitResponse,
		Error: errString(err),
	}
	if result != nil {
		resp.AgentConnected = result.AgentConnected
	}
	return resp
}

func (e *Engine) handleFormatReview(msg *protocol.FormatReviewMsg) *protocol.FormatReviewResponse {
	formatted, err := e.FormatReview(msg.Action, msg.Body)
	return &protocol.FormatReviewResponse{
		Type:      protocol.TypeFormatReviewResponse,
		Formatted: formatted,
		Error:     errString(err),
	}
}

func (e *Engine) handleGetSubmissions(_ *protocol.GetSubmissionsMsg) *protocol.GetSubmissionsResponse {
	submissions, err := e.GetSubmissions()
	return &protocol.GetSubmissionsResponse{
		Type:        protocol.TypeGetSubmissionsResponse,
		Submissions: submissions,
		Error:       errString(err),
	}
}

// --- Base ref ---

func (e *Engine) handleSetBaseRef(msg *protocol.SetBaseRefMsg) *protocol.SetBaseRefResponse {
	err := e.SetBaseRef(msg.Ref)
	return &protocol.SetBaseRefResponse{
		Type:  protocol.TypeSetBaseRefResponse,
		Error: errString(err),
	}
}

func (e *Engine) handleSetAutoAdvanceRef(msg *protocol.SetAutoAdvanceRefMsg) *protocol.SetAutoAdvanceRefResponse {
	e.SetAutoAdvanceRef(msg.Enabled)
	return &protocol.SetAutoAdvanceRefResponse{
		Type: protocol.TypeSetAutoAdvanceRefResponse,
	}
}

func (e *Engine) handleIsAutoAdvanceRef(_ *protocol.IsAutoAdvanceRefMsg) *protocol.IsAutoAdvanceRefResponse {
	return &protocol.IsAutoAdvanceRefResponse{
		Type:    protocol.TypeIsAutoAdvanceRefResponse,
		Enabled: e.IsAutoAdvanceRef(),
	}
}

func (e *Engine) handleSelectedBaseRef(_ *protocol.SelectedBaseRefMsg) *protocol.SelectedBaseRefResponse {
	return &protocol.SelectedBaseRefResponse{
		Type: protocol.TypeSelectedBaseRefResponse,
		Ref:  e.SelectedBaseRef(),
	}
}

func (e *Engine) handleRecentCommits(msg *protocol.RecentCommitsMsg) *protocol.RecentCommitsResponse {
	entries, err := e.RecentCommits(msg.Count)
	out := make([]protocol.LogEntry, len(entries))
	for i, entry := range entries {
		out[i] = protocol.LogEntry{Hash: entry.Hash, Subject: entry.Subject}
	}
	return &protocol.RecentCommitsResponse{
		Type:    protocol.TypeRecentCommitsResponse,
		Commits: out,
		Error:   errString(err),
	}
}

// --- Snapshots ---

func (e *Engine) handleGetSnapshots(_ *protocol.GetSnapshotsMsg) *protocol.GetSnapshotsResponse {
	snaps, err := e.GetSnapshots()
	return &protocol.GetSnapshotsResponse{
		Type:      protocol.TypeGetSnapshotsResponse,
		Snapshots: snaps,
		Error:     errString(err),
	}
}

func (e *Engine) handleSetSnapshotBase(msg *protocol.SetSnapshotBaseMsg) *protocol.SetSnapshotBaseResponse {
	err := e.SetSnapshotBase(msg.SnapshotID)
	return &protocol.SetSnapshotBaseResponse{
		Type:  protocol.TypeSetSnapshotBaseResponse,
		Error: errString(err),
	}
}

func (e *Engine) handleClearSnapshotBase(_ *protocol.ClearSnapshotBaseMsg) *protocol.ClearSnapshotBaseResponse {
	e.ClearSnapshotBase()
	return &protocol.ClearSnapshotBaseResponse{
		Type: protocol.TypeClearSnapshotBaseResponse,
	}
}

func (e *Engine) handleGetActiveSnapshot(_ *protocol.GetActiveSnapshotMsg) *protocol.GetActiveSnapshotResponse {
	return &protocol.GetActiveSnapshotResponse{
		Type:     protocol.TypeGetActiveSnapshotResponse,
		Snapshot: e.GetActiveSnapshot(),
	}
}

func (e *Engine) handleHasSnapshots(_ *protocol.HasSnapshotsMsg) *protocol.HasSnapshotsResponse {
	has, err := e.HasSnapshots()
	return &protocol.HasSnapshotsResponse{
		Type:  protocol.TypeHasSnapshotsResponse,
		Has:   has,
		Error: errString(err),
	}
}

// --- Config ---

func (e *Engine) handleGetConfig(_ *protocol.GetConfigMsg) *protocol.GetConfigResponse {
	return &protocol.GetConfigResponse{
		Type:   protocol.TypeGetConfigResponse,
		Config: e.GetConfig(),
	}
}

// handleSaveConfig replaces the engine's config with the value from the
// client and persists it. The in-place copy preserves pointer identity so
// any in-process subscribers holding e.cfg keep seeing fresh fields.
func (e *Engine) handleSaveConfig(msg *protocol.SaveConfigMsg) *protocol.SaveConfigResponse {
	e.mu.Lock()
	*e.cfg = msg.Config
	e.mu.Unlock()
	err := e.SaveConfig()
	return &protocol.SaveConfigResponse{
		Type:  protocol.TypeSaveConfigResponse,
		Error: errString(err),
	}
}

func (e *Engine) handleIsReviewTrackingEnabled(_ *protocol.IsReviewTrackingEnabledMsg) *protocol.IsReviewTrackingEnabledResponse {
	return &protocol.IsReviewTrackingEnabledResponse{
		Type:    protocol.TypeIsReviewTrackingEnabledResponse,
		Enabled: e.IsReviewTrackingEnabled(),
	}
}

// --- Status ---

func (e *Engine) handleGetFeedbackStatus(_ *protocol.GetFeedbackStatusMsg) *protocol.GetFeedbackStatusResponse {
	return &protocol.GetFeedbackStatusResponse{
		Type:   protocol.TypeGetFeedbackStatusResponse,
		Status: e.GetFeedbackStatus(),
	}
}

func (e *Engine) handleGetQueuedCount(_ *protocol.GetQueuedCountMsg) *protocol.GetQueuedCountResponse {
	return &protocol.GetQueuedCountResponse{
		Type:  protocol.TypeGetQueuedCountResponse,
		Count: e.GetQueuedCount(),
	}
}

func (e *Engine) handleReloadPendingFeedback(_ *protocol.ReloadPendingFeedbackMsg) *protocol.ReloadPendingFeedbackResponse {
	e.ReloadPendingFeedback()
	return &protocol.ReloadPendingFeedbackResponse{
		Type: protocol.TypeReloadPendingFeedbackResponse,
	}
}

func (e *Engine) handleGetSubscriberCount(_ *protocol.GetSubscriberCountMsg) *protocol.GetSubscriberCountResponse {
	return &protocol.GetSubscriberCountResponse{
		Type:  protocol.TypeGetSubscriberCountResponse,
		Count: e.GetSubscriberCount(),
	}
}

func (e *Engine) handleGetSocketPath(_ *protocol.GetSocketPathMsg) *protocol.GetSocketPathResponse {
	return &protocol.GetSocketPathResponse{
		Type: protocol.TypeGetSocketPathResponse,
		Path: e.GetSocketPath(),
	}
}

// handleSetPause routes the TUI's pause keybind through the daemon so the
// real pause flag flips on the engine that the agent's `monocle review
// status` and `get-feedback --wait` consult. Pre-fix, the client's
// RequestPause/CancelPause were no-op stubs and the agent never saw
// pause_requested.
func (e *Engine) handleSetPause(msg *protocol.SetPauseMsg) *protocol.SetPauseResponse {
	if msg.Requested {
		e.RequestPause()
	} else {
		e.CancelPause()
	}
	return &protocol.SetPauseResponse{Type: protocol.TypeSetPauseResponse}
}

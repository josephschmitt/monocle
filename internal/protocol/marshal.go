package protocol

import (
	"encoding/json"
	"fmt"
)

// Encode marshals a message to a JSON line (with trailing newline).
func Encode(msg any) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("protocol encode: %w", err)
	}
	return append(data, '\n'), nil
}

// Decode unmarshals a JSON line, using the "type" field to discriminate.
func Decode(data []byte) (any, error) {
	var envelope struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, fmt.Errorf("protocol decode envelope: %w", err)
	}

	var msg any
	switch envelope.Type {
	case TypeGetReviewStatus:
		msg = &GetReviewStatusMsg{}
	case TypePollFeedback:
		msg = &PollFeedbackMsg{}
	case TypeSubmitContent:
		msg = &SubmitContentMsg{}
	case TypeSubscribe:
		msg = &SubscribeMsg{}
	case TypeConnect:
		msg = &ConnectMsg{}
	case TypeIdentify:
		msg = &IdentifyMsg{}
	case TypeAddAdditionalFiles:
		msg = &AddAdditionalFilesMsg{}
	case TypeMarkActivity:
		msg = &MarkActivityMsg{}
	case TypeAwaitReview:
		msg = &AwaitReviewMsg{}
	case TypeGetReviewStatusResponse:
		msg = &GetReviewStatusResponse{}
	case TypePollFeedbackResponse:
		msg = &PollFeedbackResponse{}
	case TypeSubmitContentResponse:
		msg = &SubmitContentResponse{}
	case TypeSubscribeResponse:
		msg = &SubscribeResponse{}
	case TypeConnectResponse:
		msg = &ConnectResponse{}
	case TypeEventNotification:
		msg = &EventNotification{}
	case TypeAddAdditionalFilesResponse:
		msg = &AddAdditionalFilesResponse{}
	case TypeMarkActivityResponse:
		msg = &MarkActivityResponse{}
	case TypeAwaitReviewResponse:
		msg = &AwaitReviewResponse{}

	// --- Engine surface: inbound ---
	case TypeStartSession:
		msg = &StartSessionMsg{}
	case TypeResumeSession:
		msg = &ResumeSessionMsg{}
	case TypeGetSession:
		msg = &GetSessionMsg{}
	case TypeListSessions:
		msg = &ListSessionsMsg{}
	case TypeRefreshChangedFiles:
		msg = &RefreshChangedFilesMsg{}
	case TypeGetChangedFiles:
		msg = &GetChangedFilesMsg{}
	case TypeGetFileDiff:
		msg = &GetFileDiffMsg{}
	case TypeGetFileContent:
		msg = &GetFileContentMsg{}
	case TypeGetContentItems:
		msg = &GetContentItemsMsg{}
	case TypeGetContentItem:
		msg = &GetContentItemMsg{}
	case TypeGetContentDiff:
		msg = &GetContentDiffMsg{}
	case TypeGetContentVersions:
		msg = &GetContentVersionsMsg{}
	case TypeGetContentDiffBetweenVersion:
		msg = &GetContentDiffBetweenVersionsMsg{}
	case TypeDismissArtifact:
		msg = &DismissArtifactMsg{}
	case TypeGetAdditionalFiles:
		msg = &GetAdditionalFilesMsg{}
	case TypeGetAdditionalFileContent:
		msg = &GetAdditionalFileContentMsg{}
	case TypeAddComment:
		msg = &AddCommentMsg{}
	case TypeEditComment:
		msg = &EditCommentMsg{}
	case TypeDeleteComment:
		msg = &DeleteCommentMsg{}
	case TypeResolveComment:
		msg = &ResolveCommentMsg{}
	case TypeClearComments:
		msg = &ClearCommentsMsg{}
	case TypeClearReview:
		msg = &ClearReviewMsg{}
	case TypeMarkReviewed:
		msg = &MarkReviewedMsg{}
	case TypeUnmarkReviewed:
		msg = &UnmarkReviewedMsg{}
	case TypeMarkContentReviewed:
		msg = &MarkContentReviewedMsg{}
	case TypeUnmarkContentReviewed:
		msg = &UnmarkContentReviewedMsg{}
	case TypeResetAllReviewed:
		msg = &ResetAllReviewedMsg{}
	case TypeMarkAllReviewed:
		msg = &MarkAllReviewedMsg{}
	case TypeGetReviewSummary:
		msg = &GetReviewSummaryMsg{}
	case TypeSubmit:
		msg = &SubmitMsg{}
	case TypeFormatReview:
		msg = &FormatReviewMsg{}
	case TypeGetSubmissions:
		msg = &GetSubmissionsMsg{}
	case TypeSetBaseRef:
		msg = &SetBaseRefMsg{}
	case TypeSetAutoAdvanceRef:
		msg = &SetAutoAdvanceRefMsg{}
	case TypeIsAutoAdvanceRef:
		msg = &IsAutoAdvanceRefMsg{}
	case TypeSelectedBaseRef:
		msg = &SelectedBaseRefMsg{}
	case TypeRecentCommits:
		msg = &RecentCommitsMsg{}
	case TypeGetSnapshots:
		msg = &GetSnapshotsMsg{}
	case TypeSetSnapshotBase:
		msg = &SetSnapshotBaseMsg{}
	case TypeClearSnapshotBase:
		msg = &ClearSnapshotBaseMsg{}
	case TypeGetActiveSnapshot:
		msg = &GetActiveSnapshotMsg{}
	case TypeHasSnapshots:
		msg = &HasSnapshotsMsg{}
	case TypeGetConfig:
		msg = &GetConfigMsg{}
	case TypeSaveConfig:
		msg = &SaveConfigMsg{}
	case TypeIsReviewTrackingEnabled:
		msg = &IsReviewTrackingEnabledMsg{}
	case TypeGetFeedbackStatus:
		msg = &GetFeedbackStatusMsg{}
	case TypeGetQueuedCount:
		msg = &GetQueuedCountMsg{}
	case TypeReloadPendingFeedback:
		msg = &ReloadPendingFeedbackMsg{}
	case TypeGetSubscriberCount:
		msg = &GetSubscriberCountMsg{}
	case TypeGetSocketPath:
		msg = &GetSocketPathMsg{}
	case TypeSetPause:
		msg = &SetPauseMsg{}

	// --- Engine surface: outbound ---
	case TypeStartSessionResponse:
		msg = &StartSessionResponse{}
	case TypeResumeSessionResponse:
		msg = &ResumeSessionResponse{}
	case TypeGetSessionResponse:
		msg = &GetSessionResponse{}
	case TypeListSessionsResponse:
		msg = &ListSessionsResponse{}
	case TypeRefreshChangedFilesResponse:
		msg = &RefreshChangedFilesResponse{}
	case TypeGetChangedFilesResponse:
		msg = &GetChangedFilesResponse{}
	case TypeGetFileDiffResponse:
		msg = &GetFileDiffResponse{}
	case TypeGetFileContentResponse:
		msg = &GetFileContentResponse{}
	case TypeGetContentItemsResponse:
		msg = &GetContentItemsResponse{}
	case TypeGetContentItemResponse:
		msg = &GetContentItemResponse{}
	case TypeGetContentDiffResponse:
		msg = &GetContentDiffResponse{}
	case TypeGetContentVersionsResponse:
		msg = &GetContentVersionsResponse{}
	case TypeGetContentDiffBetweenVersionResponse:
		msg = &GetContentDiffBetweenVersionsResponse{}
	case TypeDismissArtifactResponse:
		msg = &DismissArtifactResponse{}
	case TypeGetAdditionalFilesResponse:
		msg = &GetAdditionalFilesResponse{}
	case TypeGetAdditionalFileContentResponse:
		msg = &GetAdditionalFileContentResponse{}
	case TypeAddCommentResponse:
		msg = &AddCommentResponse{}
	case TypeEditCommentResponse:
		msg = &EditCommentResponse{}
	case TypeDeleteCommentResponse:
		msg = &DeleteCommentResponse{}
	case TypeResolveCommentResponse:
		msg = &ResolveCommentResponse{}
	case TypeClearCommentsResponse:
		msg = &ClearCommentsResponse{}
	case TypeClearReviewResponse:
		msg = &ClearReviewResponse{}
	case TypeMarkReviewedResponse:
		msg = &MarkReviewedResponse{}
	case TypeUnmarkReviewedResponse:
		msg = &UnmarkReviewedResponse{}
	case TypeMarkContentReviewedResponse:
		msg = &MarkContentReviewedResponse{}
	case TypeUnmarkContentReviewedResponse:
		msg = &UnmarkContentReviewedResponse{}
	case TypeResetAllReviewedResponse:
		msg = &ResetAllReviewedResponse{}
	case TypeMarkAllReviewedResponse:
		msg = &MarkAllReviewedResponse{}
	case TypeGetReviewSummaryResponse:
		msg = &GetReviewSummaryResponse{}
	case TypeSubmitResponse:
		msg = &SubmitResponse{}
	case TypeFormatReviewResponse:
		msg = &FormatReviewResponse{}
	case TypeGetSubmissionsResponse:
		msg = &GetSubmissionsResponse{}
	case TypeSetBaseRefResponse:
		msg = &SetBaseRefResponse{}
	case TypeSetAutoAdvanceRefResponse:
		msg = &SetAutoAdvanceRefResponse{}
	case TypeIsAutoAdvanceRefResponse:
		msg = &IsAutoAdvanceRefResponse{}
	case TypeSelectedBaseRefResponse:
		msg = &SelectedBaseRefResponse{}
	case TypeRecentCommitsResponse:
		msg = &RecentCommitsResponse{}
	case TypeGetSnapshotsResponse:
		msg = &GetSnapshotsResponse{}
	case TypeSetSnapshotBaseResponse:
		msg = &SetSnapshotBaseResponse{}
	case TypeClearSnapshotBaseResponse:
		msg = &ClearSnapshotBaseResponse{}
	case TypeGetActiveSnapshotResponse:
		msg = &GetActiveSnapshotResponse{}
	case TypeHasSnapshotsResponse:
		msg = &HasSnapshotsResponse{}
	case TypeGetConfigResponse:
		msg = &GetConfigResponse{}
	case TypeSaveConfigResponse:
		msg = &SaveConfigResponse{}
	case TypeIsReviewTrackingEnabledResponse:
		msg = &IsReviewTrackingEnabledResponse{}
	case TypeGetFeedbackStatusResponse:
		msg = &GetFeedbackStatusResponse{}
	case TypeGetQueuedCountResponse:
		msg = &GetQueuedCountResponse{}
	case TypeReloadPendingFeedbackResponse:
		msg = &ReloadPendingFeedbackResponse{}
	case TypeGetSubscriberCountResponse:
		msg = &GetSubscriberCountResponse{}
	case TypeGetSocketPathResponse:
		msg = &GetSocketPathResponse{}
	case TypeSetPauseResponse:
		msg = &SetPauseResponse{}

	default:
		return nil, fmt.Errorf("protocol decode: unknown type %q", envelope.Type)
	}

	if err := json.Unmarshal(data, msg); err != nil {
		return nil, fmt.Errorf("protocol decode %s: %w", envelope.Type, err)
	}
	return msg, nil
}

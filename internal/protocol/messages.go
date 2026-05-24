package protocol

import "github.com/josephschmitt/monocle/internal/types"

// Inbound message types (from CLI subcommands to engine via socket)
const (
	TypeGetReviewStatus    = "get_review_status"
	TypePollFeedback       = "poll_feedback"
	TypeSubmitContent      = "submit_content"
	TypeSubscribe          = "subscribe"
	TypeConnect            = "connect"
	TypeIdentify           = "identify"
	TypeAddAdditionalFiles = "add_additional_files"
	TypeMarkActivity       = "mark_activity"
	TypeAwaitReview        = "await_review"
)

// Outbound message types (from engine to CLI subcommands)
const (
	TypeGetReviewStatusResponse    = "get_review_status_response"
	TypePollFeedbackResponse       = "poll_feedback_response"
	TypeSubmitContentResponse      = "submit_content_response"
	TypeSubscribeResponse          = "subscribe_response"
	TypeConnectResponse            = "connect_response"
	TypeEventNotification          = "event_notification"
	TypeAddAdditionalFilesResponse = "add_additional_files_response"
	TypeMarkActivityResponse       = "mark_activity_response"
	TypeAwaitReviewResponse        = "await_review_response"
)

// GetReviewStatusMsg requests the current review state from the engine.
type GetReviewStatusMsg struct {
	Type string `json:"type"`
}

// GetReviewStatusResponse returns the current review state.
type GetReviewStatusResponse struct {
	Type         string `json:"type"`
	Status       string `json:"status"` // "no_feedback" | "pending" | "pause_requested"
	CommentCount int    `json:"comment_count,omitempty"`
	Summary      string `json:"summary,omitempty"`
}

// PollFeedbackMsg requests pending feedback, optionally blocking until available.
type PollFeedbackMsg struct {
	Type string `json:"type"`
	Wait bool   `json:"wait"`
}

// PollFeedbackResponse returns feedback if available.
type PollFeedbackResponse struct {
	Type         string `json:"type"`
	HasFeedback  bool   `json:"has_feedback"`
	Feedback     string `json:"feedback,omitempty"`
	CommentCount int    `json:"comment_count,omitempty"`
	Action       string `json:"action,omitempty"` // "approve" | "request_changes"
}

// SubmitContentMsg sends reviewable content (plans, docs) from the agent.
type SubmitContentMsg struct {
	Type        string `json:"type"`
	ID          string `json:"id"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	ContentType string `json:"content_type,omitempty"`
	IsPlan      bool   `json:"is_plan,omitempty"`
}

// SubmitContentResponse acknowledges content submission.
// ID echoes the stored content's id — the daemon mints a UUID when the
// caller sends an empty ID, so this field lets the caller address the
// just-submitted item (mark reviewed, dismiss, fetch versions).
type SubmitContentResponse struct {
	Type    string `json:"type"`
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	ID      string `json:"id,omitempty"`
}

// SubscribeMsg requests a persistent event subscription on this connection.
//
// Passive=true marks the connection as a viewer (e.g. the TUI) that should
// NOT be counted as an attached agent. The server forwards events but skips
// the subscriberCount bookkeeping that Submit() uses to pick push vs queue
// delivery, and suppresses the EventConnectionChanged broadcast that
// otherwise tells the UI "an agent is connected". The zero value (false)
// preserves the existing push-subscriber semantics for backwards
// compatibility with any agent that sends SubscribeMsg directly.
type SubscribeMsg struct {
	Type    string   `json:"type"`
	Events  []string `json:"events"`
	Passive bool     `json:"passive,omitempty"`
}

// SubscribeResponse acknowledges a subscription request.
// ProtocolVersion lets clients detect a stale daemon that predates wire
// features they rely on (e.g. Passive subscribers). A zero value means
// the daemon is older than ProtocolVersion=1 and must be respawned for
// the client to trust newer behaviour.
type SubscribeResponse struct {
	Type            string `json:"type"`
	Success         bool   `json:"success"`
	ProtocolVersion int    `json:"protocol_version,omitempty"`
}

// CurrentProtocolVersion is bumped when a wire-protocol change requires
// new client behaviour. Bump this whenever a new client semantically
// depends on a daemon feature that an older daemon would silently drop.
//
//	1 — initial versioned protocol (Passive subscribers, ContentAdded,
//	    presence-flag responses).
const CurrentProtocolVersion = 1

// EventNotification pushes an engine event to a subscribed connection.
type EventNotification struct {
	Type    string         `json:"type"`
	Event   string         `json:"event"`
	Payload map[string]any `json:"payload"`
}

// ConnectMsg requests a persistent connection with optional event forwarding
// but without becoming a push subscriber. The connection supports request/response
// for tool calls and receives event notifications, but does not increment
// subscriberCount (so Submit() always queues feedback for pull delivery).
type ConnectMsg struct {
	Type   string   `json:"type"`
	Events []string `json:"events,omitempty"`
}

// ConnectResponse acknowledges a connect request.
type ConnectResponse struct {
	Type    string `json:"type"`
	Success bool   `json:"success"`
}

// IdentifyMsg carries the agent's self-reported name (sent after MCP handshake).
type IdentifyMsg struct {
	Type  string `json:"type"`
	Agent string `json:"agent"`
}

// AddAdditionalFilesMsg sends file/directory paths to add for review.
type AddAdditionalFilesMsg struct {
	Type  string   `json:"type"`
	Paths []string `json:"paths"`
}

// AddAdditionalFilesResponse acknowledges additional files submission.
// Added carries the newly-attached files (not the cumulative list) so
// callers can distinguish a fresh add from a no-op de-dup.
//
// AddedPresent disambiguates "new daemon returned an empty Added list"
// (genuinely added zero files) from "old daemon doesn't populate Added at
// all". Without this flag, a JSON `omitempty` on an empty slice is
// indistinguishable on the wire from the field being absent, so the
// client cannot tell whether to trust Added or fall back to a cumulative
// fetch. New daemons always set AddedPresent=true; old daemons leave it
// false.
type AddAdditionalFilesResponse struct {
	Type         string                 `json:"type"`
	Success      bool                   `json:"success"`
	Message      string                 `json:"message,omitempty"`
	Count        int                    `json:"count"`
	Added        []types.AdditionalFile `json:"added,omitempty"`
	AddedPresent bool                   `json:"added_present,omitempty"`
}

// MarkActivityMsg notifies the engine that a write-tool just fired in the
// current session, marking the session as having unreviewed changes. The
// Stop-hook's AwaitReview call consults this flag to decide whether to
// block the turn or let it end normally.
type MarkActivityMsg struct {
	Type string `json:"type"`
}

// MarkActivityResponse acknowledges an activity mark.
type MarkActivityResponse struct {
	Type    string `json:"type"`
	Success bool   `json:"success"`
}

// AwaitReviewMsg is issued by the Stop hook at turn-end. If the session
// has unreviewed activity (a write-tool fired during the turn), the engine
// blocks until the reviewer submits feedback. Otherwise it returns
// immediately with HasActivity=false so the agent's turn can end cleanly.
type AwaitReviewMsg struct {
	Type string `json:"type"`
	Wait bool   `json:"wait"` // true = block on reviewer; false = snapshot query
}

// AwaitReviewResponse reports the outcome of an AwaitReview call.
// When HasActivity is false the turn may end normally; when true with
// Action="approve" the turn ends after the reviewer saw the diff; when
// true with Action="request_changes" the hook converts the feedback into
// a Stop-hook block decision that sends Claude back to work.
type AwaitReviewResponse struct {
	Type        string `json:"type"`
	HasActivity bool   `json:"has_activity"`
	Action      string `json:"action,omitempty"`
	Feedback    string `json:"feedback,omitempty"`
}

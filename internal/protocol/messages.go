package protocol

// Inbound message types (from hook shim to engine)
const (
	TypePostToolUse   = "post_tool_use"
	TypeStop          = "stop"
	TypePromptSubmit  = "prompt_submit"
	TypeContentReview = "content_review"
)

// Outbound message types (from engine to hook shim)
const (
	TypeStopResponse         = "stop_response"
	TypePromptSubmitResponse = "prompt_submit_response"
)

type PostToolUseMsg struct {
	Type       string `json:"type"`
	Agent      string `json:"agent"`
	Tool       string `json:"tool"`
	FilePath   string `json:"file_path,omitempty"`
	ToolInput  string `json:"tool_input,omitempty"`
	ToolOutput string `json:"tool_output,omitempty"`
}

type StopMsg struct {
	Type       string `json:"type"`
	Agent      string `json:"agent"`
	StopReason string `json:"stop_reason,omitempty"`
	RequestID  string `json:"request_id"`

	// Content to present for review when the agent stops (e.g., plans).
	// Adapters populate these when the agent has reviewable non-file content.
	ReviewContent      string `json:"review_content,omitempty"`
	ReviewContentTitle string `json:"review_content_title,omitempty"`
	ReviewContentType  string `json:"review_content_type,omitempty"`
}

type PromptSubmitMsg struct {
	Type      string `json:"type"`
	Agent     string `json:"agent"`
	Prompt    string `json:"prompt,omitempty"`
	RequestID string `json:"request_id"`
}

type ContentReviewMsg struct {
	Type        string `json:"type"`
	ID          string `json:"id"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	ContentType string `json:"content_type,omitempty"`
	RequestID   string `json:"request_id,omitempty"`
}

type StopResponse struct {
	Type          string `json:"type"`
	RequestID     string `json:"request_id"`
	Continue      bool   `json:"continue"`
	SystemMessage string `json:"system_message,omitempty"`
}

type PromptSubmitResponse struct {
	Type              string `json:"type"`
	RequestID         string `json:"request_id"`
	AdditionalContext string `json:"additional_context,omitempty"`
}

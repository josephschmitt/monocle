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
	case TypePostToolUse:
		msg = &PostToolUseMsg{}
	case TypeStop:
		msg = &StopMsg{}
	case TypePromptSubmit:
		msg = &PromptSubmitMsg{}
	case TypeContentReview:
		msg = &ContentReviewMsg{}
	case TypeStopResponse:
		msg = &StopResponse{}
	case TypePromptSubmitResponse:
		msg = &PromptSubmitResponse{}
	default:
		return nil, fmt.Errorf("protocol decode: unknown type %q", envelope.Type)
	}

	if err := json.Unmarshal(data, msg); err != nil {
		return nil, fmt.Errorf("protocol decode %s: %w", envelope.Type, err)
	}
	return msg, nil
}

// IsBlocking returns true for message types that require a response.
func IsBlocking(msg any) bool {
	switch msg.(type) {
	case *StopMsg, *PromptSubmitMsg:
		return true
	default:
		return false
	}
}

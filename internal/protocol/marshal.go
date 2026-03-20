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
	case TypeGetReviewStatusResponse:
		msg = &GetReviewStatusResponse{}
	case TypePollFeedbackResponse:
		msg = &PollFeedbackResponse{}
	case TypeSubmitContentResponse:
		msg = &SubmitContentResponse{}
	case TypeSubscribeResponse:
		msg = &SubscribeResponse{}
	case TypeEventNotification:
		msg = &EventNotification{}
	default:
		return nil, fmt.Errorf("protocol decode: unknown type %q", envelope.Type)
	}

	if err := json.Unmarshal(data, msg); err != nil {
		return nil, fmt.Errorf("protocol decode %s: %w", envelope.Type, err)
	}
	return msg, nil
}

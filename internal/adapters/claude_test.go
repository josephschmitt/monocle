package adapters

import (
	"testing"

	"github.com/anthropics/monocle/internal/protocol"
)

func TestClaudeParseStop(t *testing.T) {
	adapter := &ClaudeAdapter{}

	t.Run("basic stop", func(t *testing.T) {
		raw := []byte(`{"request_id":"r1","stop_reason":"end_turn"}`)
		msg, err := adapter.ParseHookInput("stop", raw)
		if err != nil {
			t.Fatal(err)
		}
		stop := msg.(*protocol.StopMsg)
		if stop.RequestID != "r1" {
			t.Errorf("expected request_id r1, got %q", stop.RequestID)
		}
		if stop.ReviewContent != "" {
			t.Error("expected no review content for basic stop")
		}
	})

	t.Run("plan mode stop", func(t *testing.T) {
		raw := []byte(`{
			"request_id": "r2",
			"permission_mode": "plan",
			"last_assistant_message": "# My Plan\n\n1. Do this\n2. Do that"
		}`)
		msg, err := adapter.ParseHookInput("stop", raw)
		if err != nil {
			t.Fatal(err)
		}
		stop := msg.(*protocol.StopMsg)
		if stop.ReviewContent != "# My Plan\n\n1. Do this\n2. Do that" {
			t.Errorf("expected plan content, got %q", stop.ReviewContent)
		}
		if stop.ReviewContentTitle != "Plan" {
			t.Errorf("expected title 'Plan', got %q", stop.ReviewContentTitle)
		}
		if stop.ReviewContentType != "markdown" {
			t.Errorf("expected type 'markdown', got %q", stop.ReviewContentType)
		}
	})

	t.Run("plan mode without message", func(t *testing.T) {
		raw := []byte(`{"request_id":"r3","permission_mode":"plan"}`)
		msg, err := adapter.ParseHookInput("stop", raw)
		if err != nil {
			t.Fatal(err)
		}
		stop := msg.(*protocol.StopMsg)
		if stop.ReviewContent != "" {
			t.Error("expected no review content when last_assistant_message is empty")
		}
	})

	t.Run("non-plan mode with message", func(t *testing.T) {
		raw := []byte(`{
			"request_id": "r4",
			"permission_mode": "default",
			"last_assistant_message": "I finished the task."
		}`)
		msg, err := adapter.ParseHookInput("stop", raw)
		if err != nil {
			t.Fatal(err)
		}
		stop := msg.(*protocol.StopMsg)
		if stop.ReviewContent != "" {
			t.Error("expected no review content for non-plan mode")
		}
	})
}

package core

import (
	"testing"
	"time"

	"github.com/anthropics/monocle/internal/protocol"
)

func TestSubmitWhileStopped(t *testing.T) {
	fq := NewFeedbackQueue()

	var resp *protocol.StopResponse
	done := make(chan struct{})

	// Simulate agent stopping (blocks)
	go func() {
		resp = fq.OnStop("req-1")
		close(done)
	}()

	// Give goroutine time to block
	time.Sleep(50 * time.Millisecond)

	if fq.GetStatus() != "none" {
		t.Errorf("expected status none, got %q", fq.GetStatus())
	}

	// User submits while agent is stopped
	fq.Submit(&FormattedReview{
		Formatted:    "## Review\nFix bug",
		CommentCount: 1,
		Action:       "request_changes",
	})

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("OnStop did not return")
	}

	if !resp.Continue {
		t.Error("expected continue=true")
	}
	if resp.SystemMessage != "## Review\nFix bug" {
		t.Errorf("unexpected message: %q", resp.SystemMessage)
	}
	if fq.GetStatus() != "delivered" {
		t.Errorf("expected status delivered, got %q", fq.GetStatus())
	}
}

func TestSubmitWhileWorking(t *testing.T) {
	fq := NewFeedbackQueue()

	// User submits while agent is working (not stopped)
	fq.Submit(&FormattedReview{
		Formatted:    "## Review\nLooks good",
		CommentCount: 1,
		Action:       "request_changes",
	})

	if fq.GetStatus() != "queued" {
		t.Errorf("expected status queued, got %q", fq.GetStatus())
	}

	// Agent stops — should get queued review immediately
	resp := fq.OnStop("req-2")

	if !resp.Continue {
		t.Error("expected continue=true")
	}
	if resp.SystemMessage != "## Review\nLooks good" {
		t.Errorf("unexpected message: %q", resp.SystemMessage)
	}
}

func TestApproveWhileStopped(t *testing.T) {
	fq := NewFeedbackQueue()

	var resp *protocol.StopResponse
	done := make(chan struct{})

	go func() {
		resp = fq.OnStop("req-3")
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)

	fq.Approve()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("OnStop did not return")
	}

	if resp.Continue {
		t.Error("expected continue=false")
	}
	if resp.SystemMessage != "" {
		t.Errorf("expected empty message, got %q", resp.SystemMessage)
	}
}

func TestApproveWhileWorking(t *testing.T) {
	fq := NewFeedbackQueue()

	// Approve while agent is working — should be no-op
	fq.Approve()

	if fq.GetStatus() != "none" {
		t.Errorf("expected status none, got %q", fq.GetStatus())
	}
}

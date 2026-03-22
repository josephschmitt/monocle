package core

import (
	"testing"
	"time"
)

func TestPollNoFeedback(t *testing.T) {
	fq := NewFeedbackQueue()

	review := fq.Poll()
	if review != nil {
		t.Error("expected nil on empty queue")
	}
}

func TestSubmitThenPoll(t *testing.T) {
	fq := NewFeedbackQueue()

	fq.Submit(&FormattedReview{
		Formatted:    "## Review\nFix bug",
		CommentCount: 1,
		Action:       "request_changes",
	})

	if fq.GetStatus() != "queued" {
		t.Errorf("expected status queued, got %q", fq.GetStatus())
	}

	review := fq.Poll()
	if review == nil {
		t.Fatal("expected review from Poll")
	}
	if review.Formatted != "## Review\nFix bug" {
		t.Errorf("unexpected review: %q", review.Formatted)
	}
	if fq.GetStatus() != "delivered" {
		t.Errorf("expected status delivered, got %q", fq.GetStatus())
	}

	// Second poll should return nil
	if fq.Poll() != nil {
		t.Error("expected nil after delivery")
	}
}

func TestWaitForFeedback(t *testing.T) {
	fq := NewFeedbackQueue()

	var review *FormattedReview
	done := make(chan struct{})

	go func() {
		review = fq.WaitForFeedback()
		close(done)
	}()

	// Give goroutine time to block
	time.Sleep(50 * time.Millisecond)

	// Submit feedback
	fq.Submit(&FormattedReview{
		Formatted:    "## Review\nFix bug",
		CommentCount: 1,
		Action:       "request_changes",
	})

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("WaitForFeedback did not return")
	}

	if review == nil {
		t.Fatal("expected review")
	}
	if review.Formatted != "## Review\nFix bug" {
		t.Errorf("unexpected review: %q", review.Formatted)
	}
}

func TestWaitForFeedbackWithPending(t *testing.T) {
	fq := NewFeedbackQueue()

	// Submit before waiting
	fq.Submit(&FormattedReview{
		Formatted:    "## Review\nLooks good",
		CommentCount: 1,
		Action:       "approve",
	})

	// WaitForFeedback should return immediately
	review := fq.WaitForFeedback()
	if review == nil {
		t.Fatal("expected review")
	}
	if review.Formatted != "## Review\nLooks good" {
		t.Errorf("unexpected review: %q", review.Formatted)
	}
}

func TestPauseRequested(t *testing.T) {
	fq := NewFeedbackQueue()

	if fq.IsPauseRequested() {
		t.Error("expected pause not requested initially")
	}

	fq.SetPauseRequested(true)

	if !fq.IsPauseRequested() {
		t.Error("expected pause requested after set")
	}

	// Submit should clear pause
	fq.Submit(&FormattedReview{
		Formatted:    "review",
		CommentCount: 1,
		Action:       "request_changes",
	})

	if fq.IsPauseRequested() {
		t.Error("expected pause cleared after Submit")
	}
}

func TestHasPending(t *testing.T) {
	fq := NewFeedbackQueue()

	if fq.HasPending() {
		t.Error("expected HasPending=false on new queue")
	}

	fq.Submit(&FormattedReview{
		Formatted:    "review",
		CommentCount: 1,
		Action:       "request_changes",
	})

	if !fq.HasPending() {
		t.Error("expected HasPending=true after Submit")
	}

	fq.Poll()

	if fq.HasPending() {
		t.Error("expected HasPending=false after Poll")
	}
}


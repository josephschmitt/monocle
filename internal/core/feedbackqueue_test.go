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
	}, false)

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
	}, false)

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
	}, false)

	// WaitForFeedback should return immediately
	review := fq.WaitForFeedback()
	if review == nil {
		t.Fatal("expected review")
	}
	if review.Formatted != "## Review\nLooks good" {
		t.Errorf("unexpected review: %q", review.Formatted)
	}
}

// TestWaitForFeedbackCancellable_PreservesFeedback reproduces the
// disconnect-mid-wait data-loss bug: a wait that is cancelled (its client
// socket died) must NOT drain the queue, so feedback submitted afterwards
// still reaches the next poller instead of being silently consumed and
// marked delivered to a dead connection.
func TestWaitForFeedbackCancellable_PreservesFeedback(t *testing.T) {
	fq := NewFeedbackQueue()

	cancel := make(chan struct{})
	got := make(chan *PollResult, 1)
	go func() {
		got <- fq.WaitForFeedbackCancellable(cancel)
	}()

	// Let the waiter park, then simulate the client disconnecting.
	time.Sleep(50 * time.Millisecond)
	close(cancel)

	select {
	case res := <-got:
		if res != nil {
			t.Fatalf("cancelled wait should return nil, got %+v", res)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("cancelled wait did not return")
	}

	// Feedback submitted after the abandoned wait must survive.
	fq.Submit(&FormattedReview{Formatted: "## Review\nFix bug", CommentCount: 1, Action: "request_changes"}, false)
	if review := fq.Poll(); review == nil {
		t.Fatal("feedback was lost — cancelled wait drained the queue")
	} else if review.Formatted != "## Review\nFix bug" {
		t.Errorf("unexpected review: %q", review.Formatted)
	}
}

// TestWaitForFeedbackCancellable_CancelAfterPending verifies that even when
// feedback is already queued, a fired cancel returns nil without consuming
// it — the disconnect wins so the review is kept for the next poller.
func TestWaitForFeedbackCancellable_CancelAfterPending(t *testing.T) {
	fq := NewFeedbackQueue()
	fq.Submit(&FormattedReview{Formatted: "pending", CommentCount: 0, Action: "approve"}, false)

	cancel := make(chan struct{})
	close(cancel)

	if res := fq.WaitForFeedbackCancellable(cancel); res != nil {
		t.Fatalf("pre-cancelled wait should return nil, got %+v", res)
	}
	if !fq.HasPending() {
		t.Fatal("pre-cancelled wait consumed already-pending feedback")
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
	}, false)

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
	}, false)

	if !fq.HasPending() {
		t.Error("expected HasPending=true after Submit")
	}

	fq.Poll()

	if fq.HasPending() {
		t.Error("expected HasPending=false after Poll")
	}
}


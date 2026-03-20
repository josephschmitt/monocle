package core

import (
	"sync"
)

// FormattedReview holds a formatted review ready for delivery.
type FormattedReview struct {
	Formatted    string
	CommentCount int
	Action       string
}

// ReviewStatusInfo holds the current review status for MCP channel queries.
type ReviewStatusInfo struct {
	Status       string // "no_feedback" | "pending" | "pause_requested"
	CommentCount int
	Summary      string
}

// FeedbackQueue manages the synchronization between user review actions
// and MCP channel feedback retrieval. Supports both non-blocking and blocking
// wait (pause flow) models.
type FeedbackQueue struct {
	mu   sync.Mutex
	cond *sync.Cond

	// pending holds a review waiting to be delivered
	pending *FormattedReview

	// pauseRequested is set when the user wants the agent to stop and wait
	pauseRequested bool

	// status tracks delivery state
	status string // "none" | "queued" | "delivered"
}

// NewFeedbackQueue creates a new FeedbackQueue.
func NewFeedbackQueue() *FeedbackQueue {
	fq := &FeedbackQueue{status: "none"}
	fq.cond = sync.NewCond(&fq.mu)
	return fq
}

// Submit stores a review for delivery. If a wait handler is blocking,
// it wakes it to deliver immediately. If the channel hasn't requested it yet,
// the review is queued for the next request.
func (fq *FeedbackQueue) Submit(review *FormattedReview) {
	fq.mu.Lock()
	defer fq.mu.Unlock()

	fq.pending = review
	fq.status = "queued"
	fq.pauseRequested = false
	fq.cond.Broadcast()
}

// Approve clears any pending review and pause state.
func (fq *FeedbackQueue) Approve() {
	fq.mu.Lock()
	defer fq.mu.Unlock()

	fq.pending = nil
	fq.pauseRequested = false
	fq.status = "none"
	fq.cond.Broadcast()
}

// Poll returns pending feedback without blocking. Returns nil if none available.
func (fq *FeedbackQueue) Poll() *FormattedReview {
	fq.mu.Lock()
	defer fq.mu.Unlock()

	if fq.pending == nil {
		return nil
	}

	review := fq.pending
	fq.pending = nil
	fq.status = "delivered"
	return review
}

// WaitForFeedback blocks until the user submits feedback. Used for the "pause" flow
// where the agent explicitly waits for review.
func (fq *FeedbackQueue) WaitForFeedback() *FormattedReview {
	fq.mu.Lock()
	defer fq.mu.Unlock()

	// If there's already pending feedback, return it immediately
	if fq.pending != nil {
		review := fq.pending
		fq.pending = nil
		fq.status = "delivered"
		fq.pauseRequested = false
		return review
	}

	// Block until feedback is submitted
	for fq.pending == nil {
		fq.cond.Wait()
	}

	review := fq.pending
	fq.pending = nil
	fq.status = "delivered"
	fq.pauseRequested = false
	return review
}

// SetPauseRequested sets the pause flag. The next review_status call
// from Claude Code will see "pause_requested".
func (fq *FeedbackQueue) SetPauseRequested(paused bool) {
	fq.mu.Lock()
	defer fq.mu.Unlock()
	fq.pauseRequested = paused
}

// IsPauseRequested returns whether the user has requested a pause.
func (fq *FeedbackQueue) IsPauseRequested() bool {
	fq.mu.Lock()
	defer fq.mu.Unlock()
	return fq.pauseRequested
}

// GetStatus returns the current feedback status.
func (fq *FeedbackQueue) GetStatus() string {
	fq.mu.Lock()
	defer fq.mu.Unlock()
	return fq.status
}

// HasPending returns true if there is a queued review waiting for delivery.
func (fq *FeedbackQueue) HasPending() bool {
	fq.mu.Lock()
	defer fq.mu.Unlock()
	return fq.pending != nil
}

// Reset clears the queue state.
func (fq *FeedbackQueue) Reset() {
	fq.mu.Lock()
	defer fq.mu.Unlock()
	fq.pending = nil
	fq.pauseRequested = false
	fq.status = "none"
}

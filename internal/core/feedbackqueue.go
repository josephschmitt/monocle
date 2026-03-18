package core

import (
	"sync"

	"github.com/anthropics/monocle/internal/protocol"
)

// FormattedReview holds a formatted review ready for delivery.
type FormattedReview struct {
	Formatted    string
	CommentCount int
	Action       string
}

// FeedbackQueue manages the synchronization between user review actions
// and agent stop events. When the agent stops, the hook handler blocks
// until the user submits or approves.
type FeedbackQueue struct {
	mu   sync.Mutex
	cond *sync.Cond

	// pending holds a review waiting to be delivered
	pending *FormattedReview

	// approved is set when user approves (release agent without feedback)
	approved bool

	// waiting is true when a stop handler is blocked waiting for user action
	waiting   bool
	requestID string

	// status tracks delivery state
	status string // "none" | "queued" | "delivered"
}

// NewFeedbackQueue creates a new FeedbackQueue.
func NewFeedbackQueue() *FeedbackQueue {
	fq := &FeedbackQueue{status: "none"}
	fq.cond = sync.NewCond(&fq.mu)
	return fq
}

// Submit stores a review for delivery. If a stop handler is waiting,
// it wakes it to deliver immediately. If the agent is still working,
// the review is queued for delivery when the agent next stops.
func (fq *FeedbackQueue) Submit(review *FormattedReview) {
	fq.mu.Lock()
	defer fq.mu.Unlock()

	fq.pending = review
	fq.approved = false

	if fq.waiting {
		fq.status = "delivered"
		fq.cond.Broadcast()
	} else {
		fq.status = "queued"
	}
}

// Approve signals that the agent should be released without feedback.
// If a stop handler is waiting, it wakes it. If the agent is working,
// this is a no-op.
func (fq *FeedbackQueue) Approve() {
	fq.mu.Lock()
	defer fq.mu.Unlock()

	if !fq.waiting {
		return // no-op if agent isn't stopped
	}

	fq.approved = true
	fq.pending = nil
	fq.status = "none"
	fq.cond.Broadcast()
}

// OnStop blocks until the user submits a review or approves.
// Called by the hook server's stop handler goroutine.
// Returns the appropriate StopResponse.
func (fq *FeedbackQueue) OnStop(requestID string) *protocol.StopResponse {
	fq.mu.Lock()
	defer fq.mu.Unlock()

	fq.requestID = requestID
	fq.waiting = true
	defer func() { fq.waiting = false }()

	// If there's already a queued review, deliver immediately
	if fq.pending != nil {
		return fq.flush(requestID)
	}

	// Block until user acts
	for fq.pending == nil && !fq.approved {
		fq.cond.Wait()
	}

	if fq.approved {
		fq.approved = false
		return &protocol.StopResponse{
			Type:      protocol.TypeStopResponse,
			RequestID: requestID,
			Continue:  false,
		}
	}

	return fq.flush(requestID)
}

// flush delivers the pending review and clears state. Must be called with mu held.
func (fq *FeedbackQueue) flush(requestID string) *protocol.StopResponse {
	review := fq.pending
	fq.pending = nil
	fq.status = "delivered"

	return &protocol.StopResponse{
		Type:          protocol.TypeStopResponse,
		RequestID:     requestID,
		Continue:      true,
		SystemMessage: review.Formatted,
	}
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
	fq.approved = false
	fq.status = "none"
}

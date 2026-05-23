package client

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/josephschmitt/monocle/internal/core"
	"github.com/josephschmitt/monocle/internal/protocol"
	"github.com/josephschmitt/monocle/internal/types"
)

// EngineClient is a socket-backed implementation of core.EngineAPI. It opens
// a single persistent connection to monocle serve, sends a SubscribeMsg so
// engine events are pushed back, then serialises request/response calls over
// the same socket. Incoming messages are demultiplexed into either the local
// event bus (for EventNotification) or the pending request channel.
type EngineClient struct {
	conn    net.Conn
	scanner *bufio.Scanner

	writeMu sync.Mutex // serialises writes to the socket
	reqMu   sync.Mutex // one request in flight at a time
	pending chan any   // read loop delivers non-event responses here
	readErr chan error // terminal read error

	subsMu      sync.Mutex
	subscribers map[core.EventKind]map[int]core.EventCallback
	nextSubID   int

	closed chan struct{}
	// cached config, mutated by the TUI between GetConfig/SaveConfig calls.
	// Pointer identity matches what the engine exposes locally.
	cfg *types.Config
}

// NewEngineClient dials the socket, subscribes to every engine event kind,
// and starts the demultiplexing read loop. The returned client satisfies
// core.EngineAPI.
func NewEngineClient(socketPath string) (*EngineClient, error) {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", socketPath, err)
	}
	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	c := &EngineClient{
		conn:        conn,
		scanner:     scanner,
		pending:     make(chan any, 1),
		readErr:     make(chan error, 1),
		subscribers: make(map[core.EventKind]map[int]core.EventCallback),
		closed:      make(chan struct{}),
	}

	// Subscribe to every event kind so the client mirrors all engine events.
	// Passive=true marks this as a viewer connection: the TUI is not an
	// attached agent, so the engine must not flip into push-delivery mode
	// just because the reviewer's UI is open. Without this flag the server
	// would count the TUI in subscriberCount, take the push branch in
	// Submit(), and silently mark the review delivered — the real agent
	// would never see it.
	sub := &protocol.SubscribeMsg{
		Type: protocol.TypeSubscribe,
		Events: []string{
			string(core.EventFileChanged),
			string(core.EventFeedbackStatusChanged),
			string(core.EventContentItemAdded),
			string(core.EventPauseChanged),
			string(core.EventFeedbackSubmitted),
			string(core.EventConnectionChanged),
			string(core.EventAdditionalFileAdded),
			string(core.EventFeedbackPickedUp),
			string(core.EventWaitStatusChanged),
		},
		Passive: true,
	}
	data, err := protocol.Encode(sub)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("encode subscribe: %w", err)
	}
	if _, err := conn.Write(data); err != nil {
		conn.Close()
		return nil, fmt.Errorf("write subscribe: %w", err)
	}

	// The very first line is the SubscribeResponse ack; consume it before
	// starting the read loop so callers can immediately issue requests.
	if !scanner.Scan() {
		conn.Close()
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("read subscribe ack: %w", err)
		}
		return nil, errors.New("server closed before subscribe ack")
	}
	ack, err := protocol.Decode(scanner.Bytes())
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("decode subscribe ack: %w", err)
	}
	if _, ok := ack.(*protocol.SubscribeResponse); !ok {
		conn.Close()
		return nil, fmt.Errorf("unexpected subscribe ack: %T", ack)
	}

	go c.readLoop()
	return c, nil
}

// readLoop demultiplexes incoming messages into events (dispatched locally)
// or request responses (sent on the pending channel for the caller to pick up).
func (c *EngineClient) readLoop() {
	for c.scanner.Scan() {
		line := c.scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		msg, err := protocol.Decode(line)
		if err != nil {
			continue // best-effort: skip garbage rather than closing the link
		}
		if notif, ok := msg.(*protocol.EventNotification); ok {
			c.dispatchEvent(notif)
			continue
		}
		select {
		case c.pending <- msg:
		case <-c.closed:
			return
		}
	}
	err := c.scanner.Err()
	if err == nil {
		err = errors.New("connection closed")
	}
	select {
	case c.readErr <- err:
	default:
	}
	close(c.closed)
}

// request sends msg and blocks until the corresponding response arrives. The
// reqMu lock guarantees one in-flight request at a time, so the order of
// non-event messages on the pending channel matches the order of sends.
//
// On timeout we tear down the underlying socket rather than just returning.
// The pending channel is shared across requests with no correlation ID, so
// a stale response that arrives after the timeout would otherwise sit in
// the buffer and be read by the next call — producing a wrong-type assertion
// panic when the wrapper does resp.(*protocol.XxxResponse). Closing the
// connection forces a clean reconnect on the next operation and prevents
// cross-talk.
func (c *EngineClient) request(msg any) (any, error) {
	data, err := protocol.Encode(msg)
	if err != nil {
		return nil, fmt.Errorf("encode: %w", err)
	}

	c.reqMu.Lock()
	defer c.reqMu.Unlock()

	c.writeMu.Lock()
	_, werr := c.conn.Write(data)
	c.writeMu.Unlock()
	if werr != nil {
		return nil, fmt.Errorf("write: %w", werr)
	}

	select {
	case resp := <-c.pending:
		return resp, nil
	case err := <-c.readErr:
		return nil, err
	case <-c.closed:
		return nil, errors.New("client closed")
	case <-time.After(DefaultTimeout):
		_ = c.conn.Close()
		return nil, errors.New("timeout waiting for response")
	}
}

func (c *EngineClient) dispatchEvent(notif *protocol.EventNotification) {
	kind := core.EventKind(notif.Event)
	payload := core.EventPayload{Kind: kind}
	if v, ok := notif.Payload["message"].(string); ok {
		payload.Message = v
	}
	if v, ok := notif.Payload["status"].(string); ok {
		payload.Status = v
	}
	if v, ok := notif.Payload["path"].(string); ok {
		payload.Path = v
	}
	if v, ok := notif.Payload["item_id"].(string); ok {
		payload.ItemID = v
	}

	c.subsMu.Lock()
	callbacks := make([]core.EventCallback, 0, len(c.subscribers[kind]))
	for _, cb := range c.subscribers[kind] {
		callbacks = append(callbacks, cb)
	}
	c.subsMu.Unlock()

	for _, cb := range callbacks {
		cb(payload)
	}
}

// Close tears down the underlying socket and terminates the read loop.
func (c *EngineClient) Close() error {
	return c.conn.Close()
}

// On satisfies core.EngineAPI. Registrations are local-only — the server
// already pushes every event kind on the subscribe connection, so the client
// just needs to fan them out to TUI callbacks.
func (c *EngineClient) On(event core.EventKind, callback core.EventCallback) core.UnsubscribeFunc {
	c.subsMu.Lock()
	if c.subscribers[event] == nil {
		c.subscribers[event] = make(map[int]core.EventCallback)
	}
	id := c.nextSubID
	c.nextSubID++
	c.subscribers[event][id] = callback
	c.subsMu.Unlock()

	return func() {
		c.subsMu.Lock()
		delete(c.subscribers[event], id)
		c.subsMu.Unlock()
	}
}

// --- EngineAPI: sessions ---

func (c *EngineClient) StartSession(opts core.SessionOptions) (*types.ReviewSession, error) {
	resp, err := c.request(&protocol.StartSessionMsg{
		Type:           protocol.TypeStartSession,
		Agent:          opts.Agent,
		RepoRoot:       opts.RepoRoot,
		BaseRef:        opts.BaseRef,
		IgnorePatterns: opts.IgnorePatterns,
	})
	if err != nil {
		return nil, err
	}
	r := resp.(*protocol.StartSessionResponse)
	if r.Error != "" {
		return nil, errors.New(r.Error)
	}
	return r.Session, nil
}

func (c *EngineClient) ResumeSession(sessionID string) (*types.ReviewSession, error) {
	resp, err := c.request(&protocol.ResumeSessionMsg{Type: protocol.TypeResumeSession, SessionID: sessionID})
	if err != nil {
		return nil, err
	}
	r := resp.(*protocol.ResumeSessionResponse)
	if r.Error != "" {
		return nil, errors.New(r.Error)
	}
	return r.Session, nil
}

func (c *EngineClient) GetSession() *types.ReviewSession {
	resp, err := c.request(&protocol.GetSessionMsg{Type: protocol.TypeGetSession})
	if err != nil {
		return nil
	}
	return resp.(*protocol.GetSessionResponse).Session
}

func (c *EngineClient) ListSessions(opts core.ListSessionsOptions) ([]types.SessionSummary, error) {
	resp, err := c.request(&protocol.ListSessionsMsg{
		Type:     protocol.TypeListSessions,
		RepoRoot: opts.RepoRoot,
		Limit:    opts.Limit,
	})
	if err != nil {
		return nil, err
	}
	r := resp.(*protocol.ListSessionsResponse)
	if r.Error != "" {
		return nil, errors.New(r.Error)
	}
	return r.Sessions, nil
}

// --- EngineAPI: files ---

func (c *EngineClient) RefreshChangedFiles() ([]types.ChangedFile, error) {
	resp, err := c.request(&protocol.RefreshChangedFilesMsg{Type: protocol.TypeRefreshChangedFiles})
	if err != nil {
		return nil, err
	}
	r := resp.(*protocol.RefreshChangedFilesResponse)
	if r.Error != "" {
		return nil, errors.New(r.Error)
	}
	return r.Files, nil
}

func (c *EngineClient) GetChangedFiles() []types.ChangedFile {
	resp, err := c.request(&protocol.GetChangedFilesMsg{Type: protocol.TypeGetChangedFiles})
	if err != nil {
		return nil
	}
	return resp.(*protocol.GetChangedFilesResponse).Files
}

func (c *EngineClient) GetFileDiff(path string) (*types.DiffResult, error) {
	resp, err := c.request(&protocol.GetFileDiffMsg{Type: protocol.TypeGetFileDiff, Path: path})
	if err != nil {
		return nil, err
	}
	r := resp.(*protocol.GetFileDiffResponse)
	if r.Error != "" {
		return nil, errors.New(r.Error)
	}
	return r.Diff, nil
}

func (c *EngineClient) GetFileContent(path string) (string, error) {
	resp, err := c.request(&protocol.GetFileContentMsg{Type: protocol.TypeGetFileContent, Path: path})
	if err != nil {
		return "", err
	}
	r := resp.(*protocol.GetFileContentResponse)
	if r.Error != "" {
		return "", errors.New(r.Error)
	}
	return r.Content, nil
}

// --- EngineAPI: content ---

func (c *EngineClient) GetContentItems() []types.ContentItem {
	resp, err := c.request(&protocol.GetContentItemsMsg{Type: protocol.TypeGetContentItems})
	if err != nil {
		return nil
	}
	return resp.(*protocol.GetContentItemsResponse).Items
}

func (c *EngineClient) GetContentItem(id string) (*types.ContentItem, error) {
	resp, err := c.request(&protocol.GetContentItemMsg{Type: protocol.TypeGetContentItem, ID: id})
	if err != nil {
		return nil, err
	}
	r := resp.(*protocol.GetContentItemResponse)
	if r.Error != "" {
		return nil, errors.New(r.Error)
	}
	return r.Item, nil
}

func (c *EngineClient) GetContentDiff(id string) (*types.DiffResult, error) {
	resp, err := c.request(&protocol.GetContentDiffMsg{Type: protocol.TypeGetContentDiff, ID: id})
	if err != nil {
		return nil, err
	}
	r := resp.(*protocol.GetContentDiffResponse)
	if r.Error != "" {
		return nil, errors.New(r.Error)
	}
	return r.Diff, nil
}

func (c *EngineClient) GetContentVersions(id string) ([]types.ContentVersion, error) {
	resp, err := c.request(&protocol.GetContentVersionsMsg{Type: protocol.TypeGetContentVersions, ID: id})
	if err != nil {
		return nil, err
	}
	r := resp.(*protocol.GetContentVersionsResponse)
	if r.Error != "" {
		return nil, errors.New(r.Error)
	}
	return r.Versions, nil
}

func (c *EngineClient) GetContentDiffBetweenVersions(id string, fromVersion, toVersion int) (*types.DiffResult, error) {
	resp, err := c.request(&protocol.GetContentDiffBetweenVersionsMsg{
		Type:        protocol.TypeGetContentDiffBetweenVersion,
		ID:          id,
		FromVersion: fromVersion,
		ToVersion:   toVersion,
	})
	if err != nil {
		return nil, err
	}
	r := resp.(*protocol.GetContentDiffBetweenVersionsResponse)
	if r.Error != "" {
		return nil, errors.New(r.Error)
	}
	return r.Diff, nil
}

func (c *EngineClient) DismissArtifact(id string) error {
	resp, err := c.request(&protocol.DismissArtifactMsg{Type: protocol.TypeDismissArtifact, ID: id})
	if err != nil {
		return err
	}
	r := resp.(*protocol.DismissArtifactResponse)
	if r.Error != "" {
		return errors.New(r.Error)
	}
	return nil
}

// --- EngineAPI: additional files ---

func (c *EngineClient) GetAdditionalFiles() []types.AdditionalFile {
	resp, err := c.request(&protocol.GetAdditionalFilesMsg{Type: protocol.TypeGetAdditionalFiles})
	if err != nil {
		return nil
	}
	return resp.(*protocol.GetAdditionalFilesResponse).Files
}

func (c *EngineClient) AddAdditionalPaths(paths []string) ([]types.AdditionalFile, error) {
	// Reuse the existing AddAdditionalFilesMsg already used by the CLI.
	resp, err := c.request(&protocol.AddAdditionalFilesMsg{
		Type:  protocol.TypeAddAdditionalFiles,
		Paths: paths,
	})
	if err != nil {
		return nil, err
	}
	r := resp.(*protocol.AddAdditionalFilesResponse)
	if !r.Success {
		if r.Message != "" {
			return nil, errors.New(r.Message)
		}
		return nil, errors.New("add additional files failed")
	}
	// The existing message doesn't return the resulting AdditionalFile list,
	// so re-fetch to keep EngineAPI parity.
	return c.GetAdditionalFiles(), nil
}

func (c *EngineClient) GetAdditionalFileContent(absPath string) (string, error) {
	resp, err := c.request(&protocol.GetAdditionalFileContentMsg{Type: protocol.TypeGetAdditionalFileContent, AbsPath: absPath})
	if err != nil {
		return "", err
	}
	r := resp.(*protocol.GetAdditionalFileContentResponse)
	if r.Error != "" {
		return "", errors.New(r.Error)
	}
	return r.Content, nil
}

// --- EngineAPI: comments ---

func (c *EngineClient) AddComment(target core.CommentTarget, commentType types.CommentType, body string) (*types.ReviewComment, error) {
	resp, err := c.request(&protocol.AddCommentMsg{
		Type:        protocol.TypeAddComment,
		TargetType:  target.TargetType,
		TargetRef:   target.TargetRef,
		LineStart:   target.LineStart,
		LineEnd:     target.LineEnd,
		CommentType: commentType,
		Body:        body,
	})
	if err != nil {
		return nil, err
	}
	r := resp.(*protocol.AddCommentResponse)
	if r.Error != "" {
		return nil, errors.New(r.Error)
	}
	return r.Comment, nil
}

func (c *EngineClient) EditComment(commentID string, commentType types.CommentType, body string) (*types.ReviewComment, error) {
	resp, err := c.request(&protocol.EditCommentMsg{
		Type:        protocol.TypeEditComment,
		CommentID:   commentID,
		CommentType: commentType,
		Body:        body,
	})
	if err != nil {
		return nil, err
	}
	r := resp.(*protocol.EditCommentResponse)
	if r.Error != "" {
		return nil, errors.New(r.Error)
	}
	return r.Comment, nil
}

func (c *EngineClient) DeleteComment(commentID string) error {
	resp, err := c.request(&protocol.DeleteCommentMsg{Type: protocol.TypeDeleteComment, CommentID: commentID})
	if err != nil {
		return err
	}
	r := resp.(*protocol.DeleteCommentResponse)
	if r.Error != "" {
		return errors.New(r.Error)
	}
	return nil
}

func (c *EngineClient) ResolveComment(commentID string) error {
	resp, err := c.request(&protocol.ResolveCommentMsg{Type: protocol.TypeResolveComment, CommentID: commentID})
	if err != nil {
		return err
	}
	r := resp.(*protocol.ResolveCommentResponse)
	if r.Error != "" {
		return errors.New(r.Error)
	}
	return nil
}

func (c *EngineClient) ClearComments() error {
	resp, err := c.request(&protocol.ClearCommentsMsg{Type: protocol.TypeClearComments})
	if err != nil {
		return err
	}
	r := resp.(*protocol.ClearCommentsResponse)
	if r.Error != "" {
		return errors.New(r.Error)
	}
	return nil
}

func (c *EngineClient) ClearReview() error {
	resp, err := c.request(&protocol.ClearReviewMsg{Type: protocol.TypeClearReview})
	if err != nil {
		return err
	}
	r := resp.(*protocol.ClearReviewResponse)
	if r.Error != "" {
		return errors.New(r.Error)
	}
	return nil
}

// --- EngineAPI: marking ---

func (c *EngineClient) MarkReviewed(path string) error {
	resp, err := c.request(&protocol.MarkReviewedMsg{Type: protocol.TypeMarkReviewed, Path: path})
	if err != nil {
		return err
	}
	r := resp.(*protocol.MarkReviewedResponse)
	if r.Error != "" {
		return errors.New(r.Error)
	}
	return nil
}

func (c *EngineClient) UnmarkReviewed(path string) error {
	resp, err := c.request(&protocol.UnmarkReviewedMsg{Type: protocol.TypeUnmarkReviewed, Path: path})
	if err != nil {
		return err
	}
	r := resp.(*protocol.UnmarkReviewedResponse)
	if r.Error != "" {
		return errors.New(r.Error)
	}
	return nil
}

func (c *EngineClient) MarkContentReviewed(id string) error {
	resp, err := c.request(&protocol.MarkContentReviewedMsg{Type: protocol.TypeMarkContentReviewed, ID: id})
	if err != nil {
		return err
	}
	r := resp.(*protocol.MarkContentReviewedResponse)
	if r.Error != "" {
		return errors.New(r.Error)
	}
	return nil
}

func (c *EngineClient) UnmarkContentReviewed(id string) error {
	resp, err := c.request(&protocol.UnmarkContentReviewedMsg{Type: protocol.TypeUnmarkContentReviewed, ID: id})
	if err != nil {
		return err
	}
	r := resp.(*protocol.UnmarkContentReviewedResponse)
	if r.Error != "" {
		return errors.New(r.Error)
	}
	return nil
}

func (c *EngineClient) ResetAllReviewed() error {
	resp, err := c.request(&protocol.ResetAllReviewedMsg{Type: protocol.TypeResetAllReviewed})
	if err != nil {
		return err
	}
	r := resp.(*protocol.ResetAllReviewedResponse)
	if r.Error != "" {
		return errors.New(r.Error)
	}
	return nil
}

func (c *EngineClient) MarkAllReviewed() error {
	resp, err := c.request(&protocol.MarkAllReviewedMsg{Type: protocol.TypeMarkAllReviewed})
	if err != nil {
		return err
	}
	r := resp.(*protocol.MarkAllReviewedResponse)
	if r.Error != "" {
		return errors.New(r.Error)
	}
	return nil
}

// --- EngineAPI: submission ---

func (c *EngineClient) GetReviewSummary() (*types.ReviewSummary, error) {
	resp, err := c.request(&protocol.GetReviewSummaryMsg{Type: protocol.TypeGetReviewSummary})
	if err != nil {
		return nil, err
	}
	r := resp.(*protocol.GetReviewSummaryResponse)
	if r.Error != "" {
		return nil, errors.New(r.Error)
	}
	return r.Summary, nil
}

func (c *EngineClient) Submit(action types.SubmitAction, body string) (*core.SubmitResult, error) {
	resp, err := c.request(&protocol.SubmitMsg{Type: protocol.TypeSubmit, Action: action, Body: body})
	if err != nil {
		return nil, err
	}
	r := resp.(*protocol.SubmitResponse)
	if r.Error != "" {
		return nil, errors.New(r.Error)
	}
	return &core.SubmitResult{AgentConnected: r.AgentConnected}, nil
}

func (c *EngineClient) FormatReview(action types.SubmitAction, body string) (string, error) {
	resp, err := c.request(&protocol.FormatReviewMsg{Type: protocol.TypeFormatReview, Action: action, Body: body})
	if err != nil {
		return "", err
	}
	r := resp.(*protocol.FormatReviewResponse)
	if r.Error != "" {
		return "", errors.New(r.Error)
	}
	return r.Formatted, nil
}

func (c *EngineClient) GetSubmissions() ([]types.ReviewSubmission, error) {
	resp, err := c.request(&protocol.GetSubmissionsMsg{Type: protocol.TypeGetSubmissions})
	if err != nil {
		return nil, err
	}
	r := resp.(*protocol.GetSubmissionsResponse)
	if r.Error != "" {
		return nil, errors.New(r.Error)
	}
	return r.Submissions, nil
}

// --- EngineAPI: base ref ---

func (c *EngineClient) SetBaseRef(ref string) error {
	resp, err := c.request(&protocol.SetBaseRefMsg{Type: protocol.TypeSetBaseRef, Ref: ref})
	if err != nil {
		return err
	}
	r := resp.(*protocol.SetBaseRefResponse)
	if r.Error != "" {
		return errors.New(r.Error)
	}
	return nil
}

func (c *EngineClient) SetAutoAdvanceRef(enabled bool) {
	_, _ = c.request(&protocol.SetAutoAdvanceRefMsg{Type: protocol.TypeSetAutoAdvanceRef, Enabled: enabled})
}

func (c *EngineClient) IsAutoAdvanceRef() bool {
	resp, err := c.request(&protocol.IsAutoAdvanceRefMsg{Type: protocol.TypeIsAutoAdvanceRef})
	if err != nil {
		return false
	}
	return resp.(*protocol.IsAutoAdvanceRefResponse).Enabled
}

func (c *EngineClient) SelectedBaseRef() string {
	resp, err := c.request(&protocol.SelectedBaseRefMsg{Type: protocol.TypeSelectedBaseRef})
	if err != nil {
		return ""
	}
	return resp.(*protocol.SelectedBaseRefResponse).Ref
}

func (c *EngineClient) RecentCommits(n int) ([]core.LogEntry, error) {
	resp, err := c.request(&protocol.RecentCommitsMsg{Type: protocol.TypeRecentCommits, Count: n})
	if err != nil {
		return nil, err
	}
	r := resp.(*protocol.RecentCommitsResponse)
	if r.Error != "" {
		return nil, errors.New(r.Error)
	}
	out := make([]core.LogEntry, len(r.Commits))
	for i, entry := range r.Commits {
		out[i] = core.LogEntry{Hash: entry.Hash, Subject: entry.Subject}
	}
	return out, nil
}

// --- EngineAPI: snapshots ---

func (c *EngineClient) GetSnapshots() ([]types.ReviewSnapshot, error) {
	resp, err := c.request(&protocol.GetSnapshotsMsg{Type: protocol.TypeGetSnapshots})
	if err != nil {
		return nil, err
	}
	r := resp.(*protocol.GetSnapshotsResponse)
	if r.Error != "" {
		return nil, errors.New(r.Error)
	}
	return r.Snapshots, nil
}

func (c *EngineClient) SetSnapshotBase(snapshotID int) error {
	resp, err := c.request(&protocol.SetSnapshotBaseMsg{Type: protocol.TypeSetSnapshotBase, SnapshotID: snapshotID})
	if err != nil {
		return err
	}
	r := resp.(*protocol.SetSnapshotBaseResponse)
	if r.Error != "" {
		return errors.New(r.Error)
	}
	return nil
}

func (c *EngineClient) ClearSnapshotBase() {
	_, _ = c.request(&protocol.ClearSnapshotBaseMsg{Type: protocol.TypeClearSnapshotBase})
}

func (c *EngineClient) GetActiveSnapshot() *types.ReviewSnapshot {
	resp, err := c.request(&protocol.GetActiveSnapshotMsg{Type: protocol.TypeGetActiveSnapshot})
	if err != nil {
		return nil
	}
	return resp.(*protocol.GetActiveSnapshotResponse).Snapshot
}

func (c *EngineClient) HasSnapshots() (bool, error) {
	resp, err := c.request(&protocol.HasSnapshotsMsg{Type: protocol.TypeHasSnapshots})
	if err != nil {
		return false, err
	}
	r := resp.(*protocol.HasSnapshotsResponse)
	if r.Error != "" {
		return false, errors.New(r.Error)
	}
	return r.Has, nil
}

// --- EngineAPI: server / feedback / status ---

// StartServer is a no-op on the client side — the serve process owns the
// socket. Kept so the client satisfies core.EngineAPI.
func (c *EngineClient) StartServer(_ string) error { return nil }

// PollFeedback, WaitForFeedback, GetReviewStatusInfo, SubmitContentForReview
// are agent-facing helpers that the TUI does not call. They're implemented
// as no-ops on the client so the interface is satisfied; any frontend that
// needed them would use the existing CLI-level PollFeedbackMsg /
// SubmitContentMsg plumbing.
func (c *EngineClient) PollFeedback() *core.FormattedReview     { return nil }
func (c *EngineClient) WaitForFeedback() *core.FormattedReview  { return nil }
func (c *EngineClient) GetReviewStatusInfo() *core.ReviewStatusInfo {
	return nil
}
func (c *EngineClient) SubmitContentForReview(_, _, _, _ string, _ bool) error {
	return errors.New("SubmitContentForReview not supported on client")
}

// RequestPause and CancelPause route the TUI's pause keybind through the
// daemon. Pre-fix these were `{}` stubs and the daemon's pause flag was
// never set, so `monocle review status` kept returning the prior status
// and `get-feedback --wait` did not block on user-initiated pause.
func (c *EngineClient) RequestPause() {
	_, _ = c.request(&protocol.SetPauseMsg{Type: protocol.TypeSetPause, Requested: true})
}
func (c *EngineClient) CancelPause() {
	_, _ = c.request(&protocol.SetPauseMsg{Type: protocol.TypeSetPause, Requested: false})
}

func (c *EngineClient) GetFeedbackStatus() string {
	resp, err := c.request(&protocol.GetFeedbackStatusMsg{Type: protocol.TypeGetFeedbackStatus})
	if err != nil {
		return ""
	}
	return resp.(*protocol.GetFeedbackStatusResponse).Status
}

func (c *EngineClient) GetQueuedCount() int {
	resp, err := c.request(&protocol.GetQueuedCountMsg{Type: protocol.TypeGetQueuedCount})
	if err != nil {
		return 0
	}
	return resp.(*protocol.GetQueuedCountResponse).Count
}

func (c *EngineClient) ReloadPendingFeedback() {
	_, _ = c.request(&protocol.ReloadPendingFeedbackMsg{Type: protocol.TypeReloadPendingFeedback})
}

func (c *EngineClient) GetSubscriberCount() int {
	resp, err := c.request(&protocol.GetSubscriberCountMsg{Type: protocol.TypeGetSubscriberCount})
	if err != nil {
		return 0
	}
	return resp.(*protocol.GetSubscriberCountResponse).Count
}

func (c *EngineClient) GetSocketPath() string {
	resp, err := c.request(&protocol.GetSocketPathMsg{Type: protocol.TypeGetSocketPath})
	if err != nil {
		return ""
	}
	return resp.(*protocol.GetSocketPathResponse).Path
}

// --- EngineAPI: config ---

// GetConfig returns a cached pointer to the engine's current config. Callers
// may mutate fields; SaveConfig ships the whole thing back to the server.
func (c *EngineClient) GetConfig() *types.Config {
	if c.cfg != nil {
		return c.cfg
	}
	resp, err := c.request(&protocol.GetConfigMsg{Type: protocol.TypeGetConfig})
	if err != nil {
		return nil
	}
	c.cfg = resp.(*protocol.GetConfigResponse).Config
	return c.cfg
}

func (c *EngineClient) SaveConfig() error {
	if c.cfg == nil {
		return errors.New("SaveConfig called before GetConfig")
	}
	resp, err := c.request(&protocol.SaveConfigMsg{Type: protocol.TypeSaveConfig, Config: *c.cfg})
	if err != nil {
		return err
	}
	r := resp.(*protocol.SaveConfigResponse)
	if r.Error != "" {
		return errors.New(r.Error)
	}
	return nil
}

func (c *EngineClient) IsReviewTrackingEnabled() bool {
	resp, err := c.request(&protocol.IsReviewTrackingEnabledMsg{Type: protocol.TypeIsReviewTrackingEnabled})
	if err != nil {
		return false
	}
	return resp.(*protocol.IsReviewTrackingEnabledResponse).Enabled
}

// Shutdown is a no-op on the client. The serve process manages its own
// lifecycle (including idle shutdown). Closing the client socket is the
// frontend's responsibility via Close.
func (c *EngineClient) Shutdown() {}

// Compile-time check: EngineClient satisfies core.EngineAPI.
var _ core.EngineAPI = (*EngineClient)(nil)

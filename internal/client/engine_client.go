package client

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/josephschmitt/monocle/internal/core"
	"github.com/josephschmitt/monocle/internal/protocol"
	"github.com/josephschmitt/monocle/internal/types"
)

// EngineClient is a socket-backed implementation of core.EngineAPI. It opens
// a persistent connection to monocle serve, sends a SubscribeMsg so engine
// events are pushed back, then serialises request/response calls over the
// same socket. Incoming messages are demultiplexed into either the local
// event bus (for EventNotification) or the pending request channel.
//
// On terminal socket error (write fail, timeout) the client tears the
// connection down and lazily re-dials on the next request — a single slow
// RPC must not permanently brick the client.
type EngineClient struct {
	socketPath string

	// active is the currently-installed connection bundle (conn + scanner
	// + per-conn closed channel + per-conn pending/readErr). It's a
	// pointer so dialAndSubscribe can atomically swap the entire bundle;
	// readLoop holds its OWN reference and tears down only that bundle,
	// which is what fixes the "stale readLoop closes fresh conn" race.
	connMu sync.Mutex
	active *clientConn

	writeMu sync.Mutex // serialises writes to the socket
	reqMu   sync.Mutex // one ordinary in-flight request at a time

	subsMu      sync.Mutex
	subscribers map[core.EventKind]map[int]core.EventCallback
	nextSubID   int

	// eventCh feeds dispatchWorker. dispatchEvent blocks on send (with a
	// quitCh fallback) so backpressure flows up to readLoop and ultimately
	// the daemon's per-conn outbound queue — instead of silently dropping
	// events the way the previous default-branch design did.
	eventCh  chan *protocol.EventNotification
	quitCh   chan struct{}
	quitOnce sync.Once

	// dispatchDone is closed when dispatchWorker exits; Close waits on it.
	dispatchDone chan struct{}

	// noTimeoutMu / noTimeoutConns track open requestNoTimeout
	// connections so Close can actively tear them down — otherwise a
	// blocked WaitForFeedback goroutine outlives Close indefinitely.
	noTimeoutMu    sync.Mutex
	noTimeoutConns map[net.Conn]struct{}

	// cfg is the cached config snapshot. Cache-on-first-call is the only
	// way to preserve the documented "GetConfig() -> mutate -> SaveConfig()"
	// round-trip: refreshing on every GetConfig would let an interleaving
	// fetch overwrite the pointer the caller is still mutating, silently
	// dropping the user's edits at save time. The trade-off is that
	// changes made by another client (a second TUI, monocle register) are
	// not visible until the EngineClient is reconstructed — a smaller bug
	// than losing live mutations.
	cfgMu sync.Mutex
	cfg   *types.Config
}

// clientConn bundles all per-connection state. By scoping closed/once to
// the bundle and letting readLoop hold its OWN bundle pointer, a stale
// readLoop's teardown can never accidentally close a freshly-redialed
// connection's channel or socket.
type clientConn struct {
	conn       net.Conn
	scanner    *bufio.Scanner
	pending    chan any   // non-event responses for this connection
	readErr    chan error // terminal read error for this connection
	closed     chan struct{}
	closedOnce sync.Once
	readDone   chan struct{} // closed when this conn's readLoop exits
}

// close tears down this specific connection. Idempotent. Operates only
// on the bundle, never on EngineClient's "current" fields.
func (cc *clientConn) close() {
	cc.closedOnce.Do(func() {
		close(cc.closed)
		_ = cc.conn.Close()
	})
}

// NewEngineClient dials the socket, subscribes to every engine event kind,
// and starts the demultiplexing read loop. The returned client satisfies
// core.EngineAPI.
func NewEngineClient(socketPath string) (*EngineClient, error) {
	c := &EngineClient{
		socketPath:     socketPath,
		subscribers:    make(map[core.EventKind]map[int]core.EventCallback),
		eventCh:        make(chan *protocol.EventNotification, 256),
		quitCh:         make(chan struct{}),
		dispatchDone:   make(chan struct{}),
		noTimeoutConns: make(map[net.Conn]struct{}),
	}

	if err := c.dialAndSubscribe(); err != nil {
		// Dial failed — close quitCh so the not-yet-started dispatch
		// worker won't leak if the caller retries. dispatchWorker isn't
		// running yet, but Close-on-failure must still leave the client
		// in a consistent state.
		close(c.quitCh)
		close(c.dispatchDone)
		return nil, err
	}
	// Start the dispatch worker AFTER the dial succeeds. Pre-fix this
	// ran inside NewEngineClient unconditionally; on dial failure the
	// caller got (nil, err) but the goroutine leaked because nothing
	// closed quitCh.
	go c.dispatchWorker()
	return c, nil
}

// subscribeAckTimeout bounds how long dialAndSubscribe waits for the
// daemon to send the SubscribeResponse. Pre-fix the read had no deadline,
// so a daemon that accepted but stalled before writing the ack would
// permanently freeze the next reconnect attempt — the TUI hung forever
// on the second request after any transient timeout.
const subscribeAckTimeout = 10 * time.Second

// dialAndSubscribe opens a fresh connection, sends the subscribe handshake,
// validates the ack (including protocol version), and starts a new readLoop
// bound to its OWN clientConn bundle.
func (c *EngineClient) dialAndSubscribe() error {
	conn, err := net.Dial("unix", c.socketPath)
	if err != nil {
		return fmt.Errorf("dial %s: %w", c.socketPath, err)
	}
	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	// Subscribe to every event kind so the client mirrors all engine events.
	// Passive=true marks this as a viewer connection: the TUI is not an
	// attached agent, so the engine must not flip into push-delivery mode
	// just because the reviewer's UI is open.
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
		return fmt.Errorf("encode subscribe: %w", err)
	}
	if _, err := conn.Write(data); err != nil {
		conn.Close()
		return fmt.Errorf("write subscribe: %w", err)
	}

	// Bound the ack read so a stalled daemon can't permanently wedge the
	// reconnect path. SetReadDeadline is cleared once the ack arrives.
	_ = conn.SetReadDeadline(time.Now().Add(subscribeAckTimeout))
	if !scanner.Scan() {
		conn.Close()
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("read subscribe ack: %w", err)
		}
		return errors.New("server closed before subscribe ack")
	}
	_ = conn.SetReadDeadline(time.Time{}) // clear deadline for steady-state reads
	ack, err := protocol.Decode(scanner.Bytes())
	if err != nil {
		conn.Close()
		return fmt.Errorf("decode subscribe ack: %w", err)
	}
	subAck, ok := ack.(*protocol.SubscribeResponse)
	if !ok {
		conn.Close()
		return fmt.Errorf("unexpected subscribe ack: %T", ack)
	}
	if subAck.ProtocolVersion < protocol.CurrentProtocolVersion {
		// Old daemon: silently drops Passive, would count this TUI as an
		// agent and route feedback to it instead of the real agent. We
		// refuse rather than silently mis-route. Caller should `monocle
		// stop` and respawn.
		conn.Close()
		return fmt.Errorf(
			"monocle serve protocol version %d is older than client (%d); run `monocle stop` and retry",
			subAck.ProtocolVersion, protocol.CurrentProtocolVersion,
		)
	}

	cc := &clientConn{
		conn:     conn,
		scanner:  scanner,
		pending:  make(chan any, 1),
		readErr:  make(chan error, 1),
		closed:   make(chan struct{}),
		readDone: make(chan struct{}),
	}

	c.connMu.Lock()
	c.active = cc
	c.connMu.Unlock()

	go c.readLoop(cc)
	return nil
}

// ensureConn re-dials if the active connection has been torn down (timeout,
// write error, server hang-up). Returns a nil error when the client is
// ready to send the next request.
func (c *EngineClient) ensureConn() error {
	c.connMu.Lock()
	cc := c.active
	c.connMu.Unlock()
	if cc == nil {
		return c.dialAndSubscribe()
	}
	select {
	case <-cc.closed:
		return c.dialAndSubscribe()
	default:
		return nil
	}
}

// readLoop demultiplexes incoming messages into events (dispatched locally)
// or request responses (sent on the connection's pending channel). The
// loop operates EXCLUSIVELY on its own clientConn bundle, never on the
// client's "current" fields — this is what stops a stale readLoop from
// tearing down a freshly-redialed connection (see the closeConn-race
// finding in PR #96's review).
func (c *EngineClient) readLoop(cc *clientConn) {
	defer close(cc.readDone)
	for cc.scanner.Scan() {
		line := cc.scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		msg, err := protocol.Decode(line)
		if err != nil {
			fmt.Fprintf(os.Stderr, "monocle client: dropping undecodable message (%v)\n", err)
			continue
		}
		if notif, ok := msg.(*protocol.EventNotification); ok {
			c.dispatchEvent(notif)
			continue
		}
		select {
		case cc.pending <- msg:
		case <-cc.closed:
			return
		case <-c.quitCh:
			return
		}
	}
	err := cc.scanner.Err()
	if err == nil {
		err = errors.New("connection closed")
	}
	select {
	case cc.readErr <- err:
	default:
	}
	// Close ONLY this connection — not whatever happens to be in
	// c.active at this moment. A redial may have already swapped in a
	// fresh clientConn by the time we reach here; that bundle is
	// completely independent of cc.
	cc.close()
}

// request sends msg and blocks until the corresponding response arrives.
// The reqMu lock guarantees one in-flight ordinary request at a time, so
// the order of non-event messages on the connection's pending channel
// matches the order of sends. On any terminal error (timeout, write
// error, server close) the connection is torn down and the next request
// lazily redials.
func (c *EngineClient) request(msg any) (any, error) {
	data, err := protocol.Encode(msg)
	if err != nil {
		return nil, fmt.Errorf("encode: %w", err)
	}

	c.reqMu.Lock()
	defer c.reqMu.Unlock()

	if err := c.ensureConn(); err != nil {
		return nil, err
	}

	c.connMu.Lock()
	cc := c.active
	c.connMu.Unlock()
	if cc == nil {
		return nil, errors.New("client closed")
	}

	c.writeMu.Lock()
	_, werr := cc.conn.Write(data)
	c.writeMu.Unlock()
	if werr != nil {
		cc.close()
		return nil, fmt.Errorf("write: %w", werr)
	}

	select {
	case resp := <-cc.pending:
		return resp, nil
	case err := <-cc.readErr:
		cc.close()
		return nil, err
	case <-cc.closed:
		return nil, errors.New("client closed")
	case <-c.quitCh:
		return nil, errors.New("client closed")
	case <-time.After(DefaultTimeout):
		cc.close()
		return nil, errors.New("timeout waiting for response")
	}
}

// requestNoTimeout is the timeout-free variant used by genuinely
// long-blocking calls (WaitForFeedback). It opens its OWN dedicated
// connection rather than serialising behind reqMu — otherwise a single
// hours-long Wait would block every other RPC on the shared client (and
// RequestPause, the very call that releases the wait, would self-deadlock).
//
// The connection is registered with the client so Close() can actively
// tear it down — pre-fix a blocked WaitForFeedback goroutine outlived
// Close indefinitely, leaking goroutines + daemon-side handler state.
func (c *EngineClient) requestNoTimeout(msg any) (any, error) {
	// Refuse to start a new no-timeout request if the client is shutting
	// down; otherwise a Close racing with a fresh call leaks the conn.
	select {
	case <-c.quitCh:
		return nil, errors.New("client closed")
	default:
	}

	conn, err := net.Dial("unix", c.socketPath)
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", c.socketPath, err)
	}

	c.noTimeoutMu.Lock()
	// Re-check quitCh under the lock so a Close that ran between the
	// quitCh check above and the lock acquisition can still cancel us.
	select {
	case <-c.quitCh:
		c.noTimeoutMu.Unlock()
		_ = conn.Close()
		return nil, errors.New("client closed")
	default:
	}
	c.noTimeoutConns[conn] = struct{}{}
	c.noTimeoutMu.Unlock()

	defer func() {
		c.noTimeoutMu.Lock()
		delete(c.noTimeoutConns, conn)
		c.noTimeoutMu.Unlock()
		_ = conn.Close()
	}()

	data, err := protocol.Encode(msg)
	if err != nil {
		return nil, fmt.Errorf("encode: %w", err)
	}
	if _, err := conn.Write(data); err != nil {
		return nil, fmt.Errorf("write: %w", err)
	}

	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("read: %w", err)
		}
		return nil, errors.New("server closed without response")
	}
	resp, err := protocol.Decode(scanner.Bytes())
	if err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	return resp, nil
}

// dispatchEvent hands the notification to the single dispatch worker.
// Blocks until eventCh accepts the message OR the client shuts down —
// pre-fix a non-blocking default branch silently dropped events when
// eventCh filled, breaking the per-subscriber 'no events lost' contract
// the comment promised. Backpressure now propagates correctly: readLoop
// stalls, which stalls the daemon's per-conn outbound queue, which (after
// its own 5s deadline) closes the slow consumer cleanly with an explicit
// error rather than silent data loss.
func (c *EngineClient) dispatchEvent(notif *protocol.EventNotification) {
	select {
	case c.eventCh <- notif:
	case <-c.quitCh:
		// Client is shutting down; drop the event.
	}
}

// dispatchWorker drains eventCh serially so per-subscriber callbacks see
// events in the same order the daemon emitted them. Exits on quitCh
// (Close) rather than channel close, because readLoop can still race
// against shutdown and would panic if we closed eventCh underneath it.
func (c *EngineClient) dispatchWorker() {
	defer close(c.dispatchDone)
	for {
		select {
		case notif := <-c.eventCh:
			c.invokeCallbacks(notif)
		case <-c.quitCh:
			return
		}
	}
}

func (c *EngineClient) invokeCallbacks(notif *protocol.EventNotification) {
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

// Close tears down the active socket, every outstanding no-timeout
// connection (e.g. a blocked WaitForFeedback), and the dispatch worker.
// Waits for all of those to exit before returning so tests / shutdown
// sequencing don't see lingering goroutines. Idempotent.
func (c *EngineClient) Close() error {
	c.quitOnce.Do(func() { close(c.quitCh) })

	c.connMu.Lock()
	cc := c.active
	c.active = nil
	c.connMu.Unlock()

	// Tear down every still-open no-timeout connection. Without this a
	// blocked WaitForFeedback (or any caller of requestNoTimeout) parks
	// on scanner.Scan forever, surviving Close.
	c.noTimeoutMu.Lock()
	conns := make([]net.Conn, 0, len(c.noTimeoutConns))
	for conn := range c.noTimeoutConns {
		conns = append(conns, conn)
	}
	c.noTimeoutMu.Unlock()
	for _, conn := range conns {
		_ = conn.Close()
	}

	if cc != nil {
		cc.close()
		<-cc.readDone
	}
	<-c.dispatchDone
	return nil
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
	r, ok := resp.(*protocol.StartSessionResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response %T", resp)
	}
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
	r, ok := resp.(*protocol.ResumeSessionResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response %T", resp)
	}
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
	r, ok := resp.(*protocol.GetSessionResponse)
	if !ok {
		return nil
	}
	return r.Session
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
	r, ok := resp.(*protocol.ListSessionsResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response %T", resp)
	}
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
	r, ok := resp.(*protocol.RefreshChangedFilesResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response %T", resp)
	}
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
	r, ok := resp.(*protocol.GetChangedFilesResponse)
	if !ok {
		return nil
	}
	return r.Files
}

func (c *EngineClient) GetFileDiff(path string) (*types.DiffResult, error) {
	resp, err := c.request(&protocol.GetFileDiffMsg{Type: protocol.TypeGetFileDiff, Path: path})
	if err != nil {
		return nil, err
	}
	r, ok := resp.(*protocol.GetFileDiffResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response %T", resp)
	}
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
	r, ok := resp.(*protocol.GetFileContentResponse)
	if !ok {
		return "", fmt.Errorf("unexpected response %T", resp)
	}
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
	r, ok := resp.(*protocol.GetContentItemsResponse)
	if !ok {
		return nil
	}
	return r.Items
}

func (c *EngineClient) GetContentItem(id string) (*types.ContentItem, error) {
	resp, err := c.request(&protocol.GetContentItemMsg{Type: protocol.TypeGetContentItem, ID: id})
	if err != nil {
		return nil, err
	}
	r, ok := resp.(*protocol.GetContentItemResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response %T", resp)
	}
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
	r, ok := resp.(*protocol.GetContentDiffResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response %T", resp)
	}
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
	r, ok := resp.(*protocol.GetContentVersionsResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response %T", resp)
	}
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
	r, ok := resp.(*protocol.GetContentDiffBetweenVersionsResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response %T", resp)
	}
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
	r, ok := resp.(*protocol.DismissArtifactResponse)
	if !ok {
		return fmt.Errorf("unexpected response %T", resp)
	}
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
	r, ok := resp.(*protocol.GetAdditionalFilesResponse)
	if !ok {
		return nil
	}
	return r.Files
}

func (c *EngineClient) AddAdditionalPaths(paths []string) ([]types.AdditionalFile, error) {
	resp, err := c.request(&protocol.AddAdditionalFilesMsg{
		Type:  protocol.TypeAddAdditionalFiles,
		Paths: paths,
	})
	if err != nil {
		return nil, err
	}
	r, ok := resp.(*protocol.AddAdditionalFilesResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response %T", resp)
	}
	if !r.Success {
		if r.Message != "" {
			return nil, errors.New(r.Message)
		}
		return nil, errors.New("add additional files failed")
	}
	// AddedPresent distinguishes "new daemon returned empty Added"
	// (legitimately added zero files) from "old daemon doesn't populate
	// Added at all". When the daemon honors the field we trust it; the
	// fallback to the cumulative GetAdditionalFiles only fires against a
	// truly old daemon — never on a new daemon returning an empty list.
	if r.AddedPresent {
		if r.Added == nil {
			return []types.AdditionalFile{}, nil
		}
		return r.Added, nil
	}
	return c.GetAdditionalFiles(), nil
}

func (c *EngineClient) GetAdditionalFileContent(absPath string) (string, error) {
	resp, err := c.request(&protocol.GetAdditionalFileContentMsg{Type: protocol.TypeGetAdditionalFileContent, AbsPath: absPath})
	if err != nil {
		return "", err
	}
	r, ok := resp.(*protocol.GetAdditionalFileContentResponse)
	if !ok {
		return "", fmt.Errorf("unexpected response %T", resp)
	}
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
	r, ok := resp.(*protocol.AddCommentResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response %T", resp)
	}
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
	r, ok := resp.(*protocol.EditCommentResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response %T", resp)
	}
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
	r, ok := resp.(*protocol.DeleteCommentResponse)
	if !ok {
		return fmt.Errorf("unexpected response %T", resp)
	}
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
	r, ok := resp.(*protocol.ResolveCommentResponse)
	if !ok {
		return fmt.Errorf("unexpected response %T", resp)
	}
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
	r, ok := resp.(*protocol.ClearCommentsResponse)
	if !ok {
		return fmt.Errorf("unexpected response %T", resp)
	}
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
	r, ok := resp.(*protocol.ClearReviewResponse)
	if !ok {
		return fmt.Errorf("unexpected response %T", resp)
	}
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
	r, ok := resp.(*protocol.MarkReviewedResponse)
	if !ok {
		return fmt.Errorf("unexpected response %T", resp)
	}
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
	r, ok := resp.(*protocol.UnmarkReviewedResponse)
	if !ok {
		return fmt.Errorf("unexpected response %T", resp)
	}
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
	r, ok := resp.(*protocol.MarkContentReviewedResponse)
	if !ok {
		return fmt.Errorf("unexpected response %T", resp)
	}
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
	r, ok := resp.(*protocol.UnmarkContentReviewedResponse)
	if !ok {
		return fmt.Errorf("unexpected response %T", resp)
	}
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
	r, ok := resp.(*protocol.ResetAllReviewedResponse)
	if !ok {
		return fmt.Errorf("unexpected response %T", resp)
	}
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
	r, ok := resp.(*protocol.MarkAllReviewedResponse)
	if !ok {
		return fmt.Errorf("unexpected response %T", resp)
	}
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
	r, ok := resp.(*protocol.GetReviewSummaryResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response %T", resp)
	}
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
	r, ok := resp.(*protocol.SubmitResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response %T", resp)
	}
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
	r, ok := resp.(*protocol.FormatReviewResponse)
	if !ok {
		return "", fmt.Errorf("unexpected response %T", resp)
	}
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
	r, ok := resp.(*protocol.GetSubmissionsResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response %T", resp)
	}
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
	r, ok := resp.(*protocol.SetBaseRefResponse)
	if !ok {
		return fmt.Errorf("unexpected response %T", resp)
	}
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
	r, ok := resp.(*protocol.IsAutoAdvanceRefResponse)
	if !ok {
		return false
	}
	return r.Enabled
}

func (c *EngineClient) SelectedBaseRef() string {
	resp, err := c.request(&protocol.SelectedBaseRefMsg{Type: protocol.TypeSelectedBaseRef})
	if err != nil {
		return ""
	}
	r, ok := resp.(*protocol.SelectedBaseRefResponse)
	if !ok {
		return ""
	}
	return r.Ref
}

func (c *EngineClient) RecentCommits(n int) ([]core.LogEntry, error) {
	resp, err := c.request(&protocol.RecentCommitsMsg{Type: protocol.TypeRecentCommits, Count: n})
	if err != nil {
		return nil, err
	}
	r, ok := resp.(*protocol.RecentCommitsResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response %T", resp)
	}
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
	r, ok := resp.(*protocol.GetSnapshotsResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response %T", resp)
	}
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
	r, ok := resp.(*protocol.SetSnapshotBaseResponse)
	if !ok {
		return fmt.Errorf("unexpected response %T", resp)
	}
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
	r, ok := resp.(*protocol.GetActiveSnapshotResponse)
	if !ok {
		return nil
	}
	return r.Snapshot
}

func (c *EngineClient) HasSnapshots() (bool, error) {
	resp, err := c.request(&protocol.HasSnapshotsMsg{Type: protocol.TypeHasSnapshots})
	if err != nil {
		return false, err
	}
	r, ok := resp.(*protocol.HasSnapshotsResponse)
	if !ok {
		return false, fmt.Errorf("unexpected response %T", resp)
	}
	if r.Error != "" {
		return false, errors.New(r.Error)
	}
	return r.Has, nil
}

// --- EngineAPI: server / feedback / status ---

// StartServer is a no-op on the client side — the serve process owns the
// socket. Kept so the client satisfies core.EngineAPI.
func (c *EngineClient) StartServer(_ string) error { return nil }

func (c *EngineClient) PollFeedback() *core.FormattedReview {
	resp, err := c.request(&protocol.PollFeedbackMsg{
		Type: protocol.TypePollFeedback,
		Wait: false,
	})
	if err != nil {
		return nil
	}
	r, ok := resp.(*protocol.PollFeedbackResponse)
	if !ok || !r.HasFeedback {
		return nil
	}
	return &core.FormattedReview{
		Formatted:    r.Feedback,
		CommentCount: r.CommentCount,
		Action:       r.Action,
	}
}

// WaitForFeedback blocks indefinitely (Wait=true) until the reviewer
// submits. Opens its own dedicated socket so the blocking call doesn't
// hold the main client's reqMu for hours and stall every other RPC —
// including RequestPause, the call that would release the wait.
func (c *EngineClient) WaitForFeedback() *core.FormattedReview {
	resp, err := c.requestNoTimeout(&protocol.PollFeedbackMsg{
		Type: protocol.TypePollFeedback,
		Wait: true,
	})
	if err != nil {
		return nil
	}
	r, ok := resp.(*protocol.PollFeedbackResponse)
	if !ok || !r.HasFeedback {
		return nil
	}
	return &core.FormattedReview{
		Formatted:    r.Feedback,
		CommentCount: r.CommentCount,
		Action:       r.Action,
	}
}

func (c *EngineClient) GetReviewStatusInfo() *core.ReviewStatusInfo {
	resp, err := c.request(&protocol.GetReviewStatusMsg{Type: protocol.TypeGetReviewStatus})
	if err != nil {
		return nil
	}
	r, ok := resp.(*protocol.GetReviewStatusResponse)
	if !ok {
		return nil
	}
	return &core.ReviewStatusInfo{
		Status:       r.Status,
		CommentCount: r.CommentCount,
		Summary:      r.Summary,
	}
}

func (c *EngineClient) SubmitContentForReview(id, title, content, contentType string, isPlan bool) error {
	resp, err := c.request(&protocol.SubmitContentMsg{
		Type:        protocol.TypeSubmitContent,
		ID:          id,
		Title:       title,
		Content:     content,
		ContentType: contentType,
		IsPlan:      isPlan,
	})
	if err != nil {
		return err
	}
	r, ok := resp.(*protocol.SubmitContentResponse)
	if !ok {
		return fmt.Errorf("unexpected response %T", resp)
	}
	if !r.Success {
		if r.Message != "" {
			return errors.New(r.Message)
		}
		return errors.New("submit content failed")
	}
	return nil
}

// RequestPause / CancelPause now propagate the socket error so the
// frontend can avoid lying to the user about whether the daemon's pause
// flag actually flipped.
func (c *EngineClient) RequestPause() error {
	_, err := c.request(&protocol.SetPauseMsg{Type: protocol.TypeSetPause, Requested: true})
	return err
}
func (c *EngineClient) CancelPause() error {
	_, err := c.request(&protocol.SetPauseMsg{Type: protocol.TypeSetPause, Requested: false})
	return err
}

func (c *EngineClient) GetFeedbackStatus() string {
	resp, err := c.request(&protocol.GetFeedbackStatusMsg{Type: protocol.TypeGetFeedbackStatus})
	if err != nil {
		return ""
	}
	r, ok := resp.(*protocol.GetFeedbackStatusResponse)
	if !ok {
		return ""
	}
	return r.Status
}

func (c *EngineClient) GetQueuedCount() int {
	resp, err := c.request(&protocol.GetQueuedCountMsg{Type: protocol.TypeGetQueuedCount})
	if err != nil {
		return 0
	}
	r, ok := resp.(*protocol.GetQueuedCountResponse)
	if !ok {
		return 0
	}
	return r.Count
}

func (c *EngineClient) ReloadPendingFeedback() {
	_, _ = c.request(&protocol.ReloadPendingFeedbackMsg{Type: protocol.TypeReloadPendingFeedback})
}

func (c *EngineClient) GetSubscriberCount() int {
	resp, err := c.request(&protocol.GetSubscriberCountMsg{Type: protocol.TypeGetSubscriberCount})
	if err != nil {
		return 0
	}
	r, ok := resp.(*protocol.GetSubscriberCountResponse)
	if !ok {
		return 0
	}
	return r.Count
}

func (c *EngineClient) GetSocketPath() string {
	resp, err := c.request(&protocol.GetSocketPathMsg{Type: protocol.TypeGetSocketPath})
	if err != nil {
		return ""
	}
	r, ok := resp.(*protocol.GetSocketPathResponse)
	if !ok {
		return ""
	}
	return r.Path
}

// --- EngineAPI: config ---

// GetConfig returns a cached pointer. The caller's documented flow is
// GetConfig() -> mutate -> SaveConfig(), so we cache the first fetch and
// return the same pointer on subsequent calls; refreshing would let an
// interleaving GetConfig overwrite the in-progress mutation pointer and
// silently drop user edits. The trade-off: changes made by a different
// client (a second TUI, `monocle register`) aren't visible until this
// EngineClient is reconstructed.
func (c *EngineClient) GetConfig() *types.Config {
	c.cfgMu.Lock()
	cached := c.cfg
	c.cfgMu.Unlock()
	if cached != nil {
		return cached
	}
	resp, err := c.request(&protocol.GetConfigMsg{Type: protocol.TypeGetConfig})
	if err != nil {
		return nil
	}
	r, ok := resp.(*protocol.GetConfigResponse)
	if !ok {
		return nil
	}
	c.cfgMu.Lock()
	// Re-check in case of race; first writer wins so the caller's
	// pointer identity is stable.
	if c.cfg == nil {
		c.cfg = r.Config
	}
	cfg := c.cfg
	c.cfgMu.Unlock()
	return cfg
}

// SaveConfig persists the cached config snapshot. Deep-copies via a JSON
// round-trip so a TUI goroutine mutating slice/map fields on the
// GetConfig-returned pointer can't race with the marshal — a concurrent
// map iteration / write would otherwise panic the encoder.
func (c *EngineClient) SaveConfig() error {
	c.cfgMu.Lock()
	cfg := c.cfg
	c.cfgMu.Unlock()
	if cfg == nil {
		return errors.New("SaveConfig called before GetConfig")
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	var safe types.Config
	if err := json.Unmarshal(data, &safe); err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}
	resp, err := c.request(&protocol.SaveConfigMsg{Type: protocol.TypeSaveConfig, Config: safe})
	if err != nil {
		return err
	}
	r, ok := resp.(*protocol.SaveConfigResponse)
	if !ok {
		return fmt.Errorf("unexpected response %T", resp)
	}
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
	r, ok := resp.(*protocol.IsReviewTrackingEnabledResponse)
	if !ok {
		return false
	}
	return r.Enabled
}

// Shutdown is a no-op on the client. The serve process manages its own
// lifecycle (including idle shutdown). Closing the client socket is the
// frontend's responsibility via Close.
func (c *EngineClient) Shutdown() {}

// Compile-time check: EngineClient satisfies core.EngineAPI.
var _ core.EngineAPI = (*EngineClient)(nil)

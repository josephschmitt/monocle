package core

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/josephschmitt/monocle/internal/protocol"
)

// DefaultIdleTimeout is how long monocle serve stays running past the 60s
// grace window after the last client disconnects. Zero disables idle
// shutdown (the serve runs until SIGINT/SIGTERM).
const DefaultIdleTimeout = 30 * time.Minute

// IdleGracePeriod is the fixed delay between "last client disconnected"
// and "start the idle countdown". Prevents thrashing when a user Ctrl-Cs
// a frontend and re-runs it within a few seconds.
const IdleGracePeriod = 60 * time.Second

// SocketServer listens on a Unix domain socket for CLI subcommand messages.
type SocketServer struct {
	listener        net.Listener
	engine          *Engine
	socketPath      string
	subscriberCount int
	queuedCount     int // active queue-mode connections (not counted in subscriberCount)
	oneshotCount    int // active one-shot request connections (CLI tool calls in flight)
	// lastDisconnectAt is the time the most recent client hung up leaving
	// the server with zero active connections. Zero value means the server
	// has always had at least one active client since startup.
	lastDisconnectAt time.Time
	subscriberMu     sync.Mutex

	// totalActiveConns counts every live accepted connection regardless of
	// type. It's the source of truth for idle detection, tracked at
	// accept/close boundaries so the monitor doesn't fire during the
	// narrow window before a new client's first message arrives.
	totalActiveConns int

	idleTimeout      time.Duration // 0 disables idle shutdown
	idleGrace        time.Duration // 0 → IdleGracePeriod; test hook only
	idleTickInterval time.Duration // 0 → 10s; test hook only
	idleStop         chan struct{} // closes when the server shuts down (stops the monitor goroutine)
	idleStopOnce     sync.Once     // guards close(idleStop) against double-close
	shutdownCh       chan struct{} // closes when idle timer fires
}

// NewSocketServer creates a new SocketServer. Call SetEngine and Start before use.
func NewSocketServer() *SocketServer {
	return &SocketServer{
		shutdownCh: make(chan struct{}),
		idleStop:   make(chan struct{}),
	}
}

// SetIdleTimeout configures how long the server stays alive past the 60s
// grace window after the last client disconnects. A zero or negative value
// disables idle shutdown.
func (s *SocketServer) SetIdleTimeout(d time.Duration) {
	s.idleTimeout = d
}

// IdleShutdownCh returns a channel that closes when the idle timer fires.
// Callers should listen on this alongside OS signals to know when to exit.
func (s *SocketServer) IdleShutdownCh() <-chan struct{} {
	return s.shutdownCh
}

// ActiveConnections returns the total number of live connections across all
// modes. Used by the idle monitor and exposed for tests/observability.
func (s *SocketServer) ActiveConnections() int {
	s.subscriberMu.Lock()
	defer s.subscriberMu.Unlock()
	return s.totalActiveConns
}

// SetEngine wires the engine to the server. Called during engine construction.
func (s *SocketServer) SetEngine(e *Engine) {
	s.engine = e
}

// Start begins listening on the given Unix domain socket path.
func (s *SocketServer) Start(socketPath string) error {
	// Probe socket: if something is listening, another monocle instance is live.
	conn, err := net.Dial("unix", socketPath)
	if err == nil {
		conn.Close()
		return fmt.Errorf("monocle is already running for this project (socket %s in use)", socketPath)
	}
	// Stale socket from a crashed process — safe to remove.
	_ = os.Remove(socketPath)

	l, err := net.Listen("unix", socketPath)
	if err != nil {
		return err
	}
	s.listener = l
	s.socketPath = socketPath

	go s.acceptLoop()
	if s.idleTimeout > 0 {
		go s.idleMonitor()
	}
	return nil
}

// idleMonitor periodically checks whether the server has been fully idle for
// grace+idleTimeout and, if so, closes shutdownCh. The main goroutine of
// monocle serve should select on IdleShutdownCh() alongside OS signals.
func (s *SocketServer) idleMonitor() {
	tick := s.idleTickInterval
	if tick == 0 {
		tick = 10 * time.Second
	}
	grace := s.idleGrace
	if grace == 0 {
		grace = IdleGracePeriod
	}
	ticker := time.NewTicker(tick)
	defer ticker.Stop()
	for {
		select {
		case <-s.idleStop:
			return
		case <-ticker.C:
			s.subscriberMu.Lock()
			active := s.totalActiveConns
			last := s.lastDisconnectAt
			s.subscriberMu.Unlock()

			if active > 0 || last.IsZero() {
				continue
			}
			if time.Since(last) >= grace+s.idleTimeout {
				close(s.shutdownCh)
				return
			}
		}
	}
}

// onDisconnect records a client departure. When it takes the active total
// to zero, the idle countdown starts.
func (s *SocketServer) onDisconnect() {
	s.subscriberMu.Lock()
	defer s.subscriberMu.Unlock()
	if s.totalActiveConns == 0 {
		s.lastDisconnectAt = time.Now()
	}
}

// onConnect cancels any in-flight idle countdown. Called when a new client
// appears so a quick UI restart within the grace window doesn't trip idle.
func (s *SocketServer) onConnect() {
	s.subscriberMu.Lock()
	defer s.subscriberMu.Unlock()
	s.lastDisconnectAt = time.Time{}
}

// SocketPath returns the path of the Unix domain socket.
func (s *SocketServer) SocketPath() string {
	return s.socketPath
}

// SubscriberCount returns the number of active subscriber connections.
func (s *SocketServer) SubscriberCount() int {
	s.subscriberMu.Lock()
	defer s.subscriberMu.Unlock()
	return s.subscriberCount
}

// Shutdown stops the server, halts the idle monitor, and removes the socket file.
func (s *SocketServer) Shutdown() error {
	// sync.Once guards against a panic when two callers race into Shutdown
	// (e.g. a SIGTERM handler and the idle-monitor-driven path both fire).
	s.idleStopOnce.Do(func() { close(s.idleStop) })
	if s.listener == nil {
		return nil
	}
	err := s.listener.Close()
	_ = os.Remove(s.socketPath)
	return err
}

func (s *SocketServer) acceptLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return // listener was closed
		}
		s.subscriberMu.Lock()
		s.totalActiveConns++
		s.subscriberMu.Unlock()
		s.onConnect()
		go func(c net.Conn) {
			defer func() {
				s.subscriberMu.Lock()
				s.totalActiveConns--
				s.subscriberMu.Unlock()
				s.onDisconnect()
			}()
			s.handleConnection(c)
		}(conn)
	}
}

// handleConnection reads the first NDJSON message to determine connection type.
// If it's a SubscribeMsg, the connection becomes persistent (bidirectional).
// Otherwise, it's a one-shot request/response as before.
func (s *SocketServer) handleConnection(conn net.Conn) {
	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	if !scanner.Scan() {
		conn.Close()
		return
	}

	line := scanner.Bytes()
	if len(line) == 0 {
		conn.Close()
		return
	}

	msg, err := protocol.Decode(line)
	if err != nil {
		conn.Close()
		return
	}

	// If the first message is a subscribe request, handle as persistent connection (push mode)
	if sub, ok := msg.(*protocol.SubscribeMsg); ok {
		s.handleSubscription(conn, scanner, sub)
		return
	}

	// If the first message is a connect request, handle as persistent connection
	// with event forwarding but without incrementing subscriberCount.
	if cm, ok := msg.(*protocol.ConnectMsg); ok {
		s.handleQueuedConnection(conn, scanner, cm)
		return
	}

	// One-shot request/response (backward compatible with CLI subcommands).
	// Track oneshotCount for observability; idle bookkeeping happens at
	// the accept/close boundary in acceptLoop.
	s.subscriberMu.Lock()
	s.oneshotCount++
	s.subscriberMu.Unlock()
	defer func() {
		s.subscriberMu.Lock()
		s.oneshotCount--
		s.subscriberMu.Unlock()
		conn.Close()
	}()

	response := s.handleMessage(msg)
	if response == nil {
		return
	}

	data, err := protocol.Encode(response)
	if err != nil {
		return
	}
	_, _ = conn.Write(data)
}

// handleSubscription manages a persistent subscription connection.
// It subscribes to requested engine events and pushes notifications,
// while also accepting request/response messages on the same connection.
func (s *SocketServer) handleSubscription(conn net.Conn, scanner *bufio.Scanner, sub *protocol.SubscribeMsg) {
	defer conn.Close()

	// Mutex for serialized writes to the connection
	var writeMu sync.Mutex
	writeMsg := func(msg any) error {
		data, err := protocol.Encode(msg)
		if err != nil {
			return err
		}
		writeMu.Lock()
		defer writeMu.Unlock()
		_, err = conn.Write(data)
		return err
	}

	// Subscribe to requested events before sending ack, so handlers are
	// registered by the time the client sees the ack and starts emitting.
	var unsubs []UnsubscribeFunc
	for _, eventName := range sub.Events {
		kind := EventKind(eventName)
		unsub := s.engine.On(kind, func(payload EventPayload) {
			if err := writeMsg(&protocol.EventNotification{
				Type:  protocol.TypeEventNotification,
				Event: string(payload.Kind),
				Payload: map[string]any{
					"message": payload.Message,
					"status":  payload.Status,
					"path":    payload.Path,
					"item_id": payload.ItemID,
				},
			}); err != nil {
				// Write failed — connection is dead. Close it to trigger
				// defer cleanup which decrements subscriberCount.
				conn.Close()
			}
		})
		unsubs = append(unsubs, unsub)
	}

	// Send ack
	if err := writeMsg(&protocol.SubscribeResponse{
		Type:    protocol.TypeSubscribeResponse,
		Success: true,
	}); err != nil {
		return
	}

	// Passive subscribers (e.g. the TUI) get events but don't count as
	// attached agents — Submit() must not flip into push mode just
	// because the reviewer's UI is open.
	if !sub.Passive {
		s.subscriberMu.Lock()
		s.subscriberCount++
		count := s.subscriberCount
		s.subscriberMu.Unlock()
		s.engine.emit(EventConnectionChanged, EventPayload{
			Kind:   EventConnectionChanged,
			Status: fmt.Sprintf("%d", count),
		})
	}

	// Clean up subscriptions (and subscriber count for non-passive subs) on exit.
	defer func() {
		if !sub.Passive {
			s.subscriberMu.Lock()
			s.subscriberCount--
			count := s.subscriberCount
			s.subscriberMu.Unlock()
			s.engine.emit(EventConnectionChanged, EventPayload{
				Kind:   EventConnectionChanged,
				Status: fmt.Sprintf("%d", count),
			})
		}
		for _, unsub := range unsubs {
			unsub()
		}
	}()

	// Read loop: incoming messages are request/response (tool calls)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		msg, err := protocol.Decode(line)
		if err != nil {
			continue
		}

		response := s.handleMessage(msg)
		if response != nil {
			_ = writeMsg(response)
		}
	}
}

// handleQueuedConnection manages a persistent connection that receives event
// notifications but does NOT increment subscriberCount. This means Submit()
// always queues feedback for pull delivery via get_feedback, while the client
// can still forward event notifications as channel hints (fire-and-forget).
func (s *SocketServer) handleQueuedConnection(conn net.Conn, scanner *bufio.Scanner, cm *protocol.ConnectMsg) {
	defer conn.Close()

	var writeMu sync.Mutex
	writeMsg := func(msg any) error {
		data, err := protocol.Encode(msg)
		if err != nil {
			return err
		}
		writeMu.Lock()
		defer writeMu.Unlock()
		_, err = conn.Write(data)
		return err
	}

	// Subscribe to requested events (like handleSubscription) but do NOT
	// increment subscriberCount. This allows event forwarding for channel
	// notifications without affecting the push/queue decision in Submit().
	var unsubs []UnsubscribeFunc
	for _, eventName := range cm.Events {
		kind := EventKind(eventName)
		unsub := s.engine.On(kind, func(payload EventPayload) {
			if err := writeMsg(&protocol.EventNotification{
				Type:  protocol.TypeEventNotification,
				Event: string(payload.Kind),
				Payload: map[string]any{
					"message": payload.Message,
					"status":  payload.Status,
					"path":    payload.Path,
					"item_id": payload.ItemID,
				},
			}); err != nil {
				conn.Close()
			}
		})
		unsubs = append(unsubs, unsub)
	}

	// Send ack
	if err := writeMsg(&protocol.ConnectResponse{
		Type:    protocol.TypeConnectResponse,
		Success: true,
	}); err != nil {
		return
	}

	// Track queue connection
	s.subscriberMu.Lock()
	s.queuedCount++
	s.subscriberMu.Unlock()

	// Notify TUI that an agent connected (without push subscription)
	s.engine.emit(EventConnectionChanged, EventPayload{
		Kind:   EventConnectionChanged,
		Status: "queue",
	})

	defer func() {
		s.subscriberMu.Lock()
		s.queuedCount--
		s.subscriberMu.Unlock()

		for _, unsub := range unsubs {
			unsub()
		}
		s.engine.emit(EventConnectionChanged, EventPayload{
			Kind:   EventConnectionChanged,
			Status: "0",
		})
	}()

	// Read loop: request/response + event notifications
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		msg, err := protocol.Decode(line)
		if err != nil {
			continue
		}

		response := s.handleMessage(msg)
		if response != nil {
			_ = writeMsg(response)
		}
	}
}

// handleMessage routes a decoded message to the appropriate engine handler.
func (s *SocketServer) handleMessage(msg any) any {
	switch m := msg.(type) {
	case *protocol.GetReviewStatusMsg:
		return s.engine.handleGetReviewStatus(m)
	case *protocol.PollFeedbackMsg:
		return s.engine.handlePollFeedback(m)
	case *protocol.SubmitContentMsg:
		return s.engine.handleSubmitContent(m)
	case *protocol.AddAdditionalFilesMsg:
		return s.engine.handleAddAdditionalFiles(m)
	case *protocol.MarkActivityMsg:
		return s.engine.handleMarkActivity(m)
	case *protocol.AwaitReviewMsg:
		return s.engine.handleAwaitReview(m)
	case *protocol.IdentifyMsg:
		s.handleIdentify(m)
		return nil

	// --- Engine surface (frontend clients) ---
	case *protocol.StartSessionMsg:
		return s.engine.handleStartSession(m)
	case *protocol.ResumeSessionMsg:
		return s.engine.handleResumeSession(m)
	case *protocol.GetSessionMsg:
		return s.engine.handleGetSession(m)
	case *protocol.ListSessionsMsg:
		return s.engine.handleListSessions(m)
	case *protocol.RefreshChangedFilesMsg:
		return s.engine.handleRefreshChangedFiles(m)
	case *protocol.GetChangedFilesMsg:
		return s.engine.handleGetChangedFiles(m)
	case *protocol.GetFileDiffMsg:
		return s.engine.handleGetFileDiff(m)
	case *protocol.GetFileContentMsg:
		return s.engine.handleGetFileContent(m)
	case *protocol.GetContentItemsMsg:
		return s.engine.handleGetContentItems(m)
	case *protocol.GetContentItemMsg:
		return s.engine.handleGetContentItem(m)
	case *protocol.GetContentDiffMsg:
		return s.engine.handleGetContentDiff(m)
	case *protocol.GetContentVersionsMsg:
		return s.engine.handleGetContentVersions(m)
	case *protocol.GetContentDiffBetweenVersionsMsg:
		return s.engine.handleGetContentDiffBetweenVersions(m)
	case *protocol.DismissArtifactMsg:
		return s.engine.handleDismissArtifact(m)
	case *protocol.GetAdditionalFilesMsg:
		return s.engine.handleGetAdditionalFiles(m)
	case *protocol.GetAdditionalFileContentMsg:
		return s.engine.handleGetAdditionalFileContent(m)
	case *protocol.AddCommentMsg:
		return s.engine.handleAddComment(m)
	case *protocol.EditCommentMsg:
		return s.engine.handleEditComment(m)
	case *protocol.DeleteCommentMsg:
		return s.engine.handleDeleteComment(m)
	case *protocol.ResolveCommentMsg:
		return s.engine.handleResolveComment(m)
	case *protocol.ClearCommentsMsg:
		return s.engine.handleClearComments(m)
	case *protocol.ClearReviewMsg:
		return s.engine.handleClearReview(m)
	case *protocol.MarkReviewedMsg:
		return s.engine.handleMarkReviewed(m)
	case *protocol.UnmarkReviewedMsg:
		return s.engine.handleUnmarkReviewed(m)
	case *protocol.MarkContentReviewedMsg:
		return s.engine.handleMarkContentReviewed(m)
	case *protocol.UnmarkContentReviewedMsg:
		return s.engine.handleUnmarkContentReviewed(m)
	case *protocol.ResetAllReviewedMsg:
		return s.engine.handleResetAllReviewed(m)
	case *protocol.MarkAllReviewedMsg:
		return s.engine.handleMarkAllReviewed(m)
	case *protocol.GetReviewSummaryMsg:
		return s.engine.handleGetReviewSummary(m)
	case *protocol.SubmitMsg:
		return s.engine.handleSubmit(m)
	case *protocol.FormatReviewMsg:
		return s.engine.handleFormatReview(m)
	case *protocol.GetSubmissionsMsg:
		return s.engine.handleGetSubmissions(m)
	case *protocol.SetBaseRefMsg:
		return s.engine.handleSetBaseRef(m)
	case *protocol.SetAutoAdvanceRefMsg:
		return s.engine.handleSetAutoAdvanceRef(m)
	case *protocol.IsAutoAdvanceRefMsg:
		return s.engine.handleIsAutoAdvanceRef(m)
	case *protocol.SelectedBaseRefMsg:
		return s.engine.handleSelectedBaseRef(m)
	case *protocol.RecentCommitsMsg:
		return s.engine.handleRecentCommits(m)
	case *protocol.GetSnapshotsMsg:
		return s.engine.handleGetSnapshots(m)
	case *protocol.SetSnapshotBaseMsg:
		return s.engine.handleSetSnapshotBase(m)
	case *protocol.ClearSnapshotBaseMsg:
		return s.engine.handleClearSnapshotBase(m)
	case *protocol.GetActiveSnapshotMsg:
		return s.engine.handleGetActiveSnapshot(m)
	case *protocol.HasSnapshotsMsg:
		return s.engine.handleHasSnapshots(m)
	case *protocol.GetConfigMsg:
		return s.engine.handleGetConfig(m)
	case *protocol.SaveConfigMsg:
		return s.engine.handleSaveConfig(m)
	case *protocol.IsReviewTrackingEnabledMsg:
		return s.engine.handleIsReviewTrackingEnabled(m)
	case *protocol.GetFeedbackStatusMsg:
		return s.engine.handleGetFeedbackStatus(m)
	case *protocol.GetQueuedCountMsg:
		return s.engine.handleGetQueuedCount(m)
	case *protocol.ReloadPendingFeedbackMsg:
		return s.engine.handleReloadPendingFeedback(m)
	case *protocol.GetSubscriberCountMsg:
		return s.engine.handleGetSubscriberCount(m)
	case *protocol.GetSocketPathMsg:
		return s.engine.handleGetSocketPath(m)

	default:
		return nil
	}
}

// handleIdentify processes an agent self-identification message and emits a
// connection event so the TUI can display the agent name.
func (s *SocketServer) handleIdentify(msg *protocol.IdentifyMsg) {
	s.subscriberMu.Lock()
	count := s.subscriberCount
	queued := s.queuedCount
	s.subscriberMu.Unlock()

	status := fmt.Sprintf("%d", count)
	if count == 0 && queued > 0 {
		status = "queue"
	}

	s.engine.emit(EventConnectionChanged, EventPayload{
		Kind:    EventConnectionChanged,
		Status:  status,
		Message: msg.Agent,
	})
}

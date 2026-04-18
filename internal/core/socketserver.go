package core

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/josephschmitt/monocle/internal/protocol"
)

// SocketServer listens on a Unix domain socket for CLI subcommand messages.
type SocketServer struct {
	listener        net.Listener
	engine          *Engine
	socketPath      string
	subscriberCount int
	queuedCount     int // active queue-mode connections (not counted in subscriberCount)
	subscriberMu    sync.Mutex
}

// NewSocketServer creates a new SocketServer. Call SetEngine and Start before use.
func NewSocketServer() *SocketServer {
	return &SocketServer{}
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
	return nil
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

// Shutdown stops the server and removes the socket file.
func (s *SocketServer) Shutdown() error {
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
		go s.handleConnection(conn)
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

	// One-shot request/response (backward compatible with CLI subcommands)
	defer conn.Close()

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

	// Track subscriber connection
	s.subscriberMu.Lock()
	s.subscriberCount++
	count := s.subscriberCount
	s.subscriberMu.Unlock()
	s.engine.emit(EventConnectionChanged, EventPayload{
		Kind:   EventConnectionChanged,
		Status: fmt.Sprintf("%d", count),
	})

	// Clean up subscriptions and subscriber count on exit
	defer func() {
		s.subscriberMu.Lock()
		s.subscriberCount--
		count := s.subscriberCount
		s.subscriberMu.Unlock()
		s.engine.emit(EventConnectionChanged, EventPayload{
			Kind:   EventConnectionChanged,
			Status: fmt.Sprintf("%d", count),
		})
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

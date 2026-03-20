package core

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/anthropics/monocle/internal/protocol"
)

// SocketServer listens on a Unix domain socket for CLI subcommand messages.
type SocketServer struct {
	listener   net.Listener
	engine     *Engine
	socketPath string
	done       chan struct{}
}

// NewSocketServer creates a new SocketServer. Call SetEngine and Start before use.
func NewSocketServer() *SocketServer {
	return &SocketServer{
		done: make(chan struct{}),
	}
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
			select {
			case <-s.done:
				return
			default:
				// listener was closed externally; stop accepting
				return
			}
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

	// If the first message is a subscribe request, handle as persistent connection
	if sub, ok := msg.(*protocol.SubscribeMsg); ok {
		s.handleSubscription(conn, scanner, sub)
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

	// Send ack
	if err := writeMsg(&protocol.SubscribeResponse{
		Type:    protocol.TypeSubscribeResponse,
		Success: true,
	}); err != nil {
		return
	}

	// Subscribe to requested events
	var unsubs []UnsubscribeFunc
	for _, eventName := range sub.Events {
		kind := EventKind(eventName)
		unsub := s.engine.On(kind, func(payload EventPayload) {
			_ = writeMsg(&protocol.EventNotification{
				Type:  protocol.TypeEventNotification,
				Event: string(payload.Kind),
				Payload: map[string]any{
					"message": payload.Message,
					"status":  payload.Status,
					"path":    payload.Path,
					"item_id": payload.ItemID,
				},
			})
		})
		unsubs = append(unsubs, unsub)
	}

	// Clean up subscriptions on exit
	defer func() {
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

// handleMessage routes a decoded message to the appropriate engine handler.
func (s *SocketServer) handleMessage(msg any) any {
	switch m := msg.(type) {
	case *protocol.GetReviewStatusMsg:
		return s.engine.handleGetReviewStatus(m)
	case *protocol.PollFeedbackMsg:
		return s.engine.handlePollFeedback(m)
	case *protocol.SubmitContentMsg:
		return s.engine.handleSubmitContent(m)
	default:
		return nil
	}
}

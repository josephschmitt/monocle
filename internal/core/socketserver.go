package core

import (
	"bufio"
	"fmt"
	"net"
	"os"

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

// handleConnection reads an NDJSON message from conn and routes it to the engine.
func (s *SocketServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	if !scanner.Scan() {
		return
	}

	line := scanner.Bytes()
	if len(line) == 0 {
		return
	}

	msg, err := protocol.Decode(line)
	if err != nil {
		return
	}

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

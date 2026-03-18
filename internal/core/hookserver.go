package core

import (
	"bufio"
	"net"
	"os"

	"github.com/anthropics/monocle/internal/protocol"
)

// HookServer listens on a Unix domain socket for hook messages from the shim.
type HookServer struct {
	listener   net.Listener
	engine     *Engine
	socketPath string
	done       chan struct{}
}

// NewHookServer creates a new HookServer. Call SetEngine and Start before use.
func NewHookServer() *HookServer {
	return &HookServer{
		done: make(chan struct{}),
	}
}

// SetEngine wires the engine to the server. Called during engine construction.
func (h *HookServer) SetEngine(e *Engine) {
	h.engine = e
}

// Start begins listening on the given Unix domain socket path.
func (h *HookServer) Start(socketPath string) error {
	// Remove a stale socket file if present.
	_ = os.Remove(socketPath)

	l, err := net.Listen("unix", socketPath)
	if err != nil {
		return err
	}
	h.listener = l
	h.socketPath = socketPath

	go h.acceptLoop()
	return nil
}

// SocketPath returns the path of the Unix domain socket.
func (h *HookServer) SocketPath() string {
	return h.socketPath
}

// Shutdown stops the server and removes the socket file.
func (h *HookServer) Shutdown() error {
	if h.listener == nil {
		return nil
	}
	err := h.listener.Close()
	_ = os.Remove(h.socketPath)
	return err
}

func (h *HookServer) acceptLoop() {
	for {
		conn, err := h.listener.Accept()
		if err != nil {
			select {
			case <-h.done:
				return
			default:
				// listener was closed externally; stop accepting
				return
			}
		}
		go h.handleConnection(conn)
	}
}

// handleConnection reads NDJSON messages from conn and routes them to the engine.
func (h *HookServer) handleConnection(conn net.Conn) {
	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		msg, err := protocol.Decode(line)
		if err != nil {
			conn.Close()
			return
		}

		if protocol.IsBlocking(msg) {
			h.handleBlocking(conn, msg)
			// After sending the response we're done with this connection.
			return
		}

		// Non-blocking: handle in goroutine and close the connection.
		go func(m any) {
			h.handleNonBlocking(m)
		}(msg)
		conn.Close()
		return
	}
	conn.Close()
}

// handleBlocking processes a message that requires a synchronous response.
func (h *HookServer) handleBlocking(conn net.Conn, msg any) {
	defer conn.Close()

	var response any
	switch m := msg.(type) {
	case *protocol.StopMsg:
		response = h.engine.handleStop(m)
	case *protocol.PromptSubmitMsg:
		response = h.engine.handlePromptSubmit(m)
	default:
		return
	}

	data, err := protocol.Encode(response)
	if err != nil {
		return
	}
	_, _ = conn.Write(data)
}

// handleNonBlocking processes a fire-and-forget message.
func (h *HookServer) handleNonBlocking(msg any) {
	switch m := msg.(type) {
	case *protocol.PostToolUseMsg:
		h.engine.handlePostToolUse(m)
	case *protocol.ContentReviewMsg:
		h.engine.handleContentReview(m)
	}
}

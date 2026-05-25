package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/josephschmitt/monocle/internal/adapters"
	"github.com/josephschmitt/monocle/internal/protocol"
	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// engineConn maintains a persistent socket connection to the Monocle engine
// for receiving push event notifications. It forwards events as MCP channel
// notifications through the captured stdio connection.
type engineConn struct {
	conn         mcp.Connection // MCP stdio connection for writing notifications
	agentName    string
	channelsOnly bool // true = reference CLI commands in notifications, false = reference tools
}

func newEngineConn(conn mcp.Connection, channelsOnly bool) *engineConn {
	return &engineConn{conn: conn, channelsOnly: channelsOnly}
}

// run connects to the engine and listens for events, forwarding them as
// channel notifications. It reconnects with backoff on connection loss.
func (e *engineConn) run(ctx context.Context) {
	socketPath := adapters.ResolveSocketPath()
	if socketPath == "" {
		return
	}

	const initialDelay = 2 * time.Second
	delay := initialDelay
	for {
		err := e.connectAndListen(ctx, socketPath)
		if ctx.Err() != nil {
			return
		}
		if err == nil {
			delay = initialDelay // reset after clean disconnect
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
			delay = min(delay*2, 10*time.Second)
		}
	}
}

// connectAndListen opens a socket connection, sends a connect message to
// subscribe to events, and forwards incoming events as channel notifications.
func (e *engineConn) connectAndListen(ctx context.Context, socketPath string) error {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	defer conn.Close()

	// Send connect message to subscribe to events
	connectMsg := protocol.ConnectMsg{
		Type: protocol.TypeConnect,
		Events: []string{
			"feedback_submitted",
			"pause_changed",
			"content_item_added",
			"additional_file_added",
		},
	}
	data, err := protocol.Encode(&connectMsg)
	if err != nil {
		return fmt.Errorf("encode connect: %w", err)
	}
	if _, err := conn.Write(data); err != nil {
		return fmt.Errorf("write connect: %w", err)
	}

	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	// Read connect response and check the daemon's protocol version. We
	// warn rather than hard-fail: tool calls go over separate one-shot
	// connections that still work against an older daemon, and returning an
	// error here would only spin the reconnect backoff. The warning tells
	// the operator to restart a stale serve (the TUI refuses it outright).
	if !scanner.Scan() {
		return fmt.Errorf("no connect response")
	}
	if ack, derr := protocol.Decode(scanner.Bytes()); derr == nil {
		if cr, ok := ack.(*protocol.ConnectResponse); ok && cr.ProtocolVersion < protocol.CurrentProtocolVersion {
			fmt.Fprintf(os.Stderr,
				"monocle: serve protocol version %d is older than this client (%d); "+
					"channel notifications may be degraded — run `monocle stop` and retry\n",
				cr.ProtocolVersion, protocol.CurrentProtocolVersion)
		}
	}

	// Send identify if we have an agent name
	if e.agentName != "" {
		e.sendIdentify(conn)
	}

	// Event loop
	for scanner.Scan() {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		msg, err := protocol.Decode(line)
		if err != nil {
			continue
		}

		if notif, ok := msg.(*protocol.EventNotification); ok {
			e.handleEvent(ctx, notif)
		}
	}

	return scanner.Err()
}

// handleEvent converts an engine event to an MCP channel notification.
func (e *engineConn) handleEvent(ctx context.Context, notif *protocol.EventNotification) {
	var content string
	var meta map[string]any

	switch notif.Event {
	case "feedback_submitted":
		msg, _ := notif.Payload["message"].(string)
		if msg == "" {
			if e.channelsOnly {
				msg = "Your reviewer has submitted feedback. Run `monocle review get-feedback` to retrieve it."
			} else {
				msg = "Your reviewer has submitted feedback. Use the get_feedback tool to retrieve it."
			}
		}
		content = msg
		meta = map[string]any{"event": "feedback_submitted"}

	case "pause_changed":
		status, _ := notif.Payload["status"].(string)
		if status != "pause_requested" {
			return
		}
		if e.channelsOnly {
			content = "Your reviewer has requested you pause and wait for feedback. " +
				"Run `monocle review get-feedback --wait` to block until feedback is ready."
		} else {
			content = "Your reviewer has requested you pause and wait for feedback. " +
				"Use the get_feedback tool with wait=true to block until feedback is ready."
		}
		meta = map[string]any{"event": "pause_requested"}

	default:
		return
	}

	e.sendChannelNotification(ctx, content, meta)
}

// sendChannelNotification sends a channel notification through the MCP connection.
func (e *engineConn) sendChannelNotification(ctx context.Context, content string, meta map[string]any) {
	params, err := json.Marshal(map[string]any{
		"content": content,
		"meta":    meta,
	})
	if err != nil {
		return
	}

	notif := &jsonrpc.Request{
		Method: "notifications/claude/channel",
		Params: params,
	}

	_ = e.conn.Write(ctx, notif)
}

// sendIdentify sends an identify message to the engine.
func (e *engineConn) sendIdentify(conn net.Conn) {
	msg := protocol.IdentifyMsg{
		Type:  protocol.TypeIdentify,
		Agent: e.agentName,
	}
	data, _ := protocol.Encode(&msg)
	conn.Write(data)
}

// identify sets the agent name and is called after the MCP handshake completes.
func (e *engineConn) identify(name string) {
	e.agentName = name
}

// close is a no-op — the engine socket connection is closed by connectAndListen's defer.
func (e *engineConn) close() {}

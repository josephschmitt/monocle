package adapters

import (
	"bufio"
	"fmt"
	"net"

	"github.com/anthropics/monocle/internal/protocol"
)

// SocketClient connects to the Monocle engine via Unix domain socket.
type SocketClient struct {
	conn    net.Conn
	scanner *bufio.Scanner
}

// NewSocketClient connects to the engine's Unix domain socket.
func NewSocketClient(socketPath string) (*SocketClient, error) {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("connect to %s: %w", socketPath, err)
	}
	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB buffer
	return &SocketClient{conn: conn, scanner: scanner}, nil
}

// Send encodes and sends a message without waiting for a response.
func (c *SocketClient) Send(msg any) error {
	data, err := protocol.Encode(msg)
	if err != nil {
		return err
	}
	_, err = c.conn.Write(data)
	return err
}

// SendAndWait sends a message and reads one response line.
func (c *SocketClient) SendAndWait(msg any) (any, error) {
	if err := c.Send(msg); err != nil {
		return nil, err
	}
	if !c.scanner.Scan() {
		if err := c.scanner.Err(); err != nil {
			return nil, fmt.Errorf("read response: %w", err)
		}
		return nil, fmt.Errorf("connection closed")
	}
	return protocol.Decode(c.scanner.Bytes())
}

// Close closes the connection.
func (c *SocketClient) Close() error {
	return c.conn.Close()
}

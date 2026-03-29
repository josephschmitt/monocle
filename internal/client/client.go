// Package client provides a socket client for communicating with a running
// Monocle engine via its Unix domain socket.
package client

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/josephschmitt/monocle/internal/adapters"
	"github.com/josephschmitt/monocle/internal/protocol"
)

// ErrNotRunning is returned when the Monocle socket is not reachable.
var ErrNotRunning = errors.New("monocle is not running — start it with 'monocle'")

// Client communicates with a running Monocle engine over a Unix domain socket.
type Client struct {
	conn    net.Conn
	scanner *bufio.Scanner
}

// Connect dials the Unix domain socket at the given path.
func Connect(socketPath string) (*Client, error) {
	if _, err := os.Stat(socketPath); errors.Is(err, os.ErrNotExist) {
		return nil, ErrNotRunning
	}

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, ErrNotRunning
	}

	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	return &Client{conn: conn, scanner: scanner}, nil
}

// ConnectDefault resolves the socket path from the current working directory
// and connects. Respects the MONOCLE_SOCKET environment variable.
func ConnectDefault() (*Client, error) {
	socketPath := os.Getenv("MONOCLE_SOCKET")
	if socketPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("get cwd: %w", err)
		}
		repoRoot := adapters.FindRepoRoot(cwd)
		socketPath = adapters.DefaultSocketPath(repoRoot)
	}
	return Connect(socketPath)
}

// ConnectWithOverride connects using the override path if non-empty, otherwise
// falls back to ConnectDefault.
func ConnectWithOverride(socketOverride string) (*Client, error) {
	if socketOverride != "" {
		return Connect(socketOverride)
	}
	return ConnectDefault()
}

// Request sends a protocol message and reads the response. The caller provides
// a timeout; use 0 for no deadline (blocking operations).
func (c *Client) Request(msg any, timeout time.Duration) (any, error) {
	data, err := protocol.Encode(msg)
	if err != nil {
		return nil, fmt.Errorf("encode: %w", err)
	}
	if _, err := c.conn.Write(data); err != nil {
		return nil, fmt.Errorf("write: %w", err)
	}

	if timeout > 0 {
		c.conn.SetReadDeadline(time.Now().Add(timeout))
	} else {
		c.conn.SetReadDeadline(time.Time{}) // no deadline
	}

	if !c.scanner.Scan() {
		if err := c.scanner.Err(); err != nil {
			return nil, fmt.Errorf("read: %w", err)
		}
		return nil, errors.New("connection closed by server")
	}

	resp, err := protocol.Decode(c.scanner.Bytes())
	if err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	return resp, nil
}

// DefaultTimeout is the read deadline for non-blocking requests.
const DefaultTimeout = 30 * time.Second

// Close closes the underlying connection.
func (c *Client) Close() error {
	return c.conn.Close()
}

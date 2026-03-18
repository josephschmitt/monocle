package main

import (
	"flag"
	"io"
	"os"

	"github.com/anthropics/monocle/internal/adapters"
	"github.com/anthropics/monocle/internal/protocol"
)

func main() {
	fs := flag.NewFlagSet("monocle-hook", flag.ContinueOnError)
	agent := fs.String("agent", "claude", "agent name")

	// Parse flags, ignoring errors — on any error, exit 0.
	if err := fs.Parse(os.Args[1:]); err != nil {
		os.Exit(0)
	}

	args := fs.Args()
	if len(args) == 0 {
		os.Exit(0)
	}
	event := args[0]

	// Read stdin completely.
	raw, err := io.ReadAll(os.Stdin)
	if err != nil {
		os.Exit(0)
	}

	// Parse the input via the adapter.
	adapter := adapters.GetAdapter(*agent)
	msg, err := adapter.ParseHookInput(event, raw)
	if err != nil {
		os.Exit(0)
	}

	// Get the socket path from the environment.
	socketPath := os.Getenv("MONOCLE_SOCKET")
	if socketPath == "" {
		os.Exit(0)
	}

	// Connect to the engine.
	client, err := adapters.NewSocketClient(socketPath)
	if err != nil {
		os.Exit(0)
	}
	defer client.Close()

	if protocol.IsBlocking(msg) {
		// Send and wait for a response.
		response, err := client.SendAndWait(msg)
		if err != nil {
			os.Exit(0)
		}

		out := adapter.FormatHookOutput(response)
		if len(out.Data) > 0 {
			os.Stdout.Write(out.Data) //nolint:errcheck
		}
		os.Exit(out.ExitCode)
	}

	// Non-blocking: fire and forget.
	_ = client.Send(msg)
	os.Exit(0)
}

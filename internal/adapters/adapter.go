package adapters

// AgentAdapter defines how to parse/format messages for a specific agent.
type AgentAdapter interface {
	// ParseHookInput parses raw hook stdin into a protocol message.
	ParseHookInput(event string, raw []byte) (any, error)

	// FormatHookOutput converts a protocol response to the agent's expected format.
	FormatHookOutput(response any) HookOutput

	// GenerateConfig produces the agent's hook configuration.
	GenerateConfig(opts SetupOptions) string

	// Capabilities returns what hook events this agent supports.
	Capabilities() AdapterCapabilities
}

// HookOutput is what gets written to the hook's stdout.
type HookOutput struct {
	Data     []byte
	ExitCode int
}

// SetupOptions configures hook generation.
type SetupOptions struct {
	HookBinaryPath string
	SocketPath     string
	ProjectLevel   bool
}

// AdapterCapabilities describes what a specific agent supports.
type AdapterCapabilities struct {
	PostToolUse  bool
	StopBlocking bool
	AsyncHooks   bool
}

// GetAdapter returns the adapter for the given agent name.
func GetAdapter(agent string) AgentAdapter {
	switch agent {
	case "claude":
		return &ClaudeAdapter{}
	default:
		return &ClaudeAdapter{} // default to Claude
	}
}

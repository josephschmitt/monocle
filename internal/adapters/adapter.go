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

// AgentInstaller manages hook configuration for a specific agent.
type AgentInstaller interface {
	Name() string
	Detect() bool
	ConfigPath(global bool) (string, bool)
	IsInstalled(configPath string) (bool, error)
	Install(configPath string, opts InstallOptions) error
	Uninstall(configPath string) error
}

// InstallOptions configures hook installation.
type InstallOptions struct {
	HookBinaryPath string
}

// InstallResult reports what happened for a single agent.
type InstallResult struct {
	Agent       string
	ConfigPath  string
	Installed   bool
	AlreadyDone bool
	Err         error
}

// GetAdapter returns the adapter for the given agent name.
func GetAdapter(agent string) AgentAdapter {
	switch agent {
	case "claude":
		return &ClaudeAdapter{}
	case "gemini":
		return &GeminiAdapter{}
	case "codex":
		return &CodexAdapter{}
	case "opencode":
		return &OpenCodeAdapter{}
	default:
		return &ClaudeAdapter{} // default to Claude
	}
}

// AllInstallers returns installers for all supported agents.
func AllInstallers() []AgentInstaller {
	return []AgentInstaller{
		&ClaudeAdapter{},
		&GeminiAdapter{},
		&CodexAdapter{},
		&OpenCodeAdapter{},
	}
}

// GetInstaller returns the installer for the given agent name, or nil.
func GetInstaller(name string) AgentInstaller {
	for _, i := range AllInstallers() {
		if i.Name() == name {
			return i
		}
	}
	return nil
}

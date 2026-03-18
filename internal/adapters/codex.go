package adapters

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/anthropics/monocle/internal/protocol"
)

// CodexAdapter handles Codex CLI's hook format.
// Codex only supports global config via TOML and a single event (agent-turn-complete).
type CodexAdapter struct{}

func (a *CodexAdapter) ParseHookInput(event string, raw []byte) (any, error) {
	switch event {
	case "agent-turn-complete":
		return &protocol.StopMsg{
			Type:  protocol.TypeStop,
			Agent: "codex",
		}, nil
	default:
		return nil, fmt.Errorf("unknown event: %s", event)
	}
}

func (a *CodexAdapter) FormatHookOutput(response any) HookOutput {
	return HookOutput{ExitCode: 0}
}

func (a *CodexAdapter) GenerateConfig(opts SetupOptions) string {
	hookCmd := opts.HookBinaryPath
	if hookCmd == "" {
		hookCmd = "monocle"
	}

	cmd := fmt.Sprintf(`sh -c 'echo "$1" | %s hook agent-turn-complete --agent codex' --`, hookCmd)
	return fmt.Sprintf("[notify]\nagent-turn-complete = %q\n", cmd)
}

func (a *CodexAdapter) Capabilities() AdapterCapabilities {
	return AdapterCapabilities{
		PostToolUse:  false,
		StopBlocking: false,
		AsyncHooks:   false,
	}
}

// AgentInstaller implementation

func (a *CodexAdapter) Name() string { return "codex" }

func (a *CodexAdapter) Detect() bool {
	if _, err := exec.LookPath("codex"); err == nil {
		return true
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	if info, err := os.Stat(filepath.Join(home, ".codex")); err == nil && info.IsDir() {
		return true
	}
	return false
}

// ConfigPath returns the config path. Codex only supports global config.
func (a *CodexAdapter) ConfigPath(global bool) (string, bool) {
	if !global {
		return "", false // Codex has no project-level config
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", false
	}
	return filepath.Join(home, ".codex", "config.toml"), true
}

func (a *CodexAdapter) IsInstalled(configPath string) (bool, error) {
	config, err := readTOMLFile(configPath)
	if err != nil {
		return false, err
	}
	notify, ok := config["notify"].(map[string]any)
	if !ok {
		return false, nil
	}
	cmd, _ := notify["agent-turn-complete"].(string)
	return isMonocleCommand(cmd), nil
}

func (a *CodexAdapter) Install(configPath string, opts InstallOptions) error {
	hookCmd := opts.HookBinaryPath
	if hookCmd == "" {
		hookCmd = "monocle"
	}
	scopeFlag := scopeFlagStr(opts.Scope)

	config, err := readTOMLFile(configPath)
	if err != nil {
		return err
	}

	notify, ok := config["notify"].(map[string]any)
	if !ok {
		notify = map[string]any{}
		config["notify"] = notify
	}

	// Check if already installed
	existing, _ := notify["agent-turn-complete"].(string)
	if isMonocleCommand(existing) {
		return nil
	}

	cmd := fmt.Sprintf(`sh -c 'echo "$1" | %s hook%s agent-turn-complete --agent codex' --`, hookCmd, scopeFlag)
	notify["agent-turn-complete"] = cmd

	return writeTOMLFile(configPath, config)
}

func (a *CodexAdapter) Uninstall(configPath string) error {
	config, err := readTOMLFile(configPath)
	if err != nil {
		return err
	}

	notify, ok := config["notify"].(map[string]any)
	if !ok {
		return nil
	}

	cmd, _ := notify["agent-turn-complete"].(string)
	if !isMonocleCommand(cmd) {
		return nil
	}

	delete(notify, "agent-turn-complete")
	if len(notify) == 0 {
		delete(config, "notify")
	}

	return writeTOMLFile(configPath, config)
}

// readTOMLFile reads a TOML file into a map. Returns empty map if file doesn't exist.
func readTOMLFile(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]any{}, nil
		}
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	if len(data) == 0 {
		return map[string]any{}, nil
	}
	var m map[string]any
	if err := toml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return m, nil
}

// writeTOMLFile atomically writes a map as TOML to path, creating parent dirs.
func writeTOMLFile(path string, data map[string]any) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create dir %s: %w", dir, err)
	}

	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}

	encoder := toml.NewEncoder(f)
	if err := encoder.Encode(data); err != nil {
		f.Close()
		os.Remove(tmp)
		return fmt.Errorf("encode toml: %w", err)
	}
	f.Close()

	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("rename %s to %s: %w", tmp, path, err)
	}
	return nil
}

// Ensure CodexAdapter implements both interfaces.
var (
	_ AgentAdapter   = (*CodexAdapter)(nil)
	_ AgentInstaller = (*CodexAdapter)(nil)
)

// Ensure ClaudeAdapter implements both interfaces.
var (
	_ AgentAdapter   = (*ClaudeAdapter)(nil)
	_ AgentInstaller = (*ClaudeAdapter)(nil)
)

// Ensure GeminiAdapter implements both interfaces.
var (
	_ AgentAdapter   = (*GeminiAdapter)(nil)
	_ AgentInstaller = (*GeminiAdapter)(nil)
)

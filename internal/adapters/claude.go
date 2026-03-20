package adapters

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// ClaudeAdapter handles Claude Code MCP channel installation.
type ClaudeAdapter struct{}

func (a *ClaudeAdapter) Name() string { return "claude" }

// Detect returns true if Claude Code appears to be installed.
func (a *ClaudeAdapter) Detect() bool {
	if _, err := exec.LookPath("claude"); err == nil {
		return true
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	if info, err := os.Stat(filepath.Join(home, ".claude")); err == nil && info.IsDir() {
		return true
	}
	return false
}

// Install writes channel.ts and configures .mcp.json.
func (a *ClaudeAdapter) Install() error {
	if err := a.installChannel(); err != nil {
		return fmt.Errorf("install channel: %w", err)
	}
	if err := a.configureMCP(); err != nil {
		return fmt.Errorf("configure mcp: %w", err)
	}
	return nil
}

// Uninstall removes channel.ts and unconfigures .mcp.json.
func (a *ClaudeAdapter) Uninstall() error {
	channelPath := channelTSPath()
	if channelPath != "" {
		if err := os.Remove(channelPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove channel.ts: %w", err)
		}
		os.Remove(filepath.Dir(channelPath)) // clean up empty parent dir
	}

	if err := a.unconfigureMCP(); err != nil {
		return fmt.Errorf("unconfigure mcp: %w", err)
	}
	return nil
}

// IsInstalled checks if the MCP channel is configured.
func (a *ClaudeAdapter) IsInstalled() (bool, error) {
	channelPath := channelTSPath()
	if channelPath == "" {
		return false, nil
	}
	_, err := os.Stat(channelPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// InstallDetails returns additional info about what was installed.
func (a *ClaudeAdapter) InstallDetails() []string {
	var details []string
	channelPath := channelTSPath()
	if channelPath != "" {
		details = append(details, fmt.Sprintf("channel → %s", channelPath))
	}
	details = append(details, "mcp → .mcp.json")

	if _, err := exec.LookPath("bun"); err != nil {
		details = append(details, "⚠ bun not found in PATH — install bun for MCP channel support")
	}

	return details
}

// installChannel writes channel.ts to the XDG config directory.
func (a *ClaudeAdapter) installChannel() error {
	path := channelTSPath()
	if path == "" {
		return fmt.Errorf("cannot determine config directory")
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(ChannelContent), 0644)
}

// configureMCP adds monocle to .mcp.json in the current project.
func (a *ClaudeAdapter) configureMCP() error {
	mcpPath := ".mcp.json"
	data, err := ReadJSONFile(mcpPath)
	if err != nil {
		return err
	}

	servers, ok := data["mcpServers"].(map[string]any)
	if !ok {
		servers = map[string]any{}
		data["mcpServers"] = servers
	}

	channelPath := channelTSPath()
	if channelPath == "" {
		return fmt.Errorf("cannot determine channel.ts path")
	}

	servers["monocle"] = map[string]any{
		"command": "bun",
		"args":    []any{channelPath},
	}

	return WriteJSONFile(mcpPath, data)
}

// unconfigureMCP removes monocle from .mcp.json.
func (a *ClaudeAdapter) unconfigureMCP() error {
	mcpPath := ".mcp.json"
	data, err := ReadJSONFile(mcpPath)
	if err != nil {
		return err
	}

	servers, ok := data["mcpServers"].(map[string]any)
	if !ok {
		return nil
	}

	delete(servers, "monocle")

	if len(servers) == 0 {
		return os.Remove(mcpPath)
	}

	return WriteJSONFile(mcpPath, data)
}

// channelTSPath returns the path for channel.ts in the XDG config directory.
func channelTSPath() string {
	cfgDir := os.Getenv("XDG_CONFIG_HOME")
	if cfgDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		cfgDir = filepath.Join(home, ".config")
	}
	return filepath.Join(cfgDir, "monocle", "channel.ts")
}

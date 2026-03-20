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
// If global is true, .mcp.json is written to ~/.mcp.json instead of the project.
func (a *ClaudeAdapter) Install(global bool) error {
	if err := a.installChannel(); err != nil {
		return fmt.Errorf("install channel: %w", err)
	}
	if err := a.configureMCP(global); err != nil {
		return fmt.Errorf("configure mcp: %w", err)
	}
	return nil
}

// Uninstall removes channel.ts, deps, and unconfigures .mcp.json.
// If global is true, removes from ~/.mcp.json instead of the project.
func (a *ClaudeAdapter) Uninstall(global bool) error {
	channelPath := channelTSPath()
	if channelPath != "" {
		dir := filepath.Dir(channelPath)
		// Remove the entire channel directory (channel.ts, package.json, node_modules, bun.lock)
		if err := os.RemoveAll(dir); err != nil {
			return fmt.Errorf("remove channel directory: %w", err)
		}
	}

	if err := a.unconfigureMCP(global); err != nil {
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
func (a *ClaudeAdapter) InstallDetails(global bool) []string {
	var details []string
	channelPath := channelTSPath()
	if channelPath != "" {
		details = append(details, fmt.Sprintf("channel → %s", channelPath))
	}
	details = append(details, fmt.Sprintf("mcp → %s", mcpJSONPath(global)))

	if _, err := exec.LookPath("bun"); err != nil {
		details = append(details, "⚠ bun not found in PATH — install bun for MCP channel support")
	}

	return details
}

// packageJSON is the package.json for the channel's npm dependencies.
const packageJSON = `{
  "name": "monocle-channel",
  "private": true,
  "dependencies": {
    "@modelcontextprotocol/sdk": "^1.12.1"
  }
}
`

// installChannel writes channel.ts, package.json, and installs deps.
func (a *ClaudeAdapter) installChannel() error {
	path := channelTSPath()
	if path == "" {
		return fmt.Errorf("cannot determine config directory")
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(ChannelContent), 0644); err != nil {
		return err
	}

	// Write package.json for channel dependencies
	pkgPath := filepath.Join(dir, "package.json")
	if err := os.WriteFile(pkgPath, []byte(packageJSON), 0644); err != nil {
		return fmt.Errorf("write package.json: %w", err)
	}

	// Run bun install to fetch dependencies
	bunPath, err := exec.LookPath("bun")
	if err != nil {
		return fmt.Errorf("bun not found in PATH — install bun for MCP channel support")
	}
	cmd := exec.Command(bunPath, "install")
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("bun install failed: %w\n%s", err, output)
	}

	return nil
}

// configureMCP adds monocle to .mcp.json.
// If global is true, writes to ~/.mcp.json; otherwise to ./.mcp.json.
func (a *ClaudeAdapter) configureMCP(global bool) error {
	mcpPath := mcpJSONPath(global)
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
// If global is true, operates on ~/.mcp.json; otherwise ./.mcp.json.
func (a *ClaudeAdapter) unconfigureMCP(global bool) error {
	mcpPath := mcpJSONPath(global)
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

// mcpJSONPath returns the path for .mcp.json.
// If global is true, returns ~/.mcp.json; otherwise ./.mcp.json.
func mcpJSONPath(global bool) string {
	if global {
		home, err := os.UserHomeDir()
		if err != nil {
			return ".mcp.json"
		}
		return filepath.Join(home, ".mcp.json")
	}
	return ".mcp.json"
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

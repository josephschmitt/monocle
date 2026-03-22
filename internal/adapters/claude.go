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
		// Remove the entire channel directory (channel.ts, package.json, node_modules, lock files)
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

// HasMCPConfig checks if monocle is configured in any .mcp.json (global or local).
func (a *ClaudeAdapter) HasMCPConfig() bool {
	for _, global := range []bool{true, false} {
		path := mcpJSONPath(global)
		data, err := ReadJSONFile(path)
		if err != nil {
			continue
		}
		servers, ok := data["mcpServers"].(map[string]any)
		if !ok {
			continue
		}
		if _, ok := servers["monocle"]; ok {
			return true
		}
	}
	return false
}

// NeedsInstall returns true if the MCP channel is not fully set up.
// It checks that both channel.ts exists AND at least one .mcp.json has a monocle entry.
func (a *ClaudeAdapter) NeedsInstall() bool {
	installed, err := a.IsInstalled()
	if err != nil || !installed {
		return true
	}
	return !a.HasMCPConfig()
}

// InstallDetails returns additional info about what was installed.
func (a *ClaudeAdapter) InstallDetails(global bool) []string {
	var details []string
	channelPath := channelTSPath()
	if channelPath != "" {
		details = append(details, fmt.Sprintf("channel → %s", channelPath))
	}
	details = append(details, fmt.Sprintf("mcp → %s", mcpJSONPath(global)))

	rt, err := detectJSRuntime()
	if err != nil {
		details = append(details, "⚠ no JavaScript runtime found — install bun, deno, or node")
	} else {
		details = append(details, fmt.Sprintf("runtime → %s", rt.name))
	}

	return details
}

// jsRuntime represents a detected JavaScript runtime.
type jsRuntime struct {
	name string // "bun", "deno", "node"
}

// detectJSRuntime finds the first available runtime in preference order: bun, deno, node.
func detectJSRuntime() (*jsRuntime, error) {
	for _, name := range []string{"bun", "deno", "node"} {
		if _, err := exec.LookPath(name); err == nil {
			return &jsRuntime{name: name}, nil
		}
	}
	return nil, fmt.Errorf("no JavaScript runtime found (install bun, deno, or node)")
}

// installDeps runs the runtime-appropriate dependency install command.
func (r *jsRuntime) installDeps(dir string) error {
	var cmd *exec.Cmd
	switch r.name {
	case "bun":
		cmd = exec.Command("bun", "install")
	case "deno":
		cmd = exec.Command("deno", "install")
	case "node":
		cmd = exec.Command("npm", "install")
	}
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s install failed: %w\n%s", r.name, err, output)
	}
	return nil
}

// mcpCommand returns the command and args for .mcp.json.
func (r *jsRuntime) mcpCommand(channelPath string) (string, []any) {
	switch r.name {
	case "deno":
		return "deno", []any{"run", "--allow-all", channelPath}
	case "node":
		return "npx", []any{"tsx", channelPath}
	default: // bun
		return "bun", []any{channelPath}
	}
}

// packageJSON is the package.json for the channel's npm dependencies.
// tsx is included for Node.js compatibility (bun/deno ignore it).
const packageJSON = `{
  "name": "monocle-channel",
  "private": true,
  "dependencies": {
    "@modelcontextprotocol/sdk": "^1.12.1",
    "tsx": "^4.0.0"
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

	// Detect runtime and install dependencies
	rt, err := detectJSRuntime()
	if err != nil {
		return err
	}
	return rt.installDeps(dir)
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

	rt, err := detectJSRuntime()
	if err != nil {
		return err
	}
	command, args := rt.mcpCommand(channelPath)
	servers["monocle"] = map[string]any{
		"command": command,
		"args":    args,
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

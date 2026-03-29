package adapters

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ClaudeAdapter handles Claude Code MCP channel registration.
type ClaudeAdapter struct{}

func (a *ClaudeAdapter) Name() string  { return "claude" }
func (a *ClaudeAdapter) Label() string { return "Claude Code" }

// ConfigPaths returns the files written by Register.
func (a *ClaudeAdapter) ConfigPaths(global bool) []string {
	paths := []string{mcpJSONPath(global)}
	paths = append(paths, SkillPaths(claudeSkillsDir(global))...)
	return paths
}

// HasConfig returns true if monocle is configured at the given scope via .mcp.json or Claude Code plugin.
func (a *ClaudeAdapter) HasConfig(global bool) bool {
	if hasMCPServersEntry(mcpJSONPath(global)) {
		return true
	}
	// Plugin config is always user-level (global)
	if global {
		return a.HasPluginConfig()
	}
	return false
}

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

// Register adds monocle to .mcp.json and installs skill files.
func (a *ClaudeAdapter) Register(global bool) error {
	if err := a.configureMCP(global); err != nil {
		return fmt.Errorf("configure mcp: %w", err)
	}
	if err := InstallSkills(claudeSkillsDir(global)); err != nil {
		return fmt.Errorf("install skills: %w", err)
	}
	return nil
}

// Unregister removes monocle from .mcp.json and removes skill files.
func (a *ClaudeAdapter) Unregister(global bool) error {
	if err := a.unconfigureMCP(global); err != nil {
		return fmt.Errorf("unconfigure mcp: %w", err)
	}
	RemoveSkills(claudeSkillsDir(global))
	return nil
}

// HasMCPConfig checks if monocle is correctly configured in any .mcp.json (global or local).
// Returns true only if the entry uses the serve-mcp-channel subcommand (not old-style bun/node configs).
func (a *ClaudeAdapter) HasMCPConfig() bool {
	for _, global := range []bool{true, false} {
		if hasMCPServersEntry(mcpJSONPath(global)) {
			return true
		}
	}
	return false
}

// HasPluginConfig checks if monocle is installed as a Claude Code plugin.
// Returns true if ~/.claude/plugins/installed_plugins.json contains a monocle@* entry
// with user scope (global) or project scope matching the current working directory.
func (a *ClaudeAdapter) HasPluginConfig() bool {
	path := installedPluginsPath()
	data, err := ReadJSONFile(path)
	if err != nil {
		return false
	}
	plugins, ok := data["plugins"].(map[string]any)
	if !ok {
		return false
	}

	cwd, _ := os.Getwd()
	if resolved, err := filepath.EvalSymlinks(cwd); err == nil {
		cwd = resolved
	}

	for key, val := range plugins {
		if !strings.HasPrefix(key, "monocle@") {
			continue
		}
		entries, ok := val.([]any)
		if !ok {
			continue
		}
		for _, e := range entries {
			entry, ok := e.(map[string]any)
			if !ok {
				continue
			}
			scope, _ := entry["scope"].(string)
			if scope == "user" {
				return true
			}
			if scope == "project" {
				projectPath, _ := entry["projectPath"].(string)
				if resolved, err := filepath.EvalSymlinks(projectPath); err == nil {
					projectPath = resolved
				}
				if projectPath != "" && cwd != "" && strings.HasPrefix(cwd, projectPath) {
					return true
				}
			}
		}
	}
	return false
}

// NeedsRegister returns true if monocle is not configured via .mcp.json or Claude Code plugin.
// This includes cases where an old-style config exists (e.g., pointing to bun/node directly).
func (a *ClaudeAdapter) NeedsRegister() bool {
	return !a.HasMCPConfig() && !a.HasPluginConfig()
}

// RegisterDetails returns info about what was registered.
func (a *ClaudeAdapter) RegisterDetails(global bool) []string {
	var details []string
	details = append(details, fmt.Sprintf("mcp → %s", mcpJSONPath(global)))

	rt, err := DetectJSRuntime()
	if err != nil {
		details = append(details, "warning: no JavaScript runtime found — install bun, deno, or node")
	} else {
		details = append(details, fmt.Sprintf("runtime → %s", rt.Name))
	}

	return details
}

// JSRuntime represents a detected JavaScript runtime.
type JSRuntime struct {
	Name string // "bun", "deno", "node"
}

// DetectJSRuntime finds the first available runtime in preference order: bun, deno, node.
func DetectJSRuntime() (*JSRuntime, error) {
	for _, name := range []string{"bun", "deno", "node"} {
		if _, err := exec.LookPath(name); err == nil {
			return &JSRuntime{Name: name}, nil
		}
	}
	return nil, fmt.Errorf("no JavaScript runtime found (install bun, deno, or node)")
}

// ExecArgs returns the binary path and argv for running the given script file.
// The binary path is an absolute path suitable for syscall.Exec.
func (r *JSRuntime) ExecArgs(scriptPath string) (string, []string, error) {
	binPath, err := exec.LookPath(r.Name)
	if err != nil {
		return "", nil, fmt.Errorf("lookup %s: %w", r.Name, err)
	}
	switch r.Name {
	case "deno":
		return binPath, []string{"deno", "run", "--allow-all", scriptPath}, nil
	default: // bun, node
		return binPath, []string{r.Name, scriptPath}, nil
	}
}

// ChannelBundlePath returns the path where the channel bundle should be written.
// Uses a content hash so the file is only written once per version.
func ChannelBundlePath() string {
	hash := sha256.Sum256(ChannelBundle)
	hexHash := hex.EncodeToString(hash[:])[:8]
	return filepath.Join(os.TempDir(), fmt.Sprintf("monocle-channel-%s.mjs", hexHash))
}

// WriteChannelBundle writes the embedded channel bundle to its temp path if needed.
// Returns the path to the written file.
func WriteChannelBundle() (string, error) {
	path := ChannelBundlePath()
	if _, err := os.Stat(path); err == nil {
		return path, nil // already exists with correct hash
	}
	if err := os.WriteFile(path, ChannelBundle, 0644); err != nil {
		return "", fmt.Errorf("write channel bundle: %w", err)
	}
	return path, nil
}

// configureMCP adds monocle to .mcp.json.
func (a *ClaudeAdapter) configureMCP(global bool) error {
	return configureMCPServersJSON(mcpJSONPath(global), ResolveCommand(global))
}

// unconfigureMCP removes monocle from .mcp.json.
func (a *ClaudeAdapter) unconfigureMCP(global bool) error {
	return unconfigureMCPServersJSON(mcpJSONPath(global))
}

// configureMCPServersJSON adds monocle to a JSON file with an "mcpServers" key.
// Used by Claude (.mcp.json) and Gemini (.gemini/settings.json).
func configureMCPServersJSON(path, command string) error {
	data, err := ReadJSONFile(path)
	if err != nil {
		return err
	}

	servers, ok := data["mcpServers"].(map[string]any)
	if !ok {
		servers = map[string]any{}
		data["mcpServers"] = servers
	}

	servers["monocle"] = map[string]any{
		"command": command,
		"args":    []any{"serve-mcp-channel"},
	}

	return WriteJSONFile(path, data)
}

// unconfigureMCPServersJSON removes monocle from a JSON file with an "mcpServers" key.
func unconfigureMCPServersJSON(path string) error {
	data, err := ReadJSONFile(path)
	if err != nil {
		return err
	}

	servers, ok := data["mcpServers"].(map[string]any)
	if !ok {
		return nil
	}

	delete(servers, "monocle")

	if len(servers) == 0 {
		delete(data, "mcpServers")
	}

	if len(data) == 0 {
		return RemoveFileIfExists(path)
	}

	return WriteJSONFile(path, data)
}

// hasMCPServersEntry checks if a JSON file has a monocle entry under "mcpServers"
// with the serve-mcp-channel subcommand.
func hasMCPServersEntry(path string) bool {
	data, err := ReadJSONFile(path)
	if err != nil {
		return false
	}
	servers, ok := data["mcpServers"].(map[string]any)
	if !ok {
		return false
	}
	entry, ok := servers["monocle"].(map[string]any)
	if !ok {
		return false
	}
	args, _ := entry["args"].([]any)
	if len(args) > 0 {
		if arg, ok := args[0].(string); ok && arg == "serve-mcp-channel" {
			return true
		}
	}
	return false
}

// installedPluginsPath returns the path to Claude Code's installed plugins registry.
func installedPluginsPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude", "plugins", "installed_plugins.json")
}

// claudeSkillsDir returns the directory for Claude Code skill files.
func claudeSkillsDir(global bool) string {
	if global {
		home, err := os.UserHomeDir()
		if err != nil {
			return filepath.Join(".claude", "skills")
		}
		return filepath.Join(home, ".claude", "skills")
	}
	return filepath.Join(".claude", "skills")
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


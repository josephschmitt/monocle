package adapters

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// ClaudeAdapter handles Claude Code MCP channel registration.
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

// Register adds monocle to .mcp.json.
// If global is true, .mcp.json is written to ~/.mcp.json instead of the project.
func (a *ClaudeAdapter) Register(global bool) error {
	if err := a.configureMCP(global); err != nil {
		return fmt.Errorf("configure mcp: %w", err)
	}
	return nil
}

// Unregister removes monocle from .mcp.json.
// If global is true, removes from ~/.mcp.json instead of the project.
func (a *ClaudeAdapter) Unregister(global bool) error {
	if err := a.unconfigureMCP(global); err != nil {
		return fmt.Errorf("unconfigure mcp: %w", err)
	}
	return nil
}

// HasMCPConfig checks if monocle is correctly configured in any .mcp.json (global or local).
// Returns true only if the entry uses the serve-mcp-channel subcommand (not old-style bun/node configs).
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
		entry, ok := servers["monocle"].(map[string]any)
		if !ok {
			continue
		}
		// Validate the entry points to serve-mcp-channel
		args, _ := entry["args"].([]any)
		if len(args) > 0 {
			if arg, ok := args[0].(string); ok && arg == "serve-mcp-channel" {
				return true
			}
		}
	}
	return false
}

// NeedsRegister returns true if the MCP channel is not correctly configured in any .mcp.json.
// This includes cases where an old-style config exists (e.g., pointing to bun/node directly).
func (a *ClaudeAdapter) NeedsRegister() bool {
	return !a.HasMCPConfig()
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

	command := "monocle"
	if global {
		// Use absolute path for global config (machine-specific)
		if exePath, err := os.Executable(); err == nil {
			command = exePath
		}
	}

	servers["monocle"] = map[string]any{
		"command": command,
		"args":    []any{"serve-mcp-channel"},
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


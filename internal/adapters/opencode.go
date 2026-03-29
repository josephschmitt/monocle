package adapters

import (
	"os"
	"os/exec"
	"path/filepath"
)

// OpenCodeAdapter handles Monocle registration for OpenCode.
// Installs skill files only — no MCP server needed.
type OpenCodeAdapter struct{}

func (a *OpenCodeAdapter) Name() string  { return "opencode" }
func (a *OpenCodeAdapter) Label() string { return "OpenCode" }

func (a *OpenCodeAdapter) ConfigPaths(global bool) []string {
	return SkillPaths(openCodeSkillsDir(global))
}

func (a *OpenCodeAdapter) HasConfig(global bool) bool {
	// Check for skill files
	dir := openCodeSkillsDir(global)
	for _, name := range SkillNames {
		if _, err := os.Stat(filepath.Join(dir, name, "SKILL.md")); err == nil {
			return true
		}
	}
	// Also detect legacy MCP config
	return hasLegacyOpenCodeMCP(global)
}

func (a *OpenCodeAdapter) Register(global bool) error {
	// Clean up legacy MCP config if present
	removeLegacyOpenCodeMCP(global)

	// Install skill files
	return InstallSkills(openCodeSkillsDir(global))
}

func (a *OpenCodeAdapter) Unregister(global bool) error {
	// Remove legacy MCP config if present
	removeLegacyOpenCodeMCP(global)

	// Remove skill files
	RemoveSkills(openCodeSkillsDir(global))

	// Also remove legacy command files
	cmdDir := openCodeCommandsDir(global)
	for _, name := range SkillNames {
		_ = RemoveFileIfExists(filepath.Join(cmdDir, name+".md"))
	}

	return nil
}

// Detect returns true if OpenCode appears to be installed.
func (a *OpenCodeAdapter) Detect() bool {
	if _, err := exec.LookPath("opencode"); err == nil {
		return true
	}
	if _, err := os.Stat("opencode.json"); err == nil {
		return true
	}
	return false
}

func openCodeSkillsDir(global bool) string {
	if global {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, ".config", "opencode", "skills")
		}
	}
	return filepath.Join(".opencode", "skills")
}

func openCodeCommandsDir(global bool) string {
	if global {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, ".config", "opencode", "commands")
		}
	}
	return filepath.Join(".opencode", "commands")
}

func openCodeConfigPath(global bool) string {
	if global {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, ".config", "opencode", "opencode.json")
		}
	}
	return "opencode.json"
}

// hasLegacyOpenCodeMCP checks if the old MCP config exists.
func hasLegacyOpenCodeMCP(global bool) bool {
	path := openCodeConfigPath(global)
	data, err := ReadJSONFile(path)
	if err != nil {
		return false
	}
	mcp, ok := data["mcp"].(map[string]any)
	if !ok {
		return false
	}
	_, ok = mcp["monocle"].(map[string]any)
	return ok
}

// removeLegacyOpenCodeMCP removes the old MCP server entry and plan permission.
func removeLegacyOpenCodeMCP(global bool) {
	path := openCodeConfigPath(global)
	data, err := ReadJSONFile(path)
	if err != nil {
		return
	}
	changed := false
	if mcp, ok := data["mcp"].(map[string]any); ok {
		if _, ok := mcp["monocle"]; ok {
			delete(mcp, "monocle")
			if len(mcp) == 0 {
				delete(data, "mcp")
			}
			changed = true
		}
	}
	if agent, ok := data["agent"].(map[string]any); ok {
		if plan, ok := agent["plan"].(map[string]any); ok {
			if perm, ok := plan["permission"].(map[string]any); ok {
				if _, ok := perm["mcp__monocle"]; ok {
					delete(perm, "mcp__monocle")
					if len(perm) == 0 {
						delete(plan, "permission")
					}
					changed = true
				}
			}
		}
	}
	if changed {
		_ = WriteJSONFile(path, data)
	}
}

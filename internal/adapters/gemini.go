package adapters

import (
	"os"
	"os/exec"
	"path/filepath"
)

// GeminiAdapter handles Monocle registration for Google Gemini CLI.
// Installs skill files only — no MCP server needed.
type GeminiAdapter struct{}

func (a *GeminiAdapter) Name() string  { return "gemini" }
func (a *GeminiAdapter) Label() string { return "Gemini CLI" }

func (a *GeminiAdapter) ConfigPaths(global bool) []string {
	return SkillPaths(geminiSkillsDir(global))
}

func (a *GeminiAdapter) HasConfig(global bool) bool {
	// Check for skill files
	dir := geminiSkillsDir(global)
	for _, name := range SkillNames {
		if _, err := os.Stat(filepath.Join(dir, name, "SKILL.md")); err == nil {
			return true
		}
	}
	// Also detect legacy MCP config
	return hasMCPServersEntry(geminiConfigPath(global))
}

func (a *GeminiAdapter) Register(global bool) error {
	// Clean up legacy MCP config if present
	_ = unconfigureMCPServersJSON(geminiConfigPath(global))

	// Remove legacy command files
	cmdDir := geminiCommandsDir(global)
	for _, name := range SkillNames {
		_ = RemoveFileIfExists(filepath.Join(cmdDir, name+".toml"))
	}

	// Install skill files
	return InstallSkills(geminiSkillsDir(global))
}

func (a *GeminiAdapter) Unregister(global bool) error {
	// Remove legacy MCP config if present
	_ = unconfigureMCPServersJSON(geminiConfigPath(global))

	// Remove legacy command files
	cmdDir := geminiCommandsDir(global)
	for _, name := range SkillNames {
		_ = RemoveFileIfExists(filepath.Join(cmdDir, name+".toml"))
	}

	// Remove skill files
	RemoveSkills(geminiSkillsDir(global))

	return nil
}

// Detect returns true if Gemini CLI appears to be installed.
func (a *GeminiAdapter) Detect() bool {
	if _, err := exec.LookPath("gemini"); err == nil {
		return true
	}
	if _, err := os.Stat(".gemini"); err == nil {
		return true
	}
	return false
}

func geminiSkillsDir(global bool) string {
	if global {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, ".gemini", "skills")
		}
	}
	return filepath.Join(".gemini", "skills")
}

func geminiConfigPath(global bool) string {
	if global {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, ".gemini", "settings.json")
		}
	}
	return filepath.Join(".gemini", "settings.json")
}

func geminiCommandsDir(global bool) string {
	if global {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, ".gemini", "commands")
		}
	}
	return filepath.Join(".gemini", "commands")
}

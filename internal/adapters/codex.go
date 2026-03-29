package adapters

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CodexAdapter handles Monocle registration for OpenAI Codex CLI.
// Installs skill files only — no MCP server needed.
type CodexAdapter struct{}

func (a *CodexAdapter) Name() string  { return "codex" }
func (a *CodexAdapter) Label() string { return "Codex CLI" }

func (a *CodexAdapter) ConfigPaths(global bool) []string {
	return SkillPaths(codexSkillsDir(global))
}

func (a *CodexAdapter) HasConfig(global bool) bool {
	// Check for skill files
	dir := codexSkillsDir(global)
	for _, name := range SkillNames {
		if _, err := os.Stat(filepath.Join(dir, name, "SKILL.md")); err == nil {
			return true
		}
	}
	// Also detect legacy MCP config
	return hasLegacyCodexMCP(global)
}

func (a *CodexAdapter) Register(global bool) error {
	// Clean up legacy MCP config if present
	removeLegacyCodexMCP(global)

	// Install skill files
	return InstallSkills(codexSkillsDir(global))
}

func (a *CodexAdapter) Unregister(global bool) error {
	// Remove legacy MCP config if present
	removeLegacyCodexMCP(global)

	// Remove skill files
	RemoveSkills(codexSkillsDir(global))

	return nil
}

// Detect returns true if Codex CLI appears to be installed.
func (a *CodexAdapter) Detect() bool {
	if _, err := exec.LookPath("codex"); err == nil {
		return true
	}
	if _, err := os.Stat(".codex"); err == nil {
		return true
	}
	return false
}

func codexSkillsDir(global bool) string {
	if global {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, ".codex", "skills")
		}
	}
	return filepath.Join(".codex", "skills")
}

func codexConfigPath(global bool) string {
	if global {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, ".codex", "config.toml")
		}
	}
	return filepath.Join(".codex", "config.toml")
}

// hasLegacyCodexMCP checks for the old MCP server TOML config.
func hasLegacyCodexMCP(global bool) bool {
	content, err := os.ReadFile(codexConfigPath(global))
	if err != nil {
		return false
	}
	return strings.Contains(string(content), "[mcp_servers.monocle]")
}

// removeLegacyCodexMCP removes the old MCP server TOML section.
func removeLegacyCodexMCP(global bool) {
	path := codexConfigPath(global)
	content, err := os.ReadFile(path)
	if err != nil {
		return
	}
	if !strings.Contains(string(content), "[mcp_servers.monocle]") {
		return
	}
	cleaned := strings.TrimRight(removeMonocleTOMLSection(string(content)), "\n") + "\n"
	if strings.TrimSpace(cleaned) == "" {
		_ = RemoveFileIfExists(path)
	} else {
		_ = WriteFileAtomic(path, []byte(cleaned))
	}
}

// removeMonocleTOMLSection strips the [mcp_servers.monocle] section and its
// key-value lines from a TOML document string.
func removeMonocleTOMLSection(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	inSection := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "[mcp_servers.monocle]" {
			inSection = true
			continue
		}
		if inSection {
			if strings.HasPrefix(trimmed, "[") {
				inSection = false
				result = append(result, line)
			}
			continue
		}
		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

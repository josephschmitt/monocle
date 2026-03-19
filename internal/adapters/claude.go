package adapters

import (
	"os"
	"os/exec"
	"path/filepath"
)

// ClaudeAdapter handles Claude Code skill installation.
type ClaudeAdapter struct{}

var _ SkillInstaller = (*ClaudeAdapter)(nil)

func (a *ClaudeAdapter) Name() string { return "claude" }

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

func (a *ClaudeAdapter) SkillPath(global bool) (string, bool) {
	if global {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", false
		}
		return filepath.Join(home, ".claude", "skills", "monocle-review", "SKILL.md"), true
	}
	return filepath.Join(".claude", "skills", "monocle-review", "SKILL.md"), true
}

func (a *ClaudeAdapter) IsInstalled(skillPath string) (bool, error) {
	_, err := os.Stat(skillPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (a *ClaudeAdapter) Install(skillPath string) error {
	dir := filepath.Dir(skillPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(skillPath, []byte(SkillContent), 0644)
}

func (a *ClaudeAdapter) Uninstall(skillPath string) error {
	if err := os.Remove(skillPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	// Clean up empty parent directories
	dir := filepath.Dir(skillPath)
	os.Remove(dir) // remove monocle-review dir if empty
	return nil
}

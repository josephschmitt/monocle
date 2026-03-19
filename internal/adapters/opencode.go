package adapters

import (
	"os"
	"os/exec"
	"path/filepath"
)

// OpenCodeAdapter handles OpenCode skill installation.
type OpenCodeAdapter struct{}

var _ SkillInstaller = (*OpenCodeAdapter)(nil)

func (a *OpenCodeAdapter) Name() string { return "opencode" }

func (a *OpenCodeAdapter) Detect() bool {
	if _, err := exec.LookPath("opencode"); err == nil {
		return true
	}
	// Check for project-level config
	if _, err := os.Stat("opencode.json"); err == nil {
		return true
	}
	return false
}

func (a *OpenCodeAdapter) SkillPath(global bool) (string, bool) {
	if global {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", false
		}
		return filepath.Join(home, ".config", "opencode", "skills", "monocle-review", "SKILL.md"), true
	}
	return filepath.Join(".opencode", "skills", "monocle-review", "SKILL.md"), true
}

func (a *OpenCodeAdapter) IsInstalled(skillPath string) (bool, error) {
	_, err := os.Stat(skillPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (a *OpenCodeAdapter) Install(skillPath string) error {
	dir := filepath.Dir(skillPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(skillPath, []byte(SkillContent), 0644)
}

func (a *OpenCodeAdapter) Uninstall(skillPath string) error {
	if err := os.Remove(skillPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	dir := filepath.Dir(skillPath)
	os.Remove(dir)
	return nil
}

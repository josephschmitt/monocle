package adapters

import (
	"os"
	"os/exec"
	"path/filepath"
)

// CodexAdapter handles Codex CLI skill installation.
type CodexAdapter struct{}

var _ SkillInstaller = (*CodexAdapter)(nil)

func (a *CodexAdapter) Name() string { return "codex" }

func (a *CodexAdapter) Detect() bool {
	if _, err := exec.LookPath("codex"); err == nil {
		return true
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	if info, err := os.Stat(filepath.Join(home, ".codex")); err == nil && info.IsDir() {
		return true
	}
	return false
}

// SkillPath returns the skill file path. Codex only supports global config.
func (a *CodexAdapter) SkillPath(global bool) (string, bool) {
	if !global {
		return "", false // Codex has no project-level config
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", false
	}
	return filepath.Join(home, ".codex", "skills", "monocle-review", "SKILL.md"), true
}

func (a *CodexAdapter) IsInstalled(skillPath string) (bool, error) {
	_, err := os.Stat(skillPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (a *CodexAdapter) Install(skillPath string) error {
	dir := filepath.Dir(skillPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(skillPath, []byte(SkillContent), 0644)
}

func (a *CodexAdapter) Uninstall(skillPath string) error {
	if err := os.Remove(skillPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	dir := filepath.Dir(skillPath)
	os.Remove(dir)
	return nil
}

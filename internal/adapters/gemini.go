package adapters

import (
	"os"
	"os/exec"
	"path/filepath"
)

// GeminiAdapter handles Gemini CLI skill installation.
type GeminiAdapter struct{}

var _ SkillInstaller = (*GeminiAdapter)(nil)

func (a *GeminiAdapter) Name() string { return "gemini" }

func (a *GeminiAdapter) Detect() bool {
	if _, err := exec.LookPath("gemini"); err == nil {
		return true
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	if info, err := os.Stat(filepath.Join(home, ".gemini")); err == nil && info.IsDir() {
		return true
	}
	return false
}

func (a *GeminiAdapter) SkillPath(global bool) (string, bool) {
	if global {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", false
		}
		return filepath.Join(home, ".gemini", "skills", "monocle-review", "SKILL.md"), true
	}
	return filepath.Join(".gemini", "skills", "monocle-review", "SKILL.md"), true
}

func (a *GeminiAdapter) IsInstalled(skillPath string) (bool, error) {
	_, err := os.Stat(skillPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (a *GeminiAdapter) Install(skillPath string) error {
	dir := filepath.Dir(skillPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(skillPath, []byte(SkillContent), 0644)
}

func (a *GeminiAdapter) Uninstall(skillPath string) error {
	if err := os.Remove(skillPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	dir := filepath.Dir(skillPath)
	os.Remove(dir)
	return nil
}

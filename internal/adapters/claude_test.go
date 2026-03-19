package adapters

import (
	"os"
	"path/filepath"
	"testing"
)

func TestClaudeSkillInstall(t *testing.T) {
	dir := t.TempDir()
	skillPath := filepath.Join(dir, ".claude", "skills", "monocle-review", "SKILL.md")

	adapter := &ClaudeAdapter{}

	// Not installed initially
	installed, err := adapter.IsInstalled(skillPath)
	if err != nil {
		t.Fatalf("IsInstalled error: %v", err)
	}
	if installed {
		t.Fatal("should not be installed initially")
	}

	// Install
	if err := adapter.Install(skillPath); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	// Verify installed
	installed, err = adapter.IsInstalled(skillPath)
	if err != nil {
		t.Fatalf("IsInstalled error: %v", err)
	}
	if !installed {
		t.Fatal("should be installed after Install()")
	}

	// Verify content
	data, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("read skill file: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("skill file should not be empty")
	}
}

func TestClaudeSkillInstall_Idempotent(t *testing.T) {
	dir := t.TempDir()
	skillPath := filepath.Join(dir, ".claude", "skills", "monocle-review", "SKILL.md")

	adapter := &ClaudeAdapter{}
	if err := adapter.Install(skillPath); err != nil {
		t.Fatalf("first install: %v", err)
	}
	if err := adapter.Install(skillPath); err != nil {
		t.Fatalf("second install: %v", err)
	}

	installed, _ := adapter.IsInstalled(skillPath)
	if !installed {
		t.Fatal("should be installed")
	}
}

func TestClaudeSkillUninstall(t *testing.T) {
	dir := t.TempDir()
	skillPath := filepath.Join(dir, ".claude", "skills", "monocle-review", "SKILL.md")

	adapter := &ClaudeAdapter{}
	adapter.Install(skillPath)
	adapter.Uninstall(skillPath)

	installed, _ := adapter.IsInstalled(skillPath)
	if installed {
		t.Fatal("should not be installed after uninstall")
	}
}

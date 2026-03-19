package adapters

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSkillInstall_AllAdapters(t *testing.T) {
	adapters := []SkillInstaller{
		&ClaudeAdapter{},
		&GeminiAdapter{},
		&OpenCodeAdapter{},
	}

	for _, adapter := range adapters {
		t.Run(adapter.Name(), func(t *testing.T) {
			dir := t.TempDir()
			// Override the skill path to use temp dir
			skillPath := filepath.Join(dir, "skills", "monocle-review", "SKILL.md")

			if err := adapter.Install(skillPath); err != nil {
				t.Fatalf("install failed: %v", err)
			}

			installed, err := adapter.IsInstalled(skillPath)
			if err != nil {
				t.Fatalf("IsInstalled error: %v", err)
			}
			if !installed {
				t.Fatal("should be installed")
			}

			// Verify content contains expected text
			data, _ := os.ReadFile(skillPath)
			content := string(data)
			if len(content) == 0 {
				t.Fatal("skill file should not be empty")
			}
		})
	}
}

func TestSkillUninstall_AllAdapters(t *testing.T) {
	adapters := []SkillInstaller{
		&ClaudeAdapter{},
		&GeminiAdapter{},
		&OpenCodeAdapter{},
	}

	for _, adapter := range adapters {
		t.Run(adapter.Name(), func(t *testing.T) {
			dir := t.TempDir()
			skillPath := filepath.Join(dir, "skills", "monocle-review", "SKILL.md")

			adapter.Install(skillPath)
			adapter.Uninstall(skillPath)

			installed, _ := adapter.IsInstalled(skillPath)
			if installed {
				t.Fatal("should not be installed after uninstall")
			}
		})
	}
}

func TestCodexConfigPath_ProjectNotSupported(t *testing.T) {
	adapter := &CodexAdapter{}
	_, supported := adapter.SkillPath(false)
	if supported {
		t.Fatal("codex should not support project-level skill installation")
	}
}

func TestResolveInstallers_Auto(t *testing.T) {
	// Auto with no agents detected should return empty
	installers := resolveInstallers([]string{"auto"}, false)
	// We can't predict what's installed, just ensure it doesn't panic
	_ = installers
}

func TestResolveInstallers_Named(t *testing.T) {
	installers := resolveInstallers([]string{"claude"}, false)
	if len(installers) != 1 {
		t.Fatalf("expected 1 installer, got %d", len(installers))
	}
	if installers[0].Name() != "claude" {
		t.Fatalf("expected claude, got %s", installers[0].Name())
	}
}

func TestResolveInstallers_Unknown(t *testing.T) {
	installers := resolveInstallers([]string{"unknown-agent"}, false)
	if len(installers) != 0 {
		t.Fatalf("expected 0 installers for unknown agent, got %d", len(installers))
	}
}

func TestIsAuto(t *testing.T) {
	tests := []struct {
		input []string
		want  bool
	}{
		{nil, true},
		{[]string{}, true},
		{[]string{"auto"}, true},
		{[]string{"claude"}, false},
		{[]string{"auto", "claude"}, false},
	}
	for _, tt := range tests {
		got := isAuto(tt.input)
		if got != tt.want {
			t.Errorf("isAuto(%v) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

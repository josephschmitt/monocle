package adapters

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCodexRegister(t *testing.T) {
	dir := t.TempDir()
	projDir := filepath.Join(dir, "project")
	os.MkdirAll(projDir, 0755)

	origDir, _ := os.Getwd()
	os.Chdir(projDir)
	defer os.Chdir(origDir)

	adapter := &CodexAdapter{}
	if err := adapter.Register(false); err != nil {
		t.Fatalf("register: %v", err)
	}

	// Verify skill files are installed
	for _, name := range SkillNames {
		path := filepath.Join(projDir, ".codex", "skills", name, "SKILL.md")
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("skill %s not found: %v", name, err)
		}
	}
}

func TestCodexRegister_Idempotent(t *testing.T) {
	dir := t.TempDir()
	projDir := filepath.Join(dir, "project")
	os.MkdirAll(projDir, 0755)

	origDir, _ := os.Getwd()
	os.Chdir(projDir)
	defer os.Chdir(origDir)

	adapter := &CodexAdapter{}
	if err := adapter.Register(false); err != nil {
		t.Fatalf("first register: %v", err)
	}
	if err := adapter.Register(false); err != nil {
		t.Fatalf("second register: %v", err)
	}

	for _, name := range SkillNames {
		path := filepath.Join(projDir, ".codex", "skills", name, "SKILL.md")
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("skill %s not found after re-register: %v", name, err)
		}
	}
}

func TestCodexRegister_CleansLegacyMCP(t *testing.T) {
	dir := t.TempDir()
	projDir := filepath.Join(dir, "project")
	codexDir := filepath.Join(projDir, ".codex")
	os.MkdirAll(codexDir, 0755)

	origDir, _ := os.Getwd()
	os.Chdir(projDir)
	defer os.Chdir(origDir)

	// Write legacy MCP config
	oldConfig := "[mcp_servers.monocle]\ncommand = \"monocle\"\nargs = [\"serve-mcp-channel\"]\n"
	if err := os.WriteFile(filepath.Join(codexDir, "config.toml"), []byte(oldConfig), 0644); err != nil {
		t.Fatalf("write old config: %v", err)
	}

	adapter := &CodexAdapter{}
	if err := adapter.Register(false); err != nil {
		t.Fatalf("register: %v", err)
	}

	// Legacy config should be cleaned up (file removed since it was only monocle)
	if _, err := os.Stat(filepath.Join(codexDir, "config.toml")); !os.IsNotExist(err) {
		content, _ := os.ReadFile(filepath.Join(codexDir, "config.toml"))
		if strings.Contains(string(content), "[mcp_servers.monocle]") {
			t.Fatal("legacy MCP config should have been removed")
		}
	}

	// Skills should be installed
	for _, name := range SkillNames {
		path := filepath.Join(projDir, ".codex", "skills", name, "SKILL.md")
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("skill %s not found: %v", name, err)
		}
	}
}

func TestCodexRegister_PreservesOtherSections(t *testing.T) {
	dir := t.TempDir()
	projDir := filepath.Join(dir, "project")
	codexDir := filepath.Join(projDir, ".codex")
	os.MkdirAll(codexDir, 0755)

	origDir, _ := os.Getwd()
	os.Chdir(projDir)
	defer os.Chdir(origDir)

	// Write config with monocle and another tool
	oldConfig := "[mcp_servers.other_tool]\ncommand = \"other\"\nargs = [\"run\"]\n\n[mcp_servers.monocle]\ncommand = \"old-monocle\"\nargs = [\"serve-mcp-channel\"]\n"
	if err := os.WriteFile(filepath.Join(codexDir, "config.toml"), []byte(oldConfig), 0644); err != nil {
		t.Fatalf("write old config: %v", err)
	}

	adapter := &CodexAdapter{}
	if err := adapter.Register(false); err != nil {
		t.Fatalf("register: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(codexDir, "config.toml"))
	if err != nil {
		t.Fatalf("read config: %v", err)
	}

	s := string(content)
	if !strings.Contains(s, "[mcp_servers.other_tool]") {
		t.Fatal("other_tool section should be preserved")
	}
	if !strings.Contains(s, `command = "other"`) {
		t.Fatal("other_tool command should be preserved")
	}
	if strings.Contains(s, "[mcp_servers.monocle]") {
		t.Fatal("monocle MCP section should have been removed")
	}
}

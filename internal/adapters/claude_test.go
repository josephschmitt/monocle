package adapters

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func requireBun(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("bun"); err != nil {
		t.Skip("bun not available, skipping channel install test")
	}
}

func TestClaudeChannelInstall(t *testing.T) {
	requireBun(t)
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "config"))

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	projDir := filepath.Join(dir, "project")
	os.MkdirAll(projDir, 0755)
	os.Chdir(projDir)
	defer os.Chdir(origDir)

	adapter := &ClaudeAdapter{}

	// Not installed initially
	installed, err := adapter.IsInstalled()
	if err != nil {
		t.Fatalf("IsInstalled error: %v", err)
	}
	if installed {
		t.Fatal("should not be installed initially")
	}

	// Install
	if err := adapter.Install(false); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	// Verify channel.ts exists
	installed, err = adapter.IsInstalled()
	if err != nil {
		t.Fatalf("IsInstalled error: %v", err)
	}
	if !installed {
		t.Fatal("should be installed after Install()")
	}

	channelPath := channelTSPath()
	data, err := os.ReadFile(channelPath)
	if err != nil {
		t.Fatalf("read channel.ts: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("channel.ts should not be empty")
	}

	// Verify package.json exists
	channelDir := filepath.Dir(channelPath)
	pkgData, err := os.ReadFile(filepath.Join(channelDir, "package.json"))
	if err != nil {
		t.Fatalf("read package.json: %v", err)
	}
	if len(pkgData) == 0 {
		t.Fatal("package.json should not be empty")
	}

	// Verify node_modules was created (bun install ran)
	if _, err := os.Stat(filepath.Join(channelDir, "node_modules")); err != nil {
		t.Fatalf("node_modules should exist after install: %v", err)
	}

	// Verify .mcp.json exists with monocle entry
	mcpData, err := os.ReadFile(filepath.Join(projDir, ".mcp.json"))
	if err != nil {
		t.Fatalf("read .mcp.json: %v", err)
	}
	var mcpConfig map[string]any
	if err := json.Unmarshal(mcpData, &mcpConfig); err != nil {
		t.Fatalf("parse .mcp.json: %v", err)
	}
	servers, ok := mcpConfig["mcpServers"].(map[string]any)
	if !ok {
		t.Fatal("mcpServers should exist in .mcp.json")
	}
	if _, ok := servers["monocle"]; !ok {
		t.Fatal("monocle should be in mcpServers")
	}
}

func TestHasMCPConfig_NoFiles(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "config"))
	t.Setenv("HOME", filepath.Join(dir, "home"))

	origDir, _ := os.Getwd()
	projDir := filepath.Join(dir, "project")
	os.MkdirAll(projDir, 0755)
	os.Chdir(projDir)
	defer os.Chdir(origDir)

	adapter := &ClaudeAdapter{}
	if adapter.HasMCPConfig() {
		t.Fatal("should return false when no .mcp.json exists")
	}
}

func TestHasMCPConfig_GlobalExists(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "config"))

	homeDir := filepath.Join(dir, "home")
	os.MkdirAll(homeDir, 0755)
	t.Setenv("HOME", homeDir)

	origDir, _ := os.Getwd()
	projDir := filepath.Join(dir, "project")
	os.MkdirAll(projDir, 0755)
	os.Chdir(projDir)
	defer os.Chdir(origDir)

	// Write global .mcp.json with monocle entry
	mcpData := map[string]any{
		"mcpServers": map[string]any{
			"monocle": map[string]any{"command": "bun"},
		},
	}
	data, _ := json.Marshal(mcpData)
	os.WriteFile(filepath.Join(homeDir, ".mcp.json"), data, 0644)

	adapter := &ClaudeAdapter{}
	if !adapter.HasMCPConfig() {
		t.Fatal("should return true when global .mcp.json has monocle")
	}
}

func TestHasMCPConfig_LocalExists(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "config"))
	t.Setenv("HOME", filepath.Join(dir, "home"))

	origDir, _ := os.Getwd()
	projDir := filepath.Join(dir, "project")
	os.MkdirAll(projDir, 0755)
	os.Chdir(projDir)
	defer os.Chdir(origDir)

	// Write local .mcp.json with monocle entry
	mcpData := map[string]any{
		"mcpServers": map[string]any{
			"monocle": map[string]any{"command": "bun"},
		},
	}
	data, _ := json.Marshal(mcpData)
	os.WriteFile(filepath.Join(projDir, ".mcp.json"), data, 0644)

	adapter := &ClaudeAdapter{}
	if !adapter.HasMCPConfig() {
		t.Fatal("should return true when local .mcp.json has monocle")
	}
}

func TestNeedsInstall_ChannelMissing(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "config"))
	t.Setenv("HOME", filepath.Join(dir, "home"))

	adapter := &ClaudeAdapter{}
	if !adapter.NeedsInstall() {
		t.Fatal("should need install when channel.ts is missing")
	}
}

func TestNeedsInstall_ConfigMissing(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "config"))
	t.Setenv("HOME", filepath.Join(dir, "home"))

	origDir, _ := os.Getwd()
	projDir := filepath.Join(dir, "project")
	os.MkdirAll(projDir, 0755)
	os.Chdir(projDir)
	defer os.Chdir(origDir)

	// Create channel.ts but no .mcp.json
	channelDir := filepath.Join(dir, "config", "monocle")
	os.MkdirAll(channelDir, 0755)
	os.WriteFile(filepath.Join(channelDir, "channel.ts"), []byte("// channel"), 0644)

	adapter := &ClaudeAdapter{}
	if !adapter.NeedsInstall() {
		t.Fatal("should need install when .mcp.json config is missing")
	}
}

func TestNeedsInstall_FullyInstalled(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "config"))

	homeDir := filepath.Join(dir, "home")
	os.MkdirAll(homeDir, 0755)
	t.Setenv("HOME", homeDir)

	origDir, _ := os.Getwd()
	projDir := filepath.Join(dir, "project")
	os.MkdirAll(projDir, 0755)
	os.Chdir(projDir)
	defer os.Chdir(origDir)

	// Create channel.ts
	channelDir := filepath.Join(dir, "config", "monocle")
	os.MkdirAll(channelDir, 0755)
	os.WriteFile(filepath.Join(channelDir, "channel.ts"), []byte("// channel"), 0644)

	// Create global .mcp.json with monocle entry
	mcpData := map[string]any{
		"mcpServers": map[string]any{
			"monocle": map[string]any{"command": "bun"},
		},
	}
	data, _ := json.Marshal(mcpData)
	os.WriteFile(filepath.Join(homeDir, ".mcp.json"), data, 0644)

	adapter := &ClaudeAdapter{}
	if adapter.NeedsInstall() {
		t.Fatal("should not need install when fully set up")
	}
}

func TestClaudeChannelInstall_Idempotent(t *testing.T) {
	requireBun(t)
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "config"))

	origDir, _ := os.Getwd()
	projDir := filepath.Join(dir, "project")
	os.MkdirAll(projDir, 0755)
	os.Chdir(projDir)
	defer os.Chdir(origDir)

	adapter := &ClaudeAdapter{}
	if err := adapter.Install(false); err != nil {
		t.Fatalf("first install: %v", err)
	}
	if err := adapter.Install(false); err != nil {
		t.Fatalf("second install: %v", err)
	}

	installed, _ := adapter.IsInstalled()
	if !installed {
		t.Fatal("should be installed")
	}
}

func TestClaudeChannelUninstall(t *testing.T) {
	requireBun(t)
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "config"))

	origDir, _ := os.Getwd()
	projDir := filepath.Join(dir, "project")
	os.MkdirAll(projDir, 0755)
	os.Chdir(projDir)
	defer os.Chdir(origDir)

	adapter := &ClaudeAdapter{}
	if err := adapter.Install(false); err != nil {
		t.Fatalf("install: %v", err)
	}
	if err := adapter.Uninstall(false); err != nil {
		t.Fatalf("uninstall: %v", err)
	}

	installed, _ := adapter.IsInstalled()
	if installed {
		t.Fatal("should not be installed after uninstall")
	}

	// Channel directory should be completely removed
	channelPath := channelTSPath()
	if _, err := os.Stat(filepath.Dir(channelPath)); !os.IsNotExist(err) {
		t.Fatal("channel directory should be removed after uninstall")
	}

	// .mcp.json should be removed (was only entry)
	if _, err := os.Stat(filepath.Join(projDir, ".mcp.json")); !os.IsNotExist(err) {
		t.Fatal(".mcp.json should be removed after uninstall")
	}
}

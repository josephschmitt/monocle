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

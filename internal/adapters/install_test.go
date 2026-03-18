package adapters

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- Claude installer tests ---

func TestClaudeInstall_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".claude", "settings.json")

	adapter := &ClaudeAdapter{}
	if err := adapter.Install(configPath, InstallOptions{}); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	config, err := ReadJSONFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}

	hooks, ok := config["hooks"].(map[string]any)
	if !ok {
		t.Fatal("expected hooks key")
	}

	// Check PostToolUse
	ptu, ok := hooks["PostToolUse"].([]any)
	if !ok || len(ptu) == 0 {
		t.Fatal("expected PostToolUse hooks")
	}

	// Check Stop
	stop, ok := hooks["Stop"].([]any)
	if !ok || len(stop) == 0 {
		t.Fatal("expected Stop hooks")
	}
}

func TestClaudeInstall_ExistingConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "settings.json")

	// Write pre-existing config
	existing := map[string]any{
		"permissions": map[string]any{"allow": []any{"Read"}},
		"hooks": map[string]any{
			"PreToolUse": []any{
				map[string]any{
					"matcher": "",
					"hooks": []any{
						map[string]any{"type": "command", "command": "my-tool check"},
					},
				},
			},
		},
	}
	if err := WriteJSONFile(configPath, existing); err != nil {
		t.Fatalf("write existing config: %v", err)
	}

	adapter := &ClaudeAdapter{}
	if err := adapter.Install(configPath, InstallOptions{}); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	config, err := ReadJSONFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}

	// Verify existing settings are preserved
	if _, ok := config["permissions"]; !ok {
		t.Fatal("permissions key was lost")
	}

	hooks := config["hooks"].(map[string]any)

	// Verify existing hook is preserved
	if _, ok := hooks["PreToolUse"]; !ok {
		t.Fatal("PreToolUse hook was lost")
	}

	// Verify monocle hooks were added
	if _, ok := hooks["PostToolUse"]; !ok {
		t.Fatal("PostToolUse not added")
	}
	if _, ok := hooks["Stop"]; !ok {
		t.Fatal("Stop not added")
	}
}

func TestClaudeInstall_Idempotent(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "settings.json")

	adapter := &ClaudeAdapter{}

	// Install twice
	if err := adapter.Install(configPath, InstallOptions{}); err != nil {
		t.Fatalf("first install failed: %v", err)
	}
	if err := adapter.Install(configPath, InstallOptions{}); err != nil {
		t.Fatalf("second install failed: %v", err)
	}

	config, err := ReadJSONFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}

	hooks := config["hooks"].(map[string]any)

	// Should have exactly 1 PostToolUse entry, not 2
	ptu := hooks["PostToolUse"].([]any)
	if len(ptu) != 1 {
		t.Fatalf("expected 1 PostToolUse entry, got %d", len(ptu))
	}

	stop := hooks["Stop"].([]any)
	if len(stop) != 1 {
		t.Fatalf("expected 1 Stop entry, got %d", len(stop))
	}
}

func TestClaudeInstall_IsInstalled(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "settings.json")

	adapter := &ClaudeAdapter{}

	// Not installed initially
	installed, err := adapter.IsInstalled(configPath)
	if err != nil {
		t.Fatalf("IsInstalled error: %v", err)
	}
	if installed {
		t.Fatal("should not be installed initially")
	}

	// Install
	if err := adapter.Install(configPath, InstallOptions{}); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	// Now installed
	installed, err = adapter.IsInstalled(configPath)
	if err != nil {
		t.Fatalf("IsInstalled error: %v", err)
	}
	if !installed {
		t.Fatal("should be installed after Install()")
	}
}

func TestClaudeUninstall(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "settings.json")

	adapter := &ClaudeAdapter{}

	// Install first
	if err := adapter.Install(configPath, InstallOptions{}); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	// Uninstall
	if err := adapter.Uninstall(configPath); err != nil {
		t.Fatalf("uninstall failed: %v", err)
	}

	// Verify hooks removed
	installed, err := adapter.IsInstalled(configPath)
	if err != nil {
		t.Fatalf("IsInstalled error: %v", err)
	}
	if installed {
		t.Fatal("should not be installed after Uninstall()")
	}
}

func TestClaudeUninstall_PreservesOtherHooks(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "settings.json")

	// Write config with both monocle and other hooks
	config := map[string]any{
		"hooks": map[string]any{
			"PostToolUse": []any{
				map[string]any{
					"matcher": "Write",
					"hooks": []any{
						map[string]any{"type": "command", "command": "monocle hook post-tool-use --agent claude", "async": true},
					},
				},
			},
			"Stop": []any{
				map[string]any{
					"matcher": "",
					"hooks": []any{
						map[string]any{"type": "command", "command": "monocle hook stop --agent claude"},
					},
				},
			},
			"PreToolUse": []any{
				map[string]any{
					"matcher": "",
					"hooks": []any{
						map[string]any{"type": "command", "command": "other-tool check"},
					},
				},
			},
		},
	}
	if err := WriteJSONFile(configPath, config); err != nil {
		t.Fatalf("write config: %v", err)
	}

	adapter := &ClaudeAdapter{}
	if err := adapter.Uninstall(configPath); err != nil {
		t.Fatalf("uninstall failed: %v", err)
	}

	result, err := ReadJSONFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}

	hooks := result["hooks"].(map[string]any)

	// PreToolUse should still be there
	if _, ok := hooks["PreToolUse"]; !ok {
		t.Fatal("PreToolUse was removed")
	}

	// PostToolUse and Stop should be gone
	if _, ok := hooks["PostToolUse"]; ok {
		t.Fatal("PostToolUse should have been removed")
	}
	if _, ok := hooks["Stop"]; ok {
		t.Fatal("Stop should have been removed")
	}
}

func TestClaudeInstall_CustomHookBinary(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "settings.json")

	adapter := &ClaudeAdapter{}
	if err := adapter.Install(configPath, InstallOptions{HookBinaryPath: "/usr/local/bin/monocle"}); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	config, err := ReadJSONFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}

	hooks := config["hooks"].(map[string]any)
	ptu := hooks["PostToolUse"].([]any)
	entry := ptu[0].(map[string]any)
	innerHooks := entry["hooks"].([]any)
	hook := innerHooks[0].(map[string]any)
	cmd := hook["command"].(string)

	if cmd != "/usr/local/bin/monocle hook post-tool-use --agent claude" {
		t.Fatalf("unexpected command: %s", cmd)
	}
}

func TestClaudeInstall_ScopeDefault(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "settings.json")

	adapter := &ClaudeAdapter{}
	if err := adapter.Install(configPath, InstallOptions{}); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	config, _ := ReadJSONFile(configPath)
	hooks := config["hooks"].(map[string]any)
	ptu := hooks["PostToolUse"].([]any)
	entry := ptu[0].(map[string]any)
	innerHooks := entry["hooks"].([]any)
	hook := innerHooks[0].(map[string]any)
	cmd := hook["command"].(string)

	// Default scope should NOT include --scope cwd
	if strings.Contains(cmd, "--scope") {
		t.Fatalf("default scope should not include --scope flag, got: %s", cmd)
	}
}

func TestClaudeInstall_ScopeCwd(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "settings.json")

	adapter := &ClaudeAdapter{}
	if err := adapter.Install(configPath, InstallOptions{Scope: "cwd"}); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	config, _ := ReadJSONFile(configPath)
	hooks := config["hooks"].(map[string]any)

	// Check PostToolUse command
	ptu := hooks["PostToolUse"].([]any)
	entry := ptu[0].(map[string]any)
	innerHooks := entry["hooks"].([]any)
	hook := innerHooks[0].(map[string]any)
	cmd := hook["command"].(string)
	if cmd != "monocle hook --scope cwd post-tool-use --agent claude" {
		t.Fatalf("unexpected PostToolUse command: %s", cmd)
	}

	// Check Stop command
	stop := hooks["Stop"].([]any)
	stopEntry := stop[0].(map[string]any)
	stopHooks := stopEntry["hooks"].([]any)
	stopHook := stopHooks[0].(map[string]any)
	stopCmd := stopHook["command"].(string)
	if stopCmd != "monocle hook --scope cwd stop --agent claude" {
		t.Fatalf("unexpected Stop command: %s", stopCmd)
	}
}

func TestGeminiInstall_ScopeCwd(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "settings.json")

	adapter := &GeminiAdapter{}
	if err := adapter.Install(configPath, InstallOptions{Scope: "cwd"}); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	config, _ := ReadJSONFile(configPath)
	hooks := config["hooks"].(map[string]any)
	at := hooks["AfterTool"].([]any)
	entry := at[0].(map[string]any)
	innerHooks := entry["hooks"].([]any)
	hook := innerHooks[0].(map[string]any)
	cmd := hook["command"].(string)
	if cmd != "monocle hook --scope cwd after-tool --agent gemini" {
		t.Fatalf("unexpected command: %s", cmd)
	}
}

func TestCodexInstall_ScopeCwd(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")

	adapter := &CodexAdapter{}
	if err := adapter.Install(configPath, InstallOptions{Scope: "cwd"}); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	config, _ := readTOMLFile(configPath)
	notify := config["notify"].(map[string]any)
	cmd := notify["agent-turn-complete"].(string)
	if !strings.Contains(cmd, "hook --scope cwd agent-turn-complete") {
		t.Fatalf("expected --scope cwd in command, got: %s", cmd)
	}
}

func TestOpenCodeInstall_ScopeCwd(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "opencode.json")

	adapter := &OpenCodeAdapter{}
	if err := adapter.Install(configPath, InstallOptions{Scope: "cwd"}); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	config, _ := ReadJSONFile(configPath)
	hookCfg, _ := GetNestedKey(config, "experimental.hook")
	hookMap := hookCfg.(map[string]any)
	fe := hookMap["file_edited"].(map[string]any)
	cmd := fe["command"].(string)
	if cmd != "monocle hook --scope cwd file-edited --agent opencode" {
		t.Fatalf("unexpected command: %s", cmd)
	}
}

// --- Gemini installer tests ---

func TestGeminiInstall_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "settings.json")

	adapter := &GeminiAdapter{}
	if err := adapter.Install(configPath, InstallOptions{}); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	config, err := ReadJSONFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}

	hooks := config["hooks"].(map[string]any)

	if _, ok := hooks["AfterTool"]; !ok {
		t.Fatal("expected AfterTool hooks")
	}
	if _, ok := hooks["AfterAgent"]; !ok {
		t.Fatal("expected AfterAgent hooks")
	}
}

func TestGeminiInstall_Idempotent(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "settings.json")

	adapter := &GeminiAdapter{}
	adapter.Install(configPath, InstallOptions{})
	adapter.Install(configPath, InstallOptions{})

	config, _ := ReadJSONFile(configPath)
	hooks := config["hooks"].(map[string]any)

	at := hooks["AfterTool"].([]any)
	if len(at) != 1 {
		t.Fatalf("expected 1 AfterTool entry, got %d", len(at))
	}
}

func TestGeminiUninstall(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "settings.json")

	adapter := &GeminiAdapter{}
	adapter.Install(configPath, InstallOptions{})
	adapter.Uninstall(configPath)

	installed, _ := adapter.IsInstalled(configPath)
	if installed {
		t.Fatal("should not be installed after uninstall")
	}
}

// --- Codex installer tests ---

func TestCodexConfigPath_ProjectNotSupported(t *testing.T) {
	adapter := &CodexAdapter{}
	_, supported := adapter.ConfigPath(false)
	if supported {
		t.Fatal("codex should not support project-level config")
	}
}

func TestCodexInstall_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")

	adapter := &CodexAdapter{}
	if err := adapter.Install(configPath, InstallOptions{}); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	config, err := readTOMLFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}

	notify, ok := config["notify"].(map[string]any)
	if !ok {
		t.Fatal("expected notify section")
	}

	cmd, ok := notify["agent-turn-complete"].(string)
	if !ok {
		t.Fatal("expected agent-turn-complete key")
	}
	if cmd == "" {
		t.Fatal("command should not be empty")
	}
}

func TestCodexInstall_Idempotent(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")

	adapter := &CodexAdapter{}
	adapter.Install(configPath, InstallOptions{})
	adapter.Install(configPath, InstallOptions{})

	installed, _ := adapter.IsInstalled(configPath)
	if !installed {
		t.Fatal("should be installed")
	}
}

func TestCodexUninstall(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")

	adapter := &CodexAdapter{}
	adapter.Install(configPath, InstallOptions{})
	adapter.Uninstall(configPath)

	installed, _ := adapter.IsInstalled(configPath)
	if installed {
		t.Fatal("should not be installed after uninstall")
	}
}

// --- OpenCode installer tests ---

func TestOpenCodeInstall_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "opencode.json")

	adapter := &OpenCodeAdapter{}
	if err := adapter.Install(configPath, InstallOptions{}); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	config, err := ReadJSONFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}

	hookCfg, ok := GetNestedKey(config, "experimental.hook")
	if !ok {
		t.Fatal("expected experimental.hook key")
	}

	hookMap := hookCfg.(map[string]any)
	if _, ok := hookMap["file_edited"]; !ok {
		t.Fatal("expected file_edited hook")
	}
	if _, ok := hookMap["session_completed"]; !ok {
		t.Fatal("expected session_completed hook")
	}
}

func TestOpenCodeInstall_Idempotent(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "opencode.json")

	adapter := &OpenCodeAdapter{}
	adapter.Install(configPath, InstallOptions{})
	adapter.Install(configPath, InstallOptions{})

	config, _ := ReadJSONFile(configPath)
	hookCfg, _ := GetNestedKey(config, "experimental.hook")
	hookMap := hookCfg.(map[string]any)

	// Should have exactly 2 events, not duplicates
	if len(hookMap) != 2 {
		t.Fatalf("expected 2 hook events, got %d", len(hookMap))
	}
}

func TestOpenCodeUninstall(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "opencode.json")

	adapter := &OpenCodeAdapter{}
	adapter.Install(configPath, InstallOptions{})
	adapter.Uninstall(configPath)

	installed, _ := adapter.IsInstalled(configPath)
	if installed {
		t.Fatal("should not be installed after uninstall")
	}

	// Verify experimental key is cleaned up
	config, _ := ReadJSONFile(configPath)
	if _, ok := config["experimental"]; ok {
		t.Fatal("empty experimental key should be cleaned up")
	}
}

func TestOpenCodeInstall_PreservesExistingConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "opencode.json")

	existing := map[string]any{
		"model": "gpt-4",
		"experimental": map[string]any{
			"other": "setting",
		},
	}
	WriteJSONFile(configPath, existing)

	adapter := &OpenCodeAdapter{}
	adapter.Install(configPath, InstallOptions{})

	config, _ := ReadJSONFile(configPath)

	// Existing settings preserved
	if config["model"] != "gpt-4" {
		t.Fatal("model setting was lost")
	}

	exp := config["experimental"].(map[string]any)
	if exp["other"] != "setting" {
		t.Fatal("experimental.other was lost")
	}

	// Hooks added
	if _, ok := exp["hook"]; !ok {
		t.Fatal("experimental.hook not added")
	}
}

// --- Orchestrator tests ---

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

// --- Integration test: full install/uninstall cycle via orchestrator ---

func TestInstallAgents_FullCycle(t *testing.T) {
	dir := t.TempDir()

	// Create a fake .claude dir so it looks like Claude is installed
	claudeConfigDir := filepath.Join(dir, ".claude")
	os.MkdirAll(claudeConfigDir, 0755)
	configPath := filepath.Join(claudeConfigDir, "settings.json")

	// Direct install via adapter (since orchestrator uses Detect() which checks real system)
	adapter := &ClaudeAdapter{}
	err := adapter.Install(configPath, InstallOptions{HookBinaryPath: "monocle"})
	if err != nil {
		t.Fatalf("install failed: %v", err)
	}

	// Verify installed
	installed, err := adapter.IsInstalled(configPath)
	if err != nil {
		t.Fatalf("check installed: %v", err)
	}
	if !installed {
		t.Fatal("should be installed")
	}

	// Uninstall
	err = adapter.Uninstall(configPath)
	if err != nil {
		t.Fatalf("uninstall failed: %v", err)
	}

	// Verify uninstalled
	installed, err = adapter.IsInstalled(configPath)
	if err != nil {
		t.Fatalf("check installed: %v", err)
	}
	if installed {
		t.Fatal("should not be installed")
	}
}

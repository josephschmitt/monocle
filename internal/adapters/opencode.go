package adapters

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/anthropics/monocle/internal/protocol"
)

// OpenCodeAdapter handles OpenCode's hook format.
// Config lives at opencode.json (project) or ~/.config/opencode/opencode.json (global),
// with hooks under the "experimental.hook" nested path.
type OpenCodeAdapter struct{}

func (a *OpenCodeAdapter) ParseHookInput(event string, raw []byte) (any, error) {
	switch event {
	case "file-edited":
		var input map[string]any
		if err := json.Unmarshal(raw, &input); err != nil {
			return nil, fmt.Errorf("parse opencode input: %w", err)
		}
		msg := &protocol.PostToolUseMsg{
			Type:  protocol.TypePostToolUse,
			Agent: "opencode",
		}
		if fp, ok := input["file_path"].(string); ok {
			msg.FilePath = fp
		}
		return msg, nil

	case "session-completed":
		return &protocol.StopMsg{
			Type:  protocol.TypeStop,
			Agent: "opencode",
		}, nil

	default:
		return nil, fmt.Errorf("unknown event: %s", event)
	}
}

func (a *OpenCodeAdapter) FormatHookOutput(response any) HookOutput {
	return HookOutput{ExitCode: 0}
}

func (a *OpenCodeAdapter) GenerateConfig(opts SetupOptions) string {
	hookCmd := opts.HookBinaryPath
	if hookCmd == "" {
		hookCmd = "monocle"
	}

	config := map[string]any{
		"experimental": map[string]any{
			"hook": map[string]any{
				"file_edited": map[string]any{
					"command": hookCmd + " hook file-edited --agent opencode",
					"async":   true,
				},
				"session_completed": map[string]any{
					"command": hookCmd + " hook session-completed --agent opencode",
				},
			},
		},
	}

	data, _ := json.MarshalIndent(config, "", "  ")
	return string(data)
}

func (a *OpenCodeAdapter) Capabilities() AdapterCapabilities {
	return AdapterCapabilities{
		PostToolUse:  true,
		StopBlocking: false,
		AsyncHooks:   true,
	}
}

// AgentInstaller implementation

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

func (a *OpenCodeAdapter) ConfigPath(global bool) (string, bool) {
	if global {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", false
		}
		return filepath.Join(home, ".config", "opencode", "opencode.json"), true
	}
	return "opencode.json", true
}

func (a *OpenCodeAdapter) IsInstalled(configPath string) (bool, error) {
	config, err := ReadJSONFile(configPath)
	if err != nil {
		return false, err
	}
	hookCfg, ok := GetNestedKey(config, "experimental.hook")
	if !ok {
		return false, nil
	}
	hookMap, ok := hookCfg.(map[string]any)
	if !ok {
		return false, nil
	}

	for _, eventCfg := range hookMap {
		em, ok := eventCfg.(map[string]any)
		if !ok {
			continue
		}
		cmd, _ := em["command"].(string)
		if isMonocleCommand(cmd) {
			return true, nil
		}
	}
	return false, nil
}

func (a *OpenCodeAdapter) Install(configPath string, opts InstallOptions) error {
	hookCmd := opts.HookBinaryPath
	if hookCmd == "" {
		hookCmd = "monocle"
	}
	scopeFlag := scopeFlagStr(opts.Scope)

	config, err := ReadJSONFile(configPath)
	if err != nil {
		return err
	}

	// Get or create experimental.hook map
	var hookMap map[string]any
	if existing, ok := GetNestedKey(config, "experimental.hook"); ok {
		hookMap, _ = existing.(map[string]any)
	}
	if hookMap == nil {
		hookMap = map[string]any{}
	}

	// Check if already installed
	for _, eventCfg := range hookMap {
		em, ok := eventCfg.(map[string]any)
		if !ok {
			continue
		}
		cmd, _ := em["command"].(string)
		if isMonocleCommand(cmd) {
			return nil // already installed
		}
	}

	hookMap["file_edited"] = map[string]any{
		"command": hookCmd + " hook" + scopeFlag + " file-edited --agent opencode",
		"async":   true,
	}
	hookMap["session_completed"] = map[string]any{
		"command": hookCmd + " hook" + scopeFlag + " session-completed --agent opencode",
	}

	SetNestedKey(config, "experimental.hook", hookMap)
	return WriteJSONFile(configPath, config)
}

func (a *OpenCodeAdapter) Uninstall(configPath string) error {
	config, err := ReadJSONFile(configPath)
	if err != nil {
		return err
	}

	existing, ok := GetNestedKey(config, "experimental.hook")
	if !ok {
		return nil
	}
	hookMap, ok := existing.(map[string]any)
	if !ok {
		return nil
	}

	// Remove entries containing monocle hook commands
	for key, eventCfg := range hookMap {
		em, ok := eventCfg.(map[string]any)
		if !ok {
			continue
		}
		cmd, _ := em["command"].(string)
		if isMonocleCommand(cmd) {
			delete(hookMap, key)
		}
	}

	if len(hookMap) == 0 {
		DeleteNestedKey(config, "experimental.hook")
		// Clean up empty experimental map
		if exp, ok := config["experimental"].(map[string]any); ok && len(exp) == 0 {
			delete(config, "experimental")
		}
	} else {
		SetNestedKey(config, "experimental.hook", hookMap)
	}

	return WriteJSONFile(configPath, config)
}

// Ensure OpenCodeAdapter implements both interfaces.
var (
	_ AgentAdapter   = (*OpenCodeAdapter)(nil)
	_ AgentInstaller = (*OpenCodeAdapter)(nil)
)

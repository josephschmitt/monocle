package adapters

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/anthropics/monocle/internal/protocol"
)

// GeminiAdapter handles Gemini CLI's hook format.
type GeminiAdapter struct{}

func (a *GeminiAdapter) ParseHookInput(event string, raw []byte) (any, error) {
	var input map[string]any
	if err := json.Unmarshal(raw, &input); err != nil {
		return nil, fmt.Errorf("parse gemini input: %w", err)
	}

	switch event {
	case "after-tool":
		return &protocol.PostToolUseMsg{
			Type:  protocol.TypePostToolUse,
			Agent: "gemini",
		}, nil

	case "after-agent":
		return &protocol.StopMsg{
			Type:  protocol.TypeStop,
			Agent: "gemini",
		}, nil

	default:
		return nil, fmt.Errorf("unknown event: %s", event)
	}
}

func (a *GeminiAdapter) FormatHookOutput(response any) HookOutput {
	switch r := response.(type) {
	case *protocol.StopResponse:
		if r.Continue && r.SystemMessage != "" {
			data, _ := json.Marshal(map[string]string{
				"decision": "block",
				"reason":   r.SystemMessage,
			})
			return HookOutput{Data: data, ExitCode: 0}
		}
		return HookOutput{ExitCode: 0}
	default:
		return HookOutput{ExitCode: 0}
	}
}

func (a *GeminiAdapter) GenerateConfig(opts SetupOptions) string {
	hookCmd := opts.HookBinaryPath
	if hookCmd == "" {
		hookCmd = "monocle"
	}

	config := map[string]any{
		"hooks": map[string]any{
			"AfterTool": []map[string]any{
				{
					"matcher": "",
					"hooks": []map[string]any{
						{
							"type":    "command",
							"command": hookCmd + " hook after-tool --agent gemini",
							"async":   true,
						},
					},
				},
			},
			"AfterAgent": []map[string]any{
				{
					"matcher": "",
					"hooks": []map[string]any{
						{
							"type":    "command",
							"command": hookCmd + " hook after-agent --agent gemini",
						},
					},
				},
			},
		},
	}

	data, _ := json.MarshalIndent(config, "", "  ")
	return string(data)
}

func (a *GeminiAdapter) Capabilities() AdapterCapabilities {
	return AdapterCapabilities{
		PostToolUse:  true,
		StopBlocking: true,
		AsyncHooks:   true,
	}
}

// AgentInstaller implementation

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

func (a *GeminiAdapter) ConfigPath(global bool) (string, bool) {
	if global {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", false
		}
		return filepath.Join(home, ".gemini", "settings.json"), true
	}
	return filepath.Join(".gemini", "settings.json"), true
}

func (a *GeminiAdapter) IsInstalled(configPath string) (bool, error) {
	config, err := ReadJSONFile(configPath)
	if err != nil {
		return false, err
	}
	return jsonHooksContainMonocle(config, "AfterTool") ||
		jsonHooksContainMonocle(config, "AfterAgent"), nil
}

func (a *GeminiAdapter) Install(configPath string, opts InstallOptions) error {
	hookCmd := opts.HookBinaryPath
	if hookCmd == "" {
		hookCmd = "monocle"
	}
	scopeFlag := scopeFlagStr(opts.Scope)

	config, err := ReadJSONFile(configPath)
	if err != nil {
		return err
	}

	hooks := getOrCreateMap(config, "hooks")

	mergeJSONHookEvent(hooks, "AfterTool", map[string]any{
		"matcher": "",
		"hooks": []any{
			map[string]any{
				"type":    "command",
				"command": hookCmd + " hook" + scopeFlag + " after-tool --agent gemini",
				"async":   true,
			},
		},
	})

	mergeJSONHookEvent(hooks, "AfterAgent", map[string]any{
		"matcher": "",
		"hooks": []any{
			map[string]any{
				"type":    "command",
				"command": hookCmd + " hook" + scopeFlag + " after-agent --agent gemini",
			},
		},
	})

	return WriteJSONFile(configPath, config)
}

func (a *GeminiAdapter) Uninstall(configPath string) error {
	config, err := ReadJSONFile(configPath)
	if err != nil {
		return err
	}

	hooks, ok := config["hooks"].(map[string]any)
	if !ok {
		return nil
	}

	removeMonocleHooks(hooks, "AfterTool")
	removeMonocleHooks(hooks, "AfterAgent")

	if len(hooks) == 0 {
		delete(config, "hooks")
	}

	return WriteJSONFile(configPath, config)
}

// scopeFlagStr returns the --scope flag fragment to insert into hook commands.
// Returns "" for the default "repo" scope, " --scope cwd" for cwd scope.
func scopeFlagStr(scope string) string {
	if scope == "cwd" {
		return " --scope cwd"
	}
	return ""
}

// isMonocleCommand checks if a command string is a monocle hook command.
func isMonocleCommand(cmd string) bool {
	return strings.Contains(cmd, "monocle hook ")
}

// jsonHooksContainMonocle checks if any hook under the given event name
// contains a monocle hook command. Shared by JSON-based adapters.
func jsonHooksContainMonocle(config map[string]any, eventName string) bool {
	hooks, ok := config["hooks"].(map[string]any)
	if !ok {
		return false
	}
	entries, ok := hooks[eventName].([]any)
	if !ok {
		return false
	}
	for _, entry := range entries {
		em, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		innerHooks, ok := em["hooks"].([]any)
		if !ok {
			continue
		}
		for _, h := range innerHooks {
			hm, ok := h.(map[string]any)
			if !ok {
				continue
			}
			cmd, _ := hm["command"].(string)
			if isMonocleCommand(cmd) {
				return true
			}
		}
	}
	return false
}

// getOrCreateMap gets or creates a nested map[string]any at the given key.
func getOrCreateMap(m map[string]any, key string) map[string]any {
	if existing, ok := m[key].(map[string]any); ok {
		return existing
	}
	newMap := map[string]any{}
	m[key] = newMap
	return newMap
}

// mergeJSONHookEvent adds a hook entry to an event array if no monocle hook
// already exists for that event. Used by JSON-based adapters (Claude, Gemini).
func mergeJSONHookEvent(hooks map[string]any, eventName string, entry map[string]any) {
	var entries []any
	if existing, ok := hooks[eventName].([]any); ok {
		entries = existing
	}

	// Check if monocle hook already exists for this event
	for _, e := range entries {
		em, ok := e.(map[string]any)
		if !ok {
			continue
		}
		innerHooks, ok := em["hooks"].([]any)
		if !ok {
			continue
		}
		for _, h := range innerHooks {
			hm, ok := h.(map[string]any)
			if !ok {
				continue
			}
			cmd, _ := hm["command"].(string)
			if isMonocleCommand(cmd) {
				return // already installed
			}
		}
	}

	hooks[eventName] = append(entries, entry)
}

// removeMonocleHooks removes hook entries containing monocle hook commands
// from the given event array, and deletes the event key if empty.
func removeMonocleHooks(hooks map[string]any, eventName string) {
	entries, ok := hooks[eventName].([]any)
	if !ok {
		return
	}

	var kept []any
	for _, entry := range entries {
		em, ok := entry.(map[string]any)
		if !ok {
			kept = append(kept, entry)
			continue
		}
		innerHooks, ok := em["hooks"].([]any)
		if !ok {
			kept = append(kept, entry)
			continue
		}

		// Filter out monocle hooks from inner hooks array
		var keptInner []any
		for _, h := range innerHooks {
			hm, ok := h.(map[string]any)
			if !ok {
				keptInner = append(keptInner, h)
				continue
			}
			cmd, _ := hm["command"].(string)
			if !isMonocleCommand(cmd) {
				keptInner = append(keptInner, h)
			}
		}

		if len(keptInner) > 0 {
			em["hooks"] = keptInner
			kept = append(kept, em)
		}
	}

	if len(kept) == 0 {
		delete(hooks, eventName)
	} else {
		hooks[eventName] = kept
	}
}

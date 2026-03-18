package adapters

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/anthropics/monocle/internal/protocol"
)

// ClaudeAdapter handles Claude Code's hook format.
type ClaudeAdapter struct{}

// claudeHookInput represents Claude Code's hook stdin JSON.
type claudeHookInput struct {
	SessionID           string          `json:"session_id,omitempty"`
	Tool                *claudeToolInfo `json:"tool,omitempty"`
	StopReason          string          `json:"stop_reason,omitempty"`
	RequestID           string          `json:"request_id,omitempty"`
	Prompt              string          `json:"prompt,omitempty"`
	PermissionMode      string          `json:"permission_mode,omitempty"`
	LastAssistantMessage string         `json:"last_assistant_message,omitempty"`
}

type claudeToolInfo struct {
	Name   string `json:"name"`
	Input  string `json:"input,omitempty"`
	Output string `json:"output,omitempty"`
}

// claudeStopOutput is Claude Code's expected stdout for stop hooks.
type claudeStopOutput struct {
	Decision string `json:"decision,omitempty"` // "block" to send feedback, omit to release
	Reason   string `json:"reason,omitempty"`   // system message when blocking
}

func (a *ClaudeAdapter) ParseHookInput(event string, raw []byte) (any, error) {
	var input claudeHookInput
	if err := json.Unmarshal(raw, &input); err != nil {
		return nil, fmt.Errorf("parse claude input: %w", err)
	}

	switch event {
	case "post-tool-use":
		msg := &protocol.PostToolUseMsg{
			Type:  protocol.TypePostToolUse,
			Agent: "claude",
		}
		if input.Tool != nil {
			msg.Tool = input.Tool.Name
			msg.ToolInput = input.Tool.Input
			msg.ToolOutput = input.Tool.Output
		}
		return msg, nil

	case "stop":
		msg := &protocol.StopMsg{
			Type:       protocol.TypeStop,
			Agent:      "claude",
			StopReason: input.StopReason,
			RequestID:  input.RequestID,
		}
		if input.PermissionMode == "plan" && input.LastAssistantMessage != "" {
			msg.ReviewContent = input.LastAssistantMessage
			msg.ReviewContentTitle = "Plan"
			msg.ReviewContentType = "markdown"
		}
		return msg, nil

	case "prompt-submit":
		return &protocol.PromptSubmitMsg{
			Type:      protocol.TypePromptSubmit,
			Agent:     "claude",
			Prompt:    input.Prompt,
			RequestID: input.RequestID,
		}, nil

	default:
		return nil, fmt.Errorf("unknown event: %s", event)
	}
}

func (a *ClaudeAdapter) FormatHookOutput(response any) HookOutput {
	switch r := response.(type) {
	case *protocol.StopResponse:
		out := claudeStopOutput{}
		if r.Continue && r.SystemMessage != "" {
			out.Decision = "block"
			out.Reason = r.SystemMessage
		}
		data, _ := json.Marshal(out)
		return HookOutput{Data: data, ExitCode: 0}

	case *protocol.PromptSubmitResponse:
		data, _ := json.Marshal(map[string]string{
			"additional_context": r.AdditionalContext,
		})
		return HookOutput{Data: data, ExitCode: 0}

	default:
		return HookOutput{ExitCode: 0}
	}
}

func (a *ClaudeAdapter) GenerateConfig(opts SetupOptions) string {
	hookCmd := opts.HookBinaryPath
	if hookCmd == "" {
		hookCmd = "monocle"
	}

	config := map[string]any{
		"hooks": map[string]any{
			"PostToolUse": []map[string]any{
				{
					"matcher": "Write|Edit|MultiEdit",
					"hooks": []map[string]any{
						{
							"type":    "command",
							"command": hookCmd + " hook post-tool-use --agent claude",
							"async":   true,
						},
					},
				},
			},
			"Stop": []map[string]any{
				{
					"matcher": "",
					"hooks": []map[string]any{
						{
							"type":    "command",
							"command": hookCmd + " hook stop --agent claude",
						},
					},
				},
			},
		},
	}

	data, _ := json.MarshalIndent(config, "", "  ")
	return string(data)
}

func (a *ClaudeAdapter) Capabilities() AdapterCapabilities {
	return AdapterCapabilities{
		PostToolUse:  true,
		StopBlocking: true,
		AsyncHooks:   true,
	}
}

// AgentInstaller implementation

func (a *ClaudeAdapter) Name() string { return "claude" }

func (a *ClaudeAdapter) Detect() bool {
	if _, err := exec.LookPath("claude"); err == nil {
		return true
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	if info, err := os.Stat(filepath.Join(home, ".claude")); err == nil && info.IsDir() {
		return true
	}
	return false
}

func (a *ClaudeAdapter) ConfigPath(global bool) (string, bool) {
	if global {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", false
		}
		return filepath.Join(home, ".claude", "settings.json"), true
	}
	return filepath.Join(".claude", "settings.json"), true
}

func (a *ClaudeAdapter) IsInstalled(configPath string) (bool, error) {
	config, err := ReadJSONFile(configPath)
	if err != nil {
		return false, err
	}
	return jsonHooksContainMonocle(config, "PostToolUse") ||
		jsonHooksContainMonocle(config, "Stop"), nil
}

func (a *ClaudeAdapter) Install(configPath string, opts InstallOptions) error {
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

	mergeJSONHookEvent(hooks, "PostToolUse", map[string]any{
		"matcher": "Write|Edit|MultiEdit",
		"hooks": []any{
			map[string]any{
				"type":    "command",
				"command": hookCmd + " hook" + scopeFlag + " post-tool-use --agent claude",
				"async":   true,
			},
		},
	})

	mergeJSONHookEvent(hooks, "Stop", map[string]any{
		"matcher": "",
		"hooks": []any{
			map[string]any{
				"type":    "command",
				"command": hookCmd + " hook" + scopeFlag + " stop --agent claude",
			},
		},
	})

	return WriteJSONFile(configPath, config)
}

func (a *ClaudeAdapter) Uninstall(configPath string) error {
	config, err := ReadJSONFile(configPath)
	if err != nil {
		return err
	}

	hooks, ok := config["hooks"].(map[string]any)
	if !ok {
		return nil
	}

	removeMonocleHooks(hooks, "PostToolUse")
	removeMonocleHooks(hooks, "Stop")

	if len(hooks) == 0 {
		delete(config, "hooks")
	}

	return WriteJSONFile(configPath, config)
}

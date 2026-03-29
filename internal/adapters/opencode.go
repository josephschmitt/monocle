package adapters

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// OpenCodeAdapter handles MCP registration for OpenCode.
type OpenCodeAdapter struct{}

func (a *OpenCodeAdapter) Name() string  { return "opencode" }
func (a *OpenCodeAdapter) Label() string { return "OpenCode" }

func (a *OpenCodeAdapter) ConfigPaths(global bool) []string {
	paths := []string{openCodeConfigPath(global)}
	for _, name := range openCodeCommandNames {
		paths = append(paths, filepath.Join(openCodeCommandsDir(global), name+".md"))
	}
	return paths
}

func (a *OpenCodeAdapter) HasConfig(global bool) bool {
	path := openCodeConfigPath(global)
	data, err := ReadJSONFile(path)
	if err != nil {
		return false
	}
	mcp, ok := data["mcp"].(map[string]any)
	if !ok {
		return false
	}
	_, ok = mcp["monocle"].(map[string]any)
	return ok
}

func (a *OpenCodeAdapter) Register(global bool) error {
	command := ResolveCommand(global)

	// Write MCP config
	path := openCodeConfigPath(global)
	data, err := ReadJSONFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	if data["$schema"] == nil {
		data["$schema"] = "https://opencode.ai/config.json"
	}
	mcp, ok := data["mcp"].(map[string]any)
	if !ok {
		mcp = map[string]any{}
		data["mcp"] = mcp
	}
	mcp["monocle"] = map[string]any{
		"type":    "local",
		"command": []any{command, "serve-mcp-channel"},
		"enabled": true,
	}

	// Allow monocle MCP tools in plan mode so the agent can submit
	// content for review and check review status while planning.
	agent, ok := data["agent"].(map[string]any)
	if !ok {
		agent = map[string]any{}
		data["agent"] = agent
	}
	plan, ok := agent["plan"].(map[string]any)
	if !ok {
		plan = map[string]any{}
		agent["plan"] = plan
	}
	perm, ok := plan["permission"].(map[string]any)
	if !ok {
		perm = map[string]any{}
		plan["permission"] = perm
	}
	perm["mcp__monocle"] = "allow"

	if err := WriteJSONFile(path, data); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}

	// Write slash commands
	cmdDir := openCodeCommandsDir(global)
	for _, name := range openCodeCommandNames {
		content := openCodeCommands[name]
		cmdPath := filepath.Join(cmdDir, name+".md")
		if err := WriteFileAtomic(cmdPath, []byte(content)); err != nil {
			return fmt.Errorf("write %s: %w", cmdPath, err)
		}
	}

	return nil
}

func (a *OpenCodeAdapter) Unregister(global bool) error {
	// Remove MCP entry
	path := openCodeConfigPath(global)
	data, err := ReadJSONFile(path)
	if err != nil {
		return err
	}
	mcp, ok := data["mcp"].(map[string]any)
	if ok {
		delete(mcp, "monocle")
		if len(mcp) == 0 {
			delete(data, "mcp")
		}
	}
	// Remove $schema if it's the only remaining key
	if len(data) <= 1 {
		_ = RemoveFileIfExists(path)
	} else {
		_ = WriteJSONFile(path, data)
	}

	// Remove command files
	cmdDir := openCodeCommandsDir(global)
	for _, name := range openCodeCommandNames {
		_ = RemoveFileIfExists(filepath.Join(cmdDir, name+".md"))
	}

	return nil
}

// Detect returns true if OpenCode appears to be installed.
func (a *OpenCodeAdapter) Detect() bool {
	if _, err := exec.LookPath("opencode"); err == nil {
		return true
	}
	if _, err := os.Stat("opencode.json"); err == nil {
		return true
	}
	return false
}

func openCodeConfigPath(global bool) string {
	if global {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, ".config", "opencode", "opencode.json")
		}
	}
	return "opencode.json"
}

func openCodeCommandsDir(global bool) string {
	if global {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, ".config", "opencode", "commands")
		}
	}
	return filepath.Join(".opencode", "commands")
}

var openCodeCommandNames = []string{"get-feedback", "review-plan", "review-plan-wait"}

var openCodeCommands = map[string]string{
	"get-feedback": `---
description: Retrieve review feedback from Monocle
---

Run ` + "`monocle review get-feedback`" + ` to retrieve pending review feedback from your reviewer. If feedback is available, read it carefully and address the comments. If no feedback is pending, let the user know.
`,
	"review-plan": `---
description: Send a plan to Monocle for review
---

Submit a plan file to Monocle so your reviewer can see it. This does NOT wait for feedback — use ` + "`/review-plan-wait`" + ` if you need to block until the reviewer responds.

1. If the user provided a file path as an argument, use that. Otherwise, find the most recently modified plan file in the project.
2. Read the plan file to get its content and filename.
3. Run ` + "`monocle review send-artifact`" + ` with:
   - ` + "`--title`" + `: The first markdown heading from the plan, or the filename if no heading found
   - ` + "`--file`" + `: Absolute path to the plan file
   - ` + "`--id`" + `: The plan filename (so updates replace the previous version)
   - ` + "`--type`" + `: ` + "`md`" + `
4. Confirm to the user that the plan was sent to Monocle.
`,
	"review-plan-wait": `---
description: Send a plan to Monocle and wait for review feedback
---

Submit a plan file to Monocle and block until the reviewer responds with feedback. Use this when you need reviewer approval before proceeding.

1. If the user provided a file path as an argument, use that. Otherwise, find the most recently modified plan file in the project.
2. Read the plan file to get its content and filename.
3. Run ` + "`monocle review send-artifact --wait`" + ` with:
   - ` + "`--title`" + `: The first markdown heading from the plan, or the filename if no heading found
   - ` + "`--file`" + `: Absolute path to the plan file
   - ` + "`--id`" + `: The plan filename (so updates replace the previous version)
   - ` + "`--type`" + `: ` + "`md`" + `
4. Handle the response:
   - If approved with no comments, inform the user and continue.
   - If feedback requests changes, act on it — update the plan, then run ` + "`monocle review send-artifact --wait`" + ` again.
   - Keep iterating until the reviewer approves.
`,
}

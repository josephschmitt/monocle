package adapters

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// GeminiAdapter handles MCP registration for Google Gemini CLI.
type GeminiAdapter struct{}

func (a *GeminiAdapter) Name() string  { return "gemini" }
func (a *GeminiAdapter) Label() string { return "Gemini CLI" }

func (a *GeminiAdapter) ConfigPaths(global bool) []string {
	paths := []string{geminiConfigPath(global)}
	for _, name := range geminiCommandNames {
		paths = append(paths, filepath.Join(geminiCommandsDir(global), name+".toml"))
	}
	return paths
}

func (a *GeminiAdapter) HasConfig(global bool) bool {
	return hasMCPServersEntry(geminiConfigPath(global))
}

func (a *GeminiAdapter) Register(global bool) error {
	command := ResolveCommand(global)

	// Write MCP config (same mcpServers format as Claude)
	if err := configureMCPServersJSON(geminiConfigPath(global), command); err != nil {
		return fmt.Errorf("configure gemini: %w", err)
	}

	// Write slash commands
	cmdDir := geminiCommandsDir(global)
	for _, name := range geminiCommandNames {
		content := geminiCommands[name]
		cmdPath := filepath.Join(cmdDir, name+".toml")
		if err := WriteFileAtomic(cmdPath, []byte(content)); err != nil {
			return fmt.Errorf("write %s: %w", cmdPath, err)
		}
	}

	return nil
}

func (a *GeminiAdapter) Unregister(global bool) error {
	// Remove MCP entry
	if err := unconfigureMCPServersJSON(geminiConfigPath(global)); err != nil {
		return fmt.Errorf("unconfigure gemini: %w", err)
	}

	// Remove command files
	cmdDir := geminiCommandsDir(global)
	for _, name := range geminiCommandNames {
		_ = RemoveFileIfExists(filepath.Join(cmdDir, name+".toml"))
	}

	return nil
}

// Detect returns true if Gemini CLI appears to be installed.
func (a *GeminiAdapter) Detect() bool {
	if _, err := exec.LookPath("gemini"); err == nil {
		return true
	}
	if _, err := os.Stat(".gemini"); err == nil {
		return true
	}
	return false
}

func geminiConfigPath(global bool) string {
	if global {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, ".gemini", "settings.json")
		}
	}
	return filepath.Join(".gemini", "settings.json")
}

func geminiCommandsDir(global bool) string {
	if global {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, ".gemini", "commands")
		}
	}
	return filepath.Join(".gemini", "commands")
}

var geminiCommandNames = []string{"get-feedback", "review-plan", "review-plan-wait"}

var geminiCommands = map[string]string{
	"get-feedback": `description = "Retrieve review feedback from Monocle"
prompt = """
Run ` + "`monocle review get-feedback`" + ` to retrieve pending review feedback from your reviewer. If feedback is available, read it carefully and address the comments. If no feedback is pending, let the user know.
"""
`,
	"review-plan": `description = "Send a plan to Monocle for review"
prompt = """
Submit a plan file to Monocle so your reviewer can see it. This does NOT wait for feedback — use /review-plan-wait if you need to block until the reviewer responds.

Find the most recently modified plan file in the project (or use the path the user provided). Read it to get the filename and first heading.

Run ` + "`monocle review send-artifact`" + ` with:
- ` + "`--title`" + `: The first markdown heading from the plan, or the filename
- ` + "`--file`" + `: Absolute path to the plan file
- ` + "`--id`" + `: The plan filename (so updates replace the previous version)
- ` + "`--type`" + `: md

Confirm to the user that the plan was sent.
"""
`,
	"review-plan-wait": `description = "Send a plan to Monocle and wait for review feedback"
prompt = """
Submit a plan file to Monocle and block until the reviewer responds. Use this when you need reviewer approval before proceeding.

Find the most recently modified plan file in the project (or use the path the user provided). Read it to get the filename and first heading.

Run ` + "`monocle review send-artifact --wait`" + ` with:
- ` + "`--title`" + `: The first markdown heading from the plan, or the filename
- ` + "`--file`" + `: Absolute path to the plan file
- ` + "`--id`" + `: The plan filename (so updates replace the previous version)
- ` + "`--type`" + `: md

Handle the response:
- If approved, inform the user and continue.
- If feedback requests changes, update the plan and run ` + "`monocle review send-artifact --wait`" + ` again.
- Keep iterating until approved.
"""
`,
}

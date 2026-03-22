package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/alecthomas/kong"

	"github.com/anthropics/monocle/internal/adapters"
	"github.com/anthropics/monocle/internal/core"
	"github.com/anthropics/monocle/internal/db"
	"github.com/anthropics/monocle/internal/tui"
)

type CLI struct {
	Run       RunCmd       `cmd:"" default:"withargs" help:"Start a review session"`
	Install   InstallCmd   `cmd:"" help:"Install MCP channel for Claude Code"`
	Uninstall UninstallCmd `cmd:"" help:"Remove MCP channel for Claude Code"`
}

type RunCmd struct{}

type InstallCmd struct {
	Global bool `help:"Install to user-level ~/.mcp.json instead of project" default:"false"`
}

type UninstallCmd struct {
	Global bool `help:"Remove from user-level ~/.mcp.json instead of project" default:"false"`
}

func main() {
	cli := CLI{}
	ctx := kong.Parse(&cli,
		kong.Name("monocle"),
		kong.Description("Terminal-based code review companion for Claude Code"),
		kong.UsageOnError(),
	)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}

func (cmd *RunCmd) Run() error {
	return runTUI()
}

func (cmd *InstallCmd) Run() error {
	adapter := &adapters.ClaudeAdapter{}

	if !adapter.Detect() {
		fmt.Println("Claude Code not detected. Install Claude Code first.")
		return nil
	}

	installed, err := adapter.IsInstalled()
	if err != nil {
		return fmt.Errorf("check install: %w", err)
	}
	if installed {
		fmt.Println("  ✓ claude: already installed")
		return nil
	}

	if err := adapter.Install(cmd.Global); err != nil {
		return fmt.Errorf("install: %w", err)
	}

	fmt.Println("  ✓ claude: MCP channel installed")
	for _, detail := range adapter.InstallDetails(cmd.Global) {
		fmt.Printf("    %s\n", detail)
	}

	fmt.Println("\nMake sure 'monocle' is in your PATH.")
	return nil
}

func (cmd *UninstallCmd) Run() error {
	adapter := &adapters.ClaudeAdapter{}

	installed, err := adapter.IsInstalled()
	if err != nil {
		return fmt.Errorf("check install: %w", err)
	}
	if !installed {
		fmt.Println("  ✓ claude: nothing to remove")
		return nil
	}

	if err := adapter.Uninstall(cmd.Global); err != nil {
		return fmt.Errorf("uninstall: %w", err)
	}

	fmt.Println("  ✓ claude: MCP channel removed")
	return nil
}

func runTUI() error {
	// Load config
	cfg, err := core.LoadConfig()
	if err != nil {
		cfg = core.DefaultConfig()
	}

	// Open database
	database, err := db.Open(db.DBPath())
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer database.Close()

	// Get repo root
	repoRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get cwd: %w", err)
	}
	repoRoot = adapters.FindRepoRoot(repoRoot)

	// Create engine
	engine, err := core.NewEngine(cfg, database, repoRoot)
	if err != nil {
		return fmt.Errorf("create engine: %w", err)
	}

	// Start session
	opts := core.SessionOptions{
		Agent:    "claude",
		RepoRoot: repoRoot,
	}
	if _, err := engine.StartSession(opts); err != nil {
		return fmt.Errorf("start session: %w", err)
	}

	// Start socket server
	socketPath := adapters.DefaultSocketPath(repoRoot)
	if err := engine.StartServer(socketPath); err != nil {
		return fmt.Errorf("start server: %w", err)
	}

	// Check if MCP channel needs installation
	var appOpts tui.AppOptions
	adapter := &adapters.ClaudeAdapter{}
	if adapter.Detect() && adapter.NeedsInstall() {
		appOpts.MCPInstallFn = func() error {
			return adapter.Install(true) // global: ~/.mcp.json
		}
	}

	// Create TUI model
	app := tui.NewApp(engine, appOpts)

	// Create Bubble Tea program
	p := tea.NewProgram(app)

	// Bridge engine events to TUI
	tui.BridgeEngineEvents(engine, p)

	// Run program
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("run tui: %w", err)
	}

	// Cleanup
	engine.Shutdown()
	return nil
}

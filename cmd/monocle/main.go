package main

import (
	"fmt"
	"os"
	"syscall"

	tea "charm.land/bubbletea/v2"
	"github.com/alecthomas/kong"

	"github.com/anthropics/monocle/internal/adapters"
	"github.com/anthropics/monocle/internal/core"
	"github.com/anthropics/monocle/internal/db"
	"github.com/anthropics/monocle/internal/tui"
)

var version = "dev"

type CLI struct {
	Run             RunCmd             `cmd:"" default:"withargs" help:"Start a review session"`
	Register        RegisterCmd        `cmd:"" help:"Register MCP channel for Claude Code"`
	Unregister      UnregisterCmd      `cmd:"" help:"Remove MCP channel registration"`
	ServeMcpChannel ServeMCPChannelCmd `cmd:"serve-mcp-channel" help:"Run the MCP channel server" hidden:""`
	Install         InstallCmd         `cmd:"" help:"Install MCP channel (alias for register)" hidden:""`
	Uninstall       UninstallCmd       `cmd:"" help:"Remove MCP channel (alias for unregister)" hidden:""`
	Version         kong.VersionFlag   `help:"Print version" name:"version"`
}

type RunCmd struct {
	Socket string `help:"Override socket path for MCP channel connection" env:"MONOCLE_SOCKET" default:""`
}

type RegisterCmd struct {
	Global bool `help:"Register in user-level ~/.mcp.json instead of project" default:"false"`
}

type UnregisterCmd struct {
	Global bool `help:"Remove from user-level ~/.mcp.json instead of project" default:"false"`
}

type ServeMCPChannelCmd struct{}

type InstallCmd struct {
	Global bool `help:"Register in user-level ~/.mcp.json instead of project" default:"false"`
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
		kong.Vars{"version": version},
	)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}

func (cmd *RunCmd) Run() error {
	return runTUI(cmd.Socket)
}

func (cmd *RegisterCmd) Run() error {
	adapter := &adapters.ClaudeAdapter{}

	if !adapter.Detect() {
		fmt.Println("Claude Code not detected. Install Claude Code first.")
		return nil
	}

	if adapter.HasMCPConfig() {
		fmt.Println("  ✓ claude: MCP channel already registered")
		return nil
	}

	if err := adapter.Register(cmd.Global); err != nil {
		return fmt.Errorf("register: %w", err)
	}

	fmt.Println("  ✓ claude: MCP channel registered")
	for _, detail := range adapter.RegisterDetails(cmd.Global) {
		fmt.Printf("    %s\n", detail)
	}

	return nil
}

func (cmd *UnregisterCmd) Run() error {
	adapter := &adapters.ClaudeAdapter{}

	if !adapter.HasMCPConfig() {
		fmt.Println("  ✓ claude: nothing to remove")
		return nil
	}

	if err := adapter.Unregister(cmd.Global); err != nil {
		return fmt.Errorf("unregister: %w", err)
	}

	fmt.Println("  ✓ claude: MCP channel removed")
	return nil
}

func (cmd *ServeMCPChannelCmd) Run() error {
	// Write the embedded bundle to a temp file
	bundlePath, err := adapters.WriteChannelBundle()
	if err != nil {
		return err
	}

	// Detect JS runtime
	rt, err := adapters.DetectJSRuntime()
	if err != nil {
		return fmt.Errorf("monocle serve-mcp-channel requires a JavaScript runtime: %w", err)
	}

	// Exec into the JS runtime, replacing this process
	binPath, argv, err := rt.ExecArgs(bundlePath)
	if err != nil {
		return err
	}

	return syscall.Exec(binPath, argv, os.Environ())
}

// Deprecated: use 'monocle register' instead.
func (cmd *InstallCmd) Run() error {
	fmt.Fprintln(os.Stderr, "Note: 'monocle install' is deprecated, use 'monocle register' instead")
	return (&RegisterCmd{Global: cmd.Global}).Run()
}

// Deprecated: use 'monocle unregister' instead.
func (cmd *UninstallCmd) Run() error {
	fmt.Fprintln(os.Stderr, "Note: 'monocle uninstall' is deprecated, use 'monocle unregister' instead")
	return (&UnregisterCmd{Global: cmd.Global}).Run()
}

func runTUI(socketOverride string) error {
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
	socketPath := socketOverride
	if socketPath == "" {
		socketPath = adapters.DefaultSocketPath(repoRoot)
	}
	if err := engine.StartServer(socketPath); err != nil {
		return fmt.Errorf("start server: %w", err)
	}

	// Check if MCP channel needs registration
	var appOpts tui.AppOptions
	adapter := &adapters.ClaudeAdapter{}
	if adapter.Detect() && adapter.NeedsRegister() {
		appOpts.MCPRegisterFn = func(global bool) error {
			return adapter.Register(global)
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

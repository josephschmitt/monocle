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
	Start    StartCmd    `cmd:"" default:"withargs" help:"Start a new review session"`
	Resume   ResumeCmd   `cmd:"" help:"Resume an existing session"`
	Sessions SessionsCmd `cmd:"" help:"List sessions"`
	Install  InstallCmd  `cmd:"" help:"Install MCP channel for Claude Code"`
	Uninstall UninstallCmd `cmd:"" help:"Remove MCP channel for Claude Code"`
}

type StartCmd struct {
	Agent string `help:"Agent type" default:"claude"`
}

type ResumeCmd struct {
	SessionID string `arg:"" help:"Session ID to resume"`
}

type SessionsCmd struct {
	Repo string `help:"Filter by repo root" default:"."`
}

type InstallCmd struct{}

type UninstallCmd struct{}

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

func (cmd *StartCmd) Run() error {
	return runTUI(cmd.Agent, "")
}

func (cmd *ResumeCmd) Run() error {
	return runTUI("", cmd.SessionID)
}

func (cmd *SessionsCmd) Run() error {
	database, err := db.Open(db.DBPath())
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer database.Close()

	cfg := core.DefaultConfig()
	repoRoot, _ := os.Getwd()
	engine, err := core.NewEngine(cfg, database, repoRoot)
	if err != nil {
		return fmt.Errorf("create engine: %w", err)
	}

	sessions, err := engine.ListSessions(core.ListSessionsOptions{RepoRoot: cmd.Repo})
	if err != nil {
		return fmt.Errorf("list sessions: %w", err)
	}

	if len(sessions) == 0 {
		fmt.Println("No sessions found.")
		return nil
	}

	for _, s := range sessions {
		fmt.Printf("%s  %s  %s  %d files  %d comments  round %d  %s\n",
			s.ID[:8], s.Agent, s.RepoRoot, s.FileCount, s.CommentCount, s.ReviewRound, s.UpdatedAt.Format("2006-01-02 15:04"))
	}
	return nil
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

	if err := adapter.Install(); err != nil {
		return fmt.Errorf("install: %w", err)
	}

	fmt.Println("  ✓ claude: MCP channel installed")
	for _, detail := range adapter.InstallDetails() {
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

	if err := adapter.Uninstall(); err != nil {
		return fmt.Errorf("uninstall: %w", err)
	}

	fmt.Println("  ✓ claude: MCP channel removed")
	return nil
}

func runTUI(agent, sessionID string) error {
	// Load config
	cfg, err := core.LoadConfig()
	if err != nil {
		cfg = core.DefaultConfig()
	}

	if agent != "" {
		cfg.DefaultAgent = agent
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

	// Start or resume session
	if sessionID != "" {
		if _, err := engine.ResumeSession(sessionID); err != nil {
			return fmt.Errorf("resume session: %w", err)
		}
	} else {
		opts := core.SessionOptions{
			Agent:    cfg.DefaultAgent,
			RepoRoot: repoRoot,
		}
		if _, err := engine.StartSession(opts); err != nil {
			return fmt.Errorf("start session: %w", err)
		}
	}

	// Start socket server
	socketPath := adapters.DefaultSocketPath(repoRoot)
	if err := engine.StartServer(socketPath); err != nil {
		return fmt.Errorf("start server: %w", err)
	}

	// Create TUI model
	app := tui.NewApp(engine)

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

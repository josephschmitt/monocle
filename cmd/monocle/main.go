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
	Setup    SetupCmd    `cmd:"" help:"Install agent hooks"`
	Review   ReviewCmd   `cmd:"" help:"Send content for review"`
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

type SetupCmd struct {
	Agent     string `arg:"" help:"Agent to configure (claude|gemini|codex|opencode)"`
	Uninstall bool   `help:"Remove hooks instead of installing"`
}

type ReviewCmd struct {
	File  string `arg:"" help:"File to review"`
	ID    string `help:"Content item ID"`
	Title string `help:"Content title"`
	Type  string `help:"Content type" default:"text"`
}

func main() {
	cli := CLI{}
	ctx := kong.Parse(&cli,
		kong.Name("monocle"),
		kong.Description("Terminal-based code review companion for AI coding agents"),
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

func (cmd *SetupCmd) Run() error {
	adapter := adapters.GetAdapter(cmd.Agent)
	config := adapter.GenerateConfig(adapters.SetupOptions{})

	if cmd.Uninstall {
		fmt.Printf("To uninstall, remove the monocle hooks from your %s configuration.\n", cmd.Agent)
		return nil
	}

	fmt.Printf("Add the following to your %s hooks configuration:\n\n", cmd.Agent)
	fmt.Println(config)
	fmt.Println()
	fmt.Println("Make sure 'monocle-hook' is in your PATH.")
	return nil
}

func (cmd *ReviewCmd) Run() error {
	// TODO: implement content review piping
	fmt.Println("Review command not yet implemented")
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

	// Start hook server
	socketPath := cfg.Hooks.SocketPath
	if socketPath == "" {
		socketPath = fmt.Sprintf("/tmp/monocle-%d.sock", os.Getpid())
	}
	if err := engine.StartHookServer(socketPath); err != nil {
		return fmt.Errorf("start hook server: %w", err)
	}

	// Set env var for hook shim
	os.Setenv("MONOCLE_SOCKET", socketPath) //nolint:errcheck

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

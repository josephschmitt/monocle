package main

import (
	"fmt"
	"io"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/alecthomas/kong"

	"github.com/anthropics/monocle/internal/adapters"
	"github.com/anthropics/monocle/internal/core"
	"github.com/anthropics/monocle/internal/db"
	"github.com/anthropics/monocle/internal/protocol"
	"github.com/anthropics/monocle/internal/tui"
)

type CLI struct {
	Start     StartCmd     `cmd:"" default:"withargs" help:"Start a new review session"`
	Resume    ResumeCmd    `cmd:"" help:"Resume an existing session"`
	Sessions  SessionsCmd  `cmd:"" help:"List sessions"`
	Install   InstallCmd   `cmd:"" help:"Install agent hooks"`
	Uninstall UninstallCmd `cmd:"" help:"Remove agent hooks"`
	Hook      HookCmd      `cmd:"" hidden:"" help:"Hook shim invoked by agent hooks"`
	Review    ReviewCmd    `cmd:"" help:"Send content for review"`
}

type StartCmd struct {
	Agent string `help:"Agent type" default:"claude"`
	Scope string `help:"Socket scope: repo (default) or cwd (for monorepos)" enum:"repo,cwd" default:"repo"`
}

type ResumeCmd struct {
	SessionID string `arg:"" help:"Session ID to resume"`
}

type SessionsCmd struct {
	Repo string `help:"Filter by repo root" default:"."`
}

type InstallCmd struct {
	Agents []string `arg:"" optional:"" default:"auto" help:"Agents to install (auto, claude, codex, gemini, opencode)"`
	Global bool     `help:"Install to global/user config instead of project" default:"false"`
	Scope  string   `help:"Socket scope: repo (default) or cwd (for monorepos)" enum:"repo,cwd" default:"repo"`
}

type UninstallCmd struct {
	Agents []string `arg:"" optional:"" default:"auto" help:"Agents to uninstall (auto, claude, codex, gemini, opencode)"`
	Global bool     `help:"Remove from global/user config" default:"false"`
}

type HookCmd struct {
	Event string `arg:"" help:"Hook event name"`
	Agent string `help:"Agent name" default:"claude"`
	Scope string `help:"Socket scope: repo or cwd" default:"repo" enum:"repo,cwd"`
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
	return runTUI(cmd.Agent, "", cmd.Scope)
}

func (cmd *ResumeCmd) Run() error {
	return runTUI("", cmd.SessionID, "")
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
	cfg, err := core.LoadConfig()
	if err != nil {
		cfg = core.DefaultConfig()
	}
	scope := resolveScope(cmd.Scope, cfg.Hooks.Scope)
	results := adapters.InstallAgents(cmd.Agents, cmd.Global, scope)

	if len(results) == 0 {
		fmt.Println("No agents detected. Specify an agent explicitly: monocle install claude")
		return nil
	}

	for _, r := range results {
		if r.Err != nil {
			fmt.Printf("  ✗ %s: %s\n", r.Agent, r.Err)
		} else if r.AlreadyDone {
			fmt.Printf("  ✓ %s: already installed (%s)\n", r.Agent, r.ConfigPath)
		} else if r.Installed {
			fmt.Printf("  ✓ %s: hooks installed → %s\n", r.Agent, r.ConfigPath)
		}
	}

	fmt.Println("\nMake sure 'monocle' is in your PATH.")
	return nil
}

func (cmd *UninstallCmd) Run() error {
	results := adapters.UninstallAgents(cmd.Agents, cmd.Global)

	if len(results) == 0 {
		fmt.Println("No agents detected. Specify an agent explicitly: monocle uninstall claude")
		return nil
	}

	for _, r := range results {
		if r.Err != nil {
			fmt.Printf("  ✗ %s: %s\n", r.Agent, r.Err)
		} else if r.AlreadyDone {
			fmt.Printf("  ✓ %s: no hooks to remove\n", r.Agent)
		} else if r.Installed {
			fmt.Printf("  ✓ %s: hooks removed from %s\n", r.Agent, r.ConfigPath)
		}
	}
	return nil
}

func (cmd *HookCmd) Run() error {
	// Read stdin completely.
	raw, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil // never block the agent
	}

	// Parse the input via the adapter.
	adapter := adapters.GetAdapter(cmd.Agent)
	msg, err := adapter.ParseHookInput(cmd.Event, raw)
	if err != nil {
		return nil
	}

	// Get the socket path from the environment or compute deterministically.
	socketPath := os.Getenv("MONOCLE_SOCKET")
	if socketPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil
		}
		dir := cwd
		if cmd.Scope != "cwd" {
			dir = adapters.FindRepoRoot(cwd)
		}
		socketPath = adapters.DefaultSocketPath(dir)
	}

	// Connect to the engine.
	client, err := adapters.NewSocketClient(socketPath)
	if err != nil {
		return nil
	}
	defer client.Close()

	if protocol.IsBlocking(msg) {
		// Send and wait for a response.
		response, err := client.SendAndWait(msg)
		if err != nil {
			return nil
		}

		out := adapter.FormatHookOutput(response)
		if len(out.Data) > 0 {
			os.Stdout.Write(out.Data) //nolint:errcheck
		}
		if out.ExitCode != 0 {
			os.Exit(out.ExitCode)
		}
		return nil
	}

	// Non-blocking: fire and forget.
	_ = client.Send(msg)
	return nil
}

func (cmd *ReviewCmd) Run() error {
	// TODO: implement content review piping
	fmt.Println("Review command not yet implemented")
	return nil
}

func runTUI(agent, sessionID, cmdScope string) error {
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

	// Get repo root, resolving scope
	repoRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get cwd: %w", err)
	}

	scope := resolveScope(cmdScope, cfg.Hooks.Scope)
	if scope != "cwd" {
		repoRoot = adapters.FindRepoRoot(repoRoot)
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
		socketPath = adapters.DefaultSocketPath(repoRoot)
	}
	if err := engine.StartHookServer(socketPath); err != nil {
		return fmt.Errorf("start hook server: %w", err)
	}

	// Set env var for hook shim (child process inheritance)
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

// resolveScope returns the effective scope: CLI flag > config > "repo" default.
func resolveScope(flagScope, cfgScope string) string {
	if flagScope == "cwd" {
		return "cwd"
	}
	if cfgScope == "cwd" {
		return "cwd"
	}
	return "repo"
}

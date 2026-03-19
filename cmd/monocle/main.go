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
	Start         StartCmd         `cmd:"" default:"withargs" help:"Start a new review session"`
	Resume        ResumeCmd        `cmd:"" help:"Resume an existing session"`
	Sessions      SessionsCmd      `cmd:"" help:"List sessions"`
	Install       InstallCmd       `cmd:"" help:"Install agent skills"`
	Uninstall     UninstallCmd     `cmd:"" help:"Remove agent skills"`
	ReviewStatus  ReviewStatusCmd  `cmd:"" name:"review-status" help:"Check for pending review feedback"`
	GetFeedback   GetFeedbackCmd   `cmd:"" name:"get-feedback" help:"Retrieve review feedback"`
	SubmitContent SubmitContentCmd `cmd:"" name:"submit-content" help:"Submit content for review"`
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

type InstallCmd struct {
	Agents []string `arg:"" optional:"" default:"auto" help:"Agents to install (auto, claude, codex, gemini, opencode)"`
	Global bool     `help:"Install to global/user config instead of project" default:"false"`
}

type UninstallCmd struct {
	Agents []string `arg:"" optional:"" default:"auto" help:"Agents to uninstall (auto, claude, codex, gemini, opencode)"`
	Global bool     `help:"Remove from global/user config" default:"false"`
}

type ReviewStatusCmd struct{}

type GetFeedbackCmd struct {
	Wait bool `help:"Block until feedback is available" default:"false"`
}

type SubmitContentCmd struct {
	Title string `help:"Content title" required:""`
	ID    string `help:"Content item ID (for updates)" default:""`
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

func (cmd *InstallCmd) Run() error {
	results := adapters.InstallAgents(cmd.Agents, cmd.Global)

	if len(results) == 0 {
		fmt.Println("No agents detected. Specify an agent explicitly: monocle install claude")
		return nil
	}

	for _, r := range results {
		if r.Err != nil {
			fmt.Printf("  ✗ %s: %s\n", r.Agent, r.Err)
		} else if r.AlreadyDone {
			fmt.Printf("  ✓ %s: already installed (%s)\n", r.Agent, r.SkillPath)
		} else if r.Installed {
			fmt.Printf("  ✓ %s: skill installed → %s\n", r.Agent, r.SkillPath)
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
			fmt.Printf("  ✓ %s: no skill to remove\n", r.Agent)
		} else if r.Installed {
			fmt.Printf("  ✓ %s: skill removed from %s\n", r.Agent, r.SkillPath)
		}
	}
	return nil
}

func (cmd *ReviewStatusCmd) Run() error {
	socketPath, err := resolveSocketPath()
	if err != nil {
		fmt.Println("No reviewer connected.")
		return nil
	}

	client, err := adapters.NewSocketClient(socketPath)
	if err != nil {
		fmt.Println("No reviewer connected.")
		return nil
	}
	defer client.Close()

	resp, err := client.SendAndWait(&protocol.GetReviewStatusMsg{
		Type: protocol.TypeGetReviewStatus,
	})
	if err != nil {
		fmt.Println("No reviewer connected.")
		return nil
	}

	status, ok := resp.(*protocol.GetReviewStatusResponse)
	if !ok {
		fmt.Println("No reviewer connected.")
		return nil
	}

	fmt.Println(status.Summary)
	return nil
}

func (cmd *GetFeedbackCmd) Run() error {
	socketPath, err := resolveSocketPath()
	if err != nil {
		fmt.Println("No reviewer connected.")
		return nil
	}

	client, err := adapters.NewSocketClient(socketPath)
	if err != nil {
		fmt.Println("No reviewer connected.")
		return nil
	}
	defer client.Close()

	if cmd.Wait {
		fmt.Fprintln(os.Stderr, "Waiting for reviewer to submit feedback...")
	}

	resp, err := client.SendAndWait(&protocol.PollFeedbackMsg{
		Type: protocol.TypePollFeedback,
		Wait: cmd.Wait,
	})
	if err != nil {
		fmt.Println("No reviewer connected.")
		return nil
	}

	feedback, ok := resp.(*protocol.PollFeedbackResponse)
	if !ok {
		fmt.Println("No reviewer connected.")
		return nil
	}

	if feedback.HasFeedback {
		fmt.Print(feedback.Feedback)
	} else {
		fmt.Println("No feedback pending.")
	}
	return nil
}

func (cmd *SubmitContentCmd) Run() error {
	// Read content from stdin
	content, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("read stdin: %w", err)
	}

	socketPath, err := resolveSocketPath()
	if err != nil {
		fmt.Println("No reviewer connected.")
		return nil
	}

	client, err := adapters.NewSocketClient(socketPath)
	if err != nil {
		fmt.Println("No reviewer connected.")
		return nil
	}
	defer client.Close()

	resp, err := client.SendAndWait(&protocol.SubmitContentMsg{
		Type:    protocol.TypeSubmitContent,
		ID:      cmd.ID,
		Title:   cmd.Title,
		Content: string(content),
	})
	if err != nil {
		fmt.Println("No reviewer connected.")
		return nil
	}

	submitResp, ok := resp.(*protocol.SubmitContentResponse)
	if !ok {
		fmt.Println("No reviewer connected.")
		return nil
	}

	fmt.Println(submitResp.Message)
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

// resolveSocketPath discovers the socket path for CLI subcommands.
func resolveSocketPath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := adapters.FindRepoRoot(cwd)
	return adapters.DefaultSocketPath(dir), nil
}

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"

	tea "charm.land/bubbletea/v2"
	"github.com/alecthomas/kong"

	"github.com/josephschmitt/monocle/internal/adapters"
	"github.com/josephschmitt/monocle/internal/client"
	"github.com/josephschmitt/monocle/internal/core"
	"github.com/josephschmitt/monocle/internal/db"
	"github.com/josephschmitt/monocle/internal/protocol"
	"github.com/josephschmitt/monocle/internal/tui"
)

var version = "dev"

type CLI struct {
	Run             RunCmd             `cmd:"" default:"withargs" help:"Start a review session"`
	Review          ReviewCmd          `cmd:"review" help:"Commands for interacting with a Monocle review session"`
	Register        RegisterCmd        `cmd:"" help:"Register Monocle for an agent"`
	Unregister      UnregisterCmd      `cmd:"" help:"Remove Monocle registration"`
	ServeMcpChannel ServeMCPChannelCmd `cmd:"serve-mcp-channel" help:"Run the MCP channel server" hidden:""`
	Install         InstallCmd         `cmd:"" help:"Install MCP channel (alias for register)" hidden:""`
	Uninstall       UninstallCmd       `cmd:"" help:"Remove MCP channel (alias for unregister)" hidden:""`
	Version         kong.VersionFlag   `help:"Print version" name:"version"`
}

// ReviewCmd groups agent-facing subcommands for interacting with a running Monocle session.
type ReviewCmd struct {
	Status       ReviewStatusCmd       `cmd:"status" help:"Check the current review status"`
	GetFeedback  ReviewGetFeedbackCmd  `cmd:"get-feedback" help:"Retrieve review feedback"`
	SendArtifact ReviewSendArtifactCmd `cmd:"send-artifact" help:"Send content to the reviewer"`
	AddFiles     ReviewAddFilesCmd     `cmd:"add-files" help:"Add files to the review session"`
}

type ReviewStatusCmd struct {
	Socket string `help:"Override socket path" env:"MONOCLE_SOCKET" default:""`
	JSON   bool   `help:"Output as JSON" default:"false"`
}

type ReviewGetFeedbackCmd struct {
	Socket string `help:"Override socket path" env:"MONOCLE_SOCKET" default:""`
	Wait   bool   `help:"Block until feedback is available" default:"false"`
	JSON   bool   `help:"Output as JSON" default:"false"`
}

type ReviewSendArtifactCmd struct {
	Socket      string `help:"Override socket path" env:"MONOCLE_SOCKET" default:""`
	Title       string `help:"Title for the content" required:""`
	File        string `help:"Path to file to submit" type:"path" default:""`
	ID          string `help:"ID for updating existing content" default:""`
	ContentType string `help:"File extension for syntax highlighting (md, go, py, ts)" name:"type" default:""`
	Wait        bool   `help:"Block until reviewer responds with feedback" default:"false"`
	JSON        bool   `help:"Output as JSON" default:"false"`
}

type ReviewAddFilesCmd struct {
	Socket string   `help:"Override socket path" env:"MONOCLE_SOCKET" default:""`
	Paths  []string `arg:"" required:"" help:"File or directory paths to add for review"`
	JSON   bool     `help:"Output as JSON" default:"false"`
}

type RunCmd struct {
	Socket         string   `help:"Override socket path for MCP channel connection" env:"MONOCLE_SOCKET" default:""`
	AdditionalPath []string `help:"Additional file or directory paths to include for review (repeatable)" name:"additional-path" short:"a" type:"path"`
	Continue       bool     `help:"Resume the most recent session for this repo" name:"continue" short:"c" xor:"session-mode"`
	Resume         bool     `help:"Show a picker to resume a previous session" name:"resume" short:"r" xor:"session-mode"`
	Session        string   `help:"Resume a specific session by ID" name:"session" short:"s" default:""`
}

type RegisterCmd struct {
	Agent  string `arg:"" optional:"" help:"Agent to register (claude, opencode, codex, gemini, all)"`
	Global bool   `help:"Register in user-level config instead of project" default:"false"`
}

type UnregisterCmd struct {
	Agent  string `arg:"" optional:"" help:"Agent to unregister (claude, opencode, codex, gemini, all)"`
	Global bool   `help:"Remove from user-level config instead of project" default:"false"`
}

type ServeMCPChannelCmd struct{}

type InstallCmd struct {
	Agent  string `arg:"" optional:"" help:"Agent to register (claude, opencode, codex, gemini, all)"`
	Global bool   `help:"Register in user-level config instead of project" default:"false"`
}

type UninstallCmd struct {
	Agent  string `arg:"" optional:"" help:"Agent to unregister (claude, opencode, codex, gemini, all)"`
	Global bool   `help:"Remove from user-level config instead of project" default:"false"`
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
	return runTUI(cmd.Socket, cmd.AdditionalPath, cmd.Continue, cmd.Resume, cmd.Session)
}

func (cmd *RegisterCmd) Run() error {
	agents, err := resolveAgents(cmd.Agent, "Select agents to register")
	if err != nil {
		return err
	}
	if len(agents) == 0 {
		return nil // user cancelled picker
	}

	for _, a := range agents {
		wasRegistered := a.HasConfig(cmd.Global)
		if err := a.Register(cmd.Global); err != nil {
			return fmt.Errorf("register %s: %w", a.Name(), err)
		}
		if wasRegistered {
			fmt.Printf("  ✓ %s: updated\n", a.Name())
		} else {
			fmt.Printf("  ✓ %s: registered\n", a.Name())
		}
		for _, p := range a.ConfigPaths(cmd.Global) {
			fmt.Printf("    → %s\n", p)
		}
	}
	return nil
}

func (cmd *UnregisterCmd) Run() error {
	agents, err := resolveAgents(cmd.Agent, "Select agents to unregister")
	if err != nil {
		return err
	}
	if len(agents) == 0 {
		return nil
	}

	for _, a := range agents {
		if !a.HasConfig(cmd.Global) {
			fmt.Printf("  ✓ %s: nothing to remove\n", a.Name())
			continue
		}
		if err := a.Unregister(cmd.Global); err != nil {
			return fmt.Errorf("unregister %s: %w", a.Name(), err)
		}
		fmt.Printf("  ✓ %s: removed\n", a.Name())
	}
	return nil
}

func resolveAgents(name, pickerTitle string) ([]adapters.AgentAdapter, error) {
	switch name {
	case "":
		return adapters.PickAgents(adapters.AllAdapters(), pickerTitle)
	case "all":
		return adapters.AllAdapters(), nil
	default:
		a, err := adapters.GetAdapter(name)
		if err != nil {
			return nil, err
		}
		return []adapters.AgentAdapter{a}, nil
	}
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
	return (&RegisterCmd{Agent: cmd.Agent, Global: cmd.Global}).Run()
}

// Deprecated: use 'monocle unregister' instead.
func (cmd *UninstallCmd) Run() error {
	fmt.Fprintln(os.Stderr, "Note: 'monocle uninstall' is deprecated, use 'monocle unregister' instead")
	return (&UnregisterCmd{Agent: cmd.Agent, Global: cmd.Global}).Run()
}

// -- Review subcommand implementations --

func (cmd *ReviewStatusCmd) Run() error {
	c, err := client.ConnectWithOverride(cmd.Socket)
	if err != nil {
		if errors.Is(err, client.ErrNotRunning) {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return err
	}
	defer c.Close()

	resp, err := c.Request(
		&protocol.GetReviewStatusMsg{Type: protocol.TypeGetReviewStatus},
		client.DefaultTimeout,
	)
	if err != nil {
		return fmt.Errorf("review status: %w", err)
	}

	status := resp.(*protocol.GetReviewStatusResponse)
	if cmd.JSON {
		return printJSON(status)
	}
	if status.Summary != "" {
		fmt.Println(status.Summary)
	} else {
		fmt.Println(status.Status)
	}
	return nil
}

func (cmd *ReviewGetFeedbackCmd) Run() error {
	c, err := client.ConnectWithOverride(cmd.Socket)
	if err != nil {
		if errors.Is(err, client.ErrNotRunning) {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return err
	}
	defer c.Close()

	timeout := client.DefaultTimeout
	if cmd.Wait {
		timeout = 0 // no deadline — block until feedback
	}

	resp, err := c.Request(
		&protocol.PollFeedbackMsg{Type: protocol.TypePollFeedback, Wait: cmd.Wait},
		timeout,
	)
	if err != nil {
		return fmt.Errorf("get feedback: %w", err)
	}

	feedback := resp.(*protocol.PollFeedbackResponse)
	if cmd.JSON {
		return printJSON(feedback)
	}
	if !feedback.HasFeedback {
		fmt.Println("No feedback pending.")
		return nil
	}
	fmt.Println(feedback.Feedback)
	return nil
}

func (cmd *ReviewSendArtifactCmd) Run() error {
	// Resolve content: --file, or stdin
	var content string
	if cmd.File != "" {
		data, err := os.ReadFile(cmd.File)
		if err != nil {
			return fmt.Errorf("read file: %w", err)
		}
		content = string(data)
		// Default ID to filename if not set
		if cmd.ID == "" {
			cmd.ID = filepath.Base(cmd.File)
		}
	} else {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("read stdin: %w", err)
		}
		content = string(data)
		if content == "" {
			return fmt.Errorf("no content: provide --file or pipe content to stdin")
		}
	}

	c, err := client.ConnectWithOverride(cmd.Socket)
	if err != nil {
		if errors.Is(err, client.ErrNotRunning) {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return err
	}
	defer c.Close()

	resp, err := c.Request(
		&protocol.SubmitContentMsg{
			Type:        protocol.TypeSubmitContent,
			ID:          cmd.ID,
			Title:       cmd.Title,
			Content:     content,
			ContentType: cmd.ContentType,
			IsPlan:      true,
		},
		client.DefaultTimeout,
	)
	if err != nil {
		return fmt.Errorf("send artifact: %w", err)
	}

	submit := resp.(*protocol.SubmitContentResponse)
	if !cmd.Wait {
		if cmd.JSON {
			return printJSON(submit)
		}
		fmt.Println(submit.Message)
		return nil
	}

	// --wait: open a new connection and block for feedback
	c.Close()
	c2, err := client.ConnectWithOverride(cmd.Socket)
	if err != nil {
		return fmt.Errorf("reconnect for wait: %w", err)
	}
	defer c2.Close()

	feedbackResp, err := c2.Request(
		&protocol.PollFeedbackMsg{Type: protocol.TypePollFeedback, Wait: true},
		0, // no deadline
	)
	if err != nil {
		return fmt.Errorf("wait for feedback: %w", err)
	}

	feedback := feedbackResp.(*protocol.PollFeedbackResponse)
	if cmd.JSON {
		return printJSON(feedback)
	}
	if !feedback.HasFeedback {
		fmt.Println("Approved. No feedback from reviewer.")
		return nil
	}
	fmt.Println(feedback.Feedback)
	return nil
}

func (cmd *ReviewAddFilesCmd) Run() error {
	// Resolve paths to absolute
	absPaths := make([]string, len(cmd.Paths))
	for i, p := range cmd.Paths {
		abs, err := filepath.Abs(p)
		if err != nil {
			return fmt.Errorf("resolve path %q: %w", p, err)
		}
		absPaths[i] = abs
	}

	c, err := client.ConnectWithOverride(cmd.Socket)
	if err != nil {
		if errors.Is(err, client.ErrNotRunning) {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return err
	}
	defer c.Close()

	resp, err := c.Request(
		&protocol.AddAdditionalFilesMsg{
			Type:  protocol.TypeAddAdditionalFiles,
			Paths: absPaths,
		},
		client.DefaultTimeout,
	)
	if err != nil {
		return fmt.Errorf("add files: %w", err)
	}

	add := resp.(*protocol.AddAdditionalFilesResponse)
	if cmd.JSON {
		return printJSON(add)
	}
	fmt.Println(add.Message)
	return nil
}

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func startNewSession(engine core.EngineAPI, repoRoot string) error {
	opts := core.SessionOptions{
		Agent:    "claude",
		RepoRoot: repoRoot,
	}
	if _, err := engine.StartSession(opts); err != nil {
		return fmt.Errorf("start session: %w", err)
	}
	return nil
}

func resolveSession(engine core.EngineAPI, repoRoot string, continueSession bool, resumePicker bool, sessionID string) error {
	switch {
	case sessionID != "":
		// Direct session ID provided via --session
		if _, err := engine.ResumeSession(sessionID); err != nil {
			return fmt.Errorf("resume session %s: %w", sessionID, err)
		}
		return nil

	case continueSession:
		sessions, err := engine.ListSessions(core.ListSessionsOptions{
			RepoRoot: repoRoot,
			Limit:    1,
		})
		if err != nil || len(sessions) == 0 {
			return startNewSession(engine, repoRoot)
		}
		if _, err := engine.ResumeSession(sessions[0].ID); err != nil {
			return fmt.Errorf("resume session: %w", err)
		}
		return nil

	case resumePicker:
		// --resume: defer session creation until user picks in the TUI modal
		return nil

	default:
		return startNewSession(engine, repoRoot)
	}
}

func runTUI(socketOverride string, additionalPaths []string, continueSession bool, resumePicker bool, sessionID string) error {
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
	nonGitMode := !adapters.IsGitRepo(repoRoot)

	// Create engine
	engine, err := core.NewEngine(cfg, database, repoRoot, nonGitMode)
	if err != nil {
		return fmt.Errorf("create engine: %w", err)
	}

	// Resolve session: continue, resume, or new
	if nonGitMode {
		// In non-git mode, always start a fresh session
		if err := startNewSession(engine, repoRoot); err != nil {
			return err
		}
	} else if err := resolveSession(engine, repoRoot, continueSession, resumePicker, sessionID); err != nil {
		return err
	}

	// Reload any pending (undelivered) feedback from a previous session
	if continueSession || sessionID != "" {
		engine.ReloadPendingFeedback()
	}

	// Add additional file paths if provided (only for new sessions)
	if len(additionalPaths) > 0 && !continueSession && !resumePicker && sessionID == "" {
		if _, err := engine.AddAdditionalPaths(additionalPaths); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not add additional paths: %v\n", err)
		}
	}

	// Start socket server (deferred when showing session picker)
	socketPath := socketOverride
	if socketPath == "" {
		socketPath = adapters.DefaultSocketPath(repoRoot)
	}
	if !resumePicker {
		if err := engine.StartServer(socketPath); err != nil {
			return fmt.Errorf("start server: %w", err)
		}
	}

	// Check if MCP channel needs registration
	var appOpts tui.AppOptions
	appOpts.NonGitMode = nonGitMode
	adapter := &adapters.ClaudeAdapter{}
	if adapter.Detect() && adapter.NeedsRegister() {
		appOpts.MCPRegisterFn = func(global bool) error {
			return adapter.Register(global)
		}
	}
	if resumePicker {
		appOpts.ShowSessionPicker = true
		appOpts.RepoRoot = repoRoot
		appOpts.DeferredSocket = socketPath
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

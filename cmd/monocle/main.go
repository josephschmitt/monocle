package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"
	"github.com/alecthomas/kong"

	"github.com/josephschmitt/monocle/internal/adapters"
	"github.com/josephschmitt/monocle/internal/client"
	"github.com/josephschmitt/monocle/internal/core"
	monocleMCP "github.com/josephschmitt/monocle/internal/mcp"
	"github.com/josephschmitt/monocle/internal/protocol"
	"github.com/josephschmitt/monocle/internal/tui"
	"github.com/josephschmitt/monocle/internal/tui/register"
)

var version = "dev"

type CLI struct {
	Run             RunCmd             `cmd:"" default:"withargs" help:"Start a review session" hidden:""`
	Review          ReviewCmd          `cmd:"review" help:"Commands for interacting with a Monocle review session"`
	Register        RegisterCmd        `cmd:"" help:"Register Monocle for an agent"`
	Unregister      UnregisterCmd      `cmd:"" help:"Remove Monocle registration"`
	Hooks           HooksCmd           `cmd:"" help:"Hook handlers for agent lifecycle events (invoked by the agent harness)"`
	Serve           ServeCmd           `cmd:"serve" help:"Run a headless Monocle engine for this repo (socket server, no TUI)"`
	Stop            StopCmd            `cmd:"stop" help:"Stop a running monocle serve process for this repo"`
	ServeMcp        ServeMCPCmd        `cmd:"serve-mcp" help:"Run the MCP server" hidden:""`
	ServeMcpChannel ServeMCPChannelCmd `cmd:"serve-mcp-channel" help:"Run the MCP channel server (deprecated)" hidden:""`
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

// WorkDirFlag is embedded by commands that support --workdir.
type WorkDirFlag struct {
	WorkDir string `help:"Override working directory (pair with a repo at this path)" name:"workdir" short:"C" type:"path" default:"" env:"MONOCLE_WORKDIR"`
}

type ReviewStatusCmd struct {
	WorkDirFlag
	Socket string `help:"Override socket path" env:"MONOCLE_SOCKET" default:""`
	JSON   bool   `help:"Output as JSON" default:"false"`
}

type ReviewGetFeedbackCmd struct {
	WorkDirFlag
	Socket string `help:"Override socket path" env:"MONOCLE_SOCKET" default:""`
	Wait   bool   `help:"Block until feedback is available" default:"false"`
	JSON   bool   `help:"Output as JSON" default:"false"`
}

type ReviewSendArtifactCmd struct {
	WorkDirFlag
	Socket      string `help:"Override socket path" env:"MONOCLE_SOCKET" default:""`
	Title       string `help:"Title for the content" required:""`
	File        string `help:"Path to file to submit" type:"path" default:""`
	ID          string `help:"ID for updating existing content" default:""`
	ContentType string `help:"File extension for syntax highlighting (md, go, py, ts)" name:"type" default:""`
	Wait        bool   `help:"Block until reviewer responds with feedback" default:"false"`
	JSON        bool   `help:"Output as JSON" default:"false"`
}

type ReviewAddFilesCmd struct {
	WorkDirFlag
	Socket string   `help:"Override socket path" env:"MONOCLE_SOCKET" default:""`
	Paths  []string `arg:"" required:"" help:"File or directory paths to add for review"`
	JSON   bool     `help:"Output as JSON" default:"false"`
}

type RunCmd struct {
	WorkDirFlag
	Socket         string   `help:"Override socket path for MCP channel connection" env:"MONOCLE_SOCKET" default:""`
	AdditionalPath []string `help:"Additional file or directory paths to include for review (repeatable)" name:"additional-path" short:"a" type:"path"`
	Continue       bool     `help:"Resume the most recent session for this repo" name:"continue" short:"c" xor:"session-mode"`
	Resume         bool     `help:"Show a picker to resume a previous session" name:"resume" short:"r" xor:"session-mode"`
	Session        string   `help:"Resume a specific session by ID" name:"session" short:"s" default:""`
}

type RegisterCmd struct {
	Agent           string `arg:"" optional:"" help:"Agent to register (claude, opencode, codex, gemini, all)"`
	Global          bool   `help:"Register in user-level config instead of project" default:"false"`
	IntegrationMode string `help:"Override the default integration mode (auto, mcp, or skills)" enum:"auto,mcp,skills" default:"auto"`
	NoPlanHook      bool   `help:"Skip installing the Claude Code ExitPlanMode hook" name:"no-plan-hook" default:"false"`
	NoReviewGate    bool   `help:"Skip installing the Claude Code turn-end review-gate hooks (PostToolUse mark-activity + Stop on-stop)" name:"no-review-gate" default:"false"`
	NoTUI           bool   `help:"Skip the interactive wizard and run headlessly" name:"no-tui" default:"false"`
}

type UnregisterCmd struct {
	Agent          string `arg:"" optional:"" help:"Agent to unregister (claude, opencode, codex, gemini, all)"`
	Global         bool   `help:"Remove from user-level config instead of project" default:"false"`
	KeepPlanHook   bool   `help:"Leave the Claude Code ExitPlanMode hook entries in settings.json" name:"keep-plan-hook" default:"false"`
	KeepReviewGate bool   `help:"Leave the Claude Code turn-end review-gate hooks in settings.json" name:"keep-review-gate" default:"false"`
	NoTUI          bool   `help:"Skip the interactive wizard and run headlessly" name:"no-tui" default:"false"`
}

type ServeMCPCmd struct {
	ExperimentalChannels     bool `help:"Enable experimental MCP channel push notifications alongside tools" default:"false"`
	ExperimentalChannelsOnly bool `help:"Enable channels only (no tools) for agents using skills" default:"false"`
}

// ServeMCPChannelCmd is the deprecated MCP channel command.
// Kept as a hidden alias for backward compatibility; will be wired to serve-mcp.
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
	adapters.Version = version

	cli := CLI{}
	ctx := kong.Parse(&cli,
		kong.Name("monocle"),
		kong.Description("Terminal-based code review companion for Claude Code"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{Tree: true}),
		kong.Vars{"version": version},
	)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}

func (cmd *RunCmd) Run() error {
	return runTUI(cmd.Socket, cmd.WorkDir, cmd.AdditionalPath, cmd.Continue, cmd.Resume, cmd.Session)
}

func (cmd *RegisterCmd) Run() error {
	allAdapters := adapters.AllAdapters()

	// Interactive wizard path: no positional agent and --no-tui wasn't set.
	if cmd.Agent == "" && !cmd.NoTUI {
		return cmd.runWizard(allAdapters)
	}

	return cmd.runHeadless(allAdapters)
}

// runWizard launches the register TUI, letting the user pick agents and
// options, then runs the registrations from within the wizard.
func (cmd *RegisterCmd) runWizard(allAdapters []adapters.AgentAdapter) error {
	// Pre-apply modes so ConfigPaths() previews correctly in the wizard.
	for _, a := range allAdapters {
		a.SetMode(cmd.resolveMode(a))
	}
	opts := register.Options{
		Mode:                  register.ModeRegister,
		Adapters:              allAdapters,
		Global:                cmd.Global,
		GlobalLocked:          cmd.Global, // only locked when explicitly set to true
		IntegrationMode:       integrationChoice(cmd.IntegrationMode),
		IntegrationModeLocked: cmd.IntegrationMode != "auto",
		SkipPlanHook:          cmd.NoPlanHook,
		SkipPlanHookLocked:    cmd.NoPlanHook,
		SkipReviewGate:        cmd.NoReviewGate,
		SkipReviewGateLocked:  cmd.NoReviewGate,
	}
	res, err := register.Run(opts)
	if err != nil {
		return err
	}
	if res.Cancelled {
		return nil
	}
	return reportWizardResults(res)
}

// runHeadless preserves the pre-wizard behavior for scripted use.
func (cmd *RegisterCmd) runHeadless(allAdapters []adapters.AgentAdapter) error {
	for _, a := range allAdapters {
		a.SetMode(cmd.resolveMode(a))
	}

	agents, err := resolveAgentsFrom(allAdapters, cmd.Agent)
	if err != nil {
		return err
	}
	if len(agents) == 0 {
		return nil
	}

	for _, a := range agents {
		claude, ok := a.(*adapters.ClaudeAdapter)
		if !ok {
			continue
		}
		if cmd.NoPlanHook {
			claude.SkipPlanHook = true
		}
		if cmd.NoReviewGate {
			claude.SkipReviewGate = true
		}
	}

	for _, a := range agents {
		wasRegistered := a.HasConfig(cmd.Global)
		if err := a.Register(cmd.Global); err != nil {
			return fmt.Errorf("register %s: %w", a.Name(), err)
		}

		action := "registered"
		if wasRegistered {
			action = "updated"
		}
		mode := cmd.resolveMode(a)
		modeLabel := "skills"
		if mode == adapters.ModeMCPTools {
			modeLabel = "mcp tools"
		}
		fmt.Printf("  ✓ %s: %s (%s)\n", a.Label(), action, modeLabel)
		for _, p := range a.ConfigPaths(cmd.Global) {
			fmt.Printf("    → %s\n", p)
		}
	}
	return nil
}

// resolveMode returns the integration mode for the given agent.
// "auto" uses per-agent defaults: Claude → MCP tools, others → skills.
func (cmd *RegisterCmd) resolveMode(a adapters.AgentAdapter) adapters.IntegrationMode {
	switch cmd.IntegrationMode {
	case "mcp":
		return adapters.ModeMCPTools
	case "skills":
		return adapters.ModeSkills
	default: // auto
		if a.Name() == "claude" {
			return adapters.ModeMCPTools
		}
		return adapters.ModeSkills
	}
}

func (cmd *UnregisterCmd) Run() error {
	allAdapters := adapters.AllAdapters()

	if cmd.Agent == "" && !cmd.NoTUI {
		return cmd.runWizard(allAdapters)
	}

	return cmd.runHeadless(allAdapters)
}

func (cmd *UnregisterCmd) runWizard(allAdapters []adapters.AgentAdapter) error {
	// Offer only adapters that actually have config at the requested scope —
	// the wizard's "nothing to remove" state is the empty-list rendering.
	var registered []adapters.AgentAdapter
	for _, a := range allAdapters {
		if a.HasConfig(cmd.Global) {
			registered = append(registered, a)
		}
	}
	if len(registered) == 0 {
		fmt.Println("Nothing to unregister at this scope.")
		return nil
	}
	opts := register.Options{
		Mode:                 register.ModeUnregister,
		Adapters:             registered,
		Global:               cmd.Global,
		GlobalLocked:         cmd.Global,
		KeepPlanHook:         cmd.KeepPlanHook,
		KeepPlanHookLocked:   cmd.KeepPlanHook,
		KeepReviewGate:       cmd.KeepReviewGate,
		KeepReviewGateLocked: cmd.KeepReviewGate,
	}
	res, err := register.Run(opts)
	if err != nil {
		return err
	}
	if res.Cancelled {
		return nil
	}
	return reportWizardResults(res)
}

func (cmd *UnregisterCmd) runHeadless(allAdapters []adapters.AgentAdapter) error {
	agents, err := resolveAgentsFrom(allAdapters, cmd.Agent)
	if err != nil {
		return err
	}
	if len(agents) == 0 {
		return nil
	}

	for _, a := range agents {
		if claude, ok := a.(*adapters.ClaudeAdapter); ok {
			claude.KeepPlanHook = cmd.KeepPlanHook
			claude.KeepReviewGate = cmd.KeepReviewGate
		}
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

// integrationChoice maps the --integration-mode flag string to the wizard's
// IntegrationChoice enum.
func integrationChoice(flag string) register.IntegrationChoice {
	switch flag {
	case "mcp":
		return register.IntegrationMCP
	case "skills":
		return register.IntegrationSkills
	}
	return register.IntegrationAuto
}

// reportWizardResults prints a summary line for each agent the wizard ran.
// Mirrors the headless path's output so scripts that parse stdout (loosely)
// see the same shape regardless of which path ran.
func reportWizardResults(res register.Result) error {
	for _, r := range res.Results {
		if r.Err != nil {
			fmt.Printf("  ✗ %s: %v\n", r.Label, r.Err)
			continue
		}
		switch r.Action {
		case "nothing":
			fmt.Printf("  ✓ %s: nothing to remove\n", r.Label)
		case "removed":
			fmt.Printf("  ✓ %s: removed\n", r.Label)
		default:
			fmt.Printf("  ✓ %s: %s\n", r.Label, r.Action)
			for _, p := range r.Paths {
				fmt.Printf("    → %s\n", p)
			}
		}
	}
	return nil
}

// resolveAgentsFrom picks agents by name for the headless register/unregister
// paths. The empty-name case (interactive selection) is handled by the wizard
// callers; reaching it here means --no-tui was passed without an agent, which
// we treat as "all" for CI friendliness.
func resolveAgentsFrom(agents []adapters.AgentAdapter, name string) ([]adapters.AgentAdapter, error) {
	switch name {
	case "", "all":
		return agents, nil
	default:
		for _, a := range agents {
			if a.Name() == name {
				return []adapters.AgentAdapter{a}, nil
			}
		}
		return nil, fmt.Errorf("unknown agent %q (valid: claude, opencode, codex, gemini)", name)
	}
}

func (cmd *ServeMCPCmd) Run() error {
	monocleMCP.Version = version
	return monocleMCP.Run(monocleMCP.Options{
		EnableChannels: cmd.ExperimentalChannels,
		ChannelsOnly:   cmd.ExperimentalChannelsOnly,
	})
}

// Deprecated: use 'monocle serve-mcp --experimental-channels-only' instead.
func (cmd *ServeMCPChannelCmd) Run() error {
	fmt.Fprintln(os.Stderr, "Note: 'monocle serve-mcp-channel' is deprecated, use 'monocle serve-mcp --experimental-channels-only' instead")
	monocleMCP.Version = version
	return monocleMCP.Run(monocleMCP.Options{ChannelsOnly: true})
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
	socketPath, err := resolveSocketForWorkDir(cmd.Socket, cmd.WorkDir)
	if err != nil {
		return err
	}
	c, err := client.Connect(socketPath)
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
	socketPath, err := resolveSocketForWorkDir(cmd.Socket, cmd.WorkDir)
	if err != nil {
		return err
	}
	c, err := client.Connect(socketPath)
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

	socketPath, err := resolveSocketForWorkDir(cmd.Socket, cmd.WorkDir)
	if err != nil {
		return err
	}
	c, err := client.Connect(socketPath)
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
		// Echo the (possibly server-minted) id on its own line so the
		// agent can later address the artifact by id — without this the
		// minted UUID returned in submit.ID is silently lost.
		if submit.ID != "" {
			fmt.Printf("id: %s\n", submit.ID)
		}
		return nil
	}

	// Echo the (possibly server-minted) id before blocking on feedback so
	// an agent that sent an empty --id can still address the artifact later
	// (mark reviewed, dismiss, fetch versions). Without this the minted
	// UUID in submit.ID is lost on the --wait path, mirroring the bug the
	// non-wait branch already fixed.
	if !cmd.JSON && submit.ID != "" {
		fmt.Printf("id: %s\n", submit.ID)
	}

	// --wait: open a new connection and block for feedback
	c.Close()
	c2, err := client.Connect(socketPath)
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

	socketPath, err := resolveSocketForWorkDir(cmd.Socket, cmd.WorkDir)
	if err != nil {
		return err
	}
	c, err := client.Connect(socketPath)
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

// resolveRepoRoot returns the repo root and git mode.
// If workdir is non-empty it is used instead of the current working directory.
func resolveRepoRoot(workdir string) (repoRoot string, nonGitMode bool, err error) {
	if workdir != "" {
		info, err := os.Stat(workdir)
		if err != nil {
			return "", false, fmt.Errorf("--workdir: %w", err)
		}
		if !info.IsDir() {
			return "", false, fmt.Errorf("--workdir: %s is not a directory", workdir)
		}
		repoRoot = adapters.FindRepoRoot(workdir)
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			return "", false, fmt.Errorf("get cwd: %w", err)
		}
		repoRoot = adapters.FindRepoRoot(cwd)
	}
	// Check for .git directly — repoRoot is already resolved, so skip
	// IsGitRepo which would redundantly call FindRepoRoot again.
	_, statErr := os.Stat(filepath.Join(repoRoot, ".git"))
	nonGitMode = statErr != nil
	return repoRoot, nonGitMode, nil
}

// resolveSocketForWorkDir computes the socket path considering --socket and --workdir.
// Precedence: explicit socket > workdir-derived > CWD-derived.
func resolveSocketForWorkDir(socketOverride, workdir string) (string, error) {
	if socketOverride != "" {
		return socketOverride, nil
	}
	if workdir != "" {
		info, err := os.Stat(workdir)
		if err != nil {
			return "", fmt.Errorf("--workdir: %w", err)
		}
		if !info.IsDir() {
			return "", fmt.Errorf("--workdir: %s is not a directory", workdir)
		}
		return adapters.DefaultSocketPath(adapters.FindRepoRoot(workdir)), nil
	}
	return adapters.ResolveSocketPath(), nil
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

func runTUI(socketOverride string, workdir string, additionalPaths []string, continueSession bool, resumePicker bool, sessionID string) error {
	// Resolve repo root — use --workdir if provided, otherwise CWD.
	// nonGitMode informs the sidebar rendering but the real engine lives
	// in `monocle serve`, which performs its own repo-root resolution.
	repoRoot, nonGitMode, err := resolveRepoRoot(workdir)
	if err != nil {
		return err
	}

	// Auto-spawn `monocle serve` for this repo if none is running, then
	// connect as a thin client. The engine (SQLite, session, feedback
	// queue, socket server) all live in the serve process; the TUI holds
	// no engine state.
	socketPath, _, err := adapters.EnsureServe(adapters.AutoSpawnOptions{
		RepoRoot: repoRoot,
		Socket:   socketOverride,
	})
	if err != nil {
		return fmt.Errorf("ensure serve: %w", err)
	}

	engine, err := client.NewEngineClient(socketPath)
	if err != nil {
		return fmt.Errorf("connect engine: %w", err)
	}
	defer engine.Close()

	// Session selection. monocle serve defaults to "continue latest or
	// start new", so the zero-flag case needs no client-side action.
	switch {
	case sessionID != "":
		if _, err := engine.ResumeSession(sessionID); err != nil {
			return fmt.Errorf("resume session %s: %w", sessionID, err)
		}
		engine.ReloadPendingFeedback()
	case continueSession:
		sessions, err := engine.ListSessions(core.ListSessionsOptions{RepoRoot: repoRoot, Limit: 1})
		if err == nil && len(sessions) > 0 {
			if _, err := engine.ResumeSession(sessions[0].ID); err != nil {
				return fmt.Errorf("resume session: %w", err)
			}
		}
		engine.ReloadPendingFeedback()
	case resumePicker:
		// Picker modal drives ListSessions + ResumeSession itself.
	}

	// Only attach additional paths on a fresh-start flow. Resume variants
	// preserve the paths that were persisted with the session.
	if len(additionalPaths) > 0 && !continueSession && !resumePicker && sessionID == "" {
		if _, err := engine.AddAdditionalPaths(additionalPaths); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not add additional paths: %v\n", err)
		}
	}

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
	}

	app := tui.NewApp(engine, appOpts)
	p := tea.NewProgram(app)
	tui.BridgeEngineEvents(engine, p)

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("run tui: %w", err)
	}
	// Note: don't call engine.Shutdown() — serve owns its own lifecycle
	// (idle timeout, PID file) and other clients may still be attached.
	return nil
}

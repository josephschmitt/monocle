package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/josephschmitt/monocle/internal/client"
	"github.com/josephschmitt/monocle/internal/protocol"
)

// blockingHookMaxWait bounds how long the Stop and ExitPlanMode hooks block
// on the reviewer when no pause has been requested. The engine can outlive an
// active reviewer (autospawn keeps the socket alive for a grace period + idle
// timeout), so without a bound the hook would block for the installed hook
// timeout (historically 96h) and the agent would appear hung between turns.
// When the reviewer has explicitly requested a pause, the engine ignores this
// bound and blocks until they submit, preserving the pause flow.
const blockingHookMaxWait = 10 * time.Second

// HooksCmd groups subcommands invoked by an agent harness in response to
// lifecycle events (plan mode entry/exit, etc). Each subcommand reads the
// agent's hook payload on stdin and emits the agent's expected decision
// JSON on stdout. The caller is always an automated hook runner, not a
// human; all error paths exit 0 with empty stdout so the agent degrades
// to its default behavior rather than hard-blocking.
type HooksCmd struct {
	ExitPlan     ExitPlanHookCmd     `cmd:"" name:"exit-plan" help:"Handle an agent's plan-mode exit: send the plan to the Monocle reviewer and approve or deny based on feedback."`
	EnterPlan    EnterPlanHookCmd    `cmd:"" name:"enter-plan" help:"Inject review context into the agent right before it begins planning."`
	MarkActivity MarkActivityHookCmd `cmd:"" name:"mark-activity" help:"Notify the Monocle engine that a write-tool just fired, marking the current session as having unreviewed changes."`
	OnStop       OnStopHookCmd       `cmd:"" name:"on-stop" help:"Block the agent's turn-end until the reviewer approves or requests changes \u2014 but only if the turn included file changes."`
}

// ExitPlanHookCmd handles the agent's plan-mode exit event. For Claude Code
// this is the PermissionRequest hook matched on the ExitPlanMode tool. The
// hook blocks until the reviewer submits in the Monocle TUI, then emits an
// allow/deny decision in the format the invoking agent expects.
type ExitPlanHookCmd struct {
	WorkDirFlag
	Agent  string `help:"Agent whose hook is invoking this command." required:"" enum:"claude"`
	Socket string `help:"Override socket path" env:"MONOCLE_SOCKET" default:""`
}

// EnterPlanHookCmd handles the agent's pre-plan event. For Claude Code this
// is the PreToolUse hook matched on the ExitPlanMode tool — it fires right
// before Claude begins drafting its plan. We inject a short context string
// pointing out that the eventual ExitPlanMode will be gated by a Monocle
// reviewer, plus any pending reviewer feedback the agent should address.
type EnterPlanHookCmd struct {
	WorkDirFlag
	Agent  string `help:"Agent whose hook is invoking this command." required:"" enum:"claude"`
	Socket string `help:"Override socket path" env:"MONOCLE_SOCKET" default:""`
}

// hookInput is the common subset of Claude Code's hook payload we care about.
// Other agents will need their own decoder when they're added to the --agent
// enum.
type hookInput struct {
	SessionID string        `json:"session_id"`
	CWD       string        `json:"cwd"`
	ToolName  string        `json:"tool_name"`
	ToolInput hookToolInput `json:"tool_input"`
}

type hookToolInput struct {
	Plan string `json:"plan"`
	// PlanFilePath is Claude Code's path to the on-disk plan file backing
	// this ExitPlanMode call (e.g. "/.../claude/plans/my-plan.md"). When
	// present it is stable across revisions of the same plan within a
	// session, so we use its basename as the engine ID for upsert-style
	// versioning.
	PlanFilePath string `json:"planFilePath"`
}

// hookDebug writes a debug line to the file named by MONOCLE_HOOK_DEBUG.
// No-op when the env var is unset. Used to diagnose silent-fallback paths
// when Claude Code invokes the hook but nothing appears to happen.
func hookDebug(format string, args ...any) {
	path := os.Getenv("MONOCLE_HOOK_DEBUG")
	if path == "" {
		return
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "[%s] ", time.Now().Format("15:04:05.000"))
	fmt.Fprintf(f, format, args...)
	fmt.Fprintln(f)
}

func decodeHookInput(r io.Reader) (hookInput, error) {
	var in hookInput
	data, err := io.ReadAll(r)
	if err != nil {
		return in, err
	}
	if len(data) == 0 {
		return in, errors.New("empty stdin")
	}
	hookDebug("stdin payload: %s", string(data))
	if err := json.Unmarshal(data, &in); err != nil {
		return in, fmt.Errorf("parse hook input: %w", err)
	}
	return in, nil
}

// firstHeading returns the first markdown H1/H2 from the plan body, stripped
// of leading "#" characters. Falls back to empty string.
func firstHeading(body string) string {
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") || strings.HasPrefix(trimmed, "## ") {
			return strings.TrimSpace(strings.TrimLeft(trimmed, "#"))
		}
	}
	return ""
}

func (cmd *ExitPlanHookCmd) Run() error {
	hookDebug("exit-plan: invoked, agent=%q socket=%q workdir=%q", cmd.Agent, cmd.Socket, cmd.WorkDir)
	in, err := decodeHookInput(os.Stdin)
	if err != nil {
		hookDebug("exit-plan: decode stdin failed: %v", err)
		return nil
	}
	hookDebug("exit-plan: decoded session=%q cwd=%q tool=%q plan_len=%d", in.SessionID, in.CWD, in.ToolName, len(in.ToolInput.Plan))

	switch cmd.Agent {
	case "claude":
		return cmd.runClaude(in)
	default:
		hookDebug("exit-plan: unsupported agent %q", cmd.Agent)
		return nil
	}
}

func (cmd *ExitPlanHookCmd) runClaude(in hookInput) error {
	plan := in.ToolInput.Plan
	if plan == "" {
		hookDebug("exit-plan/claude: empty plan body, exiting silently")
		return nil
	}

	workdir := cmd.WorkDir
	if workdir == "" && in.CWD != "" {
		workdir = in.CWD
	}

	socketPath, err := resolveSocketForWorkDir(cmd.Socket, workdir)
	if err != nil {
		hookDebug("exit-plan/claude: socket resolve failed: %v (workdir=%q)", err, workdir)
		return nil
	}
	hookDebug("exit-plan/claude: resolved socket=%q (workdir=%q)", socketPath, workdir)

	planID := filepath.Base(in.ToolInput.PlanFilePath)
	if planID == "" || planID == "." || planID == "/" {
		planID = fmt.Sprintf("exit-plan-%s.md", in.SessionID)
	}
	hookDebug("exit-plan/claude: using plan id=%q (from planFilePath=%q)", planID, in.ToolInput.PlanFilePath)
	title := firstHeading(plan)
	if title == "" {
		title = planID
	}

	submit, err := client.Connect(socketPath)
	if err != nil {
		hookDebug("exit-plan/claude: connect failed: %v", err)
		return nil
	}
	if _, err := submit.Request(
		&protocol.SubmitContentMsg{
			Type:        protocol.TypeSubmitContent,
			ID:          planID,
			Title:       title,
			Content:     plan,
			ContentType: "md",
			IsPlan:      true,
		},
		client.DefaultTimeout,
	); err != nil {
		hookDebug("exit-plan/claude: submit request failed: %v", err)
		submit.Close()
		return nil
	}
	submit.Close()
	hookDebug("exit-plan/claude: submitted, now blocking on feedback")

	// Second connection for the blocking poll — the socket server rejects
	// overlapping blocking calls on the same connection.
	wait, err := client.Connect(socketPath)
	if err != nil {
		hookDebug("exit-plan/claude: reconnect for poll failed: %v", err)
		return nil
	}
	defer wait.Close()

	resp, err := wait.Request(
		&protocol.PollFeedbackMsg{
			Type:      protocol.TypePollFeedback,
			Wait:      true,
			MaxWaitMs: int(blockingHookMaxWait.Milliseconds()),
		},
		0,
	)
	if err != nil {
		hookDebug("exit-plan/claude: poll request failed: %v", err)
		return nil
	}
	feedback, ok := resp.(*protocol.PollFeedbackResponse)
	if !ok {
		hookDebug("exit-plan/claude: poll response is not PollFeedbackResponse (got %T)", resp)
		return nil
	}
	hookDebug("exit-plan/claude: got feedback action=%q has_feedback=%v comment_count=%d body_len=%d",
		feedback.Action, feedback.HasFeedback, feedback.CommentCount, len(feedback.Feedback))

	return emitClaudePermissionDecision(os.Stdout, feedback)
}

// emitClaudePermissionDecision writes the Claude Code PermissionRequest hook
// response that reflects the reviewer's action.
func emitClaudePermissionDecision(w io.Writer, feedback *protocol.PollFeedbackResponse) error {
	decision := map[string]any{"behavior": "allow"}
	if feedback.Action == "request_changes" {
		msg := feedback.Feedback
		if msg == "" {
			msg = "Reviewer requested changes."
		}
		decision = map[string]any{"behavior": "deny", "message": msg}
	}
	payload := map[string]any{
		"hookSpecificOutput": map[string]any{
			"hookEventName": "PermissionRequest",
			"decision":      decision,
		},
	}
	buf, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	hookDebug("exit-plan/claude: emitting decision=%s", string(buf))
	_, err = fmt.Fprintln(w, string(buf))
	return err
}

func (cmd *EnterPlanHookCmd) Run() error {
	in, err := decodeHookInput(os.Stdin)
	if err != nil {
		return nil
	}

	switch cmd.Agent {
	case "claude":
		return cmd.runClaude(in)
	default:
		return nil
	}
}

func (cmd *EnterPlanHookCmd) runClaude(in hookInput) error {
	workdir := cmd.WorkDir
	if workdir == "" && in.CWD != "" {
		workdir = in.CWD
	}

	socketPath, err := resolveSocketForWorkDir(cmd.Socket, workdir)
	if err != nil {
		return nil
	}

	// If the engine isn't reachable, stay out of Claude's way — the user
	// isn't running monocle and doesn't want plan-mode injection.
	c, err := client.Connect(socketPath)
	if err != nil {
		return nil
	}
	resp, reqErr := c.Request(
		&protocol.GetReviewStatusMsg{Type: protocol.TypeGetReviewStatus},
		2*time.Second,
	)
	c.Close()

	context := "Monocle is running for this session. When you submit a plan via ExitPlanMode, it will be sent to a human reviewer who can approve or request changes — the approval flow is automatic, you do not need to run any review commands yourself."
	if reqErr == nil {
		if status, ok := resp.(*protocol.GetReviewStatusResponse); ok {
			if status.Status == "pending" || status.CommentCount > 0 {
				context += fmt.Sprintf(" There are %d unaddressed reviewer comment(s) — read them before finalizing the plan.", status.CommentCount)
			}
		}
	}

	return emitClaudePreToolUseContext(os.Stdout, context)
}

func emitClaudePreToolUseContext(w io.Writer, context string) error {
	payload := map[string]any{
		"hookSpecificOutput": map[string]any{
			"hookEventName":     "PreToolUse",
			"additionalContext": context,
		},
	}
	return json.NewEncoder(w).Encode(payload)
}

// MarkActivityHookCmd notifies the Monocle engine that a write-tool just
// fired in the current session. Registered as the PostToolUse hook with
// matcher Edit|Write|NotebookEdit|MultiEdit. Runs with a tight timeout
// (5s) and must not block the agent — exits 0 empty on every error.
type MarkActivityHookCmd struct {
	WorkDirFlag
	Agent  string `help:"Agent whose hook is invoking this command." required:"" enum:"claude"`
	Socket string `help:"Override socket path" env:"MONOCLE_SOCKET" default:""`
}

func (cmd *MarkActivityHookCmd) Run() error {
	hookDebug("mark-activity: invoked, agent=%q", cmd.Agent)
	in, err := decodeHookInput(os.Stdin)
	if err != nil {
		hookDebug("mark-activity: decode stdin failed: %v", err)
		return nil
	}
	if cmd.Agent != "claude" {
		return nil
	}

	workdir := cmd.WorkDir
	if workdir == "" && in.CWD != "" {
		workdir = in.CWD
	}
	socketPath, err := resolveSocketForWorkDir(cmd.Socket, workdir)
	if err != nil {
		hookDebug("mark-activity: socket resolve failed: %v", err)
		return nil
	}

	c, err := client.Connect(socketPath)
	if err != nil {
		hookDebug("mark-activity: connect failed: %v", err)
		return nil
	}
	defer c.Close()

	if _, err := c.Request(
		&protocol.MarkActivityMsg{Type: protocol.TypeMarkActivity},
		client.DefaultTimeout,
	); err != nil {
		hookDebug("mark-activity: request failed: %v", err)
		return nil
	}
	hookDebug("mark-activity: session marked dirty")
	return nil
}

// OnStopHookCmd gates Claude's turn-end on reviewer approval when the turn
// made reviewable file changes. Registered as the Stop hook. If the session
// isn't dirty (no write-tools fired), the hook exits 0 and the turn ends
// normally. If the session is dirty, the hook blocks until the reviewer
// submits feedback; "request_changes" becomes a Stop-hook block decision
// that sends Claude back to address the feedback.
type OnStopHookCmd struct {
	WorkDirFlag
	Agent  string `help:"Agent whose hook is invoking this command." required:"" enum:"claude"`
	Socket string `help:"Override socket path" env:"MONOCLE_SOCKET" default:""`
}

func (cmd *OnStopHookCmd) Run() error {
	hookDebug("on-stop: invoked, agent=%q", cmd.Agent)
	in, err := decodeHookInput(os.Stdin)
	if err != nil {
		hookDebug("on-stop: decode stdin failed: %v", err)
		return nil
	}
	if cmd.Agent != "claude" {
		return nil
	}

	workdir := cmd.WorkDir
	if workdir == "" && in.CWD != "" {
		workdir = in.CWD
	}
	socketPath, err := resolveSocketForWorkDir(cmd.Socket, workdir)
	if err != nil {
		hookDebug("on-stop: socket resolve failed: %v", err)
		return nil
	}

	c, err := client.Connect(socketPath)
	if err != nil {
		hookDebug("on-stop: connect failed (engine down, turn ends normally): %v", err)
		return nil
	}
	defer c.Close()

	resp, err := c.Request(
		&protocol.AwaitReviewMsg{
			Type:      protocol.TypeAwaitReview,
			Wait:      true,
			MaxWaitMs: int(blockingHookMaxWait.Milliseconds()),
		},
		0,
	)
	if err != nil {
		hookDebug("on-stop: await-review request failed: %v", err)
		return nil
	}
	review, ok := resp.(*protocol.AwaitReviewResponse)
	if !ok {
		hookDebug("on-stop: unexpected response type %T", resp)
		return nil
	}
	hookDebug("on-stop: has_activity=%v action=%q feedback_len=%d", review.HasActivity, review.Action, len(review.Feedback))

	if !review.HasActivity {
		return nil
	}
	if review.Action == "request_changes" {
		return emitClaudeStopBlock(os.Stdout, review.Feedback)
	}
	// HasActivity=true but approved (or no explicit action) — turn ends normally.
	return nil
}

// emitClaudeStopBlock writes a Claude Code Stop-hook response that blocks
// the stop with the reviewer's feedback injected as the reason. Claude
// sees the reason and continues the conversation to address it.
func emitClaudeStopBlock(w io.Writer, reason string) error {
	if reason == "" {
		reason = "Reviewer requested changes."
	}
	payload := map[string]any{
		"decision": "block",
		"reason":   reason,
	}
	buf, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	hookDebug("on-stop: emitting block decision=%s", string(buf))
	_, err = fmt.Fprintln(w, string(buf))
	return err
}

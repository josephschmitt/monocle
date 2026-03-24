# o_(◉) monocle

**Review your AI agent's code as it writes it.** Leave comments on diffs, submit structured feedback, and watch the agent fix things in real time — all from your terminal.

You can run monocle either side-by-side with your agent:
![image](https://github.com/user-attachments/assets/c9c2b7ed-ce1b-417e-9fcb-b5086c6e94d4)

Or full-screen to get maximum space for your review:
![image](https://github.com/user-attachments/assets/91daaa6f-c838-46df-a699-aaedae62b240)

Monocle is a TUI that runs alongside [Claude Code](https://claude.com/claude-code). It connects via an [MCP channel](https://code.claude.com/docs/en/channels-reference) that pushes your review feedback directly into the agent's context. No copy-pasting, no window switching, no waiting.

## Why

Without something like Monocle, reviewing agent-written code means rubber-stamping diffs you didn't read, copy-pasting feedback into a chat window, or just hoping the agent got it right. There's no way to say "fix these three issues and show me again."

Monocle gives you a proper review loop without slowing the agent down. It doesn't gate each file change behind an approval — your agent keeps working while you review at your own pace. When you're ready, leave line-level comments and submit. The agent receives your feedback immediately and starts addressing it. You see the updated diffs, review again, and iterate — like PR reviews, but in real time.

## Requirements

- [Claude Code](https://claude.com/claude-code) v2.1.80+ (channels require claude.ai login, not API keys)
- A JavaScript runtime for the MCP channel: [Bun](https://bun.sh), [Deno](https://deno.com), or [Node.js](https://nodejs.org) (auto-detected in that order)
- A terminal with 256-color or true color support
- A [Nerd Font](https://www.nerdfonts.com/) for file icons (optional but recommended)

## Installation

### Homebrew (macOS/Linux)

```bash
brew install josephschmitt/tap/monocle
```

<details>
<summary>Other installation methods</summary>

#### Go Install

```bash
go install github.com/josephschmitt/monocle/cmd/monocle@latest
```

#### Pre-built Binaries

Download from [GitHub Releases](https://github.com/josephschmitt/monocle/releases):

**macOS:**
```bash
# Apple Silicon
# x-release-please-start-version
curl -Lo monocle.tar.gz https://github.com/josephschmitt/monocle/releases/download/v0.21.0/monocle_darwin_arm64.tar.gz
# x-release-please-end
tar xzf monocle.tar.gz
sudo mv monocle /usr/local/bin/

# Intel
# x-release-please-start-version
curl -Lo monocle.tar.gz https://github.com/josephschmitt/monocle/releases/download/v0.21.0/monocle_darwin_amd64.tar.gz
# x-release-please-end
tar xzf monocle.tar.gz
sudo mv monocle /usr/local/bin/
```

**Linux:**
```bash
# x86_64
# x-release-please-start-version
curl -Lo monocle.tar.gz https://github.com/josephschmitt/monocle/releases/download/v0.21.0/monocle_linux_amd64.tar.gz
# x-release-please-end
tar xzf monocle.tar.gz
sudo mv monocle /usr/local/bin/

# ARM64
# x-release-please-start-version
curl -Lo monocle.tar.gz https://github.com/josephschmitt/monocle/releases/download/v0.21.0/monocle_linux_arm64.tar.gz
# x-release-please-end
tar xzf monocle.tar.gz
sudo mv monocle /usr/local/bin/
```

#### From Source

```bash
git clone https://github.com/josephschmitt/monocle.git
cd monocle
devbox run -- make build
# Binaries are in bin/
```

</details>

## Quick Start

### 1. Install the plugin

In Claude Code, add Monocle as a plugin marketplace and install:

```
/plugin marketplace add josephschmitt/monocle
/plugin install monocle@monocle
```

### 2. Start reviewing

In one terminal, start Claude Code with the channel enabled (the flag is required during the [channels research preview](https://code.claude.com/docs/en/channels)):

```bash
claude --dangerously-load-development-channels plugin:monocle@monocle
```

In another, start Monocle:
```bash
monocle
```

Claude Code gets tools for checking review status, retrieving feedback, submitting plans, and more — and starts receiving your review feedback as push notifications.

> **Note:** The `--dangerously-load-development-channels` flag is required during the [channels research preview](https://code.claude.com/docs/en/channels-reference).

### 3. The review loop

Navigate with `j`/`k`, add comments with `c`, and use `v` for visual (multi-line) selections. Press `?` to see all keybindings, or see the full [Keybindings](#keybindings) reference.

**Submit** (`S`): Your review is formatted and pushed to Claude Code via the MCP channel. If there are no comments, it's treated as an approval. Toggle the "Copy to clipboard" checkbox with `Shift+Tab` in the submit modal to also copy the formatted review when submitting.

**Yank** (`Ctrl+y`): In the submit modal, copies the formatted review to your system clipboard without submitting, then closes the modal.

**Pause** (`P`): Claude Code receives a notification to stop and wait. It calls `get_feedback` with `wait=true` and blocks until you submit your review. This is for when you want to review before the agent moves on.

### 4. Plan review

Monocle isn't limited to reviewing file changes. Claude Code can submit **plans, architecture decisions, and other content** directly to Monocle for review using the `submit_plan` tool. These show up alongside your file diffs in the sidebar, and you can leave line-level comments on them the same way.

This means you can review the agent's *thinking* before it writes code — not just the output. Ask Claude Code to submit its plan first, review it, leave feedback, and only then let it start implementing.

When Claude Code enters [plan mode](https://docs.anthropic.com/en/docs/claude-code/plan-mode), Monocle provides the `submit_plan_and_wait` tool which submits the plan to your TUI **and blocks** until you respond with feedback. If you approve, the agent starts implementing. If you request changes, the agent updates the plan and submits again. See [Plan mode setup](#plan-mode-setup) for configuration details.

## Features

- **MCP channel integration** — Push-based feedback delivery to Claude Code, no polling or copy-pasting
- **Pause flow** — Ask Claude Code to stop and wait while you review, then release it when ready
- **Live diff viewer** — Unified and split (side-by-side) views with syntax highlighting and intra-line diffs
- **Structured comments** — Tag feedback as issues, suggestions, notes, or praise with line-level or file-level precision
- **Visual selection** — Select line ranges for comments with vim-style visual mode
- **Plan review** — Claude Code can submit plans for your review before writing code, with markdown rendering
- **Plan mode gating** — `submit_plan_and_wait` blocks the agent until you approve the plan before implementation begins
- **Markdown rendering** — Plans and changed `.md` files render with styled headings, bold, italic, lists, and code blocks
- **Horizontal scrolling & line wrapping** — Navigate wide diffs with `h`/`l` or toggle wrapping with `w`
- **Responsive layout** — Automatically stacks panes vertically in narrow terminals
- **Ref picker** — Change the base ref on the fly to compare against any branch or commit
- **Comment resolution** — Mark individual comments as resolved (`x`); resolved comments are excluded from submitted reviews
- **Submission history** — View past review submissions with `:history`
- **Mouse support** — Click to focus panes, scroll with the wheel, click files to select, drag to make visual selections, and interact with modal controls
- **Configurable keybindings** — Override any navigation or action key via config
- **Feedback queue** — Submit reviews while the agent is working; delivered when Claude Code next checks
- **Connection indicator** — See at a glance whether Claude Code is connected, with manual socket override for troubleshooting
- **Session persistence** — Reviews survive restarts via SQLite

## Plan mode setup

The channel's built-in instructions tell Claude Code about `submit_plan_and_wait`, but for reliable plan mode behavior you should also add instructions to your project's `CLAUDE.md`. Add the following to your project's `CLAUDE.md` (create one at the root of your repo if you don't have one):

````markdown
## Monocle Integration

When the Monocle MCP channel is connected:
- Use the `submit_plan` MCP tool to send plans or content for the reviewer to see
- Use the plan filename as the `id` parameter so updates replace the previous version

**Plan mode (important):** When in plan mode, use `submit_plan_and_wait` instead of `submit_plan`. This tool submits the plan AND blocks until the reviewer responds with feedback. If the reviewer approves, proceed to call ExitPlanMode. If they request changes, update the plan and call `submit_plan_and_wait` again. Only call ExitPlanMode after the reviewer has approved.
````

> **Why CLAUDE.md?** Claude Code reads your project's `CLAUDE.md` at the start of every conversation. While the MCP channel provides its own instructions, having the plan mode workflow in `CLAUDE.md` ensures the agent follows it consistently — especially at the critical moment of deciding whether to exit plan mode or submit for review first.

## Keybindings

| Key | Action |
|-----|--------|
| `j`/`k` | Move up/down |
| `J`/`K` | Scroll diff up/down (any pane) |
| `Ctrl+d`/`u` | Scroll diff half page (any pane) |
| `g`/`G` | Top/bottom |
| `h`/`l` | Scroll diff left/right |
| `H`/`L` | Scroll diff left/right (any pane) |
| `0` | Scroll to column 0 (any pane) |
| `^` | Scroll to first non-space (any pane) |
| `$` | Scroll to line end (any pane) |
| `[`/`]` | Previous/next file (any pane) |
| `{`/`}` | Previous/next sidebar section (any pane) |
| `Enter` | Focus diff pane / toggle dir |
| `Tab` | Switch pane focus |
| `\` | Toggle sidebar visibility |
| `1`/`2` | Jump to pane |
| `w` | Toggle line wrapping (any pane) |
| `f` | Toggle flat/tree view |
| `z`/`e` | Collapse/expand all (tree) |
| `b` | Change base ref |
| `c` | Add comment at cursor (edit if on a comment) |
| `C` | Add file-level comment |
| `v` | Visual select (multi-line comments) |
| `x` | Toggle comment resolved (on a comment line) |
| `d` | Delete comment (on a comment line) |
| `r` | Toggle file reviewed |
| `t` | Cycle diff style (unified/split/file) (any pane) |
| `T` | Cycle layout (auto/side-by-side/stacked) |
| `R` | Force reload files |
| `S` / `:submit` | Submit review |
| `Ctrl+y` | Copy review to clipboard |
| `P` / `:pause` | Pause Claude Code (wait for your review) |
| `D` / `:dismiss-outdated` | Dismiss outdated comments |
| `:discard` | Discard all pending comments |
| `:history` | View past review submissions |
| `I` | Connection info (socket path, subscriber count) |
| `?` | Show all keybindings |

## CLI

```
monocle [--socket PATH]        Start a review session
monocle register [--global]    Register MCP channel for Claude Code
monocle unregister [--global]  Remove MCP channel registration
monocle --version              Print version
```

### Manual Socket Override

If auto-pairing fails (e.g., Claude Code's working directory differs from Monocle's), you can manually specify the socket path on either side:

- **Monocle:** `monocle --socket /tmp/monocle-abc123.sock`
- **Channel (env var):** Set `MONOCLE_SOCKET` in your `.mcp.json`:
  ```json
  {
    "mcpServers": {
      "monocle": {
        "env": { "MONOCLE_SOCKET": "/tmp/monocle-abc123.sock" }
      }
    }
  }
  ```

Press `I` in the TUI to see the current socket path and connection status.

## Configuration

Monocle loads settings from JSON config files:

1. **Global:** `~/.config/monocle/config.json` (or `$XDG_CONFIG_HOME/monocle/config.json`)
2. **Project:** `.monocle/config.json` in the working directory (overrides global)

```json
{
  "layout": "auto",
  "diff_style": "unified",
  "sidebar_style": "flat",
  "wrap": false,
  "tab_size": 4,
  "context_lines": 3,
  "ignore_patterns": [],
  "keybindings": {},
  "mouse": true,
  "clear_after_submit": "ask",
  "plan_review_mode": false,
  "review_format": {
    "include_snippets": true,
    "max_snippet_lines": 10,
    "include_summary": true
  }
}
```

| Setting | Values | Default | Description |
|---------|--------|---------|-------------|
| `layout` | `"auto"`, `"side-by-side"`, `"stacked"` | `"auto"` | Pane arrangement (`auto` switches based on terminal width) |
| `diff_style` | `"unified"`, `"split"`, `"file"` | `"unified"` | Diff display mode (`file` shows raw content) |
| `sidebar_style` | `"flat"`, `"tree"` | `"flat"` | File list display mode |
| `wrap` | `true`, `false` | `false` | Word-wrap long lines in diffs |
| `tab_size` | integer | `4` | Spaces per tab character |
| `context_lines` | integer | `3` | Unchanged lines shown around diff hunks |
| `ignore_patterns` | string array | `[]` | Glob patterns for files to exclude |
| `mouse` | `true`, `false` | `true` | Enable mouse interactions (click, scroll, drag) |
| `clear_after_submit` | `"ask"`, `"always"`, `"never"` | `"ask"` | Whether to clear comments after submitting a review |
| `plan_review_mode` | `true`, `false` | `false` | Auto-hide sidebar and enable wrap when reviewing plans |
| `keybindings` | object | `{}` | Custom key overrides (see below) |
| `review_format.include_snippets` | `true`, `false` | `true` | Include code snippets in formatted reviews |
| `review_format.max_snippet_lines` | integer | `10` | Truncate snippets longer than this |
| `review_format.include_summary` | `true`, `false` | `true` | Include comment count summary in formatted reviews |

Toggle keybindings (`T`, `t`, `w`, `f`) change settings for the current session only. Edit the config file to persist your preferences.

### Custom Keybindings

Override any action key by mapping the action name to a new key string:

```json
{
  "keybindings": {
    "quit": "Q",
    "submit": "ctrl+s",
    "scroll_down": "ctrl+j"
  }
}
```

Available action names: `up`, `down`, `top`, `bottom`, `half_up`, `half_down`, `prev_file`, `next_file`, `select`, `focus_swap`, `toggle_sidebar`, `scroll_down`, `scroll_up`, `scroll_left`, `scroll_right`, `scroll_home`, `scroll_first_char`, `scroll_end`, `wrap`, `toggle_diff`, `tree_mode`, `collapse_all`, `expand_all`, `prev_section`, `next_section`, `comment`, `file_comment`, `visual`, `reviewed`, `submit`, `pause`, `dismiss_outdated`, `base_ref`, `cycle_layout`, `refresh`, `help`, `quit`, `command_mode`.

The help overlay (`?`) dynamically reflects your custom bindings. Modal keys (Enter, Esc, Tab in overlays) are not configurable.

## How it works

```
┌─────────────┐                ┌───────────────┐              ┌──────────┐
│ Claude Code │<--stdio/MCP--->│  channel.ts   │<---socket--->│ monocle  │
│             │                │ (MCP server)  │              │  (TUI)   │
└─────────────┘                └───────────────┘              └──────────┘
```

1. You leave line-level comments on diffs — issues, suggestions, notes, praise
2. You press `S` to submit your review
3. Monocle formats the review and pushes it through the MCP channel as a notification
4. Claude Code receives the feedback immediately and starts addressing your comments
5. You see the updated diffs in real time, review again, and iterate

The key difference from other approaches: **Claude Code doesn't have to stop and ask for feedback.** The channel pushes your review into its context the moment you submit. And if you want Claude Code to pause and wait for you to finish reviewing, just press `P` — it receives a pause notification and blocks until your review is ready.

## License

MIT

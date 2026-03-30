# o_(◉) monocle

**Review your AI agent's code as it writes it.** Leave comments on diffs, submit structured feedback, and watch the agent fix things in real time — all from your terminal.

![IMG_0689](https://github.com/user-attachments/assets/92a877a8-ac89-4561-bd05-9bb916820943)


[More screenshots →](screenshots/)

Monocle is a TUI that runs alongside your AI coding agent. You review diffs in real time as the agent writes code, leave line-level comments — issues, suggestions, notes — and submit a structured review in one batch. The agent receives your feedback and starts fixing things immediately, just like a PR review but live.

Monocle connects to your agent via [MCP](https://modelcontextprotocol.io/). With [Claude Code](https://claude.com/claude-code) and [MCP channels](https://code.claude.com/docs/en/channels-reference), feedback is pushed directly into the agent's context the moment you submit. Any other MCP-compatible agent — [OpenCode](https://opencode.ai), [Codex CLI](https://github.com/openai/codex), [Gemini CLI](https://github.com/google/gemini-cli) — retrieves feedback via the `get_feedback` tool.

## Why

Without something like Monocle, reviewing agent-written code means rubber-stamping diffs you didn't read, copy-pasting feedback into a chat window, or just hoping the agent got it right. There's no way to say "fix these three issues and show me again."

Monocle gives you a proper review loop without slowing the agent down. It doesn't gate each file change behind an approval — your agent keeps working while you review at your own pace. When you're ready, leave line-level comments and submit. The agent receives your feedback immediately and starts addressing it. You see the updated diffs, review again, and iterate — like PR reviews, but in real time.

## Requirements

- An MCP-compatible coding agent: [Claude Code](https://claude.com/claude-code), [OpenCode](https://opencode.ai), [Codex CLI](https://github.com/openai/codex), [Gemini CLI](https://github.com/google/gemini-cli), or any agent that supports MCP tool servers
- A JavaScript runtime for the MCP server: [Bun](https://bun.sh), [Deno](https://deno.com), or [Node.js](https://nodejs.org) (auto-detected in that order)
- For push notifications: [Claude Code](https://claude.com/claude-code) v2.1.80+ with [MCP channel support](https://code.claude.com/docs/en/channels-reference) (requires claude.ai login, not API keys)
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
curl -Lo monocle.tar.gz https://github.com/josephschmitt/monocle/releases/download/v0.35.0/monocle_darwin_arm64.tar.gz
# x-release-please-end
tar xzf monocle.tar.gz
sudo mv monocle /usr/local/bin/

# Intel
# x-release-please-start-version
curl -Lo monocle.tar.gz https://github.com/josephschmitt/monocle/releases/download/v0.35.0/monocle_darwin_amd64.tar.gz
# x-release-please-end
tar xzf monocle.tar.gz
sudo mv monocle /usr/local/bin/
```

**Linux:**
```bash
# x86_64
# x-release-please-start-version
curl -Lo monocle.tar.gz https://github.com/josephschmitt/monocle/releases/download/v0.35.0/monocle_linux_amd64.tar.gz
# x-release-please-end
tar xzf monocle.tar.gz
sudo mv monocle /usr/local/bin/

# ARM64
# x-release-please-start-version
curl -Lo monocle.tar.gz https://github.com/josephschmitt/monocle/releases/download/v0.35.0/monocle_linux_arm64.tar.gz
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

### Claude Code (recommended — push notifications)

Claude Code supports [MCP channels](https://code.claude.com/docs/en/channels-reference), which let Monocle push review feedback directly into the agent's context as a notification. This means the agent learns about your review the moment you submit — no manual step needed. It's the most seamless integration and the way Monocle was designed to be used.

#### 1. Install the plugin

In Claude Code, add Monocle as a plugin marketplace and install:

```
/plugin marketplace add josephschmitt/monocle
/plugin install monocle@monocle
```

#### 2. Start reviewing

In one terminal, start Claude Code with the channel enabled (the flag is required during the [channels research preview](https://code.claude.com/docs/en/channels)):

```bash
claude --dangerously-load-development-channels plugin:monocle@monocle
```

In another, start Monocle:
```bash
monocle
```

Claude Code gets tools for checking review status, retrieving feedback, submitting content for review, and more — and starts receiving your review feedback as push notifications.

> **Note:** The `--dangerously-load-development-channels` flag is required during the [channels research preview](https://code.claude.com/docs/en/channels-reference).

> **Tip:** If you start or restart Monocle while Claude Code is already running, the MCP channel may need to reconnect. Type `/mcp` in Claude Code and select Monocle to reconnect.

### Codex CLI

Install the Monocle plugin from the Codex CLI plugin browser:

```
/plugins
```

Search for "monocle" and install. Then start Monocle in a separate terminal:

```bash
monocle
```

### Gemini CLI

Install the Monocle extension:

```bash
gemini extensions install josephschmitt/monocle
```

Then start Monocle in a separate terminal:

```bash
monocle
```

### OpenCode / other agents (manual registration)

For agents without a plugin system, use `monocle register` to write MCP config and slash commands directly:

```bash
monocle register opencode   # or: codex, gemini, all
```

You can also run `monocle register` with no argument to get an interactive picker. Each agent writes to its own config location:

| Agent | Config file | Slash commands |
|-------|-------------|----------------|
| Claude Code | `.mcp.json` | `plugins/claude/commands/` |
| OpenCode | `opencode.json` | `.opencode/commands/` |
| Codex CLI | `.codex/config.toml` | `.codex-plugin/skills/` |
| Gemini CLI | `.gemini/settings.json` | `.gemini/commands/` |

> **Note:** `monocle register` is also available as a fallback for Claude Code, Codex CLI, and Gemini CLI if you prefer direct config over the plugin/extension system.

Start your agent and Monocle in separate terminals. When you submit a review, Monocle queues it for delivery. The agent retrieves it by calling the `get_feedback` tool, either via the `/get-feedback` slash command or by calling the tool directly.

### The review loop

Navigate with `j`/`k`, add comments with `c`, and use `v` for visual (multi-line) selections. Press `?` to see all keybindings, or see the full [Keybindings](#keybindings) reference.

**Submit** (`S`): Your review is formatted and queued for delivery. With Claude Code channels, a push notification prompts the agent to retrieve it immediately. With other agents, the review waits in the queue until the agent calls `get_feedback` or you trigger it with the `/get-feedback` command. Multiple reviews can accumulate in the queue — the agent receives them all combined when it pulls. If there are no comments, the review is treated as an approval. Toggle the "Copy to clipboard" checkbox with `Shift+Tab` in the submit modal to also copy the formatted review when submitting.

**External editor** (`Ctrl+g`): In the comment or submit modal, opens the current text in your `$VISUAL` or `$EDITOR` (falls back to `vi`). Edit in your preferred editor, save and quit, and the text is brought back into the modal.

**Yank** (`Ctrl+y`): In the submit modal, copies the formatted review to your system clipboard without submitting, then closes the modal.

**Pause** (`P`): The agent receives a push notification to stop and wait. It calls `get_feedback` with `wait=true` and blocks until you submit your review. This is for when you want to review before the agent moves on. Pause requires MCP channel support (currently Claude Code only).

### Plan review and focus mode

Monocle isn't limited to reviewing file changes. Your agent can submit **plans, architecture decisions, summaries, and other content** directly to Monocle for review using the `submit_for_review` tool. These show up alongside your file diffs in the sidebar, and you can leave line-level comments on them the same way. You can also trigger this yourself with the `/review-plan` or `/review-plan-wait` slash commands — useful when you want to send the agent's plan to Monocle without waiting for the agent to do it on its own.

This means you can review the agent's *thinking* before it writes code — not just the output. Ask the agent to submit its content first, review it, leave feedback, and only then let it proceed.

The `submit_for_review_and_wait` tool submits content to your TUI **and blocks** until you respond with feedback. If you approve, the agent continues. If you request changes, the agent updates and submits again — iterating across as many rounds as it takes until you're satisfied.

> **Note:** Monocle's tools are available to your agent but the agent decides when to use them on its own. If you want the agent to automatically submit plans for review, add instructions to your agent's project configuration. See the [plugin README](plugin/README.md#automatic-content-review) for a suggested prompt.

## Features

- **Works with any MCP agent** — Claude Code, OpenCode, Codex CLI, Gemini CLI, or any agent that supports MCP tool servers
- **Push notifications** — With Claude Code channels, feedback is pushed directly into the agent's context the moment you submit
- **Pull-based feedback** — Agents without channel support retrieve feedback via the `get_feedback` tool; multiple reviews queue up and are delivered together
- **Pause flow** — Ask your agent to stop and wait while you review, then release it when ready (requires MCP channel support)
- **Live diff viewer** — Unified and split (side-by-side) views with syntax highlighting and intra-line diffs
- **Structured comments** — Tag feedback as issues, suggestions, notes, or praise with line-level or file-level precision
- **Suggested edits** — Press `s` to propose exact code changes with GitHub-style `suggestion` blocks
- **Visual selection** — Select line ranges for comments with vim-style visual mode
- **Content review + focus mode** — Your agent can submit plans, summaries, and other content for your review, with markdown rendering and distraction-free focus mode
- **Review gating** — `submit_for_review_and_wait` blocks the agent until you approve the submitted content before it proceeds
- **Markdown rendering** — Plans and changed `.md` files render with styled headings, bold, italic, lists, and code blocks
- **Horizontal scrolling & line wrapping** — Navigate wide diffs with `h`/`l` or toggle wrapping with `w`
- **Responsive layout** — Automatically stacks panes vertically in narrow terminals
- **Ref picker** — Change the base ref on the fly to compare against any branch or commit
- **Comment resolution** — Mark individual comments as resolved (`x`); resolved comments are excluded from submitted reviews
- **Submission history** — View past review submissions with `:history`
- **Mouse support** — Click to focus panes, scroll with the wheel, click files to select, drag to make visual selections, and interact with modal controls
- **External editor** — Open comment or submit text in `$VISUAL`/`$EDITOR` with `Ctrl+g` for full editing power
- **Configurable keybindings** — Override any navigation or action key via config
- **Feedback queue** — Submit reviews while the agent is working; delivered when the agent next calls `get_feedback`
- **Connection indicator** — See at a glance whether your agent is connected, with manual socket override for troubleshooting
- **Review tracking** — Mark files as reviewed with `r` (auto-advances to next), filter sidebar to show only unreviewed or reviewed files with `/`, and all reviewed states reset on submit
- **Session persistence** — Reviews survive restarts via SQLite

## MCP Tools

Monocle exposes the following tools to your agent via its MCP server:

| Tool | Description |
|------|-------------|
| `review_status` | Check the current review status, including whether feedback is pending or a pause has been requested |
| `get_feedback` | Retrieve queued review feedback. When `wait` is true, blocks until feedback is available |
| `submit_for_review` | Submit content for review in Monocle. Accepts inline content or a file path. Returns immediately after submission |
| `submit_for_review_and_wait` | Submit content for review in Monocle and block until the reviewer responds. An empty response means no comments were left |
| `add_files` | Add files or directories to the review session in Monocle. Accepts absolute paths |

## Slash Commands

Monocle ships slash commands for agents that support them. These are thin wrappers around the MCP tools above.

| Command | Available for | Description |
|---------|---------------|-------------|
| `/get-feedback` | Claude Code, Codex CLI, OpenCode, Gemini CLI | Retrieve pending review feedback |
| `/review-plan` | Claude Code, Codex CLI, OpenCode, Gemini CLI | Find the active plan file and submit it for review via `submit_for_review` |
| `/review-plan-wait` | Claude Code, Codex CLI, OpenCode, Gemini CLI | Find the active plan file and submit it for review via `submit_for_review_and_wait`, then iterate on feedback |

Slash commands live in agent-specific directories (`plugins/claude/commands/` for Claude Code, `plugins/codex/skills/` for Codex CLI, `.opencode/commands/` for OpenCode, `plugins/gemini/commands/` for Gemini CLI).

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
| `s` | Suggest edit at cursor (pre-fills suggestion block) |
| `C` | Add file-level comment |
| `v` | Visual select (multi-line comments) |
| `x` | Toggle comment resolved (on a comment line) |
| `d` | Delete comment (on a comment line) |
| `r` | Toggle file reviewed (auto-advances to next unreviewed) |
| `/` | Cycle sidebar filter (all → unreviewed → reviewed) |
| `t` | Cycle diff style (unified/split/file) (any pane) |
| `T` | Cycle layout (auto/side-by-side/stacked) |
| `R` | Force reload files |
| `S` / `:submit` | Submit review |
| `Ctrl+g` | Open external editor (comment/submit modal) |
| `Ctrl+y` | Copy review to clipboard |
| `P` / `:pause` | Pause the agent (wait for your review) |
| `D` / `:clear` | Clear review (all comments, plans, reviewed states) |
| `F` | Toggle focus mode (hide sidebar, enable wrap) |
| `:mark-all-reviewed` | Mark all files as reviewed |
| `:mark-all-unreviewed` | Mark all files as unreviewed |
| `:discard` | Discard all pending comments |
| `:history` | View past review submissions |
| `I` | Connection info (socket path, subscriber count) |
| `?` | Show all keybindings |

### Comment editor

The comment editor supports standard emacs-style shortcuts:

| Key | Action |
|-----|--------|
| `←`/`→` or `Ctrl+B`/`Ctrl+F` | Move cursor left/right |
| `↑`/`↓` or `Ctrl+P`/`Ctrl+N` | Move cursor up/down (multiline) |
| `Home`/`Ctrl+A` | First non-whitespace, then start of line |
| `End`/`Ctrl+E` | End of line |
| `Ctrl+D` or `Delete` | Delete character at cursor |
| `Ctrl+K` | Kill to end of line |
| `Ctrl+U` | Kill to start of line |
| `Ctrl+W` or `Alt+Backspace` | Delete word before cursor |
| `Alt+D` | Delete word after cursor |
| `Alt+←` or `Alt+B` | Move cursor back one word |
| `Alt+→` or `Alt+F` | Move cursor forward one word |
| `Shift+Enter` or `Alt+Enter` | Insert newline |
| `Ctrl+G` | Open in external editor (`$VISUAL`/`$EDITOR`) |
| `Tab` | Cycle comment type |
| `Enter` | Save comment |
| `Esc` | Cancel |

## CLI

```
monocle [--socket PATH]              Start a review session
monocle register [agent] [--global]  Register Monocle for an agent
monocle unregister [agent] [--global] Remove Monocle registration
monocle --version                    Print version
```

The `agent` argument is one of `claude`, `opencode`, `codex`, `gemini`, or `all`. If omitted, an interactive picker lets you select which agents to register. The `--global` flag writes to the user-level config directory instead of the project.

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
  "min_diff_width": 80,
  "auto_focus_mode": false,
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
| `min_diff_width` | integer | `80` | Minimum character width for the diff viewer in side-by-side layout |
| `mouse` | `true`, `false` | `true` | Enable mouse interactions (click, scroll, drag) |
| `auto_focus_mode` | `true`, `false` | `false` | Auto-enter focus mode (hide sidebar, enable wrap) when reviewing plans |
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

Available action names: `up`, `down`, `top`, `bottom`, `half_up`, `half_down`, `prev_file`, `next_file`, `select`, `focus_swap`, `toggle_sidebar`, `scroll_down`, `scroll_up`, `scroll_left`, `scroll_right`, `scroll_home`, `scroll_first_char`, `scroll_end`, `wrap`, `toggle_diff`, `tree_mode`, `collapse_all`, `expand_all`, `prev_section`, `next_section`, `comment`, `file_comment`, `suggest`, `visual`, `reviewed`, `submit`, `pause`, `dismiss_outdated`, `base_ref`, `cycle_layout`, `refresh`, `help`, `quit`, `command_mode`.

The help overlay (`?`) dynamically reflects your custom bindings. Modal keys (Enter, Esc, Tab in overlays) are not configurable.

## How it works

```
┌─────────────┐                ┌───────────────┐              ┌──────────┐
│   Agent     │<--stdio/MCP--->│  channel.ts   │<---socket--->│ monocle  │
│             │                │ (MCP server)  │              │  (TUI)   │
└─────────────┘                └───────────────┘              └──────────┘
```

1. You leave line-level comments on diffs — issues, suggestions, notes, praise
2. You press `S` to submit your review
3. Monocle queues the review for delivery
4. The agent picks up the feedback — automatically via push notification (Claude Code with channels) or when you trigger `/get-feedback` — and starts addressing your comments
5. You see the updated diffs in real time, review again, and iterate

Feedback is always queued for reliability. How the agent learns about it depends on the integration:

- **Claude Code with channels:** A push notification is sent through the MCP channel with a summary of the review (e.g., "Your reviewer requested changes — 2 issues, 1 suggestion"). The agent calls `get_feedback` to retrieve the full review. If the push fails silently (channels not enabled), the review stays in the queue.
- **Any MCP agent:** The agent calls the `get_feedback` tool directly, either via a `/get-feedback` slash command or on its own. Multiple reviews accumulate in the queue and are delivered together.

If you want the agent to pause and wait for you to finish reviewing, press `P` — the agent receives a pause notification and blocks until your review is ready.

## License

MIT

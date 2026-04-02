# o_(◉) monocle

**Review your AI agent's code as it writes it.** Leave comments on diffs, submit structured feedback, and watch the agent fix things in real time — all from your terminal.

![IMG_0689](https://github.com/user-attachments/assets/92a877a8-ac89-4561-bd05-9bb916820943)


[More screenshots →](screenshots/)

Monocle is a TUI that runs alongside your AI coding agent. You review diffs in real time as the agent writes code, leave line-level comments — issues, suggestions, notes — and submit a structured review in one batch. The agent receives your feedback and starts fixing things immediately, just like a PR review but live.

Monocle connects to your agent using Unix sockets. But with [Claude Code](https://claude.com/claude-code) and [MCP channels](https://code.claude.com/docs/en/channels-reference), it's able to push feedback directly into the agent's context the moment you submit. Other agents — [OpenCode](https://opencode.ai), [Codex CLI](https://github.com/openai/codex), [Gemini CLI](https://github.com/google/gemini-cli) — can still use skills or bash tools to retrieve feedback from monocle after it's been submitted and queued. MCP channels just make the process smoother.

## Why

Without something like Monocle, reviewing agent-written code means rubber-stamping diffs you didn't read, copy-pasting feedback into a chat window, or just hoping the agent got it right. There's no way to say "fix these three issues and show me again."

Monocle gives you a proper review loop without slowing the agent down. It doesn't gate each file change behind an approval — your agent keeps working while you review at your own pace. When you're ready, leave line-level comments and submit. The agent receives your feedback immediately and starts addressing it. You see the updated diffs, review again, and iterate — like PR reviews, but in real time.

## Requirements

- A coding agent: [Claude Code](https://claude.com/claude-code), [OpenCode](https://opencode.ai), [Codex CLI](https://github.com/openai/codex), [Gemini CLI](https://github.com/google/gemini-cli), or any agent that supports [agent skills](https://agentskills.io)
- For push notifications (Claude Code only): a JavaScript runtime ([Bun](https://bun.sh), [Deno](https://deno.com), or [Node.js](https://nodejs.org)) and [MCP channel support](https://code.claude.com/docs/en/channels-reference)
- A terminal with 256-color or true color support
- A [Nerd Font](https://www.nerdfonts.com/) for file icons (optional but recommended)

## Installation

### Homebrew (macOS/Linux)

```bash
brew install josephschmitt/tap/monocle
```

<details>
<summary>Other installation methods</summary>

#### Pre-built Binaries

Download from [GitHub Releases](https://github.com/josephschmitt/monocle/releases):

**macOS:**
```bash
# Apple Silicon
# x-release-please-start-version
curl -Lo monocle.tar.gz https://github.com/josephschmitt/monocle/releases/download/v0.41.1/monocle_darwin_arm64.tar.gz
# x-release-please-end
tar xzf monocle.tar.gz
sudo mv monocle /usr/local/bin/

# Intel
# x-release-please-start-version
curl -Lo monocle.tar.gz https://github.com/josephschmitt/monocle/releases/download/v0.41.1/monocle_darwin_amd64.tar.gz
# x-release-please-end
tar xzf monocle.tar.gz
sudo mv monocle /usr/local/bin/
```

**Linux:**
```bash
# x86_64
# x-release-please-start-version
curl -Lo monocle.tar.gz https://github.com/josephschmitt/monocle/releases/download/v0.41.1/monocle_linux_amd64.tar.gz
# x-release-please-end
tar xzf monocle.tar.gz
sudo mv monocle /usr/local/bin/

# ARM64
# x-release-please-start-version
curl -Lo monocle.tar.gz https://github.com/josephschmitt/monocle/releases/download/v0.41.1/monocle_linux_arm64.tar.gz
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

### 1. Register Monocle with your agent

```bash
monocle register          # interactive picker
monocle register claude   # or: opencode, codex, gemini, all
```

This installs skills and, for Claude Code, the MCP channel config for push notifications. Use `--global` to write to the user-level config directory instead of the project.

| Agent       | Skills              | MCP config  |
|-------------|---------------------|-------------|
| Claude Code | `.claude/skills/`   | `.mcp.json` |
| OpenCode    | `.opencode/skills/` | -           |
| Codex CLI   | `.codex/skills/`    | -           |
| Gemini CLI  | `.gemini/skills/`   | -           |

<details>
<summary>Alternative: use your agent's plugin/extension system</summary>

#### Claude Code

```
/plugin marketplace add josephschmitt/monocle
/plugin install monocle@monocle
```

#### Codex CLI

```
/plugins
```

Search for "monocle" and install.

#### Gemini CLI

```bash
gemini extensions install josephschmitt/monocle
```

</details>

### 2. Start reviewing

Start your agent and Monocle in separate terminals:

```bash
monocle
```

Your agent gets [skills](#skills) for checking review status, retrieving feedback, and submitting content for review. When you submit a review, Monocle queues it for delivery. The agent retrieves it via the `/get-feedback` skill or on its own.

#### Push notifications (Claude Code only)

Claude Code supports [MCP channels](https://code.claude.com/docs/en/channels-reference), which automate the `/get-feedback` step. When you submit a review, a push notification prompts the agent to retrieve your feedback immediately instead of waiting for the next poll. To enable channels, start Claude Code with the channel flag:

```bash
claude --dangerously-load-development-channels plugin:monocle@monocle
```

> **Note:** The `--dangerously-load-development-channels` flag is required during the [channels research preview](https://code.claude.com/docs/en/channels-reference).

> **Tip:** If you start or restart Monocle while Claude Code is already running, the MCP channel may need to reconnect. Type `/mcp` in Claude Code and select Monocle to reconnect.

### The review loop

Navigate with `j`/`k`, add comments with `c`, and use `v` for visual (multi-line) selections. Press `?` to see all keybindings, or see the full [Keybindings](#keybindings) reference.

**Submit** (`S`): Your review is formatted and queued for delivery. With Claude Code channels, a push notification prompts the agent to retrieve it immediately. With other agents, the review waits in the queue until the agent runs `/get-feedback` or calls `monocle review get-feedback`. Multiple reviews can accumulate in the queue — the agent receives them all combined when it pulls. If there are no comments, the review is treated as an approval. Toggle the "Copy to clipboard" checkbox with `Shift+Tab` in the submit modal to also copy the formatted review when submitting.

**External editor** (`Ctrl+g`): In the comment or submit modal, opens the current text in your `$VISUAL` or `$EDITOR` (falls back to `vi`). Edit in your preferred editor, save and quit, and the text is brought back into the modal.

**Yank** (`Ctrl+y`): In the submit modal, copies the formatted review to your system clipboard without submitting, then closes the modal.

**Pause** (`P`): The agent receives a push notification to stop and wait. It runs `monocle review get-feedback --wait` and blocks until you submit your review. This is for when you want to review before the agent moves on. Pause requires MCP channel support (currently Claude Code only).

### Plan review and focus mode

Monocle isn't limited to reviewing file changes. Your agent can submit **plans, architecture decisions, summaries, and other content** directly to Monocle for review using `monocle review send-artifact`. These show up alongside your file diffs in the sidebar, and you can leave line-level comments on them the same way. You can also trigger this yourself with the `/review-plan` or `/review-plan-wait` skills — useful when you want to send the agent's plan to Monocle without waiting for the agent to do it on its own.

This means you can review the agent's *thinking* before it writes code — not just the output. Ask the agent to submit its content first, review it, leave feedback, and only then let it proceed.

The `/review-plan-wait` skill submits content to your TUI **and blocks** until you respond with feedback. If you approve, the agent continues. If you request changes, the agent updates and submits again — iterating across as many rounds as it takes until you're satisfied.

> **Note:** Monocle's skills are available to your agent but the agent decides when to use them on its own. If you want the agent to automatically submit plans for review, add instructions to your agent's project configuration. See [Automatic content review](#automatic-content-review) below for a suggested prompt.

## Features

- **Works with any coding agent** — Claude Code, OpenCode, Codex CLI, Gemini CLI, or any agent that supports [skills](https://agentskills.io) or can call CLI commands
- **Push notifications** — With Claude Code channels, feedback is pushed directly into the agent's context the moment you submit
- **Pull-based feedback** — Agents without channel support retrieve feedback via the `/get-feedback` skill or `monocle review get-feedback`; multiple reviews queue up and are delivered together
- **Plan & architecture review** — Your agent can submit plans, architecture decisions, and other content for review with markdown rendering. When iterating, Monocle shows diffs between plan versions so you can see exactly what changed. Use focus mode (`F`) for distraction-free reading
- **Review gating** — `/review-plan-wait` blocks the agent until you approve the submitted content before it proceeds
- **Pause flow** — Ask your agent to stop and wait while you review, then release it when ready (requires MCP channel support)
- **Live diff viewer** — Unified and split (side-by-side) views with syntax highlighting and intra-line diffs
- **Structured comments** — Tag feedback as issues, suggestions, notes, or praise with line-level or file-level precision
- **Suggested edits** — Press `s` to propose exact code changes with GitHub-style `suggestion` blocks
- **Visual selection** — Select line ranges for comments with vim-style visual mode
- **Markdown rendering** — Plans and changed `.md` files render with styled headings, bold, italic, lists, and code blocks
- **Horizontal scrolling & line wrapping** — Navigate wide diffs with `h`/`l` or toggle wrapping with `w`
- **Responsive layout** — Automatically stacks panes vertically in narrow terminals
- **Ref picker** — Change the base ref on the fly to compare against any branch or commit
- **Version history** — Browse all versions of a plan or artifact and diff any version against the latest
- **Comment resolution** — Mark individual comments as resolved (`x`); resolved comments are excluded from submitted reviews
- **Submission history** — View past review submissions with `:history`
- **Mouse support** — Click to focus panes, scroll with the wheel, click files to select, drag to make visual selections, and interact with modal controls
- **External editor** — Open comment or submit text in `$VISUAL`/`$EDITOR` with `Ctrl+g` for full editing power
- **Configurable keybindings** — Override any navigation or action key via config
- **Feedback queue** — Submit reviews while the agent is working; delivered when the agent next runs `/get-feedback`
- **Connection indicator** — See at a glance whether your agent is connected, with manual socket override for troubleshooting
- **Review tracking** — Mark files as reviewed with `r` (auto-advances to next), filter sidebar to show only unreviewed or reviewed files with `/`, and all reviewed states reset on submit
- **Session persistence** — Reviews survive restarts via SQLite

## Skills

Monocle ships skills for all supported agents. These are installed by `monocle register` into agent-specific skill directories.

| Skill               | Available for                                | Description                                                                                                    |
|---------------------|----------------------------------------------|----------------------------------------------------------------------------------------------------------------|
| `/get-feedback`     | Claude Code, Codex CLI, OpenCode, Gemini CLI | Retrieve pending review feedback                                                                               |
| `/review-plan`      | Claude Code, Codex CLI, OpenCode, Gemini CLI | Find the active plan file and submit it for review via `monocle review send-artifact`                          |
| `/review-plan-wait` | Claude Code, Codex CLI, OpenCode, Gemini CLI | Find the active plan file, submit for review, and iterate on feedback until approved                           |

Skills live in agent-specific directories (`.claude/skills/` for Claude Code, `.codex/skills/` for Codex CLI, `.opencode/skills/` for OpenCode, `.gemini/skills/` for Gemini CLI).

## Keybindings

| Key                    | Action                                                    |
|------------------------|-----------------------------------------------------------|
| `j`/`k`                | Move up/down                                              |
| `J`/`K`                | Scroll diff up/down (any pane)                            |
| `Ctrl+d`/`u`           | Scroll diff half page (any pane)                          |
| `g`/`G`                | Top/bottom                                                |
| `h`/`l`                | Scroll diff left/right                                    |
| `H`/`L`                | Scroll diff left/right (any pane)                         |
| `0`                    | Scroll to column 0 (any pane)                             |
| `^`                    | Scroll to first non-space (any pane)                      |
| `$`                    | Scroll to line end (any pane)                             |
| `[`/`]`                | Previous/next file (any pane)                             |
| `{`/`}`                | Previous/next sidebar section (any pane)                  |
| `Enter`                | Focus diff pane / toggle dir                              |
| `Tab`                  | Switch pane focus                                         |
| `\`                    | Toggle sidebar visibility                                 |
| `1`/`2`                | Jump to pane                                              |
| `w`                    | Toggle line wrapping (any pane)                           |
| `f`                    | Toggle flat/tree view                                     |
| `z`/`e`                | Collapse/expand all (tree)                                |
| `b`                    | Change base ref                                           |
| `B`                    | Base artifact version to diff against                             |
| `c`                    | Add comment at cursor (edit if on a comment)              |
| `s`                    | Suggest edit at cursor (pre-fills suggestion block)       |
| `C`                    | Add file-level comment                                    |
| `v`                    | Visual select (multi-line comments)                       |
| `x`                    | Toggle comment resolved (on a comment line)               |
| `d`                    | Delete comment (on a comment line)                        |
| `r`                    | Toggle file reviewed (auto-advances to next unreviewed)   |
| `/`                    | Cycle sidebar filter (all -> unreviewed -> reviewed)      |
| `t`                    | Cycle diff style (unified/split/file) (any pane)          |
| `T`                    | Cycle layout (auto/side-by-side/stacked)                  |
| `R`                    | Force reload files                                        |
| `S` / `:submit`        | Submit review                                             |
| `Ctrl+g`               | Open external editor (comment/submit modal)               |
| `Ctrl+y`               | Copy review to clipboard                                  |
| `P` / `:pause`         | Pause the agent (wait for your review)                    |
| `D` / `:clear`         | Clear review (all comments, plans, reviewed states)       |
| `F`                    | Toggle focus mode (hide sidebar, enable wrap)             |
| `:mark-all-reviewed`   | Mark all files as reviewed                                |
| `:mark-all-unreviewed` | Mark all files as unreviewed                              |
| `:discard`             | Discard all pending comments                              |
| `:history`             | View past review submissions                              |
| `:base-artifact-version` | Base artifact version to diff against                           |
| `:base-ref`              | Base ref to diff against (same as `b`)                  |
| `I`                    | Connection info (socket path, subscriber count)           |
| `?`                    | Show all keybindings                                      |

### Comment editor

The comment editor supports standard emacs-style shortcuts:

| Key                                    | Action                                          |
|----------------------------------------|-------------------------------------------------|
| `<-`/`->` or `Ctrl+B`/`Ctrl+F`         | Move cursor left/right                          |
| `up`/`down` or `Ctrl+P`/`Ctrl+N`       | Move cursor up/down (multiline)                 |
| `Home`/`Ctrl+A`                        | First non-whitespace, then start of line        |
| `End`/`Ctrl+E`                         | End of line                                     |
| `Ctrl+D` or `Delete`                   | Delete character at cursor                      |
| `Ctrl+K`                               | Kill to end of line                             |
| `Ctrl+U`                               | Kill to start of line                           |
| `Ctrl+W` or `Alt+Backspace`            | Delete word before cursor                       |
| `Alt+D`                                | Delete word after cursor                        |
| `Alt+<-` or `Alt+B`                    | Move cursor back one word                       |
| `Alt+->` or `Alt+F`                    | Move cursor forward one word                    |
| `Shift+Enter` or `Alt+Enter`           | Insert newline                                  |
| `Ctrl+G`                               | Open in external editor (`$VISUAL`/`$EDITOR`)   |
| `Tab`                                  | Cycle comment type                              |
| `Enter`                                | Save comment                                    |
| `Esc`                                  | Cancel                                          |

## CLI

```
monocle [--socket PATH]              Start a review session
monocle register [agent] [--global]  Register Monocle for an agent
monocle unregister [agent] [--global] Remove Monocle registration
monocle --version                    Print version
```

The `agent` argument is one of `claude`, `opencode`, `codex`, `gemini`, or `all`. If omitted, an interactive picker lets you select which agents to register. The `--global` flag writes to the user-level config directory instead of the project.

### Agent-Facing Commands

These commands are used by agents (via skills) to interact with a running Monocle session:

```
monocle review status [--json]                          Check review status
monocle review get-feedback [--wait] [--json]            Retrieve review feedback
monocle review send-artifact --title T [--file F] [--id ID] [--type EXT] [--wait] [--json]
                                                         Send content for review
monocle review add-files <paths...> [--json]             Add files to review session
```

- `--wait` blocks until the reviewer responds (used by `/review-plan-wait` skill)
- `--json` outputs structured JSON for programmatic use
- `send-artifact` reads from `--file` or stdin

### Manual Socket Override

If auto-pairing fails (e.g., the agent's working directory differs from Monocle's), you can manually specify the socket path:

- **Monocle:** `monocle --socket /tmp/monocle-abc123.sock`
- **Agent commands:** `MONOCLE_SOCKET=/tmp/monocle-abc123.sock monocle review status`
- **MCP channel (Claude):** Set `MONOCLE_SOCKET` in `.mcp.json` env

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
  "comment_expand": true,
  "comment_expand_delay": 2000,
  "review_format": {
    "include_snippets": true,
    "max_snippet_lines": 10,
    "include_summary": true
  }
}
```

| Setting                              | Values                                     | Default      | Description                                                              |
|--------------------------------------|--------------------------------------------|--------------|--------------------------------------------------------------------------|
| `layout`                             | `"auto"`, `"side-by-side"`, `"stacked"`    | `"auto"`     | Pane arrangement (`auto` switches based on terminal width)               |
| `diff_style`                         | `"unified"`, `"split"`, `"file"`           | `"unified"`  | Diff display mode (`file` shows raw content)                             |
| `sidebar_style`                      | `"flat"`, `"tree"`                         | `"flat"`     | File list display mode                                                   |
| `wrap`                               | `true`, `false`                            | `false`      | Word-wrap long lines in diffs                                            |
| `tab_size`                           | integer                                    | `4`          | Spaces per tab character                                                 |
| `context_lines`                      | integer                                    | `3`          | Unchanged lines shown around diff hunks                                  |
| `ignore_patterns`                    | string array                               | `[]`         | Glob patterns for files to exclude                                       |
| `min_diff_width`                     | integer                                    | `80`         | Minimum character width for the diff viewer in side-by-side layout       |
| `mouse`                              | `true`, `false`                            | `true`       | Enable mouse interactions (click, scroll, drag)                          |
| `auto_focus_mode`                    | `true`, `false`                            | `false`      | Auto-enter focus mode (hide sidebar, enable wrap) when reviewing plans   |
| `comment_expand`                     | `true`, `false`                            | `true`       | Auto-expand comments on hover                                            |
| `comment_expand_delay`               | integer (ms)                               | `2000`       | Delay before auto-expanding a selected comment (0 = instant)             |
| `keybindings`                        | object                                     | `{}`         | Custom key overrides (see below)                                         |
| `review_format.include_snippets`     | `true`, `false`                            | `true`       | Include code snippets in formatted reviews                               |
| `review_format.max_snippet_lines`    | integer                                    | `10`         | Truncate snippets longer than this                                       |
| `review_format.include_summary`      | `true`, `false`                            | `true`       | Include comment count summary in formatted reviews                       |

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

## Automatic content review

By default, Monocle's skills are available to your agent but the agent decides when to use them on its own. If you want the agent to automatically submit plans or other content for review, add instructions to your agent's project configuration (e.g. `CLAUDE.md`, `AGENTS.md`, etc.):

````markdown
## Monocle Integration

When Monocle is running:
- Use the `/review-plan` skill to send content (plans, decisions, summaries) for the reviewer to see
- Use the content's filename as the identifier so updates replace the previous version
- In plan mode, use `/review-plan-wait` instead — it blocks until the reviewer responds. If they request changes, update and resubmit until approved.
````

You can also use the `/review-plan` and `/review-plan-wait` skills manually at any time.

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

- **Claude Code with channels:** A push notification is sent through the MCP channel with a summary of the review (e.g., "Your reviewer requested changes — 2 issues, 1 suggestion"). The agent runs `/get-feedback` to retrieve the full review. If the push fails silently (channels not enabled), the review stays in the queue.
- **Any agent:** The agent calls `monocle review get-feedback` via a skill or CLI command, either via a `/get-feedback` slash command or on its own. Multiple reviews accumulate in the queue and are delivered together.

If you want the agent to pause and wait for you to finish reviewing, press `P` — the agent receives a pause notification and blocks until your review is ready.

## License

MIT

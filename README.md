# o_(◉) monocle

**Review your AI agent's code as it writes it.** Leave comments on diffs, submit structured feedback, and watch the agent fix things in real time — all from your terminal.

![image](https://github.com/user-attachments/assets/30580911-35ee-4d82-9eb0-5dde13663741)

Monocle is a TUI that runs alongside [Claude Code](https://claude.com/claude-code). It connects via an [MCP channel](https://code.claude.com/docs/en/channels-reference) that pushes your review feedback directly into the agent's context. No copy-pasting, no window switching, no waiting.

## Why

Without something like Monocle, reviewing agent-written code means rubber-stamping diffs you didn't read, copy-pasting feedback into a chat window, or just hoping the agent got it right. There's no way to say "fix these three issues and show me again."

Monocle gives you a proper review loop without slowing the agent down. It doesn't gate each file change behind an approval — your agent keeps working while you review at your own pace. When you're ready, leave line-level comments and submit. The agent receives your feedback immediately and starts addressing it. You see the updated diffs, review again, and iterate — like PR reviews, but in real time.

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

### Not just diffs

Monocle isn't limited to reviewing file changes. Claude Code can submit **plans, architecture decisions, and other content** directly to Monocle for review using the `submit_plan` tool. These show up alongside your file diffs in the sidebar, and you can leave line-level comments on them the same way.

This means you can review the agent's *thinking* before it writes code — not just the output. Ask Claude Code to submit its plan first, review it, leave feedback, and only then let it start implementing.

## Features

- **MCP channel integration** — Push-based feedback delivery to Claude Code, no polling or copy-pasting
- **Pause flow** — Ask Claude Code to stop and wait while you review, then release it when ready
- **Live diff viewer** — Unified and split (side-by-side) views with syntax highlighting and intra-line diffs
- **Structured comments** — Tag feedback as issues, suggestions, notes, or praise with line-level or file-level precision
- **Visual selection** — Select line ranges for comments with vim-style visual mode
- **Plan review** — Claude Code can submit plans for your review before writing code
- **Horizontal scrolling & line wrapping** — Navigate wide diffs with `h`/`l` or toggle wrapping with `w`
- **Responsive layout** — Automatically stacks panes vertically in narrow terminals
- **Ref picker** — Change the base ref on the fly to compare against any branch or commit
- **Feedback queue** — Submit reviews while the agent is working; delivered when Claude Code next checks
- **Connection indicator** — See at a glance whether Claude Code is connected, with manual socket override for troubleshooting
- **Session persistence** — Reviews survive restarts via SQLite

## Installation

### Homebrew (macOS/Linux)

```bash
brew install josephschmitt/tap/monocle
```

### Go Install

```bash
go install github.com/josephschmitt/monocle/cmd/monocle@latest
```

### Pre-built Binaries

Download from [GitHub Releases](https://github.com/josephschmitt/monocle/releases):

**macOS:**
```bash
# Apple Silicon
# x-release-please-start-version
curl -Lo monocle.tar.gz https://github.com/josephschmitt/monocle/releases/download/v0.10.1/monocle_0.8.0_darwin_arm64.tar.gz
# x-release-please-end
tar xzf monocle.tar.gz
sudo mv monocle /usr/local/bin/

# Intel
# x-release-please-start-version
curl -Lo monocle.tar.gz https://github.com/josephschmitt/monocle/releases/download/v0.10.1/monocle_0.8.0_darwin_amd64.tar.gz
# x-release-please-end
tar xzf monocle.tar.gz
sudo mv monocle /usr/local/bin/
```

**Linux:**
```bash
# x86_64
# x-release-please-start-version
curl -Lo monocle.tar.gz https://github.com/josephschmitt/monocle/releases/download/v0.10.1/monocle_0.8.0_linux_amd64.tar.gz
# x-release-please-end
tar xzf monocle.tar.gz
sudo mv monocle /usr/local/bin/

# ARM64
# x-release-please-start-version
curl -Lo monocle.tar.gz https://github.com/josephschmitt/monocle/releases/download/v0.10.1/monocle_0.8.0_linux_arm64.tar.gz
# x-release-please-end
tar xzf monocle.tar.gz
sudo mv monocle /usr/local/bin/
```

### From Source

```bash
git clone https://github.com/josephschmitt/monocle.git
cd monocle
devbox run -- make build
# Binaries are in bin/
```

## Quick Start

### 1. Install the MCP channel

```bash
monocle install
```

This registers Monocle as an MCP server in your project's `.mcp.json`. Use `--global` to install to `~/.mcp.json` instead (applies to all projects).

### 2. Start reviewing

In one terminal, start Monocle:
```bash
monocle
```

In another, start Claude Code with the development channels flag (required during the [channels research preview](https://code.claude.com/docs/en/channels)):

```bash
claude --dangerously-load-development-channels server:monocle
```

This tells Claude Code to load the monocle MCP server as a channel. Claude Code gets three new tools (`review_status`, `get_feedback`, `submit_plan`) and starts receiving your review feedback as push notifications.

> **Note:** The `--dangerously-load-development-channels` flag is only needed during the channels research preview. Once channels are generally available, `monocle install` will be all you need.

### 3. The review loop

| Key | Action |
|-----|--------|
| `j`/`k` | Move up/down |
| `J`/`K` | Scroll diff up/down (any pane) |
| `Ctrl+d`/`u` | Scroll diff half page (any pane) |
| `g`/`G` | Top/bottom |
| `h`/`l` | Scroll diff left/right |
| `H`/`L` | Scroll diff left/right (any pane) |
| `[`/`]` | Previous/next file (any pane) |
| `Enter` | Focus diff pane / toggle dir |
| `Tab` | Switch pane focus |
| `1`/`2` | Jump to pane |
| `w` | Toggle line wrapping |
| `f` | Toggle flat/tree view |
| `z`/`e` | Collapse/expand all (tree) |
| `b` | Change base ref |
| `c` | Add comment at cursor |
| `C` | Add file-level comment |
| `v` | Visual select (multi-line comments) |
| `r` | Toggle file reviewed |
| `t` | Toggle unified/split diff |
| `T` | Cycle layout (auto/side-by-side/stacked) |
| `S` | Submit review |
| `P` | Pause Claude Code (wait for your review) |
| `D` | Dismiss outdated comments |
| `I` | Connection info (socket path, subscriber count) |
| `?` | Show all keybindings |

**Submit** (`S`): Your review is formatted and pushed to Claude Code via the MCP channel. If there are no comments, it's treated as an approval.

**Pause** (`P`): Claude Code receives a notification to stop and wait. It calls `get_feedback` with `wait=true` and blocks until you submit your review. This is for when you want to review before the agent moves on.

## CLI

```
monocle [--socket PATH]     Start a review session
monocle install [--global]  Install MCP channel for Claude Code
monocle uninstall [--global] Remove MCP channel
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
  "theme": "default",
  "ignore_patterns": [],
  "keybindings": {},
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
| `diff_style` | `"unified"`, `"split"` | `"unified"` | Diff display mode |
| `sidebar_style` | `"flat"`, `"tree"` | `"flat"` | File list display mode |
| `wrap` | `true`, `false` | `false` | Word-wrap long lines in diffs |
| `tab_size` | integer | `4` | Spaces per tab character |
| `context_lines` | integer | `3` | Unchanged lines shown around diff hunks |
| `theme` | `"default"` | `"default"` | Color theme |
| `ignore_patterns` | string array | `[]` | Glob patterns for files to exclude |
| `keybindings` | object | `{}` | Custom key overrides |
| `review_format` | object | see above | Controls how review feedback is formatted |

Toggle keybindings (`T`, `t`, `w`, `f`) change settings for the current session only. Edit the config file to persist your preferences.

## Requirements

- [Claude Code](https://claude.com/claude-code) v2.1.80+ (channels require claude.ai login, not API keys)
- A JavaScript runtime for the MCP channel: [Bun](https://bun.sh), [Deno](https://deno.com), or [Node.js](https://nodejs.org) (auto-detected in that order)
- A terminal with 256-color or true color support
- A [Nerd Font](https://www.nerdfonts.com/) for file icons (optional but recommended)

## License

MIT

# Monocle

A terminal-based code review tool that creates a real feedback loop between you and Claude Code.

Monocle uses an [MCP channel](https://code.claude.com/docs/en/channels-reference) to push your review feedback directly into Claude Code's context — no copy-pasting, no polling, no window switching. You review diffs and leave structured comments. Claude Code receives them as they happen and addresses the issues. It's PR reviews, but between you and your agent, in real time.

## The Problem

Without a tool like Monocle, reviewing agent-written code is painful:

- **Rubber-stamping** — You approve without reading because there's no good way to review in-progress work
- **Copy-pasting** — You read a diff somewhere, then manually paste feedback into the agent's prompt
- **Context switching** — You bounce between terminal, editor, git diff, and the agent window
- **No iteration** — There's no way to say "fix these three issues and show me again"

## How Monocle Solves It

Monocle runs alongside Claude Code as a dedicated review TUI. Under the hood, it connects via an **MCP channel** — a push-based communication layer that lets Monocle deliver feedback directly into Claude Code's context without the agent needing to poll or check for updates.

Here's the flow:

```
┌─────────────┐    stdio/MCP    ┌───────────────┐    socket     ┌──────────┐
│ Claude Code │ ◂──────────────▸│  channel.ts   │ ◂───────────▸│ monocle  │
│             │                 │  (MCP server)  │               │  (TUI)   │
└─────────────┘                 └───────────────┘               └──────────┘
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
- **Live diff viewer** — Unified and side-by-side views that update as the agent makes changes
- **Structured comments** — Tag feedback as issues, suggestions, notes, or praise with line-level precision
- **Visual selection** — Select line ranges for comments with vim-style visual mode
- **Plan review** — Claude Code can submit plans for your review before writing code
- **Feedback queue** — Submit reviews while the agent is working; delivered when Claude Code next checks
- **Session persistence** — Reviews survive restarts via SQLite

## Installation

### Homebrew (macOS/Linux)

```bash
brew install josephschmitt/tap/monocle
```

### Go Install

```bash
go install github.com/anthropics/monocle/cmd/monocle@latest
```

### Pre-built Binaries

Download from [GitHub Releases](https://github.com/josephschmitt/monocle/releases):

**macOS:**
```bash
# Apple Silicon
# x-release-please-start-version
curl -Lo monocle.tar.gz https://github.com/josephschmitt/monocle/releases/download/v0.5.0/monocle_0.2.0_darwin_arm64.tar.gz
# x-release-please-end
tar xzf monocle.tar.gz
sudo mv monocle /usr/local/bin/

# Intel
# x-release-please-start-version
curl -Lo monocle.tar.gz https://github.com/josephschmitt/monocle/releases/download/v0.5.0/monocle_0.2.0_darwin_amd64.tar.gz
# x-release-please-end
tar xzf monocle.tar.gz
sudo mv monocle /usr/local/bin/
```

**Linux:**
```bash
# x86_64
# x-release-please-start-version
curl -Lo monocle.tar.gz https://github.com/josephschmitt/monocle/releases/download/v0.5.0/monocle_0.2.0_linux_amd64.tar.gz
# x-release-please-end
tar xzf monocle.tar.gz
sudo mv monocle /usr/local/bin/

# ARM64
# x-release-please-start-version
curl -Lo monocle.tar.gz https://github.com/josephschmitt/monocle/releases/download/v0.5.0/monocle_0.2.0_linux_arm64.tar.gz
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
| `j`/`k` | Navigate files |
| `Enter` | Focus diff pane |
| `c` | Add comment at cursor |
| `v` | Visual select (multi-line comments) |
| `S` | Submit review |
| `P` | Pause Claude Code (wait for your review) |
| `?` | Show all keybindings |

**Submit** (`S`): Your review is formatted and pushed to Claude Code via the MCP channel. If there are no comments, it's treated as an approval.

**Pause** (`P`): Claude Code receives a notification to stop and wait. It calls `get_feedback` with `wait=true` and blocks until you submit your review. This is for when you want to review before the agent moves on.

## CLI

```
monocle                     Start a review session
monocle install [--global]  Install MCP channel for Claude Code
monocle uninstall [--global] Remove MCP channel
```

## Requirements

- [Claude Code](https://claude.com/claude-code) v2.1.80+ (channels require claude.ai login, not API keys)
- [Bun](https://bun.sh) (runtime for the MCP channel server)
- A terminal with 256-color or true color support
- A [Nerd Font](https://www.nerdfonts.com/) for file icons (optional but recommended)

## License

MIT

# Monocle

A terminal-based code review tool purpose-built for the agentic coding era.

As AI agents take over the act of writing code, the developer's role shifts from author to **reviewer and director**. But today's tools don't serve that role well. You either rubber-stamp everything the agent produces or interrupt it on every file change. There's no equivalent of "leave a thoughtful PR review and let the author address it."

Monocle fills that gap. It runs alongside Claude Code and provides a structured, bidirectional code review workflow — like GitHub PR reviews, but between you and your agent, right in the terminal.

The agent writes code. You review diffs, leave structured feedback (issues, suggestions, notes). Monocle delivers that feedback directly back into Claude Code via an MCP channel. The agent addresses your comments and re-presents its changes. No copy-pasting, no window switching, no breaking flow.

## Why Monocle?

Without Monocle, the agentic review loop is broken:
- **Rubber-stamping** — You approve without reading because there's no good way to review
- **Copy-pasting** — You manually paste feedback from a diff viewer into the agent's prompt
- **Context switching** — You bounce between editor, terminal, git diff, and agent window
- **No iteration** — There's no way to say "fix these issues and show me again"

Monocle integrates with Claude Code via an MCP channel to create a real review loop. When you're ready, you see the diffs, leave line-level comments, and submit a formatted review — Claude Code receives it as a channel notification, addresses the issues, and re-presents its changes for another round.

## Features

- **Live diff viewer** — Unified and side-by-side diff views with syntax-aware coloring
- **Structured comments** — Tag feedback as issues, suggestions, notes, or praise with line-level precision
- **Visual selection** — Select line ranges for comments with vim-style visual mode
- **Automatic refresh** — File list and diffs update live as the agent makes changes
- **Feedback queue** — Submit reviews while the agent is working; they're delivered when Claude Code next checks
- **Session persistence** — Reviews survive restarts via SQLite
- **MCP channel** — Push-based integration with Claude Code
- **Nerd Font icons** — File type icons with true color in the sidebar

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
curl -Lo monocle.tar.gz https://github.com/josephschmitt/monocle/releases/download/v0.1.0/monocle_0.1.0_darwin_arm64.tar.gz
# x-release-please-end
tar xzf monocle.tar.gz
sudo mv monocle /usr/local/bin/

# Intel
# x-release-please-start-version
curl -Lo monocle.tar.gz https://github.com/josephschmitt/monocle/releases/download/v0.1.0/monocle_0.1.0_darwin_amd64.tar.gz
# x-release-please-end
tar xzf monocle.tar.gz
sudo mv monocle /usr/local/bin/
```

**Linux:**
```bash
# x86_64
# x-release-please-start-version
curl -Lo monocle.tar.gz https://github.com/josephschmitt/monocle/releases/download/v0.1.0/monocle_0.1.0_linux_amd64.tar.gz
# x-release-please-end
tar xzf monocle.tar.gz
sudo mv monocle /usr/local/bin/

# ARM64
# x-release-please-start-version
curl -Lo monocle.tar.gz https://github.com/josephschmitt/monocle/releases/download/v0.1.0/monocle_0.1.0_linux_arm64.tar.gz
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

This writes `channel.ts` to `~/.config/monocle/` and adds the monocle MCP server to `.mcp.json` in your project.

### 2. Start a review session

In one terminal pane, start monocle:
```bash
monocle
```

In another pane, run Claude Code as usual. The MCP channel connects Monocle and Claude Code automatically.

### 3. Review and comment

| Key | Action |
|-----|--------|
| `j`/`k` | Navigate up/down |
| `Enter` | Focus diff pane |
| `Tab` | Switch pane focus |
| `1`/`2` | Jump to sidebar/diff |
| `c` | Add comment at cursor |
| `v` | Visual select mode |
| `r` | Mark file as reviewed |
| `t` | Toggle unified/split diff |
| `b` | Change base ref |
| `S` | Submit review |
| `P` | Pause (ask Claude Code to wait) |
| `D` | Dismiss outdated comments |
| `?` | Show all keybindings |
| `q` | Quit |

### 4. Submit feedback

Press `S` to submit your review. If Claude Code is waiting (paused), the review is delivered immediately. If it's still working, the review is queued and delivered when Claude Code next checks for feedback.

Press `P` to request a pause — Claude Code receives a notification and waits for your review before continuing.

## How It Works

```
┌─────────────┐    stdio/MCP    ┌───────────────┐    socket     ┌──────────┐
│ Claude Code │ ◂──────────────▸│  channel.ts   │ ◂───────────▸│ monocle  │
│             │                 │  (MCP server)  │               │  (TUI)   │
└─────────────┘                 └───────────────┘               └──────────┘
```

1. `monocle install` writes a `channel.ts` MCP server and registers it in `.mcp.json`
2. Claude Code spawns `channel.ts` as a subprocess and communicates via stdio
3. `channel.ts` connects to the running Monocle TUI via a Unix domain socket
4. As the agent makes changes, Monocle updates its diff view in real time
5. When you submit a review, `channel.ts` pushes a notification to Claude Code
6. Claude Code sees the feedback and addresses the issues

## CLI Commands

```
monocle                     Start a review session (default)
monocle install             Install MCP channel for Claude Code
monocle uninstall           Remove MCP channel
```

## Requirements

- A terminal with 256-color or true color support
- A [Nerd Font](https://www.nerdfonts.com/) for file icons (optional but recommended)
- [Bun](https://bun.sh) for the MCP channel server
- Go 1.23+ (for building from source)

## License

MIT

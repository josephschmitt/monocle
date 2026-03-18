# Monocle

A terminal-based code review companion for AI coding agents.

Monocle runs alongside Claude Code, Codex, Gemini CLI, or any AI coding agent. The agent writes code, you review diffs and leave structured feedback — like a GitHub PR review, but in your terminal — and Monocle delivers that feedback directly back into the agent's context. No more rubber-stamping, no more copy-pasting.

## Why Monocle?

AI coding agents are powerful, but the review loop is broken. You either:
- **Rubber-stamp** agent output without reading it
- **Copy-paste** feedback from a separate terminal into the agent's prompt
- **Lose context** switching between your editor, git diff, and the agent

Monocle fixes this by providing a structured review workflow that integrates directly with your agent via hooks. When the agent stops, you see the diffs, leave comments (issues, suggestions, notes), and submit a formatted review — all without leaving the terminal.

## Features

- **Live diff viewer** — Unified and side-by-side diff views with syntax-aware coloring
- **Structured comments** — Tag feedback as issues, suggestions, notes, or praise with line-level precision
- **Visual selection** — Select line ranges for comments with vim-style visual mode
- **Automatic refresh** — File list and diffs update live as the agent makes changes
- **Feedback queue** — Submit reviews while the agent is working; they're delivered when it next stops
- **Session persistence** — Reviews survive restarts via SQLite
- **Agent hooks** — Direct integration with Claude Code (more agents coming)
- **Nerd Font icons** — File type icons with true color in the sidebar

## Installation

### Homebrew (macOS/Linux)

```bash
brew install josephschmitt/tap/monocle
```

### Go Install

```bash
go install github.com/anthropics/monocle/cmd/monocle@latest
go install github.com/anthropics/monocle/cmd/monocle-hook@latest
```

### Pre-built Binaries

<!-- x-release-please-start-version -->

Download from [GitHub Releases](https://github.com/josephschmitt/monocle/releases):

**macOS:**
```bash
# Apple Silicon
curl -Lo monocle.tar.gz https://github.com/josephschmitt/monocle/releases/download/v0.1.0/monocle_0.1.0_darwin_arm64.tar.gz
tar xzf monocle.tar.gz
sudo mv monocle monocle-hook /usr/local/bin/

# Intel
curl -Lo monocle.tar.gz https://github.com/josephschmitt/monocle/releases/download/v0.1.0/monocle_0.1.0_darwin_amd64.tar.gz
tar xzf monocle.tar.gz
sudo mv monocle monocle-hook /usr/local/bin/
```

**Linux:**
```bash
# x86_64
curl -Lo monocle.tar.gz https://github.com/josephschmitt/monocle/releases/download/v0.1.0/monocle_0.1.0_linux_amd64.tar.gz
tar xzf monocle.tar.gz
sudo mv monocle monocle-hook /usr/local/bin/

# ARM64
curl -Lo monocle.tar.gz https://github.com/josephschmitt/monocle/releases/download/v0.1.0/monocle_0.1.0_linux_arm64.tar.gz
tar xzf monocle.tar.gz
sudo mv monocle monocle-hook /usr/local/bin/
```

<!-- x-release-please-end -->

### From Source

```bash
git clone https://github.com/josephschmitt/monocle.git
cd monocle
devbox run -- make build
# Binaries are in bin/
```

## Quick Start

### 1. Install hooks for your agent

```bash
monocle setup claude
```

This outputs the hook configuration to add to your Claude Code settings. Copy it into `.claude/settings.json` (project) or `~/.claude/settings.json` (global).

### 2. Start a review session

In one terminal pane, start monocle:
```bash
monocle
```

In another pane, run your agent as usual. Monocle picks up file changes via hooks automatically.

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
| `S` | Submit review |
| `A` | Approve (release agent) |
| `D` | Dismiss outdated comments |
| `?` | Show all keybindings |
| `q` | Quit |

### 4. Submit feedback

Press `S` to submit your review. If the agent is stopped, the review is delivered immediately. If it's still working, the review is queued and delivered when the agent next stops.

Press `A` to approve and release the agent without feedback.

## How It Works

```
┌─────────────┐     hooks      ┌──────────────┐    socket     ┌──────────┐
│  AI Agent   │ ──────────────▸│ monocle-hook │ ─────────────▸│ monocle  │
│ (Claude,etc)│                │   (shim)     │               │  (TUI)   │
└─────────────┘                └──────────────┘               └──────────┘
       ▲                                                           │
       │                     formatted review                      │
       └───────────────────────────────────────────────────────────┘
```

1. Your AI agent fires hooks as it works (file edits, stop events)
2. `monocle-hook` (a lightweight shim) forwards these to the running `monocle` instance via a Unix domain socket
3. Monocle updates its diff view in real time
4. When you submit a review, it's formatted as structured markdown and injected back into the agent's context
5. The agent sees your feedback and can address the issues

## CLI Commands

```
monocle                     Start a new review session (default)
monocle start --agent X     Start with a specific agent type
monocle resume <session-id> Resume a previous session
monocle sessions            List past sessions
monocle setup claude        Show hook configuration for Claude Code
```

## Requirements

- A terminal with 256-color or true color support
- A [Nerd Font](https://www.nerdfonts.com/) for file icons (optional but recommended)
- Go 1.23+ (for building from source)

## License

MIT

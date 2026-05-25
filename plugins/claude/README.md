# Monocle

Terminal-based code review companion for AI coding agents. Run Monocle alongside your agent — the agent writes code, you review diffs and leave structured feedback, and Monocle delivers your review back via skills and push notifications.

## Prerequisites

- [Monocle](https://github.com/josephschmitt/monocle) — install via Homebrew:
  ```
  brew install --cask josephschmitt/tap/monocle
  ```

## Setup

1. Add the marketplace:
   ```
   /plugin marketplace add josephschmitt/monocle
   ```

2. Install the plugin:
   ```
   /plugin install monocle@monocle
   ```

3. Start Claude Code with the channel enabled:
   ```
   claude --channels plugin:monocle@monocle
   ```

4. In a separate terminal, start Monocle in the same repo:
   ```
   monocle
   ```

Monocle's TUI will display diffs as Claude makes changes. Add comments, then submit your review — Claude receives the feedback automatically via push notification.

## Skills

| Skill | Description |
|-------|-------------|
| `/get-feedback` | Retrieve pending review feedback |
| `/review-plan` | Send a plan file to Monocle for review. Returns immediately |
| `/review-plan-wait` | Send a plan file to Monocle and block until the reviewer responds |

## MCP channel

The plugin includes an MCP channel server (`channel.ts`) that sends push notifications to Claude Code when review events occur. It does not expose any MCP tools — all operations use skills that run CLI commands under the hood.

| Event | Notification |
|-------|--------------|
| `feedback_submitted` | Prompts Claude to run `monocle review get-feedback` to retrieve the review |
| `pause_requested` | Prompts Claude to run `monocle review get-feedback --wait` to block until the reviewer submits |

## How it works

Monocle runs a TUI that watches your repo for changes. When Claude modifies files, Monocle shows you the diffs. You review, add inline comments, and submit. The MCP channel server connects to the Monocle engine via a Unix domain socket and pushes notifications to Claude when you submit feedback or request a pause. Claude then runs the appropriate CLI command to retrieve your review.

Claude can also submit content (plans, summaries, decisions) for your review — these appear in Monocle's TUI so you can provide early feedback.

## Pause flow

Press **P** in Monocle to request Claude pause. Claude receives a `pause_requested` notification and blocks until you submit your review — giving you time to catch up without the agent racing ahead.
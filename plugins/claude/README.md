# Monocle

Terminal-based code review companion for AI coding agents. Run Monocle alongside your agent — the agent writes code, you review diffs and leave structured feedback, and Monocle delivers your review back via MCP.

## Prerequisites

- [Monocle](https://github.com/josephschmitt/monocle) — install via Homebrew:
  ```
  brew install josephschmitt/tap/monocle
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

Monocle's TUI will display diffs as Claude makes changes. Add comments, then submit your review — Claude receives the feedback automatically.

## Tools

| Tool | Description |
|------|-------------|
| `review_status` | Check the current review status, including whether feedback is pending or a pause has been requested |
| `get_feedback` | Retrieve queued review feedback. When `wait` is true, blocks until feedback is available |
| `submit_for_review` | Submit content for review in Monocle. Accepts inline content or a file path. Returns immediately |
| `submit_for_review_and_wait` | Submit content for review in Monocle and block until the reviewer responds |
| `add_files` | Add files or directories to the review session in Monocle. Accepts absolute paths |

## How it works

Monocle runs a TUI that watches your repo for changes. When Claude modifies files, Monocle shows you the diffs. You review, add inline comments, and submit. The plugin registers Monocle's built-in MCP channel server, which connects to the engine via a Unix domain socket and pushes your feedback to Claude as channel notifications.

Claude can also submit content (plans, summaries, decisions) for your review — these appear in Monocle's TUI so you can provide early feedback.

## Pause flow

Press **P** in Monocle to request Claude pause. Claude receives a notification and blocks until you submit your review — giving you time to catch up without the agent racing ahead.

## Automatic content review

By default, Monocle's tools are available to your agent but the agent decides when to use them on its own. If you want the agent to automatically submit plans or other content for review, add instructions to your agent's project configuration (e.g. `CLAUDE.md`, `AGENTS.md`, etc.):

````markdown
## Monocle Integration

When Monocle's MCP tools are available:
- Use the `submit_for_review` MCP tool to send content (plans, decisions, summaries) for the reviewer to see
- Use the content's filename as the `id` parameter so updates replace the previous version
- In plan mode, use `submit_for_review_and_wait` instead — it blocks until the reviewer responds. If they request changes, update and call again until approved.
````

You can also use the `/review-plan` and `/review-plan-wait` slash commands to manually send content at any time.

See the [main README](https://github.com/josephschmitt/monocle#plan-review-and-focus-mode) for more on content review workflows.

# Monocle

Terminal-based code review companion for AI coding agents. Run Monocle alongside your agent — the agent writes code, you review diffs and leave structured feedback, and Monocle delivers your review back via skills.

## Prerequisites

- [Monocle](https://github.com/josephschmitt/monocle) — install via Homebrew:
  ```
  brew install --cask josephschmitt/tap/monocle
  ```

## Setup

1. Install the extension:
   ```bash
   gemini extensions install josephschmitt/monocle
   ```

2. Start Gemini CLI and Monocle in separate terminals:
   ```
   monocle
   ```

Monocle's TUI will display diffs as Gemini makes changes. Add comments, then submit your review — Gemini retrieves the feedback via the `/get-feedback` skill.

## Skills

| Skill | Description |
|-------|-------------|
| `/get-feedback` | Retrieve pending review feedback |
| `/review-plan` | Send a plan file to Monocle for review. Returns immediately |
| `/review-plan-wait` | Send a plan file to Monocle and block until the reviewer responds |

## How it works

Monocle runs a TUI that watches your repo for changes. When Gemini modifies files, Monocle shows you the diffs. You review, add inline comments, and submit. Gemini retrieves your feedback by running the `/get-feedback` skill.

Gemini can also submit content (plans, summaries, decisions) for your review using `/review-plan` — these appear in Monocle's TUI so you can provide early feedback.
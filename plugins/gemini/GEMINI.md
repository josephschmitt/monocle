# Monocle

Terminal-based code review companion for AI coding agents. Your reviewer is watching your code changes in real time using Monocle.

## How it works

Monocle exposes MCP tools for interacting with your reviewer:

- `review_status` — Check if the reviewer has pending feedback or has requested a pause
- `get_feedback` — Retrieve queued review feedback. Use `wait=true` to block until feedback is available
- `submit_for_review` — Submit content (plans, summaries, decisions) for the reviewer to see and comment on
- `submit_for_review_and_wait` — Submit content and block until the reviewer responds
- `add_files` — Add additional files for the reviewer to see

When the reviewer submits feedback, retrieve it with `get_feedback`. If the reviewer requests a pause, stop and call `get_feedback` with `wait=true` to block until they submit their review.

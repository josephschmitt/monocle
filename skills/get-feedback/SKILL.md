---
name: get-feedback
description: Retrieves pending review feedback from Monocle. Use when the reviewer has submitted feedback through the Monocle TUI, or after receiving a feedback_submitted notification.
---

# Get Review Feedback

Run `monocle review get-feedback` to retrieve pending review feedback.

## Handling the response

- If feedback is available, read it carefully and act on it — the feedback contains your reviewer's comments, issues, and suggestions about your code changes
- If no feedback is pending, inform the user that no review feedback is available yet

After receiving feedback, address the reviewer's comments in your code, then continue with your work.

# Get Review Feedback

Retrieve pending review feedback from your Monocle reviewer. Use this when your reviewer has submitted feedback through the Monocle TUI and you need to pick it up.

## Instructions

1. **Run `monocle review get-feedback`** using the Bash tool (non-blocking poll)

2. **Handle the response:**
   - If feedback is available, read it carefully and act on it — the feedback contains your reviewer's comments, issues, and suggestions about your code changes
   - If no feedback is pending, inform the user that no review feedback is available yet

3. **After receiving feedback**, address the reviewer's comments in your code, then continue with your work

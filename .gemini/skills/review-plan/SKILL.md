---
name: review-plan
description: Sends a plan file to Monocle for the reviewer to see. Returns immediately without waiting for feedback. Use when submitting a plan for review and you do not need to block for approval.
---

# Send Plan to Monocle

Submits a plan file to Monocle so the reviewer can see it. Does NOT wait for feedback — use `/review-plan-wait` to block until the reviewer responds.

## Steps

1. **Find the plan file** — if the user provided a path via `$ARGUMENTS`, use that. Otherwise, find the most recently modified plan file in the project.

2. **Read the plan file** to confirm it exists and get its filename.

3. **Run `monocle review send-artifact`** with:
   - `--title`: The first markdown heading from the plan, or the filename if no heading found
   - `--file`: Absolute path to the plan file
   - `--id`: The plan filename (e.g. `my-plan.md`) — ensures updates replace the previous version
   - `--type`: `md`

4. **Confirm** to the user that the plan was sent to Monocle.

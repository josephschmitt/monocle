---
name: review-plan-wait
description: Sends a plan file to Monocle and blocks until the reviewer responds with feedback. Use in plan mode or whenever reviewer approval is needed before proceeding.
---

# Send Plan to Monocle and Wait for Review

Submits a plan file to Monocle and blocks until the reviewer responds with feedback.

## Steps

1. **Find the plan file** — if the user provided a path via `$ARGUMENTS`, use that. Otherwise, find the most recently modified plan file in the project.

2. **Read the plan file** to confirm it exists and get its filename.

3. **Run `monocle review send-artifact --wait`** with:
   - `--title`: The first markdown heading from the plan, or the filename if no heading found
   - `--file`: Absolute path to the plan file
   - `--id`: The plan filename (e.g. `my-plan.md`) — ensures updates replace the previous version
   - `--type`: `md`
   - `--wait`: Blocks until the reviewer responds

4. **Handle the response:**
   - If the reviewer approved with no comments, inform the user and continue
   - If the reviewer provided feedback requesting changes, share the feedback with the user and act on it — update the plan, then run `monocle review send-artifact --wait` again
   - Keep iterating until the reviewer approves

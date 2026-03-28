# Send Plan to Monocle and Wait for Review

Submit a plan file to Monocle and block until the reviewer responds with feedback. Use this in plan mode or whenever you need reviewer approval before proceeding.

## Instructions

1. **Find the plan file:**
   - If the user provided a path via `$ARGUMENTS`, use that
   - Otherwise, look for the active plan file in `.claude/plans/` (most recently modified `.md` file)

2. **Read the plan file** to confirm it exists and get its filename

3. **Call the `submit_for_review_and_wait` MCP tool** with:
   - `title`: The first markdown heading from the plan, or the filename if no heading found
   - `file_path`: Absolute path to the plan file
   - `id`: The plan filename (e.g. `my-plan.md`) — this ensures updates replace the previous version
   - `content_type`: `"md"`

4. **Handle the response:**
   - If the reviewer approved with no comments, inform the user and continue
   - If the reviewer provided feedback requesting changes, share the feedback with the user and act on it — update the plan, then call `submit_for_review_and_wait` again to start another review round
   - Keep iterating until the reviewer approves

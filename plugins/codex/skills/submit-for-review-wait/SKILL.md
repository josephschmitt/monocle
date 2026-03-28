---
name: submit-for-review-wait
description: Send content to Monocle and wait for review feedback
---

When the user asks you to submit content and wait for review feedback, find the most recently modified plan file in the project (or use the path the user provided). Read it to get the filename and first heading.

Call the monocle `submit_for_review_and_wait` tool with:
- `title`: The first markdown heading from the content, or the filename if no heading found
- `file_path`: Absolute path to the file
- `id`: The filename (so updates replace the previous version)
- `content_type`: `"md"`

Handle the response:
- If approved with no comments, inform the user and continue.
- If feedback requests changes, act on it — update the content, then call `submit_for_review_and_wait` again.
- Keep iterating until the reviewer approves.

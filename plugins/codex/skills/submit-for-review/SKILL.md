---
name: submit-for-review
description: Send content to Monocle for review
---

When the user asks you to submit content for review, find the most recently modified plan file in the project (or use the path the user provided). Read it to get the filename and first heading.

Call the monocle `submit_for_review` tool with:
- `title`: The first markdown heading from the content, or the filename if no heading found
- `file_path`: Absolute path to the file
- `id`: The filename (so updates replace the previous version)
- `content_type`: `"md"`

This does NOT wait for feedback. If the user needs to block until the reviewer responds, use the submit-for-review-wait skill instead.

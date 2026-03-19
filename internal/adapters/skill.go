package adapters

// SkillContent is the SKILL.md content installed for all agents.
const SkillContent = `---
name: monocle-review
description: Check for and receive code review feedback from your human reviewer who is watching your changes in real-time using Monocle. Invoke after completing a logical unit of work.
allowed-tools: Bash(monocle *)
---

Your reviewer is watching your code changes in real-time using Monocle. They may
leave comments on specific files and lines, or on your plans.

## Check for pending feedback

Run this after completing a logical unit of work (implementing a feature, fixing
a bug, finishing a refactoring step) or before starting a new task:

` + "```" + `
monocle review-status
` + "```" + `

This returns the current review state: no feedback, pending feedback (with count),
or a pause request from your reviewer.

## Retrieve feedback

If review-status shows pending feedback:

` + "```" + `
monocle get-feedback
` + "```" + `

This returns formatted review comments with file paths, line numbers, and code
snippets. Address the feedback before continuing.

## When your reviewer requests a pause

If review-status returns "pause_requested", your reviewer wants you to stop and
wait for their feedback. Run:

` + "```" + `
monocle get-feedback --wait
` + "```" + `

This blocks until your reviewer submits their review. Do not continue working
until you receive the feedback.

## Submit a plan or content for review

When you produce a plan, architecture decision, or other content you want your
reviewer to see:

` + "```" + `
monocle submit-content --title "Your Title" <<'MONOCLE_EOF'
[your content here]
MONOCLE_EOF
` + "```" + `

Your reviewer will see this alongside file diffs and can leave line-level
comments on it.
`

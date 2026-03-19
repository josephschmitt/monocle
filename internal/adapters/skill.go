package adapters

// SkillContent is the SKILL.md content installed for all agents.
const SkillContent = `---
name: monocle-review
description: Check for and receive code review feedback from your human reviewer who is watching your changes in real-time using Monocle. Invoke at the start of every new task AND after completing a logical unit of work.
allowed-tools: Bash(monocle *)
---

Your reviewer is watching your code changes in real-time using Monocle. They may
leave comments on specific files and lines, or on your plans.

## Wait for review feedback

When your user asks you to wait for review, or when you want to check in with
your reviewer after completing a logical unit of work, run:

` + "```" + `
monocle get-feedback --wait
` + "```" + `

This blocks until your reviewer submits their review. The reviewer will see your
code changes and any plans you have submitted. Do not continue working until you
receive the feedback. Address all issues before proceeding.

## Quick status check

To check if feedback is pending without blocking:

` + "```" + `
monocle review-status
` + "```" + `

If feedback is pending, retrieve it with:

` + "```" + `
monocle get-feedback
` + "```" + `

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

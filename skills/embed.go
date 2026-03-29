// Package skills provides embedded SKILL.md files following the agentskills.io
// standard. These are shared by all agent adapters and installed by monocle register.
package skills

import "embed"

//go:embed all:get-feedback all:review-plan all:review-plan-wait
var FS embed.FS

// Names lists the skill directories embedded in the binary.
var Names = []string{"get-feedback", "review-plan", "review-plan-wait"}

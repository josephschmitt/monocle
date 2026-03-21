package tui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// renderSplash renders the monocle logo and getting-started hints centered in the given dimensions.
func renderSplash(width, height int) string {
	var lines []string
	if width >= 48 && height >= 16 {
		lines = splashFull()
	} else {
		lines = splashSmall()
	}
	return centerBlock(lines, width, height)
}

// splashFull renders the logo with getting-started instructions.
func splashFull() []string {
	logo := lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Bold(true)
	dim := lipgloss.NewStyle().Faint(true)
	cmd := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))

	const cmdCol = 3

	hint := func(command, desc string) string {
		return dim.Render("press ") + cmd.Render(lipgloss.NewStyle().Width(cmdCol).Render(command)) + dim.Render(desc)
	}

	return []string{
		logo.Render("o_(◉) monocle"),
		dim.Render("code review companion for Claude Code"),
		"",
		dim.Render("Run " + cmd.Render("monocle install") + dim.Render(" to set up the MCP channel.")),
		dim.Render("Then launch Claude Code in this repo with:"),
		dim.Render("  " + cmd.Render("claude --dangerously-allow-channels")),
		"",
		dim.Render("Diffs appear here as Claude Code works."),
		"",
		hint("c", "to comment on a line"),
		hint("C", "to comment on a file"),
		hint("S", "to submit your review"),
		"",
		hint("?", "for keybinding help"),
		hint("q", "to quit"),
	}
}

// splashSmall renders a minimal fallback for narrow/short panes.
func splashSmall() []string {
	logo := lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Bold(true)
	dim := lipgloss.NewStyle().Faint(true)
	return []string{
		logo.Render("o_(◉) monocle"),
		dim.Render("Press ? for help"),
	}
}

// centerBlock centers a block of pre-styled lines both vertically and horizontally.
// Uses lipgloss.Width() for correct ANSI-aware width measurement.
func centerBlock(lines []string, width, height int) string {
	if width == 0 || height == 0 {
		return strings.Join(lines, "\n")
	}

	// Find the widest line (visual width, ANSI-safe).
	maxW := 0
	for _, line := range lines {
		if w := lipgloss.Width(line); w > maxW {
			maxW = w
		}
	}

	var b strings.Builder

	// Vertical centering: place block at 1/3 from top.
	blockHeight := len(lines)
	padTop := (height - blockHeight) / 3
	if padTop < 0 {
		padTop = 0
	}
	for i := 0; i < padTop; i++ {
		b.WriteString("\n")
	}

	// Horizontal centering: center the block as a whole.
	blockLeft := (width - maxW) / 2
	if blockLeft < 0 {
		blockLeft = 0
	}

	for i, line := range lines {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(strings.Repeat(" ", blockLeft))
		b.WriteString(line)
	}

	return b.String()
}

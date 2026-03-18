package tui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/anthropics/monocle/internal/types"
)

type statusBarModel struct {
	agentStatus    types.AgentStatus
	agentName      string
	baseRef        string
	fileCount      int
	commentCount   int
	feedbackStatus string
	commandMode    bool
	commandBuffer  string
	width          int
	theme          Theme
}

func newStatusBarModel(theme Theme) statusBarModel {
	return statusBarModel{
		agentStatus: types.AgentStatusIdle,
		theme:       theme,
	}
}

func (m statusBarModel) View() string {
	if m.width == 0 {
		return ""
	}

	if m.commandMode {
		cmdLine := fmt.Sprintf(":%s█", m.commandBuffer)
		return m.theme.StatusBar.Width(m.width).Render(cmdLine)
	}

	// Agent status
	var statusStr string
	var statusStyle lipgloss.Style
	switch m.agentStatus {
	case types.AgentStatusIdle:
		statusStr = "IDLE"
		statusStyle = m.theme.StatusIdle
	case types.AgentStatusWorking:
		statusStr = "WORKING"
		statusStyle = m.theme.StatusWorking
	case types.AgentStatusStopped:
		statusStr = "STOPPED"
		statusStyle = m.theme.StatusStopped
	default:
		statusStr = "IDLE"
		statusStyle = m.theme.StatusIdle
	}
	status := statusStyle.Bold(true).Render(fmt.Sprintf("[%s]", statusStr))

	// Info sections
	parts := []string{status}

	if m.baseRef != "" {
		ref := m.baseRef
		if len(ref) > 8 {
			ref = ref[:8]
		}
		parts = append(parts, fmt.Sprintf("ref:%s", ref))
	}

	if m.agentName != "" {
		parts = append(parts, m.agentName)
	}

	parts = append(parts, fmt.Sprintf("%d files", m.fileCount))
	parts = append(parts, fmt.Sprintf("%d comments", m.commentCount))

	if m.feedbackStatus != "" && m.feedbackStatus != "none" {
		fbStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
		parts = append(parts, fbStyle.Render(m.feedbackStatus))
	}

	// Key hints (right-aligned, collapse to ?:help when narrow)
	fullHints := "c:comment  S:submit  D:dismiss  q:quit"
	shortHints := "?:help"
	left := strings.Join(parts, "  ")

	leftW := lipgloss.Width(left)
	hints := fullHints
	if leftW+len(fullHints)+2 > m.width {
		hints = shortHints
	}

	gap := m.width - leftW - len(hints) - 2
	if gap < 1 {
		gap = 1
	}

	styledHints := lipgloss.NewStyle().Faint(true).Render(hints)
	bar := left + strings.Repeat(" ", gap) + styledHints
	return m.theme.StatusBar.Render(bar)
}

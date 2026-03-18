package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/anthropics/monocle/internal/types"
)

type reviewSummaryModel struct {
	active       bool
	summary      *types.ReviewSummary
	agentStopped bool
	width        int
	height       int
	theme        Theme
}

func newReviewSummaryModel(theme Theme) reviewSummaryModel {
	return reviewSummaryModel{theme: theme}
}

type confirmSubmitMsg struct{}
type cancelSubmitMsg struct{}

func (m reviewSummaryModel) Init() tea.Cmd {
	return nil
}

func (m reviewSummaryModel) Update(msg tea.Msg) (reviewSummaryModel, tea.Cmd) {
	if !m.active {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			m.active = false
			return m, func() tea.Msg { return confirmSubmitMsg{} }
		case "esc":
			m.active = false
			return m, func() tea.Msg { return cancelSubmitMsg{} }
		}
	}
	return m, nil
}

func (m reviewSummaryModel) View() string {
	if !m.active || m.summary == nil {
		return ""
	}

	modalWidth := m.width * 2 / 3
	if modalWidth < 50 {
		modalWidth = 50
	}
	if modalWidth > m.width-4 {
		modalWidth = m.width - 4
	}

	var b strings.Builder

	b.WriteString(lipgloss.NewStyle().Bold(true).Render("Review Summary"))
	b.WriteString("\n\n")

	// Counts
	if m.summary.IssueCt > 0 {
		b.WriteString(fmt.Sprintf("  Issues:      %d\n", m.summary.IssueCt))
	}
	if m.summary.SuggestionCt > 0 {
		b.WriteString(fmt.Sprintf("  Suggestions: %d\n", m.summary.SuggestionCt))
	}
	if m.summary.NoteCt > 0 {
		b.WriteString(fmt.Sprintf("  Notes:       %d\n", m.summary.NoteCt))
	}
	if m.summary.PraiseCt > 0 {
		b.WriteString(fmt.Sprintf("  Praise:      %d\n", m.summary.PraiseCt))
	}
	b.WriteString("\n")

	// Comments by file
	if len(m.summary.FileComments) > 0 {
		b.WriteString(lipgloss.NewStyle().Bold(true).Render("Files:"))
		b.WriteString("\n")
		for path, cmts := range m.summary.FileComments {
			b.WriteString(fmt.Sprintf("  %s (%d comments)\n", path, len(cmts)))
		}
		b.WriteString("\n")
	}

	// Comments on content items
	if len(m.summary.ContentComments) > 0 {
		b.WriteString(lipgloss.NewStyle().Bold(true).Render("Content Items:"))
		b.WriteString("\n")
		for id, cmts := range m.summary.ContentComments {
			b.WriteString(fmt.Sprintf("  %s (%d comments)\n", id, len(cmts)))
		}
		b.WriteString("\n")
	}

	// Delivery status
	if m.agentStopped {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render("Review will be sent immediately"))
	} else {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render("Review will be queued until agent stops"))
	}
	b.WriteString("\n\n")

	b.WriteString(lipgloss.NewStyle().Faint(true).Render("Enter: submit  Escape: cancel"))

	return m.theme.ModalBorder.Width(modalWidth).Render(b.String())
}

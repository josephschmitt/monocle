package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/anthropics/monocle/internal/types"
)

type historyModel struct {
	active       bool
	submissions  []types.ReviewSubmission
	cursor       int
	viewDetail   bool // true when viewing a single submission's content
	scrollOffset int
	width        int
	height       int
	theme        Theme
}

func newHistoryModel(theme Theme) historyModel {
	return historyModel{theme: theme}
}

type openHistoryMsg struct {
	submissions []types.ReviewSubmission
}

type closeHistoryMsg struct{}

func (m *historyModel) open(submissions []types.ReviewSubmission) {
	m.active = true
	m.submissions = submissions
	m.cursor = 0
	m.viewDetail = false
	m.scrollOffset = 0
}

func (m historyModel) Update(msg tea.Msg) (historyModel, tea.Cmd) {
	if !m.active {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc", "q":
			if m.viewDetail {
				m.viewDetail = false
				m.scrollOffset = 0
				return m, nil
			}
			m.active = false
			return m, func() tea.Msg { return closeHistoryMsg{} }
		case "j", "down":
			if m.viewDetail {
				m.scrollOffset++
			} else if m.cursor < len(m.submissions)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.viewDetail {
				if m.scrollOffset > 0 {
					m.scrollOffset--
				}
			} else if m.cursor > 0 {
				m.cursor--
			}
		case "enter":
			if !m.viewDetail && len(m.submissions) > 0 {
				m.viewDetail = true
				m.scrollOffset = 0
			}
		}
	}
	return m, nil
}

func (m historyModel) View() string {
	if !m.active {
		return ""
	}

	modalWidth := calcModalWidth(m.width, 0)

	var b strings.Builder

	if m.viewDetail && m.cursor < len(m.submissions) {
		// Detail view: show the full formatted review
		sub := m.submissions[m.cursor]
		b.WriteString(lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("Round %d — %s", sub.ReviewRound, sub.Action)))
		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().Faint(true).Render(sub.SubmittedAt.Format("2006-01-02 15:04:05")))
		b.WriteString("\n\n")

		// Show the formatted review with scrolling
		lines := strings.Split(sub.FormattedReview, "\n")
		vpH := m.viewportHeight()
		if m.scrollOffset > len(lines)-vpH {
			m.scrollOffset = len(lines) - vpH
		}
		if m.scrollOffset < 0 {
			m.scrollOffset = 0
		}
		end := m.scrollOffset + vpH
		if end > len(lines) {
			end = len(lines)
		}
		visible := lines[m.scrollOffset:end]
		b.WriteString(strings.Join(visible, "\n"))

		b.WriteString("\n\n")
		b.WriteString(lipgloss.NewStyle().Faint(true).Render("Esc: back  j/k: scroll"))
	} else {
		// List view
		b.WriteString(lipgloss.NewStyle().Bold(true).Render("Review History"))
		b.WriteString("\n\n")

		if len(m.submissions) == 0 {
			b.WriteString(lipgloss.NewStyle().Faint(true).Render("No submissions yet."))
		} else {
			actionStyle := func(action types.SubmitAction) lipgloss.Style {
				if action == types.ActionApprove {
					return lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
				}
				return lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
			}

			for i, sub := range m.submissions {
				prefix := "  "
				if i == m.cursor {
					prefix = "> "
				}
				line := fmt.Sprintf("%sRound %d  %s  %d comment(s)  %s",
					prefix,
					sub.ReviewRound,
					actionStyle(sub.Action).Render(string(sub.Action)),
					sub.CommentCount,
					lipgloss.NewStyle().Faint(true).Render(sub.SubmittedAt.Format("15:04:05")),
				)
				if i == m.cursor {
					b.WriteString(lipgloss.NewStyle().Reverse(true).Render(line))
				} else {
					b.WriteString(line)
				}
				b.WriteString("\n")
			}
		}

		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().Faint(true).Render("Enter: view details  Esc: close"))
	}

	return m.theme.ModalBorder.Width(modalWidth).Render(b.String())
}

func (m historyModel) viewportHeight() int {
	const chrome = 10
	h := m.height - chrome
	if h < 1 {
		h = 1
	}
	return h
}

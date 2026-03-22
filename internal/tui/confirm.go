package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type confirmAction int

const (
	confirmClearAfterSubmit confirmAction = iota
	confirmDiscard
)

type confirmModel struct {
	active  bool
	title   string
	message string
	action  confirmAction
	width   int
	theme   Theme
}

func newConfirmModel(theme Theme) confirmModel {
	return confirmModel{theme: theme}
}

type confirmActionMsg struct {
	action confirmAction
}

type cancelConfirmMsg struct{}

func (m *confirmModel) open(title, message string, action confirmAction) {
	m.active = true
	m.title = title
	m.message = message
	m.action = action
}

func (m confirmModel) Update(msg tea.Msg) (confirmModel, tea.Cmd) {
	if !m.active {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter", "y":
			m.active = false
			action := m.action
			return m, func() tea.Msg { return confirmActionMsg{action: action} }
		case "esc", "n":
			m.active = false
			return m, func() tea.Msg { return cancelConfirmMsg{} }
		}
	}
	return m, nil
}

func (m confirmModel) View() string {
	if !m.active {
		return ""
	}

	modalWidth := calcModalWidth(m.width, 0)

	var b strings.Builder

	b.WriteString(lipgloss.NewStyle().Bold(true).Render(m.title))
	b.WriteString("\n\n")
	b.WriteString(m.message)
	b.WriteString("\n\n")
	b.WriteString(lipgloss.NewStyle().Faint(true).Render("Y/Enter: confirm  N/Esc: cancel"))

	return m.theme.ModalBorder.Width(modalWidth).Render(b.String())
}

package tui

import (
	"fmt"
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
	active      bool
	title       string
	message     string
	action      confirmAction
	showDontAsk bool
	dontAsk     bool
	width       int
	height      int
	theme       Theme
}

func newConfirmModel(theme Theme) confirmModel {
	return confirmModel{theme: theme}
}

type confirmActionMsg struct {
	action  confirmAction
	dontAsk bool
}

type cancelConfirmMsg struct {
	dontAsk bool
}

func (m *confirmModel) open(title, message string, action confirmAction) {
	m.active = true
	m.title = title
	m.message = message
	m.action = action
	m.showDontAsk = false
	m.dontAsk = false
}

func (m *confirmModel) openWithDontAsk(title, message string, action confirmAction) {
	m.open(title, message, action)
	m.showDontAsk = true
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
			dontAsk := m.dontAsk
			return m, func() tea.Msg { return confirmActionMsg{action: action, dontAsk: dontAsk} }
		case "esc", "n":
			m.active = false
			dontAsk := m.dontAsk
			return m, func() tea.Msg { return cancelConfirmMsg{dontAsk: dontAsk} }
		case "shift+tab":
			if m.showDontAsk {
				m.dontAsk = !m.dontAsk
			}
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

	if m.showDontAsk {
		check := " "
		if m.dontAsk {
			check = "x"
		}
		b.WriteString(fmt.Sprintf("[%s] Don't ask again this session\n\n", check))
	}

	hints := "Y/Enter: confirm  N/Esc: cancel"
	if m.showDontAsk {
		hints += "  Shift+Tab: don't ask again"
	}
	b.WriteString(lipgloss.NewStyle().Faint(true).Render(hints))

	return m.theme.ModalBorder.Width(modalWidth).Render(b.String())
}

package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type connectionInfoModel struct {
	active          bool
	socketPath      string
	subscriberCount int
	width           int
	height          int
	theme           Theme
}

func newConnectionInfoModel(theme Theme) connectionInfoModel {
	return connectionInfoModel{theme: theme}
}

type closeConnectionInfoMsg struct{}

func (m connectionInfoModel) Update(msg tea.Msg) (connectionInfoModel, tea.Cmd) {
	if !m.active {
		return m, nil
	}
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc", "I":
			m.active = false
			return m, func() tea.Msg { return closeConnectionInfoMsg{} }
		}
	}
	return m, nil
}

func (m connectionInfoModel) View() string {
	if !m.active {
		return ""
	}

	modalWidth := calcModalWidth(m.width, 60)

	var b strings.Builder

	b.WriteString(lipgloss.NewStyle().Bold(true).Render("Connection Info"))
	b.WriteString("\n\n")

	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Bold(true)
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("7"))

	// Connection status
	var statusStr string
	if m.subscriberCount > 0 {
		statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
		statusStr = statusStyle.Render(fmt.Sprintf("Connected (%d subscriber%s)", m.subscriberCount, pluralS(m.subscriberCount)))
	} else {
		statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true)
		statusStr = statusStyle.Render("No subscribers connected")
	}
	b.WriteString(labelStyle.Render("Status:  "))
	b.WriteString(statusStr)
	b.WriteString("\n\n")

	// Socket path
	b.WriteString(labelStyle.Render("Socket:  "))
	b.WriteString(valueStyle.Render(m.socketPath))
	b.WriteString("\n\n")

	// Manual override hint
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	b.WriteString(hintStyle.Render("Manual override:"))
	b.WriteString("\n")
	codeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	b.WriteString("  Monocle:    ")
	b.WriteString(codeStyle.Render(fmt.Sprintf("monocle --socket %s", m.socketPath)))
	b.WriteString("\n")
	b.WriteString("  Channel:    ")
	b.WriteString(codeStyle.Render(fmt.Sprintf("MONOCLE_SOCKET=%s", m.socketPath)))
	b.WriteString("\n\n")

	b.WriteString(lipgloss.NewStyle().Faint(true).Render("Press I or Esc to close"))

	return m.theme.ModalBorder.Width(modalWidth).Render(b.String())
}

func pluralS(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

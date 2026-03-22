package tui

import (
	"image/color"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type installPromptModel struct {
	active bool
	global bool
	width  int
	height int
	theme  Theme
}

func newInstallPromptModel(theme Theme) installPromptModel {
	return installPromptModel{theme: theme}
}

type installMCPMsg struct {
	global bool
}

type cancelInstallMsg struct{}

func (m *installPromptModel) open() {
	m.active = true
	m.global = false
}

func (m installPromptModel) Update(msg tea.Msg) (installPromptModel, tea.Cmd) {
	if !m.active {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			m.active = false
			global := m.global
			return m, func() tea.Msg { return installMCPMsg{global: global} }
		case "esc":
			m.active = false
			return m, func() tea.Msg { return cancelInstallMsg{} }
		case "tab":
			m.global = !m.global
		}
	}
	return m, nil
}

func (m installPromptModel) View() string {
	if !m.active {
		return ""
	}

	modalWidth := calcModalWidth(m.width, 0)

	var b strings.Builder

	b.WriteString(lipgloss.NewStyle().Bold(true).Render("Install MCP Channel"))
	b.WriteString("\n\n")
	b.WriteString("Monocle's MCP channel is not installed. This is needed to directly communicate with Claude Code during reviews.")
	b.WriteString("\n\n")

	// Scope selector (Tab to cycle)
	options := []struct {
		global bool
		label  string
		desc   string
		color  color.Color
	}{
		{false, "LOCAL", "./.mcp.json", lipgloss.Color("3")},
		{true, "GLOBAL", "~/.mcp.json", lipgloss.Color("4")},
	}
	for i, opt := range options {
		var style lipgloss.Style
		if opt.global == m.global {
			style = lipgloss.NewStyle().
				Background(opt.color).
				Foreground(lipgloss.Color("0")).
				Bold(true).
				Padding(0, 1)
		} else {
			style = lipgloss.NewStyle().
				Foreground(opt.color).
				Padding(0, 1)
		}
		b.WriteString(style.Render(opt.label))
		b.WriteString(" ")
		b.WriteString(lipgloss.NewStyle().Faint(true).Render(opt.desc))
		if i < len(options)-1 {
			b.WriteString("  ")
		}
	}
	b.WriteString("  ")
	b.WriteString(lipgloss.NewStyle().Faint(true).Render("(Tab)"))
	b.WriteString("\n\n")

	b.WriteString(lipgloss.NewStyle().Faint(true).Render("Enter: install  Tab: cycle scope  Esc: skip"))

	return m.theme.ModalBorder.Width(modalWidth).Render(b.String())
}

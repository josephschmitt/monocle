package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type helpModel struct {
	active bool
	width  int
	height int
	theme  Theme
}

func newHelpModel(theme Theme) helpModel {
	return helpModel{theme: theme}
}

type closeHelpMsg struct{}

func (m helpModel) Update(msg tea.Msg) (helpModel, tea.Cmd) {
	if !m.active {
		return m, nil
	}
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc", "?", "q":
			m.active = false
			return m, func() tea.Msg { return closeHelpMsg{} }
		}
	}
	return m, nil
}

func (m helpModel) View() string {
	if !m.active {
		return ""
	}

	modalWidth := m.width * 2 / 3
	if modalWidth < 40 {
		modalWidth = 40
	}
	if modalWidth > 60 {
		modalWidth = 60
	}

	var b strings.Builder

	b.WriteString(lipgloss.NewStyle().Bold(true).Render("Keybindings"))
	b.WriteString("\n\n")

	sections := []struct {
		title string
		keys  []struct{ key, desc string }
	}{
		{"Navigation", []struct{ key, desc string }{
			{"j/k", "Move up/down"},
			{"Ctrl+d/u", "Half page down/up"},
			{"g/G", "Top/bottom"},
			{"J/K", "Scroll diff (from any pane)"},
			{"[/]", "Previous/next file"},
			{"Enter", "Focus diff pane / toggle dir"},
			{"Tab", "Switch pane focus"},
			{"1/2", "Jump to pane"},
			{"b", "Change base ref"},
			{"f", "Toggle flat/tree view"},
			{"z/e", "Collapse/expand all (tree)"},
		}},
		{"Review", []struct{ key, desc string }{
			{"c", "Add comment at cursor"},
			{"v", "Visual select mode"},
			{"r", "Toggle file reviewed"},
			{"S / :submit", "Submit review (approve if no comments)"},
			{"P / :pause", "Toggle pause (ask Claude Code to wait)"},
			{"D / :dismiss-outdated", "Dismiss outdated comments"},
		}},
		{"General", []struct{ key, desc string }{
			{"t", "Toggle unified/split diff"},
			{"?", "Show this help"},
			{"q", "Quit"},
		}},
	}

	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Bold(true).Width(20)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	sectionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Bold(true)

	for i, section := range sections {
		b.WriteString(sectionStyle.Render(section.title))
		b.WriteString("\n")
		for _, k := range section.keys {
			b.WriteString("  " + keyStyle.Render(k.key) + descStyle.Render(k.desc) + "\n")
		}
		if i < len(sections)-1 {
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Faint(true).Render("Press ? or Esc to close"))

	return m.theme.ModalBorder.Width(modalWidth).Render(b.String())
}

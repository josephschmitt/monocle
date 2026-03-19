package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/anthropics/monocle/internal/core"
)

type refPickerModel struct {
	entries    []core.LogEntry
	cursor     int
	width      int
	height     int
	active     bool
	autoActive bool // whether auto-advance is currently on
	theme      Theme
}

func newRefPickerModel(theme Theme) refPickerModel {
	return refPickerModel{theme: theme}
}

type openRefPickerMsg struct {
	entries    []core.LogEntry
	autoActive bool
}

type selectRefMsg struct {
	hash string
	auto bool
}

type cancelRefPickerMsg struct{}

func (m refPickerModel) Update(msg tea.Msg) (refPickerModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "j", "down":
			if m.cursor < len(m.entries) {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "enter":
			if m.cursor == 0 {
				// "Auto (HEAD)" option
				return m, func() tea.Msg { return selectRefMsg{auto: true} }
			}
			idx := m.cursor - 1
			if idx < len(m.entries) {
				return m, func() tea.Msg { return selectRefMsg{hash: m.entries[idx].Hash} }
			}
		case "esc", "q":
			return m, func() tea.Msg { return cancelRefPickerMsg{} }
		}
	}
	return m, nil
}

func (m refPickerModel) View() string {
	if !m.active {
		return ""
	}

	var b strings.Builder

	title := lipgloss.NewStyle().Bold(true).Render("Select Base Ref")
	b.WriteString(title + "\n\n")

	// Auto option
	autoLabel := "  Auto (follow HEAD)"
	if m.autoActive {
		autoLabel = "  Auto (follow HEAD) ✓"
	}
	if m.cursor == 0 {
		b.WriteString(lipgloss.NewStyle().Reverse(true).Render(autoLabel))
	} else {
		b.WriteString(autoLabel)
	}
	b.WriteString("\n\n")

	// Commit entries
	maxVisible := 15
	if maxVisible > len(m.entries) {
		maxVisible = len(m.entries)
	}
	for i := 0; i < maxVisible; i++ {
		entry := m.entries[i]
		line := fmt.Sprintf("  %s %s", entry.Hash, entry.Subject)
		if len(line) > m.width-6 && m.width > 10 {
			line = line[:m.width-9] + "..."
		}
		if m.cursor == i+1 {
			b.WriteString(lipgloss.NewStyle().Reverse(true).Render(line))
		} else {
			b.WriteString(line)
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Faint(true).Render("  enter:select  esc:cancel"))

	boxW := m.width - 20
	if boxW < 40 {
		boxW = 40
	}
	if boxW > 80 {
		boxW = 80
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("4")).
		Padding(1, 2).
		Width(boxW).
		Render(b.String())
}

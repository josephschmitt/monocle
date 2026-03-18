package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/anthropics/monocle/internal/types"
)

type sidebarModel struct {
	files        []types.ChangedFile
	contentItems []types.ContentItem
	cursor       int
	width        int
	height       int
	focused      bool
	recentPaths  map[string]bool
}

func newSidebarModel() sidebarModel {
	return sidebarModel{
		recentPaths: make(map[string]bool),
	}
}

type sidebarSelectMsg struct {
	path        string
	isContent   bool
	contentID   string
}

type recentFadeMsg struct {
	path string
}

func (m sidebarModel) Init() tea.Cmd {
	return nil
}

func (m sidebarModel) Update(msg tea.Msg) (sidebarModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if !m.focused {
			return m, nil
		}
		switch msg.String() {
		case "j", "down":
			if m.cursor < m.totalItems()-1 {
				m.cursor++
			}
			return m, m.selectCurrent()
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, m.selectCurrent()
		case "g":
			m.cursor = 0
		case "G":
			if total := m.totalItems(); total > 0 {
				m.cursor = total - 1
			}
		case "enter":
			return m, m.selectCurrent()
		case "]":
			if m.cursor < m.totalItems()-1 {
				m.cursor++
			}
			return m, m.selectCurrent()
		case "[":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, m.selectCurrent()
		}
	}
	return m, nil
}

func (m sidebarModel) View() string {
	if m.width == 0 {
		return ""
	}

	var b strings.Builder

	// Header
	fileCount := len(m.files)
	reviewedCount := 0
	for _, f := range m.files {
		if f.Reviewed {
			reviewedCount++
		}
	}
	header := fmt.Sprintf(" Files  %d / %d", reviewedCount, fileCount)
	headerStyle := lipgloss.NewStyle().Bold(true).Width(m.width)
	b.WriteString(headerStyle.Render(header))
	b.WriteString("\n")

	// File list
	idx := 0
	for _, f := range m.files {
		line := m.renderFileItem(f, idx == m.cursor)
		b.WriteString(line)
		b.WriteString("\n")
		idx++
	}

	// Content items section
	if len(m.contentItems) > 0 {
		b.WriteString("\n")
		divider := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Width(m.width)
		b.WriteString(divider.Render(" Review Items"))
		b.WriteString("\n")

		for _, item := range m.contentItems {
			line := m.renderContentItem(item, idx == m.cursor)
			b.WriteString(line)
			b.WriteString("\n")
			idx++
		}
	}

	return b.String()
}

func (m sidebarModel) renderFileItem(f types.ChangedFile, selected bool) string {
	// Status indicator
	var statusChar string
	switch f.Status {
	case types.FileAdded:
		statusChar = "A"
	case types.FileModified:
		statusChar = "M"
	case types.FileDeleted:
		statusChar = "D"
	case types.FileRenamed:
		statusChar = "R"
	default:
		statusChar = "?"
	}

	// Review status
	reviewChar := "○"
	if f.Reviewed {
		reviewChar = "✓"
	}

	// Recent indicator
	recentChar := " "
	if m.recentPaths[f.Path] {
		recentChar = "~"
	}

	// Build line
	icon := fileIcon(f.Path)
	name := truncatePath(f.Path, m.width-11)
	line := fmt.Sprintf(" %s %s%s %s %s", statusChar, recentChar, icon, name, reviewChar)

	if selected && m.focused {
		style := lipgloss.NewStyle().Reverse(true).Width(m.width)
		return style.Render(line)
	}

	// Color status character
	var statusStyle lipgloss.Style
	switch f.Status {
	case types.FileAdded:
		statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	case types.FileDeleted:
		statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	case types.FileModified:
		statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	default:
		statusStyle = lipgloss.NewStyle()
	}
	_ = statusStyle // Used in full render; keeping simple for now

	return fmt.Sprintf("%-*s", m.width, line)
}

func (m sidebarModel) renderContentItem(item types.ContentItem, selected bool) string {
	reviewChar := "○"
	if item.Reviewed {
		reviewChar = "✓"
	}

	name := truncatePath(item.Title, m.width-6)
	line := fmt.Sprintf("   %s %s", name, reviewChar)

	if selected && m.focused {
		style := lipgloss.NewStyle().Reverse(true).Width(m.width)
		return style.Render(line)
	}
	return fmt.Sprintf("%-*s", m.width, line)
}

func (m sidebarModel) totalItems() int {
	return len(m.files) + len(m.contentItems)
}

func (m sidebarModel) selectCurrent() tea.Cmd {
	if m.cursor < len(m.files) {
		path := m.files[m.cursor].Path
		return func() tea.Msg {
			return sidebarSelectMsg{path: path}
		}
	}
	contentIdx := m.cursor - len(m.files)
	if contentIdx < len(m.contentItems) {
		item := m.contentItems[contentIdx]
		return func() tea.Msg {
			return sidebarSelectMsg{isContent: true, contentID: item.ID}
		}
	}
	return nil
}

func truncatePath(path string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	if len(path) <= maxLen {
		return path
	}
	if maxLen <= 3 {
		return path[:maxLen]
	}
	return "..." + path[len(path)-(maxLen-3):]
}

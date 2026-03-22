package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/anthropics/monocle/internal/core"
)

const refPickerPageSize = 20

type refPickerModel struct {
	entries    []core.LogEntry
	cursor     int
	offset     int // scroll offset for visible entries
	width      int
	height     int
	active     bool
	autoActive bool // whether auto-advance is currently on
	hasMore    bool // true if last fetch returned a full page
	loading    bool // true while loading more entries
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

type loadMoreRefsMsg struct {
	entries []core.LogEntry
	hasMore bool
}

func (m refPickerModel) maxCursor() int {
	n := len(m.entries) // last commit is at cursor position len(entries)
	if m.hasMore {
		n++ // "Load more..." option
	}
	return n
}

func (m refPickerModel) Update(msg tea.Msg) (refPickerModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "j", "down":
			if m.cursor < m.maxCursor() {
				m.cursor++
				m.ensureVisible()
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
				m.ensureVisible()
			}
		case "enter":
			if m.cursor == 0 {
				// "Auto (HEAD)" option
				return m, func() tea.Msg { return selectRefMsg{auto: true} }
			}
			// "Load more..." option
			if m.hasMore && m.cursor == len(m.entries)+1 {
				m.loading = true
				return m, nil // app.go handles the actual fetch
			}
			idx := m.cursor - 1
			if idx < len(m.entries) {
				return m, func() tea.Msg { return selectRefMsg{hash: m.entries[idx].Hash} }
			}
		case "esc", "q":
			return m, func() tea.Msg { return cancelRefPickerMsg{} }
		}

	case loadMoreRefsMsg:
		m.entries = msg.entries
		m.hasMore = msg.hasMore
		m.loading = false
		m.ensureVisible()
	}
	return m, nil
}

func (m refPickerModel) View() string {
	if !m.active {
		return ""
	}

	boxW := calcModalWidth(m.width, 80)
	contentW := boxW - 6 // 2 border + 4 padding (2 each side)

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

	// Commit entries with scrolling
	vh := m.viewportHeight()
	end := m.offset + vh
	if end > len(m.entries) {
		end = len(m.entries)
	}

	faintStyle := lipgloss.NewStyle().Faint(true)
	if m.offset > 0 {
		b.WriteString(faintStyle.Render("  ▲ more") + "\n")
	}

	// Determine hash column width from the first entry (typically 7 chars)
	hashWidth := 7
	if len(m.entries) > 0 {
		hashWidth = len(m.entries[0].Hash)
	}
	hashColW := hashWidth + 3 // "  " prefix + hash + " " gap
	subjectW := contentW - hashColW
	if subjectW < 10 {
		subjectW = 10
	}
	hashStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Width(hashColW).
		PaddingLeft(2).PaddingRight(1)
	subjectStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("7")).Width(subjectW)
	plainHashStyle := lipgloss.NewStyle().Width(hashColW).PaddingLeft(2).PaddingRight(1)
	plainSubjectStyle := lipgloss.NewStyle().Width(subjectW)
	for i := m.offset; i < end; i++ {
		entry := m.entries[i]
		var line string
		if m.cursor == i+1 {
			// Render without foreground colors so Reverse works cleanly
			hashBlock := plainHashStyle.Render(entry.Hash)
			subjectBlock := plainSubjectStyle.Render(entry.Subject)
			line = lipgloss.JoinHorizontal(lipgloss.Top, hashBlock, subjectBlock)
			// Pad each line to full width so the highlight spans the row
			parts := strings.Split(line, "\n")
			for j, l := range parts {
				if w := lipgloss.Width(l); w < contentW {
					parts[j] = l + strings.Repeat(" ", contentW-w)
				}
			}
			line = lipgloss.NewStyle().Reverse(true).Render(strings.Join(parts, "\n"))
		} else {
			hashBlock := hashStyle.Render(entry.Hash)
			subjectBlock := subjectStyle.Render(entry.Subject)
			line = lipgloss.JoinHorizontal(lipgloss.Top, hashBlock, subjectBlock)
		}
		b.WriteString(line + "\n")
	}

	if m.hasMore {
		loadMoreLabel := "  Load more..."
		if m.loading {
			loadMoreLabel = "  Loading..."
		}
		if m.cursor == len(m.entries)+1 {
			b.WriteString(lipgloss.NewStyle().Reverse(true).Render(loadMoreLabel))
		} else {
			b.WriteString(faintStyle.Render(loadMoreLabel))
		}
		b.WriteString("\n")
	} else if end < len(m.entries) {
		b.WriteString(faintStyle.Render("  ▼ more") + "\n")
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Faint(true).Render("  enter:select  esc:cancel"))

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("4")).
		Padding(1, 2).
		Width(boxW).
		Render(b.String())
}

// viewportHeight returns how many commit entries fit in the ref picker.
// Uses a stable calculation based only on screen size so the modal
// doesn't resize while scrolling.
func (m refPickerModel) viewportHeight() int {
	// Cap modal to ~60% of screen height, minus chrome
	// Chrome: title(1) + blank(1) + auto line(1) + blank(1) + load-more/footer(2) + padding(2) + border(2) = 10
	maxModalH := m.height * 3 / 5
	availableLines := maxModalH - 10
	if availableLines < 1 {
		return 1
	}
	// Budget ~2/3 of available lines for entries to account for wrapping
	vh := availableLines * 2 / 3
	if vh < 1 {
		vh = 1
	}
	return vh
}

// ensureVisible adjusts the scroll offset so the cursor stays within the
// visible viewport.
func (m *refPickerModel) ensureVisible() {
	// cursor 0 is the "Auto" option which is always visible above the list
	if m.cursor <= 0 {
		m.offset = 0
		return
	}
	// entries use 0-based indexing, cursor 1 = entries[0]
	entryIdx := m.cursor - 1
	if entryIdx < m.offset {
		m.offset = entryIdx
	}
	vh := m.viewportHeight()
	if vh > 0 && entryIdx >= m.offset+vh {
		m.offset = entryIdx - vh + 1
	}
}

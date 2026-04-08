package tui

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/josephschmitt/monocle/internal/core"
	"github.com/josephschmitt/monocle/internal/types"
)

// formatTimeAgo returns a human-readable relative time string.
func formatTimeAgo(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		if m == 1 {
			return "1m ago"
		}
		return fmt.Sprintf("%dm ago", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		if h == 1 {
			return "1h ago"
		}
		return fmt.Sprintf("%dh ago", h)
	default:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "1d ago"
		}
		return fmt.Sprintf("%dd ago", days)
	}
}

const refPickerPageSize = 20

type refPickerModel struct {
	entries    []core.LogEntry
	snapshots  []types.ReviewSnapshot
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
	snapshots  []types.ReviewSnapshot
	autoActive bool
}

type selectRefMsg struct {
	hash string
	auto bool
}

type selectSnapshotMsg struct {
	snapshotID int
}

type cancelRefPickerMsg struct{}

type loadMoreRefsMsg struct {
	entries []core.LogEntry
	hasMore bool
}

// snapshotCount returns the number of snapshot entries in the picker.
func (m refPickerModel) snapshotCount() int {
	return len(m.snapshots)
}

// commitStartCursor returns the cursor position of the first commit entry.
func (m refPickerModel) commitStartCursor() int {
	return 1 + m.snapshotCount() // 0=auto, then snapshots
}

func (m refPickerModel) maxCursor() int {
	n := m.commitStartCursor() + len(m.entries) - 1
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
		case "g":
			m.cursor = 0
			m.ensureVisible()
		case "G":
			m.cursor = m.maxCursor()
			m.ensureVisible()
		case "enter":
			if m.cursor == 0 {
				// "Latest Changes" option
				return m, func() tea.Msg { return selectRefMsg{auto: true} }
			}
			// Snapshot entries
			if m.cursor >= 1 && m.cursor <= m.snapshotCount() {
				snapIdx := m.cursor - 1
				id := m.snapshots[snapIdx].ID
				return m, func() tea.Msg { return selectSnapshotMsg{snapshotID: id} }
			}
			// "Load more..." option
			commitStart := m.commitStartCursor()
			if m.hasMore && m.cursor == commitStart+len(m.entries) {
				m.loading = true
				return m, nil // app.go handles the actual fetch
			}
			idx := m.cursor - commitStart
			if idx >= 0 && idx < len(m.entries) {
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

	// "Latest Changes" option (was "Auto (follow HEAD)")
	autoLabel := "  Latest Changes"
	if m.autoActive {
		autoLabel = "  Latest Changes ✓"
	}
	if m.cursor == 0 {
		b.WriteString(lipgloss.NewStyle().Reverse(true).Render(autoLabel))
	} else {
		b.WriteString(autoLabel)
	}
	b.WriteString("\n")

	// Snapshot entries
	if len(m.snapshots) > 0 {
		b.WriteString("\n")
		faintSep := lipgloss.NewStyle().Faint(true)
		b.WriteString(faintSep.Render("  Since Review") + "\n")
		for i, snap := range m.snapshots {
			label := fmt.Sprintf("  Round %d (%s)", snap.ReviewRound, formatTimeAgo(snap.CreatedAt))
			cursorPos := 1 + i
			if m.cursor == cursorPos {
				b.WriteString(lipgloss.NewStyle().Reverse(true).Render(label))
			} else {
				b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Render(label))
			}
			b.WriteString("\n")
		}
	}
	b.WriteString("\n")

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
	commitStart := m.commitStartCursor()
	for i := m.offset; i < end; i++ {
		entry := m.entries[i]
		var line string
		if m.cursor == commitStart+i {
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
		if m.cursor == commitStart+len(m.entries) {
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

// handleClick processes a mouse click at content-relative coordinates.
// Returns a command and whether the click was handled.
func (m *refPickerModel) handleClick(contentY int) (tea.Cmd, bool) {
	// Content layout:
	// Line 0: "Select Base Ref"
	// Line 1: blank
	// Line 2: Latest Changes option
	// Line 3: blank (or snapshot header if snapshots exist)
	// Then: snapshot entries, blank, commit entries

	// Latest Changes option
	if contentY == 2 {
		m.cursor = 0
		return func() tea.Msg { return selectRefMsg{auto: true} }, true
	}

	// Snapshot entries (if any)
	if len(m.snapshots) > 0 {
		// Line 3: blank, Line 4: "Since Review" header, Lines 5+: snapshot entries
		snapStartLine := 5
		for i := range m.snapshots {
			if contentY == snapStartLine+i {
				m.cursor = 1 + i
				id := m.snapshots[i].ID
				return func() tea.Msg { return selectSnapshotMsg{snapshotID: id} }, true
			}
		}
	}

	// Commit entries
	commitHeaderLines := 4 // title + blank + auto + blank
	if len(m.snapshots) > 0 {
		commitHeaderLines += 1 + len(m.snapshots) + 1 // header + entries + blank
	}
	entryStartLine := commitHeaderLines
	if m.offset > 0 {
		entryStartLine++ // "▲ more" indicator line
	}

	clickedIdx := contentY - entryStartLine + m.offset
	if clickedIdx >= 0 && clickedIdx < len(m.entries) {
		m.cursor = m.commitStartCursor() + clickedIdx
		hash := m.entries[clickedIdx].Hash
		return func() tea.Msg { return selectRefMsg{hash: hash} }, true
	}

	return nil, false
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
	// cursor 0 is the "Latest Changes" option which is always visible above the list
	// Snapshot entries are also always visible (not scrollable)
	commitStart := m.commitStartCursor()
	if m.cursor < commitStart {
		m.offset = 0
		return
	}
	// entries use 0-based indexing relative to commitStart
	entryIdx := m.cursor - commitStart
	if entryIdx < m.offset {
		m.offset = entryIdx
	}
	vh := m.viewportHeight()
	if vh > 0 && entryIdx >= m.offset+vh {
		m.offset = entryIdx - vh + 1
	}
}

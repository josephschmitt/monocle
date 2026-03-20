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
	offset       int // scroll offset for viewport
	width        int
	height       int
	focused      bool
	recentPaths  map[string]bool

	// Tree mode state
	treeMode     bool
	treeRoots    []*fileTreeNode
	collapsed    map[string]bool
	visibleItems []visibleItem
}

func newSidebarModel() sidebarModel {
	return sidebarModel{
		recentPaths: make(map[string]bool),
		collapsed:   make(map[string]bool),
	}
}

type sidebarSelectMsg struct {
	path      string
	isContent bool
	contentID string
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
			m.ensureVisible()
			return m, m.selectCurrent()
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
			m.ensureVisible()
			return m, m.selectCurrent()
		case "g":
			m.cursor = 0
			m.ensureVisible()
		case "G":
			if total := m.totalItems(); total > 0 {
				m.cursor = total - 1
			}
			m.ensureVisible()
		case "enter":
			if m.treeMode {
				idx := m.cursor
				if idx < len(m.visibleItems) && m.visibleItems[idx].isDir {
					path := m.visibleItems[idx].node.Path
					if m.collapsed[path] {
						delete(m.collapsed, path)
					} else {
						m.collapsed[path] = true
					}
					m.visibleItems = flattenTree(m.treeRoots, m.collapsed)
					// Clamp cursor
					if total := m.totalItems(); total > 0 && m.cursor >= total {
						m.cursor = total - 1
					}
					m.ensureVisible()
					return m, nil
				}
			}
			return m, m.selectCurrent()
		case "]":
			if m.cursor < m.totalItems()-1 {
				m.cursor++
			}
			m.ensureVisible()
			return m, m.selectCurrent()
		case "[":
			if m.cursor > 0 {
				m.cursor--
			}
			m.ensureVisible()
			return m, m.selectCurrent()
		case "f":
			currentPath := ""
			if f := m.selectedFile(); f != nil {
				currentPath = f.Path
			}
			m.treeMode = !m.treeMode
			if m.treeMode {
				m.rebuildTree()
			}
			if currentPath != "" {
				m.selectPath(currentPath)
			}
			if total := m.totalItems(); total > 0 && m.cursor >= total {
				m.cursor = total - 1
			}
			m.ensureVisible()
			return m, m.selectCurrent()
		case "z":
			if m.treeMode {
				m.collapseAll()
				return m, m.selectCurrent()
			}
		case "e":
			if m.treeMode {
				m.collapsed = make(map[string]bool)
				m.visibleItems = flattenTree(m.treeRoots, m.collapsed)
				if total := m.totalItems(); total > 0 && m.cursor >= total {
					m.cursor = total - 1
				}
				m.ensureVisible()
				return m, m.selectCurrent()
			}
		}
	}
	return m, nil
}

func (m sidebarModel) View() string {
	if m.width == 0 {
		return ""
	}

	var b strings.Builder

	// Header (always visible, not scrollable)
	fileCount := len(m.files)
	reviewedCount := 0
	for _, f := range m.files {
		if f.Reviewed {
			reviewedCount++
		}
	}
	modeIndicator := ""
	if m.treeMode {
		modeIndicator = " "
	}
	header := fmt.Sprintf(" Files%s  %d / %d", modeIndicator, reviewedCount, fileCount)
	headerStyle := lipgloss.NewStyle().Bold(true).Width(m.width)
	b.WriteString(headerStyle.Render(header))
	b.WriteString("\n")

	// Render only items within the viewport [offset, offset+viewportHeight)
	fileItemCt := m.fileItemCount()
	totalItems := m.totalItems()
	availableLines := m.viewportHeight()

	linesUsed := 0
	for idx := m.offset; idx < totalItems && linesUsed < availableLines; idx++ {
		// Content section divider (when crossing from files to content items)
		if idx == fileItemCt && len(m.contentItems) > 0 {
			if linesUsed+2 > availableLines {
				break
			}
			b.WriteString("\n")
			divider := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Width(m.width)
			b.WriteString(divider.Render(" Review Items"))
			b.WriteString("\n")
			linesUsed += 2
			if linesUsed >= availableLines {
				break
			}
		}

		var line string
		if idx < fileItemCt {
			if m.treeMode {
				item := m.visibleItems[idx]
				if item.isDir {
					line = m.renderDirItem(item, idx == m.cursor)
				} else {
					line = m.renderTreeFileItem(item, idx == m.cursor)
				}
			} else {
				line = m.renderFileItem(m.files[idx], idx == m.cursor)
			}
		} else {
			contentIdx := idx - fileItemCt
			line = m.renderContentItem(m.contentItems[contentIdx], idx == m.cursor)
		}

		b.WriteString(line)
		b.WriteString("\n")
		linesUsed++
	}

	return b.String()
}

func (m sidebarModel) renderFileItem(f types.ChangedFile, selected bool) string {
	// Status indicator (lazygit-style colors)
	var statusChar, statusColor string
	switch f.Status {
	case types.FileAdded:
		statusChar = "A"
		statusColor = "#2ea043"
	case types.FileModified:
		statusChar = "M"
		statusColor = "#d29922"
	case types.FileDeleted:
		statusChar = "D"
		statusColor = "#f85149"
	case types.FileRenamed:
		statusChar = "R"
		statusColor = "#a371f7"
	default:
		statusChar = "?"
		statusColor = "7"
	}
	styledStatus := lipgloss.NewStyle().Foreground(lipgloss.Color(statusColor)).Bold(true).Render(statusChar)

	// Review status
	reviewChar := "○"
	if f.Reviewed {
		reviewChar = lipgloss.NewStyle().Foreground(lipgloss.Color("#2ea043")).Render("✓")
	}

	// Recent indicator
	recentChar := " "
	if m.recentPaths[f.Path] {
		recentChar = "~"
	}

	// Layout: " {status} {recent}{icon} {name...}  {review} "
	// Icon glyphs render as width 2 in terminals but lipgloss measures them
	// as width 1. We account for this by subtracting iconSlack from nameW
	// and always padding name to a fixed width so alignment is consistent.
	icon := fileIcon(f.Path)
	glyph := iconLookup(f.Path).glyph
	const iconSlack = 2

	if selected && m.focused {
		plainReview := "○"
		if f.Reviewed {
			plainReview = "✓"
		}
		right := " " + plainReview + " "
		prefix := fmt.Sprintf(" %s %s%s ", statusChar, recentChar, glyph)
		nameW := m.width - lipgloss.Width(prefix) - lipgloss.Width(right) - iconSlack
		if nameW < 1 {
			nameW = 1
		}
		name := fmt.Sprintf("%-*s", nameW, truncatePath(f.Path, nameW))
		padded := prefix + name + right
		return lipgloss.NewStyle().Reverse(true).Render(padded)
	}

	right := " " + reviewChar + " "
	prefix := fmt.Sprintf(" %s %s%s ", styledStatus, recentChar, icon)
	nameW := m.width - lipgloss.Width(prefix) - lipgloss.Width(right) - iconSlack
	if nameW < 1 {
		nameW = 1
	}
	name := fmt.Sprintf("%-*s", nameW, truncatePath(f.Path, nameW))
	return prefix + name + right
}

// renderDirItem renders a directory node in tree mode.
func (m sidebarModel) renderDirItem(item visibleItem, selected bool) string {
	indent := strings.Repeat("  ", item.depth)
	arrow := "▼"
	if m.collapsed[item.node.Path] {
		arrow = "▶"
	}

	// Folder icon
	const folderGlyph = "\uf07b" // nf-fa-folder
	const folderColor = "#e8a838"
	const iconSlack = 2

	if selected && m.focused {
		prefix := fmt.Sprintf(" %s%s %s ", indent, arrow, folderGlyph)
		nameW := m.width - lipgloss.Width(prefix) - iconSlack
		if nameW < 1 {
			nameW = 1
		}
		name := fmt.Sprintf("%-*s", nameW, truncatePath(item.node.Name, nameW))
		padded := prefix + name
		return lipgloss.NewStyle().Reverse(true).Render(padded)
	}

	styledArrow := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(arrow)
	styledFolder := lipgloss.NewStyle().Foreground(lipgloss.Color(folderColor)).Render(folderGlyph)
	prefix := fmt.Sprintf(" %s%s %s ", indent, styledArrow, styledFolder)
	nameW := m.width - lipgloss.Width(prefix) - iconSlack
	if nameW < 1 {
		nameW = 1
	}
	dirStyle := lipgloss.NewStyle().Bold(true)
	name := fmt.Sprintf("%-*s", nameW, truncatePath(item.node.Name, nameW))
	return prefix + dirStyle.Render(name)
}

// renderTreeFileItem renders a file node in tree mode with indentation.
func (m sidebarModel) renderTreeFileItem(item visibleItem, selected bool) string {
	f := item.node.File
	indent := strings.Repeat("  ", item.depth)

	var statusChar, statusColor string
	switch f.Status {
	case types.FileAdded:
		statusChar = "A"
		statusColor = "#2ea043"
	case types.FileModified:
		statusChar = "M"
		statusColor = "#d29922"
	case types.FileDeleted:
		statusChar = "D"
		statusColor = "#f85149"
	case types.FileRenamed:
		statusChar = "R"
		statusColor = "#a371f7"
	default:
		statusChar = "?"
		statusColor = "7"
	}

	reviewChar := "○"
	if f.Reviewed {
		reviewChar = lipgloss.NewStyle().Foreground(lipgloss.Color("#2ea043")).Render("✓")
	}

	recentChar := " "
	if m.recentPaths[f.Path] {
		recentChar = "~"
	}

	icon := fileIcon(f.Path)
	glyph := iconLookup(f.Path).glyph
	const iconSlack = 2

	if selected && m.focused {
		plainReview := "○"
		if f.Reviewed {
			plainReview = "✓"
		}
		right := " " + plainReview + " "
		prefix := fmt.Sprintf(" %s%s %s%s ", indent, statusChar, recentChar, glyph)
		nameW := m.width - lipgloss.Width(prefix) - lipgloss.Width(right) - iconSlack
		if nameW < 1 {
			nameW = 1
		}
		name := fmt.Sprintf("%-*s", nameW, truncatePath(item.node.Name, nameW))
		padded := prefix + name + right
		return lipgloss.NewStyle().Reverse(true).Render(padded)
	}

	styledStatus := lipgloss.NewStyle().Foreground(lipgloss.Color(statusColor)).Bold(true).Render(statusChar)
	right := " " + reviewChar + " "
	prefix := fmt.Sprintf(" %s%s %s%s ", indent, styledStatus, recentChar, icon)
	nameW := m.width - lipgloss.Width(prefix) - lipgloss.Width(right) - iconSlack
	if nameW < 1 {
		nameW = 1
	}
	name := fmt.Sprintf("%-*s", nameW, truncatePath(item.node.Name, nameW))
	return prefix + name + right
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
	return m.fileItemCount() + len(m.contentItems)
}

// fileItemCount returns the number of file-related items (files in flat mode,
// visible items in tree mode).
func (m sidebarModel) fileItemCount() int {
	if m.treeMode {
		return len(m.visibleItems)
	}
	return len(m.files)
}

func (m sidebarModel) selectCurrent() tea.Cmd {
	fileCount := m.fileItemCount()

	if m.cursor < fileCount {
		if m.treeMode {
			item := m.visibleItems[m.cursor]
			if item.isDir {
				return nil // Don't send selection for directories
			}
			path := item.node.File.Path
			return func() tea.Msg {
				return sidebarSelectMsg{path: path}
			}
		}
		path := m.files[m.cursor].Path
		return func() tea.Msg {
			return sidebarSelectMsg{path: path}
		}
	}
	contentIdx := m.cursor - fileCount
	if contentIdx < len(m.contentItems) {
		item := m.contentItems[contentIdx]
		return func() tea.Msg {
			return sidebarSelectMsg{isContent: true, contentID: item.ID}
		}
	}
	return nil
}

// selectedFile returns the ChangedFile at the current cursor position,
// or nil if the cursor is on a directory or content item.
func (m sidebarModel) selectedFile() *types.ChangedFile {
	fileCount := m.fileItemCount()
	if m.cursor >= fileCount {
		return nil
	}
	if m.treeMode {
		item := m.visibleItems[m.cursor]
		if item.isDir {
			return nil
		}
		return item.node.File
	}
	return &m.files[m.cursor]
}

// rebuildTree reconstructs the tree from the current file list and updates
// visible items. Safe to call when treeMode is false (no-op).
func (m *sidebarModel) rebuildTree() {
	if !m.treeMode {
		return
	}
	m.treeRoots = buildFileTree(m.files)
	m.visibleItems = flattenTree(m.treeRoots, m.collapsed)
}

// selectPath moves the cursor to the item matching the given file path.
func (m *sidebarModel) selectPath(path string) {
	if m.treeMode {
		for i, item := range m.visibleItems {
			if !item.isDir && item.node.File != nil && item.node.File.Path == path {
				m.cursor = i
				return
			}
		}
	} else {
		for i, f := range m.files {
			if f.Path == path {
				m.cursor = i
				return
			}
		}
	}
}

// collapseAll collapses all directory nodes in the tree.
func (m *sidebarModel) collapseAll() {
	currentPath := ""
	if f := m.selectedFile(); f != nil {
		currentPath = f.Path
	}

	m.collapsed = make(map[string]bool)
	var markCollapsed func(nodes []*fileTreeNode)
	markCollapsed = func(nodes []*fileTreeNode) {
		for _, n := range nodes {
			if n.File == nil {
				m.collapsed[n.Path] = true
				markCollapsed(n.Children)
			}
		}
	}
	markCollapsed(m.treeRoots)
	m.visibleItems = flattenTree(m.treeRoots, m.collapsed)

	if currentPath != "" {
		m.selectPath(currentPath)
	}
	if total := m.totalItems(); total > 0 && m.cursor >= total {
		m.cursor = total - 1
	}
	m.ensureVisible()
}

// ensureVisible adjusts the scroll offset so the cursor stays within the
// visible viewport, mirroring diffViewModel.ensureVisible.
func (m *sidebarModel) ensureVisible() {
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	vh := m.viewportHeight()
	if vh > 0 && m.cursor >= m.offset+vh {
		m.offset = m.cursor - vh + 1
	}
}

// viewportHeight returns how many item lines fit in the sidebar viewport.
// The header always takes 1 line; remaining space is for scrollable items.
func (m sidebarModel) viewportHeight() int {
	h := m.height - 1 // header line
	if h < 0 {
		h = 0
	}
	return h
}

// clampOffset ensures offset and cursor are within valid bounds after the
// item list changes externally.
func (m *sidebarModel) clampOffset() {
	total := m.totalItems()
	if total == 0 {
		m.cursor = 0
		m.offset = 0
		return
	}
	if m.cursor >= total {
		m.cursor = total - 1
	}
	if m.offset >= total {
		m.offset = total - 1
	}
	if m.offset < 0 {
		m.offset = 0
	}
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

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
				idx := m.cursor - len(m.contentItems)
				if idx >= 0 && idx < len(m.visibleItems) && m.visibleItems[idx].isDir {
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

	sectionStyle := lipgloss.NewStyle().Bold(true).Width(m.width)

	// Render only items within the viewport [offset, offset+viewportHeight)
	contentItemCt := len(m.contentItems)
	totalItems := m.totalItems()
	availableLines := m.viewportHeight()

	linesUsed := 0

	// Review Items section header (if any content items exist)
	if contentItemCt > 0 {
		contentReviewed := 0
		for _, item := range m.contentItems {
			if item.Reviewed {
				contentReviewed++
			}
		}
		header := fmt.Sprintf(" Review Items  %d / %d", contentReviewed, contentItemCt)
		b.WriteString(sectionStyle.Render(header))
		b.WriteString("\n")
		linesUsed++
	}

	for idx := m.offset; idx < totalItems && linesUsed < availableLines; idx++ {
		// Files section header (when crossing from content items to files)
		if idx == contentItemCt && contentItemCt > 0 {
			if linesUsed > 0 {
				if linesUsed+1 > availableLines {
					break
				}
				b.WriteString("\n")
				linesUsed++
			}

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
			b.WriteString(sectionStyle.Render(header))
			b.WriteString("\n")
			linesUsed++
			if linesUsed >= availableLines {
				break
			}
		}

		var line string
		if idx < contentItemCt {
			line = m.renderContentItem(m.contentItems[idx], idx == m.cursor)
		} else {
			fileIdx := idx - contentItemCt
			if m.treeMode {
				item := m.visibleItems[fileIdx]
				if item.isDir {
					line = m.renderDirItem(item, idx == m.cursor)
				} else {
					line = m.renderTreeFileItem(item, idx == m.cursor)
				}
			} else {
				line = m.renderFileItem(m.files[fileIdx], idx == m.cursor)
			}
		}

		b.WriteString(line)
		b.WriteString("\n")
		linesUsed++
	}

	// If no content items, show the Files header at the top
	if contentItemCt == 0 {
		var header strings.Builder
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
		headerStr := fmt.Sprintf(" Files%s  %d / %d", modeIndicator, reviewedCount, fileCount)
		header.WriteString(sectionStyle.Render(headerStr))
		header.WriteString("\n")
		header.WriteString(b.String())
		return header.String()
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

	leftPad := " "
	if selected {
		leftPad = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Render("▎")
	}

	right := " " + reviewChar + " "
	prefix := fmt.Sprintf("%s%s %s%s ", leftPad, styledStatus, recentChar, icon)
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

	leftPad := " "
	if selected {
		leftPad = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Render("▎")
	}

	styledArrow := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(arrow)
	styledFolder := lipgloss.NewStyle().Foreground(lipgloss.Color(folderColor)).Render(folderGlyph)
	prefix := fmt.Sprintf("%s%s%s %s ", leftPad, indent, styledArrow, styledFolder)
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

	leftPad := " "
	if selected {
		leftPad = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Render("▎")
	}

	styledStatus := lipgloss.NewStyle().Foreground(lipgloss.Color(statusColor)).Bold(true).Render(statusChar)
	right := " " + reviewChar + " "
	prefix := fmt.Sprintf("%s%s%s %s%s ", leftPad, indent, styledStatus, recentChar, icon)
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

	// Build icon path from content type (e.g. "md" → "content.md")
	iconPath := item.Title
	if item.ContentType != "" {
		ext := item.ContentType
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		iconPath = "content" + ext
	}
	icon := fileIcon(iconPath)
	glyph := iconLookup(iconPath).glyph
	const iconSlack = 2

	if selected && m.focused {
		plainReview := reviewChar
		if item.Reviewed {
			plainReview = "+"
		}
		right := " " + plainReview + " "
		prefix := fmt.Sprintf("  %s ", glyph)
		nameW := m.width - lipgloss.Width(prefix) - lipgloss.Width(right) - iconSlack
		if nameW < 1 {
			nameW = 1
		}
		name := truncatePath(item.Title, nameW)
		line := fmt.Sprintf("%s%-*s%s", prefix, nameW, name, right)
		return lipgloss.NewStyle().Reverse(true).Render(line)
	}

	leftPad := " "
	if selected {
		leftPad = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Render("▎")
	}
	right := " " + reviewChar + " "
	prefix := fmt.Sprintf("%s %s ", leftPad, icon)
	nameW := m.width - lipgloss.Width(prefix) - lipgloss.Width(right) - iconSlack
	if nameW < 1 {
		nameW = 1
	}
	name := truncatePath(item.Title, nameW)
	return fmt.Sprintf("%s%-*s%s", prefix, nameW, name, right)
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
	contentCount := len(m.contentItems)

	// Content items come first
	if m.cursor < contentCount {
		item := m.contentItems[m.cursor]
		return func() tea.Msg {
			return sidebarSelectMsg{isContent: true, contentID: item.ID}
		}
	}

	// Then file items
	fileIdx := m.cursor - contentCount
	if fileIdx < m.fileItemCount() {
		if m.treeMode {
			item := m.visibleItems[fileIdx]
			if item.isDir {
				return nil // Don't send selection for directories
			}
			path := item.node.File.Path
			return func() tea.Msg {
				return sidebarSelectMsg{path: path}
			}
		}
		path := m.files[fileIdx].Path
		return func() tea.Msg {
			return sidebarSelectMsg{path: path}
		}
	}
	return nil
}

// selectedFile returns the ChangedFile at the current cursor position,
// or nil if the cursor is on a directory or content item.
func (m sidebarModel) selectedFile() *types.ChangedFile {
	contentCount := len(m.contentItems)
	if m.cursor < contentCount {
		return nil // content item, not a file
	}
	fileIdx := m.cursor - contentCount
	if fileIdx >= m.fileItemCount() {
		return nil
	}
	if m.treeMode {
		item := m.visibleItems[fileIdx]
		if item.isDir {
			return nil
		}
		return item.node.File
	}
	return &m.files[fileIdx]
}

// navigateFile moves the cursor to the next (dir=+1) or previous (dir=-1)
// file, skipping directory nodes in tree mode. Returns a selectCurrent()
// command if a file was found, or nil if navigation is not possible.
func (m *sidebarModel) navigateFile(dir int) tea.Cmd {
	total := m.totalItems()
	if total == 0 {
		return nil
	}
	contentCount := len(m.contentItems)
	next := m.cursor + dir
	for next >= 0 && next < total {
		// Skip directory nodes in tree mode (file items start at contentCount)
		fileIdx := next - contentCount
		if m.treeMode && fileIdx >= 0 && fileIdx < m.fileItemCount() {
			if m.visibleItems[fileIdx].isDir {
				next += dir
				continue
			}
		}
		break
	}
	if next < 0 || next >= total {
		return nil
	}
	m.cursor = next
	m.ensureVisible()
	return m.selectCurrent()
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
	contentCount := len(m.contentItems)
	if m.treeMode {
		for i, item := range m.visibleItems {
			if !item.isDir && item.node.File != nil && item.node.File.Path == path {
				m.cursor = i + contentCount
				return
			}
		}
	} else {
		for i, f := range m.files {
			if f.Path == path {
				m.cursor = i + contentCount
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
// Accounts for section headers and dividers that consume vertical space.
func (m sidebarModel) viewportHeight() int {
	headerLines := 0
	if len(m.contentItems) > 0 {
		headerLines += 1 // "Review Items" header
		headerLines += 1 // blank line between sections
		headerLines += 1 // "Files" header
	} else {
		headerLines += 1 // "Files" header only
	}
	h := m.height - headerLines
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

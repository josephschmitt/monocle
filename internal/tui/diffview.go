package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/anthropics/monocle/internal/types"
)

type diffStyle int

const (
	diffStyleUnified diffStyle = iota
	diffStyleSplit
)

// diffViewLine represents a rendered line in the diff view.
type diffViewLine struct {
	kind       types.DiffLineKind
	oldLineNum int
	newLineNum int
	content    string
	isHunk     bool
	hunkHeader string
	isComment  bool
	comment    *types.ReviewComment

	// Split diff: right side
	isSplit      bool
	rightKind    types.DiffLineKind
	rightLineNum int
	rightContent string
	rightEmpty   bool // true if this side is a blank filler
	leftEmpty    bool
}

type diffViewModel struct {
	path      string
	hunks     []types.DiffHunk
	comments  []types.ReviewComment
	lines     []diffViewLine
	cursor    int
	offset    int // scroll offset
	width     int
	height    int
	focused   bool
	style     diffStyle

	// Visual mode
	visualMode  bool
	visualStart int

	// Content view mode (for plans/docs)
	contentMode bool
	contentID   string
	contentTitle string
}

func newDiffViewModel() diffViewModel {
	return diffViewModel{}
}

type loadDiffMsg struct {
	path     string
	result   *types.DiffResult
	comments []types.ReviewComment
}

func (m diffViewModel) Init() tea.Cmd {
	return nil
}

func (m diffViewModel) Update(msg tea.Msg) (diffViewModel, tea.Cmd) {
	switch msg := msg.(type) {
	case loadDiffMsg:
		m.contentMode = false
		m.contentID = ""
		m.contentTitle = ""
		sameFile := msg.path == m.path
		if msg.result != nil {
			m.hunks = msg.result.Hunks
		} else {
			m.hunks = nil
		}
		m.path = msg.path
		m.comments = msg.comments
		prevCursor := m.cursor
		prevOffset := m.offset
		m.buildLines()
		if sameFile && prevCursor < len(m.lines) {
			m.cursor = prevCursor
			m.offset = prevOffset
		} else {
			m.cursor = m.nearestSelectable(0, 1)
			m.offset = 0
			m.visualMode = false
		}
		return m, nil

	case loadContentMsg:
		m.contentMode = true
		m.contentID = msg.id
		m.contentTitle = msg.title
		m.path = msg.id
		m.hunks = nil
		m.comments = nil
		m.buildContentLines(msg.content)
		m.cursor = m.nearestSelectable(0, 1)
		m.offset = 0
		m.visualMode = false
		return m, nil

	case tea.KeyPressMsg:
		if !m.focused {
			return m, nil
		}
		switch msg.String() {
		case "j", "down":
			m.cursor = m.nextSelectable(m.cursor, 1)
			m.ensureVisible()
		case "k", "up":
			m.cursor = m.nextSelectable(m.cursor, -1)
			m.ensureVisible()
		case "ctrl+d":
			m.cursor += m.height / 2
			if m.cursor >= len(m.lines) {
				m.cursor = len(m.lines) - 1
			}
			m.cursor = m.nearestSelectable(m.cursor, 1)
			m.ensureVisible()
		case "ctrl+u":
			m.cursor -= m.height / 2
			if m.cursor < 0 {
				m.cursor = 0
			}
			m.cursor = m.nearestSelectable(m.cursor, -1)
			m.ensureVisible()
		case "g":
			m.cursor = m.nearestSelectable(0, 1)
			m.ensureVisible()
		case "G":
			if len(m.lines) > 0 {
				m.cursor = m.nearestSelectable(len(m.lines)-1, -1)
			}
			m.ensureVisible()
		case "v":
			if !m.visualMode {
				m.visualMode = true
				m.visualStart = m.cursor
			} else {
				m.visualMode = false
			}
		case "esc":
			m.visualMode = false
		case "t":
			// Toggle diff style
			if m.style == diffStyleUnified {
				m.style = diffStyleSplit
			} else {
				m.style = diffStyleUnified
			}
			m.buildLines()
		case "c":
			// Open comment editor
			if m.contentMode {
				// Content mode: comment on the content item
				if m.visualMode {
					start, end := m.visualRange()
					return m, openCommentCmd(m.contentID, start, end)
				}
				line := m.currentDiffLine()
				if line > 0 {
					return m, openCommentCmd(m.contentID, line, line)
				}
			} else {
				if m.visualMode {
					start, end := m.visualRange()
					return m, openCommentCmd(m.path, start, end)
				}
				line := m.currentDiffLine()
				if line > 0 {
					return m, openCommentCmd(m.path, line, line)
				}
			}
		}
	}
	return m, nil
}

func (m diffViewModel) View() string {
	if m.width == 0 || len(m.lines) == 0 {
		if m.path == "" {
			return centerText("Select a file to view diff", m.width, m.height)
		}
		if m.contentMode {
			return centerText("Empty content", m.width, m.height)
		}
		return centerText("No changes", m.width, m.height)
	}

	var b strings.Builder

	visibleLines := m.height
	end := m.offset + visibleLines
	if end > len(m.lines) {
		end = len(m.lines)
	}

	for i := m.offset; i < end; i++ {
		line := m.lines[i]
		selected := i == m.cursor
		inVisual := m.visualMode && m.inVisualRange(i)

		var rendered string
		if line.isHunk {
			rendered = m.renderHunkHeader(line, selected)
		} else if line.isComment {
			rendered = m.renderCommentLine(line, selected)
		} else if line.isSplit {
			rendered = m.renderSplitLine(line, selected, inVisual)
		} else {
			gutterWidth := 10
			contentWidth := m.width - gutterWidth
			rendered = m.renderDiffLine(line, gutterWidth, contentWidth, selected, inVisual)
		}

		b.WriteString(rendered)
		if i < end-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

func (m *diffViewModel) buildLines() {
	m.lines = nil

	if m.style == diffStyleSplit {
		m.buildSplitLines()
		return
	}

	for _, hunk := range m.hunks {
		// Hunk header
		m.lines = append(m.lines, diffViewLine{
			isHunk:     true,
			hunkHeader: hunk.Header,
			content:    fmt.Sprintf("@@ -%d,%d +%d,%d @@ %s", hunk.OldStart, hunk.OldCount, hunk.NewStart, hunk.NewCount, hunk.Header),
		})

		// Diff lines
		for _, dl := range hunk.Lines {
			m.lines = append(m.lines, diffViewLine{
				kind:       dl.Kind,
				oldLineNum: dl.OldLineNum,
				newLineNum: dl.NewLineNum,
				content:    dl.Content,
			})
		}

		// Inline comments for this hunk
		for _, c := range m.comments {
			if c.TargetRef == m.path && c.LineStart >= hunk.NewStart && c.LineStart <= hunk.NewStart+hunk.NewCount {
				m.lines = append(m.lines, diffViewLine{
					isComment: true,
					comment:   &c,
					content:   formatInlineComment(&c),
				})
			}
		}
	}
}

// buildContentLines builds lines for a content item (plan/doc) displayed as a document.
func (m *diffViewModel) buildContentLines(content string) {
	m.lines = nil
	rawLines := strings.Split(content, "\n")
	for i, line := range rawLines {
		m.lines = append(m.lines, diffViewLine{
			kind:       types.DiffLineContext,
			newLineNum: i + 1,
			content:    line,
		})
	}
}

func (m *diffViewModel) buildSplitLines() {
	for _, hunk := range m.hunks {
		m.lines = append(m.lines, diffViewLine{
			isHunk:     true,
			hunkHeader: hunk.Header,
			content:    fmt.Sprintf("@@ -%d,%d +%d,%d @@ %s", hunk.OldStart, hunk.OldCount, hunk.NewStart, hunk.NewCount, hunk.Header),
		})

		// Collect removed and added runs, pair them up
		var removed, added []types.DiffLine
		flushPairs := func() {
			maxLen := len(removed)
			if len(added) > maxLen {
				maxLen = len(added)
			}
			for i := 0; i < maxLen; i++ {
				sl := diffViewLine{isSplit: true}
				if i < len(removed) {
					sl.kind = types.DiffLineRemoved
					sl.oldLineNum = removed[i].OldLineNum
					sl.content = removed[i].Content
				} else {
					sl.leftEmpty = true
					sl.kind = types.DiffLineContext
				}
				if i < len(added) {
					sl.rightKind = types.DiffLineAdded
					sl.rightLineNum = added[i].NewLineNum
					sl.rightContent = added[i].Content
				} else {
					sl.rightEmpty = true
					sl.rightKind = types.DiffLineContext
				}
				m.lines = append(m.lines, sl)
			}
			removed = removed[:0]
			added = added[:0]
		}

		for _, dl := range hunk.Lines {
			switch dl.Kind {
			case types.DiffLineRemoved:
				removed = append(removed, dl)
			case types.DiffLineAdded:
				added = append(added, dl)
			case types.DiffLineContext:
				flushPairs()
				m.lines = append(m.lines, diffViewLine{
					isSplit:      true,
					kind:         types.DiffLineContext,
					oldLineNum:   dl.OldLineNum,
					content:      dl.Content,
					rightKind:    types.DiffLineContext,
					rightLineNum: dl.NewLineNum,
					rightContent: dl.Content,
				})
			}
		}
		flushPairs()

		// Inline comments
		for _, c := range m.comments {
			if c.TargetRef == m.path && c.LineStart >= hunk.NewStart && c.LineStart <= hunk.NewStart+hunk.NewCount {
				m.lines = append(m.lines, diffViewLine{
					isComment: true,
					comment:   &c,
					content:   formatInlineComment(&c),
				})
			}
		}
	}
}

func (m diffViewModel) renderHunkHeader(line diffViewLine, selected bool) string {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Faint(true)
	content := style.Render(line.content)
	if selected && m.focused {
		content = lipgloss.NewStyle().Reverse(true).Render(line.content)
	}
	return fmt.Sprintf("%-*s", m.width, content)
}

func (m diffViewModel) renderCommentLine(line diffViewLine, selected bool) string {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	if selected {
		style = style.Reverse(true)
	}
	return style.Render(fmt.Sprintf("%-*s", m.width, line.content))
}

func (m diffViewModel) renderDiffLine(line diffViewLine, gutterWidth, contentWidth int, selected, inVisual bool) string {
	// Gutter
	var gutter string
	switch line.kind {
	case types.DiffLineContext:
		gutter = fmt.Sprintf("%4d %4d ", line.oldLineNum, line.newLineNum)
	case types.DiffLineAdded:
		gutter = fmt.Sprintf("     %4d ", line.newLineNum)
	case types.DiffLineRemoved:
		gutter = fmt.Sprintf("%4d      ", line.oldLineNum)
	}

	// Content
	content := line.content
	if len(content) > contentWidth {
		content = content[:contentWidth-1] + "…"
	}

	full := gutter + fmt.Sprintf("%-*s", contentWidth, content)

	if (selected || inVisual) && m.focused {
		return lipgloss.NewStyle().Reverse(true).Render(full)
	}

	// Color by diff type
	switch line.kind {
	case types.DiffLineAdded:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render(full)
	case types.DiffLineRemoved:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render(full)
	default:
		return full
	}
}

func (m diffViewModel) renderSplitLine(line diffViewLine, selected, inVisual bool) string {
	halfW := m.width / 2
	gutterW := 5 // "NNNN "
	contentW := halfW - gutterW - 1 // -1 for divider
	if contentW < 1 {
		contentW = 1
	}

	// Left side
	var leftGutter, leftContent string
	if line.leftEmpty {
		leftGutter = strings.Repeat(" ", gutterW)
		leftContent = strings.Repeat(" ", contentW)
	} else {
		if line.oldLineNum > 0 {
			leftGutter = fmt.Sprintf("%4d ", line.oldLineNum)
		} else {
			leftGutter = strings.Repeat(" ", gutterW)
		}
		lc := line.content
		if len(lc) > contentW {
			lc = lc[:contentW-1] + "…"
		}
		leftContent = fmt.Sprintf("%-*s", contentW, lc)
	}

	// Right side
	var rightGutter, rightContent string
	if line.rightEmpty {
		rightGutter = strings.Repeat(" ", gutterW)
		rightContent = strings.Repeat(" ", contentW)
	} else {
		if line.rightLineNum > 0 {
			rightGutter = fmt.Sprintf("%4d ", line.rightLineNum)
		} else {
			rightGutter = strings.Repeat(" ", gutterW)
		}
		rc := line.rightContent
		if len(rc) > contentW {
			rc = rc[:contentW-1] + "…"
		}
		rightContent = fmt.Sprintf("%-*s", contentW, rc)
	}

	leftFull := leftGutter + leftContent
	rightFull := rightGutter + rightContent
	divider := "│"

	if (selected || inVisual) && m.focused {
		return lipgloss.NewStyle().Reverse(true).Render(leftFull + divider + rightFull)
	}

	// Color each side
	leftStyled := m.colorSide(leftFull, line.kind, line.leftEmpty)
	rightStyled := m.colorSide(rightFull, line.rightKind, line.rightEmpty)
	divStyled := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(divider)

	return leftStyled + divStyled + rightStyled
}

func (m diffViewModel) colorSide(text string, kind types.DiffLineKind, empty bool) string {
	if empty {
		return lipgloss.NewStyle().Faint(true).Render(text)
	}
	switch kind {
	case types.DiffLineAdded:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render(text)
	case types.DiffLineRemoved:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render(text)
	default:
		return text
	}
}

func (m *diffViewModel) ensureVisible() {
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+m.height {
		m.offset = m.cursor - m.height + 1
	}
}

// isSelectable returns true if the line at idx is a diff content line (not a hunk header or comment).
func (m diffViewModel) isSelectable(idx int) bool {
	if idx < 0 || idx >= len(m.lines) {
		return false
	}
	line := m.lines[idx]
	return !line.isHunk && !line.isComment
}

// nextSelectable moves from current position by dir (+1 or -1), skipping non-selectable lines.
func (m diffViewModel) nextSelectable(from, dir int) int {
	next := from + dir
	for next >= 0 && next < len(m.lines) && !m.isSelectable(next) {
		next += dir
	}
	if next < 0 || next >= len(m.lines) {
		return from // stay put if nothing selectable in that direction
	}
	return next
}

// nearestSelectable finds the closest selectable line from pos, preferring the given direction.
func (m diffViewModel) nearestSelectable(pos, dir int) int {
	if pos < 0 {
		pos = 0
	}
	if pos >= len(m.lines) {
		pos = len(m.lines) - 1
	}
	if m.isSelectable(pos) {
		return pos
	}
	return m.nextSelectable(pos, dir)
}

func (m diffViewModel) visualRange() (int, int) {
	start := m.visualStart
	end := m.cursor
	if start > end {
		start, end = end, start
	}
	// Map to line numbers
	startLine := m.lineNumAt(start)
	endLine := m.lineNumAt(end)
	if startLine == 0 {
		startLine = endLine
	}
	if endLine == 0 {
		endLine = startLine
	}
	return startLine, endLine
}

func (m diffViewModel) inVisualRange(idx int) bool {
	if !m.visualMode {
		return false
	}
	start := m.visualStart
	end := m.cursor
	if start > end {
		start, end = end, start
	}
	return idx >= start && idx <= end
}

func (m diffViewModel) lineNumAt(idx int) int {
	if idx < 0 || idx >= len(m.lines) {
		return 0
	}
	line := m.lines[idx]
	if line.newLineNum > 0 {
		return line.newLineNum
	}
	return line.oldLineNum
}

func (m diffViewModel) currentDiffLine() int {
	return m.lineNumAt(m.cursor)
}

type openCommentMsg struct {
	path      string
	lineStart int
	lineEnd   int
}

func openCommentCmd(path string, start, end int) tea.Cmd {
	return func() tea.Msg {
		return openCommentMsg{path: path, lineStart: start, lineEnd: end}
	}
}

func formatInlineComment(c *types.ReviewComment) string {
	typeLabel := strings.ToUpper(string(c.Type))
	prefix := "│"
	if c.Outdated {
		prefix = "│ ⚠"
	}
	body := c.Body
	if len(body) > 60 {
		body = body[:57] + "..."
	}
	return fmt.Sprintf("  ┌─── %s %s", typeLabel, strings.Repeat("─", 20)) + "\n" +
		fmt.Sprintf("  %s %s", prefix, body) + "\n" +
		fmt.Sprintf("  └───%s", strings.Repeat("─", 25))
}

func centerText(text string, width, height int) string {
	if width == 0 || height == 0 {
		return text
	}
	var b strings.Builder
	padTop := height / 3
	for i := 0; i < padTop; i++ {
		b.WriteString("\n")
	}
	padLeft := (width - len(text)) / 2
	if padLeft < 0 {
		padLeft = 0
	}
	b.WriteString(strings.Repeat(" ", padLeft) + text)
	return b.String()
}

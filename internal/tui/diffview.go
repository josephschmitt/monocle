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
		m.path = msg.path
		if msg.result != nil {
			m.hunks = msg.result.Hunks
		} else {
			m.hunks = nil
		}
		m.comments = msg.comments
		m.buildLines()
		m.cursor = 0
		m.offset = 0
		m.visualMode = false
		return m, nil

	case tea.KeyPressMsg:
		if !m.focused {
			return m, nil
		}
		switch msg.String() {
		case "j", "down":
			if m.cursor < len(m.lines)-1 {
				m.cursor++
				m.ensureVisible()
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
				m.ensureVisible()
			}
		case "ctrl+d":
			m.cursor += m.height / 2
			if m.cursor >= len(m.lines) {
				m.cursor = len(m.lines) - 1
			}
			m.ensureVisible()
		case "ctrl+u":
			m.cursor -= m.height / 2
			if m.cursor < 0 {
				m.cursor = 0
			}
			m.ensureVisible()
		case "g":
			m.cursor = 0
			m.ensureVisible()
		case "G":
			if len(m.lines) > 0 {
				m.cursor = len(m.lines) - 1
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
	return m, nil
}

func (m diffViewModel) View() string {
	if m.width == 0 || len(m.lines) == 0 {
		if m.path == "" {
			return centerText("Select a file to view diff", m.width, m.height)
		}
		return centerText("No changes", m.width, m.height)
	}

	var b strings.Builder
	gutterWidth := 10 // "NNNN NNNN "
	contentWidth := m.width - gutterWidth

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
		} else {
			rendered = m.renderDiffLine(line, gutterWidth, contentWidth, selected, inVisual)
		}

		b.WriteString(rendered)
		if i < end-1 {
			b.WriteString("\n")
		}
	}

	// Pad remaining height
	rendered := end - m.offset
	for i := rendered; i < visibleLines; i++ {
		b.WriteString("\n")
	}

	return b.String()
}

func (m *diffViewModel) buildLines() {
	m.lines = nil

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

func (m *diffViewModel) ensureVisible() {
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+m.height {
		m.offset = m.cursor - m.height + 1
	}
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

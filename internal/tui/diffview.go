package tui

import (
	"fmt"
	"image/color"
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

	// Paired line content for intra-line diff highlighting (unified mode)
	pairContent string

	// Markdown rendering state
	mdInCodeBlock bool
	mdIsFence     bool
	mdCodeLang    string

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
	theme     *Theme
	hl        *highlighter

	hOffset int  // horizontal scroll offset (runes)
	wrap    bool // soft-wrap long lines
	tabSize int  // spaces per tab character

	// Visual mode
	visualMode  bool
	visualStart int

	// Content view mode (for plans/docs)
	contentMode bool
	contentID   string
	contentTitle string
	mdStyler    *markdownStyler
}

func newDiffViewModel(theme *Theme) diffViewModel {
	return diffViewModel{
		theme:    theme,
		hl:       newHighlighter(),
		mdStyler: newMarkdownStyler(*theme),
	}
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
			m.hOffset = 0
			m.visualMode = false
		}
		return m, nil

	case loadContentMsg:
		isReload := m.contentMode && m.contentID == msg.id
		m.contentMode = true
		m.contentID = msg.id
		m.contentTitle = msg.title
		if msg.contentType != "" {
			ext := msg.contentType
			if !strings.HasPrefix(ext, ".") {
				ext = "." + ext
			}
			m.path = "content" + ext
		} else {
			m.path = msg.id
		}
		m.hunks = nil
		m.comments = msg.comments
		prevCursor := m.cursor
		prevOffset := m.offset
		m.buildContentLines(msg.content)
		if isReload && prevCursor < len(m.lines) {
			m.cursor = m.nearestSelectable(prevCursor, 1)
			m.offset = prevOffset
		} else {
			m.cursor = m.nearestSelectable(0, 1)
			m.offset = 0
		}
		m.hOffset = 0
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
		case "h", "left":
			m.ScrollLeft()
		case "l", "right":
			m.ScrollRight()
		case "0":
			m.hOffset = 0
		case "w":
			m.wrap = !m.wrap
			if m.wrap {
				m.hOffset = 0
			}
			m.ensureVisible()
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
					return m, openCommentCmd(m.contentID, start, end, types.TargetContent)
				}
				line := m.currentDiffLine()
				if line > 0 {
					return m, openCommentCmd(m.contentID, line, line, types.TargetContent)
				}
			} else {
				if m.visualMode {
					start, end := m.visualRange()
					return m, openCommentCmd(m.path, start, end, types.TargetFile)
				}
				line := m.currentDiffLine()
				if line > 0 {
					return m, openCommentCmd(m.path, line, line, types.TargetFile)
				}
			}
		case "C":
			// File-level comment
			if m.contentMode {
				return m, openFileCommentCmd(m.contentID, types.TargetContent)
			}
			if m.path != "" {
				return m, openFileCommentCmd(m.path, types.TargetFile)
			}
		}
	}
	return m, nil
}

func (m diffViewModel) View() string {
	if m.width == 0 || len(m.lines) == 0 {
		if m.path == "" {
			return renderSplash(m.width, m.height)
		}
		if m.contentMode {
			return centerBlock([]string{"Empty content"}, m.width, m.height)
		}
		return centerBlock([]string{"No changes"}, m.width, m.height)
	}

	var b strings.Builder
	screenUsed := 0

	for i := m.offset; i < len(m.lines) && screenUsed < m.height; i++ {
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
		} else if m.contentMode {
			gutterWidth := 4
			contentWidth := m.width - gutterWidth
			rendered = m.renderContentLine(line, gutterWidth, contentWidth, selected, inVisual)
		} else {
			gutterWidth := 10
			contentWidth := m.width - gutterWidth
			rendered = m.renderDiffLine(line, gutterWidth, contentWidth, selected, inVisual)
		}

		// rendered may contain multiple lines in wrap mode
		renderedLines := strings.Split(rendered, "\n")
		for _, rl := range renderedLines {
			if screenUsed >= m.height {
				break
			}
			if screenUsed > 0 {
				b.WriteString("\n")
			}
			b.WriteString(rl)
			screenUsed++
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

	// File-level comments (LineStart == 0) rendered before hunks
	for i := range m.comments {
		c := &m.comments[i]
		if c.TargetRef == m.path && c.LineStart == 0 {
			m.lines = append(m.lines, diffViewLine{
				isComment: true,
				comment:   c,
				content:   formatInlineComment(c),
			})
		}
	}

	for _, hunk := range m.hunks {
		// Hunk header
		m.lines = append(m.lines, diffViewLine{
			isHunk:     true,
			hunkHeader: hunk.Header,
			content:    fmt.Sprintf("@@ -%d,%d +%d,%d @@ %s", hunk.OldStart, hunk.OldCount, hunk.NewStart, hunk.NewCount, hunk.Header),
		})

		// Diff lines with inline comments inserted after target line
		for _, dl := range hunk.Lines {
			m.lines = append(m.lines, diffViewLine{
				kind:       dl.Kind,
				oldLineNum: dl.OldLineNum,
				newLineNum: dl.NewLineNum,
				content:    m.expandTabs(dl.Content),
			})

			// Insert comments after their last targeted line
			for i := range m.comments {
				c := &m.comments[i]
				anchor := c.LineEnd
				if anchor == 0 {
					anchor = c.LineStart
				}
				if c.TargetRef == m.path && anchor == dl.NewLineNum && dl.NewLineNum > 0 {
					m.lines = append(m.lines, diffViewLine{
						isComment: true,
						comment:   c,
						content:   formatInlineComment(c),
					})
				}
			}
		}
	}

	m.pairLines()
}

// buildContentLines builds lines for a content item (plan/doc) displayed as a document.
func (m *diffViewModel) buildContentLines(content string) {
	m.lines = nil

	// File-level comments (LineStart == 0) rendered before content
	for i := range m.comments {
		c := &m.comments[i]
		if c.TargetRef == m.contentID && c.LineStart == 0 {
			m.lines = append(m.lines, diffViewLine{
				isComment: true,
				comment:   c,
				content:   formatInlineComment(c),
			})
		}
	}

	isMd := isMarkdownContent(m.path)
	inCodeBlock := false
	codeLang := ""
	rawLines := strings.Split(content, "\n")
	for i, line := range rawLines {
		lineNum := i + 1

		// Track code fence state for markdown rendering
		isFence := false
		if isMd {
			if fence := codeFencePattern.FindStringSubmatch(line); fence != nil {
				isFence = true
				if !inCodeBlock {
					inCodeBlock = true
					codeLang = fence[1]
				} else {
					inCodeBlock = false
					codeLang = ""
				}
			}
		}

		m.lines = append(m.lines, diffViewLine{
			kind:          types.DiffLineContext,
			newLineNum:    lineNum,
			content:       m.expandTabs(line),
			mdInCodeBlock: inCodeBlock && isMd && !isFence,
			mdIsFence:     isFence,
			mdCodeLang:    codeLang,
		})

		// Insert comments after their last targeted line
		for j := range m.comments {
			c := &m.comments[j]
			anchor := c.LineEnd
			if anchor == 0 {
				anchor = c.LineStart
			}
			if c.TargetRef == m.contentID && anchor == lineNum {
				m.lines = append(m.lines, diffViewLine{
					isComment: true,
					comment:   c,
					content:   formatInlineComment(c),
				})
			}
		}
	}
}

func (m *diffViewModel) buildSplitLines() {
	// File-level comments (LineStart == 0) rendered before hunks
	for i := range m.comments {
		c := &m.comments[i]
		if c.TargetRef == m.path && c.LineStart == 0 {
			m.lines = append(m.lines, diffViewLine{
				isComment: true,
				comment:   c,
				content:   formatInlineComment(c),
			})
		}
	}

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
					sl.content = m.expandTabs(removed[i].Content)
				} else {
					sl.leftEmpty = true
					sl.kind = types.DiffLineContext
				}
				if i < len(added) {
					sl.rightKind = types.DiffLineAdded
					sl.rightLineNum = added[i].NewLineNum
					sl.rightContent = m.expandTabs(added[i].Content)
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
				expanded := m.expandTabs(dl.Content)
				m.lines = append(m.lines, diffViewLine{
					isSplit:      true,
					kind:         types.DiffLineContext,
					oldLineNum:   dl.OldLineNum,
					content:      expanded,
					rightKind:    types.DiffLineContext,
					rightLineNum: dl.NewLineNum,
					rightContent: expanded,
				})
			}
		}
		flushPairs()

		// Insert inline comments after their target lines
		m.insertInlineComments(hunk)
	}
}

// pairLines pairs consecutive removed/added line runs for intra-line diff highlighting.
func (m *diffViewModel) pairLines() {
	i := 0
	for i < len(m.lines) {
		if m.lines[i].isHunk || m.lines[i].isComment {
			i++
			continue
		}

		// Find run of removed lines
		removeStart := i
		for i < len(m.lines) && m.lines[i].kind == types.DiffLineRemoved &&
			!m.lines[i].isHunk && !m.lines[i].isComment {
			i++
		}
		removeEnd := i

		// Find run of added lines immediately after
		addStart := i
		for i < len(m.lines) && m.lines[i].kind == types.DiffLineAdded &&
			!m.lines[i].isHunk && !m.lines[i].isComment {
			i++
		}
		addEnd := i

		// Pair them up
		removeCount := removeEnd - removeStart
		addCount := addEnd - addStart
		pairCount := removeCount
		if addCount < pairCount {
			pairCount = addCount
		}
		for j := 0; j < pairCount; j++ {
			m.lines[removeStart+j].pairContent = m.lines[addStart+j].content
			m.lines[addStart+j].pairContent = m.lines[removeStart+j].content
		}

		// If we didn't advance past any removed/added, skip forward
		if removeStart == removeEnd && addStart == addEnd {
			i++
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
	// Pick color based on comment type
	var clr color.Color
	if line.comment != nil {
		switch line.comment.Type {
		case types.CommentIssue:
			clr = lipgloss.Color("1")
		case types.CommentSuggestion:
			clr = lipgloss.Color("3")
		case types.CommentNote:
			clr = lipgloss.Color("4")
		case types.CommentPraise:
			clr = lipgloss.Color("2")
		default:
			clr = lipgloss.Color("3")
		}
	} else {
		clr = lipgloss.Color("3")
	}

	style := lipgloss.NewStyle().Foreground(clr)
	if selected {
		style = style.Reverse(true)
	}

	// Render each sub-line individually to preserve multi-line box structure
	subLines := strings.Split(line.content, "\n")
	var b strings.Builder
	for i, sl := range subLines {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(style.Render(fmt.Sprintf("%-*s", m.width, sl)))
	}
	return b.String()
}

func (m diffViewModel) renderContentLine(line diffViewLine, _, contentWidth int, selected, inVisual bool) string {
	gutterWidth := 4
	gutter := fmt.Sprintf("%-3d ", line.newLineNum)
	isMd := m.contentMode && isMarkdownContent(m.path)

	// Wrap mode
	if m.wrap {
		return m.renderWrappedLine(gutter, line.content, gutterWidth, contentWidth,
			nil, nil, selected || inVisual)
	}

	// Scroll mode: apply horizontal offset, then clip
	content := line.content
	if m.hOffset > 0 {
		content, _ = applyHOffset(content, m.hOffset)
	}
	contentRunes := []rune(content)
	if len(contentRunes) > contentWidth {
		content = string(contentRunes[:contentWidth])
	}

	if (selected || inVisual) && m.focused {
		padded := gutter + padToWidth(content, contentWidth)
		return lipgloss.NewStyle().Reverse(true).Render(padded)
	}

	// Render gutter
	gutterStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	if len(gutter) < gutterWidth {
		gutter = fmt.Sprintf("%-*s", gutterWidth, gutter)
	}
	renderedGutter := gutterStyle.Render(gutter)

	// Render content: markdown styling or syntax highlighting
	var renderedContent string
	if isMd && line.mdIsFence {
		// Code fence markers → render as horizontal rule
		renderedContent = m.mdStyler.theme.MarkdownRule.Render(strings.Repeat("─", min(40, contentWidth)))
		renderedContent = padToWidth(renderedContent, contentWidth)
	} else if isMd && line.mdInCodeBlock && line.mdCodeLang != "" {
		// Code block with language → use Chroma syntax highlighting
		fakePath := "code." + line.mdCodeLang
		renderedContent = m.hl.highlightLine(fakePath, content, nil, nil, nil, contentWidth)
	} else if isMd && line.mdInCodeBlock {
		// Code block without language → code block style
		renderedContent = m.mdStyler.theme.MarkdownCodeBlock.Render(content)
		renderedContent = padToWidth(renderedContent, contentWidth)
	} else if isMd {
		// Regular markdown line
		renderedContent = m.mdStyler.StyleLine(content)
		renderedContent = padToWidth(renderedContent, contentWidth)
	} else {
		renderedContent = m.hl.highlightLine(m.path, content, nil, nil, nil, contentWidth)
	}

	return renderedGutter + renderedContent
}

func (m diffViewModel) renderDiffLine(line diffViewLine, _, contentWidth int, selected, inVisual bool) string {
	gutterWidth := 10

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

	// Determine backgrounds
	var lineBg, changeBg color.Color
	switch line.kind {
	case types.DiffLineAdded:
		lineBg = m.theme.AddedBg
		changeBg = m.theme.AddedChangeBg
	case types.DiffLineRemoved:
		lineBg = m.theme.RemovedBg
		changeBg = m.theme.RemovedChangeBg
	}

	// Wrap mode: render line as multiple screen lines
	if m.wrap {
		return m.renderWrappedLine(gutter, line.content, gutterWidth, contentWidth,
			lineBg, changeBg, selected || inVisual)
	}

	// Scroll mode: apply horizontal offset, then clip
	content := line.content
	if m.hOffset > 0 {
		content, _ = applyHOffset(content, m.hOffset)
	}
	contentRunes := []rune(content)
	if len(contentRunes) > contentWidth {
		content = string(contentRunes[:contentWidth])
	}

	// Selected: reverse the full plain line
	if (selected || inVisual) && m.focused {
		padded := gutter + padToWidth(content, contentWidth)
		return lipgloss.NewStyle().Reverse(true).Render(padded)
	}

	// Render gutter
	gutterStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	if lineBg != nil {
		gutterStyle = gutterStyle.Background(lineBg)
	}
	if len(gutter) < gutterWidth {
		gutter = fmt.Sprintf("%-*s", gutterWidth, gutter)
	}
	renderedGutter := gutterStyle.Render(gutter)

	// Compute intra-line change ranges
	var changes []changeRange
	if line.pairContent != "" {
		if line.kind == types.DiffLineRemoved {
			changes, _ = computeChangeRanges(line.content, line.pairContent)
		} else if line.kind == types.DiffLineAdded {
			_, changes = computeChangeRanges(line.pairContent, line.content)
		}
		if m.hOffset > 0 {
			changes = shiftChangeRanges(changes, m.hOffset)
		}
		changes = clipChangeRanges(changes, contentWidth)
	}

	renderedContent := m.hl.highlightLine(m.path, content, lineBg, changeBg, changes, contentWidth)

	return renderedGutter + renderedContent
}

func (m diffViewModel) renderSplitLine(line diffViewLine, selected, inVisual bool) string {
	halfW := (m.width - 1) / 2 // subtract divider, then halve
	gutterW := 5               // "NNNN "
	contentW := halfW - gutterW
	if contentW < 1 {
		contentW = 1
	}

	// Prepare left side raw content
	var leftGutter, leftRawContent string
	leftTruncatedAt := -1
	if line.leftEmpty {
		leftGutter = strings.Repeat(" ", gutterW)
		leftRawContent = ""
	} else {
		if line.oldLineNum > 0 {
			leftGutter = fmt.Sprintf("%4d ", line.oldLineNum)
		} else {
			leftGutter = strings.Repeat(" ", gutterW)
		}
		leftRawContent = line.content
		if m.hOffset > 0 {
			leftRawContent, _ = applyHOffset(leftRawContent, m.hOffset)
		}
		leftRunes := []rune(leftRawContent)
		if len(leftRunes) > contentW {
			leftTruncatedAt = contentW
			leftRawContent = string(leftRunes[:contentW])
		}
	}

	// Prepare right side raw content
	var rightGutter, rightRawContent string
	rightTruncatedAt := -1
	if line.rightEmpty {
		rightGutter = strings.Repeat(" ", gutterW)
		rightRawContent = ""
	} else {
		if line.rightLineNum > 0 {
			rightGutter = fmt.Sprintf("%4d ", line.rightLineNum)
		} else {
			rightGutter = strings.Repeat(" ", gutterW)
		}
		rightRawContent = line.rightContent
		if m.hOffset > 0 {
			rightRawContent, _ = applyHOffset(rightRawContent, m.hOffset)
		}
		rightRunes := []rune(rightRawContent)
		if len(rightRunes) > contentW {
			rightTruncatedAt = contentW
			rightRawContent = string(rightRunes[:contentW])
		}
	}

	divider := "│"

	// Selected: reverse the full plain line
	if (selected || inVisual) && m.focused {
		leftFull := leftGutter + padToWidth(leftRawContent, contentW)
		rightFull := rightGutter + padToWidth(rightRawContent, contentW)
		return lipgloss.NewStyle().Reverse(true).Render(leftFull + divider + rightFull)
	}

	// Compute intra-line change ranges for paired sides
	var leftChanges, rightChanges []changeRange
	if !line.leftEmpty && !line.rightEmpty &&
		line.kind == types.DiffLineRemoved && line.rightKind == types.DiffLineAdded {
		leftChanges, rightChanges = computeChangeRanges(line.content, line.rightContent)
		if m.hOffset > 0 {
			leftChanges = shiftChangeRanges(leftChanges, m.hOffset)
			rightChanges = shiftChangeRanges(rightChanges, m.hOffset)
		}
		if leftTruncatedAt >= 0 {
			leftChanges = clipChangeRanges(leftChanges, leftTruncatedAt)
		}
		if rightTruncatedAt >= 0 {
			rightChanges = clipChangeRanges(rightChanges, rightTruncatedAt)
		}
	}

	// Render each side
	leftStyled := m.renderSplitSide(leftGutter, leftRawContent, line.kind, line.leftEmpty, leftChanges, gutterW, contentW)
	rightStyled := m.renderSplitSide(rightGutter, rightRawContent, line.rightKind, line.rightEmpty, rightChanges, gutterW, contentW)
	divStyled := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(divider)

	return leftStyled + divStyled + rightStyled
}

func (m diffViewModel) renderSplitSide(gutter, content string, kind types.DiffLineKind, empty bool, changes []changeRange, gutterW, contentW int) string {
	if empty {
		full := strings.Repeat(" ", gutterW) + strings.Repeat(" ", contentW)
		return lipgloss.NewStyle().Faint(true).Render(full)
	}

	// Determine backgrounds
	var lineBg, changeBg color.Color
	switch kind {
	case types.DiffLineAdded:
		lineBg = m.theme.AddedBg
		changeBg = m.theme.AddedChangeBg
	case types.DiffLineRemoved:
		lineBg = m.theme.RemovedBg
		changeBg = m.theme.RemovedChangeBg
	}

	// Render gutter
	gutterStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	if lineBg != nil {
		gutterStyle = gutterStyle.Background(lineBg)
	}
	if len(gutter) < gutterW {
		gutter = fmt.Sprintf("%-*s", gutterW, gutter)
	}
	renderedGutter := gutterStyle.Render(gutter)

	// Render content with syntax highlighting
	renderedContent := m.hl.highlightLine(m.path, content, lineBg, changeBg, changes, contentW)

	return renderedGutter + renderedContent
}

// padToWidth pads a string with spaces to reach the target visual width,
// using lipgloss.Width for correct measurement of multi-byte characters.
func padToWidth(s string, width int) string {
	visWidth := lipgloss.Width(s)
	if visWidth >= width {
		return s
	}
	return s + strings.Repeat(" ", width-visWidth)
}

// renderWrappedLine renders a single logical line wrapped across multiple screen lines.
// Used by both renderDiffLine and renderContentLine in wrap mode.
func (m diffViewModel) renderWrappedLine(gutter, content string, gutterWidth, contentWidth int,
	lineBg, changeBg color.Color, highlight bool) string {

	chunks := wrapContent(content, contentWidth)
	blankGutter := strings.Repeat(" ", gutterWidth)

	var parts []string
	for ci, chunk := range chunks {
		chunkGutter := gutter
		if ci > 0 {
			chunkGutter = blankGutter
		}

		if highlight && m.focused {
			full := chunkGutter + fmt.Sprintf("%-*s", contentWidth, chunk)
			parts = append(parts, lipgloss.NewStyle().Reverse(true).Render(full))
			continue
		}

		// Render gutter
		gutterStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
		if lineBg != nil {
			gutterStyle = gutterStyle.Background(lineBg)
		}
		if len(chunkGutter) < gutterWidth {
			chunkGutter = fmt.Sprintf("%-*s", gutterWidth, chunkGutter)
		}
		renderedGutter := gutterStyle.Render(chunkGutter)

		// Highlight each chunk independently (change ranges not mapped across chunks)
		renderedContent := m.hl.highlightLine(m.path, chunk, lineBg, changeBg, nil, contentWidth)

		parts = append(parts, renderedGutter+renderedContent)
	}
	return strings.Join(parts, "\n")
}

// ScrollRight scrolls the diff content right by a tab stop.
func (m *diffViewModel) ScrollRight() {
	if m.wrap {
		return
	}
	m.hOffset += 4
}

// ScrollLeft scrolls the diff content left by a tab stop.
func (m *diffViewModel) ScrollLeft() {
	if m.wrap {
		return
	}
	m.hOffset -= 4
	if m.hOffset < 0 {
		m.hOffset = 0
	}
}

// contentWidthFor returns the available content width (excluding gutter) for a line.
func (m diffViewModel) contentWidthFor(line diffViewLine) int {
	if line.isSplit {
		return (m.width-1)/2 - 5 // subtract divider, then halve, minus gutter
	}
	if m.contentMode {
		return m.width - 4 // gutterWidth=4
	}
	return m.width - 10 // gutterWidth=10
}

// screenLinesFor returns how many screen lines a logical line occupies.
// In non-wrap mode or for split/hunk/comment lines, this is always 1.
func (m diffViewModel) screenLinesFor(idx int) int {
	if !m.wrap {
		return 1
	}
	if idx < 0 || idx >= len(m.lines) {
		return 1
	}
	line := m.lines[idx]
	if line.isHunk || line.isComment || line.isSplit {
		return 1
	}
	cw := m.contentWidthFor(line)
	if cw <= 0 {
		return 1
	}
	return len(wrapContent(line.content, cw))
}

// applyHOffset slices content at the horizontal offset (rune-aware).
// Returns the sliced content and whether there is hidden content to the left.
func applyHOffset(content string, hOffset int) (string, bool) {
	if hOffset <= 0 {
		return content, false
	}
	runes := []rune(content)
	if hOffset >= len(runes) {
		return "", true
	}
	return string(runes[hOffset:]), true
}

// shiftChangeRanges adjusts byte-offset change ranges by a rune offset.
// This is approximate since rune offset != byte offset for multi-byte chars,
// but works correctly for ASCII content (the common case for code).
func shiftChangeRanges(changes []changeRange, runeOffset int) []changeRange {
	if runeOffset <= 0 || len(changes) == 0 {
		return changes
	}
	var result []changeRange
	for _, cr := range changes {
		shifted := changeRange{start: cr.start - runeOffset, end: cr.end - runeOffset}
		if shifted.end <= 0 {
			continue
		}
		if shifted.start < 0 {
			shifted.start = 0
		}
		result = append(result, shifted)
	}
	return result
}

// expandTabs replaces tab characters with spaces for consistent width calculation.
// Tabs are 1 rune but render as multiple visual columns in the terminal, which
// breaks rune-based width truncation in the diff view.
func (m *diffViewModel) expandTabs(s string) string {
	tabSize := m.tabSize
	if tabSize <= 0 {
		tabSize = 4
	}
	return strings.ReplaceAll(s, "\t", strings.Repeat(" ", tabSize))
}

// wrapContent splits content into lines that fit within width, preferring
// to break at word boundaries (spaces). Falls back to character-based
// wrapping when a single word exceeds the available width.
func wrapContent(content string, width int) []string {
	if width <= 0 {
		return []string{content}
	}
	runes := []rune(content)
	if len(runes) <= width {
		return []string{content}
	}

	var chunks []string
	lineStart := 0
	lastSpace := -1 // index of last space seen on the current line

	for i := 0; i < len(runes); i++ {
		if runes[i] == ' ' {
			lastSpace = i
		}

		lineLen := i - lineStart + 1
		if lineLen > width {
			if lastSpace > lineStart {
				// Break after the last space (space stays at end of current line)
				chunks = append(chunks, string(runes[lineStart:lastSpace+1]))
				lineStart = lastSpace + 1
				lastSpace = -1
			} else {
				// No space on this line — force break at width (character fallback)
				chunks = append(chunks, string(runes[lineStart:lineStart+width]))
				lineStart = lineStart + width
				lastSpace = -1
				i = lineStart - 1 // will be incremented by loop
			}
		}
	}

	// Emit remaining content
	if lineStart < len(runes) {
		chunks = append(chunks, string(runes[lineStart:]))
	}

	return chunks
}

// ScrollDown scrolls the diff viewport down by one line.
func (m *diffViewModel) ScrollDown() {
	// In wrap mode, compute max offset accounting for wrapped lines
	if m.wrap {
		// Check if there's content below to scroll to
		screenLines := 0
		for i := m.offset; i < len(m.lines); i++ {
			screenLines += m.screenLinesFor(i)
			if screenLines > m.height {
				m.offset++
				return
			}
		}
		return
	}
	maxOffset := len(m.lines) - m.height
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.offset < maxOffset {
		m.offset++
		if m.cursor < m.offset {
			m.cursor = m.nearestSelectable(m.offset, 1)
		}
	}
}

// ScrollUp scrolls the diff viewport up by one line.
func (m *diffViewModel) ScrollUp() {
	if m.offset > 0 {
		m.offset--
		if !m.wrap && m.cursor >= m.offset+m.height {
			m.cursor = m.nearestSelectable(m.offset+m.height-1, -1)
		}
	}
}

// ScrollDownHalfPage scrolls the diff viewport down by half a page.
func (m *diffViewModel) ScrollDownHalfPage() {
	jump := m.height / 2
	if jump < 1 {
		jump = 1
	}
	if m.wrap {
		for i := 0; i < jump; i++ {
			screenLines := 0
			canScroll := false
			for j := m.offset; j < len(m.lines); j++ {
				screenLines += m.screenLinesFor(j)
				if screenLines > m.height {
					m.offset++
					canScroll = true
					break
				}
			}
			if !canScroll {
				break
			}
		}
		return
	}
	maxOffset := len(m.lines) - m.height
	if maxOffset < 0 {
		maxOffset = 0
	}
	m.offset += jump
	if m.offset > maxOffset {
		m.offset = maxOffset
	}
	if m.cursor < m.offset {
		m.cursor = m.nearestSelectable(m.offset, 1)
	}
}

// ScrollUpHalfPage scrolls the diff viewport up by half a page.
func (m *diffViewModel) ScrollUpHalfPage() {
	jump := m.height / 2
	if jump < 1 {
		jump = 1
	}
	m.offset -= jump
	if m.offset < 0 {
		m.offset = 0
	}
	if !m.wrap && m.cursor >= m.offset+m.height {
		m.cursor = m.nearestSelectable(m.offset+m.height-1, -1)
	}
}

func (m *diffViewModel) ensureVisible() {
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if !m.wrap {
		if m.cursor >= m.offset+m.height {
			m.offset = m.cursor - m.height + 1
		}
		return
	}
	// Wrap mode: count screen lines from offset to cursor
	screenLines := 0
	for i := m.offset; i <= m.cursor && i < len(m.lines); i++ {
		screenLines += m.screenLinesFor(i)
	}
	for screenLines > m.height && m.offset < m.cursor {
		screenLines -= m.screenLinesFor(m.offset)
		m.offset++
	}
}

// isSelectable returns true if the line at idx is a diff content line (not a hunk header or comment).
func (m diffViewModel) isSelectable(idx int) bool {
	if idx < 0 || idx >= len(m.lines) {
		return false
	}
	line := m.lines[idx]
	if line.isHunk || line.isComment {
		return false
	}
	// Skip removed lines — they have no new-file line number and can't be commented on
	if line.kind == types.DiffLineRemoved && line.newLineNum == 0 {
		return false
	}
	return true
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
	// Only return new-file line numbers — comments reference lines that
	// exist in the current working tree so the agent can act on them.
	return line.newLineNum
}

func (m diffViewModel) currentDiffLine() int {
	return m.lineNumAt(m.cursor)
}

type openCommentMsg struct {
	path       string
	lineStart  int
	lineEnd    int
	targetType types.TargetType
}

func openCommentCmd(path string, start, end int, targetType types.TargetType) tea.Cmd {
	return func() tea.Msg {
		return openCommentMsg{path: path, lineStart: start, lineEnd: end, targetType: targetType}
	}
}

func openFileCommentCmd(path string, targetType types.TargetType) tea.Cmd {
	return func() tea.Msg {
		return openCommentMsg{path: path, lineStart: 0, lineEnd: 0, targetType: targetType}
	}
}

// insertInlineComments inserts comment lines after the diff line they target.
// It walks the existing lines (from the current hunk) in reverse-insertion order.
func (m *diffViewModel) insertInlineComments(hunk types.DiffHunk) {
	// Collect comments for this hunk
	var hunkComments []*types.ReviewComment
	for i := range m.comments {
		c := &m.comments[i]
		if c.TargetRef == m.path && c.LineStart >= hunk.NewStart && c.LineStart <= hunk.NewStart+hunk.NewCount {
			hunkComments = append(hunkComments, c)
		}
	}
	if len(hunkComments) == 0 {
		return
	}

	// Walk lines and insert comments after matching lines
	var result []diffViewLine
	for _, line := range m.lines {
		result = append(result, line)

		// Match on new-file line number (rightLineNum in split mode)
		lineNum := line.rightLineNum
		if lineNum == 0 {
			lineNum = line.newLineNum
		}
		if lineNum == 0 {
			continue
		}

		for _, c := range hunkComments {
			anchor := c.LineEnd
			if anchor == 0 {
				anchor = c.LineStart
			}
			if anchor == lineNum {
				result = append(result, diffViewLine{
					isComment: true,
					comment:   c,
					content:   formatInlineComment(c),
				})
			}
		}
	}
	m.lines = result
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


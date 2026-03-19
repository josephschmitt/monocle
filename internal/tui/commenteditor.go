package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/anthropics/monocle/internal/types"
)

type commentEditorModel struct {
	active      bool
	path        string
	lineStart   int
	lineEnd     int
	commentType types.CommentType
	body        string
	cursorPos   int
	width       int
	height      int
	theme       Theme
	editingID   string // non-empty when editing existing comment
}

func newCommentEditorModel(theme Theme) commentEditorModel {
	return commentEditorModel{
		commentType: types.CommentIssue,
		theme:       theme,
	}
}

type saveCommentMsg struct {
	path        string
	lineStart   int
	lineEnd     int
	commentType types.CommentType
	body        string
	editingID   string
}

type cancelCommentMsg struct{}

func (m commentEditorModel) Init() tea.Cmd {
	return nil
}

func (m commentEditorModel) Update(msg tea.Msg) (commentEditorModel, tea.Cmd) {
	if !m.active {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			m.active = false
			return m, func() tea.Msg { return cancelCommentMsg{} }
		case "enter":
			if strings.TrimSpace(m.body) == "" {
				return m, nil
			}
			saveMsg := saveCommentMsg{
				path:        m.path,
				lineStart:   m.lineStart,
				lineEnd:     m.lineEnd,
				commentType: m.commentType,
				body:        m.body,
				editingID:   m.editingID,
			}
			m.active = false
			return m, func() tea.Msg { return saveMsg }
		case "shift+enter", "alt+enter":
			m.body += "\n"
		case "tab":
			// Cycle comment type
			switch m.commentType {
			case types.CommentIssue:
				m.commentType = types.CommentSuggestion
			case types.CommentSuggestion:
				m.commentType = types.CommentNote
			case types.CommentNote:
				m.commentType = types.CommentPraise
			case types.CommentPraise:
				m.commentType = types.CommentIssue
			}
		case "backspace":
			if len(m.body) > 0 {
				m.body = m.body[:len(m.body)-1]
			}
		case "space":
			m.body += " "
		default:
			// Only add printable characters
			key := msg.String()
			if len(key) == 1 {
				m.body += key
			}
		}
	}
	return m, nil
}

func (m commentEditorModel) View() string {
	if !m.active {
		return ""
	}

	modalWidth := m.width * 2 / 3
	if modalWidth < 40 {
		modalWidth = 40
	}
	if modalWidth > m.width-4 {
		modalWidth = m.width - 4
	}

	var b strings.Builder

	// Title
	title := "New Comment"
	if m.editingID != "" {
		title = "Edit Comment"
	}
	b.WriteString(lipgloss.NewStyle().Bold(true).Render(title))
	b.WriteString("\n\n")

	// Target
	if m.lineStart > 0 {
		if m.lineEnd > m.lineStart {
			b.WriteString(fmt.Sprintf("File: %s (lines %d-%d)\n", m.path, m.lineStart, m.lineEnd))
		} else {
			b.WriteString(fmt.Sprintf("File: %s (line %d)\n", m.path, m.lineStart))
		}
	} else {
		b.WriteString(fmt.Sprintf("File: %s\n", m.path))
	}
	b.WriteString("\n")

	// Type selector
	typeLabels := []struct {
		t     types.CommentType
		label string
	}{
		{types.CommentIssue, "ISSUE"},
		{types.CommentSuggestion, "SUGGESTION"},
		{types.CommentNote, "NOTE"},
		{types.CommentPraise, "PRAISE"},
	}
	for _, tl := range typeLabels {
		if tl.t == m.commentType {
			b.WriteString(fmt.Sprintf("[%s] ", tl.label))
		} else {
			b.WriteString(fmt.Sprintf(" %s  ", strings.ToLower(tl.label)))
		}
	}
	b.WriteString("  (Tab to cycle)\n\n")

	// Text area
	bodyDisplay := m.body + "█"
	b.WriteString(bodyDisplay)
	b.WriteString("\n\n")

	// Hints
	b.WriteString(lipgloss.NewStyle().Faint(true).Render("Enter: save  Shift+Enter: newline  Esc: cancel  Tab: cycle type"))

	return m.theme.ModalBorder.Width(modalWidth).Render(b.String())
}

func (m *commentEditorModel) open(path string, lineStart, lineEnd int) {
	m.active = true
	m.path = path
	m.lineStart = lineStart
	m.lineEnd = lineEnd
	m.commentType = types.CommentIssue
	m.body = ""
	m.editingID = ""
}

func (m *commentEditorModel) openEdit(comment *types.ReviewComment) {
	m.active = true
	m.path = comment.TargetRef
	m.lineStart = comment.LineStart
	m.lineEnd = comment.LineEnd
	m.commentType = comment.Type
	m.body = comment.Body
	m.editingID = comment.ID
}

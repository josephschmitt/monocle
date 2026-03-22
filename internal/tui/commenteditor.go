package tui

import (
	"fmt"
	"image/color"
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
	targetType  types.TargetType
	commentType types.CommentType
	body        string
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
	targetType  types.TargetType
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
				targetType:  m.targetType,
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

	modalWidth := calcModalWidth(m.width, 0)

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
		b.WriteString(fmt.Sprintf("File: %s (file-level comment)\n", m.path))
	}
	b.WriteString("\n")

	// Type selector — each type has a color; selected gets solid bg, unselected gets colored text
	typeLabels := []struct {
		t     types.CommentType
		label string
		color color.Color
	}{
		{types.CommentIssue, "ISSUE", lipgloss.Color("1")},
		{types.CommentSuggestion, "SUGGESTION", lipgloss.Color("3")},
		{types.CommentNote, "NOTE", lipgloss.Color("4")},
		{types.CommentPraise, "PRAISE", lipgloss.Color("2")},
	}
	for i, tl := range typeLabels {
		var style lipgloss.Style
		if tl.t == m.commentType {
			style = lipgloss.NewStyle().
				Background(tl.color).
				Foreground(lipgloss.Color("0")).
				Bold(true).
				Padding(0, 1)
		} else {
			style = lipgloss.NewStyle().
				Foreground(tl.color).
				Padding(0, 1)
		}
		b.WriteString(style.Render(tl.label))
		if i < len(typeLabels)-1 {
			b.WriteString(" ")
		}
	}
	b.WriteString("  ")
	b.WriteString(lipgloss.NewStyle().Faint(true).Render("(Tab)"))
	b.WriteString("\n\n")

	// Text area
	bodyDisplay := m.body + "█"
	b.WriteString(bodyDisplay)
	b.WriteString("\n\n")

	// Hints
	b.WriteString(lipgloss.NewStyle().Faint(true).Render("Enter: save  Shift+Enter: newline  Esc: cancel  Tab: cycle type"))

	return m.theme.ModalBorder.Width(modalWidth).Render(b.String())
}

func (m *commentEditorModel) open(path string, lineStart, lineEnd int, targetType types.TargetType) {
	m.active = true
	m.path = path
	m.lineStart = lineStart
	m.lineEnd = lineEnd
	m.targetType = targetType
	m.commentType = types.CommentIssue
	m.body = ""
	m.editingID = ""
}

func (m *commentEditorModel) openEdit(comment *types.ReviewComment) {
	m.active = true
	m.path = comment.TargetRef
	m.lineStart = comment.LineStart
	m.lineEnd = comment.LineEnd
	m.targetType = comment.TargetType
	m.commentType = comment.Type
	m.body = comment.Body
	m.editingID = comment.ID
}

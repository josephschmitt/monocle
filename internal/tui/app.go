package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/anthropics/monocle/internal/core"
	"github.com/anthropics/monocle/internal/types"
)

// focusTarget identifies which pane holds keyboard focus.
type focusTarget int

const (
	focusSidebar focusTarget = iota
	focusMain
)

// overlayKind identifies which (if any) overlay is shown.
type overlayKind int

const (
	overlayNone overlayKind = iota
	overlayComment
	overlayReview
)

// Engine event messages bridged from core.EngineAPI callbacks.

type fileChangedMsg struct {
	path string
}

type agentStatusMsg struct {
	status string
}

type feedbackStatusMsg struct {
	status string
}

type contentItemMsg struct {
	id string
}

// appModel is the root model that composes all sub-models.
type appModel struct {
	engine core.EngineAPI

	sidebar       sidebarModel
	diffView      diffViewModel
	statusBar     statusBarModel
	commentEditor commentEditorModel
	reviewSummary reviewSummaryModel

	focus   focusTarget
	overlay overlayKind

	commandMode   bool
	commandBuffer string

	width  int
	height int

	theme Theme
	keys  KeyMap
}

// NewApp creates the root appModel.
func NewApp(engine core.EngineAPI) appModel {
	theme := DefaultTheme()
	return appModel{
		engine:        engine,
		sidebar:       newSidebarModel(),
		diffView:      newDiffViewModel(),
		statusBar:     newStatusBarModel(theme),
		commentEditor: newCommentEditorModel(theme),
		reviewSummary: newReviewSummaryModel(theme),
		focus:         focusSidebar,
		overlay:       overlayNone,
		theme:         theme,
		keys:          DefaultKeyMap(),
	}
}

// Init loads the initial file list from the engine.
func (m appModel) Init() tea.Cmd {
	return func() tea.Msg {
		files := m.engine.GetChangedFiles()
		items := m.engine.GetContentItems()
		return initialLoadMsg{files: files, items: items}
	}
}

// initialLoadMsg carries the initial file and content item lists.
type initialLoadMsg struct {
	files []types.ChangedFile
	items []types.ContentItem
}

// Update handles all incoming messages and routes them appropriately.
func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		const sidebarWidth = 30
		const statusBarHeight = 1
		mainWidth := m.width - sidebarWidth
		if mainWidth < 0 {
			mainWidth = 0
		}
		contentHeight := m.height - statusBarHeight
		if contentHeight < 0 {
			contentHeight = 0
		}
		m.sidebar.width = sidebarWidth
		m.sidebar.height = contentHeight
		m.diffView.width = mainWidth
		m.diffView.height = contentHeight
		m.statusBar.width = m.width
		m.commentEditor.width = m.width
		m.commentEditor.height = m.height
		m.reviewSummary.width = m.width
		m.reviewSummary.height = m.height
		return m, nil

	case initialLoadMsg:
		m.sidebar.files = msg.files
		m.sidebar.contentItems = msg.items
		// Sync status bar file count
		session := m.engine.GetSession()
		if session != nil {
			m.statusBar.baseRef = session.BaseRef
			m.statusBar.agentName = session.Agent
		}
		m.statusBar.fileCount = len(msg.files)
		m.statusBar.agentStatus = m.engine.GetAgentStatus()
		m.statusBar.feedbackStatus = m.engine.GetFeedbackStatus()
		return m, nil

	// Engine events
	case fileChangedMsg:
		m.sidebar.files = m.engine.GetChangedFiles()
		m.statusBar.fileCount = len(m.sidebar.files)
		return m, nil

	case agentStatusMsg:
		m.statusBar.agentStatus = m.engine.GetAgentStatus()
		return m, nil

	case feedbackStatusMsg:
		m.statusBar.feedbackStatus = msg.status
		return m, nil

	case contentItemMsg:
		m.sidebar.contentItems = m.engine.GetContentItems()
		return m, nil

	// Diff loading
	case loadDiffMsg:
		var cmd tea.Cmd
		m.diffView, cmd = m.diffView.Update(msg)
		return m, cmd

	// Sidebar selection → load diff
	case sidebarSelectMsg:
		return m, m.handleSidebarSelect(msg)

	// Comment overlay
	case openCommentMsg:
		m.commentEditor.open(msg.path, msg.lineStart, msg.lineEnd)
		m.overlay = overlayComment
		return m, nil

	case saveCommentMsg:
		m.overlay = overlayNone
		return m, m.handleSaveComment(msg)

	case cancelCommentMsg:
		m.overlay = overlayNone
		return m, nil

	// Review summary overlay open
	case openReviewMsg:
		m.reviewSummary.summary = msg.summary
		m.reviewSummary.active = true
		m.reviewSummary.agentStopped = msg.agentStopped
		m.overlay = overlayReview
		return m, nil

	// Review summary overlay
	case confirmSubmitMsg:
		m.overlay = overlayNone
		return m, func() tea.Msg {
			_, err := m.engine.Submit()
			if err != nil {
				return agentStatusMsg{status: "submit_error"}
			}
			return agentStatusMsg{status: "submitted"}
		}

	case cancelSubmitMsg:
		m.overlay = overlayNone
		return m, nil

	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

// handleKey processes keyboard input when no overlay is active.
func (m appModel) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	// If an overlay is active, route key to the overlay.
	if m.overlay == overlayComment {
		var cmd tea.Cmd
		m.commentEditor, cmd = m.commentEditor.Update(msg)
		return m, cmd
	}
	if m.overlay == overlayReview {
		var cmd tea.Cmd
		m.reviewSummary, cmd = m.reviewSummary.Update(msg)
		return m, cmd
	}

	// Command mode input.
	if m.commandMode {
		return m.handleCommandModeKey(msg)
	}

	key := msg.String()

	switch key {
	case ":":
		m.commandMode = true
		m.commandBuffer = ""
		m.statusBar.commandMode = true
		m.statusBar.commandBuffer = ""
		return m, nil

	case "q":
		return m, tea.Quit

	case "tab":
		if m.focus == focusSidebar {
			m.focus = focusMain
			m.sidebar.focused = false
			m.diffView.focused = true
		} else {
			m.focus = focusSidebar
			m.sidebar.focused = true
			m.diffView.focused = false
		}
		return m, nil

	case "r":
		return m, m.handleMarkReviewed()
	}

	// Route to focused sub-model.
	if m.focus == focusSidebar {
		var cmd tea.Cmd
		m.sidebar, cmd = m.sidebar.Update(msg)
		return m, cmd
	}

	var cmd tea.Cmd
	m.diffView, cmd = m.diffView.Update(msg)
	return m, cmd
}

// handleCommandModeKey processes keystrokes while in command mode.
func (m appModel) handleCommandModeKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "escape":
		m.commandMode = false
		m.commandBuffer = ""
		m.statusBar.commandMode = false
		m.statusBar.commandBuffer = ""
		return m, nil

	case "enter":
		cmd := m.executeCommand(m.commandBuffer)
		m.commandMode = false
		m.commandBuffer = ""
		m.statusBar.commandMode = false
		m.statusBar.commandBuffer = ""
		return m, cmd

	case "backspace":
		if len(m.commandBuffer) > 0 {
			m.commandBuffer = m.commandBuffer[:len(m.commandBuffer)-1]
			m.statusBar.commandBuffer = m.commandBuffer
		}
		return m, nil

	default:
		if len(key) == 1 || key == " " {
			m.commandBuffer += key
			m.statusBar.commandBuffer = m.commandBuffer
		}
		return m, nil
	}
}

// openReviewMsg carries the data needed to open the review summary overlay.
type openReviewMsg struct {
	summary     *types.ReviewSummary
	agentStopped bool
}

// executeCommand runs a named command entered in command mode.
func (m appModel) executeCommand(cmd string) tea.Cmd {
	engine := m.engine
	switch strings.TrimSpace(cmd) {
	case "submit":
		return func() tea.Msg {
			summary, err := engine.GetReviewSummary()
			if err != nil || summary == nil {
				return cancelSubmitMsg{}
			}
			session := engine.GetSession()
			agentStopped := session != nil && session.AgentStatus == types.AgentStatusStopped
			return openReviewMsg{summary: summary, agentStopped: agentStopped}
		}

	case "submit!":
		return func() tea.Msg {
			_, err := engine.Submit()
			if err != nil {
				return agentStatusMsg{status: "submit_error"}
			}
			return agentStatusMsg{status: "submitted"}
		}

	case "approve":
		return func() tea.Msg {
			_, err := engine.Approve()
			if err != nil {
				return agentStatusMsg{status: "approve_error"}
			}
			return agentStatusMsg{status: "approved"}
		}

	case "dismiss-outdated":
		return func() tea.Msg {
			_ = engine.DismissOutdated()
			return fileChangedMsg{}
		}
	}

	return nil
}

// handleSidebarSelect loads the diff for the selected file or content item.
func (m appModel) handleSidebarSelect(msg sidebarSelectMsg) tea.Cmd {
	if msg.isContent {
		return func() tea.Msg {
			item, err := m.engine.GetContentItem(msg.contentID)
			if err != nil || item == nil {
				return loadDiffMsg{path: msg.contentID}
			}
			return loadDiffMsg{path: item.Title}
		}
	}
	return func() tea.Msg {
		result, err := m.engine.GetFileDiff(msg.path)
		if err != nil {
			return loadDiffMsg{path: msg.path}
		}
		session := m.engine.GetSession()
		var comments []types.ReviewComment
		if session != nil {
			for _, c := range session.Comments {
				if c.TargetRef == msg.path {
					comments = append(comments, c)
				}
			}
		}
		return loadDiffMsg{
			path:     msg.path,
			result:   result,
			comments: comments,
		}
	}
}

// handleSaveComment persists a new or edited comment then reloads the diff.
func (m appModel) handleSaveComment(msg saveCommentMsg) tea.Cmd {
	return func() tea.Msg {
		target := core.CommentTarget{
			TargetRef: msg.path,
			LineStart: msg.lineStart,
			LineEnd:   msg.lineEnd,
		}
		if target.LineStart > 0 {
			target.TargetType = types.TargetFile
		} else {
			target.TargetType = types.TargetContent
		}

		if msg.editingID != "" {
			_, _ = m.engine.EditComment(msg.editingID, msg.body)
		} else {
			_, _ = m.engine.AddComment(target, msg.commentType, msg.body)
		}

		// Reload diff for the file
		result, err := m.engine.GetFileDiff(msg.path)
		if err != nil {
			return loadDiffMsg{path: msg.path}
		}
		session := m.engine.GetSession()
		var comments []types.ReviewComment
		if session != nil {
			for _, c := range session.Comments {
				if c.TargetRef == msg.path {
					comments = append(comments, c)
				}
			}
		}
		return loadDiffMsg{
			path:     msg.path,
			result:   result,
			comments: comments,
		}
	}
}

// handleMarkReviewed toggles the reviewed status of the currently selected file.
func (m appModel) handleMarkReviewed() tea.Cmd {
	if m.focus != focusSidebar {
		return nil
	}
	if m.sidebar.cursor >= len(m.sidebar.files) {
		return nil
	}
	file := m.sidebar.files[m.sidebar.cursor]
	return func() tea.Msg {
		if file.Reviewed {
			_ = m.engine.UnmarkReviewed(file.Path)
		} else {
			_ = m.engine.MarkReviewed(file.Path)
		}
		return fileChangedMsg{path: file.Path}
	}
}

// View renders the full TUI layout.
func (m appModel) View() tea.View {
	sidebarView := m.theme.SidebarBorder.
		Width(m.sidebar.width).
		Height(m.sidebar.height).
		Render(m.sidebar.View())

	mainView := m.theme.MainPane.
		Width(m.diffView.width).
		Height(m.diffView.height).
		Render(m.diffView.View())

	body := lipgloss.JoinHorizontal(lipgloss.Top, sidebarView, mainView)
	statusView := m.statusBar.View()
	full := lipgloss.JoinVertical(lipgloss.Left, body, statusView)

	// Render overlay centered on top of the layout if active.
	if m.overlay == overlayComment {
		overlayContent := m.commentEditor.View()
		if overlayContent != "" {
			full = overlayOn(full, overlayContent, m.width, m.height)
		}
	} else if m.overlay == overlayReview {
		overlayContent := m.reviewSummary.View()
		if overlayContent != "" {
			full = overlayOn(full, overlayContent, m.width, m.height)
		}
	}

	v := tea.NewView(full)
	v.AltScreen = true
	return v
}

// overlayOn centers overlayContent over baseContent.
func overlayOn(base, overlay string, width, height int) string {
	overlayLines := strings.Split(overlay, "\n")
	overlayH := len(overlayLines)
	overlayW := 0
	for _, l := range overlayLines {
		if w := lipgloss.Width(l); w > overlayW {
			overlayW = w
		}
	}

	topPad := (height - overlayH) / 2
	if topPad < 0 {
		topPad = 0
	}
	leftPad := (width - overlayW) / 2
	if leftPad < 0 {
		leftPad = 0
	}

	padLeft := strings.Repeat(" ", leftPad)
	var overlayBlock strings.Builder
	for i := 0; i < topPad; i++ {
		overlayBlock.WriteString("\n")
	}
	for i, line := range overlayLines {
		overlayBlock.WriteString(padLeft + line)
		if i < len(overlayLines)-1 {
			overlayBlock.WriteString("\n")
		}
	}

	// Overlay the text on top of base by rebuilding line by line.
	baseLines := strings.Split(base, "\n")
	oLines := strings.Split(overlayBlock.String(), "\n")

	result := make([]string, len(baseLines))
	copy(result, baseLines)
	for i, ol := range oLines {
		if i >= len(result) {
			break
		}
		if strings.TrimSpace(ol) != "" {
			result[i] = ol
		}
	}
	return strings.Join(result, "\n")
}

// BridgeEngineEvents subscribes to engine events and forwards them to the
// Bubble Tea program as messages. Call this after tea.NewProgram but before
// p.Run().
func BridgeEngineEvents(engine core.EngineAPI, p *tea.Program) {
	engine.On(core.EventFileChanged, func(e core.EventPayload) {
		p.Send(fileChangedMsg{path: e.Path})
	})
	engine.On(core.EventAgentStatusChanged, func(e core.EventPayload) {
		p.Send(agentStatusMsg{status: e.Status})
	})
	engine.On(core.EventFeedbackStatusChanged, func(e core.EventPayload) {
		p.Send(feedbackStatusMsg{status: e.Status})
	})
	engine.On(core.EventContentItemAdded, func(e core.EventPayload) {
		p.Send(contentItemMsg{id: e.ItemID})
	})
}

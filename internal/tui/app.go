package tui

import (
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/anthropics/monocle/internal/clipboard"
	"github.com/anthropics/monocle/internal/core"
	"github.com/anthropics/monocle/internal/types"
)

// focusTarget identifies which pane holds keyboard focus.
type focusTarget int

const (
	focusSidebar focusTarget = iota
	focusMain
)

// layoutMode determines whether panes are arranged horizontally or stacked vertically.
type layoutMode int

const (
	layoutHorizontal layoutMode = iota
	layoutStacked
)

const layoutBreakpoint = 110

// overlayKind identifies which (if any) overlay is shown.
type overlayKind int

const (
	overlayNone overlayKind = iota
	overlayComment
	overlayReview
	overlayHelp
	overlayRefPicker
	overlayConfirm
	overlayInstallPrompt
	overlayConnectionInfo
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

type connectionChangedMsg struct {
	count int
}

type pauseChangedMsg struct {
	status string
}

type submitSuccessMsg struct{}

type commentsClearedMsg struct {
	reloadPath string
	isContent  bool
}

type resolveCommentMsg struct {
	commentID string
}

type openConfirmMsg struct {
	title   string
	message string
	action  confirmAction
}

type mcpInstallPromptMsg struct{}

type mcpInstallResultMsg struct {
	err error
}

type refreshTickMsg struct{}

func refreshTick() tea.Cmd {
	return tea.Tick(2*time.Second, func(time.Time) tea.Msg {
		return refreshTickMsg{}
	})
}

// AppOptions configures optional behavior for the TUI app.
type AppOptions struct {
	MCPInstallFn func(global bool) error // if non-nil, offer MCP auto-install on startup
}

// appModel is the root model that composes all sub-models.
type appModel struct {
	engine core.EngineAPI

	sidebar       sidebarModel
	diffView      diffViewModel
	statusBar     statusBarModel
	commentEditor commentEditorModel
	reviewSummary reviewSummaryModel
	help          helpModel
	refPicker      refPickerModel
	confirm        confirmModel
	connectionInfo connectionInfoModel

	focus        focusTarget
	overlay      overlayKind
	layout       layoutMode
	layoutConfig string

	commandMode   bool
	commandBuffer string

	width  int
	height int

	theme Theme
	keys  KeyMap

	mcpInstallFn    func(global bool) error
	installPrompt   installPromptModel
}

// NewApp creates the root appModel.
func NewApp(engine core.EngineAPI, opts ...AppOptions) appModel {
	var o AppOptions
	if len(opts) > 0 {
		o = opts[0]
	}

	theme := DefaultTheme()
	keys := DefaultKeyMap()
	sidebar := newSidebarModel(&keys)
	sidebar.focused = true
	dv := newDiffViewModel(&theme, &keys)
	var layoutCfg string

	if engine != nil {
		if cfg := engine.GetConfig(); cfg != nil {
			if cfg.Keybindings != nil {
				keys = keys.ApplyOverrides(cfg.Keybindings)
			}
			if cfg.SidebarStyle == "tree" {
				sidebar.treeMode = true
			}
			if cfg.DiffStyle == "split" {
				dv.style = diffStyleSplit
			}
			if cfg.Layout != "" {
				layoutCfg = cfg.Layout
			}
			if cfg.Wrap {
				dv.wrap = true
			}
			if cfg.TabSize > 0 {
				dv.tabSize = cfg.TabSize
			}
		}
	}

	return appModel{
		engine:        engine,
		sidebar:       sidebar,
		diffView:      dv,
		statusBar:     newStatusBarModel(theme),
		commentEditor: newCommentEditorModel(theme),
		reviewSummary: newReviewSummaryModel(theme),
		help:          newHelpModel(theme, &keys),
		refPicker:     newRefPickerModel(theme),
		confirm:        newConfirmModel(theme),
		connectionInfo: newConnectionInfoModel(theme),
		installPrompt: newInstallPromptModel(theme),
		focus:         focusSidebar,
		overlay:       overlayNone,
		layoutConfig:  layoutCfg,
		theme:         theme,
		keys:          keys,
		mcpInstallFn:  o.MCPInstallFn,
	}
}

// Init loads the initial file list from the engine and starts the refresh tick.
func (m appModel) Init() tea.Cmd {
	cmds := []tea.Cmd{
		func() tea.Msg {
			files := m.engine.GetChangedFiles()
			items := m.engine.GetContentItems()
			return initialLoadMsg{files: files, items: items}
		},
		refreshTick(),
	}
	if m.mcpInstallFn != nil {
		cmds = append(cmds, func() tea.Msg {
			return mcpInstallPromptMsg{}
		})
	}
	return tea.Batch(cmds...)
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

		const borderW = 2 // left + right border
		const borderH = 2 // top + bottom border
		const titleHeight = 1
		const statusBarHeight = 1
		const chrome = titleHeight + statusBarHeight + borderH

		contentHeight := m.height - chrome
		if contentHeight < 0 {
			contentHeight = 0
		}

		switch m.layoutConfig {
		case "side-by-side":
			m.layout = layoutHorizontal
		case "stacked":
			m.layout = layoutStacked
		default: // "auto" or ""
			if m.width < layoutBreakpoint {
				m.layout = layoutStacked
			} else {
				m.layout = layoutHorizontal
			}
		}

		if m.layout == layoutStacked {
			contentW := m.width - borderW
			if contentW < 0 {
				contentW = 0
			}

			sidebarH := stackedSidebarHeight(contentHeight, len(m.sidebar.files), len(m.sidebar.contentItems))
			diffH := contentHeight - sidebarH - borderH // account for sidebar border
			if diffH < 0 {
				diffH = 0
			}

			m.sidebar.width = contentW
			m.sidebar.height = sidebarH
			m.diffView.width = contentW
			m.diffView.height = diffH
		} else {
			// Prioritize diff area: guarantee 80 chars for diff content,
			// then let sidebar grow up to 1/3 of width (clamped to [30, 50]).
			const minDiffContent = 80
			maxSidebarForDiff := m.width - minDiffContent - 2*borderW // room left after diff + both borders
			sidebarContentW := m.width / 3
			if sidebarContentW > maxSidebarForDiff {
				sidebarContentW = maxSidebarForDiff
			}
			if sidebarContentW < 30 {
				sidebarContentW = 30
			}
			if sidebarContentW > 50 {
				sidebarContentW = 50
			}

			sidebarOuter := sidebarContentW + borderW
			mainOuter := m.width - sidebarOuter
			if mainOuter < 0 {
				mainOuter = 0
			}

			m.sidebar.width = sidebarContentW
			m.sidebar.height = contentHeight
			m.diffView.width = mainOuter - borderW
			m.diffView.height = contentHeight
		}

		m.statusBar.width = m.width
		m.commentEditor.width = m.width
		m.commentEditor.height = m.height
		m.reviewSummary.width = m.width
		m.reviewSummary.height = m.height
		m.help.width = m.width
		m.help.height = m.height
		m.confirm.width = m.width
		m.confirm.height = m.height
		m.installPrompt.width = m.width
		m.installPrompt.height = m.height
		m.connectionInfo.width = m.width
		m.connectionInfo.height = m.height
		return m, nil

	case initialLoadMsg:
		m.sidebar.files = msg.files
		m.sidebar.contentItems = msg.items
		m.sidebar.rebuildTree()
		m.sidebar.clampOffset()
		recalcStackedLayout(&m)
		// Sync status bar file count
		session := m.engine.GetSession()
		if session != nil {
			m.statusBar.baseRef = session.BaseRef
			m.statusBar.agentName = session.Agent
		}
		m.statusBar.fileCount = len(msg.files)
		m.statusBar.agentStatus = m.engine.GetAgentStatus()
		m.statusBar.feedbackStatus = m.engine.GetFeedbackStatus()
		// Auto-select the first file, or first content item if no files
		if len(msg.files) > 0 {
			return m, m.handleSidebarSelect(sidebarSelectMsg{path: msg.files[0].Path})
		}
		if len(msg.items) > 0 {
			return m, m.handleSidebarSelect(sidebarSelectMsg{isContent: true, contentID: msg.items[0].ID})
		}
		return m, nil

	// Periodic refresh
	case refreshTickMsg:
		return m, tea.Batch(m.refreshFiles(), refreshTick())

	case refreshResultMsg:
		m.sidebar.files = msg.files
		m.sidebar.rebuildTree()
		m.sidebar.clampOffset()
		recalcStackedLayout(&m)
		m.statusBar.fileCount = len(msg.files)
		if msg.path != "" && msg.result != nil {
			m.diffView, _ = m.diffView.Update(loadDiffMsg{
				path:     msg.path,
				result:   msg.result,
				comments: msg.comments,
			})
		}
		// Auto-select first file if current view is stale
		if len(msg.files) > 0 && !m.diffViewShowsValidFile() {
			return m, m.handleSidebarSelect(sidebarSelectMsg{path: msg.files[0].Path})
		}
		return m, nil

	// Engine events
	case fileChangedMsg:
		m.sidebar.files = m.engine.GetChangedFiles()
		m.sidebar.rebuildTree()
		m.sidebar.clampOffset()
		recalcStackedLayout(&m)
		m.statusBar.fileCount = len(m.sidebar.files)
		session := m.engine.GetSession()
		if session != nil {
			m.statusBar.baseRef = session.BaseRef
			m.statusBar.commentCount = len(session.Comments)
		}
		// Auto-select first file if the current view is empty or stale
		if len(m.sidebar.files) > 0 && !m.diffViewShowsValidFile() {
			return m, m.handleSidebarSelect(sidebarSelectMsg{path: m.sidebar.files[0].Path})
		}
		return m, nil

	case agentStatusMsg:
		m.statusBar.agentStatus = m.engine.GetAgentStatus()
		return m, nil

	case connectionChangedMsg:
		m.statusBar.connected = msg.count > 0
		return m, nil

	case feedbackStatusMsg:
		m.statusBar.feedbackStatus = msg.status
		return m, nil

	case contentItemMsg:
		m.sidebar.contentItems = m.engine.GetContentItems()
		m.sidebar.clampOffset()
		recalcStackedLayout(&m)
		// Auto-select only if nothing else is available (no files, no current view)
		if m.diffView.path == "" && len(m.sidebar.files) == 0 && msg.id != "" {
			return m, m.handleSidebarSelect(sidebarSelectMsg{isContent: true, contentID: msg.id})
		}
		return m, nil

	case pauseChangedMsg:
		m.statusBar.agentStatus = m.engine.GetAgentStatus()
		return m, nil

	case baseRefChangedMsg:
		session := m.engine.GetSession()
		if session != nil {
			m.statusBar.baseRef = session.BaseRef
		}
		return m, m.refreshFiles()

	case openRefPickerMsg:
		m.refPicker.entries = msg.entries
		m.refPicker.autoActive = msg.autoActive
		m.refPicker.active = true
		m.refPicker.cursor = 0
		m.refPicker.width = m.width
		m.refPicker.height = m.height
		m.overlay = overlayRefPicker
		return m, nil

	case selectRefMsg:
		m.overlay = overlayNone
		m.refPicker.active = false
		if msg.auto {
			return m, m.executeCommand("ref auto")
		}
		return m, m.executeCommand("ref " + msg.hash)

	case cancelRefPickerMsg:
		m.overlay = overlayNone
		m.refPicker.active = false
		return m, nil

	// Diff loading
	case loadDiffMsg:
		var cmd tea.Cmd
		m.diffView, cmd = m.diffView.Update(msg)
		return m, cmd

	// Content item loading (plans, docs)
	case loadContentMsg:
		var cmd tea.Cmd
		m.diffView, cmd = m.diffView.Update(msg)
		return m, cmd

	// Sidebar selection → load diff (focus stays where it is)
	case sidebarSelectMsg:
		return m, m.handleSidebarSelect(msg)

	// Comment overlay
	case openCommentMsg:
		m.commentEditor.open(msg.path, msg.lineStart, msg.lineEnd, msg.targetType)
		m.overlay = overlayComment
		return m, nil

	case saveCommentMsg:
		m.overlay = overlayNone
		m.diffView.visualMode = false
		return m, m.handleSaveComment(msg)

	case cancelCommentMsg:
		m.overlay = overlayNone
		return m, nil

	case closeHelpMsg:
		m.overlay = overlayNone
		return m, nil

	case closeConnectionInfoMsg:
		m.overlay = overlayNone
		return m, nil

	case resolveCommentMsg:
		engine := m.engine
		id := msg.commentID
		currentPath := m.diffView.path
		isContent := m.diffView.contentMode
		return m, func() tea.Msg {
			_ = engine.ResolveComment(id)
			if isContent {
				return fileChangedMsg{}
			}
			// Reload diff to update comment display
			result, _ := engine.GetFileDiff(currentPath)
			session := engine.GetSession()
			var comments []types.ReviewComment
			if session != nil {
				for _, c := range session.Comments {
					if c.TargetRef == currentPath {
						comments = append(comments, c)
					}
				}
			}
			return loadDiffMsg{path: currentPath, result: result, comments: comments}
		}

	// Review summary overlay open
	case openReviewMsg:
		m.reviewSummary.open(msg.summary, msg.agentStopped)
		m.overlay = overlayReview
		return m, nil

	// Review summary → submit with user-chosen action and body
	case confirmSubmitMsg:
		m.overlay = overlayNone
		action := msg.action
		body := msg.body
		copyToClip := msg.copyToClipboard
		engine := m.engine
		return m, func() tea.Msg {
			if err := engine.Submit(action, body); err != nil {
				return agentStatusMsg{status: "submit_error"}
			}
			if copyToClip {
				if text, err := engine.FormatReview(action, body); err == nil {
					clipboard.Copy(text)
				}
			}
			return submitSuccessMsg{}
		}

	case yankReviewMsg:
		m.overlay = overlayNone
		action := msg.action
		body := msg.body
		engine := m.engine
		return m, func() tea.Msg {
			text, err := engine.FormatReview(action, body)
			if err != nil {
				return yankFailMsg{err: err.Error()}
			}
			if copyErr := clipboard.Copy(text); copyErr != nil {
				return yankFailMsg{err: copyErr.Error()}
			}
			return yankSuccessMsg{}
		}

	case yankSuccessMsg:
		m.statusBar.feedbackStatus = "copied"
		return m, nil

	case yankFailMsg:
		m.statusBar.feedbackStatus = "copy_failed"
		return m, nil

	case cancelSubmitMsg:
		m.overlay = overlayNone
		return m, nil

	// Post-submit: offer to clear comments
	case submitSuccessMsg:
		m.statusBar.feedbackStatus = m.engine.GetFeedbackStatus()
		session := m.engine.GetSession()
		if session != nil {
			m.statusBar.commentCount = len(session.Comments)
		}
		// Only ask to clear if there are active comments
		hasActive := false
		if session != nil {
			for _, c := range session.Comments {
				if !c.Outdated {
					hasActive = true
					break
				}
			}
		}
		if hasActive {
			m.confirm.open("Review Submitted", "Clear all comments from this review?", confirmClearAfterSubmit)
			m.overlay = overlayConfirm
		}
		return m, nil

	// Confirm overlay actions
	case confirmActionMsg:
		m.overlay = overlayNone
		engine := m.engine
		currentPath := m.diffView.path
		isContent := m.diffView.contentMode
		switch msg.action {
		case confirmClearAfterSubmit, confirmDiscard:
			return m, func() tea.Msg {
				_ = engine.ClearComments()
				return commentsClearedMsg{reloadPath: currentPath, isContent: isContent}
			}
		}
		return m, nil

	case mcpInstallPromptMsg:
		m.installPrompt.open()
		m.overlay = overlayInstallPrompt
		return m, nil

	case installMCPMsg:
		m.overlay = overlayNone
		installFn := m.mcpInstallFn
		global := msg.global
		m.statusBar.feedbackStatus = "Installing MCP channel..."
		return m, func() tea.Msg {
			return mcpInstallResultMsg{err: installFn(global)}
		}

	case cancelInstallMsg:
		m.overlay = overlayNone
		return m, nil

	case mcpInstallResultMsg:
		if msg.err != nil {
			m.statusBar.feedbackStatus = "MCP install failed"
		} else {
			m.statusBar.feedbackStatus = "MCP channel installed"
			m.mcpInstallFn = nil
		}
		return m, nil

	case openConfirmMsg:
		m.confirm.open(msg.title, msg.message, msg.action)
		m.overlay = overlayConfirm
		return m, nil

	case cancelConfirmMsg:
		m.overlay = overlayNone
		return m, nil

	// After comments are cleared, refresh sidebar + diff
	case commentsClearedMsg:
		m.sidebar.files = m.engine.GetChangedFiles()
		m.sidebar.rebuildTree()
		m.sidebar.clampOffset()
		recalcStackedLayout(&m)
		m.statusBar.fileCount = len(m.sidebar.files)
		session := m.engine.GetSession()
		if session != nil {
			m.statusBar.baseRef = session.BaseRef
			m.statusBar.commentCount = len(session.Comments)
		}
		// Reload current view to remove inline comment markers
		if msg.reloadPath != "" && msg.isContent {
			return m, m.handleSidebarSelect(sidebarSelectMsg{isContent: true, contentID: msg.reloadPath})
		} else if msg.reloadPath != "" {
			return m, m.handleSidebarSelect(sidebarSelectMsg{path: msg.reloadPath})
		}
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
	if m.overlay == overlayHelp {
		var cmd tea.Cmd
		m.help, cmd = m.help.Update(msg)
		return m, cmd
	}
	if m.overlay == overlayRefPicker {
		var cmd tea.Cmd
		m.refPicker, cmd = m.refPicker.Update(msg)
		return m, cmd
	}
	if m.overlay == overlayConfirm {
		var cmd tea.Cmd
		m.confirm, cmd = m.confirm.Update(msg)
		return m, cmd
	}
	if m.overlay == overlayInstallPrompt {
		var cmd tea.Cmd
		m.installPrompt, cmd = m.installPrompt.Update(msg)
		return m, cmd
	}
	if m.overlay == overlayConnectionInfo {
		var cmd tea.Cmd
		m.connectionInfo, cmd = m.connectionInfo.Update(msg)
		return m, cmd
	}

	// Command mode input.
	if m.commandMode {
		return m.handleCommandModeKey(msg)
	}

	key := msg.String()
	km := m.keys

	// Check for pane-number shortcuts (1, 2, etc.)
	if pane, ok := km.FocusPaneN[key]; ok {
		switch pane {
		case 1:
			m.focus = focusSidebar
			m.sidebar.focused = true
			m.diffView.focused = false
		case 2:
			m.focus = focusMain
			m.sidebar.focused = false
			m.diffView.focused = true
		}
		return m, nil
	}

	switch {
	case Matches(key, km.CommandMode):
		m.commandMode = true
		m.commandBuffer = ""
		m.statusBar.commandMode = true
		m.statusBar.commandBuffer = ""
		return m, nil

	case Matches(key, km.Quit):
		return m, tea.Quit

	case Matches(key, km.Help):
		m.help.active = true
		m.help.scrollOffset = 0
		m.overlay = overlayHelp
		return m, nil

	case key == "I":
		m.connectionInfo.active = true
		m.connectionInfo.socketPath = m.engine.GetSocketPath()
		m.connectionInfo.subscriberCount = m.engine.GetSubscriberCount()
		m.overlay = overlayConnectionInfo
		return m, nil

	case Matches(key, km.FocusSwap):
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

	case Matches(key, km.FileComment):
		// File-level comment from sidebar
		if m.focus == focusSidebar {
			if f := m.sidebar.selectedFile(); f != nil {
				return m, openFileCommentCmd(f.Path, types.TargetFile)
			}
			return m, nil
		}
		// Delegate to diff view when focused on main
		var cmd tea.Cmd
		m.diffView, cmd = m.diffView.Update(msg)
		return m, cmd

	case Matches(key, km.Reviewed):
		return m, m.handleMarkReviewed()

	case Matches(key, km.Submit):
		return m, m.executeCommand("submit")

	case Matches(key, km.DismissOutdated):
		return m, m.executeCommand("dismiss-outdated")

	case Matches(key, km.Pause):
		return m, m.executeCommand("pause")

	case Matches(key, km.CycleLayout):
		switch m.layoutConfig {
		case "", "auto":
			m.layoutConfig = "side-by-side"
		case "side-by-side":
			m.layoutConfig = "stacked"
		default:
			m.layoutConfig = "auto"
		}
		return m, func() tea.Msg {
			return tea.WindowSizeMsg{Width: m.width, Height: m.height}
		}

	case Matches(key, km.BaseRef):
		engine := m.engine
		return m, func() tea.Msg {
			entries, err := engine.RecentCommits(20)
			if err != nil {
				return nil
			}
			return openRefPickerMsg{
				entries:    entries,
				autoActive: engine.IsAutoAdvanceRef(),
			}
		}

	case Matches(key, km.ScrollDown):
		m.diffView.ScrollDown()
		return m, nil

	case Matches(key, km.ScrollUp):
		m.diffView.ScrollUp()
		return m, nil

	case Matches(key, km.ScrollLeft):
		m.diffView.ScrollLeft()
		return m, nil

	case Matches(key, km.ScrollRight):
		m.diffView.ScrollRight()
		return m, nil

	case Matches(key, km.HalfDown):
		m.diffView.ScrollDownHalfPage()
		return m, nil

	case Matches(key, km.HalfUp):
		m.diffView.ScrollUpHalfPage()
		return m, nil

	case Matches(key, km.PrevFile):
		cmd := m.sidebar.navigateFile(-1)
		return m, cmd

	case Matches(key, km.NextFile):
		cmd := m.sidebar.navigateFile(+1)
		return m, cmd

	case Matches(key, km.Select):
		if m.focus == focusSidebar {
			// In tree mode, enter on a directory toggles collapse
			if m.sidebar.treeMode && m.sidebar.selectedFile() == nil {
				var cmd tea.Cmd
				m.sidebar, cmd = m.sidebar.Update(msg)
				return m, cmd
			}
			m.focus = focusMain
			m.sidebar.focused = false
			m.diffView.focused = true
			return m, nil
		}
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
	case "esc":
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
	summary      *types.ReviewSummary
	agentStopped bool
}

// executeCommand runs a named command entered in command mode.
func (m appModel) executeCommand(cmd string) tea.Cmd {
	engine := m.engine
	switch strings.TrimSpace(cmd) {
	case "submit":
		return func() tea.Msg {
			summary, err := engine.GetReviewSummary()
			if err != nil {
				return cancelSubmitMsg{}
			}
			if summary == nil {
				summary = &types.ReviewSummary{
					FileComments:    map[string][]types.ReviewComment{},
					ContentComments: map[string][]types.ReviewComment{},
				}
			}
			session := engine.GetSession()
			agentStopped := session != nil && session.AgentStatus == types.AgentStatusPaused
			return openReviewMsg{summary: summary, agentStopped: agentStopped}
		}

	case "submit!":
		return func() tea.Msg {
			// Auto-detect action: request_changes if issues/suggestions, approve otherwise
			action := types.ActionApprove
			summary, _ := engine.GetReviewSummary()
			if summary != nil && (summary.IssueCt+summary.SuggestionCt > 0) {
				action = types.ActionRequestChanges
			}
			if err := engine.Submit(action, ""); err != nil {
				return agentStatusMsg{status: "submit_error"}
			}
			return submitSuccessMsg{}
		}

	case "discard":
		return func() tea.Msg {
			session := engine.GetSession()
			hasActive := false
			if session != nil {
				for _, c := range session.Comments {
					if !c.Outdated {
						hasActive = true
						break
					}
				}
			}
			if !hasActive {
				return nil
			}
			return openConfirmMsg{
				title:   "Discard Review",
				message: "Discard all pending comments? This cannot be undone.",
				action:  confirmDiscard,
			}
		}

	case "dismiss-outdated":
		return func() tea.Msg {
			_ = engine.DismissOutdated()
			return fileChangedMsg{}
		}

	case "pause":
		return func() tea.Msg {
			engine.RequestPause()
			return pauseChangedMsg{status: "pause_requested"}
		}

	case "unpause":
		return func() tea.Msg {
			engine.CancelPause()
			return pauseChangedMsg{status: "cancelled"}
		}
	}

	// Handle :ref commands
	trimmed := strings.TrimSpace(cmd)
	if strings.HasPrefix(trimmed, "ref ") {
		arg := strings.TrimSpace(trimmed[4:])
		if arg == "auto" {
			return func() tea.Msg {
				engine.SetAutoAdvanceRef(true)
				return baseRefChangedMsg{}
			}
		}
		return func() tea.Msg {
			if err := engine.SetBaseRef(arg); err != nil {
				return baseRefChangedMsg{err: err.Error()}
			}
			return baseRefChangedMsg{}
		}
	}

	return nil
}

type baseRefChangedMsg struct {
	err string
}

// stackedSidebarHeight returns the height for the sidebar in stacked mode.
// It accounts for the header line plus one line per file/content item, with a
// minimum of 4 rows and at most 40% of totalHeight.
func stackedSidebarHeight(totalHeight, fileCount, contentItemCount int) int {
	// 1 header line + 1 per item
	h := 1 + fileCount + contentItemCount
	if h < 4 {
		h = 4
	}
	maxH := totalHeight * 40 / 100
	if maxH < 4 {
		maxH = 4
	}
	if h > maxH {
		h = maxH
	}
	return h
}

// recalcStackedLayout recalculates sidebar and diff view heights for stacked
// mode based on the current file/content item counts. No-op in horizontal mode.
func recalcStackedLayout(m *appModel) {
	if m.layout != layoutStacked {
		return
	}
	const borderH = 2
	const titleHeight = 1
	const statusBarHeight = 1
	const chrome = titleHeight + statusBarHeight + borderH

	contentHeight := m.height - chrome
	if contentHeight < 0 {
		contentHeight = 0
	}

	sidebarH := stackedSidebarHeight(contentHeight, len(m.sidebar.files), len(m.sidebar.contentItems))
	diffH := contentHeight - sidebarH - borderH
	if diffH < 0 {
		diffH = 0
	}

	m.sidebar.height = sidebarH
	m.diffView.height = diffH
}

// diffViewShowsValidFile returns true if the diff view is showing a valid
// view — either a file still in the file list or a content item.
func (m appModel) diffViewShowsValidFile() bool {
	if m.diffView.path == "" {
		return false
	}
	if m.diffView.contentMode {
		return true // content items are always valid
	}
	for _, f := range m.sidebar.files {
		if f.Path == m.diffView.path {
			return true
		}
	}
	return false
}

// handleSidebarSelect loads the diff for the selected file or content item.
func (m appModel) handleSidebarSelect(msg sidebarSelectMsg) tea.Cmd {
	if msg.isContent {
		return func() tea.Msg {
			item, err := m.engine.GetContentItem(msg.contentID)
			if err != nil || item == nil {
				return loadDiffMsg{path: msg.contentID}
			}
			session := m.engine.GetSession()
			var comments []types.ReviewComment
			if session != nil {
				for _, c := range session.Comments {
					if c.TargetRef == item.ID && c.TargetType == types.TargetContent {
						comments = append(comments, c)
					}
				}
			}
			return loadContentMsg{
				id:          item.ID,
				title:       item.Title,
				content:     item.Content,
				contentType: item.ContentType,
				comments:    comments,
			}
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
			TargetType: msg.targetType,
			TargetRef:  msg.path,
			LineStart:  msg.lineStart,
			LineEnd:    msg.lineEnd,
		}

		if msg.editingID != "" {
			_, _ = m.engine.EditComment(msg.editingID, msg.body)
		} else {
			_, _ = m.engine.AddComment(target, msg.commentType, msg.body)
		}

		// Content items: reload as content, not as diff
		if msg.targetType == types.TargetContent {
			item, err := m.engine.GetContentItem(msg.path)
			if err != nil || item == nil {
				return loadContentMsg{id: msg.path}
			}
			session := m.engine.GetSession()
			var comments []types.ReviewComment
			if session != nil {
				for _, c := range session.Comments {
					if c.TargetRef == item.ID && c.TargetType == types.TargetContent {
						comments = append(comments, c)
					}
				}
			}
			return loadContentMsg{
				id:          item.ID,
				title:       item.Title,
				content:     item.Content,
				contentType: item.ContentType,
				comments:    comments,
			}
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
	var filePath string
	var reviewed bool

	switch m.focus {
	case focusSidebar:
		file := m.sidebar.selectedFile()
		if file == nil {
			return nil
		}
		filePath = file.Path
		reviewed = file.Reviewed
	case focusMain:
		if m.diffView.path == "" || m.diffView.contentMode {
			return nil
		}
		filePath = m.diffView.path
		for _, f := range m.sidebar.files {
			if f.Path == filePath {
				reviewed = f.Reviewed
				break
			}
		}
	default:
		return nil
	}
	return func() tea.Msg {
		if reviewed {
			_ = m.engine.UnmarkReviewed(filePath)
		} else {
			_ = m.engine.MarkReviewed(filePath)
		}
		return fileChangedMsg{path: filePath}
	}
}

// refreshFiles returns a Cmd that refreshes the file list and current diff from git.
func (m appModel) refreshFiles() tea.Cmd {
	engine := m.engine
	currentPath := m.diffView.path
	inContentMode := m.diffView.contentMode
	return func() tea.Msg {
		// Refresh the file list from git
		files, err := engine.RefreshChangedFiles()
		if err != nil {
			return nil
		}
		session := engine.GetSession()

		// Don't reload diff when viewing a content item — it's not a file
		var result *types.DiffResult
		var comments []types.ReviewComment
		if currentPath != "" && !inContentMode {
			result, _ = engine.GetFileDiff(currentPath)
			if session != nil {
				for _, c := range session.Comments {
					if c.TargetRef == currentPath {
						comments = append(comments, c)
					}
				}
			}
		}

		return refreshResultMsg{
			files:    files,
			path:     currentPath,
			result:   result,
			comments: comments,
		}
	}
}

type refreshResultMsg struct {
	files    []types.ChangedFile
	path     string
	result   *types.DiffResult
	comments []types.ReviewComment
}

// loadContentMsg carries content item data for rendering in the diff view.
type loadContentMsg struct {
	id          string
	title       string
	content     string
	contentType string
	comments    []types.ReviewComment
}

// View renders the full TUI layout.
func (m appModel) View() tea.View {
	// Title bar
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("4")).Render(" o_(◉) monocle")
	titleBar := lipgloss.NewStyle().Width(m.width).Render(title)

	sidebarStyle := m.theme.SidebarBorder
	if m.focus == focusSidebar {
		sidebarStyle = m.theme.SidebarBorderFocused
	}
	mainStyle := m.theme.MainPane
	if m.focus == focusMain {
		mainStyle = m.theme.MainPaneFocused
	}

	var body string

	// lipgloss v2: Width/Height set the OUTER dimensions (including border).
	// Our content dimensions (sidebar.width, diffView.width, etc.) are the
	// inner content size, so we add borderW/borderH to get the outer size.
	const bw = 2 // border left + right
	const bh = 2 // border top + bottom

	if m.layout == layoutStacked {
		sidebarView := sidebarStyle.
			Width(m.sidebar.width + bw).
			Height(m.sidebar.height + bh).
			Render(m.sidebar.View())

		mainView := mainStyle.
			Width(m.diffView.width + bw).
			Height(m.diffView.height + bh).
			Render(m.diffView.View())

		body = lipgloss.JoinVertical(lipgloss.Left, sidebarView, mainView)
	} else {
		sidebarView := sidebarStyle.
			Width(m.sidebar.width + bw).
			Height(m.sidebar.height + bh).
			Render(m.sidebar.View())

		// Measure actual rendered sidebar width and give diff view the rest
		sidebarRenderedW := lipgloss.Width(sidebarView)
		diffOuterW := m.width - sidebarRenderedW
		diffContentW := diffOuterW - bw
		if diffContentW < 0 {
			diffContentW = 0
		}
		m.diffView.width = diffContentW

		mainView := mainStyle.
			Width(diffOuterW).
			Height(m.diffView.height + bh).
			Render(m.diffView.View())

		body = lipgloss.JoinHorizontal(lipgloss.Top, sidebarView, mainView)
	}
	m.statusBar.width = m.width
	statusView := m.statusBar.View()
	full := lipgloss.JoinVertical(lipgloss.Left, titleBar, body, statusView)

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
	} else if m.overlay == overlayHelp {
		overlayContent := m.help.View()
		if overlayContent != "" {
			full = overlayOn(full, overlayContent, m.width, m.height)
		}
	} else if m.overlay == overlayRefPicker {
		overlayContent := m.refPicker.View()
		if overlayContent != "" {
			full = overlayOn(full, overlayContent, m.width, m.height)
		}
	} else if m.overlay == overlayConfirm {
		overlayContent := m.confirm.View()
		if overlayContent != "" {
			full = overlayOn(full, overlayContent, m.width, m.height)
		}
	} else if m.overlay == overlayInstallPrompt {
		overlayContent := m.installPrompt.View()
		if overlayContent != "" {
			full = overlayOn(full, overlayContent, m.width, m.height)
		}
	} else if m.overlay == overlayConnectionInfo {
		overlayContent := m.connectionInfo.View()
		if overlayContent != "" {
			full = overlayOn(full, overlayContent, m.width, m.height)
		}
	}

	v := tea.NewView(full)
	v.AltScreen = true
	return v
}

// overlayOn centers overlay content over base content, preserving base content
// (including borders and styling) on both sides of the overlay.
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
	if topPad < 2 {
		topPad = 2
	}
	leftPad := (width - overlayW) / 2
	if leftPad < 0 {
		leftPad = 0
	}

	baseLines := strings.Split(base, "\n")
	result := make([]string, len(baseLines))
	copy(result, baseLines)

	for i, oLine := range overlayLines {
		baseIdx := topPad + i
		if baseIdx >= len(result) {
			break
		}

		baseLine := result[baseIdx]

		// Left: preserve base content before overlay
		leftPart := ansi.Cut(baseLine, 0, leftPad)
		if leftW := lipgloss.Width(leftPart); leftW < leftPad {
			leftPart += strings.Repeat(" ", leftPad-leftW)
		}

		// Right: preserve base content after overlay
		rightPart := ansi.TruncateLeft(baseLine, leftPad+overlayW, "")

		result[baseIdx] = leftPart + oLine + rightPart
	}
	return strings.Join(result, "\n")
}

// calcModalWidth computes modal width as max(screenWidth*2/3, 65), capped by
// maxWidth (pass 0 for no cap) and screen bounds (screenWidth-10 for margin).
func calcModalWidth(screenWidth, maxWidth int) int {
	w := screenWidth * 2 / 3
	if w < 65 {
		w = 65
	}
	if maxWidth > 0 && w > maxWidth {
		w = maxWidth
	}
	if w > screenWidth-10 {
		w = screenWidth - 10
	}
	if w < 0 {
		w = 0
	}
	return w
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
	engine.On(core.EventPauseChanged, func(e core.EventPayload) {
		p.Send(pauseChangedMsg{status: e.Status})
	})
	engine.On(core.EventConnectionChanged, func(e core.EventPayload) {
		count, _ := strconv.Atoi(e.Status)
		p.Send(connectionChangedMsg{count: count})
	})
}

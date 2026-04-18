// app.go is the main Bubble Tea application model and event loop.
package tui

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/josephschmitt/monocle/internal/clipboard"
	"github.com/josephschmitt/monocle/internal/core"
	"github.com/josephschmitt/monocle/internal/types"
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

const defaultMinDiffWidth = 80

// overlayKind identifies which (if any) overlay is shown.
type overlayKind int

const (
	overlayNone overlayKind = iota
	overlayComment
	overlayReview
	overlayHelp
	overlayRefPicker
	overlayConfirm
	overlayRegisterPrompt
	overlayConnectionInfo
	overlayHistory
	overlaySessionPicker
	overlayInfo
	overlayVersionPicker
)

// Engine event messages bridged from core.EngineAPI callbacks.

type fileChangedMsg struct {
	path    string
	advance bool // auto-advance to next unreviewed item
}

type submitErrorMsg struct{}

type feedbackStatusMsg struct {
	status string
}

type contentItemMsg struct {
	id string
}

type additionalFileAddedMsg struct {
	path    string
	advance bool // auto-advance to next unreviewed item
}

type connectionChangedMsg struct {
	count     int
	agentName string
	mode      string // "queue" for queue-mode connections, subscriber count for push
}

// requestContentDiffMsg requests async computation of a content item diff.
type requestContentDiffMsg struct {
	contentID      string
	preferredStyle diffStyle // style to use when diff loads (0=unified for manual cycle)
}

// loadContentDiffMsg carries the computed content diff result.
type loadContentDiffMsg struct {
	contentID      string
	result         *types.DiffResult
	comments       []types.ReviewComment
	err            error
	preferredStyle diffStyle // style to render the diff in
	fromVersion    int       // base version for the diff (0 = default latest-vs-previous)
	toVersion      int       // target version for the diff
}

type feedbackPickedUpMsg struct{}

type pauseChangedMsg struct {
	status string
}

type waitStatusMsg struct {
	waiting bool
}

type contentReviewedMsg struct {
	id      string
	advance bool // auto-advance to next unreviewed item
}

type editCommentMsg struct {
	comment *types.ReviewComment
}

type deleteCommentMsg struct {
	commentID string
}

type submitSuccessMsg struct {
	agentConnected bool // whether an agent was connected to receive the review
}

type commentsClearedMsg struct {
	reloadPath       string
	isContent        bool
	isAdditionalFile bool
}

type reviewClearedMsg struct {
	reloadPath       string
	isContent        bool
	isAdditionalFile bool
}

type artifactDismissedMsg struct {
	id  string
	err error
}

type resolveCommentMsg struct {
	commentID string
}

type openConfirmMsg struct {
	title   string
	message string
	action  confirmAction
}

type mcpRegisterPromptMsg struct{}

type openInfoBannerMsg struct {
	title   string
	message string
}

type mcpRegisterResultMsg struct {
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
	MCPRegisterFn     func(global bool) error // if non-nil, offer MCP auto-registration on startup
	ShowSessionPicker bool                    // if true, show session picker modal on startup
	RepoRoot          string                  // repo root path, used by session picker to list sessions
	NonGitMode        bool                    // if true, directory mode (no git, show file contents instead of diffs)
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
	history        historyModel
	sessionPicker  sessionPickerModel
	versionPicker  versionPickerModel

	focus         focusTarget
	overlay       overlayKind
	layout        layoutMode
	layoutConfig  string
	sidebarHidden     bool
	sidebarAutoHidden bool // true when sidebar was hidden due to empty state (not user action)
	sidebarUserShown  bool // true when user explicitly showed the sidebar (prevents auto-hide)

	commandMode   bool
	commandBuffer string

	width  int
	height int

	theme Theme
	keys  KeyMap

	mcpRegisterFn    func(global bool) error
	registerPrompt   registerPromptModel

	pendingDismissArtifactID string // set while the dismiss-artifact confirm modal is open

	focusModeActive       bool // currently in focus mode
	focusModeSavedSidebar bool // sidebar visibility before entering focus mode
	focusModeSavedWrap    bool // wrap state before entering focus mode

	mouseEnabled bool // whether mouse mode is active
	minDiffWidth int  // minimum diff viewer content width in horizontal layout

	showSessionPicker bool   // open session picker on startup
	repoRoot          string // repo root for session listing

	nonGitMode bool             // directory mode (no git)
	infoBanner infoBannerModel  // info modal for non-git startup
}

// NewApp creates the root appModel and wires up all subsystems.
func NewApp(engine core.EngineAPI, opts ...AppOptions) appModel {
	var o AppOptions
	if len(opts) > 0 {
		o = opts[0]
	}

	theme := DefaultTheme()
	keys := DefaultKeyMap()
	sidebar := newSidebarModel(&keys)
	sidebar.focused = true
	help := newHelpModel(theme, &keys)
	dv := newDiffViewModel(&theme, &keys)
	var layoutCfg string

	mouseEnabled := true
	minDiffW := defaultMinDiffWidth
	if engine != nil {
		if cfg := engine.GetConfig(); cfg != nil {
			if cfg.Keybindings != nil {
				keys = keys.ApplyOverrides(cfg.Keybindings)
			}
			if cfg.SidebarStyle == "tree" {
				sidebar.treeMode = true
			}
			switch cfg.DiffStyle {
			case "split":
				dv.style = diffStyleSplit
				dv.preferredContentDiffStyle = diffStyleSplit
				dv.autoContentDiff = true
			case "file":
				dv.style = diffStyleFile
				// autoContentDiff stays false: "file" mode = no auto-switch for plans
			default:
				dv.autoContentDiff = true
				// preferredContentDiffStyle stays at diffStyleUnified (zero value)
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
			if cfg.Mouse != nil && !*cfg.Mouse {
				mouseEnabled = false
			}
			if cfg.MinDiffWidth > 0 {
				minDiffW = cfg.MinDiffWidth
			}
			if cfg.CommentExpand != nil && !*cfg.CommentExpand {
				dv.commentExpandDelay = -1
			} else if cfg.CommentExpandDelay > 0 {
				dv.commentExpandDelay = time.Duration(cfg.CommentExpandDelay) * time.Millisecond
			}
			sidebar.reviewTracking = cfg.ReviewTracking
			help.reviewTracking = cfg.ReviewTracking
		}
	}

	if o.NonGitMode {
		dv.style = diffStyleFile
	}

	return appModel{
		engine:        engine,
		sidebar:       sidebar,
		diffView:      dv,
		statusBar:     newStatusBarModel(theme),
		commentEditor: newCommentEditorModel(theme),
		reviewSummary: newReviewSummaryModel(theme),
		help:          help,
		refPicker:     newRefPickerModel(theme),
		confirm:        newConfirmModel(theme),
		connectionInfo: newConnectionInfoModel(theme),
		history:        newHistoryModel(theme),
		sessionPicker:  newSessionPickerModel(theme),
		versionPicker:  newVersionPickerModel(theme),
		registerPrompt: newRegisterPromptModel(theme),
		infoBanner:     newInfoBannerModel(theme),
		focus:         focusSidebar,
		overlay:       overlayNone,
		layoutConfig:  layoutCfg,
		theme:         theme,
		keys:          keys,
		mcpRegisterFn:     o.MCPRegisterFn,
		mouseEnabled:      mouseEnabled,
		minDiffWidth:      minDiffW,
		showSessionPicker: o.ShowSessionPicker,
		repoRoot:          o.RepoRoot,
		nonGitMode:        o.NonGitMode,
	}
}

// Init loads the initial file list from the engine, starts the refresh tick, and kicks off the TUI event loop.
func (m appModel) Init() tea.Cmd {
	cmds := []tea.Cmd{refreshTick()}

	if m.showSessionPicker {
		// Defer file loading until user picks a session
		engine := m.engine
		repoRoot := m.repoRoot
		startCmd := m.startSessionAndLoad()
		cmds = append(cmds, func() tea.Msg {
			sessions, err := engine.ListSessions(core.ListSessionsOptions{
				RepoRoot: repoRoot,
				Limit:    20,
			})
			if err != nil || len(sessions) == 0 {
				// No sessions to pick from — create a new one
				return startCmd()
			}
			return openSessionPickerMsg{sessions: sessions}
		})
	} else {
		cmds = append(cmds, func() tea.Msg {
			files := m.engine.GetChangedFiles()
			items := m.engine.GetContentItems()
			additional := m.engine.GetAdditionalFiles()
			return initialLoadMsg{files: files, items: items, additionalFiles: additional}
		})
	}

	if m.mcpRegisterFn != nil {
		cmds = append(cmds, func() tea.Msg {
			return mcpRegisterPromptMsg{}
		})
	}
	if m.nonGitMode {
		cmds = append(cmds, func() tea.Msg {
			return openInfoBannerMsg{
				title:   "Directory Mode",
				message: "This directory is not a Git repository.\nMonocle will display file contents instead of diffs.",
			}
		})
	}
	return tea.Batch(cmds...)
}

// startSessionAndLoad creates a new session on the engine and returns a cmd
// that loads the initial file/content lists. Used when the session picker
// opens with no prior sessions or the user cancels.
func (m appModel) startSessionAndLoad() tea.Cmd {
	engine := m.engine
	repoRoot := m.repoRoot
	return func() tea.Msg {
		engine.StartSession(core.SessionOptions{RepoRoot: repoRoot})
		files := engine.GetChangedFiles()
		items := engine.GetContentItems()
		additional := engine.GetAdditionalFiles()
		return initialLoadMsg{files: files, items: items, additionalFiles: additional}
	}
}

// resumeSessionAndLoad resumes an existing session and loads its data.
func (m appModel) resumeSessionAndLoad(sessionID string) tea.Cmd {
	engine := m.engine
	return func() tea.Msg {
		if _, err := engine.ResumeSession(sessionID); err != nil {
			return nil
		}
		files := engine.GetChangedFiles()
		items := engine.GetContentItems()
		additional := engine.GetAdditionalFiles()
		return initialLoadMsg{files: files, items: items, additionalFiles: additional}
	}
}

// initialLoadMsg carries the initial file and content item lists.
type initialLoadMsg struct {
	files           []types.ChangedFile
	items           []types.ContentItem
	additionalFiles []types.AdditionalFile
}

// Update handles all incoming messages and routes them appropriately.
func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		recalcPaneDimensions(&m)

		m.statusBar.width = m.width
		m.commentEditor.width = m.width
		m.commentEditor.height = m.height
		m.reviewSummary.width = m.width
		m.reviewSummary.height = m.height
		m.help.width = m.width
		m.help.height = m.height
		m.confirm.width = m.width
		m.confirm.height = m.height
		m.registerPrompt.width = m.width
		m.registerPrompt.height = m.height
		m.connectionInfo.width = m.width
		m.connectionInfo.height = m.height
		m.sessionPicker.width = m.width
		m.sessionPicker.height = m.height
		m.infoBanner.width = m.width
		m.infoBanner.height = m.height
		return m, nil

	case initialLoadMsg:
		m.sidebar.files = msg.files
		m.sidebar.contentItems = msg.items
		m.sidebar.additionalFiles = msg.additionalFiles
		m.sidebar.applyReviewedFilter()
		m.sidebar.rebuildTree()
		m.sidebar.clampOffset()
		recalcStackedLayout(&m)
		// Sync status bar file count
		session := m.engine.GetSession()
		if session != nil {
			m.statusBar.baseRef = m.displayBaseRef(session)
			m.statusBar.agentName = session.Agent
		}
		m.statusBar.fileCount = len(msg.files)
		m.statusBar.socketStarted = m.engine.GetSocketPath() != ""
		m.statusBar.subscriberCount = m.engine.GetSubscriberCount()
		m.statusBar.feedbackStatus = m.engine.GetFeedbackStatus()
		// Auto-hide sidebar when empty, auto-show when items arrive
		if m.autoToggleSidebar() {
			recalcPaneDimensions(&m)
		}
		// Auto-select the first file, or first content item if no files
		if len(msg.files) > 0 {
			m.sidebar.selectPath(msg.files[0].Path)
			return m, m.handleSidebarSelect(sidebarSelectMsg{path: msg.files[0].Path})
		}
		if len(msg.items) > 0 {
			m.sidebar.selectContentByID(msg.items[0].ID)
			return m, m.handleSidebarSelect(sidebarSelectMsg{isContent: true, contentID: msg.items[0].ID})
		}
		if len(msg.additionalFiles) > 0 {
			return m, m.handleSidebarSelect(sidebarSelectMsg{path: msg.additionalFiles[0].Path, isAdditionalFile: true})
		}
		return m, nil

	// Periodic refresh — fires on a timer to keep the file list and diff in sync.
	case refreshTickMsg:
		return m, tea.Batch(m.refreshFiles(), refreshTick())

	case refreshResultMsg:
		prevKind, prevID := m.sidebar.currentItemKey()

		m.sidebar.files = msg.files
		m.sidebar.applyReviewedFilter()
		m.sidebar.rebuildTree()

		m.sidebar.selectByKey(prevKind, prevID)
		m.sidebar.clampOffset()
		recalcStackedLayout(&m)
		m.statusBar.fileCount = len(msg.files)
		if m.autoToggleSidebar() {
			recalcPaneDimensions(&m)
		}
		var diffCmd tea.Cmd
		if msg.contentItem != nil && m.diffView.isViewingContentItem() && m.diffView.contentID == msg.contentItem.ID {
			// Only update if content or comments actually changed to avoid
			// flicker from the content→auto-switch→diff cycle on every refresh tick.
			if contentItemChanged(msg.contentItem, msg.contentComments, &m.diffView) {
				m.diffView, diffCmd = m.diffView.Update(loadContentMsg{
					id:                 msg.contentItem.ID,
					title:              msg.contentItem.Title,
					content:            msg.contentItem.Content,
					contentType:        msg.contentItem.ContentType,
					comments:           msg.contentComments,
					versionCount:   msg.contentItem.VersionCount,
					autoSwitchDiff: msg.contentItem.VersionCount > 1 && m.diffView.contentMode,
				})
			}
		} else if msg.path != "" && msg.result != nil && msg.path == m.diffView.path {
			if fileDiffChanged(msg.result, msg.comments, &m.diffView) {
				m.diffView, diffCmd = m.diffView.Update(loadDiffMsg{
					path:     msg.path,
					result:   msg.result,
					comments: msg.comments,
				})
			}
		}
		// Auto-select first file if current view is stale
		if len(msg.files) > 0 && !m.diffViewShowsValidFile() {
			m.sidebar.selectPath(msg.files[0].Path)
			return m, m.handleSidebarSelect(sidebarSelectMsg{path: msg.files[0].Path})
		} else if len(msg.files) == 0 && !m.diffView.isViewingContentItem() && m.diffView.path != "" {
			m.diffView.clearFileState()
		}
		return m, diffCmd

	// Engine events
	case fileChangedMsg:
		m.sidebar.files = m.engine.GetChangedFiles()
		m.sidebar.applyReviewedFilter()
		m.sidebar.rebuildTree()
		m.sidebar.clampOffset()
		recalcStackedLayout(&m)
		m.statusBar.fileCount = len(m.sidebar.files)
		if m.autoToggleSidebar() {
			recalcPaneDimensions(&m)
		}
		session := m.engine.GetSession()
		if session != nil {
			m.statusBar.baseRef = m.displayBaseRef(session)
			m.statusBar.commentCount = len(session.Comments)
		}
		// Auto-advance to next unreviewed item after marking reviewed
		if msg.advance {
			if cmd := m.sidebar.nextUnreviewed(); cmd != nil {
				return m, cmd
			}
			return m, nil
		}
		// Auto-select first file if the current view is empty or stale
		if len(m.sidebar.files) > 0 && !m.diffViewShowsValidFile() {
			m.sidebar.selectPath(m.sidebar.files[0].Path)
			return m, m.handleSidebarSelect(sidebarSelectMsg{path: m.sidebar.files[0].Path})
		} else if len(m.sidebar.files) == 0 && !m.diffView.isViewingContentItem() && m.diffView.path != "" {
			m.diffView.clearFileState()
		}
		return m, nil

	case contentReviewedMsg:
		m.sidebar.contentItems = m.engine.GetContentItems()
		m.sidebar.applyReviewedFilter()
		m.sidebar.clampOffset()
		// Auto-advance to next unreviewed item after marking reviewed
		if msg.advance {
			if cmd := m.sidebar.nextUnreviewed(); cmd != nil {
				return m, cmd
			}
		}
		return m, nil

	case connectionChangedMsg:
		m.statusBar.subscriberCount = msg.count
		m.statusBar.connectionMode = msg.mode
		m.statusBar.socketStarted = m.engine.GetSocketPath() != ""
		if msg.agentName != "" {
			m.statusBar.agentName = msg.agentName
		}
		if msg.count == 0 && msg.mode == "" {
			m.statusBar.agentName = ""
		}
		m.reviewSummary.agentConnected = msg.count > 0 || msg.mode == "queue"
		return m, nil

	case feedbackStatusMsg:
		m.statusBar.feedbackStatus = msg.status
		return m, nil

	case contentItemMsg:
		m.sidebar.contentItems = m.engine.GetContentItems()
		m.sidebar.applyReviewedFilter()
		m.sidebar.rebuildTree()
		m.sidebar.clampOffset()
		recalcStackedLayout(&m)
		if m.autoToggleSidebar() {
			recalcPaneDimensions(&m)
		}
		// Auto-enter focus mode if enabled and this is a plan
		if !m.focusModeActive && msg.id != "" {
			if cfg := m.engine.GetConfig(); cfg != nil && cfg.AutoFocusMode {
				if item, err := m.engine.GetContentItem(msg.id); err == nil && item != nil && item.IsPlan {
					m.focusModeSavedSidebar = m.sidebarHidden
					m.focusModeSavedWrap = m.diffView.wrap
					m.sidebarHidden = true
					m.sidebarAutoHidden = false // focus mode owns the hidden state
					m.diffView.wrap = true
					m.diffView.hOffset = 0
					m.focus = focusMain
					m.sidebar.focused = false
					m.diffView.focused = true
					m.focusModeActive = true
					m.sidebar.selectContentByID(msg.id)
					selectCmd := m.handleSidebarSelect(sidebarSelectMsg{isContent: true, contentID: msg.id})
					resizeCmd := func() tea.Msg {
						return tea.WindowSizeMsg{Width: m.width, Height: m.height}
					}
					return m, tea.Batch(selectCmd, resizeCmd)
				}
			}
		}
		// Auto-select the latest content item in sidebar and viewer
		if msg.id != "" {
			m.sidebar.selectContentByID(msg.id)
			return m, m.handleSidebarSelect(sidebarSelectMsg{isContent: true, contentID: msg.id})
		}
		return m, nil

	case additionalFileAddedMsg:
		m.sidebar.additionalFiles = m.engine.GetAdditionalFiles()
		m.sidebar.applyReviewedFilter()
		m.sidebar.clampOffset()
		recalcStackedLayout(&m)
		if m.autoToggleSidebar() {
			recalcPaneDimensions(&m)
		}
		// Auto-advance to next unreviewed item after marking reviewed
		if msg.advance {
			if cmd := m.sidebar.nextUnreviewed(); cmd != nil {
				return m, cmd
			}
			return m, nil
		}
		// Auto-select only if nothing else is showing
		if m.diffView.path == "" && len(m.sidebar.files) == 0 && len(m.sidebar.contentItems) == 0 && msg.path != "" {
			return m, m.handleSidebarSelect(sidebarSelectMsg{path: msg.path, isAdditionalFile: true})
		}
		return m, nil

	case pauseChangedMsg:
		return m, nil

	case waitStatusMsg:
		m.statusBar.waitingForReview = msg.waiting
		return m, nil

	case baseRefChangedMsg:
		if session := m.engine.GetSession(); session != nil {
			m.statusBar.baseRef = m.displayBaseRef(session)
		}
		return m, m.refreshFiles()

	case openRefPickerMsg:
		m.refPicker.entries = msg.entries
		m.refPicker.snapshots = msg.snapshots
		m.refPicker.autoActive = msg.autoActive
		m.refPicker.active = true
		m.refPicker.width = m.width
		m.refPicker.height = m.height
		m.refPicker.offset = 0
		m.refPicker.hasMore = len(msg.entries) >= refPickerPageSize
		m.refPicker.loading = false

		// Pre-select the currently active ref or snapshot
		m.refPicker.cursor = m.refPicker.workingTreeCursor()
		m.refPicker.snapshotActive = false
		m.refPicker.activeSnapshotID = 0
		if snap := m.engine.GetActiveSnapshot(); snap != nil {
			m.refPicker.snapshotActive = true
			m.refPicker.activeSnapshotID = snap.ID
			// Find the active snapshot in the list
			for i, s := range msg.snapshots {
				if s.ID == snap.ID {
					m.refPicker.cursor = i
					break
				}
			}
		} else if !msg.autoActive {
			// Use the user's selected ref for matching (not the parent used for diffing)
			displayRef := m.engine.SelectedBaseRef()
			if displayRef == "" {
				if session := m.engine.GetSession(); session != nil {
					displayRef = session.BaseRef
				}
			}
			if displayRef != "" {
				commitStart := m.refPicker.commitStartCursor()
				for i, entry := range msg.entries {
					if strings.HasPrefix(entry.Hash, displayRef) || strings.HasPrefix(displayRef, entry.Hash) {
						m.refPicker.cursor = commitStart + i
						break
					}
				}
			}
		}
		m.refPicker.ensureVisible()
		m.overlay = overlayRefPicker
		return m, nil

	case selectRefMsg:
		m.overlay = overlayNone
		m.refPicker.active = false
		// Selecting a git ref clears snapshot mode
		m.engine.ClearSnapshotBase()
		if msg.auto {
			return m, m.executeCommand("ref auto")
		}
		return m, m.executeCommand("ref " + msg.hash)

	case selectSnapshotMsg:
		m.overlay = overlayNone
		m.refPicker.active = false
		if err := m.engine.SetSnapshotBase(msg.snapshotID); err != nil {
			return m, nil
		}
		// Refresh to show snapshot-based diff
		session := m.engine.GetSession()
		if session != nil {
			m.statusBar.baseRef = m.displayBaseRef(session)
		}
		return m, m.refreshFiles()

	case loadMoreRefsMsg:
		m.refPicker, _ = m.refPicker.Update(msg)
		return m, nil

	case cancelRefPickerMsg:
		m.overlay = overlayNone
		m.refPicker.active = false
		return m, nil

	case openVersionPickerMsg:
		// Reverse to newest-first order (matching ref picker convention)
		reversed := make([]types.ContentVersion, len(msg.versions))
		for i, v := range msg.versions {
			reversed[len(msg.versions)-1-i] = v
		}
		m.versionPicker.versions = reversed
		m.versionPicker.contentID = msg.contentID
		m.versionPicker.active = true
		m.versionPicker.width = m.width
		m.versionPicker.height = m.height
		m.versionPicker.offset = 0
		m.versionPicker.cursor = 1 // start at first selectable (index 0 is current)
		m.versionPicker.ensureVisible()
		m.overlay = overlayVersionPicker
		return m, nil

	case selectVersionMsg:
		m.overlay = overlayNone
		m.versionPicker.active = false
		engine := m.engine
		contentID := msg.contentID
		fromVersion := msg.fromVersion
		toVersion := msg.toVersion
		return m, func() tea.Msg {
			result, err := engine.GetContentDiffBetweenVersions(contentID, fromVersion, toVersion)
			if err != nil || result == nil {
				return nil
			}
			item, err := engine.GetContentItem(contentID)
			if err != nil {
				return nil
			}
			return loadContentDiffMsg{
				contentID:      contentID,
				result:         result,
				comments:       item.Comments,
				preferredStyle: diffStyleUnified,
				fromVersion:    fromVersion,
				toVersion:      toVersion,
			}
		}

	case cancelVersionPickerMsg:
		m.overlay = overlayNone
		m.versionPicker.active = false
		return m, nil

	case openSessionPickerMsg:
		m.sessionPicker.sessions = msg.sessions
		m.sessionPicker.active = true
		m.sessionPicker.width = m.width
		m.sessionPicker.height = m.height
		m.sessionPicker.cursor = 0
		m.sessionPicker.offset = 0
		m.overlay = overlaySessionPicker
		return m, nil

	case selectSessionMsg:
		m.overlay = overlayNone
		m.sessionPicker.active = false
		if msg.id == "" {
			return m, m.startSessionAndLoad()
		}
		return m, m.resumeSessionAndLoad(msg.id)

	case cancelSessionPickerMsg:
		m.overlay = overlayNone
		m.sessionPicker.active = false
		return m, m.startSessionAndLoad()

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

	// Content diff request (from diff style cycle on content items)
	case requestContentDiffMsg:
		engine := m.engine
		contentID := msg.contentID
		preferredStyle := msg.preferredStyle
		return m, func() tea.Msg {
			result, err := engine.GetContentDiff(contentID)
			if err != nil || result == nil {
				return loadContentDiffMsg{contentID: contentID, err: err}
			}
			session := engine.GetSession()
			var comments []types.ReviewComment
			if session != nil {
				for _, c := range session.Comments {
					if c.TargetRef == contentID && c.TargetType == types.TargetContent {
						comments = append(comments, c)
					}
				}
			}
			return loadContentDiffMsg{contentID: contentID, result: result, comments: comments, preferredStyle: preferredStyle}
		}

	// Content diff loaded
	case loadContentDiffMsg:
		var cmd tea.Cmd
		m.diffView, cmd = m.diffView.Update(msg)
		return m, cmd

	// File content request (from diff style cycle)
	case requestFileContentMsg:
		engine := m.engine
		path := msg.path
		selectID := msg.selectCommentID
		return m, func() tea.Msg {
			content, err := engine.GetFileContent(path)
			if err != nil {
				return loadFileContentMsg{path: path, err: err}
			}
			session := engine.GetSession()
			var comments []types.ReviewComment
			if session != nil {
				for _, c := range session.Comments {
					if c.TargetRef == path {
						comments = append(comments, c)
					}
				}
			}
			return loadFileContentMsg{
				path:            path,
				content:         content,
				comments:        comments,
				selectCommentID: selectID,
			}
		}

	// File content loaded
	case loadFileContentMsg:
		var cmd tea.Cmd
		m.diffView, cmd = m.diffView.Update(msg)
		return m, cmd

	// Additional file loaded
	case loadAdditionalFileMsg:
		var cmd tea.Cmd
		m.diffView, cmd = m.diffView.Update(msg)
		return m, cmd

	// Sidebar selection → load diff (focus stays where it is)
	case sidebarSelectMsg:
		return m, m.handleSidebarSelect(msg)

	// Comment overlay
	case openCommentMsg:
		if msg.prefillBody != "" {
			m.commentEditor.openSuggest(msg.path, msg.lineStart, msg.lineEnd, msg.targetType, msg.prefillBody, msg.prefillType)
		} else {
			m.commentEditor.open(msg.path, msg.lineStart, msg.lineEnd, msg.targetType)
		}
		m.overlay = overlayComment
		return m, nil

	case editCommentMsg:
		m.commentEditor.openEdit(msg.comment)
		m.overlay = overlayComment
		return m, nil

	case deleteCommentMsg:
		engine := m.engine
		id := msg.commentID
		currentPath := m.diffView.path
		isContent := m.diffView.contentMode
		additionalPath := m.diffView.additionalFilePath
		return m, func() tea.Msg {
			_ = engine.DeleteComment(id)
			if isContent {
				return fileChangedMsg{}
			}
			if additionalPath != "" {
				content, err := engine.GetAdditionalFileContent(additionalPath)
				if err != nil {
					return loadAdditionalFileMsg{path: additionalPath, err: err}
				}
				session := engine.GetSession()
				var comments []types.ReviewComment
				if session != nil {
					for _, c := range session.Comments {
						if c.TargetRef == additionalPath && c.TargetType == types.TargetAdditionalFile {
							comments = append(comments, c)
						}
					}
				}
				return loadAdditionalFileMsg{path: additionalPath, content: content, comments: comments}
			}
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

	case saveCommentMsg:
		m.overlay = overlayNone
		m.diffView.visualMode = false
		return m, m.handleSaveComment(msg)

	case cancelCommentMsg:
		m.overlay = overlayNone
		return m, nil

	case externalEditorRequestMsg:
		return m, openExternalEditor(msg.body, msg.origin)

	case externalEditorResultMsg:
		if msg.err != nil {
			return m, nil
		}
		switch msg.origin {
		case overlayComment:
			m.commentEditor.body = msg.body
			m.commentEditor.cursor = len([]rune(msg.body))
		case overlayReview:
			m.reviewSummary.body = msg.body
		}
		return m, nil

	case openFileInEditorDoneMsg:
		return m, m.refreshFiles()

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
		additionalPath := m.diffView.additionalFilePath
		return m, func() tea.Msg {
			_ = engine.ResolveComment(id)
			if isContent {
				return fileChangedMsg{}
			}
			if additionalPath != "" {
				content, err := engine.GetAdditionalFileContent(additionalPath)
				if err != nil {
					return loadAdditionalFileMsg{path: additionalPath, err: err}
				}
				session := engine.GetSession()
				var comments []types.ReviewComment
				if session != nil {
					for _, c := range session.Comments {
						if c.TargetRef == additionalPath && c.TargetType == types.TargetAdditionalFile {
							comments = append(comments, c)
						}
					}
				}
				return loadAdditionalFileMsg{path: additionalPath, content: content, comments: comments}
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
		m.reviewSummary.open(msg.summary, msg.agentConnected)
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
			result, err := engine.Submit(action, body)
			if err != nil {
				return submitErrorMsg{}
			}
			if copyToClip {
				if text, err := engine.FormatReview(action, body); err == nil {
					clipboard.Copy(text)
				}
			}
			return submitSuccessMsg{agentConnected: result.AgentConnected}
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

	case submitErrorMsg:
		m.statusBar.feedbackStatus = "submit_failed"
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
			m.statusBar.fileCount = len(session.ChangedFiles)
			m.statusBar.baseRef = m.displayBaseRef(session)
		}

		// If no push-mode agent was connected, feedback is queued for pull delivery.
		// Clear comments (they're frozen in the submission) but don't advance
		// the round or clear content items — that happens when the agent pulls.
		if !msg.agentConnected {
			count := m.engine.GetQueuedCount()
			if count == 1 {
				m.statusBar.feedbackStatus = "1 review queued"
			} else {
				m.statusBar.feedbackStatus = fmt.Sprintf("%d reviews queued", count)
			}

			// Clear comments — they're now frozen in the submission record
			if session != nil && len(session.Comments) > 0 {
				_ = m.engine.ClearComments()
				m.statusBar.commentCount = 0
			}

			// Reload diff view to remove inline comment annotations
			if m.diffView.path != "" {
				m.diffView.comments = nil
			}

			// Restore focus mode state
			if m.focusModeActive {
				m.sidebarHidden = m.focusModeSavedSidebar
				m.diffView.wrap = m.focusModeSavedWrap
				m.focusModeActive = false
				m.sidebarAutoHidden = m.sidebarHidden && !m.sidebarHasItems()
				m.autoToggleSidebar()
				recalcPaneDimensions(&m)
			}
			return m, nil
		}

		m.syncArtifactsAfterSubmit(session)

		if session != nil && len(session.Comments) > 0 {
			_ = m.engine.ClearComments()
			m.statusBar.commentCount = 0
		}

		// Restore focus mode state
		if m.focusModeActive {
			m.sidebarHidden = m.focusModeSavedSidebar
			m.diffView.wrap = m.focusModeSavedWrap
			m.focusModeActive = false
			m.sidebarAutoHidden = m.sidebarHidden && !m.sidebarHasItems()
			m.autoToggleSidebar()
			recalcPaneDimensions(&m)
			return m, nil
		}

		recalcStackedLayout(&m)
		if m.autoToggleSidebar() {
			recalcPaneDimensions(&m)
		}
		return m, nil

	// Queued feedback was picked up by the agent (pull delivery completed)
	case feedbackPickedUpMsg:
		m.statusBar.feedbackStatus = "delivered"
		session := m.engine.GetSession()

		m.syncArtifactsAfterSubmit(session)

		if session != nil {
			m.statusBar.commentCount = 0
			m.statusBar.fileCount = len(session.ChangedFiles)
		}

		if m.focusModeActive {
			m.sidebarHidden = m.focusModeSavedSidebar
			m.diffView.wrap = m.focusModeSavedWrap
			m.focusModeActive = false
			m.sidebarAutoHidden = m.sidebarHidden && !m.sidebarHasItems()
			m.autoToggleSidebar()
			return m, func() tea.Msg {
				return tea.WindowSizeMsg{Width: m.width, Height: m.height}
			}
		}

		if m.autoToggleSidebar() {
			return m, func() tea.Msg {
				return tea.WindowSizeMsg{Width: m.width, Height: m.height}
			}
		}

		return m, nil

	// Confirm overlay actions
	case confirmActionMsg:
		m.overlay = overlayNone
		engine := m.engine
		currentPath := m.diffView.path
		isContent := m.diffView.contentMode
		switch msg.action {
		case confirmDiscard:
			return m, func() tea.Msg {
				_ = engine.ClearComments()
				return commentsClearedMsg{reloadPath: currentPath, isContent: isContent, isAdditionalFile: m.diffView.additionalFilePath != ""}
			}
		case confirmClear:
			isAdditional := m.diffView.additionalFilePath != ""
			return m, func() tea.Msg {
				_ = engine.ClearReview()
				return reviewClearedMsg{reloadPath: currentPath, isContent: isContent, isAdditionalFile: isAdditional}
			}
		case confirmDismissArtifact:
			id := m.pendingDismissArtifactID
			m.pendingDismissArtifactID = ""
			if id == "" {
				return m, nil
			}
			return m, func() tea.Msg {
				err := engine.DismissArtifact(id)
				return artifactDismissedMsg{id: id, err: err}
			}
		}
		return m, nil

	case mcpRegisterPromptMsg:
		m.registerPrompt.open()
		m.overlay = overlayRegisterPrompt
		return m, nil

	case openInfoBannerMsg:
		m.infoBanner.open(msg.title, msg.message)
		m.overlay = overlayInfo
		return m, nil

	case closeInfoBannerMsg:
		m.overlay = overlayNone
		if msg.quit {
			return m, tea.Quit
		}
		return m, nil

	case registerMCPMsg:
		m.overlay = overlayNone
		registerFn := m.mcpRegisterFn
		global := msg.global
		m.statusBar.feedbackStatus = "Registering MCP channel..."
		return m, func() tea.Msg {
			return mcpRegisterResultMsg{err: registerFn(global)}
		}

	case cancelRegisterMsg:
		m.overlay = overlayNone
		return m, nil

	case mcpRegisterResultMsg:
		if msg.err != nil {
			m.statusBar.feedbackStatus = "MCP registration failed"
		} else {
			m.statusBar.feedbackStatus = "MCP channel registered"
			m.mcpRegisterFn = nil
		}
		return m, nil

	case openConfirmMsg:
		m.confirm.open(msg.title, msg.message, msg.action)
		m.overlay = overlayConfirm
		return m, nil

	case cancelConfirmMsg:
		m.overlay = overlayNone
		m.pendingDismissArtifactID = ""
		return m, nil

	case openHistoryMsg:
		m.history.open(msg.submissions)
		m.history.width = m.width
		m.history.height = m.height
		m.overlay = overlayHistory
		return m, nil

	case closeHistoryMsg:
		m.overlay = overlayNone
		return m, nil

	// After comments are cleared (e.g. :discard), refresh sidebar + diff
	case commentsClearedMsg:
		m.sidebar.files = m.engine.GetChangedFiles()
		m.sidebar.contentItems = m.engine.GetContentItems()
		m.sidebar.applyReviewedFilter()
		m.sidebar.rebuildTree()
		m.sidebar.clampOffset()
		recalcStackedLayout(&m)
		m.statusBar.fileCount = len(m.sidebar.files)
		session := m.engine.GetSession()
		if session != nil {
			m.statusBar.baseRef = m.displayBaseRef(session)
			m.statusBar.commentCount = len(session.Comments)
		}
		// Reload current view to remove inline comment markers
		if msg.reloadPath != "" && msg.isContent {
			return m, m.handleSidebarSelect(sidebarSelectMsg{isContent: true, contentID: msg.reloadPath})
		} else if msg.reloadPath != "" && msg.isAdditionalFile {
			return m, m.handleSidebarSelect(sidebarSelectMsg{path: msg.reloadPath, isAdditionalFile: true})
		} else if msg.reloadPath != "" {
			return m, m.handleSidebarSelect(sidebarSelectMsg{path: msg.reloadPath})
		}
		return m, nil

	// After full review clear (:clear), refresh everything
	case reviewClearedMsg:
		m.sidebar.contentItems = nil
		m.sidebar.files = m.engine.GetChangedFiles()
		m.sidebar.applyReviewedFilter()
		m.sidebar.rebuildTree()
		m.sidebar.clampOffset()
		recalcStackedLayout(&m)
		m.statusBar.fileCount = len(m.sidebar.files)
		session := m.engine.GetSession()
		if session != nil {
			m.statusBar.baseRef = m.displayBaseRef(session)
			m.statusBar.commentCount = len(session.Comments)
		}
		// If viewing a content item, it no longer exists — clear the view
		if msg.isContent {
			m.diffView.contentMode = false
			m.diffView.contentID = ""
			m.diffView.path = ""
			m.diffView.hunks = nil
			m.diffView.lines = nil
			m.diffView.comments = nil
			return m, nil
		}
		// Reload current file view to remove inline comment markers
		if msg.reloadPath != "" && msg.isAdditionalFile {
			return m, m.handleSidebarSelect(sidebarSelectMsg{path: msg.reloadPath, isAdditionalFile: true})
		} else if msg.reloadPath != "" {
			return m, m.handleSidebarSelect(sidebarSelectMsg{path: msg.reloadPath})
		}
		return m, nil

	case artifactDismissedMsg:
		if msg.err != nil {
			m.statusBar.feedbackStatus = fmt.Sprintf("dismiss artifact failed: %v", msg.err)
			return m, nil
		}
		m.sidebar.contentItems = m.engine.GetContentItems()
		m.sidebar.applyReviewedFilter()
		m.sidebar.rebuildTree()
		m.sidebar.clampOffset()
		recalcStackedLayout(&m)
		session := m.engine.GetSession()
		if session != nil {
			m.statusBar.commentCount = len(session.Comments)
		}
		if m.diffView.contentMode && m.diffView.contentID == msg.id {
			m.diffView.contentMode = false
			m.diffView.contentID = ""
			m.diffView.path = ""
			m.diffView.hunks = nil
			m.diffView.lines = nil
			m.diffView.comments = nil
		}
		return m, nil

	case tea.KeyPressMsg:
		return m.handleKey(msg)

	case tea.MouseClickMsg:
		if m.mouseEnabled {
			return m.handleMouseClick(msg)
		}
	case tea.MouseWheelMsg:
		if m.mouseEnabled {
			return m.handleMouseWheel(msg)
		}
	case tea.MouseMotionMsg:
		if m.mouseEnabled {
			return m.handleMouseMotion(msg)
		}
	case tea.MouseReleaseMsg:
		if m.mouseEnabled {
			return m.handleMouseRelease(msg)
		}

	case commentExpandTickMsg:
		var cmd tea.Cmd
		m.diffView, cmd = m.diffView.Update(msg)
		return m, cmd
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
		if m.refPicker.loading {
			engine := m.engine
			count := len(m.refPicker.entries) + refPickerPageSize
			cmd = func() tea.Msg {
				entries, err := engine.RecentCommits(count)
				if err != nil {
					return nil
				}
				return loadMoreRefsMsg{
					entries: entries,
					hasMore: len(entries) >= count,
				}
			}
		}
		return m, cmd
	}
	if m.overlay == overlayVersionPicker {
		var cmd tea.Cmd
		m.versionPicker, cmd = m.versionPicker.Update(msg)
		return m, cmd
	}
	if m.overlay == overlayConfirm {
		var cmd tea.Cmd
		m.confirm, cmd = m.confirm.Update(msg)
		return m, cmd
	}
	if m.overlay == overlayRegisterPrompt {
		var cmd tea.Cmd
		m.registerPrompt, cmd = m.registerPrompt.Update(msg)
		return m, cmd
	}
	if m.overlay == overlayConnectionInfo {
		var cmd tea.Cmd
		m.connectionInfo, cmd = m.connectionInfo.Update(msg)
		return m, cmd
	}
	if m.overlay == overlayHistory {
		var cmd tea.Cmd
		m.history, cmd = m.history.Update(msg)
		return m, cmd
	}
	if m.overlay == overlaySessionPicker {
		var cmd tea.Cmd
		m.sessionPicker, cmd = m.sessionPicker.Update(msg)
		return m, cmd
	}
	if m.overlay == overlayInfo {
		var cmd tea.Cmd
		m.infoBanner, cmd = m.infoBanner.Update(msg)
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
			if m.sidebarHidden {
				return m, nil
			}
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
		if m.sidebarHidden {
			return m, nil
		}
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

	case Matches(key, km.ToggleSidebar):
		m.sidebarHidden = !m.sidebarHidden
		m.sidebarAutoHidden = false          // user explicitly toggled
		m.sidebarUserShown = !m.sidebarHidden // track if user forced it visible
		if m.sidebarHidden {
			m.focus = focusMain
			m.sidebar.focused = false
			m.diffView.focused = true
		}
		recalcPaneDimensions(&m)
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
		if !m.sidebar.reviewTracking {
			return m, nil
		}
		return m, m.handleMarkReviewed()

	case Matches(key, km.FilterReviewed):
		if !m.sidebar.reviewTracking {
			return m, nil
		}
		m.sidebar.cycleReviewFilter()
		m.sidebar.files = m.engine.GetChangedFiles()
		m.sidebar.contentItems = m.engine.GetContentItems()
		m.sidebar.additionalFiles = m.engine.GetAdditionalFiles()
		m.sidebar.applyReviewedFilter()
		m.sidebar.rebuildTree()
		m.sidebar.clampOffset()
		recalcStackedLayout(&m)
		m.statusBar.fileCount = len(m.sidebar.files)
		return m, nil

	case Matches(key, km.Submit):
		return m, m.executeCommand("submit")

	case Matches(key, km.Pause):
		return m, m.executeCommand("pause")

	case Matches(key, km.ClearReview):
		return m, m.executeCommand("clear")

	case Matches(key, km.DismissArtifact) && m.focus == focusSidebar && m.sidebar.selectedContentItem() != nil:
		item := m.sidebar.selectedContentItem()
		m.pendingDismissArtifactID = item.ID
		m.confirm.open(
			"Dismiss artifact?",
			fmt.Sprintf("Remove %q from the sidebar. Version history and comments will be deleted. This can't be undone.", item.Title),
			confirmDismissArtifact,
		)
		m.overlay = overlayConfirm
		return m, nil

	case Matches(key, km.ToggleFocusMode):
		if m.focusModeActive {
			// Exit focus mode: restore saved state
			m.sidebarHidden = m.focusModeSavedSidebar
			m.diffView.wrap = m.focusModeSavedWrap
			m.focusModeActive = false
			m.sidebarAutoHidden = m.sidebarHidden && !m.sidebarHasItems()
			m.autoToggleSidebar()
			return m, func() tea.Msg {
				return tea.WindowSizeMsg{Width: m.width, Height: m.height}
			}
		}
		// Enter focus mode
		m.focusModeSavedSidebar = m.sidebarHidden
		m.focusModeSavedWrap = m.diffView.wrap
		m.sidebarHidden = true
		m.sidebarAutoHidden = false // focus mode owns the hidden state
		m.diffView.hOffset = 0
		m.focus = focusMain
		m.sidebar.focused = false
		m.diffView.focused = true
		m.focusModeActive = true
		return m, func() tea.Msg {
			return tea.WindowSizeMsg{Width: m.width, Height: m.height}
		}

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

	case Matches(key, km.OpenInEditor):
		var filePath string
		var line int
		if m.focus == focusSidebar {
			if f := m.sidebar.selectedFile(); f != nil {
				filePath = filepath.Join(m.repoRoot, f.Path)
			} else if af := m.sidebar.selectedAdditionalFile(); af != nil {
				filePath = af.Path
			}
			line = 1
		} else {
			if m.diffView.contentMode {
				break
			}
			if m.diffView.additionalFilePath != "" {
				filePath = m.diffView.additionalFilePath
			} else if m.diffView.path != "" {
				filePath = filepath.Join(m.repoRoot, m.diffView.path)
			}
			line = m.diffView.currentDiffLine()
			if line < 1 {
				line = 1
			}
		}
		if filePath == "" {
			break
		}
		return m, openFileInEditor(filePath, line)

	case Matches(key, km.Refresh):
		return m, m.refreshFiles()

	case Matches(key, km.BaseRef):
		if m.nonGitMode {
			return m, nil // no git refs in directory mode
		}
		engine := m.engine
		return m, func() tea.Msg {
			entries, err := engine.RecentCommits(20)
			if err != nil {
				return nil
			}
			snapshots, _ := engine.GetSnapshots()
			return openRefPickerMsg{
				entries:    entries,
				snapshots:  snapshots,
				autoActive: engine.IsAutoAdvanceRef(),
			}
		}

	case Matches(key, km.ArtifactVersions):
		if !m.diffView.isViewingContentItem() {
			return m, func() tea.Msg {
				return openInfoBannerMsg{
					title:   "No Artifact Selected",
					message: "Select a plan or artifact in the sidebar to browse its version history.",
				}
			}
		}
		if m.diffView.contentVersionCount < 2 {
			return m, func() tea.Msg {
				return openInfoBannerMsg{
					title:   "No Version History",
					message: "This artifact has only one version. Submit updated versions to build a history.",
				}
			}
		}
		contentID := m.diffView.contentID
		engine := m.engine
		return m, func() tea.Msg {
			versions, err := engine.GetContentVersions(contentID)
			if err != nil || len(versions) < 2 {
				return nil
			}
			return openVersionPickerMsg{
				contentID: contentID,
				versions:  versions,
			}
		}

	case Matches(key, km.ScrollDown):
		m.diffView.ScrollDown()
		m.diffView.ScrollDown()
		return m, nil

	case Matches(key, km.ScrollUp):
		m.diffView.ScrollUp()
		m.diffView.ScrollUp()
		return m, nil

	case Matches(key, km.ScrollLeft):
		m.diffView.ScrollLeft()
		return m, nil

	case Matches(key, km.ScrollRight):
		m.diffView.ScrollRight()
		return m, nil

	case Matches(key, km.ScrollHome):
		m.diffView.ResetHScroll()
		return m, nil

	case Matches(key, km.ScrollFirstChar):
		m.diffView.ScrollToFirstChar()
		return m, nil

	case Matches(key, km.ScrollEnd):
		m.diffView.ScrollToEnd()
		return m, nil

	case Matches(key, km.Wrap):
		m.diffView.ToggleWrap()
		return m, nil

	case Matches(key, km.ToggleDiff):
		if m.nonGitMode {
			return m, nil // always file view in directory mode
		}
		cmd := m.diffView.CycleDiffStyle()
		return m, cmd

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

	case Matches(key, km.PrevSection):
		cmd := m.sidebar.jumpToPrevSection()
		return m, cmd

	case Matches(key, km.NextSection):
		cmd := m.sidebar.jumpToNextSection()
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
	summary        *types.ReviewSummary
	agentConnected bool
}

// executeCommand runs a named command entered in command mode.
func (m appModel) executeCommand(cmd string) tea.Cmd {
	engine := m.engine
	switch strings.TrimSpace(cmd) {
	case "submit":
		agentConnected := m.statusBar.subscriberCount > 0 || m.statusBar.connectionMode == "queue"
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
			return openReviewMsg{summary: summary, agentConnected: agentConnected}
		}

	case "submit!":
		return func() tea.Msg {
			// Auto-detect action: request_changes if issues/suggestions, approve otherwise
			action := types.ActionApprove
			summary, _ := engine.GetReviewSummary()
			if summary != nil && (summary.IssueCt+summary.SuggestionCt > 0) {
				action = types.ActionRequestChanges
			}
			result, err := engine.Submit(action, "")
			if err != nil {
				return submitErrorMsg{}
			}
			return submitSuccessMsg{agentConnected: result.AgentConnected}
		}

	case "discard":
		return func() tea.Msg {
			session := engine.GetSession()
			if session == nil || len(session.Comments) == 0 {
				return nil
			}
			return openConfirmMsg{
				title:   "Discard Review",
				message: "Discard all pending comments? This cannot be undone.",
				action:  confirmDiscard,
			}
		}

	case "clear":
		return func() tea.Msg {
			session := engine.GetSession()
			if session == nil {
				return nil
			}
			hasComments := len(session.Comments) > 0
			hasContent := len(session.ContentItems) > 0
			hasReviewed := false
			for _, f := range session.ChangedFiles {
				if f.Reviewed {
					hasReviewed = true
					break
				}
			}
			if !hasReviewed {
				for _, f := range session.AdditionalFiles {
					if f.Reviewed {
						hasReviewed = true
						break
					}
				}
			}
			if !hasComments && !hasContent && !hasReviewed {
				return nil
			}
			return openConfirmMsg{
				title:   "Clear Review",
				message: "Clear all comments, plans, and reviewed states? This cannot be undone.",
				action:  confirmClear,
			}
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

	case "history":
		return func() tea.Msg {
			subs, err := engine.GetSubmissions()
			if err != nil {
				return nil
			}
			return openHistoryMsg{submissions: subs}
		}

	case "mark-all-reviewed":
		return func() tea.Msg {
			_ = engine.MarkAllReviewed()
			return fileChangedMsg{}
		}

	case "mark-all-unreviewed":
		return func() tea.Msg {
			_ = engine.ResetAllReviewed()
			return fileChangedMsg{}
		}

	case "base-artifact-version":
		if !m.diffView.isViewingContentItem() {
			return func() tea.Msg {
				return openInfoBannerMsg{
					title:   "No Artifact Selected",
					message: "Select a plan or artifact in the sidebar to browse its version history.",
				}
			}
		}
		if m.diffView.contentVersionCount < 2 {
			return func() tea.Msg {
				return openInfoBannerMsg{
					title:   "No Version History",
					message: "This artifact has only one version. Submit updated versions to build a history.",
				}
			}
		}
		contentID := m.diffView.contentID
		return func() tea.Msg {
			versions, err := engine.GetContentVersions(contentID)
			if err != nil || len(versions) < 2 {
				return nil
			}
			return openVersionPickerMsg{
				contentID: contentID,
				versions:  versions,
			}
		}

	case "base-ref":
		if m.nonGitMode {
			return nil
		}
		return func() tea.Msg {
			entries, err := engine.RecentCommits(20)
			if err != nil {
				return nil
			}
			snapshots, _ := engine.GetSnapshots()
			return openRefPickerMsg{
				entries:    entries,
				snapshots:  snapshots,
				autoActive: engine.IsAutoAdvanceRef(),
			}
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

// displayBaseRef returns the ref to show in the status bar.
// Prefers the user's selected ref over the internal diff baseline (which is
// the parent commit used for git diff).
func (m appModel) displayBaseRef(session *types.ReviewSession) string {
	if snap := m.engine.GetActiveSnapshot(); snap != nil {
		return fmt.Sprintf("R%d (%s)", snap.ReviewRound, relativeTime(snap.CreatedAt))
	}
	if selected := m.engine.SelectedBaseRef(); selected != "" {
		return selected
	}
	return session.BaseRef
}

// stackedSidebarHeight returns the height for the sidebar in stacked mode.
// It accounts for section headers, separators, and one line per item, with a
// minimum of 8 rows and at most 35% of totalHeight.
func stackedSidebarHeight(totalHeight, fileCount, contentItemCount, additionalFileCount int) int {
	h := sidebarHeaderLines(contentItemCount, additionalFileCount) + fileCount + contentItemCount + additionalFileCount
	if h < 8 {
		h = 8
	}
	maxH := totalHeight * 35 / 100
	if maxH < 8 {
		maxH = 8
	}
	if h > maxH {
		h = maxH
	}
	return h
}

// recalcPaneDimensions recalculates sidebar and diff view dimensions for the
// current layout mode and sidebar visibility. Use this when sidebarHidden or
// layout mode may have changed (e.g. restoring from focus mode). For simple
// content-count changes in stacked mode, prefer recalcStackedLayout.
func recalcPaneDimensions(m *appModel) {
	const borderW = 2
	const borderH = 2
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
		if m.width < m.minDiffWidth+30 {
			m.layout = layoutStacked
		} else {
			m.layout = layoutHorizontal
		}
	}

	if m.sidebarHidden {
		m.sidebar.width = 0
		m.sidebar.height = 0
		diffContentW := m.width - borderW
		if diffContentW < 0 {
			diffContentW = 0
		}
		m.diffView.width = diffContentW
		m.diffView.height = contentHeight
	} else if m.layout == layoutStacked {
		contentW := m.width - borderW
		if contentW < 0 {
			contentW = 0
		}

		sidebarH := stackedSidebarHeight(contentHeight, len(m.sidebar.files), len(m.sidebar.contentItems), len(m.sidebar.additionalFiles))
		diffH := contentHeight - sidebarH - borderH
		if diffH < 0 {
			diffH = 0
		}

		m.sidebar.width = contentW
		m.sidebar.height = sidebarH
		m.diffView.width = contentW
		m.diffView.height = diffH
	} else {
		maxSidebarForDiff := m.width - m.minDiffWidth - 2*borderW
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
}

// recalcStackedLayout recalculates sidebar and diff view heights for stacked
// mode based on the current file/content item counts. No-op in horizontal mode.
func recalcStackedLayout(m *appModel) {
	if m.layout != layoutStacked || m.sidebarHidden {
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

	sidebarH := stackedSidebarHeight(contentHeight, len(m.sidebar.files), len(m.sidebar.contentItems), len(m.sidebar.additionalFiles))
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
	if m.diffView.isViewingContentItem() {
		for _, ci := range m.sidebar.contentItems {
			if ci.ID == m.diffView.contentID {
				return true
			}
		}
		return false
	}
	if m.diffView.additionalFilePath != "" {
		for _, af := range m.sidebar.additionalFiles {
			if af.Path == m.diffView.additionalFilePath {
				return true
			}
		}
		return false
	}
	for _, f := range m.sidebar.files {
		if f.Path == m.diffView.path {
			return true
		}
	}
	return false
}

// sidebarHasItems returns true if the sidebar has any displayable items.
func (m appModel) sidebarHasItems() bool {
	return len(m.sidebar.files) > 0 || len(m.sidebar.contentItems) > 0 || len(m.sidebar.additionalFiles) > 0
}

// syncArtifactsAfterSubmit refreshes the sidebar's artifact list from the
// session so it reflects reviewed-state changes applied during submit, and
// clears inline comment annotations from the active view (comments are
// frozen in the submission record once submitted).
func (m *appModel) syncArtifactsAfterSubmit(session *types.ReviewSession) {
	if session != nil {
		m.sidebar.contentItems = session.ContentItems
	}
	m.sidebar.rebuildTree()
	m.sidebar.clampOffset()
	m.diffView.comments = nil
}

// autoToggleSidebar hides the sidebar when it has no items, or shows it when
// items arrive and it was auto-hidden. Returns true if layout recalculation
// is needed.
func (m *appModel) autoToggleSidebar() bool {
	hasItems := m.sidebarHasItems()
	// Clear user override once items arrive — auto behavior resumes
	if hasItems {
		m.sidebarUserShown = false
	}
	if !hasItems && !m.sidebarHidden && !m.sidebarUserShown {
		// Auto-hide: no items to show and user hasn't forced it visible
		m.sidebarHidden = true
		m.sidebarAutoHidden = true
		m.focus = focusMain
		m.sidebar.focused = false
		m.diffView.focused = true
		return true
	}
	if hasItems && m.sidebarHidden && m.sidebarAutoHidden {
		// Auto-show: items arrived and we were the ones who hid it
		m.sidebarHidden = false
		m.sidebarAutoHidden = false
		return true
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
				id:             item.ID,
				title:          item.Title,
				content:        item.Content,
				contentType:    item.ContentType,
				comments:       comments,
				versionCount:   item.VersionCount,
				autoSwitchDiff: item.VersionCount > 1,
			}
		}
	}
	if msg.isAdditionalFile {
		return func() tea.Msg {
			content, err := m.engine.GetAdditionalFileContent(msg.path)
			if err != nil {
				return loadAdditionalFileMsg{path: msg.path, err: err}
			}
			session := m.engine.GetSession()
			var comments []types.ReviewComment
			if session != nil {
				for _, c := range session.Comments {
					if c.TargetRef == msg.path && c.TargetType == types.TargetAdditionalFile {
						comments = append(comments, c)
					}
				}
			}
			return loadAdditionalFileMsg{
				path:     msg.path,
				content:  content,
				comments: comments,
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

		var commentID string
		if msg.editingID != "" {
			_, _ = m.engine.EditComment(msg.editingID, msg.commentType, msg.body)
			commentID = msg.editingID
		} else {
			created, _ := m.engine.AddComment(target, msg.commentType, msg.body)
			if created != nil {
				commentID = created.ID
			}
		}

		// Additional files: reload as additional file view
		if msg.targetType == types.TargetAdditionalFile {
			content, err := m.engine.GetAdditionalFileContent(msg.path)
			if err != nil {
				return loadAdditionalFileMsg{path: msg.path, err: err}
			}
			session := m.engine.GetSession()
			var comments []types.ReviewComment
			if session != nil {
				for _, c := range session.Comments {
					if c.TargetRef == msg.path && c.TargetType == types.TargetAdditionalFile {
						comments = append(comments, c)
					}
				}
			}
			return loadAdditionalFileMsg{
				path:            msg.path,
				content:         content,
				comments:        comments,
				selectCommentID: commentID,
			}
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
				id:              item.ID,
				title:           item.Title,
				content:         item.Content,
				contentType:     item.ContentType,
				comments:        comments,
				versionCount:    item.VersionCount,
				selectCommentID: commentID,
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
			path:            msg.path,
			result:          result,
			comments:        comments,
			selectCommentID: commentID,
		}
	}
}

// handleMarkReviewed toggles the reviewed status of the currently selected file or content item.
func (m appModel) handleMarkReviewed() tea.Cmd {
	// Content item: check diff viewer content mode first, then sidebar cursor
	if item := m.contentItemForReview(); item != nil {
		engine := m.engine
		id := item.ID
		reviewed := item.Reviewed
		return func() tea.Msg {
			if reviewed {
				_ = engine.UnmarkContentReviewed(id)
			} else {
				_ = engine.MarkContentReviewed(id)
			}
			return contentReviewedMsg{id: id, advance: !reviewed}
		}
	}

	// File
	var filePath string
	var reviewed bool
	var isAdditional bool

	switch m.focus {
	case focusSidebar:
		if af := m.sidebar.selectedAdditionalFile(); af != nil {
			filePath = af.Path
			reviewed = af.Reviewed
			isAdditional = true
		} else if file := m.sidebar.selectedFile(); file != nil {
			filePath = file.Path
			reviewed = file.Reviewed
		} else {
			return nil
		}
	case focusMain:
		if m.diffView.path == "" {
			return nil
		}
		if m.diffView.additionalFilePath != "" {
			filePath = m.diffView.additionalFilePath
			isAdditional = true
			for _, af := range m.sidebar.additionalFiles {
				if af.Path == filePath {
					reviewed = af.Reviewed
					break
				}
			}
		} else {
			filePath = m.diffView.path
			for _, f := range m.sidebar.files {
				if f.Path == filePath {
					reviewed = f.Reviewed
					break
				}
			}
		}
	default:
		return nil
	}
	willAdvance := !reviewed
	return func() tea.Msg {
		if reviewed {
			_ = m.engine.UnmarkReviewed(filePath)
		} else {
			_ = m.engine.MarkReviewed(filePath)
		}
		if isAdditional {
			return additionalFileAddedMsg{path: filePath, advance: willAdvance}
		}
		return fileChangedMsg{path: filePath, advance: willAdvance}
	}
}

// contentItemForReview returns the content item that should be toggled, or nil.
func (m appModel) contentItemForReview() *types.ContentItem {
	// Main pane showing content
	if m.focus == focusMain && m.diffView.contentMode {
		for i := range m.sidebar.contentItems {
			if m.sidebar.contentItems[i].ID == m.diffView.contentID {
				return &m.sidebar.contentItems[i]
			}
		}
		return nil
	}
	// Sidebar cursor on content item
	if m.focus == focusSidebar {
		return m.sidebar.selectedContentItem()
	}
	return nil
}

// refreshFiles returns a Cmd that refreshes the file list and current diff from git.
func (m appModel) refreshFiles() tea.Cmd {
	engine := m.engine
	currentPath := m.diffView.path
	isContentItem := m.diffView.isViewingContentItem()
	contentID := m.diffView.contentID
	inAdditionalFileMode := m.diffView.additionalFilePath != ""
	return func() tea.Msg {
		// Refresh the file list from git
		files, err := engine.RefreshChangedFiles()
		if err != nil {
			return nil
		}
		session := engine.GetSession()

		// Refresh content item if one is currently displayed
		if isContentItem && contentID != "" {
			item, itemErr := engine.GetContentItem(contentID)
			var contentComments []types.ReviewComment
			if session != nil {
				for _, c := range session.Comments {
					if c.TargetRef == contentID && c.TargetType == types.TargetContent {
						contentComments = append(contentComments, c)
					}
				}
			}
			if itemErr == nil && item != nil {
				return refreshResultMsg{
					files:       files,
					contentItem: item,
					contentComments: contentComments,
				}
			}
			return refreshResultMsg{files: files}
		}

		// Don't reload diff when viewing an additional file — it's not a git file
		var result *types.DiffResult
		var comments []types.ReviewComment
		if currentPath != "" && !inAdditionalFileMode {
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

// contentItemChanged reports whether a refreshed content item differs from what
// the diff view is currently showing (content text or comment state).
func contentItemChanged(item *types.ContentItem, comments []types.ReviewComment, dv *diffViewModel) bool {
	if item.Content != dv.contentDiffContent {
		return true
	}
	return commentsChanged(comments, dv.comments)
}

// fileDiffChanged reports whether a refreshed file diff differs from what the
// diff view is currently showing (hunks or comment state).
func fileDiffChanged(result *types.DiffResult, comments []types.ReviewComment, dv *diffViewModel) bool {
	if result == nil {
		return len(dv.hunks) != 0
	}
	if len(result.Hunks) != len(dv.hunks) {
		return true
	}
	for i, h := range result.Hunks {
		prev := dv.hunks[i]
		if h.OldStart != prev.OldStart || h.OldCount != prev.OldCount ||
			h.NewStart != prev.NewStart || h.NewCount != prev.NewCount ||
			h.Header != prev.Header || len(h.Lines) != len(prev.Lines) {
			return true
		}
		for j, l := range h.Lines {
			pl := prev.Lines[j]
			if l.Kind != pl.Kind || l.OldLineNum != pl.OldLineNum ||
				l.NewLineNum != pl.NewLineNum || l.Content != pl.Content {
				return true
			}
		}
	}
	return commentsChanged(comments, dv.comments)
}

func commentsChanged(next, prev []types.ReviewComment) bool {
	if len(next) != len(prev) {
		return true
	}
	for i, c := range next {
		p := prev[i]
		if c.ID != p.ID || c.Body != p.Body || c.Resolved != p.Resolved {
			return true
		}
	}
	return false
}

type refreshResultMsg struct {
	files           []types.ChangedFile
	path            string
	result          *types.DiffResult
	comments        []types.ReviewComment
	contentItem     *types.ContentItem
	contentComments []types.ReviewComment
}

// loadContentMsg carries content item data for rendering in the diff view.
type loadContentMsg struct {
	id              string
	title           string
	content         string
	contentType     string
	comments        []types.ReviewComment
	versionCount    int    // number of versions stored for this content item
	autoSwitchDiff  bool   // true to auto-switch to preferred diff style
	selectCommentID string // if set, auto-select and expand this comment after loading
}

// View renders the full TUI layout.
func (m appModel) View() tea.View {
	// Title bar
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("4")).Render(" o_(◉) monocle")
	if m.focusModeActive {
		badge := lipgloss.NewStyle().
			Background(lipgloss.Color("5")).
			Foreground(lipgloss.Color("0")).
			Bold(true).
			Padding(0, 1).
			Render("FOCUS MODE")
		title = title + " " + badge
	}
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

	if m.sidebarHidden {
		mainView := mainStyle.
			Width(m.diffView.width + bw).
			Height(m.diffView.height + bh).
			Render(m.diffView.View())
		body = mainView
	} else if m.layout == layoutStacked {
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
	if m.focus == focusMain && m.diffView.CursorComment() != nil {
		m.statusBar.contextHints = "c:edit  d:delete  x:resolve  ?:help"
	} else {
		m.statusBar.contextHints = ""
	}
	m.statusBar.diffStyle = m.diffView.style
	m.statusBar.contentMode = m.diffView.contentMode
	m.statusBar.contentID = m.diffView.contentID
	m.statusBar.diffBaseVersion = m.diffView.diffBaseVersion
	m.statusBar.diffToVersion = m.diffView.diffToVersion
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
	} else if m.overlay == overlayVersionPicker {
		overlayContent := m.versionPicker.View()
		if overlayContent != "" {
			full = overlayOn(full, overlayContent, m.width, m.height)
		}
	} else if m.overlay == overlayConfirm {
		overlayContent := m.confirm.View()
		if overlayContent != "" {
			full = overlayOn(full, overlayContent, m.width, m.height)
		}
	} else if m.overlay == overlayRegisterPrompt {
		overlayContent := m.registerPrompt.View()
		if overlayContent != "" {
			full = overlayOn(full, overlayContent, m.width, m.height)
		}
	} else if m.overlay == overlayConnectionInfo {
		overlayContent := m.connectionInfo.View()
		if overlayContent != "" {
			full = overlayOn(full, overlayContent, m.width, m.height)
		}
	} else if m.overlay == overlayHistory {
		overlayContent := m.history.View()
		if overlayContent != "" {
			full = overlayOn(full, overlayContent, m.width, m.height)
		}
	} else if m.overlay == overlaySessionPicker {
		overlayContent := m.sessionPicker.View()
		if overlayContent != "" {
			full = overlayOn(full, overlayContent, m.width, m.height)
		}
	} else if m.overlay == overlayInfo {
		overlayContent := m.infoBanner.View()
		if overlayContent != "" {
			full = overlayOn(full, overlayContent, m.width, m.height)
		}
	}

	v := tea.NewView(full)
	v.AltScreen = true
	if m.mouseEnabled {
		v.MouseMode = tea.MouseModeCellMotion
	}
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

// CalcModalWidth computes modal width as max(screenWidth*2/3, 65), capped by
// maxWidth (pass 0 for no cap) and screen bounds (screenWidth-10 for margin).
func CalcModalWidth(screenWidth, maxWidth int) int {
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
	engine.On(core.EventFeedbackStatusChanged, func(e core.EventPayload) {
		p.Send(feedbackStatusMsg{status: e.Status})
	})
	engine.On(core.EventContentItemAdded, func(e core.EventPayload) {
		p.Send(contentItemMsg{id: e.ItemID})
	})
	engine.On(core.EventAdditionalFileAdded, func(e core.EventPayload) {
		p.Send(additionalFileAddedMsg{path: e.Path})
	})
	engine.On(core.EventPauseChanged, func(e core.EventPayload) {
		p.Send(pauseChangedMsg{status: e.Status})
	})
	engine.On(core.EventConnectionChanged, func(e core.EventPayload) {
		var msg connectionChangedMsg
		msg.agentName = e.Message
		if e.Status == "queue" {
			msg.mode = "queue"
		} else {
			msg.count, _ = strconv.Atoi(e.Status)
		}
		p.Send(msg)
	})
	engine.On(core.EventFeedbackPickedUp, func(e core.EventPayload) {
		p.Send(feedbackPickedUpMsg{})
	})
	engine.On(core.EventWaitStatusChanged, func(e core.EventPayload) {
		p.Send(waitStatusMsg{waiting: e.Status == "waiting"})
	})
}

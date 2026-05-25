package tui

import (
	"fmt"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/josephschmitt/monocle/internal/core"
	"github.com/josephschmitt/monocle/internal/types"
)

// stubEngine is a minimal EngineAPI stub for testing TUI behavior.
type stubEngine struct {
	core.EngineAPI // embed to satisfy interface; panics on unimplemented methods
	cfg            *types.Config
	session        *types.ReviewSession
	contentItems   []types.ContentItem
	cleared        bool
	dismissCalled  bool
}

func (s *stubEngine) GetConfig() *types.Config                  { return s.cfg }
func (s *stubEngine) GetSession() *types.ReviewSession           { return s.session }
func (s *stubEngine) GetFeedbackStatus() string { return "" }
func (s *stubEngine) GetQueuedCount() int        { return 0 }
func (s *stubEngine) ReloadPendingFeedback()     {}
func (s *stubEngine) SelectedBaseRef() string    { return "" }
func (s *stubEngine) GetChangedFiles() []types.ChangedFile       { return nil }
func (s *stubEngine) GetAdditionalFiles() []types.AdditionalFile { return nil }
func (s *stubEngine) MarkContentReviewed(id string) error        { return nil }
func (s *stubEngine) UnmarkContentReviewed(id string) error      { return nil }
func (s *stubEngine) GetContentItems() []types.ContentItem       { return s.contentItems }
func (s *stubEngine) GetContentItem(id string) (*types.ContentItem, error) {
	for i := range s.contentItems {
		if s.contentItems[i].ID == id {
			return &s.contentItems[i], nil
		}
	}
	return nil, fmt.Errorf("not found")
}
func (s *stubEngine) GetSnapshots() ([]types.ReviewSnapshot, error)   { return nil, nil }
func (s *stubEngine) SetSnapshotBase(snapshotID int) error             { return nil }
func (s *stubEngine) ClearSnapshotBase()                               {}
func (s *stubEngine) GetActiveSnapshot() *types.ReviewSnapshot         { return nil }
func (s *stubEngine) HasSnapshots() (bool, error)                      { return false, nil }
func (s *stubEngine) IsReviewTrackingEnabled() bool                    { return s.cfg != nil && s.cfg.ReviewTracking }
func (s *stubEngine) ClearComments() error {
	s.cleared = true
	return nil
}
func (s *stubEngine) DismissArtifact(id string) error {
	s.dismissCalled = true
	filtered := s.contentItems[:0]
	for _, item := range s.contentItems {
		if item.ID != id {
			filtered = append(filtered, item)
		}
	}
	s.contentItems = filtered
	if s.session != nil {
		sessionItems := s.session.ContentItems[:0]
		for _, item := range s.session.ContentItems {
			if item.ID != id {
				sessionItems = append(sessionItems, item)
			}
		}
		s.session.ContentItems = sessionItems
	}
	return nil
}
func (s *stubEngine) ClearReview() error {
	s.cleared = true
	s.session.Comments = nil
	s.session.ContentItems = nil
	s.contentItems = nil
	for i := range s.session.ChangedFiles {
		s.session.ChangedFiles[i].Reviewed = false
	}
	return nil
}

func newTestSession(withComments bool) *types.ReviewSession {
	session := &types.ReviewSession{ID: "test-session"}
	if withComments {
		session.Comments = []types.ReviewComment{
			{ID: "c1", Body: "fix this"},
		}
	}
	return session
}

func TestSubmitSuccess_AlwaysClearsComments(t *testing.T) {
	engine := &stubEngine{
		cfg:     &types.Config{},
		session: newTestSession(true),
	}
	m := NewApp(engine)

	result, _ := m.Update(submitSuccessMsg{})
	app := result.(appModel)

	if app.overlay == overlayConfirm {
		t.Error("expected no confirm modal — comments should always auto-clear")
	}
	if !engine.cleared {
		t.Error("expected ClearComments to be called")
	}
}

func TestSubmitSuccess_NoComments_SkipsClear(t *testing.T) {
	session := &types.ReviewSession{
		ID:       "test",
		Comments: nil,
	}
	engine := &stubEngine{
		cfg:     &types.Config{},
		session: session,
	}
	m := NewApp(engine)

	_, _ = m.Update(submitSuccessMsg{})

	if engine.cleared {
		t.Error("expected ClearComments NOT to be called when no comments")
	}
}

func TestSubmitSuccess_AgentDisconnected_ClearsComments(t *testing.T) {
	engine := &stubEngine{
		cfg:     &types.Config{},
		session: newTestSession(true),
	}
	m := NewApp(engine)

	_, cmd := m.Update(submitSuccessMsg{})

	if cmd != nil {
		t.Error("expected no command when agent disconnected")
	}
	// Comments are cleared even without agent — they're frozen in the
	// queued submission record and should not remain in the UI.
	if !engine.cleared {
		t.Error("expected ClearComments to be called for queued submission")
	}
}

func TestSubmitSuccess_PreservesContentView(t *testing.T) {
	engine := &stubEngine{
		cfg:     &types.Config{},
		session: newTestSession(false),
	}
	m := NewApp(engine)
	m.diffView.contentMode = true
	m.diffView.contentID = "plan-1"
	m.diffView.path = "plan-1"
	m.diffView.comments = []types.ReviewComment{{ID: "c1"}}

	result, _ := m.Update(submitSuccessMsg{})
	app := result.(appModel)

	// Artifacts persist across rounds — keep the content view on screen so
	// the reviewer can keep referring to the plan after submitting.
	if !app.diffView.contentMode {
		t.Error("expected contentMode to be preserved after submit")
	}
	if app.diffView.contentID != "plan-1" {
		t.Errorf("expected contentID preserved, got %q", app.diffView.contentID)
	}
	// Inline comment annotations should be cleared — comments are frozen
	// in the submission record.
	if len(app.diffView.comments) != 0 {
		t.Errorf("expected inline comments cleared, got %d", len(app.diffView.comments))
	}
}

func TestClearReview_OpensConfirmWhenHasState(t *testing.T) {
	engine := &stubEngine{
		cfg: &types.Config{},
		session: &types.ReviewSession{
			ID:       "test",
			Comments: []types.ReviewComment{{ID: "c1", Body: "fix"}},
		},
	}
	m := NewApp(engine)

	cmd := m.executeCommand("clear")
	if cmd == nil {
		t.Fatal("expected a command from clear")
	}
	msg := cmd()
	confirm, ok := msg.(openConfirmMsg)
	if !ok {
		t.Fatalf("expected openConfirmMsg, got %T", msg)
	}
	if confirm.action != confirmClear {
		t.Errorf("expected confirmClear action, got %v", confirm.action)
	}
}

func TestClearReview_NoopWhenEmpty(t *testing.T) {
	engine := &stubEngine{
		cfg:     &types.Config{},
		session: &types.ReviewSession{ID: "test"},
	}
	m := NewApp(engine)

	cmd := m.executeCommand("clear")
	if cmd == nil {
		t.Fatal("expected a command from clear")
	}
	msg := cmd()
	if msg != nil {
		t.Errorf("expected nil message when nothing to clear, got %T", msg)
	}
}

func TestSubmitSuccess_RecalcsStackedLayout(t *testing.T) {
	engine := &stubEngine{
		cfg:     &types.Config{},
		session: newTestSession(true),
	}
	m := NewApp(engine)
	// Set initial dimensions — 80 wide triggers stacked layout
	result, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 40})
	m = result.(appModel)
	if m.layout != layoutStacked {
		t.Fatalf("expected stacked layout, got %v", m.layout)
	}

	// Add files and content items to establish a baseline sidebar height.
	item := types.ContentItem{ID: "plan-1", Title: "Plan"}
	m.sidebar.files = []types.ChangedFile{{Path: "file.go", Status: "M"}}
	m.sidebar.contentItems = []types.ContentItem{item}
	engine.session.ContentItems = []types.ContentItem{item}
	m.sidebar.rebuildTree()
	recalcStackedLayout(&m)

	// Submit feedback — artifacts persist across rounds.
	result, _ = m.Update(submitSuccessMsg{})
	app := result.(appModel)

	if len(app.sidebar.contentItems) != 1 {
		t.Errorf("expected artifact to persist after submit, got %d", len(app.sidebar.contentItems))
	}
	if app.sidebar.height == 0 {
		t.Error("expected non-zero sidebar height after submit")
	}
	if app.diffView.height == 0 {
		t.Error("expected non-zero diffView height after submit")
	}
}

func TestSubmitSuccess_FocusModeRestoresDimensions(t *testing.T) {
	engine := &stubEngine{
		cfg:     &types.Config{},
		session: newTestSession(true),
	}
	m := NewApp(engine)
	// Set initial dimensions
	result, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 40})
	m = result.(appModel)

	// Add files so the sidebar stays visible after focus mode restore
	m.sidebar.files = []types.ChangedFile{{Path: "file.go", Status: "M"}}

	// Enter focus mode (sidebar hidden)
	m.focusModeSavedSidebar = false
	m.focusModeSavedWrap = false
	m.sidebarHidden = true
	m.focusModeActive = true
	result, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 40})
	m = result.(appModel)
	if m.sidebar.width != 0 || m.sidebar.height != 0 {
		t.Fatal("expected zero sidebar dimensions in focus mode")
	}

	// Submit feedback (restores focus mode)
	result, cmd := m.Update(submitSuccessMsg{})
	app := result.(appModel)

	if app.sidebarHidden {
		t.Error("expected sidebar to be visible after focus mode restore")
	}
	if app.sidebar.width == 0 {
		t.Error("expected non-zero sidebar width after focus mode restore")
	}
	if app.sidebar.height == 0 {
		t.Error("expected non-zero sidebar height after focus mode restore")
	}
	if cmd != nil {
		t.Error("expected nil command (inline recalc, no deferred WindowSizeMsg)")
	}
}

func TestSubmitSuccess_NoAgent_FocusModeRestoresDimensions(t *testing.T) {
	engine := &stubEngine{
		cfg:     &types.Config{},
		session: newTestSession(true),
	}
	m := NewApp(engine)
	result, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 40})
	m = result.(appModel)

	// Add files so the sidebar stays visible after focus mode restore
	m.sidebar.files = []types.ChangedFile{{Path: "file.go", Status: "M"}}

	// Enter focus mode
	m.focusModeSavedSidebar = false
	m.sidebarHidden = true
	m.focusModeActive = true
	result, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 40})
	m = result.(appModel)

	// Submit with no agent connected
	result, _ = m.Update(submitSuccessMsg{})
	app := result.(appModel)

	if app.sidebarHidden {
		t.Error("expected sidebar visible after no-agent focus restore")
	}
	if app.sidebar.width == 0 {
		t.Error("expected non-zero sidebar width")
	}
	if app.sidebar.height == 0 {
		t.Error("expected non-zero sidebar height")
	}
}

func TestDismissArtifact_ConfirmAcceptCallsEngine(t *testing.T) {
	item := types.ContentItem{ID: "plan-1", Title: "Plan"}
	engine := &stubEngine{
		cfg: &types.Config{},
		session: &types.ReviewSession{
			ID:           "test",
			ContentItems: []types.ContentItem{item},
		},
		contentItems: []types.ContentItem{item},
	}
	m := NewApp(engine)
	m.pendingDismissArtifactID = "plan-1"

	_, cmd := m.Update(confirmActionMsg{action: confirmDismissArtifact})
	if cmd == nil {
		t.Fatal("expected command from confirmDismissArtifact accept")
	}
	msg := cmd()
	result, ok := msg.(artifactDismissedMsg)
	if !ok {
		t.Fatalf("expected artifactDismissedMsg, got %T", msg)
	}
	if result.id != "plan-1" {
		t.Errorf("expected id plan-1, got %q", result.id)
	}
	if !engine.dismissCalled {
		t.Error("expected engine.DismissArtifact to be called")
	}
}

func TestDismissArtifact_ConfirmCancelClearsPending(t *testing.T) {
	engine := &stubEngine{cfg: &types.Config{}, session: &types.ReviewSession{ID: "test"}}
	m := NewApp(engine)
	m.pendingDismissArtifactID = "plan-1"

	result, _ := m.Update(cancelConfirmMsg{})
	app := result.(appModel)
	if app.pendingDismissArtifactID != "" {
		t.Errorf("expected pendingDismissArtifactID cleared on cancel, got %q", app.pendingDismissArtifactID)
	}
	if engine.dismissCalled {
		t.Error("expected engine.DismissArtifact NOT to be called on cancel")
	}
}

func TestArtifactDismissed_ClearsContentViewIfViewing(t *testing.T) {
	engine := &stubEngine{
		cfg:     &types.Config{},
		session: &types.ReviewSession{ID: "test"},
	}
	m := NewApp(engine)
	m.diffView.contentMode = true
	m.diffView.contentID = "plan-1"
	m.diffView.path = "plan-1"

	result, _ := m.Update(artifactDismissedMsg{id: "plan-1"})
	app := result.(appModel)

	if app.diffView.contentMode {
		t.Error("expected contentMode cleared when the viewed artifact is dismissed")
	}
	if app.diffView.contentID != "" {
		t.Errorf("expected contentID cleared, got %q", app.diffView.contentID)
	}
}

func TestClearReview_ClearsContentView(t *testing.T) {
	engine := &stubEngine{
		cfg: &types.Config{},
		session: &types.ReviewSession{
			ID:           "test",
			ContentItems: []types.ContentItem{{ID: "plan-1", Title: "Plan"}},
		},
		contentItems: []types.ContentItem{{ID: "plan-1", Title: "Plan"}},
	}
	m := NewApp(engine)
	m.diffView.contentMode = true
	m.diffView.contentID = "plan-1"
	m.diffView.path = "plan-1"

	result, _ := m.Update(reviewClearedMsg{reloadPath: "plan-1", isContent: true})
	app := result.(appModel)

	if app.diffView.contentMode {
		t.Error("expected contentMode to be cleared")
	}
	if app.diffView.contentID != "" {
		t.Errorf("expected contentID to be cleared, got %q", app.diffView.contentID)
	}
}

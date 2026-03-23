package tui

import (
	"testing"

	"github.com/anthropics/monocle/internal/core"
	"github.com/anthropics/monocle/internal/types"
)

// stubEngine is a minimal EngineAPI stub for testing TUI behavior.
type stubEngine struct {
	core.EngineAPI // embed to satisfy interface; panics on unimplemented methods
	cfg            *types.Config
	session        *types.ReviewSession
	cleared        bool
}

func (s *stubEngine) GetConfig() *types.Config          { return s.cfg }
func (s *stubEngine) GetSession() *types.ReviewSession   { return s.session }
func (s *stubEngine) GetFeedbackStatus() string          { return "" }
func (s *stubEngine) GetChangedFiles() []types.ChangedFile { return nil }
func (s *stubEngine) ClearComments() error {
	s.cleared = true
	return nil
}

func newTestSession(withComments bool) *types.ReviewSession {
	session := &types.ReviewSession{ID: "test-session"}
	if withComments {
		session.Comments = []types.ReviewComment{
			{ID: "c1", Body: "fix this", Outdated: false},
		}
	}
	return session
}

func TestSubmitSuccess_ConfigAsk_ShowsModal(t *testing.T) {
	engine := &stubEngine{
		cfg:     &types.Config{ClearAfterSubmit: "ask"},
		session: newTestSession(true),
	}
	m := NewApp(engine)
	m.width = 80
	m.height = 40

	result, _ := m.Update(submitSuccessMsg{})
	app := result.(appModel)

	if app.overlay != overlayConfirm {
		t.Errorf("expected overlayConfirm, got %d", app.overlay)
	}
	if !app.confirm.showDontAsk {
		t.Error("expected confirm modal to show 'don't ask' checkbox")
	}
}

func TestSubmitSuccess_ConfigAlways_AutoClears(t *testing.T) {
	engine := &stubEngine{
		cfg:     &types.Config{ClearAfterSubmit: "always"},
		session: newTestSession(true),
	}
	m := NewApp(engine)

	result, cmd := m.Update(submitSuccessMsg{})
	app := result.(appModel)

	if app.overlay == overlayConfirm {
		t.Error("expected no confirm modal for 'always'")
	}
	if cmd == nil {
		t.Fatal("expected a command to clear comments")
	}

	// Execute the command and verify it clears
	msg := cmd()
	if _, ok := msg.(commentsClearedMsg); !ok {
		t.Errorf("expected commentsClearedMsg, got %T", msg)
	}
	if !engine.cleared {
		t.Error("expected ClearComments to be called")
	}
}

func TestSubmitSuccess_ConfigNever_SkipsModal(t *testing.T) {
	engine := &stubEngine{
		cfg:     &types.Config{ClearAfterSubmit: "never"},
		session: newTestSession(true),
	}
	m := NewApp(engine)

	result, cmd := m.Update(submitSuccessMsg{})
	app := result.(appModel)

	if app.overlay == overlayConfirm {
		t.Error("expected no confirm modal for 'never'")
	}
	if cmd != nil {
		t.Error("expected no command for 'never'")
	}
	if engine.cleared {
		t.Error("expected ClearComments NOT to be called")
	}
}

func TestSubmitSuccess_NoActiveComments_SkipsModal(t *testing.T) {
	session := &types.ReviewSession{
		ID: "test",
		Comments: []types.ReviewComment{
			{ID: "c1", Body: "old", Outdated: true},
		},
	}
	engine := &stubEngine{
		cfg:     &types.Config{ClearAfterSubmit: "ask"},
		session: session,
	}
	m := NewApp(engine)

	result, _ := m.Update(submitSuccessMsg{})
	app := result.(appModel)

	if app.overlay == overlayConfirm {
		t.Error("expected no modal when all comments are outdated")
	}
}

func TestSubmitSuccess_SessionOverrideAlways(t *testing.T) {
	engine := &stubEngine{
		cfg:     &types.Config{ClearAfterSubmit: "ask"},
		session: newTestSession(true),
	}
	m := NewApp(engine)
	m.clearAfterSubmitOverride = "always"

	result, cmd := m.Update(submitSuccessMsg{})
	app := result.(appModel)

	if app.overlay == overlayConfirm {
		t.Error("expected no modal when session override is 'always'")
	}
	if cmd == nil {
		t.Fatal("expected a command to clear comments")
	}
	msg := cmd()
	if _, ok := msg.(commentsClearedMsg); !ok {
		t.Errorf("expected commentsClearedMsg, got %T", msg)
	}
}

func TestSubmitSuccess_SessionOverrideNever(t *testing.T) {
	engine := &stubEngine{
		cfg:     &types.Config{ClearAfterSubmit: "ask"},
		session: newTestSession(true),
	}
	m := NewApp(engine)
	m.clearAfterSubmitOverride = "never"

	result, cmd := m.Update(submitSuccessMsg{})
	app := result.(appModel)

	if app.overlay == overlayConfirm {
		t.Error("expected no modal when session override is 'never'")
	}
	if cmd != nil {
		t.Error("expected no command for 'never'")
	}
}

func TestConfirmWithDontAsk_SetsSessionOverrideAlways(t *testing.T) {
	engine := &stubEngine{
		cfg:     &types.Config{ClearAfterSubmit: "ask"},
		session: newTestSession(true),
	}
	m := NewApp(engine)
	m.overlay = overlayConfirm

	result, _ := m.Update(confirmActionMsg{action: confirmClearAfterSubmit, dontAsk: true})
	app := result.(appModel)

	if app.clearAfterSubmitOverride != "always" {
		t.Errorf("expected override 'always', got %q", app.clearAfterSubmitOverride)
	}
}

func TestCancelWithDontAsk_SetsSessionOverrideNever(t *testing.T) {
	engine := &stubEngine{
		cfg:     &types.Config{ClearAfterSubmit: "ask"},
		session: newTestSession(true),
	}
	m := NewApp(engine)
	m.overlay = overlayConfirm

	result, _ := m.Update(cancelConfirmMsg{dontAsk: true})
	app := result.(appModel)

	if app.clearAfterSubmitOverride != "never" {
		t.Errorf("expected override 'never', got %q", app.clearAfterSubmitOverride)
	}
}

func TestConfirmWithoutDontAsk_NoOverride(t *testing.T) {
	engine := &stubEngine{
		cfg:     &types.Config{ClearAfterSubmit: "ask"},
		session: newTestSession(true),
	}
	m := NewApp(engine)
	m.overlay = overlayConfirm

	result, _ := m.Update(confirmActionMsg{action: confirmClearAfterSubmit, dontAsk: false})
	app := result.(appModel)

	if app.clearAfterSubmitOverride != "" {
		t.Errorf("expected no override, got %q", app.clearAfterSubmitOverride)
	}
}

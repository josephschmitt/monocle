package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func shiftTabKey() tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}
}

func keyPress(code rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: code}
}

func TestConfirmModel_OpenSetsFields(t *testing.T) {
	m := newConfirmModel(DefaultTheme())
	m.open("Title", "Message", confirmDiscard)

	if !m.active {
		t.Error("expected active after open")
	}
	if m.showDontAsk {
		t.Error("expected showDontAsk=false for regular open")
	}
}

func TestConfirmModel_OpenWithDontAskShowsCheckbox(t *testing.T) {
	m := newConfirmModel(DefaultTheme())
	m.openWithDontAsk("Title", "Message", confirmClearAfterSubmit)

	if !m.active {
		t.Error("expected active after openWithDontAsk")
	}
	if !m.showDontAsk {
		t.Error("expected showDontAsk=true for openWithDontAsk")
	}
	if m.dontAsk {
		t.Error("expected dontAsk=false initially")
	}
}

func TestConfirmModel_ShiftTabTogglesDontAsk(t *testing.T) {
	m := newConfirmModel(DefaultTheme())
	m.openWithDontAsk("Title", "Message", confirmClearAfterSubmit)

	// Toggle on
	m, _ = m.Update(shiftTabKey())
	if !m.dontAsk {
		t.Error("expected dontAsk=true after first shift+tab")
	}

	// Toggle off
	m, _ = m.Update(shiftTabKey())
	if m.dontAsk {
		t.Error("expected dontAsk=false after second shift+tab")
	}
}

func TestConfirmModel_ShiftTabIgnoredWithoutShowDontAsk(t *testing.T) {
	m := newConfirmModel(DefaultTheme())
	m.open("Title", "Message", confirmDiscard)

	m, _ = m.Update(shiftTabKey())
	if m.dontAsk {
		t.Error("shift+tab should be ignored when showDontAsk is false")
	}
}

func TestConfirmModel_ConfirmIncludesDontAsk(t *testing.T) {
	m := newConfirmModel(DefaultTheme())
	m.openWithDontAsk("Title", "Message", confirmClearAfterSubmit)

	// Toggle don't ask on
	m, _ = m.Update(shiftTabKey())

	// Confirm
	var cmd tea.Cmd
	m, cmd = m.Update(keyPress('y'))

	if m.active {
		t.Error("expected inactive after confirm")
	}

	msg := cmd()
	action, ok := msg.(confirmActionMsg)
	if !ok {
		t.Fatalf("expected confirmActionMsg, got %T", msg)
	}
	if !action.dontAsk {
		t.Error("expected dontAsk=true in confirmActionMsg")
	}
	if action.action != confirmClearAfterSubmit {
		t.Error("expected confirmClearAfterSubmit action")
	}
}

func TestConfirmModel_CancelIncludesDontAsk(t *testing.T) {
	m := newConfirmModel(DefaultTheme())
	m.openWithDontAsk("Title", "Message", confirmClearAfterSubmit)

	// Toggle don't ask on
	m, _ = m.Update(shiftTabKey())

	// Cancel
	var cmd tea.Cmd
	m, cmd = m.Update(keyPress(tea.KeyEscape))

	if m.active {
		t.Error("expected inactive after cancel")
	}

	msg := cmd()
	cancel, ok := msg.(cancelConfirmMsg)
	if !ok {
		t.Fatalf("expected cancelConfirmMsg, got %T", msg)
	}
	if !cancel.dontAsk {
		t.Error("expected dontAsk=true in cancelConfirmMsg")
	}
}

func TestConfirmModel_ViewShowsCheckbox(t *testing.T) {
	m := newConfirmModel(DefaultTheme())
	m.width = 80
	m.height = 40
	m.openWithDontAsk("Review Submitted", "Clear all comments?", confirmClearAfterSubmit)

	view := m.View()
	if view == "" {
		t.Fatal("expected non-empty view")
	}

	if !strings.Contains(view, "Don't ask again") {
		t.Error("expected view to contain 'Don't ask again'")
	}
	if !strings.Contains(view, "Shift+Tab") {
		t.Error("expected view to contain 'Shift+Tab' hint")
	}
}

func TestConfirmModel_ViewHidesCheckboxForRegularOpen(t *testing.T) {
	m := newConfirmModel(DefaultTheme())
	m.width = 80
	m.height = 40
	m.open("Confirm", "Are you sure?", confirmDiscard)

	view := m.View()
	if strings.Contains(view, "Don't ask again") {
		t.Error("expected view to NOT contain 'Don't ask again' for regular open")
	}
	if strings.Contains(view, "Shift+Tab") {
		t.Error("expected view to NOT contain 'Shift+Tab' hint for regular open")
	}
}

package tui

import "charm.land/bubbles/v2/key"

// KeyMap defines all keybindings.
type KeyMap struct {
	Up        key.Binding
	Down      key.Binding
	HalfUp   key.Binding
	HalfDown key.Binding
	Top       key.Binding
	Bottom    key.Binding
	PrevFile  key.Binding
	NextFile  key.Binding
	Select    key.Binding
	Comment     key.Binding
	FileComment key.Binding
	Visual      key.Binding
	VisualLine  key.Binding
	Reviewed  key.Binding
	ToggleDiff key.Binding
	TreeMode   key.Binding
	Collapse   key.Binding
	ExpandAll  key.Binding
	ScrollMainDown  key.Binding
	ScrollMainUp    key.Binding
	ScrollMainLeft  key.Binding
	ScrollMainRight key.Binding
	Wrap            key.Binding
	FocusSwap       key.Binding
	Quit      key.Binding
}

// DefaultKeyMap returns the default keybindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up:        key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("k", "up")),
		Down:      key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("j", "down")),
		HalfUp:   key.NewBinding(key.WithKeys("ctrl+u"), key.WithHelp("C-u", "scroll diff half page up")),
		HalfDown: key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("C-d", "scroll diff half page down")),
		Top:       key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "top")),
		Bottom:    key.NewBinding(key.WithKeys("G"), key.WithHelp("G", "bottom")),
		PrevFile:  key.NewBinding(key.WithKeys("["), key.WithHelp("[", "prev file (any pane)")),
		NextFile:  key.NewBinding(key.WithKeys("]"), key.WithHelp("]", "next file (any pane)")),
		Select:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
		Comment:     key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "comment")),
		FileComment: key.NewBinding(key.WithKeys("C"), key.WithHelp("C", "file comment")),
		Visual:      key.NewBinding(key.WithKeys("v"), key.WithHelp("v", "visual")),
		VisualLine: key.NewBinding(key.WithKeys("V"), key.WithHelp("V", "visual line")),
		Reviewed:  key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "mark reviewed")),
		ToggleDiff: key.NewBinding(key.WithKeys("t"), key.WithHelp("t", "toggle diff style")),
		TreeMode:   key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "tree mode")),
		Collapse:   key.NewBinding(key.WithKeys("z"), key.WithHelp("z", "collapse all")),
		ExpandAll:  key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "expand all")),
		ScrollMainDown:  key.NewBinding(key.WithKeys("J"), key.WithHelp("J", "scroll diff down")),
		ScrollMainUp:    key.NewBinding(key.WithKeys("K"), key.WithHelp("K", "scroll diff up")),
		ScrollMainLeft:  key.NewBinding(key.WithKeys("H"), key.WithHelp("H", "scroll diff left")),
		ScrollMainRight: key.NewBinding(key.WithKeys("L"), key.WithHelp("L", "scroll diff right")),
		Wrap:            key.NewBinding(key.WithKeys("w"), key.WithHelp("w", "toggle line wrap")),
		FocusSwap:       key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "switch pane")),
		Quit:      key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
	}
}

package tui

// KeyMap defines all configurable keybindings. Each field holds one or more
// key strings that trigger that action. Users override individual actions
// via Config.Keybindings (map action name → key string).
type KeyMap struct {
	// Navigation
	Up       []string
	Down     []string
	Top      []string
	Bottom   []string
	HalfUp   []string
	HalfDown []string
	PrevFile []string
	NextFile []string
	Select   []string

	// Pane focus
	FocusSwap   []string
	FocusPaneN  map[string]int // key → pane number (1=sidebar, 2=diff)

	// Diff view
	ScrollDown  []string
	ScrollUp    []string
	ScrollLeft  []string
	ScrollRight []string
	ScrollHome  []string
	Wrap        []string
	ToggleDiff  []string

	// Sidebar
	TreeMode    []string
	CollapseAll []string
	ExpandAll   []string

	// Review actions
	Comment     []string
	FileComment []string
	Visual      []string
	Reviewed    []string
	Submit      []string
	Pause       []string
	DismissOutdated []string

	// General
	BaseRef     []string
	CycleLayout []string
	Help        []string
	Quit        []string
	CommandMode []string
}

// DefaultKeyMap returns the built-in default keybindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up:       []string{"k", "up"},
		Down:     []string{"j", "down"},
		Top:      []string{"g"},
		Bottom:   []string{"G"},
		HalfUp:   []string{"ctrl+u"},
		HalfDown: []string{"ctrl+d"},
		PrevFile: []string{"["},
		NextFile: []string{"]"},
		Select:   []string{"enter"},

		FocusSwap:  []string{"tab"},
		FocusPaneN: map[string]int{"1": 1, "2": 2},

		ScrollDown:  []string{"J"},
		ScrollUp:    []string{"K"},
		ScrollLeft:  []string{"H"},
		ScrollRight: []string{"L"},
		ScrollHome:  []string{"0"},
		Wrap:        []string{"w"},
		ToggleDiff:  []string{"t"},

		TreeMode:    []string{"f"},
		CollapseAll: []string{"z"},
		ExpandAll:   []string{"e"},

		Comment:         []string{"c"},
		FileComment:     []string{"C"},
		Visual:          []string{"v"},
		Reviewed:        []string{"r"},
		Submit:          []string{"S"},
		Pause:           []string{"P"},
		DismissOutdated: []string{"D"},

		BaseRef:     []string{"b"},
		CycleLayout: []string{"T"},
		Help:        []string{"?"},
		Quit:        []string{"q"},
		CommandMode: []string{":"},
	}
}

// actionBindings maps config action names to pointers into the KeyMap.
// This is used by ApplyOverrides to find which field to update.
var actionNames = []string{
	"up", "down", "top", "bottom", "half_up", "half_down",
	"prev_file", "next_file", "select",
	"focus_swap",
	"scroll_down", "scroll_up", "scroll_left", "scroll_right", "scroll_home",
	"wrap", "toggle_diff",
	"tree_mode", "collapse_all", "expand_all",
	"comment", "file_comment", "visual", "reviewed",
	"submit", "pause", "dismiss_outdated",
	"base_ref", "cycle_layout", "help", "quit", "command_mode",
}

// ApplyOverrides merges user-configured keybinding overrides into the keymap.
// Each key in overrides is an action name (e.g. "quit"), value is the key string.
func (km KeyMap) ApplyOverrides(overrides map[string]string) KeyMap {
	for action, key := range overrides {
		switch action {
		case "up":
			km.Up = []string{key}
		case "down":
			km.Down = []string{key}
		case "top":
			km.Top = []string{key}
		case "bottom":
			km.Bottom = []string{key}
		case "half_up":
			km.HalfUp = []string{key}
		case "half_down":
			km.HalfDown = []string{key}
		case "prev_file":
			km.PrevFile = []string{key}
		case "next_file":
			km.NextFile = []string{key}
		case "select":
			km.Select = []string{key}
		case "focus_swap":
			km.FocusSwap = []string{key}
		case "scroll_down":
			km.ScrollDown = []string{key}
		case "scroll_up":
			km.ScrollUp = []string{key}
		case "scroll_left":
			km.ScrollLeft = []string{key}
		case "scroll_right":
			km.ScrollRight = []string{key}
		case "scroll_home":
			km.ScrollHome = []string{key}
		case "wrap":
			km.Wrap = []string{key}
		case "toggle_diff":
			km.ToggleDiff = []string{key}
		case "tree_mode":
			km.TreeMode = []string{key}
		case "collapse_all":
			km.CollapseAll = []string{key}
		case "expand_all":
			km.ExpandAll = []string{key}
		case "comment":
			km.Comment = []string{key}
		case "file_comment":
			km.FileComment = []string{key}
		case "visual":
			km.Visual = []string{key}
		case "reviewed":
			km.Reviewed = []string{key}
		case "submit":
			km.Submit = []string{key}
		case "pause":
			km.Pause = []string{key}
		case "dismiss_outdated":
			km.DismissOutdated = []string{key}
		case "base_ref":
			km.BaseRef = []string{key}
		case "cycle_layout":
			km.CycleLayout = []string{key}
		case "help":
			km.Help = []string{key}
		case "quit":
			km.Quit = []string{key}
		case "command_mode":
			km.CommandMode = []string{key}
		}
	}
	return km
}

// Matches returns true if the key string matches any of the given bindings.
func Matches(key string, bindings []string) bool {
	for _, b := range bindings {
		if key == b {
			return true
		}
	}
	return false
}

// Label returns the display label for a keybinding (first key, or joined with /).
func Label(bindings []string) string {
	if len(bindings) == 0 {
		return ""
	}
	if len(bindings) == 1 {
		return bindings[0]
	}
	return bindings[0] + "/" + bindings[1]
}

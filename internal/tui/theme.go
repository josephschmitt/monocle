package tui

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// Theme holds all styles for the TUI.
type Theme struct {
	// Layout
	SidebarBorder        lipgloss.Style
	SidebarBorderFocused lipgloss.Style
	MainPane             lipgloss.Style
	MainPaneFocused      lipgloss.Style

	// Diff colors
	Added          lipgloss.Style
	Removed        lipgloss.Style
	Context        lipgloss.Style
	HunkHeader     lipgloss.Style
	LineNumber     lipgloss.Style

	// Diff backgrounds (true color for syntax highlighting overlay)
	AddedBg         color.Color
	RemovedBg       color.Color
	AddedChangeBg   color.Color
	RemovedChangeBg color.Color

	// Comment styles
	CommentBorder  lipgloss.Style
	CommentIssue   lipgloss.Style
	CommentSuggest lipgloss.Style
	CommentNote    lipgloss.Style
	CommentPraise  lipgloss.Style

	// Status
	StatusBar      lipgloss.Style
	StatusIdle     lipgloss.Style
	StatusWorking  lipgloss.Style
	StatusStopped  lipgloss.Style

	// Modal
	ModalOverlay   lipgloss.Style
	ModalBorder    lipgloss.Style

	// Markdown
	MarkdownH1         lipgloss.Style
	MarkdownH2         lipgloss.Style
	MarkdownH3         lipgloss.Style
	MarkdownBlockquote lipgloss.Style
	MarkdownCode       lipgloss.Style
	MarkdownCodeBlock  lipgloss.Style
	MarkdownRule       lipgloss.Style
	MarkdownBullet     lipgloss.Style
}

// DefaultTheme returns a theme using 16-color ANSI for maximum compatibility.
func DefaultTheme() Theme {
	return Theme{
		SidebarBorder:        lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("8")),
		SidebarBorderFocused: lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("4")),
		MainPane:             lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("8")),
		MainPaneFocused:      lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("4")),

		Added:          lipgloss.NewStyle().Foreground(lipgloss.Color("2")),
		Removed:        lipgloss.NewStyle().Foreground(lipgloss.Color("1")),
		Context:        lipgloss.NewStyle(),
		HunkHeader:     lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Faint(true),
		LineNumber:     lipgloss.NewStyle().Foreground(lipgloss.Color("8")),

		CommentBorder:  lipgloss.NewStyle().Foreground(lipgloss.Color("3")),
		CommentIssue:   lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true),
		CommentSuggest: lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Bold(true),
		CommentNote:    lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Bold(true),
		CommentPraise:  lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true),

		StatusBar:      lipgloss.NewStyle().Background(lipgloss.Color("0")).Foreground(lipgloss.Color("7")),
		StatusIdle:     lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
		StatusWorking:  lipgloss.NewStyle().Foreground(lipgloss.Color("6")),
		StatusStopped:  lipgloss.NewStyle().Foreground(lipgloss.Color("3")),

		ModalOverlay:   lipgloss.NewStyle().Background(lipgloss.Color("0")),
		ModalBorder:    lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("4")).Padding(1, 2),

		MarkdownH1:         lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Bold(true),
		MarkdownH2:         lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Bold(true),
		MarkdownH3:         lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true),
		MarkdownBlockquote: lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Italic(true),
		MarkdownCode:       lipgloss.NewStyle().Foreground(lipgloss.Color("3")),
		MarkdownCodeBlock:  lipgloss.NewStyle().Foreground(lipgloss.Color("7")).Faint(true),
		MarkdownRule:       lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
		MarkdownBullet:     lipgloss.NewStyle().Foreground(lipgloss.Color("5")),

		AddedBg:         lipgloss.Color("#132a13"),
		RemovedBg:       lipgloss.Color("#2a1313"),
		AddedChangeBg:   lipgloss.Color("#1f4a1f"),
		RemovedChangeBg: lipgloss.Color("#4a1f1f"),
	}
}

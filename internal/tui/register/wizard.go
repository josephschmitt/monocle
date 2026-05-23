package register

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/josephschmitt/monocle/internal/adapters"
	"github.com/josephschmitt/monocle/internal/core"
	"github.com/josephschmitt/monocle/internal/tui"
)

// Model is the Bubble Tea root model for the register/unregister wizard.
type Model struct {
	state WizardState
}

// NewModel constructs the wizard from Options. An empty KeyMap triggers a
// full defaults fallback (keys + theme), so simple callers can pass a zero
// Options{Mode, Adapters} and get a working wizard. The theme fallback
// honors the user's `theme` config so a `light`-configured user gets a
// light wizard too.
func NewModel(opts Options) Model {
	if len(opts.Keys.WizardAdvance) == 0 {
		opts.Keys = tui.DefaultKeyMap()
		opts.Theme = tui.DefaultTheme()
		if cfg, err := core.LoadConfig(); err == nil && cfg != nil && cfg.Theme == "light" {
			opts.Theme = tui.LightTheme()
		}
	}
	return Model{state: NewWizardState(opts)}
}

// Result summarizes what the wizard produced for the caller.
type Result struct {
	Cancelled bool
	Results   []AgentResult
}

// AgentResult holds the outcome for one adapter after StepExecute runs it.
type AgentResult struct {
	Name   string
	Label  string
	Paths  []string
	Action string // "registered", "updated", "removed", "nothing"
	Err    error
}

// Run launches the wizard as a standalone tea.Program and returns the final
// result. Callers (cmd/monocle) use this from RegisterCmd.Run /
// UnregisterCmd.Run when no positional agent was given.
func Run(opts Options) (Result, error) {
	m := NewModel(opts)
	p := tea.NewProgram(m)
	final, err := p.Run()
	if err != nil {
		return Result{}, fmt.Errorf("wizard: %w", err)
	}
	fm, ok := final.(Model)
	if !ok {
		return Result{}, fmt.Errorf("wizard: unexpected model type %T", final)
	}
	return Result{Cancelled: fm.state.cancelled, Results: fm.state.results}, nil
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.state.width = msg.Width
		m.state.height = msg.Height
		return m, nil

	case cancelMsg:
		m.state.cancelled = true
		return m, tea.Quit

	case backMsg:
		if len(m.state.history) > 0 {
			m.state.step = m.state.history[len(m.state.history)-1]
			m.state.history = m.state.history[:len(m.state.history)-1]
		}
		return m, nil

	case advanceMsg:
		// Validate forward transitions.
		if m.state.step == StepAgents && !m.state.anySelected() {
			return m, nil
		}
		m.state.history = append(m.state.history, m.state.step)
		m.state.step = m.state.nextStep(m.state.step)
		if m.state.step == StepExecute {
			return m, enterExecute(&m.state)
		}
		return m, nil

	case runNextMsg:
		return m, runAgent(&m.state)

	case agentFinishedMsg:
		for len(m.state.results) <= msg.index {
			m.state.results = append(m.state.results, AgentResult{})
		}
		m.state.results[msg.index] = msg.result
		m.state.runIndex = msg.index + 1
		if m.state.runIndex >= len(m.state.selectedAdapters()) {
			return m, func() tea.Msg { return executeDoneMsg{} }
		}
		return m, func() tea.Msg { return runNextMsg{} }

	case executeDoneMsg:
		// User presses enter to close from here.
		return m, nil

	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Global keys: esc/q cancel from anywhere. The wizard has no text-entry
	// fields, so `q` is safe to treat as a quit alias.
	if key == "esc" || key == "q" {
		return m, func() tea.Msg { return cancelMsg{} }
	}

	// After execute finishes, enter closes the wizard.
	if m.state.step == StepExecute && m.state.runIndex >= len(m.state.selectedAdapters()) && m.state.runIndex > 0 {
		if key == "enter" {
			return m, tea.Quit
		}
		return m, nil
	}

	// Back navigation (shift+tab, backspace) — not on execute.
	if m.state.step != StepExecute && tui.Matches(key, m.state.keys.WizardBack) {
		return m, func() tea.Msg { return backMsg{} }
	}

	switch m.state.step {
	case StepAgents:
		return updateAgents(m, key)
	case StepClaude:
		return updateClaude(m, key)
	case StepConfirm:
		return updateConfirm(m, key)
	}
	return m, nil
}

func (m Model) View() tea.View {
	if m.state.width <= 0 {
		return tea.NewView("")
	}
	// lipgloss v2: Width() sets the *outer* rendered width (including border
	// and padding). Pass the full terminal width so the box fills it, then
	// subtract the chrome to know how many cols are actually available for
	// content — that's the width children use for wrapping.
	frame := m.state.theme.ModalBorder.GetHorizontalFrameSize()
	outerWidth := m.state.width
	contentWidth := outerWidth - frame
	if contentWidth < 40 {
		contentWidth = 40
		outerWidth = contentWidth + frame
	}

	header := renderHeader(m.state)

	var body string
	switch m.state.step {
	case StepAgents:
		body = viewAgents(m.state)
	case StepClaude:
		body = viewClaude(m.state, contentWidth)
	case StepConfirm:
		body = viewConfirm(m.state)
	case StepExecute:
		body = viewExecute(m.state)
	}

	footer := renderFooter()

	content := strings.Join([]string{header, body, footer}, "\n")
	return tea.NewView(m.state.theme.ModalBorder.Width(outerWidth).Render(content))
}

func renderHeader(s WizardState) string {
	verb := titleRegister
	if s.mode == ModeUnregister {
		verb = titleUnregister
	}
	title := styleLogo.Render("o_(◉) monocle") + styleFaint.Render("  │  ") + styleBold.Render(verb)
	steps := []string{"Agents", "Claude", "Confirm", "Run"}
	if !s.claudeSelected() {
		steps = []string{"Agents", "Confirm", "Run"}
	}
	active := 0
	switch s.step {
	case StepAgents:
		active = 0
	case StepClaude:
		active = 1
	case StepConfirm:
		if s.claudeSelected() {
			active = 2
		} else {
			active = 1
		}
	case StepExecute:
		active = len(steps) - 1
	}
	trail := make([]string, 0, len(steps))
	for i, label := range steps {
		if i == active {
			trail = append(trail, styleBreadcrumb.Render(label))
		} else {
			trail = append(trail, styleFaint.Render(label))
		}
	}
	trailStr := strings.Join(trail, styleFaint.Render(" › "))
	return title + "\n" + trailStr + "\n"
}

func renderFooter() string {
	return "\n" + styleFaint.Render(helpHint)
}

// lockNote renders the "(via --flag)" annotation for pre-filled fields.
func lockNote(flag string) string {
	return styleFaintItalic.Render("  (via " + flag + ")")
}

// Nerd Font glyphs for the wizard's toggle rows. Using circular variants to
// match the main TUI's icon vocabulary and read as a modern checkbox rather
// than the ASCII `[x]`/`[ ]` the old picker used.
const (
	glyphChecked   = "\uf058" //  nf-fa-check_circle
	glyphUnchecked = "\uf10c" //  nf-fa-circle_o
)

// Shared styles hoisted to package scope so View() renders (every keypress)
// don't reallocate them. All lipgloss styles are immutable values — safe to
// share across goroutines or re-render calls.
var (
	styleFaint        = lipgloss.NewStyle().Faint(true)
	styleBold         = lipgloss.NewStyle().Bold(true)
	styleFaintItalic  = lipgloss.NewStyle().Faint(true).Italic(true)
	styleCursorBar    = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true)
	styleLabelActive  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))
	styleCheckboxOn   = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	styleCheckboxLock = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Faint(true)
	styleOk           = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	styleBad          = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	styleLogo         = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("4"))
	styleBreadcrumb   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	styleCode         = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
)

func checkbox(checked, locked bool) string {
	if checked {
		if locked {
			return styleCheckboxLock.Render(glyphChecked)
		}
		return styleCheckboxOn.Render(glyphChecked)
	}
	return styleFaint.Render(glyphUnchecked)
}

// rowCursor is a 2-char gutter rendered at the start of every selectable row.
// Active rows get a colored left bar that reads as "this is where you are";
// inactive rows get blank space so the alignment stays stable.
func rowCursor(active bool) string {
	if active {
		return styleCursorBar.Render("▌") + " "
	}
	return "  "
}

// highlightLabel renders a label with the emphasis level appropriate for its
// row state. Active rows get bright + bold; inactive rows render as-is.
func highlightLabel(label string, active bool) string {
	if active {
		return styleLabelActive.Render(label)
	}
	return label
}

// indentedWrap word-wraps `text` to fit in `width` columns with each line
// prefixed by `indent` spaces. Used for descriptions that need a hanging
// indent — lipgloss's PaddingLeft+Width combo doesn't survive the outer
// ModalBorder re-flow, so we do the wrapping and prefixing ourselves.
func indentedWrap(text string, indent, width int) string {
	available := width - indent
	if available < 10 {
		available = 10
	}
	pad := strings.Repeat(" ", indent)
	words := strings.Fields(text)
	if len(words) == 0 {
		return pad
	}
	var lines []string
	cur := words[0]
	for _, w := range words[1:] {
		if lipgloss.Width(cur)+1+lipgloss.Width(w) > available {
			lines = append(lines, cur)
			cur = w
			continue
		}
		cur += " " + w
	}
	lines = append(lines, cur)
	for i, l := range lines {
		lines[i] = pad + l
	}
	return strings.Join(lines, "\n")
}

// advanceCmd emits the advanceMsg used by step updaters to walk forward.
func advanceCmd() tea.Cmd { return func() tea.Msg { return advanceMsg{} } }

// resolveIntegrationModeForAgent returns the effective mode for an agent in
// the wizard's current state.
func (s WizardState) resolveIntegrationModeForAgent(a adapters.AgentAdapter) adapters.IntegrationMode {
	return resolveIntegrationMode(a.Name(), s.integration[a.Name()])
}

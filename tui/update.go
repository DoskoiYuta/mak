package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/doskoiyuta/mak/fuzzy"
	"github.com/doskoiyuta/mak/mkfile"
)

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil

	case tea.KeyMsg:
		if m.mode == modeHelp {
			// Any key closes the help overlay.
			m.mode = modeNormal
			return m, nil
		}
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global keys.
	switch msg.String() {
	case "ctrl+c":
		m.quitting = true
		m.result = Result{Action: ExitCancel}
		return m, tea.Quit
	case "ctrl+y":
		opts, ok := m.currentResult()
		if ok {
			m.quitting = true
			m.result = opts
			m.result.Action = ExitCopy
			return m, tea.Quit
		}
		return m, nil
	case "?":
		if m.focus == focusSearch && m.search.Value() == "" {
			m.mode = modeHelp
			return m, nil
		}
	}

	switch m.focus {
	case focusSearch:
		return m.handleSearchKey(msg)
	case focusArgs:
		return m.handleArgsKey(msg)
	}
	return m, nil
}

func (m Model) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		// `q` quits only when the search field is empty, to allow typing
		// target names that contain the letter q.
		if m.search.Value() == "" {
			m.quitting = true
			m.result = Result{Action: ExitCancel}
			return m, tea.Quit
		}
	case "esc":
		m.quitting = true
		m.result = Result{Action: ExitCancel}
		return m, tea.Quit
	case "up", "ctrl+p":
		if m.selected > 0 {
			m.selected--
			m.syncVarInputsForSelection()
		}
		return m, nil
	case "down", "ctrl+n":
		if m.selected < len(m.matches)-1 {
			m.selected++
			m.syncVarInputsForSelection()
		}
		return m, nil
	case "tab":
		if len(m.varInputs) > 0 {
			m.focus = focusArgs
			m.search.Blur()
			m.varFocus = 0
			m.varInputs[0].Focus()
			return m, textinput.Blink
		}
		// No vars — fall through and treat Tab as a no-op. Enter is still
		// the way to execute.
		return m, nil
	case "enter":
		res, ok := m.currentResult()
		if ok {
			m.quitting = true
			m.result = res
			m.result.Action = ExitRun
			return m, tea.Quit
		}
		return m, nil
	}

	// Default: forward to the textinput.
	var cmd tea.Cmd
	prev := m.search.Value()
	m.search, cmd = m.search.Update(msg)
	if m.search.Value() != prev {
		m.matches = fuzzy.Filter(m.search.Value(), m.allNames)
		m.selected = 0
		m.syncVarInputsForSelection()
	}
	return m, cmd
}

func (m Model) handleArgsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.quitting = true
		m.result = Result{Action: ExitCancel}
		return m, tea.Quit
	case "shift+tab":
		if m.varFocus > 0 {
			m.varInputs[m.varFocus].Blur()
			m.varFocus--
			m.varInputs[m.varFocus].Focus()
			return m, textinput.Blink
		}
		// First variable: move back to search.
		m.varInputs[m.varFocus].Blur()
		m.focus = focusSearch
		m.search.Focus()
		return m, textinput.Blink
	case "tab":
		if len(m.varInputs) == 0 {
			return m, nil
		}
		m.varInputs[m.varFocus].Blur()
		m.varFocus = (m.varFocus + 1) % len(m.varInputs)
		m.varInputs[m.varFocus].Focus()
		return m, textinput.Blink
	case "up":
		if m.varFocus > 0 {
			m.varInputs[m.varFocus].Blur()
			m.varFocus--
			m.varInputs[m.varFocus].Focus()
			return m, textinput.Blink
		}
		return m, nil
	case "down":
		if m.varFocus < len(m.varInputs)-1 {
			m.varInputs[m.varFocus].Blur()
			m.varFocus++
			m.varInputs[m.varFocus].Focus()
			return m, textinput.Blink
		}
		return m, nil
	case "enter":
		// Enter on the last variable runs; otherwise moves to next field.
		if m.varFocus < len(m.varInputs)-1 {
			m.varInputs[m.varFocus].Blur()
			m.varFocus++
			m.varInputs[m.varFocus].Focus()
			return m, textinput.Blink
		}
		res, ok := m.currentResult()
		if ok {
			m.quitting = true
			m.result = res
			m.result.Action = ExitRun
			return m, tea.Quit
		}
		return m, nil
	}

	if len(m.varInputs) == 0 {
		return m, nil
	}
	var cmd tea.Cmd
	m.varInputs[m.varFocus], cmd = m.varInputs[m.varFocus].Update(msg)
	return m, cmd
}

// currentResult collects the in-flight selection into a Result (without
// assigning an Action — the caller sets that).
func (m *Model) currentResult() (Result, bool) {
	t, ok := m.currentTarget()
	if !ok {
		return Result{}, false
	}
	vars := make([]mkfile.VarAssignment, 0, len(m.varInputs))
	for i, name := range m.varNames {
		vars = append(vars, mkfile.VarAssignment{
			Name:  name,
			Value: m.varInputs[i].Value(),
		})
	}
	return Result{
		Target:    t.Name,
		Variables: vars,
	}, true
}

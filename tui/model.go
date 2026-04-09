// Package tui implements the Bubble Tea based user interface for mak.
package tui

import (
	"github.com/charmbracelet/bubbles/textinput"

	"github.com/doskoiyuta/mak/fuzzy"
	"github.com/doskoiyuta/mak/makefile"
)

// focusArea identifies which UI section currently has input focus.
type focusArea int

const (
	focusSearch focusArea = iota
	focusArgs
)

// mode is the overall TUI mode.
type mode int

const (
	modeNormal mode = iota
	modeHelp
)

// ExitAction tells the caller what to do after the TUI exits.
type ExitAction int

const (
	// ExitCancel means the user quit without choosing anything.
	ExitCancel ExitAction = iota
	// ExitRun means the user wants to execute the selected command.
	ExitRun
	// ExitCopy means the user wants to copy the command to the clipboard.
	ExitCopy
)

// Result is returned from Run once the TUI exits.
type Result struct {
	Action    ExitAction
	Target    string
	Variables []makefile.VarAssignment
}

// Model is the Bubble Tea model for the mak TUI.
type Model struct {
	parsed    *makefile.ParsedMakefile
	styles    Styles
	search    textinput.Model
	allNames  []string
	targetIdx map[string]int
	matches   []fuzzy.Match
	selected  int

	varInputs []textinput.Model
	varNames  []string
	varFocus  int

	focus focusArea
	mode  mode

	width  int
	height int

	quitting   bool
	result     Result
	showHelp   bool
	errMessage string
}

// New returns an initialised Model.
func New(parsed *makefile.ParsedMakefile) Model {
	ti := textinput.New()
	ti.Placeholder = "Type to search targets..."
	ti.Prompt = ""
	ti.CharLimit = 256
	ti.Focus()

	names := make([]string, len(parsed.Targets))
	idx := make(map[string]int, len(parsed.Targets))
	for i, t := range parsed.Targets {
		names[i] = t.Name
		idx[t.Name] = i
	}

	m := Model{
		parsed:    parsed,
		styles:    DefaultStyles(),
		search:    ti,
		allNames:  names,
		targetIdx: idx,
		focus:     focusSearch,
	}
	m.matches = fuzzy.Filter("", names)
	m.syncVarInputsForSelection()
	return m
}

// syncVarInputsForSelection rebuilds the variable input widgets for the
// currently selected target.
func (m *Model) syncVarInputsForSelection() {
	m.varInputs = nil
	m.varNames = nil
	m.varFocus = 0
	t, ok := m.currentTarget()
	if !ok {
		return
	}
	for _, v := range t.Variables {
		ti := textinput.New()
		ti.Prompt = ""
		ti.CharLimit = 1024
		ti.Width = 32
		if def, ok := m.parsed.Defaults[v]; ok {
			ti.SetValue(def)
		}
		m.varInputs = append(m.varInputs, ti)
		m.varNames = append(m.varNames, v)
	}
}

// currentTarget returns the target currently highlighted by the selection
// cursor.
func (m *Model) currentTarget() (makefile.Target, bool) {
	if len(m.matches) == 0 {
		return makefile.Target{}, false
	}
	if m.selected < 0 || m.selected >= len(m.matches) {
		return makefile.Target{}, false
	}
	name := m.matches[m.selected].Target
	idx, ok := m.targetIdx[name]
	if !ok {
		return makefile.Target{}, false
	}
	return m.parsed.Targets[idx], true
}

// buildRunOptions collects the current target + variable inputs into a
// makefile.RunOptions suitable for BuildArgs/PreviewCommand.
func (m *Model) buildRunOptions(makefilePath, directory string) (makefile.RunOptions, bool) {
	t, ok := m.currentTarget()
	if !ok {
		return makefile.RunOptions{}, false
	}
	vars := make([]makefile.VarAssignment, 0, len(m.varInputs))
	for i, name := range m.varNames {
		val := m.varInputs[i].Value()
		if val == "" {
			continue
		}
		vars = append(vars, makefile.VarAssignment{Name: name, Value: val})
	}
	return makefile.RunOptions{
		Makefile:  makefilePath,
		Directory: directory,
		Target:    t.Name,
		Variables: vars,
	}, true
}

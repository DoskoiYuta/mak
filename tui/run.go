package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/doskoiyuta/mak/mkfile"
)

// Run launches the TUI and blocks until the user quits. It returns the
// Result describing what the user chose to do.
func Run(parsed *mkfile.ParsedMakefile) (Result, error) {
	m := New(parsed)
	prog := tea.NewProgram(m)
	finalModel, err := prog.Run()
	if err != nil {
		return Result{}, fmt.Errorf("run tui: %w", err)
	}
	fm, ok := finalModel.(Model)
	if !ok {
		return Result{}, fmt.Errorf("unexpected model type: %T", finalModel)
	}
	return fm.result, nil
}

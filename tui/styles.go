package tui

import "github.com/charmbracelet/lipgloss"

// Styles holds all lipgloss styles used by the TUI.
type Styles struct {
	Title         lipgloss.Style
	Subtitle      lipgloss.Style
	Border        lipgloss.Style
	SearchPrompt  lipgloss.Style
	SearchInput   lipgloss.Style
	TargetCursor  lipgloss.Style
	TargetName    lipgloss.Style
	TargetNameSel lipgloss.Style
	TargetDesc    lipgloss.Style
	TargetDescSel lipgloss.Style
	MatchChar     lipgloss.Style
	MatchCharSel  lipgloss.Style
	VarName       lipgloss.Style
	VarNameFocus  lipgloss.Style
	VarInput      lipgloss.Style
	VarInputFocus lipgloss.Style
	Preview       lipgloss.Style
	PreviewHint   lipgloss.Style
	Help          lipgloss.Style
	Error         lipgloss.Style
	Empty         lipgloss.Style
}

// DefaultStyles returns the default style set.
func DefaultStyles() Styles {
	const (
		colorAccent    = "205" // pink
		colorAccent2   = "81"  // cyan
		colorSubtle    = "241"
		colorMatch     = "214" // orange
		colorMatchSel  = "226" // yellow
		colorBorder    = "238"
		colorPreviewBg = "236"
		colorError     = "196"
	)

	return Styles{
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colorAccent)),
		Subtitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorSubtle)),
		Border: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorBorder)),
		SearchPrompt: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorAccent)).
			Bold(true),
		SearchInput: lipgloss.NewStyle(),
		TargetCursor: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorAccent)).
			Bold(true),
		TargetName: lipgloss.NewStyle(),
		TargetNameSel: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorAccent2)).
			Bold(true),
		TargetDesc: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorSubtle)),
		TargetDescSel: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorSubtle)).
			Italic(true),
		MatchChar: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorMatch)).
			Bold(true),
		MatchCharSel: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorMatchSel)).
			Bold(true),
		VarName: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorSubtle)),
		VarNameFocus: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorAccent)).
			Bold(true),
		VarInput: lipgloss.NewStyle(),
		VarInputFocus: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorAccent2)),
		Preview: lipgloss.NewStyle().
			Background(lipgloss.Color(colorPreviewBg)).
			Padding(0, 1),
		PreviewHint: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorSubtle)).
			Italic(true),
		Help: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorSubtle)),
		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorError)).
			Bold(true),
		Empty: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorSubtle)).
			Italic(true),
	}
}

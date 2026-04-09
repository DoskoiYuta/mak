package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/doskoiyuta/mak/fuzzy"
	"github.com/doskoiyuta/mak/makefile"
)

const (
	minWidth        = 60
	maxTargetsShown = 8
)

// View implements tea.Model.
func (m Model) View() string {
	if m.quitting {
		return ""
	}
	if m.mode == modeHelp {
		return m.renderHelp()
	}
	if m.width > 0 && m.width < minWidth {
		return m.renderNarrow()
	}

	var sections []string

	sections = append(sections, m.renderHeader())
	sections = append(sections, m.renderSearch())
	sections = append(sections, m.renderTargets())
	if len(m.varInputs) > 0 {
		sections = append(sections, m.renderVars())
	}
	sections = append(sections, m.renderPreview())
	sections = append(sections, m.renderHelpLine())

	return strings.Join(sections, "\n")
}

func (m Model) renderHeader() string {
	title := m.styles.Title.Render("mak")
	path := m.styles.Subtitle.Render(" " + m.parsed.Path)
	return title + path
}

func (m Model) renderSearch() string {
	prompt := m.styles.SearchPrompt.Render("> ")
	return prompt + m.search.View()
}

func (m Model) renderTargets() string {
	if len(m.matches) == 0 {
		return m.styles.Empty.Render("  (ターゲットが見つかりません)")
	}

	var b strings.Builder

	start, end := visibleWindow(m.selected, len(m.matches), maxTargetsShown)

	// Determine the column width for the target name so descriptions
	// align.
	nameWidth := 0
	for i := start; i < end; i++ {
		n := lipgloss.Width(m.matches[i].Target)
		if n > nameWidth {
			nameWidth = n
		}
	}

	for i := start; i < end; i++ {
		match := m.matches[i]
		selected := i == m.selected

		cursor := "  "
		if selected {
			cursor = m.styles.TargetCursor.Render("▶ ")
		}

		name := highlightMatches(match, m.styles, selected)
		pad := strings.Repeat(" ", nameWidth-lipgloss.Width(match.Target)+2)

		desc := ""
		if idx, ok := m.targetIdx[match.Target]; ok {
			desc = m.parsed.Targets[idx].Description
		}
		if selected {
			desc = m.styles.TargetDescSel.Render(desc)
		} else {
			desc = m.styles.TargetDesc.Render(desc)
		}

		b.WriteString(cursor)
		b.WriteString(name)
		b.WriteString(pad)
		b.WriteString(desc)
		b.WriteString("\n")
	}
	// Trim trailing newline.
	return strings.TrimRight(b.String(), "\n")
}

func visibleWindow(selected, total, size int) (int, int) {
	if total <= size {
		return 0, total
	}
	half := size / 2
	start := selected - half
	if start < 0 {
		start = 0
	}
	end := start + size
	if end > total {
		end = total
		start = end - size
	}
	return start, end
}

// highlightMatches renders the target name with matched characters in a
// distinct style.
func highlightMatches(match fuzzy.Match, s Styles, selected bool) string {
	baseStyle := s.TargetName
	matchStyle := s.MatchChar
	if selected {
		baseStyle = s.TargetNameSel
		matchStyle = s.MatchCharSel
	}

	if len(match.Positions) == 0 {
		return baseStyle.Render(match.Target)
	}

	set := match.PositionSet()
	var b strings.Builder
	for i, r := range []rune(match.Target) {
		if _, ok := set[i]; ok {
			b.WriteString(matchStyle.Render(string(r)))
		} else {
			b.WriteString(baseStyle.Render(string(r)))
		}
	}
	return b.String()
}

func (m Model) renderVars() string {
	if len(m.varInputs) == 0 {
		return ""
	}
	nameWidth := 0
	for _, n := range m.varNames {
		if len(n) > nameWidth {
			nameWidth = len(n)
		}
	}

	var b strings.Builder
	for i, name := range m.varNames {
		paddedName := name + strings.Repeat(" ", nameWidth-len(name))
		focused := m.focus == focusArgs && i == m.varFocus
		if focused {
			b.WriteString(m.styles.VarNameFocus.Render(paddedName))
		} else {
			b.WriteString(m.styles.VarName.Render(paddedName))
		}
		b.WriteString("  ")
		b.WriteString(m.varInputs[i].View())
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func (m Model) renderPreview() string {
	t, ok := m.currentTarget()
	if !ok {
		return m.styles.Preview.Render("$ make")
	}
	vars := make([]makefile.VarAssignment, 0, len(m.varInputs))
	for i, name := range m.varNames {
		val := m.varInputs[i].Value()
		if val == "" {
			continue
		}
		vars = append(vars, makefile.VarAssignment{Name: name, Value: val})
	}
	cmd := makefile.PreviewCommand(makefile.RunOptions{
		Target:    t.Name,
		Variables: vars,
	})
	return m.styles.Preview.Render(cmd)
}

func (m Model) renderHelpLine() string {
	hints := []string{
		"[Enter] 実行",
		"[Tab] 変数入力",
		"[Ctrl+Y] コピー",
		"[?] ヘルプ",
		"[Ctrl+C] 終了",
	}
	return m.styles.Help.Render(strings.Join(hints, "  "))
}

func (m Model) renderHelp() string {
	lines := []string{
		m.styles.Title.Render("mak  — キーバインド"),
		"",
		"  文字入力           fuzzy 検索",
		"  ↑ / ↓ / Ctrl+P/N   ターゲット選択",
		"  Tab                変数入力欄へフォーカス移動",
		"  Shift+Tab          検索欄へ戻る",
		"  Enter              実行（変数欄では次へ、最後で実行）",
		"  Ctrl+Y             コマンドをクリップボードにコピーして終了",
		"  Ctrl+C / q / Esc   終了",
		"  ?                  このヘルプを表示",
		"",
		m.styles.PreviewHint.Render("何かキーを押すと戻ります"),
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderNarrow() string {
	// Minimal fallback layout for narrow terminals.
	var b strings.Builder
	b.WriteString(m.styles.Title.Render("mak"))
	b.WriteString("\n")
	b.WriteString("> ")
	b.WriteString(m.search.View())
	b.WriteString("\n")
	if len(m.matches) == 0 {
		b.WriteString(m.styles.Empty.Render("(no targets)"))
		return b.String()
	}
	start, end := visibleWindow(m.selected, len(m.matches), 5)
	for i := start; i < end; i++ {
		prefix := "  "
		if i == m.selected {
			prefix = "▶ "
		}
		b.WriteString(prefix)
		b.WriteString(m.matches[i].Target)
		b.WriteString("\n")
	}
	t, ok := m.currentTarget()
	if ok {
		b.WriteString(fmt.Sprintf("$ make %s", t.Name))
	}
	return b.String()
}

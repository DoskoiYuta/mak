// Package makefile parses Makefiles to extract targets, descriptions, and
// the set of variables referenced inside each target's recipe.
package makefile

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

// Target represents a single Makefile target.
type Target struct {
	// Name is the target name.
	Name string
	// Description is the text following a `## ` comment immediately above
	// the target declaration. May be empty.
	Description string
	// Variables is the list of $(VAR) / ${VAR} references that appear in
	// the recipe lines of this target, in first-seen order.
	Variables []string
	// LineNumber is the 1-based line number where the target was declared.
	LineNumber int
}

// ParsedMakefile is the result of parsing a Makefile.
type ParsedMakefile struct {
	// Path is the path that was parsed.
	Path string
	// Targets is the list of parsed targets in source order, de-duplicated
	// (first definition wins).
	Targets []Target
	// Defaults maps variable name to default value, derived from top-level
	// variable assignments in the Makefile.
	Defaults map[string]string
}

var (
	// targetLine matches a target declaration like `name:` or `name: deps`.
	// We intentionally anchor at the start of the line (no leading
	// whitespace) so recipe lines (which begin with a tab) are excluded.
	targetLine = regexp.MustCompile(`^([A-Za-z0-9_][A-Za-z0-9_.\-]*)\s*:(?:[^=]|$)`)

	// varAssignment matches `VAR = ...`, `VAR ?= ...`, `VAR := ...`.
	// `VAR != ...` is matched separately.
	varAssignment    = regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_]*)\s*(\?=|:=|=)\s*(.*)$`)
	varShellAssign   = regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_]*)\s*!=\s*(.*)$`)
	descriptionLine  = regexp.MustCompile(`^##\s?(.*)$`)
	variableInRecipe = regexp.MustCompile(`\$[({]([A-Za-z_][A-Za-z0-9_]*)[)}]`)
)

// Parse reads the Makefile at path and returns the parsed result.
func Parse(path string) (*ParsedMakefile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open makefile: %w", err)
	}
	defer f.Close()

	pm, err := parseReader(f)
	if err != nil {
		return nil, err
	}
	pm.Path = path
	return pm, nil
}

func parseReader(r io.Reader) (*ParsedMakefile, error) {
	scanner := bufio.NewScanner(r)
	// Allow long recipe lines.
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	pm := &ParsedMakefile{
		Defaults: map[string]string{},
	}

	var (
		lineNum      int
		pendingDesc  string
		seenTargets  = map[string]struct{}{}
		currentIdx   = -1 // index into pm.Targets for the in-progress target
		seenVarInTgt = map[string]struct{}{}
	)

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Recipe line (tab-indented) belongs to the current target.
		if strings.HasPrefix(line, "\t") {
			if currentIdx >= 0 {
				for _, m := range variableInRecipe.FindAllStringSubmatch(line, -1) {
					name := m[1]
					if _, ok := seenVarInTgt[name]; ok {
						continue
					}
					seenVarInTgt[name] = struct{}{}
					pm.Targets[currentIdx].Variables = append(pm.Targets[currentIdx].Variables, name)
				}
			}
			continue
		}

		trimmed := strings.TrimSpace(line)

		// Blank line resets the pending description.
		if trimmed == "" {
			pendingDesc = ""
			currentIdx = -1
			seenVarInTgt = map[string]struct{}{}
			continue
		}

		// Comment line: capture the most recent `## ` comment as a pending
		// description. Regular `#` comments reset it.
		if strings.HasPrefix(trimmed, "##") {
			if m := descriptionLine.FindStringSubmatch(trimmed); m != nil {
				pendingDesc = stripTargetPrefix(m[1])
			}
			continue
		}
		if strings.HasPrefix(trimmed, "#") {
			pendingDesc = ""
			continue
		}

		// Variable assignments (only capture if we are at top level, i.e.
		// not inside a target recipe — which we already ensured by the tab
		// check above).
		if m := varAssignment.FindStringSubmatch(trimmed); m != nil {
			name, value := m[1], strings.TrimSpace(m[3])
			if _, ok := pm.Defaults[name]; !ok {
				pm.Defaults[name] = stripInlineComment(value)
			}
			pendingDesc = ""
			currentIdx = -1
			continue
		}
		if m := varShellAssign.FindStringSubmatch(trimmed); m != nil {
			name := m[1]
			if _, ok := pm.Defaults[name]; !ok {
				pm.Defaults[name] = ""
			}
			pendingDesc = ""
			currentIdx = -1
			continue
		}

		// Target declaration.
		if m := targetLine.FindStringSubmatch(line); m != nil {
			name := m[1]
			// Skip dot-prefixed targets like .PHONY.
			if strings.HasPrefix(name, ".") {
				pendingDesc = ""
				currentIdx = -1
				continue
			}
			if _, dup := seenTargets[name]; dup {
				pendingDesc = ""
				currentIdx = -1
				continue
			}
			seenTargets[name] = struct{}{}
			pm.Targets = append(pm.Targets, Target{
				Name:        name,
				Description: pendingDesc,
				LineNumber:  lineNum,
			})
			currentIdx = len(pm.Targets) - 1
			seenVarInTgt = map[string]struct{}{}
			pendingDesc = ""
			continue
		}

		// Anything else resets pending state.
		pendingDesc = ""
		currentIdx = -1
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan makefile: %w", err)
	}
	return pm, nil
}

// stripTargetPrefix removes a leading `<target>:` from a description
// comment, turning "## build: アプリをビルドする" into "アプリをビルドする".
func stripTargetPrefix(s string) string {
	s = strings.TrimSpace(s)
	if idx := strings.Index(s, ":"); idx > 0 {
		head := s[:idx]
		// Only strip if the head looks like a target name.
		if isIdentifier(head) {
			return strings.TrimSpace(s[idx+1:])
		}
	}
	return s
}

func isIdentifier(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		if i == 0 {
			if !(isAlpha(r) || r == '_') {
				return false
			}
			continue
		}
		if !(isAlpha(r) || isDigit(r) || r == '_' || r == '-' || r == '.') {
			return false
		}
	}
	return true
}

func isAlpha(r rune) bool { return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') }
func isDigit(r rune) bool { return r >= '0' && r <= '9' }

// stripInlineComment removes a trailing `# comment` from a value, honouring
// escaped hashes (`\#`).
func stripInlineComment(v string) string {
	out := make([]byte, 0, len(v))
	prev := byte(0)
	for i := 0; i < len(v); i++ {
		c := v[i]
		if c == '#' && prev != '\\' {
			break
		}
		out = append(out, c)
		prev = c
	}
	return strings.TrimRight(string(out), " \t")
}

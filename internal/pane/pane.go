// Package pane models a single directory view: its path, entries, and cursor.
package pane

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/kiruva/lazyfiles/internal/ui"
)

// Model is one dual-pane side.
type Model struct {
	Path    string
	Entries []Entry
	Cursor  int

	width, height int

	sort       SortMode
	showHidden bool
	selected   map[string]bool // keys are entry names in the current dir
}

// New builds a pane rooted at path and loads its contents.
func New(path string) Model {
	m := Model{Path: path, selected: map[string]bool{}}
	m.reload()
	return m
}

// SetSize records the pane's outer dimensions (used for paging math and render).
func (m *Model) SetSize(w, h int) {
	m.width, m.height = w, h
}

func (m *Model) reload() {
	hasParent := filepath.Dir(m.Path) != m.Path
	entries, err := readRaw(m.Path, hasParent)
	if err != nil {
		m.Entries = nil
		m.Cursor = 0
		return
	}
	if !m.showHidden {
		filtered := entries[:0]
		for _, e := range entries {
			if !e.isDotfile() {
				filtered = append(filtered, e)
			}
		}
		entries = filtered
	}
	sortEntries(entries, m.sort)
	m.Entries = entries
	m.clampCursor()
}

func (m *Model) clampCursor() {
	if m.Cursor >= len(m.Entries) {
		m.Cursor = len(m.Entries) - 1
	}
	if m.Cursor < 0 {
		m.Cursor = 0
	}
}

// visibleRows is how many entry rows fit in the pane body (excludes borders/title).
func (m Model) visibleRows() int {
	return max(m.height-2-1, 1)
}

// Current returns the entry under the cursor, if any.
func (m *Model) Current() (Entry, bool) {
	if m.Cursor < 0 || m.Cursor >= len(m.Entries) {
		return Entry{}, false
	}
	return m.Entries[m.Cursor], true
}

// Cursor movement ------------------------------------------------------------

func (m *Model) MoveUp() {
	if m.Cursor > 0 {
		m.Cursor--
	}
}

func (m *Model) MoveDown() {
	if m.Cursor < len(m.Entries)-1 {
		m.Cursor++
	}
}

func (m *Model) PageUp()   { m.Cursor = max(m.Cursor-m.visibleRows(), 0) }
func (m *Model) PageDown() { m.Cursor = min(m.Cursor+m.visibleRows(), len(m.Entries)-1) }
func (m *Model) Top()      { m.Cursor = 0 }
func (m *Model) Bottom()   { m.Cursor = max(len(m.Entries)-1, 0) }

// Navigation -----------------------------------------------------------------

// Enter descends into the highlighted directory (or ".." to the parent).
func (m *Model) Enter() {
	cur, ok := m.Current()
	if !ok || !cur.IsDir {
		return
	}
	if cur.Name == ".." {
		m.Path = filepath.Dir(m.Path)
	} else {
		m.Path = filepath.Join(m.Path, cur.Name)
	}
	m.enterDir()
}

// Ascend moves to the parent directory.
func (m *Model) Ascend() {
	if parent := filepath.Dir(m.Path); parent != m.Path {
		m.Path = parent
		m.enterDir()
	}
}

func (m *Model) enterDir() {
	m.Cursor = 0
	m.selected = map[string]bool{} // selection is per-directory
	m.reload()
}

// Selection ------------------------------------------------------------------

// ToggleSelect flips the selection state of the current entry (never ".."),
// then advances the cursor — the familiar "tap Space down a list" flow.
func (m *Model) ToggleSelect() {
	cur, ok := m.Current()
	if !ok || cur.Name == ".." {
		return
	}
	if m.selected[cur.Name] {
		delete(m.selected, cur.Name)
	} else {
		m.selected[cur.Name] = true
	}
	m.MoveDown()
}

// SelectedCount returns how many entries are selected.
func (m *Model) SelectedCount() int { return len(m.selected) }

// SelectedNames returns the selected entry names, or the current entry's name
// if nothing is explicitly selected (the "act on cursor" fallback).
func (m *Model) SelectedNames() []string {
	if len(m.selected) > 0 {
		names := make([]string, 0, len(m.selected))
		for n := range m.selected {
			names = append(names, n)
		}
		return names
	}
	if cur, ok := m.Current(); ok && cur.Name != ".." {
		return []string{cur.Name}
	}
	return nil
}

// View options ---------------------------------------------------------------

// ToggleHidden shows or hides dotfiles.
func (m *Model) ToggleHidden() {
	m.showHidden = !m.showHidden
	m.reload()
}

// CycleSort advances to the next sort mode.
func (m *Model) CycleSort() {
	m.sort = m.sort.Next()
	m.reload()
}

// SortMode returns the active sort mode (for display).
func (m *Model) SortModeLabel() string { return m.sort.String() }

// HiddenShown reports whether dotfiles are visible.
func (m *Model) HiddenShown() bool { return m.showHidden }

// Rendering ------------------------------------------------------------------

// View renders the pane using its stored size.
func (m Model) View(active bool) string {
	border := ui.InactiveBorder
	if active {
		border = ui.ActiveBorder
	}

	innerW := max(m.width-2, 1)
	innerH := max(m.height-2, 1)

	title := ui.Title.Render(truncate(m.Path, innerW))
	rows := max(innerH-1, 1)

	start := 0
	if m.Cursor >= rows {
		start = m.Cursor - rows + 1
	}
	end := min(start+rows, len(m.Entries))

	var b strings.Builder
	b.WriteString(title)
	for i := start; i < end; i++ {
		e := m.Entries[i]
		b.WriteByte('\n')
		line := formatEntry(e, m.selected[e.Name], innerW)
		switch {
		case i == m.Cursor && active:
			line = ui.Cursor.Width(innerW).Render(line)
		case i == m.Cursor:
			line = ui.CursorInactive.Width(innerW).Render(line)
		case m.selected[e.Name]:
			line = ui.Selected.Render(line)
		case e.IsDir:
			line = ui.DirName.Render(line)
		}
		b.WriteString(line)
	}

	return border.Width(innerW).Height(innerH).Render(b.String())
}

// formatEntry lays out "gutter name ........ size" padded to width.
func formatEntry(e Entry, selected bool, width int) string {
	gutter := " "
	if selected {
		gutter = "●"
	}

	name := e.Name
	if e.IsDir && e.Name != ".." {
		name += "/"
	}

	size := ""
	if !e.IsDir {
		size = humanize(e.Size)
	}

	avail := width - lipgloss.Width(gutter) - 1 // gutter + one space after it
	gap := avail - lipgloss.Width(name) - lipgloss.Width(size)
	if gap < 1 {
		name = truncate(name, max(avail-lipgloss.Width(size)-1, 1))
		gap = max(avail-lipgloss.Width(name)-lipgloss.Width(size), 0)
	}
	return gutter + " " + name + strings.Repeat(" ", gap) + size
}

func truncate(s string, w int) string {
	if w <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= w {
		return s
	}
	if w == 1 {
		return "…"
	}
	var out strings.Builder
	for _, r := range s {
		if lipgloss.Width(out.String()+string(r)) > w-1 {
			break
		}
		out.WriteRune(r)
	}
	return out.String() + "…"
}

func humanize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%dB", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB", float64(size)/float64(div), "KMGTPE"[exp])
}

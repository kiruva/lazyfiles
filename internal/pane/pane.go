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
}

// New builds a pane rooted at path and loads its contents.
func New(path string) Model {
	m := Model{Path: path}
	m.reload()
	return m
}

func (m *Model) reload() {
	entries, err := ReadDir(m.Path)
	if err != nil {
		m.Entries = nil
		m.Cursor = 0
		return
	}
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

// Current returns the entry under the cursor, if any.
func (m *Model) Current() (Entry, bool) {
	if m.Cursor < 0 || m.Cursor >= len(m.Entries) {
		return Entry{}, false
	}
	return m.Entries[m.Cursor], true
}

// MoveUp/MoveDown move the cursor within bounds.
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
	m.Cursor = 0
	m.reload()
}

// Ascend moves to the parent directory.
func (m *Model) Ascend() {
	if parent := filepath.Dir(m.Path); parent != m.Path {
		m.Path = parent
		m.Cursor = 0
		m.reload()
	}
}

// View renders the pane into a box of the given outer width/height.
func (m Model) View(width, height int, active bool) string {
	border := ui.InactiveBorder
	if active {
		border = ui.ActiveBorder
	}

	innerW := max(width-2, 1)  // minus left/right border
	innerH := max(height-2, 1) // minus top/bottom border

	title := ui.Title.Render(truncate(m.Path, innerW))
	rows := max(innerH-1, 1) // reserve one line for the title

	start := 0
	if m.Cursor >= rows {
		start = m.Cursor - rows + 1
	}
	end := min(start+rows, len(m.Entries))

	var b strings.Builder
	b.WriteString(title)
	for i := start; i < end; i++ {
		b.WriteByte('\n')
		line := formatEntry(m.Entries[i], innerW)
		switch {
		case i == m.Cursor && active:
			line = ui.Cursor.Width(innerW).Render(line)
		case i == m.Cursor:
			line = ui.CursorInactive.Width(innerW).Render(line)
		case m.Entries[i].IsDir:
			line = ui.DirName.Render(line)
		}
		b.WriteString(line)
	}

	return border.Width(innerW).Height(innerH).Render(b.String())
}

// formatEntry lays out "name ........ size" padded to width.
func formatEntry(e Entry, width int) string {
	name := e.Name
	if e.IsDir && e.Name != ".." {
		name += "/"
	}

	size := ""
	if !e.IsDir {
		size = humanize(e.Size)
	}

	gap := width - lipgloss.Width(name) - lipgloss.Width(size)
	if gap < 1 {
		name = truncate(name, max(width-lipgloss.Width(size)-1, 1))
		gap = max(width-lipgloss.Width(name)-lipgloss.Width(size), 0)
	}
	return name + strings.Repeat(" ", gap) + size
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

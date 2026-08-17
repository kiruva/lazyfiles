package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/kiruva/lazyfiles/internal/ui"
)

// View implements tea.Model.
func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "loading…"
	}

	body := lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.panes[0].View(m.active == 0),
		m.panes[1].View(m.active == 1),
	)
	return lipgloss.JoinVertical(lipgloss.Left, body, m.statusBar())
}

func (m Model) statusBar() string {
	p := &m.panes[m.active]

	left := p.Path

	right := fmt.Sprintf("%d items", len(p.Entries))
	if n := p.SelectedCount(); n > 0 {
		right += fmt.Sprintf(" · %d selected", n)
	}
	right += fmt.Sprintf(" · sort:%s", p.SortModeLabel())
	if p.HiddenShown() {
		right += " · hidden"
	}

	pad := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if pad < 1 {
		pad = 1
	}
	bar := left + strings.Repeat(" ", pad) + right
	return ui.StatusBar.Width(m.width).Render(bar)
}

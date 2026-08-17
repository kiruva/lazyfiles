package app

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/kiruva/lazyfiles/internal/ui"
)

// View implements tea.Model.
func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "loading…"
	}

	bodyH := m.height - 1 // reserve one line for the status bar
	leftW := m.width / 2
	rightW := m.width - leftW

	left := m.panes[0].View(leftW, bodyH, m.active == 0)
	right := m.panes[1].View(rightW, bodyH, m.active == 1)
	body := lipgloss.JoinHorizontal(lipgloss.Top, left, right)

	return lipgloss.JoinVertical(lipgloss.Left, body, m.statusBar())
}

func (m Model) statusBar() string {
	loc := m.panes[m.active].Path
	help := "tab: switch   j/k: move   l: open   h: up   q: quit"

	pad := m.width - lipgloss.Width(loc) - lipgloss.Width(help)
	if pad < 1 {
		pad = 1
	}
	bar := loc + strings.Repeat(" ", pad) + help
	return ui.StatusBar.Width(m.width).Render(bar)
}

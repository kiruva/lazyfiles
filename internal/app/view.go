package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/kiruva/lazyfiles/internal/fileops"
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
	base := lipgloss.JoinVertical(lipgloss.Left, body, m.statusBar())

	switch m.mode {
	case modeConfirm:
		return overlay(m.width, m.height, m.renderConfirm())
	case modeProgress:
		return overlay(m.width, m.height, m.renderProgress())
	default:
		return base
	}
}

// overlay centers a modal box on a blank screen of the given size.
func overlay(w, h int, dialog string) string {
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, dialog)
}

func (m Model) statusBar() string {
	if m.errText != "" {
		return ui.ErrorBar.Width(m.width).Render(" " + m.errText)
	}

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
	return ui.StatusBar.Width(m.width).Render(left + strings.Repeat(" ", pad) + right)
}

func (m Model) renderConfirm() string {
	j := m.pending
	n := len(j.Srcs)

	var title, body string
	if j.Op == fileops.OpDelete {
		title = ui.Danger.Render(fmt.Sprintf("Delete %d %s?", n, items(n)))
		body = "This cannot be undone."
	} else {
		title = ui.DialogTitle.Render(fmt.Sprintf("%s %d %s", j.Op, n, items(n)))
		body = "→ " + truncTail(j.Dest, 44)
		if m.willOverwrite {
			body += "\n" + ui.Danger.Render("Existing files will be overwritten.")
		}
	}

	prompt := ui.DialogHint.Render("y") + " confirm    " + ui.DialogHint.Render("n") + " cancel"
	content := lipgloss.JoinVertical(lipgloss.Left, title, "", body, "", prompt)
	return ui.Dialog.Render(content)
}

func (m Model) renderProgress() string {
	p := m.progress
	title := ui.DialogTitle.Render(m.pending.Op.Present() + "…")
	bar := progressBar(p.Done, p.Total, 30)
	stat := fmt.Sprintf("%d / %d", p.Done, p.Total)
	cur := ui.Faint.Render(truncTail(p.Current, 44))

	content := lipgloss.JoinVertical(lipgloss.Left, title, "", bar+"  "+stat, cur)
	return ui.Dialog.Render(content)
}

func progressBar(done, total, width int) string {
	if total <= 0 {
		total = 1
	}
	filled := done * width / total
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}
	return ui.BarFilled.Render(strings.Repeat("█", filled)) +
		ui.BarEmpty.Render(strings.Repeat("░", width-filled))
}

func items(n int) string {
	if n == 1 {
		return "item"
	}
	return "items"
}

// truncTail keeps the tail of a path, prefixing "…" when it's too long.
func truncTail(s string, w int) string {
	if lipgloss.Width(s) <= w {
		return s
	}
	r := []rune(s)
	if w <= 1 {
		return "…"
	}
	return "…" + string(r[len(r)-(w-1):])
}

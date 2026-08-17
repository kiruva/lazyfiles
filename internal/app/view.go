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
	switch j.Op {
	case fileops.OpDelete:
		title = ui.Danger.Render(fmt.Sprintf("Delete %d %s?", n, items(n)))
		body = "This cannot be undone."
	case fileops.OpPack:
		title = ui.DialogTitle.Render(fmt.Sprintf("Pack %d %s", n, items(n)))
		body = "→ " + truncTail(j.Out, 44)
	case fileops.OpUnpack:
		title = ui.DialogTitle.Render(fmt.Sprintf("Unpack %d %s", n, archives(n)))
		body = "→ " + truncTail(j.Dest, 44)
	case fileops.OpUnwrap:
		title = ui.DialogTitle.Render(fmt.Sprintf("Unpack %d %s here", n, archives(n)))
		body = "→ " + truncTail(j.Dest, 44)
	default: // copy / move
		title = ui.DialogTitle.Render(fmt.Sprintf("%s %d %s", j.Op, n, items(n)))
		body = "→ " + truncTail(j.Dest, 44)
	}
	if m.willOverwrite {
		body += "\n" + ui.Danger.Render("Existing files will be overwritten.")
	}

	prompt := ui.DialogHint.Render("y") + " confirm    " + ui.DialogHint.Render("n") + " cancel"
	content := lipgloss.JoinVertical(lipgloss.Left, title, "", body, "", prompt)
	return ui.Dialog.Render(content)
}

func (m Model) renderProgress() string {
	p := m.progress
	title := ui.DialogTitle.Render(m.pending.Op.Present() + "…")

	var line string
	if p.Total > 0 {
		line = progressBar(p.Done, p.Total, 30) + fmt.Sprintf("  %d / %d", p.Done, p.Total)
	} else {
		// Total unknown (e.g. 7z/rar): show a static bar and a running count.
		line = ui.BarEmpty.Render(strings.Repeat("░", 30)) + fmt.Sprintf("  %d %s", p.Done, items(p.Done))
	}
	cur := ui.Faint.Render(truncTail(p.Current, 44))

	content := lipgloss.JoinVertical(lipgloss.Left, title, "", line, cur)
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

func archives(n int) string {
	if n == 1 {
		return "archive"
	}
	return "archives"
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

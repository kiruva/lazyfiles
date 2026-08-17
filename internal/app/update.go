package app

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resizePanes()

	case tea.KeyMsg:
		// &m.panes[...] points into this value-receiver copy, so mutating it
		// and returning m is safe and idiomatic.
		p := &m.panes[m.active]
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Up):
			p.MoveUp()
		case key.Matches(msg, m.keys.Down):
			p.MoveDown()
		case key.Matches(msg, m.keys.PageUp):
			p.PageUp()
		case key.Matches(msg, m.keys.PageDown):
			p.PageDown()
		case key.Matches(msg, m.keys.Top):
			p.Top()
		case key.Matches(msg, m.keys.Bottom):
			p.Bottom()
		case key.Matches(msg, m.keys.Enter):
			p.Enter()
		case key.Matches(msg, m.keys.Back):
			p.Ascend()
		case key.Matches(msg, m.keys.Select):
			p.ToggleSelect()
		case key.Matches(msg, m.keys.Hidden):
			p.ToggleHidden()
		case key.Matches(msg, m.keys.Sort):
			p.CycleSort()
		case key.Matches(msg, m.keys.Switch):
			m.active = 1 - m.active
		}
	}
	return m, nil
}

// resizePanes splits the terminal into two side-by-side panes above the status bar.
func (m *Model) resizePanes() {
	bodyH := m.height - 1 // reserve one line for the status bar
	leftW := m.width / 2
	m.panes[0].SetSize(leftW, bodyH)
	m.panes[1].SetSize(m.width-leftW, bodyH)
}

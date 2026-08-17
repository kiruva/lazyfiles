// Package app is the root Bubble Tea model wiring the two panes together.
package app

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/kiruva/lazyfiles/internal/pane"
)

// Model is the top-level application state.
type Model struct {
	panes  [2]pane.Model
	active int // 0 = left, 1 = right
	width  int
	height int
	keys   keyMap
}

// New constructs the app with both panes rooted at the current directory.
func New() Model {
	wd, err := os.Getwd()
	if err != nil {
		wd = string(os.PathSeparator)
	}
	return Model{
		panes:  [2]pane.Model{pane.New(wd), pane.New(wd)},
		active: 0,
		keys:   defaultKeys(),
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd { return nil }

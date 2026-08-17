// Package ui holds the Lip Gloss theme shared across lazyfiles.
package ui

import "github.com/charmbracelet/lipgloss"

// Palette
var (
	accent   = lipgloss.Color("39")  // bright blue — active pane / cursor
	dim      = lipgloss.Color("240") // grey — inactive borders
	fgBright = lipgloss.Color("231") // near-white text
	barBg    = lipgloss.Color("237") // status bar background
)

var (
	// ActiveBorder wraps the pane that currently has focus.
	ActiveBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(accent)

	// InactiveBorder wraps the pane without focus.
	InactiveBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(dim)

	// Cursor highlights the selected row in the active pane.
	Cursor = lipgloss.NewStyle().
		Background(accent).
		Foreground(fgBright).
		Bold(true)

	// CursorInactive highlights the selected row in the unfocused pane.
	CursorInactive = lipgloss.NewStyle().
			Background(dim).
			Foreground(fgBright)

	// DirName styles directory entries.
	DirName = lipgloss.NewStyle().
		Foreground(accent).
		Bold(true)

	// Title styles the path shown at the top of each pane.
	Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Bold(true)

	// StatusBar styles the bottom bar.
	StatusBar = lipgloss.NewStyle().
			Foreground(fgBright).
			Background(barBg)
)

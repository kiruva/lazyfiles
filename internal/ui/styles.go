// Package ui holds the Lip Gloss theme shared across lazyfiles.
package ui

import "github.com/charmbracelet/lipgloss"

// Palette
var (
	accent   = lipgloss.Color("39")  // bright blue — active pane / cursor
	dim      = lipgloss.Color("240") // grey — inactive borders
	fgBright = lipgloss.Color("231") // near-white text
	barBg    = lipgloss.Color("237") // status bar background
	danger   = lipgloss.Color("196") // red — destructive actions / errors
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

	// Selected styles entries the user has marked with Space.
	Selected = lipgloss.NewStyle().
			Foreground(lipgloss.Color("220")). // amber
			Bold(true)

	// Title styles the path shown at the top of each pane.
	Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Bold(true)

	// StatusBar styles the bottom bar.
	StatusBar = lipgloss.NewStyle().
			Foreground(fgBright).
			Background(barBg)

	// ErrorBar styles the bottom bar when an operation failed.
	ErrorBar = lipgloss.NewStyle().
			Foreground(fgBright).
			Background(danger).
			Bold(true)

	// Dialog is the bordered box for confirm/progress modals.
	Dialog = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(accent).
		Padding(1, 3)

	// DialogTitle styles the modal heading.
	DialogTitle = lipgloss.NewStyle().
			Foreground(accent).
			Bold(true)

	// DialogHint styles the [y]/[n] key chips in a modal.
	DialogHint = lipgloss.NewStyle().
			Foreground(fgBright).
			Background(accent).
			Bold(true).
			Padding(0, 1)

	// Danger styles destructive/overwrite warnings.
	Danger = lipgloss.NewStyle().
		Foreground(danger).
		Bold(true)

	// Faint styles secondary text (e.g. the current file in progress).
	Faint = lipgloss.NewStyle().Foreground(dim)

	// BarFilled / BarEmpty render the progress bar.
	BarFilled = lipgloss.NewStyle().Foreground(accent)
	BarEmpty  = lipgloss.NewStyle().Foreground(dim)
)

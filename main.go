package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/kiruva/lazyfiles/internal/app"
)

// version is overridden at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	for _, arg := range os.Args[1:] {
		if arg == "--version" || arg == "-v" {
			fmt.Println("lazyfiles", version)
			return
		}
	}

	p := tea.NewProgram(app.New(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "lazyfiles:", err)
		os.Exit(1)
	}
}

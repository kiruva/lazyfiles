package ui

import (
	"strings"
	"testing"
)

func TestThemeLookup(t *testing.T) {
	if _, ok := ThemeByName("  NORD "); !ok {
		t.Fatal("lookup should be case- and space-insensitive")
	}
	if _, ok := ThemeByName("nope"); ok {
		t.Fatal("unknown theme reported as known")
	}
	if ThemeIndex("nope") != 0 {
		t.Fatal("unknown theme should index to the default")
	}
	if got := ThemeNames()[0]; got != "default" {
		t.Fatalf("first theme = %q, want default", got)
	}
	if !strings.Contains(ThemeList(), "nord") {
		t.Fatalf("ThemeList() = %q", ThemeList())
	}
}

func TestThemesAreComplete(t *testing.T) {
	seen := map[string]bool{}
	for _, th := range Themes() {
		if seen[th.Name] {
			t.Fatalf("duplicate theme %q", th.Name)
		}
		seen[th.Name] = true

		fields := map[string]string{
			"Accent":   string(th.Accent),
			"Dim":      string(th.Dim),
			"Fg":       string(th.Fg),
			"Title":    string(th.Title),
			"Mark":     string(th.Mark),
			"Bar":      string(th.Bar),
			"Danger":   string(th.Danger),
			"CursorFg": string(th.CursorFg),
		}
		for name, v := range fields {
			if v == "" {
				t.Fatalf("theme %q has empty %s", th.Name, name)
			}
		}
	}
}

func TestApplyRebuildsStyles(t *testing.T) {
	t.Cleanup(func() { Apply(Themes()[0]) })

	nord, _ := ThemeByName("nord")
	Apply(nord)
	if Current().Name != "nord" {
		t.Fatalf("Current() = %q", Current().Name)
	}
	if got := DirName.GetForeground(); got != nord.Accent {
		t.Fatalf("DirName foreground = %v, want %v", got, nord.Accent)
	}

	dracula, _ := ThemeByName("dracula")
	Apply(dracula)
	if got := DirName.GetForeground(); got != dracula.Accent {
		t.Fatalf("styles did not follow the second Apply: %v", got)
	}
}

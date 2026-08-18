package remote

import (
	"os"
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	tests := []struct {
		in    string
		host  Host
		path  string
		match bool
	}{
		{"ssh://user@example.com/var/log", Host{User: "user", Name: "example.com"}, "/var/log", true},
		{"ssh://example.com:2222/srv", Host{Name: "example.com", Port: "2222"}, "/srv", true},
		{"ssh://example.com", Host{Name: "example.com"}, "", true},
		{"user@example.com:/var/log", Host{User: "user", Name: "example.com"}, "/var/log", true},
		{"box:", Host{Name: "box"}, "", true},
		{"box:relative/dir", Host{Name: "box"}, "relative/dir", true},

		// Local paths must never be mistaken for hosts.
		{"/etc", Host{}, "", false},
		{"./sub", Host{}, "", false},
		{"~/dev", Host{}, "", false},
		{"plainword", Host{}, "", false},
		{"", Host{}, "", false},
		{"/tmp/a:b", Host{}, "", false},
		{"@host:/x", Host{}, "", false},
	}

	for _, tc := range tests {
		h, p, ok := Parse(tc.in)
		if ok != tc.match {
			t.Fatalf("Parse(%q) matched = %v, want %v", tc.in, ok, tc.match)
		}
		if !ok {
			continue
		}
		if h != tc.host || p != tc.path {
			t.Fatalf("Parse(%q) = %+v %q, want %+v %q", tc.in, h, p, tc.host, tc.path)
		}
	}
}

func TestHostDisplay(t *testing.T) {
	h := Host{User: "kim", Name: "box", Port: "2222"}
	if got := h.Target(); got != "kim@box" {
		t.Fatalf("Target() = %q", got)
	}
	if got := h.String(); got != "kim@box:2222" {
		t.Fatalf("String() = %q", got)
	}
	if got := h.Display("/srv"); got != "kim@box:2222:/srv" {
		t.Fatalf("Display() = %q", got)
	}
	if (Host{}).IsZero() != true {
		t.Fatal("zero host should report IsZero")
	}
}

// TestShQuote guards the boundary where our strings become someone else's
// shell command.
func TestShQuote(t *testing.T) {
	tests := map[string]string{
		"plain":            "'plain'",
		"with space":       "'with space'",
		"it's":             `'it'\''s'`,
		"$(rm -rf /)":      "'$(rm -rf /)'",
		"`whoami`":         "'`whoami`'",
		"a;b&&c|d":         "'a;b&&c|d'",
		"back\\slash":      `'back\slash'`,
		"new\nline":        "'new\nline'",
		"$HOME":            "'$HOME'",
		"quote'and space'": `'quote'\''and space'\'''`,
	}
	for in, want := range tests {
		if got := shQuote(in); got != want {
			t.Fatalf("shQuote(%q) = %s, want %s", in, got, want)
		}
	}

	if got := quoteAll([]string{"a b", "c"}); got != "'a b' 'c'" {
		t.Fatalf("quoteAll = %s", got)
	}
}

func TestJoinAndParent(t *testing.T) {
	if got := Join("/srv/www", "logs"); got != "/srv/www/logs" {
		t.Fatalf("Join relative = %q", got)
	}
	if got := Join("/srv/www", "/etc"); got != "/etc" {
		t.Fatalf("Join absolute = %q", got)
	}
	if got := Join("/srv/www", ".."); got != "/srv" {
		t.Fatalf("Join .. = %q", got)
	}
	if got := Join("/srv", "~/dev"); got != "~/dev" {
		t.Fatalf("Join tilde = %q", got)
	}
	if got := Parent("/srv/www"); got != "/srv" {
		t.Fatalf("Parent = %q", got)
	}
	if got := Parent("/"); got != "/" {
		t.Fatalf("Parent(/) = %q", got)
	}
}

func TestParseListLine(t *testing.T) {
	tests := []struct {
		line  string
		name  string
		isDir bool
		size  int64
		epoch int64
		ok    bool
	}{
		{"drwxr-xr-x 2 1000 1000 4096 1712345678 projects", "projects", true, 4096, 1712345678, true},
		{"-rw-r--r-- 1 1000 1000  512 1712345678 notes.txt", "notes.txt", false, 512, 1712345678, true},
		{"-rw-r--r-- 1 1000 1000  512 1712345678 two words.txt", "two words.txt", false, 512, 1712345678, true},
		{"lrwxrwxrwx 1 0 0 7 1712345678 link -> /etc/passwd", "link", false, 7, 1712345678, true},
		{"crw-rw-rw- 1 0 0 1, 3 1712345678 null", "null", false, 0, 1712345678, true},
		{"drwxr-xr-x 4 501 20 128 Apr  5 09:31 Documents", "Documents", true, 128, 0, true},
		{"total 24", "", false, 0, 0, false},
		{"", "", false, 0, 0, false},
		{"garbage", "", false, 0, 0, false},
	}

	for _, tc := range tests {
		e, ok := parseListLine(tc.line)
		if ok != tc.ok {
			t.Fatalf("parseListLine(%q) ok = %v, want %v", tc.line, ok, tc.ok)
		}
		if !ok {
			continue
		}
		if e.Name != tc.name || e.IsDir != tc.isDir || e.Size != tc.size {
			t.Fatalf("parseListLine(%q) = %+v", tc.line, e)
		}
		if tc.epoch != 0 && !e.ModTime.Equal(time.Unix(tc.epoch, 0)) {
			t.Fatalf("parseListLine(%q) mtime = %v", tc.line, e.ModTime)
		}
	}
}

func TestParseModeFlags(t *testing.T) {
	if m := parseMode("drwxr-xr-x"); m&os.ModeDir == 0 {
		t.Fatal("directory bit missing")
	}
	if m := parseMode("lrwxrwxrwx"); m&os.ModeSymlink == 0 {
		t.Fatal("symlink bit missing")
	}
	if m := parseMode("-rwsr-xr-x"); m.Perm()&0o100 == 0 {
		t.Fatal("setuid execute bit should still count as executable")
	}
	if got := parseMode("-rw-r--r--").Perm(); got != 0o644 {
		t.Fatalf("perm = %o, want 644", got)
	}
}

func TestParseTarVerbose(t *testing.T) {
	tests := map[string]string{
		"x home/kim/notes.txt":                     "home/kim/notes.txt",
		"a src/main.go":                            "src/main.go",
		"docs/readme.md":                           "docs/readme.md",
		"tar: Removing leading '/' from names":     "",
		"Warning: Permanently added host to known": "",
	}
	for in, want := range tests {
		if got := parseTarVerbose(in); got != want {
			t.Fatalf("parseTarVerbose(%q) = %q, want %q", in, got, want)
		}
	}
}

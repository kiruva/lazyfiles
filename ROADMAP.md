# lazyfiles — Roadmap

A TUI file manager to rule them all — dual-pane, keyboard-driven, and as intuitive as lazygit.
Built for Linux first, comfortable on macOS.

## Stack

- **Language:** Go
- **TUI runtime:** [Bubble Tea](https://github.com/charmbracelet/bubbletea) (The Elm Architecture)
- **Styling / layout:** [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- **Widgets:** [Bubbles](https://github.com/charmbracelet/bubbles) (`key`, `help`, `viewport`, `textinput`)
- **Archives:** shell out to system tools (`tar`, `unzip`, `zip`, `7z`) — zero heavy deps

## Architecture

One root `Model` (The Elm Architecture) owns everything:

```
lazyfiles/
├── main.go                 # tea.NewProgram(app).Run()
└── internal/
    ├── app/                # root Model: panes[2], active, mode, keymap
    │   ├── model.go
    │   ├── update.go       # routes msgs → active pane / dialogs / ops
    │   ├── view.go         # Lip Gloss: two panes + status bar
    │   └── keys.go         # keymap (bubbles/key)
    ├── pane/               # one directory pane: cwd, entries, cursor, selection
    │   ├── pane.go
    │   └── entry.go        # FileEntry + directory reading/sorting
    ├── fs/                 # (Phase 2+) copy/move/delete/archive as tea.Cmd
    │   ├── ops.go
    │   └── archive.go
    └── ui/                 # Lip Gloss theme + dialogs
        └── styles.go
```

**Two decisions baked in from day one — the things that make it feel "kickass":**

1. **Active-pane visual language.** The active pane gets a highlighted border; the cursor
   row is reverse-video. Copy/move is always "active pane → inactive pane", so `Tab` is the
   only mental model you need.
2. **Non-blocking operations.** Every filesystem action runs in a goroutine and streams
   progress back through Bubble Tea as messages — the UI never freezes during a recursive copy.

## Phases

- [x] **Phase 0 — Scaffold.** Running two-pane skeleton: reads real dirs, `j/k` move,
      `l`/`enter` open, `h` up, `Tab` switch pane, `q` quit. ← **you are here**
- [ ] **Phase 1 — Navigation polish.** Sorting modes, hidden-file toggle, scroll/paging,
      selection with `Space`, breadcrumb + item counts in status bar.
- [ ] **Phase 2 — File operations.** `F5` copy · `F6` move · `F8`/`Del` delete — all
      recursive, all with a live progress dialog and a confirm/overwrite prompt.
- [ ] **Phase 3 — Archives.** Unpack → other pane · Pack selection · Unwrap in place,
      via the same progress plumbing (shelling out to `tar`/`zip`/`7z`).
- [ ] **Phase 4 — Public polish.** `?` help overlay, README + demo GIF, MIT license, CI
      (`vet`/`test`/`golangci-lint`), `go install` + Homebrew instructions. Tag `v0.1.0`.

## Keybindings (current)

| Key            | Action              |
| -------------- | ------------------- |
| `j` / `↓`      | Move cursor down    |
| `k` / `↑`      | Move cursor up      |
| `l` / `enter`  | Open directory      |
| `h` / `⌫`      | Go to parent dir    |
| `Tab`          | Switch active pane  |
| `q` / `Ctrl+C` | Quit                |

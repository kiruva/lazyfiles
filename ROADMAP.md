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
      `l`/`enter` open, `h` up, `Tab` switch pane, `q` quit.
- [x] **Phase 1 — Navigation polish.** Multi-select with `Space`, sort modes (`s`:
      name/size/time), hidden-file toggle (`.`), page/top/bottom paging, and a status bar
      with item count, selection count, sort mode, and hidden indicator. ← **you are here**
- [x] **Phase 2 — File operations.** `F5`/`c` copy · `F6`/`m` move · `F8`/`Del`/`d` delete —
      all recursive, active pane → other pane, off the UI thread via a `fileops` engine that
      streams progress. Confirm dialog (with overwrite warning), live progress bar, and both
      panes auto-refresh on completion. ← **you are here**
- [x] **Phase 3 — Archives.** `p` pack selection → other pane (`.tar.gz`) · `u` unpack →
      other pane · `U` unpack in place, all shelling out to `tar`/`unzip`/`7z`/`unrar` through
      the same progress plumbing (verbose output parsed for per-file progress). ← **you are here**
- [ ] **Phase 4 — Public polish.** `?` help overlay, README + demo GIF, MIT license, CI
      (`vet`/`test`/`golangci-lint`), `go install` + Homebrew instructions. Tag `v0.1.0`.

## Keybindings (current)

| Key                | Action                          |
| ------------------ | ------------------------------- |
| `j` / `↓`          | Move cursor down                |
| `k` / `↑`          | Move cursor up                  |
| `PgDn` / `Ctrl+D`  | Page down                       |
| `PgUp` / `Ctrl+U`  | Page up                         |
| `g` / `Home`       | Jump to top                     |
| `G` / `End`        | Jump to bottom                  |
| `l` / `enter`      | Open directory                  |
| `h` / `⌫`          | Go to parent dir                |
| `Tab`              | Switch active pane              |
| `Space`            | Select / deselect entry         |
| `s`                | Cycle sort (name → size → time) |
| `.`                | Toggle hidden files             |
| `F5` / `c`         | Copy selection → other pane     |
| `F6` / `m`         | Move selection → other pane     |
| `F8` / `Del` / `d` | Delete selection                |
| `p`                | Pack selection → other pane     |
| `u`                | Unpack archive → other pane     |
| `U`                | Unpack archive in place         |
| `y` / `n`          | Confirm / cancel a prompt       |
| `q` / `Ctrl+C`     | Quit                            |

Operations act on the marked selection, or on the highlighted entry when nothing is marked.

## External tools

Archive actions shell out to standard CLIs, used only when you touch that format:

| Format                         | Extract   | Create |
| ------------------------------ | --------- | ------ |
| `.tar`, `.tar.gz`/`.tgz`, `.tar.bz2`, `.tar.xz`, `.tar.zst` | `tar` | `tar` (pack target) |
| `.zip`                         | `unzip`   | —      |
| `.7z`                          | `7z`      | —      |
| `.rar`                         | `unrar`   | —      |

`pack` always writes `.tar.gz`. Extract entry-count progress uses `tar -t` / `zipinfo`; `7z`
and `unrar` show an indeterminate bar.

# lazyfiles

> A TUI file manager to rule them all — dual-pane, keyboard-driven, as intuitive as lazygit.

Built for the Linux community first, comfortable on macOS. Leans on what Total Commander is
to Windows, reimagined for the terminal.

> **Status:** early days — Phase 2 (dual-pane navigation + recursive copy/move/delete with
> live progress). See [ROADMAP.md](ROADMAP.md).

## Run it

```sh
go run .
```

## Keys

| Key                | Action                          |
| ------------------ | ------------------------------- |
| `j` / `↓`          | Move cursor down                |
| `k` / `↑`          | Move cursor up                  |
| `PgDn` / `PgUp`    | Page down / up                  |
| `g` / `G`          | Jump to top / bottom            |
| `l` / `enter`      | Open directory                  |
| `h` / `⌫`          | Go to parent dir                |
| `Tab`              | Switch active pane              |
| `Space`            | Select / deselect entry         |
| `s`                | Cycle sort (name → size → time) |
| `.`                | Toggle hidden files             |
| `F5` / `c`         | Copy selection → other pane     |
| `F6` / `m`         | Move selection → other pane     |
| `F8` / `Del` / `d` | Delete selection                |
| `y` / `n`          | Confirm / cancel a prompt       |
| `q` / `Ctrl+C`     | Quit                            |

Operations act on the marked selection (`Space`), or on the highlighted entry when nothing
is marked. Copy/move always go from the **active** pane to the **other** pane.

## Built with

[Bubble Tea](https://github.com/charmbracelet/bubbletea) ·
[Lip Gloss](https://github.com/charmbracelet/lipgloss) ·
[Bubbles](https://github.com/charmbracelet/bubbles)

## License

[MIT](LICENSE)

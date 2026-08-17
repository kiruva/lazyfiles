# lazyfiles

> A TUI file manager to rule them all — dual-pane, keyboard-driven, as intuitive as lazygit.

Built for the Linux community first, comfortable on macOS. Leans on what Total Commander is
to Windows, reimagined for the terminal.

> **Status:** early days — Phase 3 (dual-pane navigation, recursive copy/move/delete, and
> pack/unpack archives) with live progress. See [ROADMAP.md](ROADMAP.md).

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
| `p`                | Pack selection → other pane     |
| `u`                | Unpack archive → other pane     |
| `U`                | Unpack archive in place         |
| `y` / `n`          | Confirm / cancel a prompt       |
| `q` / `Ctrl+C`     | Quit                            |

Operations act on the marked selection (`Space`), or on the highlighted entry when nothing
is marked. Copy/move/pack/unpack always go from the **active** pane to the **other** pane.

Archive actions shell out to `tar`, `unzip`, `7z`, and `unrar` (only the tool for the format
you touch is required). `pack` writes a `.tar.gz`.

## Built with

[Bubble Tea](https://github.com/charmbracelet/bubbletea) ·
[Lip Gloss](https://github.com/charmbracelet/lipgloss) ·
[Bubbles](https://github.com/charmbracelet/bubbles)

## License

[MIT](LICENSE)

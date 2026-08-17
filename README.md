# lazyfiles

> A TUI file manager to rule them all — dual-pane, keyboard-driven, as intuitive as lazygit.

Built for the Linux community first, comfortable on macOS. Leans on what Total Commander is
to Windows, reimagined for the terminal.

> **Status:** early days — Phase 0 (running dual-pane skeleton). See [ROADMAP.md](ROADMAP.md).

## Run it

```sh
go run .
```

## Keys

| Key            | Action              |
| -------------- | ------------------- |
| `j` / `↓`      | Move cursor down    |
| `k` / `↑`      | Move cursor up      |
| `l` / `enter`  | Open directory      |
| `h` / `⌫`      | Go to parent dir    |
| `Tab`          | Switch active pane  |
| `q` / `Ctrl+C` | Quit                |

## Built with

[Bubble Tea](https://github.com/charmbracelet/bubbletea) ·
[Lip Gloss](https://github.com/charmbracelet/lipgloss) ·
[Bubbles](https://github.com/charmbracelet/bubbles)

## License

[MIT](LICENSE)

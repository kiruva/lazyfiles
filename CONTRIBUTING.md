# Contributing to lazyfiles

Thanks for your interest! lazyfiles aims to be a clean, keyboard-driven, dual-pane
file manager. Contributions of all sizes are welcome.

## Getting started

Go 1.26 or newer is required.

```sh
git clone https://github.com/kiruva/lazyfiles
cd lazyfiles
go run .
```

## Before opening a PR

Please make sure the following pass locally:

```sh
make fmt    # gofmt
make vet    # go vet ./...
make test   # go test ./...
make lint   # golangci-lint run (see .golangci.yml)
```

CI runs the same checks on every push and pull request.

## Project layout

| Path               | Responsibility                                           |
| ------------------ | -------------------------------------------------------- |
| `main.go`          | Entry point; starts the Bubble Tea program.              |
| `internal/app`     | Root model, update router, view, keymap (The Elm Arch.). |
| `internal/pane`    | One directory pane; also the virtual archive browser.    |
| `internal/fileops` | UI-agnostic create/copy/move/delete + archive engine.    |
| `internal/remote`  | In-process ssh: listings, transfers, `~/.ssh/config`.    |
| `internal/config`  | The `key = value` config file: theme, saved connections. |
| `internal/ui`      | Lip Gloss theme and shared styles.                       |

## Guidelines

- Keep the `fileops` package free of any Bubble Tea / UI imports — it streams
  progress over a channel that the app layer adapts into messages.
- Filesystem operations must stay off the UI thread (run as a `tea.Cmd`).
- Match the surrounding style: small functions, clear names, comments where the
  _why_ isn't obvious.
- New keybindings go in `internal/app/keys.go` and should appear in the `?` help
  overlay automatically via `keyMap.groups()`.
- Anything remote must stay off the UI thread too: the pane records where it wants
  to be and the app layer fills it in from a `tea.Cmd`.
- Note user-facing changes in `CHANGELOG.md` under `## [Unreleased]`.

## Reporting bugs

Open an issue with your OS, terminal, lazyfiles version (`lazyfiles --version`),
and steps to reproduce.

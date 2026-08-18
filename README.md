# lazyfiles

> A TUI file manager to rule them all — dual-pane, keyboard-driven, as intuitive as lazygit.

[![CI](https://github.com/kiruva/lazyfiles/actions/workflows/ci.yml/badge.svg)](https://github.com/kiruva/lazyfiles/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/kiruva/lazyfiles)](https://goreportcard.com/report/github.com/kiruva/lazyfiles)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Built for the Linux community first, comfortable on macOS. Leans on what Total Commander is
to Windows, reimagined for the terminal.

> **Status:** early days — dual-pane navigation, recursive copy/move/delete, pack/unpack
> archives, view/edit text (including files **inside** archives), and browsing/transferring
> over ssh. See [ROADMAP.md](ROADMAP.md).

![demo](demo.gif)

## Install

```sh
# With Go installed (1.24+):
go install github.com/kiruva/lazyfiles@latest

# Or build from source:
git clone https://github.com/kiruva/lazyfiles
cd lazyfiles
make build      # produces ./lazyfiles
```

Pre-built binaries for Linux and macOS (amd64/arm64) are attached to each
[release](https://github.com/kiruva/lazyfiles/releases).

## Run it

```sh
lazyfiles                  # or: go run .
lazyfiles --theme nord     # start with a theme
lazyfiles --themes         # list the built-in themes
lazyfiles --version
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
| `Ctrl+L` / `:`     | Edit address bar (jump to path) |
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
| `Enter`            | Open dir / browse into archive  |
| `v`                | View file (read-only)           |
| `e`                | Edit file (nano-style)          |
| `S`                | ssh connections                 |
| `t`                | Theme picker                    |
| `?`                | Show all keybindings            |
| `y` / `n`          | Confirm / cancel a prompt       |
| `q` / `Ctrl+C`     | Quit                            |

Press `?` any time for the full keybinding overlay.

## Address bar

Each pane's top line is its address bar. It follows the cursor as you walk the tree, and
`Ctrl+L` (or `:`) turns it into an input for jumping straight to a path — `Enter` goes,
`Esc` cancels, `Tab` completes, `↑`/`↓` cycle completions. Paths may be absolute, relative
to the current directory, `~`-rooted, or contain `$VARS`. Typing the path of a browsable
archive opens it as a virtual tree, and an `ssh://` or `host:/path` target starts the ssh
connection flow for that pane (see below); anything that isn't a directory leaves the bar
open with the reason in the status line.

Operations act on the marked selection (`Space`), or on the highlighted entry when nothing
is marked. Copy/move/pack/unpack always go from the **active** pane to the **other** pane.

Archive actions shell out to `tar`, `unzip`, `7z`, and `unrar` (only the tool for the format
you touch is required). `pack` writes a `.tar.gz`.

## Over ssh

Press `S` for the connection modal. It lists your saved connections, most recently
used first, with a `●` next to any that are already connected:

```
╭──────────────────────────────────────────────────────────╮
│   Connect over ssh                           left pane   │
│                                                          │
│   ▸ prod         deploy@web01.example.com /srv/www       │
│     backup       ● kim@nas:2222                          │
│     + new connection…                                    │
│                                                          │
│   enter connect · n new · e edit · d delete · esc        │
╰──────────────────────────────────────────────────────────╯
```

`enter` connects, `n` adds one, `e` edits, `d` deletes. The connection opens in the
pane you pressed `S` from — the modal says which one.

A connection records a name, host, user, port, starting path, and optionally a key
file. **It never records a password.** If the host needs one, lazyfiles asks each
time it starts:

```
╭──────────────────────────────────────────────────────────╮
│   Password                                   left pane   │
│                                                          │
│   for prod (deploy@web01.example.com)                    │
│                                                          │
│   ••••••••••                                             │
│                                                          │
│   not saved — asked again next time lazyfiles starts     │
╰──────────────────────────────────────────────────────────╯
```

The password is offered to the server and kept in memory for that session, so
transfers and reconnects after an idle timeout do not ask again. Quitting drops it.
Authentication is tried in order: ssh-agent, then key files, then the password —
so a key-based host never prompts at all.

A host that isn't in `~/.ssh/known_hosts` shows its fingerprint for you to check
before anything is sent to it; accepting appends it to `known_hosts`, as `ssh`
would. A known host presenting a *different key of the same type* is refused
outright rather than offered as a yes/no — that is either a rebuilt server or an
interception, and either way it wants looking at by hand. A key of an algorithm
you have no entry for is just a key you have not seen, and prompts normally.

As `ssh` does, lazyfiles asks the server for a host key algorithm it already has
recorded, so a host whose `known_hosts` line is ed25519 is not re-verified against
whatever key type the server happens to prefer.

Saved connections live in the same config file as everything else:

```ini
# lazyfiles configuration
# ssh passwords are never stored here
conn.prod.host = web01.example.com
conn.prod.user = deploy
conn.prod.path = /srv/www
```

### Transferring

With one pane remote and one local, `c`/`F5` and `m`/`F6` transfer in whichever
direction the panes describe — the **active** pane is always the source, so copying
down and copying up are the same two keys. `d`/`F8` deletes on the host. With both
panes on the same host, copy and move run there without the data crossing the wire.

The address bar (`Ctrl+L`) also accepts a remote target, which goes through the same
connect flow:

```
ssh://user@host:2222/var/log
user@host:/var/log            # scp-style
host:                         # the login directory
```

While a pane is remote, plain paths in the address bar stay on that host; `local:/path`
(or `file:///path`) brings it back to this machine. `Enter`/`h` walk the tree, `Space`
marks, `s` sorts, `.` toggles hidden files.

Not available over ssh: pack/unpack, browsing into archives, and view/edit — copy the
file across first. Copying directly between two different hosts is refused; route it
through this machine.

### How it works

lazyfiles speaks ssh in-process (`golang.org/x/crypto/ssh`) rather than shelling out
to the `ssh` binary. That is what makes the password prompt possible at all: `ssh`
reads passwords straight from the terminal, which the TUI owns, and a password passed
any other way would have to travel through a command line or an environment variable
where other processes can see it. In-process, it goes from the input straight into the
authentication exchange.

The trade-off is that a native client does not read `~/.ssh/config` for you, so
lazyfiles parses the part that decides where a connection goes: `HostName`, `User`,
`Port`, `IdentityFile`, `ProxyJump`, plus `Include` and `Host` pattern matching. An
alias that depends on anything else — `ProxyCommand`, for instance — will not resolve
the way `ssh` would. A jump host must accept key or agent authentication, since the
modal only prompts for one password.

Transfers stream a tar archive over one ssh session (`tar -cf -` on one end, `tar -xf -`
on the other) rather than using scp or sftp. Every path is quoted by lazyfiles and
interpreted by exactly one shell, and the local `tar -v` names each file as it moves,
which is what fills the progress bar. The far side needs `tar` and a POSIX shell;
nothing is installed.

## Themes

Eight built-in themes: `default`, `nord`, `dracula`, `gruvbox`, `catppuccin`, `tokyonight`,
`solarized`, `monokai`. Press `t` for the picker — moving the cursor previews the theme live,
`Enter` applies and remembers it, `Esc` puts the old one back.

Resolution order is `--theme` → `$LAZYFILES_THEME` → config file → `default`:

```sh
lazyfiles --theme gruvbox
LAZYFILES_THEME=dracula lazyfiles
```

The picker writes the choice to `$XDG_CONFIG_HOME/lazyfiles/config` (or
`~/.config/lazyfiles/config`), a plain `key = value` file you can also edit by hand:

```ini
# lazyfiles configuration
theme = nord
```

Saved ssh connections share this file; see [Over ssh](#over-ssh).

Themes are pure data — a name plus eight colours in `internal/ui/theme.go`. Adding one is a
single struct literal; every style is rebuilt from it.

## View & edit

`v` opens a file in a scrollable read-only viewer; `e` opens it in a nano-style editor
(`Ctrl+S` save, `Ctrl+Q`/`Esc` quit — `Esc` guards unsaved changes). Binary files are refused.

This works **inside archives** too: press `Enter` on a `.tar`/`.tar.gz`/`.zip` to browse its
contents as a virtual folder, then `v`/`e` on any member streams it into memory. Saving an
edited member writes it back to the archive (targeted update for zip / uncompressed tar;
transparent repack for compressed tar; `.7z`/`.rar` browsing requires unpacking).

**Copying into archives**: with one pane browsing inside an archive and the other on real
files, `F5`/`c` (copy) or `F6`/`m` (move) adds the selected files/folders as members at the
archive's current virtual directory. `pack`/`unpack` still require a real destination pane.

## Development

```sh
make run     # go run .
make test    # go test ./...
make vet     # go vet ./...
make lint    # golangci-lint run
make build   # build ./lazyfiles
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for the project layout and guidelines.

To regenerate the demo GIF, install [VHS](https://github.com/charmbracelet/vhs) and run
`vhs demo.tape`.

## Built with

[Bubble Tea](https://github.com/charmbracelet/bubbletea) ·
[Lip Gloss](https://github.com/charmbracelet/lipgloss) ·
[Bubbles](https://github.com/charmbracelet/bubbles)

## License

[MIT](LICENSE)

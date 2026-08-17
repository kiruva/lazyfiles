package app

import (
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/kiruva/lazyfiles/internal/fileops"
)

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resizePanes()
		return m, nil

	case fileops.Progress:
		m.progress = msg
		return m, waitCmd(m.progressCh)

	case fileops.Result:
		return m.finishOp(msg), nil

	case tea.KeyMsg:
		switch m.mode {
		case modeConfirm:
			return m.onConfirmKey(msg)
		case modeProgress:
			return m, nil // input is locked while an operation runs
		default:
			return m.onNormalKey(msg)
		}
	}
	return m, nil
}

// onNormalKey handles navigation and triggers operations.
func (m Model) onNormalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.errText = "" // any key clears a stale error banner
	p := &m.panes[m.active]

	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit
	case key.Matches(msg, m.keys.Up):
		p.MoveUp()
	case key.Matches(msg, m.keys.Down):
		p.MoveDown()
	case key.Matches(msg, m.keys.PageUp):
		p.PageUp()
	case key.Matches(msg, m.keys.PageDown):
		p.PageDown()
	case key.Matches(msg, m.keys.Top):
		p.Top()
	case key.Matches(msg, m.keys.Bottom):
		p.Bottom()
	case key.Matches(msg, m.keys.Enter):
		p.Enter()
	case key.Matches(msg, m.keys.Back):
		p.Ascend()
	case key.Matches(msg, m.keys.Select):
		p.ToggleSelect()
	case key.Matches(msg, m.keys.Hidden):
		p.ToggleHidden()
	case key.Matches(msg, m.keys.Sort):
		p.CycleSort()
	case key.Matches(msg, m.keys.Switch):
		m.active = 1 - m.active
	case key.Matches(msg, m.keys.Copy):
		m.beginOp(fileops.OpCopy)
	case key.Matches(msg, m.keys.Move):
		m.beginOp(fileops.OpMove)
	case key.Matches(msg, m.keys.Delete):
		m.beginOp(fileops.OpDelete)
	case key.Matches(msg, m.keys.Pack):
		m.beginOp(fileops.OpPack)
	case key.Matches(msg, m.keys.Unpack):
		m.beginOp(fileops.OpUnpack)
	case key.Matches(msg, m.keys.Unwrap):
		m.beginOp(fileops.OpUnwrap)
	}
	return m, nil
}

// onConfirmKey handles the y/n prompt for a pending operation.
func (m Model) onConfirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "enter":
		return m, m.startPending()
	case "n", "esc", "q":
		m.mode = modeNormal
	}
	return m, nil
}

// beginOp assembles a job from the active pane's selection and asks for
// confirmation. Source = active pane; destination = the other pane (except
// delete, which has none, and unwrap, which extracts in place).
func (m *Model) beginOp(op fileops.Op) {
	src := &m.panes[m.active]
	names := src.SelectedNames()
	if len(names) == 0 {
		return
	}

	srcs := make([]string, 0, len(names))
	for _, n := range names {
		srcs = append(srcs, filepath.Join(src.Path, n))
	}
	other := m.panes[1-m.active].Path

	m.willOverwrite = false
	job := fileops.Job{Op: op, Srcs: srcs}

	switch op {
	case fileops.OpCopy, fileops.OpMove:
		if other == src.Path {
			m.errText = "source and destination are the same directory"
			return
		}
		job.Dest = other
		m.willOverwrite = anyExist(other, names)

	case fileops.OpDelete:
		// destructive; no destination

	case fileops.OpUnpack, fileops.OpUnwrap:
		if bad := firstNonArchive(names); bad != "" {
			m.errText = "not a supported archive: " + bad
			return
		}
		if op == fileops.OpUnpack {
			if other == src.Path {
				m.errText = "source and destination are the same directory"
				return
			}
			job.Dest = other
		} else {
			job.Dest = src.Path // unwrap = extract in place
		}

	case fileops.OpPack:
		if other == src.Path {
			m.errText = "source and destination are the same directory"
			return
		}
		job.Dest = other
		job.Out = filepath.Join(other, packName(names))
		if _, err := os.Stat(job.Out); err == nil {
			m.willOverwrite = true
		}
	}

	m.pending = job
	m.mode = modeConfirm
}

// anyExist reports whether any of names already exists under dir.
func anyExist(dir string, names []string) bool {
	for _, n := range names {
		if _, err := os.Stat(filepath.Join(dir, n)); err == nil {
			return true
		}
	}
	return false
}

// firstNonArchive returns the first name that isn't a supported archive, or "".
func firstNonArchive(names []string) string {
	for _, n := range names {
		if !fileops.IsArchive(n) {
			return n
		}
	}
	return ""
}

// packName picks the output archive name for a pack job.
func packName(names []string) string {
	if len(names) == 1 {
		return names[0] + ".tar.gz"
	}
	return "archive.tar.gz"
}

// startPending launches the confirmed job and enters progress mode.
func (m *Model) startPending() tea.Cmd {
	ch := fileops.Run(m.pending)
	m.progressCh = ch
	m.progress = fileops.Progress{}
	m.mode = modeProgress
	return waitCmd(ch)
}

// finishOp refreshes both panes and surfaces any error.
func (m Model) finishOp(res fileops.Result) Model {
	m.mode = modeNormal
	m.progressCh = nil
	for i := range m.panes {
		m.panes[i].ClearSelection()
		m.panes[i].Refresh()
	}
	if res.Err != nil {
		m.errText = res.Op.String() + " failed: " + res.Err.Error()
	}
	return m
}

// resizePanes splits the terminal into two side-by-side panes above the status bar.
func (m *Model) resizePanes() {
	bodyH := m.height - 1 // reserve one line for the status bar
	leftW := m.width / 2
	m.panes[0].SetSize(leftW, bodyH)
	m.panes[1].SetSize(m.width-leftW, bodyH)
}

// waitCmd blocks on the next value from the operation channel and delivers it
// as a message; it returns nil once the channel is closed.
func waitCmd(ch <-chan any) tea.Cmd {
	return func() tea.Msg {
		if ch == nil {
			return nil
		}
		msg, ok := <-ch
		if !ok {
			return nil
		}
		return msg
	}
}

package app

import (
	"bytes"
	"fmt"
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
		return m.onKey(msg)
	}

	// Non-key messages (e.g. cursor blink) are routed to the active component.
	switch m.mode {
	case modeEdit:
		var cmd tea.Cmd
		m.editor, cmd = m.editor.Update(msg)
		return m, cmd
	case modeView:
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}
	return m, nil
}

// onKey dispatches a keypress to the handler for the current mode.
func (m Model) onKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case modeConfirm:
		return m.onConfirmKey(msg)
	case modeProgress:
		return m, nil // input is locked while an operation runs
	case modeView:
		return m.onViewKey(msg)
	case modeEdit:
		return m.onEditKey(msg)
	case modeHelp:
		m.mode = modeNormal // any key dismisses the help overlay
		return m, nil
	default:
		return m.onNormalKey(msg)
	}
}

// onNormalKey handles navigation and triggers operations / view / edit.
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
		m.enterOrOpen()
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
	case key.Matches(msg, m.keys.Help):
		m.mode = modeHelp
	case key.Matches(msg, m.keys.View):
		return m, m.openViewer()
	case key.Matches(msg, m.keys.Edit):
		return m, m.openEditor()
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

// onViewKey drives the read-only pager.
func (m Model) onViewKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.mode = modeNormal
		return m, nil
	case "e":
		return m, m.openEditor() // hand off to the editor on the same file
	}
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// onEditKey drives the nano-style editor.
func (m Model) onEditKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+s":
		m.saveEditor()
		return m, nil
	case "ctrl+q":
		m.closeEditor()
		return m, nil
	case "esc":
		if m.editor.Value() != m.editOrig {
			m.editStatus = "unsaved — Ctrl+S to save, Ctrl+Q to discard"
			return m, nil
		}
		m.closeEditor()
		return m, nil
	}
	m.editStatus = ""
	var cmd tea.Cmd
	m.editor, cmd = m.editor.Update(msg)
	return m, cmd
}

// enterOrOpen descends into a directory, or opens an archive for browsing.
func (m *Model) enterOrOpen() {
	p := &m.panes[m.active]
	if !p.InArchive() {
		if cur, ok := p.Current(); ok && !cur.IsDir && fileops.Browsable(cur.Name) {
			if err := p.EnterArchive(filepath.Join(p.Path, cur.Name)); err != nil {
				m.errText = "cannot open archive: " + err.Error()
			}
			return
		}
	}
	p.Enter()
}

// loadCurrent reads the highlighted file (real or in-archive) into memory.
func (m *Model) loadCurrent() (data []byte, title string, err error) {
	p := &m.panes[m.active]
	if p.InArchive() {
		mem, ok := p.CurrentMemberPath()
		if !ok {
			return nil, "", fmt.Errorf("select a file to open")
		}
		data, err = fileops.ReadMember(p.ArchivePath(), mem)
		return data, mem, err
	}
	cur, ok := p.Current()
	if !ok || cur.IsDir {
		return nil, "", fmt.Errorf("select a file to open")
	}
	full := filepath.Join(p.Path, cur.Name)
	data, err = os.ReadFile(full)
	return data, cur.Name, err
}

// openViewer loads the current file into the read-only pager.
func (m *Model) openViewer() tea.Cmd {
	data, title, err := m.loadCurrent()
	if err != nil {
		m.errText = err.Error()
		return nil
	}
	if isBinary(data) {
		m.errText = "not a text file: " + title
		return nil
	}
	m.viewTitle = title
	m.viewport.SetContent(string(data))
	m.viewport.GotoTop()
	m.mode = modeView
	return nil
}

// openEditor loads the current file into the editor and remembers where to save.
func (m *Model) openEditor() tea.Cmd {
	data, title, err := m.loadCurrent()
	if err != nil {
		m.errText = err.Error()
		return nil
	}
	if isBinary(data) {
		m.errText = "not a text file: " + title
		return nil
	}

	p := &m.panes[m.active]
	m.edit = editTarget{title: title}
	if p.InArchive() {
		mem, _ := p.CurrentMemberPath()
		m.edit.archive = p.ArchivePath()
		m.edit.member = mem
	} else {
		cur, _ := p.Current()
		m.edit.realPath = filepath.Join(p.Path, cur.Name)
	}

	m.editOrig = string(data)
	m.editStatus = ""
	m.editor.SetValue(string(data))
	m.mode = modeEdit
	return m.editor.Focus()
}

// saveEditor writes the buffer back to its origin (file or archive member).
func (m *Model) saveEditor() {
	data := []byte(m.editor.Value())

	var err error
	if m.edit.archive != "" {
		err = fileops.WriteMember(m.edit.archive, m.edit.member, data)
	} else {
		perm := os.FileMode(0o644)
		if fi, e := os.Stat(m.edit.realPath); e == nil {
			perm = fi.Mode().Perm()
		}
		err = os.WriteFile(m.edit.realPath, data, perm)
	}
	if err != nil {
		m.errText = "save failed: " + err.Error()
		m.editStatus = "save failed"
		return
	}

	m.editOrig = string(data) // buffer is now clean
	m.editStatus = "saved"
	m.panes[m.active].Refresh()
}

func (m *Model) closeEditor() {
	m.mode = modeNormal
	m.editor.Blur()
	m.editStatus = ""
}

// beginOp assembles a job from the active pane's selection and asks for
// confirmation. Source = active pane; destination = the other pane (except
// delete, which has none, and unwrap, which extracts in place).
func (m *Model) beginOp(op fileops.Op) {
	src := &m.panes[m.active]
	if src.InArchive() {
		m.errText = "not available inside an archive — unpack it first"
		return
	}
	// Pack/unpack need a real destination directory; copy/move can target an
	// archive (handled below as add-to-archive).
	if (op == fileops.OpPack || op == fileops.OpUnpack) && m.panes[1-m.active].InArchive() {
		m.errText = "destination pane is inside an archive"
		return
	}
	names := src.SelectedNames()
	if len(names) == 0 {
		return
	}

	srcs := make([]string, 0, len(names))
	for _, n := range names {
		srcs = append(srcs, filepath.Join(src.Path, n))
	}
	dst := &m.panes[1-m.active]
	other := dst.Path

	m.willOverwrite = false
	job := fileops.Job{Op: op, Srcs: srcs}

	switch op {
	case fileops.OpCopy, fileops.OpMove:
		if dst.InArchive() {
			// Copy/move real files into the archive at its current virtual dir.
			job = fileops.Job{
				Op:   fileops.OpAddToArchive,
				Srcs: srcs,
				Dest: dst.ArchivePath(),
				VDir: dst.VPath(),
				Move: op == fileops.OpMove,
			}
			break
		}
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

// resizePanes splits the terminal into two panes above the status bar and sizes
// the view/edit components to the full body area.
func (m *Model) resizePanes() {
	bodyH := m.height - 1 // reserve one line for the status bar
	leftW := m.width / 2
	m.panes[0].SetSize(leftW, bodyH)
	m.panes[1].SetSize(m.width-leftW, bodyH)

	compH := m.height - 2 // header + footer
	if compH < 1 {
		compH = 1
	}
	m.viewport.Width = m.width
	m.viewport.Height = compH
	m.editor.SetWidth(m.width)
	m.editor.SetHeight(compH)
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

// isBinary reports whether data looks non-textual (contains a NUL byte).
func isBinary(data []byte) bool {
	n := len(data)
	if n > 8000 {
		n = 8000
	}
	return bytes.IndexByte(data[:n], 0) >= 0
}

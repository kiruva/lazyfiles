// Package fileops performs recursive copy/move/delete off the UI thread and
// streams progress over a channel. It has no knowledge of Bubble Tea — the app
// layer adapts the emitted values into messages.
package fileops

import (
	"errors"
	"io"
	iofs "io/fs"
	"os"
	"path/filepath"
	"syscall"
)

// Op identifies a filesystem operation.
type Op int

const (
	OpCopy Op = iota
	OpMove
	OpDelete
)

func (o Op) String() string {
	switch o {
	case OpCopy:
		return "Copy"
	case OpMove:
		return "Move"
	case OpDelete:
		return "Delete"
	default:
		return "?"
	}
}

// Present returns the present-continuous form for progress titles.
func (o Op) Present() string {
	switch o {
	case OpCopy:
		return "Copying"
	case OpMove:
		return "Moving"
	case OpDelete:
		return "Deleting"
	default:
		return "Working"
	}
}

// Job describes work to perform. Dest is the destination directory and is
// ignored for OpDelete.
type Job struct {
	Op   Op
	Srcs []string // absolute source paths
	Dest string   // destination directory
}

// Progress is emitted repeatedly as the job runs.
type Progress struct {
	Current     string
	Done, Total int
}

// Result is emitted exactly once, last, when the job finishes.
type Result struct {
	Op  Op
	Err error
}

// Run starts the job in a goroutine and returns a channel that yields zero or
// more Progress values followed by a single Result, then closes.
func Run(job Job) <-chan any {
	ch := make(chan any)
	go func() {
		defer close(ch)

		total := 0
		for _, s := range job.Srcs {
			total += countItems(s)
		}
		r := &reporter{ch: ch, total: total}

		var err error
		for _, src := range job.Srcs {
			switch job.Op {
			case OpCopy:
				err = copyPath(src, filepath.Join(job.Dest, filepath.Base(src)), r)
			case OpMove:
				err = movePath(src, filepath.Join(job.Dest, filepath.Base(src)), r)
			case OpDelete:
				err = deletePath(src, r)
			}
			if err != nil {
				break
			}
		}
		ch <- Result{Op: job.Op, Err: err}
	}()
	return ch
}

// reporter emits one Progress per processed item.
type reporter struct {
	ch          chan<- any
	done, total int
}

func (r *reporter) step(name string) {
	r.done++
	r.ch <- Progress{Current: name, Done: r.done, Total: r.total}
}

func countItems(root string) int {
	n := 0
	_ = filepath.WalkDir(root, func(_ string, _ iofs.DirEntry, err error) error {
		if err == nil {
			n++
		}
		return nil
	})
	if n == 0 {
		return 1
	}
	return n
}

func copyPath(src, dst string, r *reporter) error {
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}

	switch {
	case info.Mode()&os.ModeSymlink != 0:
		target, err := os.Readlink(src)
		if err != nil {
			return err
		}
		_ = os.Remove(dst)
		if err := os.Symlink(target, dst); err != nil {
			return err
		}
		r.step(src)
		return nil

	case info.IsDir():
		if err := os.MkdirAll(dst, info.Mode().Perm()); err != nil {
			return err
		}
		r.step(src)
		entries, err := os.ReadDir(src)
		if err != nil {
			return err
		}
		for _, e := range entries {
			if err := copyPath(filepath.Join(src, e.Name()), filepath.Join(dst, e.Name()), r); err != nil {
				return err
			}
		}
		return nil

	default:
		if err := copyFile(src, dst, info.Mode().Perm()); err != nil {
			return err
		}
		r.step(src)
		return nil
	}
}

func copyFile(src, dst string, perm os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

func movePath(src, dst string, r *reporter) error {
	// Fast path: same filesystem → atomic rename.
	if err := os.Rename(src, dst); err == nil {
		r.step(src)
		return nil
	} else if !errors.Is(err, syscall.EXDEV) {
		return err
	}
	// Cross-device: copy then remove the original.
	if err := copyPath(src, dst, r); err != nil {
		return err
	}
	return os.RemoveAll(src)
}

func deletePath(src string, r *reporter) error {
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}
	if info.IsDir() && info.Mode()&os.ModeSymlink == 0 {
		entries, err := os.ReadDir(src)
		if err != nil {
			return err
		}
		for _, e := range entries {
			if err := deletePath(filepath.Join(src, e.Name()), r); err != nil {
				return err
			}
		}
	}
	if err := os.Remove(src); err != nil {
		return err
	}
	r.step(src)
	return nil
}

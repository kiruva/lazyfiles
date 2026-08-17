package pane

import (
	"os"
	"path/filepath"
	"sort"
)

// Entry is a single item shown in a pane.
type Entry struct {
	Name  string
	IsDir bool
	Size  int64
	Mode  os.FileMode
}

// ReadDir reads path and returns its entries sorted directories-first, then by
// name. A ".." entry is prepended unless path is a filesystem root.
func ReadDir(path string) ([]Entry, error) {
	dirents, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	entries := make([]Entry, 0, len(dirents)+1)
	if parent := filepath.Dir(path); parent != path {
		entries = append(entries, Entry{Name: "..", IsDir: true})
	}

	for _, d := range dirents {
		e := Entry{Name: d.Name(), IsDir: d.IsDir()}
		if info, err := d.Info(); err == nil {
			e.Size = info.Size()
			e.Mode = info.Mode()
		}
		entries = append(entries, e)
	}

	sort.SliceStable(entries, func(i, j int) bool {
		a, b := entries[i], entries[j]
		if a.Name == ".." {
			return true
		}
		if b.Name == ".." {
			return false
		}
		if a.IsDir != b.IsDir {
			return a.IsDir
		}
		return a.Name < b.Name
	})

	return entries, nil
}

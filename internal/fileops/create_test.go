package fileops

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateFile(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "notes.txt")

	if err := CreateFile(target); err != nil {
		t.Fatalf("CreateFile: %v", err)
	}
	fi, err := os.Stat(target)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if fi.Size() != 0 {
		t.Errorf("new file is %d bytes, want empty", fi.Size())
	}
}

// A create must never truncate what is already there — that is what the
// copy/move overwrite prompt is for.
func TestCreateFileRefusesExisting(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "keep.txt")
	if err := os.WriteFile(target, []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := CreateFile(target); err == nil {
		t.Fatal("CreateFile overwrote an existing file")
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "data" {
		t.Errorf("contents = %q, want %q", data, "data")
	}
}

func TestCreateFileMakesParents(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "a", "b", "main.go")

	if err := CreateFile(target); err != nil {
		t.Fatalf("CreateFile: %v", err)
	}
	if _, err := os.Stat(target); err != nil {
		t.Fatalf("stat: %v", err)
	}
}

func TestCreateDir(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "a", "b")

	if err := CreateDir(target); err != nil {
		t.Fatalf("CreateDir: %v", err)
	}
	fi, err := os.Stat(target)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if !fi.IsDir() {
		t.Error("created a file, want a directory")
	}
}

func TestCreateDirRefusesExisting(t *testing.T) {
	dir := t.TempDir()
	if err := CreateDir(dir); err == nil {
		t.Fatal("CreateDir accepted an existing directory")
	}

	file := filepath.Join(dir, "taken")
	if err := os.WriteFile(file, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := CreateDir(file); err == nil {
		t.Fatal("CreateDir accepted a name taken by a file")
	}
}

package browser

import (
	"os"
	"path/filepath"
	"testing"
)

func setup(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"b.mp4", "A.mkv", "notes.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestModelSortsDirsFirst(t *testing.T) {
	dir := setup(t)
	m, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Entries) != 4 {
		t.Fatalf("entries = %d", len(m.Entries))
	}
	if m.Entries[0].Name != "sub" || !m.Entries[0].Dir {
		t.Fatalf("first entry = %+v", m.Entries[0])
	}
	if m.Entries[1].Name != "A.mkv" {
		t.Fatalf("second entry = %+v", m.Entries[1])
	}
}

func TestMoveAndSelect(t *testing.T) {
	m, err := New(setup(t))
	if err != nil {
		t.Fatal(err)
	}
	m.Move(1)
	if e, _ := m.Selected(); e.Name != "A.mkv" {
		t.Fatalf("selected = %+v", e)
	}
	m.Move(-10)
	if m.Cursor != 0 {
		t.Fatalf("cursor = %d", m.Cursor)
	}
	m.Move(99)
	if m.Cursor != len(m.Entries)-1 {
		t.Fatalf("cursor = %d", m.Cursor)
	}
}

func TestEnterFileAndDirectory(t *testing.T) {
	dir := setup(t)
	m, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := m.Enter(); ok {
		t.Fatal("entering a directory should not return a file")
	}
	if m.Dir != filepath.Join(dir, "sub") {
		t.Fatalf("dir = %q", m.Dir)
	}
	m.Parent()
	if m.Dir != dir {
		t.Fatalf("parent = %q", m.Dir)
	}
	m.Move(1)
	file, ok := m.Enter()
	if !ok || file != filepath.Join(dir, "A.mkv") {
		t.Fatalf("file = %q, ok = %v", file, ok)
	}
}

func TestEmptyDirectory(t *testing.T) {
	m, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := m.Selected(); ok {
		t.Fatal("empty directory has a selection")
	}
	if file, ok := m.Enter(); ok || file != "" {
		t.Fatal("empty directory returned a file")
	}
}

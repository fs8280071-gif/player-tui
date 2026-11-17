// Package browser implements a small, testable file-list model.
package browser

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Entry struct {
	Name string
	Path string
	Dir  bool
}

type Model struct {
	Dir     string
	Entries []Entry
	Cursor  int
	Err     error
}

func New(dir string) (*Model, error) {
	if dir == "" {
		dir = "."
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	m := &Model{Dir: abs}
	if err := m.Refresh(); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *Model) Refresh() error {
	list, err := os.ReadDir(m.Dir)
	if err != nil {
		m.Err = err
		return err
	}
	entries := make([]Entry, 0, len(list))
	for _, de := range list {
		name := de.Name()
		path := filepath.Join(m.Dir, name)
		isDir := de.IsDir()
		if de.Type()&os.ModeSymlink != 0 {
			if info, err := os.Stat(path); err == nil {
				isDir = info.IsDir()
			}
		}
		entries = append(entries, Entry{Name: name, Path: path, Dir: isDir})
	}
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].Dir != entries[j].Dir {
			return entries[i].Dir
		}
		return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
	})
	m.Entries = entries
	m.Err = nil
	if m.Cursor >= len(entries) {
		m.Cursor = len(entries) - 1
	}
	if m.Cursor < 0 {
		m.Cursor = 0
	}
	return nil
}

func (m *Model) Move(delta int) {
	if len(m.Entries) == 0 {
		m.Cursor = 0
		return
	}
	m.Cursor += delta
	if m.Cursor < 0 {
		m.Cursor = 0
	}
	if m.Cursor >= len(m.Entries) {
		m.Cursor = len(m.Entries) - 1
	}
}

func (m *Model) Selected() (Entry, bool) {
	if m.Cursor < 0 || m.Cursor >= len(m.Entries) {
		return Entry{}, false
	}
	return m.Entries[m.Cursor], true
}

// Enter walks into the selected directory. It returns ok=true with a file path
// when the selection is a file.
func (m *Model) Enter() (string, bool) {
	entry, ok := m.Selected()
	if !ok {
		return "", false
	}
	if !entry.Dir {
		return entry.Path, true
	}
	next := &Model{Dir: entry.Path}
	if err := next.Refresh(); err != nil {
		m.Err = err
		return "", false
	}
	*m = *next
	return "", false
}

func (m *Model) Parent() {
	parent := filepath.Dir(m.Dir)
	if parent == m.Dir {
		return
	}
	old := m.Dir
	next := &Model{Dir: parent}
	if err := next.Refresh(); err != nil {
		m.Err = err
		return
	}
	next.selectPath(old)
	*m = *next
}

func (m *Model) selectPath(path string) {
	for i, e := range m.Entries {
		if e.Path == path {
			m.Cursor = i
			return
		}
	}
}

var ErrNotDir = errors.New("not a directory")

func (m *Model) Up() {
	m.Parent()
}

package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

type ResumeEntry struct {
	Position time.Duration `json:"position_ns"`
	Updated  time.Time     `json:"updated"`
}

type Store struct {
	path    string
	Entries map[string]ResumeEntry `json:"entries"`
}

func NewStore(path string) *Store {
	s := &Store{path: path, Entries: map[string]ResumeEntry{}}
	data, err := os.ReadFile(path)
	if err == nil {
		var loaded Store
		if json.Unmarshal(data, &loaded) == nil && loaded.Entries != nil {
			s.Entries = loaded.Entries
		}
	}
	return s
}

func (s *Store) Position(target string) time.Duration {
	key, ok := keyFor(target)
	if !ok {
		return 0
	}
	return s.Entries[key].Position
}

func (s *Store) Set(target string, position time.Duration) {
	key, ok := keyFor(target)
	if !ok || position <= 0 {
		return
	}
	s.Entries[key] = ResumeEntry{Position: position, Updated: time.Now()}
}

func (s *Store) Forget(target string) {
	if key, ok := keyFor(target); ok {
		delete(s.Entries, key)
	}
}

func (s *Store) Save() error {
	if s.path == "" {
		return errors.New("resume store has no path")
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func keyFor(target string) (string, bool) {
	if target == "" {
		return "", false
	}
	lower := target
	for _, scheme := range []string{"http://", "https://", "rtsp://", "rtmp://", "srt://", "file://"} {
		if len(lower) >= len(scheme) && lower[:len(scheme)] == scheme {
			return "", false
		}
	}
	abs, err := filepath.Abs(target)
	if err != nil {
		return "", false
	}
	return abs, true
}

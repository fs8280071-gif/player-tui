package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadDefaultsForMissingFile(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "missing.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Render != "auto" || cfg.Theme != "catppuccin" || cfg.Volume != 100 || !cfg.Resume {
		t.Fatalf("defaults = %+v", cfg)
	}
}

func TestLoadValuesAndComments(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	data := `
# player-tui settings
audio = true
render = "halfblock"   # inline comment
theme = 'gruvbox'
volume = 65
resume = false
dither = true
unknown = "ignored"
`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Audio || cfg.Render != "halfblock" || cfg.Theme != "gruvbox" ||
		cfg.Volume != 65 || cfg.Resume || !cfg.Dither {
		t.Fatalf("config = %+v", cfg)
	}
}

func TestLoadRejectsBadValues(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	for _, line := range []string{
		"volume = 999",
		"resume = maybe",
		"render = halfblock",
		"bogus",
	} {
		if err := os.WriteFile(path, []byte(line+"\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := Load(path); err == nil {
			t.Errorf("accepted %q", line)
		}
	}
}

func TestResumeStoreRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "resume.json")
	media := filepath.Join(dir, "clip.mp4")
	store := NewStore(path)
	store.Set(media, 42*time.Second)
	if err := store.Save(); err != nil {
		t.Fatal(err)
	}
	reloaded := NewStore(path)
	if got := reloaded.Position(media); got != 42*time.Second {
		t.Fatalf("position = %v", got)
	}
	reloaded.Forget(media)
	if err := reloaded.Save(); err != nil {
		t.Fatal(err)
	}
	if got := NewStore(path).Position(media); got != 0 {
		t.Fatalf("forgotten position = %v", got)
	}
}

func TestResumeStoreSkipsURLs(t *testing.T) {
	store := NewStore(filepath.Join(t.TempDir(), "resume.json"))
	store.Set("https://example.com/a.mp4", time.Minute)
	if len(store.Entries) != 0 {
		t.Fatalf("URL was stored: %+v", store.Entries)
	}
}

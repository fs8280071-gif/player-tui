package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParseTime(t *testing.T) {
	cases := []struct {
		in   string
		want time.Duration
		ok   bool
	}{
		{"0", 0, true},
		{"90", 90 * time.Second, true},
		{"1:30", 90 * time.Second, true},
		{"01:02:03", time.Hour + 2*time.Minute + 3*time.Second, true},
		{"1:60", 0, false},
		{"-1", 0, false},
		{"", 0, false},
		{"1:2:3:4", 0, false},
	}
	for _, tc := range cases {
		got, err := ParseTime(tc.in)
		if tc.ok && (err != nil || got != tc.want) {
			t.Errorf("ParseTime(%q) = %v, %v; want %v", tc.in, got, err, tc.want)
		}
		if !tc.ok && err == nil {
			t.Errorf("ParseTime(%q) unexpectedly succeeded: %v", tc.in, got)
		}
	}
}

func TestParseTargetAndTitle(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "clip.mp4")
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	opts, err := Parse([]string{"play", path, "--start", "1:02", "--title", "hello\x1b[31m"})
	if err != nil {
		t.Fatal(err)
	}
	if opts.Spec.Start != 62*time.Second {
		t.Fatalf("start = %v", opts.Spec.Start)
	}
	if opts.Spec.Title != "hello[31m" {
		t.Fatalf("title not sanitized: %q", opts.Spec.Title)
	}
	if opts.Spec.Target == "" {
		t.Fatal("target is empty")
	}
}

func TestParseRejectsMissingFile(t *testing.T) {
	if _, err := Parse([]string{"play", "/does/not/exist.mkv"}); err == nil {
		t.Fatal("expected an error")
	}
}

func TestParseHelp(t *testing.T) {
	opts, err := Parse([]string{"--help"})
	if err != nil || !opts.Help {
		t.Fatalf("help: %v %v", opts.Help, err)
	}
}

func TestParseRenderOptions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "clip.mp4")
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	opts, err := Parse([]string{"bench", path, "--render", "quadrant", "--dither", "--frames", "30", "--cols", "80", "--rows", "24"})
	if err != nil {
		t.Fatal(err)
	}
	if opts.Command != CommandBench {
		t.Fatalf("command = %v", opts.Command)
	}
	if opts.Render.String() != "quadrant" || !opts.Dither {
		t.Fatalf("render = %v dither = %v", opts.Render, opts.Dither)
	}
	if opts.Frames != 30 || opts.Cols != 80 || opts.Rows != 24 {
		t.Fatalf("bench options = %+v", opts)
	}
	opts, err = Parse([]string{"play", path, "--audio"})
	if err != nil || !opts.Audio {
		t.Fatalf("audio flag: %+v %v", opts, err)
	}
}

func TestParseRenderEquals(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "clip.mp4")
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	opts, err := Parse([]string{"play", path, "--render=sextant"})
	if err != nil {
		t.Fatal(err)
	}
	if opts.Render != 3 {
		t.Fatalf("render = %v", opts.Render)
	}
	opts, err = Parse([]string{"play", path, "--render", "kitty"})
	if err != nil || opts.Render.String() != "kitty" {
		t.Fatalf("kitty render: %v %v", opts.Render, err)
	}
	if _, err := Parse([]string{"play", path, "--render", "bogus"}); err == nil {
		t.Fatal("invalid render mode accepted")
	}
}

func TestParseSpaceContainingPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "my clip (final).mp4")
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	opts, err := Parse([]string{"play", path})
	if err != nil {
		t.Fatal(err)
	}
	if opts.Spec.Title != "my clip (final).mp4" {
		t.Fatalf("title = %q", opts.Spec.Title)
	}
}

func TestParseEmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "broken.mp4")
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := Parse([]string{"play", path})
	if err == nil || !strings.Contains(err.Error(), "empty") {
		t.Fatalf("expected empty-file error, got %v", err)
	}
}

func TestParseDoctor(t *testing.T) {
	opts, err := Parse([]string{"doctor"})
	if err != nil || opts.Command != CommandDoctor {
		t.Fatalf("doctor parse: %+v %v", opts, err)
	}
	if _, err := Parse([]string{"doctor", "extra"}); err == nil {
		t.Fatal("doctor accepted an argument")
	}
}

func TestParseQueue(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.mp4")
	b := filepath.Join(dir, "b.mp4")
	for _, p := range []string{a, b} {
		if err := os.WriteFile(p, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	opts, err := Parse([]string{"play", a, b})
	if err != nil {
		t.Fatal(err)
	}
	if len(opts.Specs) != 2 || opts.Specs[0].Target != a || opts.Specs[1].Target != b {
		t.Fatalf("queue = %+v", opts.Specs)
	}
	if opts.Spec.Target != a {
		t.Fatalf("first spec = %+v", opts.Spec)
	}
	if _, err := Parse([]string{"play", a, b, "--title", "one"}); err == nil {
		t.Fatal("--title accepted with multiple targets")
	}
}

func TestParseTheme(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "clip.mp4")
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	opts, err := Parse([]string{"play", path, "--theme", "nord"})
	if err != nil || opts.Theme != "nord" {
		t.Fatalf("theme = %q, %v", opts.Theme, err)
	}
	opts, err = Parse([]string{"play", path, "--theme=gruvbox"})
	if err != nil || opts.Theme != "gruvbox" {
		t.Fatalf("theme equals = %q, %v", opts.Theme, err)
	}
}

func TestParseBrowse(t *testing.T) {
	opts, err := Parse(nil)
	if err != nil || opts.Command != CommandBrowse {
		t.Fatalf("no-arg parse: %+v %v", opts, err)
	}
	dir := t.TempDir()
	opts, err = Parse([]string{"browse", dir})
	if err != nil || opts.Command != CommandBrowse || opts.BrowseDir == "" {
		t.Fatalf("browse parse: %+v %v", opts, err)
	}
	file := filepath.Join(dir, "clip.mp4")
	if err := os.WriteFile(file, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Parse([]string{"browse", file}); err == nil {
		t.Fatal("browse accepted a file")
	}
}

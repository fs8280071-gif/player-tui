package cli

import (
	"os"
	"path/filepath"
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
	if _, err := Parse([]string{"play", path, "--render", "kitty"}); err == nil {
		t.Fatal("invalid render mode accepted")
	}
}

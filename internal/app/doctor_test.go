package app

import (
	"strings"
	"testing"

	"github.com/fs8280071-gif/player-tui/internal/render"
)

func TestDoctor(t *testing.T) {
	env := map[string]string{
		"TERM":         "xterm-kitty",
		"COLORTERM":    "truecolor",
		"TERM_PROGRAM": "kitty",
	}
	getenv := func(k string) string { return env[k] }
	out := Doctor(getenv, DoctorOptions{Width: 120, Height: 40, TTY: true})
	for _, want := range []string{"120x40", "xterm-kitty", "truecolor", "render: kitty", "kitty graphics: true"} {
		if !strings.Contains(out, want) {
			t.Errorf("doctor output missing %q:\n%s", want, out)
		}
	}
	out = Doctor(func(string) string { return "" }, DoctorOptions{Render: render.ModeHalfblock})
	if !strings.Contains(out, "not a tty") || !strings.Contains(out, "render: halfblock (from flag)") {
		t.Errorf("unexpected doctor output:\n%s", out)
	}
}

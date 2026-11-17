package app

import (
	"fmt"
	"strings"

	"github.com/fs8280071-gif/player-tui/internal/render"
	"github.com/fs8280071-gif/player-tui/internal/sanitize"
)

type DoctorOptions struct {
	Width  int
	Height int
	TTY    bool
	Render render.Mode
}

func Doctor(getenv func(string) string, opts DoctorOptions) string {
	caps := render.Detect(getenv)
	auto := render.AutoMode(getenv)
	mode := auto
	source := "auto"
	if opts.Render != render.ModeAuto {
		mode = opts.Render
		source = "flag"
	}

	var b strings.Builder
	b.WriteString("player-tui doctor\n")
	if opts.TTY && opts.Width > 0 && opts.Height > 0 {
		fmt.Fprintf(&b, "terminal: %dx%d cells\n", opts.Width, opts.Height)
	} else {
		b.WriteString("terminal: not a tty (size unknown)\n")
	}
	fmt.Fprintf(&b, "TERM: %s\n", valueOr(caps.TERM, "(unset)"))
	fmt.Fprintf(&b, "COLORTERM: %s\n", valueOr(caps.ColorTerm, "(unset)"))
	fmt.Fprintf(&b, "TERM_PROGRAM: %s\n", valueOr(caps.TermProgram, "(unset)"))
	fmt.Fprintf(&b, "capabilities: %s\n", caps.Summary())
	fmt.Fprintf(&b, "truecolor: %t\n", caps.TrueColor)
	fmt.Fprintf(&b, "kitty graphics: %t\n", caps.Kitty)
	fmt.Fprintf(&b, "sixel: %t\n", caps.Sixel)
	fmt.Fprintf(&b, "iterm2 inline: %t\n", caps.ITerm2)
	fmt.Fprintf(&b, "render: %s (from %s), auto would pick %s\n", mode, source, auto)
	b.WriteString("fallback: quadrant -> halfblock (TERM=dumb/linux or PLAYER_TUI_ASCII=1)\n")
	b.WriteString("audio: off by default; pass --audio to enable libVLC audio output\n")
	b.WriteString("graphics encoders are unit-tested but not validated on this terminal\n")
	return b.String()
}

func valueOr(s, fallback string) string {
	s = sanitize.Text(s)
	if strings.TrimSpace(s) == "" {
		return fallback
	}
	return s
}

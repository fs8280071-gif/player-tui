package render

import "strings"

type Caps struct {
	TERM        string
	ColorTerm   string
	TermProgram string
	Kitty       bool
	Sixel       bool
	ITerm2      bool
	TrueColor   bool
}

func Detect(getenv func(string) string) Caps {
	c := Caps{
		TERM:        strings.TrimSpace(getenv("TERM")),
		ColorTerm:   strings.TrimSpace(getenv("COLORTERM")),
		TermProgram: strings.TrimSpace(getenv("TERM_PROGRAM")),
	}
	term := strings.ToLower(c.TERM)
	prog := strings.ToLower(c.TermProgram)
	colorTerm := strings.ToLower(c.ColorTerm)

	c.Kitty = getenv("KITTY_WINDOW_ID") != "" || strings.Contains(term, "kitty")
	c.ITerm2 = prog == "iterm.app" || prog == "iterm2" || prog == "wezterm"
	c.Sixel = strings.Contains(term, "sixel") || prog == "mlterm" || prog == "windows_terminal"
	c.TrueColor = strings.Contains(colorTerm, "truecolor") ||
		strings.Contains(colorTerm, "24bit") ||
		strings.Contains(term, "direct")
	return c
}

func (c Caps) PreferredMode() Mode {
	switch {
	case c.Kitty:
		return ModeKitty
	case c.ITerm2:
		return ModeITerm2
	case c.Sixel:
		return ModeSixel
	default:
		return ModeQuadrant
	}
}

func (c Caps) Summary() string {
	var b strings.Builder
	if c.Kitty {
		b.WriteString("kitty ")
	}
	if c.ITerm2 {
		b.WriteString("iterm2 ")
	}
	if c.Sixel {
		b.WriteString("sixel ")
	}
	if c.TrueColor {
		b.WriteString("truecolor ")
	}
	if b.Len() == 0 {
		return "block-glyphs"
	}
	return strings.TrimSpace(b.String())
}

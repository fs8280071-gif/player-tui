// Package theme provides semantic terminal color roles.
package theme

import (
	"fmt"
	"strconv"
	"strings"
)

type RGB struct {
	R uint8
	G uint8
	B uint8
}

type Theme struct {
	Name   string
	Fg     RGB
	Bg     RGB
	Accent RGB
	HelpBg RGB
	HelpFg RGB
}

var themes = map[string]Theme{
	"catppuccin": {
		Name: "catppuccin", Fg: RGB{0xcd, 0xd6, 0xf4}, Bg: RGB{0x1e, 0x1e, 0x2e},
		Accent: RGB{0x89, 0xb4, 0xfa}, HelpBg: RGB{0x31, 0x32, 0x44}, HelpFg: RGB{0xf5, 0xe0, 0xdc},
	},
	"gruvbox": {
		Name: "gruvbox", Fg: RGB{0xeb, 0xdb, 0xb2}, Bg: RGB{0x28, 0x28, 0x28},
		Accent: RGB{0xfa, 0xbd, 0x2f}, HelpBg: RGB{0x3c, 0x38, 0x36}, HelpFg: RGB{0xeb, 0xdb, 0xb2},
	},
	"nord": {
		Name: "nord", Fg: RGB{0xec, 0xef, 0xf4}, Bg: RGB{0x2e, 0x34, 0x40},
		Accent: RGB{0x88, 0xc0, 0xd0}, HelpBg: RGB{0x3b, 0x42, 0x52}, HelpFg: RGB{0xe5, 0xe9, 0xf0},
	},
	"high-contrast": {
		Name: "high-contrast", Fg: RGB{0xff, 0xff, 0xff}, Bg: RGB{0x00, 0x00, 0x00},
		Accent: RGB{0x00, 0xff, 0xff}, HelpBg: RGB{0x00, 0x00, 0x00}, HelpFg: RGB{0xff, 0xff, 0xff},
	},
}

func Names() []string {
	return []string{"catppuccin", "gruvbox", "nord", "high-contrast"}
}

func Parse(name string) (Theme, error) {
	key := strings.ToLower(strings.TrimSpace(name))
	if key == "" {
		key = "catppuccin"
	}
	if t, ok := themes[key]; ok {
		return t, nil
	}
	return Theme{}, fmt.Errorf("unknown theme %q (use %s)", name, strings.Join(Names(), ", "))
}

func (t Theme) StatusStyle(noColor bool) string {
	if noColor {
		return "\x1b[7m"
	}
	return fgBg(t.Fg, t.Bg)
}

func (t Theme) HelpStyle(noColor bool) string {
	if noColor {
		return "\x1b[7m"
	}
	return fgBg(t.HelpFg, t.HelpBg)
}

func fgBg(fg, bg RGB) string {
	return "\x1b[38;2;" + rgb(fg) + "m\x1b[48;2;" + rgb(bg) + "m"
}

func rgb(c RGB) string {
	return strconv.Itoa(int(c.R)) + ";" + strconv.Itoa(int(c.G)) + ";" + strconv.Itoa(int(c.B))
}

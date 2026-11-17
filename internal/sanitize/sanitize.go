// Package sanitize strips terminal-active characters from untrusted text.
package sanitize

import (
	"strings"
	"unicode"
)

const maxRunes = 200

func Text(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	n := 0
	for _, r := range s {
		if r == '\t' {
			r = ' '
		} else if isUnsafe(r) {
			continue
		}
		if n >= maxRunes {
			break
		}
		b.WriteRune(r)
		n++
	}
	out := strings.TrimSpace(b.String())
	if out == "" {
		return ""
	}
	return out
}

func isUnsafe(r rune) bool {
	if unicode.IsControl(r) {
		return true
	}
	switch {
	case r >= 0x200b && r <= 0x200f,
		r >= 0x202a && r <= 0x202e,
		r >= 0x2060 && r <= 0x2069,
		r == 0xfeff:
		return true
	}
	return false
}

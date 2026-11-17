package sanitize

import (
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"
)

func TestTextStripsTerminalControls(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"\x1b[31mred\x1b[0m", "[31mred[0m"},
		{"hello\x00world", "helloworld"},
		{"a\x07b", "ab"},
		{"c1:\u009b2J", "c1:2J"},
		{"bidi:\u202eevil\u202c", "bidi:evil"},
		{"tabs\there", "tabs here"},
		{"  padded  ", "padded"},
		{"\xff\xfe", "\ufffd\ufffd"},
	}
	for _, tc := range cases {
		if got := Text(tc.in); got != tc.want {
			t.Errorf("Text(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestTextCapsLength(t *testing.T) {
	out := Text(strings.Repeat("a", 500))
	if utf8.RuneCountInString(out) != maxRunes {
		t.Fatalf("length = %d, want %d", utf8.RuneCountInString(out), maxRunes)
	}
}

func TestTextEmpty(t *testing.T) {
	for _, in := range []string{"", "\x1b", "\x00", " \t "} {
		if got := Text(in); got != "" {
			t.Errorf("Text(%q) = %q, want empty", in, got)
		}
	}
}

func FuzzText(f *testing.F) {
	for _, seed := range []string{
		"", "clip.mp4", "\x1b[2J", "\x1b]52;c;Zm9v\x07", "\u009b31m",
		"subtitle\nline", "\u202eabc", "日本語 タイトル", "\xff\xfe",
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, in string) {
		out := Text(in)
		if utf8.RuneCountInString(out) > maxRunes {
			t.Fatalf("output too long: %d", utf8.RuneCountInString(out))
		}
		for _, r := range out {
			if unicode.IsControl(r) {
				t.Fatalf("control %U survived: %q", r, out)
			}
			switch {
			case r >= 0x200b && r <= 0x200f,
				r >= 0x202a && r <= 0x202e,
				r >= 0x2060 && r <= 0x2069,
				r == 0xfeff:
				t.Fatalf("bidi control %U survived: %q", r, out)
			}
		}
		if strings.ContainsRune(out, 0x1b) {
			t.Fatalf("ESC survived: %q", out)
		}
	})
}

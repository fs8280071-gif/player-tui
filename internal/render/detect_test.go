package render

import "testing"

func TestDetect(t *testing.T) {
	cases := []struct {
		name string
		env  map[string]string
		want Caps
		mode Mode
	}{
		{
			name: "kitty",
			env:  map[string]string{"TERM": "xterm-kitty", "KITTY_WINDOW_ID": "1", "COLORTERM": "truecolor"},
			want: Caps{TERM: "xterm-kitty", Kitty: true, TrueColor: true},
			mode: ModeKitty,
		},
		{
			name: "iterm2",
			env:  map[string]string{"TERM": "xterm-256color", "TERM_PROGRAM": "iTerm.app"},
			want: Caps{TERM: "xterm-256color", TermProgram: "iTerm.app", ITerm2: true},
			mode: ModeITerm2,
		},
		{
			name: "sixel",
			env:  map[string]string{"TERM": "xterm-sixel"},
			want: Caps{TERM: "xterm-sixel", Sixel: true},
			mode: ModeSixel,
		},
		{
			name: "plain",
			env:  map[string]string{"TERM": "xterm-256color"},
			want: Caps{TERM: "xterm-256color"},
			mode: ModeQuadrant,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			getenv := func(key string) string { return tc.env[key] }
			got := Detect(getenv)
			if got.Kitty != tc.want.Kitty || got.Sixel != tc.want.Sixel ||
				got.ITerm2 != tc.want.ITerm2 || got.TrueColor != tc.want.TrueColor {
				t.Fatalf("Detect = %+v, want %+v", got, tc.want)
			}
			if got.PreferredMode() != tc.mode {
				t.Fatalf("preferred = %v, want %v", got.PreferredMode(), tc.mode)
			}
		})
	}
}

func TestAutoModePrefersGraphics(t *testing.T) {
	if got := AutoMode(func(k string) string {
		return map[string]string{"TERM": "xterm-kitty", "KITTY_WINDOW_ID": "1"}[k]
	}); got != ModeKitty {
		t.Fatalf("kitty auto = %v", got)
	}
	if got := AutoMode(func(k string) string {
		return map[string]string{"TERM": "xterm-256color", "TERM_PROGRAM": "iTerm.app"}[k]
	}); got != ModeITerm2 {
		t.Fatalf("iterm auto = %v", got)
	}
	if got := AutoMode(func(k string) string {
		return map[string]string{"TERM": "xterm-sixel"}[k]
	}); got != ModeSixel {
		t.Fatalf("sixel auto = %v", got)
	}
}

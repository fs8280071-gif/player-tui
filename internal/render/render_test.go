package render

import (
	"bytes"
	"strings"
	"testing"

	"github.com/fs8280071-gif/player-tui/internal/video"
)

func TestFitPreservesAspect(t *testing.T) {
	cols, rows, row, col := Fit(120, 40, 16.0/9.0)
	if row < 1 || col < 1 || cols < 1 || rows < 1 {
		t.Fatalf("bad fit: %d %d %d %d", cols, rows, row, col)
	}
	got := float64(cols) / (2 * float64(rows))
	if got < 1.70 || got > 1.85 {
		t.Fatalf("aspect drifted: cols=%d rows=%d ratio=%.3f", cols, rows, got)
	}
	if row+rows-1 > 39 {
		t.Fatalf("video overlaps status row: start=%d rows=%d", row, rows)
	}
}

func TestFit1920x1088(t *testing.T) {
	cols, rows, row, col := Fit(120, 40, 1920.0/1088.0)
	if cols != 120 || rows != 34 {
		t.Fatalf("fit = %dx%d at %d,%d; want 120x34", cols, rows, row, col)
	}
	got := float64(cols) / (2 * float64(rows))
	want := 1920.0 / 1088.0
	if got < want-0.001 || got > want+0.001 {
		t.Fatalf("aspect = %.4f, want %.4f", got, want)
	}
}

func TestGlyphTablesComplete(t *testing.T) {
	seen := make(map[rune]bool)
	for mask, glyph := range quadrantGlyphs {
		if glyph == 0 {
			t.Fatalf("quadrant glyph %d is empty", mask)
		}
		if glyph != ' ' && seen[glyph] {
			t.Fatalf("duplicate quadrant glyph %U", glyph)
		}
		seen[glyph] = true
	}
	seen = make(map[rune]bool)
	for mask, glyph := range sextantGlyphs {
		if glyph == 0 {
			t.Fatalf("sextant glyph %d is empty", mask)
		}
		if glyph != ' ' && seen[glyph] {
			t.Fatalf("duplicate sextant glyph %U", glyph)
		}
		seen[glyph] = true
	}
}

func TestAppendDrawsHalfBlocks(t *testing.T) {
	f := testFrame(4, 4)
	r := New(ModeHalfblock, false)
	out := r.Append(nil, f, 4, 2, 1, 1)
	if !bytes.Contains(out, []byte("\xe2\x96\x80")) {
		t.Fatalf("no half-block glyph in output: %q", out)
	}
	if !bytes.Contains(out, []byte("\x1b[38;2;")) {
		t.Fatalf("no truecolor foreground in output")
	}
}

func TestAppendModesUseBlockGlyphs(t *testing.T) {
	f := testFrame(8, 12)
	for _, tc := range []struct {
		mode  Mode
		glyph []rune
	}{
		{ModeHalfblock, []rune{'\u2580'}},
		{ModeQuadrant, quadrantGlyphs[:]},
		{ModeSextant, sextantGlyphs[:]},
	} {
		r := New(tc.mode, false)
		out := string(r.Append(nil, f, 4, 2, 1, 1))
		if !containsRune(out, tc.glyph) {
			t.Errorf("%s output has no expected block glyph", tc.mode)
		}
	}
}

func TestAppend1920x1088(t *testing.T) {
	f := uniformFrame(1920, 1088, 100)
	r := New(ModeSextant, false)
	out := r.Append(nil, f, 1, 1, 1, 1)
	if len(out) == 0 {
		t.Fatal("empty output")
	}
	cols, rows, _, _ := Fit(120, 40, float64(f.Width)/float64(f.Height))
	if cols != 120 || rows != 34 {
		t.Fatalf("fit = %dx%d", cols, rows)
	}
}

func TestDitherChangesOutput(t *testing.T) {
	f := uniformFrame(8, 12, 100)
	plain := New(ModeQuadrant, false).Append(nil, f, 4, 2, 1, 1)
	dithered := New(ModeQuadrant, true).Append(nil, f, 4, 2, 1, 1)
	if bytes.Equal(plain, dithered) {
		t.Fatal("dither did not change output")
	}
}

func TestAutoMode(t *testing.T) {
	tests := []struct {
		env  map[string]string
		want Mode
	}{
		{map[string]string{"TERM": "dumb"}, ModeHalfblock},
		{map[string]string{"TERM": ""}, ModeHalfblock},
		{map[string]string{"TERM": "linux"}, ModeHalfblock},
		{map[string]string{"TERM": "xterm-256color"}, ModeQuadrant},
		{map[string]string{"TERM": "xterm-256color", "PLAYER_TUI_ASCII": "1"}, ModeHalfblock},
	}
	for _, tc := range tests {
		got := AutoMode(func(key string) string { return tc.env[key] })
		if got != tc.want {
			t.Errorf("AutoMode(%v) = %v; want %v", tc.env, got, tc.want)
		}
	}
}

func TestParseMode(t *testing.T) {
	for name, want := range map[string]Mode{
		"auto": ModeAuto, "halfblock": ModeHalfblock, "half-block": ModeHalfblock,
		"quadrant": ModeQuadrant, "quad": ModeQuadrant,
		"sextant": ModeSextant, "2x3": ModeSextant,
	} {
		got, err := ParseMode(name)
		if err != nil || got != want {
			t.Errorf("ParseMode(%q) = %v, %v; want %v", name, got, err, want)
		}
	}
	if _, err := ParseMode("kitty"); err == nil {
		t.Fatal("invalid mode accepted")
	}
}

func containsRune(s string, set []rune) bool {
	for _, r := range set {
		if r != ' ' && strings.ContainsRune(s, r) {
			return true
		}
	}
	return false
}

func testFrame(w, h int) *video.Frame {
	f := uniformFrame(w, h, 16)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			f.Y[y*f.StrideY+x] = uint8(x * 255 / max(1, w-1))
		}
	}
	return f
}

func uniformFrame(w, h int, y uint8) *video.Frame {
	f := &video.Frame{
		Width: w, Height: h,
		Y:       bytes.Repeat([]byte{y}, w*h),
		U:       bytes.Repeat([]byte{128}, ((w+1)/2)*((h+1)/2)),
		V:       bytes.Repeat([]byte{128}, ((w+1)/2)*((h+1)/2)),
		StrideY: w,
		StrideU: (w + 1) / 2,
		StrideV: (w + 1) / 2,
	}
	return f
}

func TestResolve(t *testing.T) {
	t.Setenv("PLAYER_TUI_RENDER", "sextant")
	t.Setenv("TERM", "xterm-256color")
	mode, source := Resolve(ModeAuto)
	if mode != ModeSextant || source != "environment" {
		t.Fatalf("Resolve env = %v, %q", mode, source)
	}
	mode, source = Resolve(ModeHalfblock)
	if mode != ModeHalfblock || source != "flag" {
		t.Fatalf("Resolve flag = %v, %q", mode, source)
	}
	t.Setenv("PLAYER_TUI_RENDER", "bogus")
	mode, source = Resolve(ModeAuto)
	if mode != ModeQuadrant || source != "auto" {
		t.Fatalf("Resolve invalid env = %v, %q", mode, source)
	}
}

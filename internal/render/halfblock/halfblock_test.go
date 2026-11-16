package halfblock

import (
	"bytes"
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
}

func TestAppendDrawsHalfBlocks(t *testing.T) {
	f := &video.Frame{
		Width: 4, Height: 4,
		Y:       bytes.Repeat([]byte{100}, 16),
		U:       bytes.Repeat([]byte{100}, 4),
		V:       bytes.Repeat([]byte{100}, 4),
		StrideY: 4, StrideU: 2, StrideV: 2,
	}
	r := New()
	out := r.Append(nil, f, 4, 2, 1, 1)
	if !bytes.Contains(out, []byte("\xe2\x96\x80")) {
		t.Fatalf("no half-block glyph in output: %q", out)
	}
	if !bytes.Contains(out, []byte("\x1b[38;2;")) {
		t.Fatalf("no truecolor foreground in output")
	}
}

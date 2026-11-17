package iterm2

import "testing"

func TestEncodeOnePixel(t *testing.T) {
	got := string(Encode(nil, 1, 1, []byte{0xff, 0x00, 0x00}, Options{}))
	want := "\x1b]1337;File=inline=1;size=3;preserveAspectRatio=1:/wAA\x07"
	if got != want {
		t.Fatalf("iterm2 mismatch\n got %q\nwant %q", got, want)
	}
}

func TestEncodePlacement(t *testing.T) {
	got := string(Encode(nil, 1, 1, []byte{0x00, 0x00, 0x00}, Options{Cols: 80, Rows: 24}))
	want := "\x1b]1337;File=inline=1;size=3;width=80;height=24;preserveAspectRatio=1:AAAA\x07"
	if got != want {
		t.Fatalf("placement mismatch\n got %q\nwant %q", got, want)
	}
}

func TestEncodeRejectsBadBuffer(t *testing.T) {
	if got := Encode(nil, 2, 2, nil, Options{}); len(got) != 0 {
		t.Fatalf("expected no output, got %q", got)
	}
}

package sixel

import "testing"

func TestEncodeOneWhitePixel(t *testing.T) {
	got := string(Encode(nil, 1, 1, []byte{255, 255, 255}))
	want := "\x1bPq\"1;1;1;1#231;2;100;100;100#231@$-\x1b\\"
	if got != want {
		t.Fatalf("sixel mismatch\n got %q\nwant %q", got, want)
	}
}

func TestEncodeOneBlackPixel(t *testing.T) {
	got := string(Encode(nil, 1, 1, []byte{0, 0, 0}))
	want := "\x1bPq\"1;1;1;1#16;2;0;0;0#16@$-\x1b\\"
	if got != want {
		t.Fatalf("sixel mismatch\n got %q\nwant %q", got, want)
	}
}

func TestEncodeTwoBands(t *testing.T) {
	img := make([]byte, 1*7*3)
	for i := range img {
		img[i] = 255
	}
	got := string(Encode(nil, 1, 7, img))
	want := "\x1bPq\"1;1;1;7#231;2;100;100;100#231~$-#231@$-\x1b\\"
	if got != want {
		t.Fatalf("band mismatch\n got %q\nwant %q", got, want)
	}
}

func TestEncodeRejectsBadBuffer(t *testing.T) {
	if got := Encode(nil, 2, 2, []byte{1, 2, 3}); len(got) != 0 {
		t.Fatalf("expected no output, got %q", got)
	}
}

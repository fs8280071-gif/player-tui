package kitty

import "testing"

func TestEncodeOnePixel(t *testing.T) {
	got := string(Encode(nil, 1, 1, []byte{0xff, 0x00, 0x00}, Options{ImageID: 1}))
	want := "\x1b_Ga=T,f=24,s=1,v=1,i=1,q=2;/wAA\x1b\\"
	if got != want {
		t.Fatalf("encode mismatch\n got %q\nwant %q", got, want)
	}
}

func TestEncodeChunks(t *testing.T) {
	got := string(Encode(nil, 1, 1, []byte{0xff, 0x00, 0x00}, Options{ImageID: 7, ChunkSize: 2}))
	want := "\x1b_Ga=T,f=24,s=1,v=1,i=7,q=2,m=1;/w\x1b\\" +
		"\x1b_Gm=0;AA\x1b\\"
	if got != want {
		t.Fatalf("chunk mismatch\n got %q\nwant %q", got, want)
	}
}

func TestEncodePlacement(t *testing.T) {
	got := string(Encode(nil, 1, 1, []byte{0x00, 0x00, 0x00}, Options{ImageID: 2, Cols: 80, Rows: 24}))
	want := "\x1b_Ga=T,f=24,s=1,v=1,i=2,c=80,r=24,q=2;AAAA\x1b\\"
	if got != want {
		t.Fatalf("placement mismatch\n got %q\nwant %q", got, want)
	}
}

func TestDelete(t *testing.T) {
	got := string(Delete(nil, 42))
	want := "\x1b_Ga=d,d=I,i=42\x1b\\"
	if got != want {
		t.Fatalf("delete mismatch\n got %q\nwant %q", got, want)
	}
}

func TestEncodeRejectsBadBuffer(t *testing.T) {
	if got := Encode(nil, 2, 2, []byte{1, 2, 3}, Options{}); len(got) != 0 {
		t.Fatalf("expected no output, got %q", got)
	}
}

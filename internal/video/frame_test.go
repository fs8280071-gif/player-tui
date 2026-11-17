package video

import (
	"math"
	"testing"
)

func TestAverageYUVUsesWholeArea(t *testing.T) {
	f := newTestFrame(8, 8, 0)
	f.Y[0] = 255
	y, u, v := f.AverageYUV(0, 0, 8, 8)
	if math.Abs(y-255.0/64.0) > 1e-9 {
		t.Fatalf("Y average = %v, want %v", y, 255.0/64.0)
	}
	if u != 128 || v != 128 {
		t.Fatalf("chroma = %v,%v; want 128,128", u, v)
	}
}

func TestAverageYUVSubrect(t *testing.T) {
	f := newTestFrame(8, 8, 16)
	for y := 0; y < 8; y++ {
		for x := 4; x < 8; x++ {
			f.Y[y*f.StrideY+x] = 240
		}
	}
	left, _, _ := f.AverageYUV(0, 0, 4, 8)
	right, _, _ := f.AverageYUV(4, 0, 8, 8)
	if left != 16 || right != 240 {
		t.Fatalf("left=%v right=%v; want 16,240", left, right)
	}
}

func TestAverageYUV1920x1088(t *testing.T) {
	f := newTestFrame(1920, 1088, 100)
	for i := range f.U {
		f.U[i] = 90
		f.V[i] = 110
	}
	y, u, v := f.AverageYUV(0, 0, f.Width, f.Height)
	if y != 100 || u != 90 || v != 110 {
		t.Fatalf("average = %v,%v,%v; want 100,90,110", y, u, v)
	}
}

func TestAverageYUVOddSize(t *testing.T) {
	f := newTestFrame(5, 5, 42)
	y, u, v := f.AverageYUV(0, 0, 5, 5)
	if y != 42 || u != 128 || v != 128 {
		t.Fatalf("average = %v,%v,%v; want 42,128,128", y, u, v)
	}
}

func newTestFrame(w, h int, y uint8) *Frame {
	f := &Frame{
		Width: w, Height: h,
		Y:       make([]byte, w*h),
		U:       make([]byte, ((w+1)/2)*((h+1)/2)),
		V:       make([]byte, ((w+1)/2)*((h+1)/2)),
		StrideY: w,
		StrideU: (w + 1) / 2,
		StrideV: (w + 1) / 2,
	}
	for i := range f.Y {
		f.Y[i] = y
	}
	for i := range f.U {
		f.U[i] = 128
		f.V[i] = 128
	}
	return f
}

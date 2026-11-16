package video

import (
	"math"
	"time"
)

type Frame struct {
	Width     int
	Height    int
	Y         []byte
	U         []byte
	V         []byte
	StrideY   int
	StrideU   int
	StrideV   int
	MediaTime time.Duration
	Arrival   time.Time
	Sequence  uint64
}

func (f *Frame) ok() bool {
	return f != nil && f.Width > 0 && f.Height > 0 &&
		len(f.Y) >= f.StrideY*f.Height &&
		len(f.U) >= f.StrideU*((f.Height+1)/2) &&
		len(f.V) >= f.StrideV*((f.Height+1)/2)
}

func (f *Frame) sample(x, y int) (float64, float64, float64) {
	if !f.ok() {
		return 16, 128, 128
	}
	if x < 0 {
		x = 0
	} else if x >= f.Width {
		x = f.Width - 1
	}
	if y < 0 {
		y = 0
	} else if y >= f.Height {
		y = f.Height - 1
	}
	yy := float64(f.Y[y*f.StrideY+x])
	cx, cy := x/2, y/2
	u := float64(f.U[cy*f.StrideU+cx])
	v := float64(f.V[cy*f.StrideV+cx])
	return yy, u, v
}

// AverageYUV box-samples the source rectangle [x0,x1) x [y0,y1).
func (f *Frame) AverageYUV(x0, y0, x1, y1 int) (float64, float64, float64) {
	if !f.ok() {
		return 16, 128, 128
	}
	if x1 <= x0 {
		x1 = x0 + 1
	}
	if y1 <= y0 {
		y1 = y0 + 1
	}
	if x0 < 0 {
		x0 = 0
	}
	if y0 < 0 {
		y0 = 0
	}
	if x1 > f.Width {
		x1 = f.Width
	}
	if y1 > f.Height {
		y1 = f.Height
	}
	sx := max(1, (x1-x0)/3)
	sy := max(1, (y1-y0)/3)
	var sumY, sumU, sumV, n float64
	for y := y0; y < y1; y += sy {
		for x := x0; x < x1; x += sx {
			yy, u, v := f.sample(x, y)
			sumY += yy
			sumU += u
			sumV += v
			n++
		}
	}
	if n == 0 {
		return f.sample((x0+x1)/2, (y0+y1)/2)
	}
	return sumY / n, sumU / n, sumV / n
}

func clamp8(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(math.Round(v))
}

// YUVToRGB converts studio-swing BT.601 I420 to 8-bit RGB.
func YUVToRGB(y, u, v float64) (uint8, uint8, uint8) {
	y = (y - 16) * 1.164
	u -= 128
	v -= 128
	r := y + 1.596*v
	g := y - 0.392*u - 0.813*v
	b := y + 2.017*u
	return clamp8(r), clamp8(g), clamp8(b)
}

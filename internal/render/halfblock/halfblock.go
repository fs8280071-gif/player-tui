package halfblock

import (
	"math"
	"strconv"

	"github.com/fs8280071-gif/player-tui/internal/video"
)

type Renderer struct {
	buf    []byte
	lastFg int32
	lastBg int32
}

func New() *Renderer {
	return &Renderer{lastFg: -1, lastBg: -1}
}

// Fit computes a centered video rectangle in terminal cells. The renderer uses
// two vertical pixels per cell, so a square-pixel video of aspect a needs
// cols/(2*rows) == a.
func Fit(termW, termH int, aspect float64) (cols, rows, startRow, startCol int) {
	if aspect <= 0 || math.IsNaN(aspect) || math.IsInf(aspect, 0) {
		aspect = 16.0 / 9.0
	}
	availRows := termH - 1
	if availRows < 1 {
		availRows = 1
	}
	if termW < 1 {
		termW = 1
	}
	cols = termW
	rows = int(math.Round(float64(cols) / (aspect * 2)))
	if rows < 1 {
		rows = 1
	}
	if rows > availRows {
		rows = availRows
		cols = int(math.Round(float64(rows) * aspect * 2))
	}
	if cols < 1 {
		cols = 1
	}
	if cols > termW {
		cols = termW
	}
	startRow = (termH-1-rows)/2 + 1
	startCol = (termW-cols)/2 + 1
	return cols, rows, startRow, startCol
}

// Append renders the frame into dst as truecolor half-block cells.
func (r *Renderer) Append(dst []byte, f *video.Frame, cols, rows, startRow, startCol int) []byte {
	dst = dst[:0]
	if f == nil || cols <= 0 || rows <= 0 {
		return dst
	}
	r.lastFg, r.lastBg = -1, -1

	for row := 0; row < rows; row++ {
		dst = appendCursor(dst, startRow+row, startCol)
		topY0 := (row * 2) * f.Height / (rows * 2)
		topY1 := (row*2 + 1) * f.Height / (rows * 2)
		botY0 := (row*2 + 1) * f.Height / (rows * 2)
		botY1 := (row*2 + 2) * f.Height / (rows * 2)
		if topY1 <= topY0 {
			topY1 = topY0 + 1
		}
		if botY1 <= botY0 {
			botY1 = botY0 + 1
		}
		for col := 0; col < cols; col++ {
			x0 := col * f.Width / cols
			x1 := (col + 1) * f.Width / cols
			if x1 <= x0 {
				x1 = x0 + 1
			}
			ty, tu, tv := f.AverageYUV(x0, topY0, x1, topY1)
			by, bu, bv := f.AverageYUV(x0, botY0, x1, botY1)
			tr, tg, tb := video.YUVToRGB(ty, tu, tv)
			br, bg, bb := video.YUVToRGB(by, bu, bv)
			dst = r.appendColors(dst, tr, tg, tb, br, bg, bb)
			dst = append(dst, 0xe2, 0x96, 0x80) // U+2580 UPPER HALF BLOCK
		}
	}
	dst = append(dst, "\x1b[0m"...)
	return dst
}

func (r *Renderer) appendColors(dst []byte, tr, tg, tb, br, bg, bb uint8) []byte {
	fg := int32(tr)<<16 | int32(tg)<<8 | int32(tb)
	bgc := int32(br)<<16 | int32(bg)<<8 | int32(bb)
	if fg != r.lastFg {
		dst = append(dst, "\x1b[38;2;"...)
		dst = strconv.AppendInt(dst, int64(tr), 10)
		dst = append(dst, ';')
		dst = strconv.AppendInt(dst, int64(tg), 10)
		dst = append(dst, ';')
		dst = strconv.AppendInt(dst, int64(tb), 10)
		dst = append(dst, 'm')
		r.lastFg = fg
	}
	if bgc != r.lastBg {
		dst = append(dst, "\x1b[48;2;"...)
		dst = strconv.AppendInt(dst, int64(br), 10)
		dst = append(dst, ';')
		dst = strconv.AppendInt(dst, int64(bg), 10)
		dst = append(dst, ';')
		dst = strconv.AppendInt(dst, int64(bb), 10)
		dst = append(dst, 'm')
		r.lastBg = bgc
	}
	return dst
}

func appendCursor(dst []byte, row, col int) []byte {
	dst = append(dst, "\x1b["...)
	dst = strconv.AppendInt(dst, int64(row), 10)
	dst = append(dst, ';')
	dst = strconv.AppendInt(dst, int64(col), 10)
	dst = append(dst, 'H')
	return dst
}

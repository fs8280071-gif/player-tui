package render

import (
	"github.com/fs8280071-gif/player-tui/internal/render/iterm2"
	"github.com/fs8280071-gif/player-tui/internal/render/kitty"
	"github.com/fs8280071-gif/player-tui/internal/render/sixel"
	"github.com/fs8280071-gif/player-tui/internal/video"
)

type Image struct {
	Width  int
	Height int
	Pix    []byte
}

func ImageFromFrame(f *video.Frame, width, height int) Image {
	if f == nil || f.Width <= 0 || f.Height <= 0 || width <= 0 || height <= 0 {
		return Image{}
	}
	img := Image{Width: width, Height: height, Pix: make([]byte, width*height*3)}
	for y := 0; y < height; y++ {
		y0, y1 := span(y, height, f.Height)
		for x := 0; x < width; x++ {
			x0, x1 := span(x, width, f.Width)
			yy, u, v := f.AverageYUV(x0, y0, x1, y1)
			r, g, b := video.YUVToRGB(yy, u, v)
			i := (y*width + x) * 3
			img.Pix[i] = r
			img.Pix[i+1] = g
			img.Pix[i+2] = b
		}
	}
	return img
}

func (r *Renderer) appendGraphics(dst []byte, f *video.Frame, cols, rows, startRow, startCol int) []byte {
	img := ImageFromFrame(f, cols*2, rows*4)
	if img.Width == 0 {
		return dst
	}
	if r.mode == ModeKitty && r.frames > 0 {
		dst = kitty.Delete(dst, r.imageID)
	}
	dst = appendCursor(dst, startRow, startCol)
	switch r.mode {
	case ModeKitty:
		dst = kitty.Encode(dst, img.Width, img.Height, img.Pix, kitty.Options{
			ImageID: r.imageID, Cols: cols, Rows: rows,
		})
	case ModeSixel:
		dst = sixel.Encode(dst, img.Width, img.Height, img.Pix)
	case ModeITerm2:
		dst = iterm2.Encode(dst, img.Width, img.Height, img.Pix, iterm2.Options{
			Cols: cols, Rows: rows,
		})
	}
	r.frames++
	return dst
}

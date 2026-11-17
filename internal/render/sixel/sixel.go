// Package sixel encodes raw RGB frames as sixel data.
package sixel

import (
	"strconv"
)

const levels = 6

func Encode(dst []byte, width, height int, rgb []byte) []byte {
	if width <= 0 || height <= 0 || len(rgb) != width*height*3 {
		return dst
	}
	indices := make([]uint8, width*height)
	for i := 0; i < width*height; i++ {
		indices[i] = uint8(paletteIndex(rgb[i*3], rgb[i*3+1], rgb[i*3+2]))
	}

	dst = append(dst, "\x1bPq"...)
	dst = append(dst, '"')
	dst = strconv.AppendInt(dst, 1, 10)
	dst = append(dst, ';')
	dst = strconv.AppendInt(dst, 1, 10)
	dst = append(dst, ';')
	dst = strconv.AppendInt(dst, int64(width), 10)
	dst = append(dst, ';')
	dst = strconv.AppendInt(dst, int64(height), 10)

	used := make([]bool, 256)
	for _, idx := range indices {
		used[idx] = true
	}
	for idx := 0; idx < 256; idx++ {
		if !used[idx] {
			continue
		}
		r, g, b := paletteRGB(idx)
		dst = append(dst, '#')
		dst = strconv.AppendInt(dst, int64(idx), 10)
		dst = append(dst, ";2;"...)
		dst = strconv.AppendInt(dst, int64(percent(r)), 10)
		dst = append(dst, ';')
		dst = strconv.AppendInt(dst, int64(percent(g)), 10)
		dst = append(dst, ';')
		dst = strconv.AppendInt(dst, int64(percent(b)), 10)
	}

	for band := 0; band*6 < height; band++ {
		bandUsed := make([]bool, 256)
		for x := 0; x < width; x++ {
			for bit := 0; bit < 6; bit++ {
				y := band*6 + bit
				if y >= height {
					break
				}
				bandUsed[indices[y*width+x]] = true
			}
		}
		for idx := 0; idx < 256; idx++ {
			if !bandUsed[idx] {
				continue
			}
			dst = append(dst, '#')
			dst = strconv.AppendInt(dst, int64(idx), 10)
			dst = appendSixelRun(dst, indices, width, height, band, uint8(idx))
			dst = append(dst, '$')
		}
		dst = append(dst, '-')
	}
	dst = append(dst, "\x1b\\"...)
	return dst
}

func appendSixelRun(dst []byte, indices []uint8, width, height, band int, color uint8) []byte {
	last := byte(0)
	run := 0
	flush := func() {
		if run == 0 {
			return
		}
		if run > 3 {
			dst = append(dst, '!')
			dst = strconv.AppendInt(dst, int64(run), 10)
			dst = append(dst, last)
		} else {
			for i := 0; i < run; i++ {
				dst = append(dst, last)
			}
		}
		run = 0
	}
	for x := 0; x < width; x++ {
		var bits byte
		for bit := 0; bit < 6; bit++ {
			y := band*6 + bit
			if y >= height {
				break
			}
			if indices[y*width+x] == color {
				bits |= 1 << uint(bit)
			}
		}
		ch := byte(0x3f) + bits
		if ch == last {
			run++
			continue
		}
		flush()
		last = ch
		run = 1
	}
	flush()
	return dst
}

func paletteIndex(r, g, b uint8) int {
	ri := quantize(r)
	gi := quantize(g)
	bi := quantize(b)
	return 16 + ri*36 + gi*6 + bi
}

func quantize(v uint8) int {
	return int((uint32(v)*uint32(levels-1) + 127) / 255)
}

func paletteRGB(idx int) (uint8, uint8, uint8) {
	i := idx - 16
	if i < 0 || i >= 216 {
		return 0, 0, 0
	}
	to8 := func(n int) uint8 { return uint8(n * 255 / (levels - 1)) }
	return to8(i / 36), to8((i / 6) % 6), to8(i % 6)
}

func percent(v uint8) int {
	return (int(v)*100 + 127) / 255
}

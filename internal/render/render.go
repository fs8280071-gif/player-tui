package render

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/fs8280071-gif/player-tui/internal/render/kitty"
	"github.com/fs8280071-gif/player-tui/internal/video"
)

type Mode int

const (
	ModeAuto Mode = iota
	ModeHalfblock
	ModeQuadrant
	ModeSextant
	ModeKitty
	ModeSixel
	ModeITerm2
)

func (m Mode) String() string {
	switch m {
	case ModeAuto:
		return "auto"
	case ModeHalfblock:
		return "halfblock"
	case ModeQuadrant:
		return "quadrant"
	case ModeSextant:
		return "sextant"
	case ModeKitty:
		return "kitty"
	case ModeSixel:
		return "sixel"
	case ModeITerm2:
		return "iterm2"
	default:
		return "unknown"
	}
}

func (m Mode) Graphics() bool {
	return m == ModeKitty || m == ModeSixel || m == ModeITerm2
}

func ParseMode(name string) (Mode, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "auto":
		return ModeAuto, nil
	case "halfblock", "half-block", "half":
		return ModeHalfblock, nil
	case "quadrant", "quad", "2x2":
		return ModeQuadrant, nil
	case "sextant", "sext", "2x3":
		return ModeSextant, nil
	case "kitty":
		return ModeKitty, nil
	case "sixel":
		return ModeSixel, nil
	case "iterm2", "iterm":
		return ModeITerm2, nil
	default:
		return ModeAuto, fmt.Errorf("unknown render mode %q (use kitty, sixel, iterm2, halfblock, quadrant, sextant, or auto)", name)
	}
}

func Resolve(m Mode) (Mode, string) {
	if m != ModeAuto {
		return m, "flag"
	}
	if value := strings.TrimSpace(os.Getenv("PLAYER_TUI_RENDER")); value != "" {
		if mode, err := ParseMode(value); err == nil && mode != ModeAuto {
			return mode, "environment"
		}
	}
	return AutoMode(os.Getenv), "auto"
}

func AutoMode(getenv func(string) string) Mode {
	term := strings.ToLower(strings.TrimSpace(getenv("TERM")))
	if term == "" || term == "dumb" || term == "linux" {
		return ModeHalfblock
	}
	if strings.EqualFold(getenv("PLAYER_TUI_ASCII"), "1") {
		return ModeHalfblock
	}
	return Detect(getenv).PreferredMode()
}

func (m Mode) pixelsPerCell() (int, int) {
	switch m {
	case ModeQuadrant:
		return 2, 2
	case ModeSextant:
		return 2, 3
	default:
		return 1, 2
	}
}

func (m Mode) glyph(mask uint8) rune {
	switch m {
	case ModeQuadrant:
		return quadrantGlyphs[int(mask)&0x0f]
	case ModeSextant:
		return sextantGlyphs[int(mask)&0x3f]
	default:
		return '\u2580'
	}
}

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
	if startRow < 1 {
		startRow = 1
	}
	startCol = (termW-cols)/2 + 1
	return cols, rows, startRow, startCol
}

type rgb struct {
	r uint8
	g uint8
	b uint8
}

type Renderer struct {
	mode      Mode
	dither    bool
	lastFg    int32
	lastBg    int32
	sampleBuf [6]rgb
	imageID   uint32
	frames    uint64
}

func New(mode Mode, dither bool) *Renderer {
	if mode == ModeAuto {
		mode, _ = Resolve(ModeAuto)
	}
	return &Renderer{mode: mode, dither: dither, lastFg: -1, lastBg: -1, imageID: 1001}
}

func (r *Renderer) Mode() Mode { return r.mode }

func (r *Renderer) Cleanup() []byte {
	if r.mode == ModeKitty && r.frames > 0 {
		return kitty.Delete(nil, r.imageID)
	}
	return nil
}

func (r *Renderer) Append(dst []byte, f *video.Frame, cols, rows, startRow, startCol int) []byte {
	dst = dst[:0]
	if f == nil || f.Width <= 0 || f.Height <= 0 || cols <= 0 || rows <= 0 {
		return dst
	}
	if r.mode.Graphics() {
		return r.appendGraphics(dst, f, cols, rows, startRow, startCol)
	}
	r.lastFg, r.lastBg = -1, -1
	pxW, pxH := r.mode.pixelsPerCell()
	gridX, gridY := cols*pxW, rows*pxH

	for row := 0; row < rows; row++ {
		dst = appendCursor(dst, startRow+row, startCol)
		for col := 0; col < cols; col++ {
			n := 0
			for sy := 0; sy < pxH; sy++ {
				y0, y1 := span(row*pxH+sy, gridY, f.Height)
				for sx := 0; sx < pxW; sx++ {
					x0, x1 := span(col*pxW+sx, gridX, f.Width)
					y, u, v := f.AverageYUV(x0, y0, x1, y1)
					cr, cg, cb := video.YUVToRGB(y, u, v)
					s := rgb{cr, cg, cb}
					if r.dither {
						s = ditherRGB(s, col*pxW+sx, row*pxH+sy)
					}
					r.sampleBuf[n] = s
					n++
				}
			}
			var fg, bg rgb
			var glyph rune
			if r.mode == ModeHalfblock {
				fg, bg = r.sampleBuf[0], r.sampleBuf[1]
				glyph = '\u2580'
			} else {
				var mask uint8
				fg, bg, mask = chooseTwo(r.sampleBuf[:n])
				glyph = r.mode.glyph(mask)
			}
			dst = r.appendColors(dst, fg, bg)
			dst = utf8.AppendRune(dst, glyph)
		}
	}
	dst = append(dst, "\x1b[0m"...)
	return dst
}

func span(i, parts, total int) (int, int) {
	if total <= 0 || parts <= 0 {
		return 0, 0
	}
	a := i * total / parts
	b := (i + 1) * total / parts
	if a >= total {
		a = total - 1
	}
	if b <= a {
		b = a + 1
	}
	if b > total {
		b = total
	}
	return a, b
}

func chooseTwo(samples []rgb) (rgb, rgb, uint8) {
	n := len(samples)
	if n == 0 {
		return rgb{}, rgb{}, 0
	}
	full := uint8((1 << uint(n)) - 1)
	ai, bi, best := 0, 0, -1
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			d := dist2(samples[i], samples[j])
			if d > best {
				best, ai, bi = d, i, j
			}
		}
	}
	if best <= 400 {
		average := averageMask(samples, full)
		return average, average, full
	}
	c0, c1 := samples[ai], samples[bi]
	m0, m1 := assignMask(samples, c0, c1)
	if m0 == 0 || m1 == 0 {
		average := averageMask(samples, full)
		return average, average, full
	}
	c0 = averageMask(samples, m0)
	c1 = averageMask(samples, m1)
	m0, m1 = assignMask(samples, c0, c1)
	if m0 == 0 || m1 == 0 {
		average := averageMask(samples, full)
		return average, average, full
	}
	return c0, c1, m0
}

func assignMask(samples []rgb, c0, c1 rgb) (uint8, uint8) {
	var m0, m1 uint8
	for i, s := range samples {
		if dist2(s, c0) <= dist2(s, c1) {
			m0 |= 1 << uint(i)
		} else {
			m1 |= 1 << uint(i)
		}
	}
	return m0, m1
}

func averageMask(samples []rgb, mask uint8) rgb {
	var sr, sg, sb, n int
	for i, s := range samples {
		if mask&(1<<uint(i)) == 0 {
			continue
		}
		sr += int(s.r)
		sg += int(s.g)
		sb += int(s.b)
		n++
	}
	if n == 0 {
		return rgb{}
	}
	return rgb{
		r: uint8((sr + n/2) / n),
		g: uint8((sg + n/2) / n),
		b: uint8((sb + n/2) / n),
	}
}

func dist2(a, b rgb) int {
	dr := int(a.r) - int(b.r)
	dg := int(a.g) - int(b.g)
	db := int(a.b) - int(b.b)
	return dr*dr + dg*dg + db*db
}

var bayer4 = [4][4]int{
	{0, 8, 2, 10},
	{12, 4, 14, 6},
	{3, 11, 1, 9},
	{15, 7, 13, 5},
}

func ditherRGB(c rgb, x, y int) rgb {
	offset := bayer4[y&3][x&3] - 8
	return rgb{
		r: clamp8(int(c.r) + offset),
		g: clamp8(int(c.g) + offset),
		b: clamp8(int(c.b) + offset),
	}
}

func clamp8(v int) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

func (r *Renderer) appendColors(dst []byte, fg, bg rgb) []byte {
	fgKey := int32(fg.r)<<16 | int32(fg.g)<<8 | int32(fg.b)
	bgKey := int32(bg.r)<<16 | int32(bg.g)<<8 | int32(bg.b)
	if fgKey != r.lastFg {
		dst = append(dst, "\x1b[38;2;"...)
		dst = strconv.AppendInt(dst, int64(fg.r), 10)
		dst = append(dst, ';')
		dst = strconv.AppendInt(dst, int64(fg.g), 10)
		dst = append(dst, ';')
		dst = strconv.AppendInt(dst, int64(fg.b), 10)
		dst = append(dst, 'm')
		r.lastFg = fgKey
	}
	if bgKey != r.lastBg {
		dst = append(dst, "\x1b[48;2;"...)
		dst = strconv.AppendInt(dst, int64(bg.r), 10)
		dst = append(dst, ';')
		dst = strconv.AppendInt(dst, int64(bg.g), 10)
		dst = append(dst, ';')
		dst = strconv.AppendInt(dst, int64(bg.b), 10)
		dst = append(dst, 'm')
		r.lastBg = bgKey
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

var quadrantGlyphs = [16]rune{
	' ',
	'\u2598',
	'\u259d',
	'\u2580',
	'\u2596',
	'\u258c',
	'\u259e',
	'\u259b',
	'\u2597',
	'\u259a',
	'\u2590',
	'\u259c',
	'\u2584',
	'\u2599',
	'\u259f',
	'\u2588',
}

var sextantGlyphs = [64]rune{
	' ',
	'🬀',
	'🬁',
	'🬂',
	'🬃',
	'🬄',
	'🬅',
	'🬆',
	'🬇',
	'🬈',
	'🬉',
	'🬊',
	'🬋',
	'🬌',
	'🬍',
	'🬎',
	'🬏',
	'🬐',
	'🬑',
	'🬒',
	'🬓',
	'▌',
	'🬔',
	'🬕',
	'🬖',
	'🬗',
	'🬘',
	'🬙',
	'🬚',
	'🬛',
	'🬜',
	'🬝',
	'🬞',
	'🬟',
	'🬠',
	'🬡',
	'🬢',
	'🬣',
	'🬤',
	'🬥',
	'🬦',
	'🬧',
	'▐',
	'🬨',
	'🬩',
	'🬪',
	'🬫',
	'🬬',
	'🬭',
	'🬮',
	'🬯',
	'🬰',
	'🬱',
	'🬲',
	'🬳',
	'🬴',
	'🬵',
	'🬶',
	'🬷',
	'🬸',
	'🬹',
	'🬺',
	'🬻',
	'█',
}

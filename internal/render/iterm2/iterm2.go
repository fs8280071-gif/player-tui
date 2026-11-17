// Package iterm2 encodes raw RGB frames for the iTerm2 inline image protocol.
package iterm2

import (
	"encoding/base64"
	"strconv"
)

type Options struct {
	Cols int
	Rows int
}

func Encode(dst []byte, width, height int, rgb []byte, opts Options) []byte {
	if width <= 0 || height <= 0 || len(rgb) != width*height*3 {
		return dst
	}
	payloadLen := base64.StdEncoding.EncodedLen(len(rgb))
	dst = append(dst, "\x1b]1337;File=inline=1;size="...)
	dst = strconv.AppendInt(dst, int64(len(rgb)), 10)
	if opts.Cols > 0 {
		dst = append(dst, ";width="...)
		dst = strconv.AppendInt(dst, int64(opts.Cols), 10)
	}
	if opts.Rows > 0 {
		dst = append(dst, ";height="...)
		dst = strconv.AppendInt(dst, int64(opts.Rows), 10)
	}
	dst = append(dst, ";preserveAspectRatio=1:"...)
	start := len(dst)
	dst = append(dst, make([]byte, payloadLen)...)
	base64.StdEncoding.Encode(dst[start:], rgb)
	dst = append(dst, '\x07')
	return dst
}

// Package kitty encodes raw RGB frames for the Kitty graphics protocol.
package kitty

import (
	"encoding/base64"
	"strconv"
)

const DefaultChunkSize = 4096

type Options struct {
	ImageID   uint32
	Cols      int
	Rows      int
	ChunkSize int
}

func Encode(dst []byte, width, height int, rgb []byte, opts Options) []byte {
	if width <= 0 || height <= 0 || len(rgb) != width*height*3 {
		return dst
	}
	if opts.ChunkSize <= 0 {
		opts.ChunkSize = DefaultChunkSize
	}
	payload := make([]byte, base64.StdEncoding.EncodedLen(len(rgb)))
	base64.StdEncoding.Encode(payload, rgb)

	dst = append(dst, "\x1b_G"...)
	dst = append(dst, "a=T,f=24,s="...)
	dst = strconv.AppendInt(dst, int64(width), 10)
	dst = append(dst, ",v="...)
	dst = strconv.AppendInt(dst, int64(height), 10)
	if opts.ImageID != 0 {
		dst = append(dst, ",i="...)
		dst = strconv.AppendUint(dst, uint64(opts.ImageID), 10)
	}
	if opts.Cols > 0 && opts.Rows > 0 {
		dst = append(dst, ",c="...)
		dst = strconv.AppendInt(dst, int64(opts.Cols), 10)
		dst = append(dst, ",r="...)
		dst = strconv.AppendInt(dst, int64(opts.Rows), 10)
	}
	dst = append(dst, ",q=2"...)

	for offset := 0; offset < len(payload); offset += opts.ChunkSize {
		end := offset + opts.ChunkSize
		if end > len(payload) {
			end = len(payload)
		}
		last := end == len(payload)
		if offset == 0 {
			if last {
				dst = append(dst, ';')
			} else {
				dst = append(dst, ",m=1;"...)
			}
		} else {
			dst = append(dst, "\x1b_G"...)
			if last {
				dst = append(dst, "m=0;"...)
			} else {
				dst = append(dst, "m=1;"...)
			}
		}
		dst = append(dst, payload[offset:end]...)
		dst = append(dst, "\x1b\\"...)
	}
	return dst
}

func Delete(dst []byte, imageID uint32) []byte {
	if imageID == 0 {
		return dst
	}
	dst = append(dst, "\x1b_Ga=d,d=I,i="...)
	dst = strconv.AppendUint(dst, uint64(imageID), 10)
	dst = append(dst, "\x1b\\"...)
	return dst
}

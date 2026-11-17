package libvlc

/*
#cgo pkg-config: libvlc
#include <stdlib.h>
#include "shim.h"
*/
import "C"

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/fs8280071-gif/player-tui/internal/engine"
	"github.com/fs8280071-gif/player-tui/internal/video"
)

type Player struct {
	inst *C.libvlc_instance_t
	mp   *C.libvlc_media_player_t
	ctx  *C.pt_vlc_ctx

	mu      sync.Mutex
	lastSeq uint64
	frame   video.Frame

	visibleW int
	visibleH int
	paddedW  int
	paddedH  int
}

func New() (*Player, error) {
	var args []string
	if os.Getenv("PLAYER_TUI_DUMMY_AUDIO") == "1" {
		args = append(args, "--aout=dummy")
	}
	if os.Getenv("PLAYER_TUI_VLC_VERBOSE") != "1" {
		args = append(args, "--quiet")
	}

	cargs := make([]*C.char, len(args))
	for i, a := range args {
		cargs[i] = C.CString(a)
	}
	defer func() {
		for _, a := range cargs {
			C.free(unsafe.Pointer(a))
		}
	}()

	var argv **C.char
	if len(cargs) > 0 {
		argv = (**C.char)(unsafe.Pointer(&cargs[0]))
	}
	inst := C.libvlc_new(C.int(len(cargs)), argv)
	if inst == nil {
		return nil, fmt.Errorf("libVLC init failed: %s", vlcErr())
	}
	return &Player{inst: inst}, nil
}

func (p *Player) Open(_ context.Context, spec engine.MediaSpec) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.inst == nil {
		return errors.New("libVLC instance is closed")
	}
	p.closeLocked()

	target := C.CString(spec.Target)
	defer C.free(unsafe.Pointer(target))

	var media *C.libvlc_media_t
	if isURL(spec.Target) {
		media = C.libvlc_media_new_location(p.inst, target)
	} else {
		media = C.libvlc_media_new_path(p.inst, target)
	}
	if media == nil {
		return fmt.Errorf("cannot open %q: %s", spec.Target, vlcErr())
	}
	defer C.libvlc_media_release(media)

	mp := C.libvlc_media_player_new_from_media(media)
	if mp == nil {
		return fmt.Errorf("cannot create player: %s", vlcErr())
	}
	ctx := C.pt_ctx_new(C.uintptr_t(0))
	if ctx == nil {
		C.libvlc_media_player_release(mp)
		return errors.New("cannot allocate video buffer context")
	}
	if rc := C.pt_ctx_attach(ctx, mp); rc != 0 {
		C.pt_ctx_free(ctx)
		C.libvlc_media_player_release(mp)
		return errors.New("cannot attach libVLC video callbacks")
	}

	if spec.SubtitleFile != "" {
		sub := C.CString(spec.SubtitleFile)
		rc := C.pt_add_subtitle(mp, sub)
		C.free(unsafe.Pointer(sub))
		if rc != 0 {
			C.pt_ctx_detach(ctx, mp)
			C.pt_ctx_free(ctx)
			C.libvlc_media_player_release(mp)
			return fmt.Errorf("cannot attach subtitle %q: %s", spec.SubtitleFile, vlcErr())
		}
	}

	p.mp = mp
	p.ctx = ctx
	p.lastSeq = 0
	p.frame = video.Frame{}
	p.visibleW, p.visibleH = 0, 0
	p.paddedW, p.paddedH = 0, 0

	if C.libvlc_media_player_play(mp) != 0 {
		p.closeLocked()
		return fmt.Errorf("playback failed: %s", vlcErr())
	}
	return nil
}

func (p *Player) Play() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.mp == nil {
		return errors.New("nothing is open")
	}
	if p.stateLocked() == engine.StatePaused {
		C.libvlc_media_player_set_pause(p.mp, 0)
		return nil
	}
	if C.libvlc_media_player_play(p.mp) != 0 {
		return fmt.Errorf("playback failed: %s", vlcErr())
	}
	return nil
}

func (p *Player) Pause() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.mp == nil {
		return errors.New("nothing is open")
	}
	C.libvlc_media_player_set_pause(p.mp, 1)
	return nil
}

func (p *Player) Toggle() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.mp == nil {
		return errors.New("nothing is open")
	}
	switch p.stateLocked() {
	case engine.StatePlaying, engine.StateBuffering, engine.StateOpening:
		C.libvlc_media_player_set_pause(p.mp, 1)
	case engine.StatePaused:
		C.libvlc_media_player_set_pause(p.mp, 0)
	}
	return nil
}

func (p *Player) Seek(pos time.Duration) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.mp == nil {
		return errors.New("nothing is open")
	}
	if pos < 0 {
		pos = 0
	}
	if total := p.durationLocked(); total > 0 && pos > total {
		pos = total
	}
	C.libvlc_media_player_set_time(p.mp, C.libvlc_time_t(pos/time.Millisecond))
	return nil
}

func (p *Player) SeekBy(delta time.Duration) error {
	p.mu.Lock()
	pos := p.timeLocked() + delta
	var total time.Duration
	if p.mp != nil {
		total = p.durationLocked()
	}
	p.mu.Unlock()
	if pos < 0 {
		pos = 0
	}
	if total > 0 && pos > total {
		pos = total
	}
	return p.Seek(pos)
}

func (p *Player) Snapshot() engine.Snapshot {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.mp == nil {
		return engine.Snapshot{State: engine.StateIdle}
	}
	snap := engine.Snapshot{
		State:    p.stateLocked(),
		Time:     p.timeLocked(),
		Duration: p.durationLocked(),
		Seekable: C.libvlc_media_player_is_seekable(p.mp) == 1,
	}
	if snap.State == engine.StateError {
		snap.Error = vlcErr()
	}
	return snap
}

// LatestFrame returns the newest decoded frame, or nil if no new frame arrived.
// The returned frame is reused by the next call; the caller must finish drawing
// before calling LatestFrame again.
func (p *Player) LatestFrame() *video.Frame {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.mp == nil || p.ctx == nil {
		return nil
	}
	seq := uint64(C.pt_ctx_seq(p.ctx))
	if seq == p.lastSeq {
		return nil
	}

	mediaTime := p.timeLocked()
	C.pt_ctx_lock(p.ctx)
	defer C.pt_ctx_unlock(p.ctx)

	var w, h, pitchY, pitchU, pitchV C.uint
	base := C.pt_ctx_data(p.ctx, &w, &h, &pitchY, &pitchU, &pitchV)
	if base == nil || w == 0 || h == 0 {
		return nil
	}
	fullW, fullH := int(w), int(h)
	if p.paddedW != fullW || p.paddedH != fullH {
		p.visibleW, p.visibleH = 0, 0
		p.paddedW, p.paddedH = fullW, fullH
	}
	width, height := fullW, fullH
	if p.visibleW == 0 || p.visibleH == 0 {
		var vw, vh C.uint
		if C.libvlc_video_get_size(p.mp, 0, &vw, &vh) == 0 && vw > 0 && vh > 0 {
			p.visibleW, p.visibleH = int(vw), int(vh)
		}
	}
	if p.visibleW > 0 && p.visibleW < width {
		width = p.visibleW
	}
	if p.visibleH > 0 && p.visibleH < height {
		height = p.visibleH
	}
	py, pu, pv := int(pitchY), int(pitchU), int(pitchV)
	ySize := py * fullH
	uvHeight := (fullH + 1) / 2
	uSize := pu * uvHeight
	vSize := pv * uvHeight

	ensureLen(&p.frame.Y, ySize)
	ensureLen(&p.frame.U, uSize)
	ensureLen(&p.frame.V, vSize)
	copyPlane(p.frame.Y, base, py, fullW, fullH)
	uBase := unsafe.Add(unsafe.Pointer(base), ySize)
	vBase := unsafe.Add(uBase, uSize)
	copyPlane(p.frame.U, (*C.uint8_t)(uBase), pu, (fullW+1)/2, uvHeight)
	copyPlane(p.frame.V, (*C.uint8_t)(vBase), pv, (fullW+1)/2, uvHeight)

	p.frame.Width = width
	p.frame.Height = height
	p.frame.StrideY = py
	p.frame.StrideU = pu
	p.frame.StrideV = pv
	p.frame.MediaTime = mediaTime
	p.frame.Arrival = time.Now()
	p.frame.Sequence = seq
	p.lastSeq = seq
	return &p.frame
}

func (p *Player) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.closeLocked()
	if p.inst != nil {
		C.libvlc_release(p.inst)
		p.inst = nil
	}
	return nil
}

func (p *Player) closeLocked() {
	if p.mp != nil {
		C.libvlc_media_player_stop(p.mp)
		if p.ctx != nil {
			C.pt_ctx_detach(p.ctx, p.mp)
		}
		C.libvlc_media_player_release(p.mp)
		p.mp = nil
	}
	if p.ctx != nil {
		C.pt_ctx_free(p.ctx)
		p.ctx = nil
	}
	p.lastSeq = 0
	p.visibleW, p.visibleH = 0, 0
	p.paddedW, p.paddedH = 0, 0
}

func (p *Player) stateLocked() engine.State {
	if p.mp == nil {
		return engine.StateIdle
	}
	switch C.libvlc_media_player_get_state(p.mp) {
	case C.libvlc_NothingSpecial:
		return engine.StateIdle
	case C.libvlc_Opening:
		return engine.StateOpening
	case C.libvlc_Buffering:
		return engine.StateBuffering
	case C.libvlc_Playing:
		return engine.StatePlaying
	case C.libvlc_Paused:
		return engine.StatePaused
	case C.libvlc_Ended:
		return engine.StateEnded
	case C.libvlc_Error:
		return engine.StateError
	default:
		return engine.StateIdle
	}
}

func (p *Player) timeLocked() time.Duration {
	if p.mp == nil {
		return 0
	}
	t := C.libvlc_media_player_get_time(p.mp)
	if t < 0 {
		return 0
	}
	return time.Duration(t) * time.Millisecond
}

func (p *Player) durationLocked() time.Duration {
	if p.mp == nil {
		return 0
	}
	t := C.libvlc_media_player_get_length(p.mp)
	if t < 0 {
		return 0
	}
	return time.Duration(t) * time.Millisecond
}

func ensureLen(b *[]byte, n int) {
	if cap(*b) < n {
		*b = make([]byte, n)
		return
	}
	*b = (*b)[:n]
}

func copyPlane(dst []byte, src *C.uint8_t, stride, width, height int) {
	row := unsafe.Slice((*byte)(unsafe.Pointer(src)), stride*height)
	for y := 0; y < height; y++ {
		copy(dst[y*stride:y*stride+width], row[y*stride:y*stride+width])
	}
}

func vlcErr() string {
	msg := C.libvlc_errmsg()
	if msg == nil {
		return "unknown libVLC error"
	}
	return C.GoString(msg)
}

func isURL(target string) bool {
	lower := strings.ToLower(target)
	for _, scheme := range []string{"http://", "https://", "rtsp://", "rtmp://", "srt://", "file://"} {
		if strings.HasPrefix(lower, scheme) {
			return true
		}
	}
	return false
}

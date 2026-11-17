package libvlc

/*
#include "shim.h"
*/
import "C"

import (
	"errors"
	"fmt"
)

func (p *Player) Rate() float64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.mp == nil {
		return 0
	}
	return float64(C.libvlc_media_player_get_rate(p.mp))
}

func (p *Player) SetRate(rate float64) error {
	if rate < 0.25 {
		rate = 0.25
	}
	if rate > 4 {
		rate = 4
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.mp == nil {
		return errors.New("nothing is open")
	}
	if C.libvlc_media_player_set_rate(p.mp, C.float(rate)) != 0 {
		return fmt.Errorf("cannot set playback rate: %s", vlcErr())
	}
	return nil
}

func (p *Player) CycleSubtitle() (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.mp == nil {
		return "", errors.New("nothing is open")
	}
	desc := C.libvlc_video_get_spu_description(p.mp)
	if desc == nil {
		return "", errors.New("no subtitle tracks")
	}
	defer C.libvlc_track_description_list_release(desc)
	id, name, ok := nextTrack(desc, C.libvlc_video_get_spu(p.mp))
	if !ok {
		return "", errors.New("no subtitle tracks")
	}
	if C.libvlc_video_set_spu(p.mp, id) != 0 {
		return "", fmt.Errorf("cannot select subtitle track: %s", vlcErr())
	}
	return name, nil
}

func (p *Player) CycleAudio() (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.mp == nil {
		return "", errors.New("nothing is open")
	}
	desc := C.libvlc_audio_get_track_description(p.mp)
	if desc == nil {
		return "", errors.New("no audio tracks")
	}
	defer C.libvlc_track_description_list_release(desc)
	id, name, ok := nextTrack(desc, C.libvlc_audio_get_track(p.mp))
	if !ok {
		return "", errors.New("no audio tracks")
	}
	if C.libvlc_audio_set_track(p.mp, id) != 0 {
		return "", fmt.Errorf("cannot select audio track: %s", vlcErr())
	}
	return name, nil
}

func nextTrack(desc *C.libvlc_track_description_t, current C.int) (C.int, string, bool) {
	type entry struct {
		id   C.int
		name string
	}
	var entries []entry
	seen := 0
	for d := desc; d != nil && seen < 64; d = d.p_next {
		entries = append(entries, entry{id: d.i_id, name: C.GoString(d.psz_name)})
		seen++
	}
	if len(entries) == 0 {
		return 0, "", false
	}
	index := 0
	for i, e := range entries {
		if e.id == current {
			index = i
			break
		}
	}
	next := entries[(index+1)%len(entries)]
	return next.id, next.name, true
}

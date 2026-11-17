package engine

import (
	"context"
	"time"

	"github.com/fs8280071-gif/player-tui/internal/video"
)

type State int

const (
	StateIdle State = iota
	StateOpening
	StateBuffering
	StatePlaying
	StatePaused
	StateEnded
	StateError
)

func (s State) String() string {
	switch s {
	case StateIdle:
		return "idle"
	case StateOpening:
		return "opening"
	case StateBuffering:
		return "buffering"
	case StatePlaying:
		return "playing"
	case StatePaused:
		return "paused"
	case StateEnded:
		return "ended"
	case StateError:
		return "error"
	default:
		return "unknown"
	}
}

type MediaSpec struct {
	Target       string
	Title        string
	SubtitleFile string
	Start        time.Duration
}

type Snapshot struct {
	State    State
	Time     time.Duration
	Duration time.Duration
	Seekable bool
	Volume   int
	Muted    bool
	Rate     float64
	Error    string
}

type Player interface {
	Open(context.Context, MediaSpec) error
	Play() error
	Pause() error
	Toggle() error
	Seek(time.Duration) error
	SeekBy(time.Duration) error
	Volume() int
	SetVolume(int) error
	Muted() bool
	SetMuted(bool) error
	Rate() float64
	SetRate(float64) error
	CycleSubtitle() (string, error)
	CycleAudio() (string, error)
	Snapshot() Snapshot
	LatestFrame() *video.Frame
	Close() error
}

package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/fs8280071-gif/player-tui/internal/engine"
	"github.com/fs8280071-gif/player-tui/internal/engine/libvlc"
	"github.com/fs8280071-gif/player-tui/internal/render"
)

type BenchOptions struct {
	Spec   engine.MediaSpec
	Render render.Mode
	Dither bool
	Frames int
	Cols   int
	Rows   int
}

type BenchStats struct {
	Mode          render.Mode
	Width         int
	Height        int
	Cols          int
	Rows          int
	Frames        int
	Wall          time.Duration
	RenderAverage time.Duration
	RenderP50     time.Duration
	RenderP95     time.Duration
	DisplayFPS    float64
	RenderFPS     float64
	Bytes         int64
}

func Bench(ctx context.Context, opts BenchOptions) (BenchStats, error) {
	mode, _ := render.Resolve(opts.Render)
	cols, rows := opts.Cols, opts.Rows
	if cols <= 0 {
		cols = 120
	}
	if rows <= 0 {
		rows = 40
	}
	_ = os.Setenv("PLAYER_TUI_DUMMY_AUDIO", "1")

	player, err := libvlc.New()
	if err != nil {
		return BenchStats{}, &Error{Code: 4, Err: err}
	}
	defer player.Close()

	if err := player.Open(ctx, opts.Spec); err != nil {
		return BenchStats{}, &Error{Code: 3, Err: err}
	}

	stats := BenchStats{Mode: mode}
	renderer := render.New(mode, opts.Dither)
	var frameBuf []byte
	var renderTotal time.Duration
	renderTimes := make([]time.Duration, 0, 256)
	var first, last time.Time
	start := time.Now()
	startApplied := opts.Spec.Start <= 0
	deadline := start.Add(2 * time.Minute)

	for {
		if err := ctx.Err(); err != nil {
			return stats, err
		}
		if time.Now().After(deadline) {
			return stats, fmt.Errorf("benchmark timed out after %d frames", stats.Frames)
		}

		snap := player.Snapshot()
		if snap.State == engine.StateError {
			msg := snap.Error
			if msg == "" {
				msg = "playback failed"
			}
			return stats, &Error{Code: 3, Err: errors.New(msg)}
		}
		if !startApplied && snap.State == engine.StatePlaying {
			if snap.Seekable {
				if err := player.Seek(opts.Spec.Start); err == nil {
					startApplied = true
				}
			} else {
				startApplied = true
			}
		}

		if f := player.LatestFrame(); f != nil {
			if stats.Width == 0 {
				stats.Width, stats.Height = f.Width, f.Height
				stats.Cols, stats.Rows, _, _ = render.Fit(cols, rows, float64(f.Width)/float64(f.Height))
			}
			t0 := time.Now()
			frameBuf = renderer.Append(frameBuf, f, stats.Cols, stats.Rows, 1, 1)
			d := time.Since(t0)
			renderTotal += d
			renderTimes = append(renderTimes, d)
			stats.Frames++
			stats.Bytes += int64(len(frameBuf))
			now := time.Now()
			if first.IsZero() {
				first = now
			}
			last = now
			if opts.Frames > 0 && stats.Frames >= opts.Frames {
				break
			}
		}

		if snap.State == engine.StateEnded {
			if stats.Frames == 0 {
				return stats, &Error{Code: 3, Err: errors.New("media ended without decoded frames")}
			}
			break
		}
		if stats.Frames == 0 && time.Since(start) > 10*time.Second {
			return stats, &Error{Code: 3, Err: errors.New("timed out waiting for the first video frame")}
		}
		if stats.Frames > 0 && time.Since(last) > 5*time.Second {
			break
		}
		time.Sleep(2 * time.Millisecond)
	}

	if stats.Frames == 0 {
		return stats, &Error{Code: 3, Err: errors.New("no video frames decoded")}
	}
	stats.Wall = last.Sub(first)
	if stats.Frames > 1 && stats.Wall > 0 {
		stats.DisplayFPS = float64(stats.Frames-1) / stats.Wall.Seconds()
	}
	if renderTotal > 0 {
		stats.RenderAverage = renderTotal / time.Duration(stats.Frames)
		sort.Slice(renderTimes, func(i, j int) bool { return renderTimes[i] < renderTimes[j] })
		stats.RenderP50 = renderTimes[len(renderTimes)/2]
		stats.RenderP95 = renderTimes[len(renderTimes)*95/100]
		stats.RenderFPS = float64(time.Second) / float64(stats.RenderP50)
	}
	return stats, nil
}

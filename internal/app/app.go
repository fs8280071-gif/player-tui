package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"unicode/utf8"

	"github.com/fs8280071-gif/player-tui/internal/engine"
	"github.com/fs8280071-gif/player-tui/internal/engine/libvlc"
	"github.com/fs8280071-gif/player-tui/internal/render"
	"github.com/fs8280071-gif/player-tui/internal/term"
	"github.com/fs8280071-gif/player-tui/internal/video"
)

type Error struct {
	Code int
	Err  error
}

func (e *Error) Error() string { return e.Err.Error() }
func (e *Error) Unwrap() error { return e.Err }

type Options struct {
	Spec   engine.MediaSpec
	Specs  []engine.MediaSpec
	Render render.Mode
	Dither bool
	Audio  bool
}

func Run(ctx context.Context, opts Options) error {
	specs := opts.Specs
	if len(specs) == 0 {
		specs = []engine.MediaSpec{opts.Spec}
	}
	index := 0
	spec := specs[index]
	player, err := libvlc.New(opts.Audio)
	if err != nil {
		return &Error{Code: 4, Err: err}
	}
	defer player.Close()

	t, err := term.Open()
	if err != nil {
		return &Error{Code: 4, Err: err}
	}
	defer t.Close()

	if err := player.Open(ctx, spec); err != nil {
		return &Error{Code: 3, Err: err}
	}

	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	keys := make(chan term.Key, 16)
	go t.ReadKeys(runCtx, keys)

	startApplied := spec.Start <= 0
	startWait := time.Now()
	var frameBuf, statusBuf []byte
	var lastFrame *video.Frame
	lastW, lastH := 0, 0
	lastState := engine.StateIdle
	statusDirty := true
	openAt := func(i int) error {
		if i < 0 || i >= len(specs) {
			return nil
		}
		if err := player.Open(runCtx, specs[i]); err != nil {
			return &Error{Code: 3, Err: err}
		}
		index = i
		spec = specs[i]
		startApplied = spec.Start <= 0
		startWait = time.Now()
		lastFrame = nil
		lastState = engine.StateIdle
		statusDirty = true
		t.WriteString("\x1b[2J")
		return nil
	}

	sigCtx, stopSignals := signal.NotifyContext(runCtx, os.Interrupt, syscall.SIGTERM)
	defer stopSignals()

	mode, _ := render.Resolve(opts.Render)
	renderer := render.New(mode, opts.Dither)
	defer func() {
		if seq := renderer.Cleanup(); len(seq) > 0 {
			t.WriteString(string(seq))
		}
	}()
	noColor := os.Getenv("NO_COLOR") != ""

	ticker := time.NewTicker(16 * time.Millisecond)
	defer ticker.Stop()
	statusTicker := time.NewTicker(250 * time.Millisecond)
	defer statusTicker.Stop()

	for {
		select {
		case <-sigCtx.Done():
			return nil
		case k := <-keys:
			switch k {
			case term.KeyQuit:
				return nil
			case term.KeySpace:
				if err := player.Toggle(); err != nil {
					statusDirty = true
				}
				statusDirty = true
			case term.KeyLeft:
				_ = player.SeekBy(-5 * time.Second)
				statusDirty = true
			case term.KeyRight:
				_ = player.SeekBy(5 * time.Second)
				statusDirty = true
			case term.KeyNext:
				if index+1 < len(specs) {
					if err := openAt(index + 1); err != nil {
						return err
					}
				}
			case term.KeyPrev:
				if index > 0 {
					if err := openAt(index - 1); err != nil {
						return err
					}
				}
			}
		case <-statusTicker.C:
			statusDirty = true
		case <-ticker.C:
		}

		snap := player.Snapshot()
		if snap.State == engine.StateError {
			msg := snap.Error
			if msg == "" {
				msg = "playback failed"
			}
			return &Error{Code: 3, Err: fmt.Errorf("%s (%s)", msg, spec.Title)}
		}
		if snap.State != lastState {
			lastState = snap.State
			statusDirty = true
		}

		if !startApplied && snap.State == engine.StatePlaying {
			if snap.Seekable {
				if err := player.Seek(spec.Start); err == nil {
					startApplied = true
				}
			} else if time.Since(startWait) > 2*time.Second {
				startApplied = true
			}
		}

		if f := player.LatestFrame(); f != nil {
			lastFrame = f
			w, h := t.Size()
			if w != lastW || h != lastH {
				lastW, lastH = w, h
				t.WriteString("\x1b[2J")
				statusDirty = true
			}
			aspect := float64(f.Width) / float64(f.Height)
			cols, rows, startRow, startCol := render.Fit(w, h, aspect)
			frameBuf = renderer.Append(frameBuf, f, cols, rows, startRow, startCol)
			t.WriteString(string(frameBuf))
		}

		w, h := t.Size()
		if w != lastW || h != lastH {
			lastW, lastH = w, h
			t.WriteString("\x1b[2J")
			statusDirty = true
			if lastFrame != nil {
				aspect := float64(lastFrame.Width) / float64(lastFrame.Height)
				cols, rows, startRow, startCol := render.Fit(w, h, aspect)
				frameBuf = renderer.Append(frameBuf, lastFrame, cols, rows, startRow, startCol)
				t.WriteString(string(frameBuf))
			}
		}

		if statusDirty {
			statusDirty = false
			title := spec.Title
			if len(specs) > 1 {
				title = fmt.Sprintf("[%d/%d] %s", index+1, len(specs), title)
			}
			statusBuf = buildStatus(statusBuf, title, snap, w, h, noColor, mode.String())
			t.WriteString(string(statusBuf))
		}
	}
}

func buildStatus(dst []byte, title string, snap engine.Snapshot, width, height int, noColor bool, mode string) []byte {
	dst = dst[:0]
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	vol := fmt.Sprintf("%d%%", snap.Volume)
	if snap.Muted {
		vol = "muted"
	}
	left := fmt.Sprintf(" %s  %s / %s  %s  vol %s  %s", title, formatDuration(snap.Time), formatDuration(snap.Duration), snap.State, vol, mode)
	right := "space pause  <- -> seek  q quit "
	line := left
	pad := width - utf8.RuneCountInString(left) - utf8.RuneCountInString(right)
	switch {
	case pad >= 1:
		line = left + strings.Repeat(" ", pad) + right
	case width > utf8.RuneCountInString(right):
		maxLeft := width - utf8.RuneCountInString(right)
		line = truncate(left, maxLeft) + right
	default:
		line = truncate(left, width)
	}
	line = truncate(line, width)
	line += strings.Repeat(" ", max(0, width-utf8.RuneCountInString(line)))

	dst = append(dst, "\x1b["...)
	dst = appendInt(dst, height)
	dst = append(dst, ";1H"...)
	dst = append(dst, "\x1b[0m"...)
	if !noColor {
		dst = append(dst, "\x1b[7m"...)
	}
	dst = append(dst, line...)
	dst = append(dst, "\x1b[0m"...)
	return dst
}

func formatDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	total := int64(d / time.Second)
	h := total / 3600
	m := (total % 3600) / 60
	s := total % 60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%02d:%02d", m, s)
}

func truncate(s string, n int) string {
	if n <= 0 {
		return ""
	}
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	runes := []rune(s)
	return string(runes[:n])
}

func appendInt(dst []byte, n int) []byte {
	if n == 0 {
		return append(dst, '0')
	}
	if n < 0 {
		dst = append(dst, '-')
		n = -n
	}
	var tmp [20]byte
	i := len(tmp)
	for n > 0 {
		i--
		tmp[i] = byte('0' + n%10)
		n /= 10
	}
	return append(dst, tmp[i:]...)
}

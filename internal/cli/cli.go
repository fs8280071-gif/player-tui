package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/fs8280071-gif/player-tui/internal/engine"
	"github.com/fs8280071-gif/player-tui/internal/render"
)

type Command int

const (
	CommandPlay Command = iota
	CommandBench
	CommandDoctor
)

type Options struct {
	Command Command
	Spec    engine.MediaSpec
	Render  render.Mode
	Dither  bool
	Frames  int
	Cols    int
	Rows    int
	Audio   bool
	Help    bool
}

type TargetError struct {
	Err error
}

func (e *TargetError) Error() string { return e.Err.Error() }
func (e *TargetError) Unwrap() error { return e.Err }

const Usage = `player-tui - play media in the terminal

Usage:
  player-tui play <file-or-url> [options]
  player-tui open <file-or-url> [options]
  player-tui <file-or-url> [options]
  player-tui bench <file-or-url> [options]
  player-tui doctor

Options:
  --render <mode>       kitty, sixel, iterm2, halfblock, quadrant,
                        sextant, or auto (default: auto)
  --dither              ordered 4x4 Bayer dithering
  --title <text>        display title (does not rename the source)
  --sub-file <path>     attach an external subtitle file
  --start <time>        start position: SS, MM:SS, or HH:MM:SS
  --audio               enable audio output (default: off)
  --frames <n>          bench: stop after n frames (0 = until end)
  --cols <n>            bench: terminal width in cells (default 120)
  --rows <n>            bench: terminal height in cells (default 40)
  -h, --help            show this help

Keys:
  space                 play / pause
  left / right          seek -5s / +5s
  q                     quit
`

func Parse(args []string) (Options, error) {
	opts := Options{Render: render.ModeAuto, Cols: 120, Rows: 40}
	if len(args) == 0 {
		return opts, errors.New("missing media target (or run: player-tui)")
	}
	if args[0] == "doctor" {
		opts.Command = CommandDoctor
		for _, arg := range args[1:] {
			switch arg {
			case "-h", "--help":
				opts.Help = true
				return opts, nil
			default:
				return opts, fmt.Errorf("doctor takes no arguments (got %q)", arg)
			}
		}
		return opts, nil
	}

	var positionals []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-h" || arg == "--help":
			opts.Help = true
			return opts, nil
		case arg == "--":
			positionals = append(positionals, args[i+1:]...)
			i = len(args)
		case arg == "--title" || arg == "--sub-file" || arg == "--start" ||
			arg == "--render" || arg == "--frames" || arg == "--cols" || arg == "--rows":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("%s needs a value", arg)
			}
			i++
			switch arg {
			case "--title":
				opts.Spec.Title = args[i]
			case "--sub-file":
				opts.Spec.SubtitleFile = args[i]
			case "--start":
				d, err := ParseTime(args[i])
				if err != nil {
					return opts, err
				}
				opts.Spec.Start = d
			case "--render":
				mode, err := render.ParseMode(args[i])
				if err != nil {
					return opts, err
				}
				opts.Render = mode
			case "--frames":
				n, err := parseCount(arg, args[i], 0)
				if err != nil {
					return opts, err
				}
				opts.Frames = n
			case "--cols":
				n, err := parseCount(arg, args[i], 1)
				if err != nil {
					return opts, err
				}
				opts.Cols = n
			case "--rows":
				n, err := parseCount(arg, args[i], 1)
				if err != nil {
					return opts, err
				}
				opts.Rows = n
			}
		case arg == "--dither":
			opts.Dither = true
		case arg == "--audio":
			opts.Audio = true
		case strings.HasPrefix(arg, "--title="):
			opts.Spec.Title = strings.TrimPrefix(arg, "--title=")
		case strings.HasPrefix(arg, "--sub-file="):
			opts.Spec.SubtitleFile = strings.TrimPrefix(arg, "--sub-file=")
		case strings.HasPrefix(arg, "--start="):
			d, err := ParseTime(strings.TrimPrefix(arg, "--start="))
			if err != nil {
				return opts, err
			}
			opts.Spec.Start = d
		case strings.HasPrefix(arg, "--render="):
			mode, err := render.ParseMode(strings.TrimPrefix(arg, "--render="))
			if err != nil {
				return opts, err
			}
			opts.Render = mode
		case strings.HasPrefix(arg, "--frames="):
			n, err := parseCount("--frames", strings.TrimPrefix(arg, "--frames="), 0)
			if err != nil {
				return opts, err
			}
			opts.Frames = n
		case strings.HasPrefix(arg, "--cols="):
			n, err := parseCount("--cols", strings.TrimPrefix(arg, "--cols="), 1)
			if err != nil {
				return opts, err
			}
			opts.Cols = n
		case strings.HasPrefix(arg, "--rows="):
			n, err := parseCount("--rows", strings.TrimPrefix(arg, "--rows="), 1)
			if err != nil {
				return opts, err
			}
			opts.Rows = n
		case strings.HasPrefix(arg, "-"):
			return opts, fmt.Errorf("unknown option %q", arg)
		default:
			positionals = append(positionals, arg)
		}
	}

	if len(positionals) > 0 {
		switch positionals[0] {
		case "play", "open":
			positionals = positionals[1:]
		case "bench":
			opts.Command = CommandBench
			positionals = positionals[1:]
		}
	}
	if len(positionals) == 0 {
		return opts, errors.New("missing media target")
	}
	if len(positionals) > 1 {
		return opts, fmt.Errorf("unexpected extra argument %q", positionals[1])
	}

	target, err := normalizeTarget(positionals[0])
	if err != nil {
		return opts, err
	}
	opts.Spec.Target = target

	if opts.Spec.Title == "" {
		if isURL(target) {
			opts.Spec.Title = target
		} else {
			opts.Spec.Title = filepath.Base(target)
		}
	}
	opts.Spec.Title = sanitizeText(opts.Spec.Title)

	if opts.Spec.SubtitleFile != "" {
		sub, err := normalizeTarget(opts.Spec.SubtitleFile)
		if err != nil {
			return opts, fmt.Errorf("subtitle: %w", err)
		}
		opts.Spec.SubtitleFile = sub
	}
	return opts, nil
}

func parseCount(name, value string, minValue int) (int, error) {
	n, err := strconv.Atoi(value)
	if err != nil || n < minValue {
		return 0, fmt.Errorf("%s needs an integer >= %d", name, minValue)
	}
	return n, nil
}

func normalizeTarget(target string) (string, error) {
	if strings.TrimSpace(target) == "" {
		return "", &TargetError{Err: errors.New("empty media target")}
	}
	for _, r := range target {
		if r == 0 || (r < 0x20 && r != '\t') || r == 0x7f {
			return "", &TargetError{Err: errors.New("media target contains control characters")}
		}
	}
	if isURL(target) {
		return target, nil
	}
	abs, err := filepath.Abs(target)
	if err != nil {
		return "", err
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		abs = resolved
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", &TargetError{Err: fmt.Errorf("cannot read %q: %w", target, err)}
	}
	if info.IsDir() {
		return "", &TargetError{Err: fmt.Errorf("%q is a directory; pass a media file (or run player-tui to browse)", target)}
	}
	if !info.Mode().IsRegular() {
		return "", &TargetError{Err: fmt.Errorf("%q is not a regular file", target)}
	}
	if info.Size() == 0 {
		return "", &TargetError{Err: fmt.Errorf("%q is empty (corrupt or unfinished download?)", target)}
	}
	return abs, nil
}

func ParseTime(s string) (time.Duration, error) {
	if s == "" {
		return 0, errors.New("empty time value")
	}
	parts := strings.Split(s, ":")
	if len(parts) > 3 {
		return 0, fmt.Errorf("invalid time %q (use SS, MM:SS, or HH:MM:SS)", s)
	}
	var total int64
	for i, part := range parts {
		if part == "" {
			return 0, fmt.Errorf("invalid time %q", s)
		}
		n, err := strconv.ParseInt(part, 10, 64)
		if err != nil || n < 0 {
			return 0, fmt.Errorf("invalid time %q (use SS, MM:SS, or HH:MM:SS)", s)
		}
		if i > 0 && n > 59 {
			return 0, fmt.Errorf("invalid time %q: minutes and seconds must be 0-59", s)
		}
		total = total*60 + n
	}
	return time.Duration(total) * time.Second, nil
}

func sanitizeText(s string) string {
	s = strings.Map(func(r rune) rune {
		if r == '\t' {
			return ' '
		}
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, s)
	s = strings.TrimSpace(s)
	runes := []rune(s)
	if len(runes) > 200 {
		s = string(runes[:200])
	}
	if s == "" {
		return "player-tui"
	}
	return s
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

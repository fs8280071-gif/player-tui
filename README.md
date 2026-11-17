# player-tui

VLC-class terminal media player: local and streamed audio/video, subtitles,
library, terminal-native rendering, and optional offline AI.

## Status

The MVP playback slice works: `player-tui play <file>` opens local media with
libVLC, renders I420 video as truecolor blocks or terminal graphics, and supports
pause, seek, volume, mute, speed, track switching, a queue, resume, themes, and a
file browser. Graphics protocol encoders are unit-tested but have not yet been
validated on real Kitty, Sixel, or iTerm2 hardware. Audio is off by default and
the PulseAudio path is documented but not exercised in CI.

- Feature inventory: [`FEATURES.md`](FEATURES.md)
- Roadmap: [`ROADMAP.md`](ROADMAP.md)
- Architecture: [`ARCHITECTURE.md`](ARCHITECTURE.md)
- Tech stack: [`TECH_STACK.md`](TECH_STACK.md)
- UI design: [`DESIGN.md`](DESIGN.md)
- Third-party licenses: [`LICENSES.md`](LICENSES.md)
- Current status and decisions: [`PROGRESS.md`](PROGRESS.md)
- Audio setup: [`docs/AUDIO-SETUP.md`](docs/AUDIO-SETUP.md)
- libVLC benchmark: [`docs/benchmarks/libvlc-callbacks.md`](docs/benchmarks/libvlc-callbacks.md)
- Render-mode benchmark: [`docs/benchmarks/render-modes.md`](docs/benchmarks/render-modes.md)

## Requirements

- Linux, macOS, or WSL (primary target: Linux)
- libVLC 3.x at runtime and `libvlc-dev` + `pkg-config` to build
- Go 1.24+ only if building from source

## Install

One command in any terminal:

```bash
curl -fsSL https://raw.githubusercontent.com/fs8280071-gif/player-tui/main/install.sh | bash
```

The installer clones or updates the source in
`${XDG_DATA_HOME:-$HOME/.local/share}/player-tui`, builds `player-tui` with
`go build` into `$HOME/.local/bin`, and adds that directory to your shell PATH
when needed.

Inspect the script first if you prefer:

```bash
curl -fsSL https://raw.githubusercontent.com/fs8280071-gif/player-tui/main/install.sh -o install.sh
less install.sh
bash install.sh
```

Or build a checkout directly:

```bash
git clone https://github.com/fs8280071-gif/player-tui.git
cd player-tui
bash install.sh
```

Installer options:

```bash
PLAYER_TUI_DIR="$HOME/src/player-tui" \
PLAYER_TUI_BIN_DIR="$HOME/.local/bin" \
bash install.sh
```

When `install.sh` runs from a source checkout it builds that checkout. Set
`PLAYER_TUI_FROM_REMOTE=1` to force a fresh clone or update.

## Usage

```bash
player-tui                                  # file browser in the current directory
player-tui browse /path/to/media            # file browser in a directory
player-tui play /path/to/video.mkv          # play one file
player-tui play a.mp4 b.mp4 c.mkv           # queue several files
player-tui play https://example.com/video.m3u8 --start 00:02:30
player-tui play /path/to/video.mkv --render quadrant --dither
player-tui doctor                           # terminal and capability report
```

`play` and `open` are aliases. `--audio` enables libVLC audio output (default
off). `--theme` selects a palette. `PLAYER_TUI_RENDER` overrides auto render
selection; `PLAYER_TUI_DUMMY_AUDIO=1` forces dummy audio.

### Options

| Option | Meaning |
|---|---|
| `--render <mode>` | `kitty`, `sixel`, `iterm2`, `halfblock`, `quadrant`, `sextant`, or `auto` |
| `--theme <name>` | `catppuccin`, `gruvbox`, `nord`, `high-contrast` |
| `--audio` | enable audio output (default off) |
| `--dither` | 4x4 Bayer ordered dithering |
| `--title <text>` | display title (single target only) |
| `--sub-file <path>` | attach an external subtitle file |
| `--start <time>` | start position: `SS`, `MM:SS`, or `HH:MM:SS` |
| `--frames <n>` | bench: stop after n frames (0 = until end) |
| `--cols <n>` / `--rows <n>` | bench: cell grid (default 120x40) |
| `-h`, `--help` | show help |

### Playback keys

| Key | Action |
|---|---|
| `space` | pause / resume |
| `left` / `right` | seek -5 s / +5 s |
| `+` / `-` | volume +/- 5% |
| `m` | mute / unmute |
| `[` / `]` | slower / faster (0.25x steps) |
| `s` / `a` | cycle subtitle / audio track |
| `n` / `p` | next / previous file in the queue |
| `?` | help overlay |
| `q` | quit |

### Browser keys

| Key | Action |
|---|---|
| `j` / `down`, `k` / `up` | move selection |
| `enter` | open file or directory |
| `h` / `backspace` | parent directory |
| `?` | help overlay |
| `q` | quit |

## Render modes

| Mode | Samples per cell | Output | Notes |
|---|---|---|---|
| `halfblock` | 1x2 | `▀` truecolor | Most portable; one foreground and one background color per cell. |
| `quadrant` | 2x2 | Quadrant Block Elements | Good detail; widely supported. |
| `sextant` | 2x3 | Sextant Block Elements (Unicode 13) | Most detail; needs a modern font. |
| `kitty` | image | Kitty graphics protocol | Raw RGB, base64 chunking, image IDs, delete on exit. Unit-tested only. |
| `sixel` | image | Sixel | 216-color cube quantization, run-length bands. Unit-tested only. |
| `iterm2` | image | OSC 1337 inline image | Width/height in cells. Unit-tested only. |
| `auto` | - | - | Prefers Kitty, then iTerm2, then Sixel, then quadrant/halfblock. |

Every block mode uses exact area averaging when downscaling, preserves the source
aspect ratio, and reserves the bottom row for status. `--dither` adds ordered
dithering before cell quantization; it improves gradients but roughly doubles or
triples ANSI output size.

Font coverage cannot be queried reliably. If quadrant or sextant glyphs render
as boxes, use `--render halfblock` or `PLAYER_TUI_RENDER=halfblock`.

### Terminal compatibility

Capabilities vary by version, build flags, font, and multiplexer configuration.
This table records what was actually exercised.

| Host | Kitty | Sixel | iTerm2 | Block modes | Verified here |
|---|---|---|---|---|---|
| xterm / Termux-style terminal | - | - | - | halfblock, quadrant, sextant | Yes (PTY, xterm-256color) |
| Kitty, Ghostty, WezTerm | encoder | - | encoder | yes | No real-device test |
| iTerm2 | encoder | - | encoder | yes | No real-device test |
| xterm/WezTerm/Windows Terminal | - | encoder | - | yes | No real-device test |
| tmux / Zellij / screen | passthrough needed | host-dependent | passthrough needed | yes | No real-device test |

Kitty, Sixel, and iTerm2 output is checked against fixed byte sequences in unit
tests and PTY-tested for startup, cleanup, and alt-screen restore, but not on a
terminal that displays the images. Treat those modes as experimental.

## Config and resume

Settings live in `~/.config/player-tui/config.toml` (override with
`PLAYER_TUI_CONFIG`). Missing keys use defaults. Comments (`#`) and quoted
strings are supported.

```toml
audio = false
dither = false
render = "auto"
theme = "catppuccin"
volume = 100
resume = true
```

Per-file resume positions are stored in `resume.json` in the same directory.
Positions under 5 seconds are ignored, and a position within 3 seconds of the end
is forgotten. `PLAYER_TUI_CONFIG` moves both files. Resume only applies to local
files, not URLs.

`NO_COLOR=1` disables palette colors in the status row and overlays (reverse
video is used instead). Video block/image colors are unchanged.

## Audio

Audio is off by default. Pass `--audio` after configuring an output. The
supported development path is PulseAudio over TCP on loopback from Termux to
Debian; exact commands for both sides are in
[`docs/AUDIO-SETUP.md`](docs/AUDIO-SETUP.md). Never bind PulseAudio to a public
interface.

## Diagnostics

```bash
player-tui doctor
```

Prints terminal size, `TERM`, `COLORTERM`, `TERM_PROGRAM`, detected Kitty/Sixel/
iTerm2/truecolor support, the auto-selected render mode, and fallbacks. Run it in
the terminal you plan to use.

## Benchmarking render modes

`bench` decodes a clip, renders frames into memory, and reports timing without
touching the terminal:

```bash
player-tui bench /path/to/clip.mp4 --render halfblock --frames 180 --cols 120 --rows 40
player-tui bench /path/to/clip.mp4 --render quadrant  --frames 180 --cols 120 --rows 40
player-tui bench /path/to/clip.mp4 --render sextant   --frames 180 --cols 120 --rows 40
```

Compare `render_p50` (median per-frame renderer time) and `render_fps` first;
those isolate the renderer. `display_fps` is end-to-end decode + render and can be
capped by the source frame rate or device throttling.

## External player handoff

The command surface accepts a local file or a direct stream URL, so another TUI,
downloader, or file manager can launch it as its external player:

```bash
player-tui open /path/to/video.mkv
player-tui open /path/to/video.mkv --title "My Video" \
  --sub-file /path/to/video.srt --start 00:02:30
player-tui open https://example.com/video.m3u8
player-tui /path/to/video.mkv
```

Generic handoff only; the player does not include a scraper or downloader for
third-party catalogues.

## Exit codes

| Code | Meaning |
|---|---|
| `0` | normal exit |
| `2` | usage, option, or config error |
| `3` | target or playback error |
| `4` | terminal or libVLC initialization error |

## Development

```bash
go test ./...
go vet ./...
go build -p 2 -o player-tui ./cmd/player-tui
```

The fuzz target for the terminal sanitizer runs with
`go test -fuzz=FuzzText ./internal/sanitize`.

## License and naming

MIT. See [`LICENSE`](LICENSE). The project does not use VLC's name, cone logo, or
branding and dynamically links libVLC under the LGPL. Third-party notices are in
[`LICENSES.md`](LICENSES.md).

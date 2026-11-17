# player-tui

VLC-class terminal media player: local and streamed audio/video, subtitles,
library, terminal-native rendering, and optional offline AI.

## Status

The first runnable MVP slice works: `player-tui play <file>` opens local media
with libVLC, renders I420 video as truecolor blocks, and supports pause, seek,
and quit. Rendering supports half-block, quadrant, and sextant modes, exact
area-average downscaling, optional ordered dithering, and aspect-correct layout
at the full terminal size. Phase C design docs are complete (`TECH_STACK.md`,
`DESIGN.md`, `LICENSES.md`). The broader player (browser, queue, subtitles,
Kitty/Sixel, library, AI) is still in progress.

- Feature inventory: [`FEATURES.md`](FEATURES.md)
- Roadmap: [`ROADMAP.md`](ROADMAP.md)
- Architecture: [`ARCHITECTURE.md`](ARCHITECTURE.md)
- Tech stack: [`TECH_STACK.md`](TECH_STACK.md)
- UI design: [`DESIGN.md`](DESIGN.md)
- Third-party licenses: [`LICENSES.md`](LICENSES.md)
- Current status and decisions: [`PROGRESS.md`](PROGRESS.md)
- libVLC benchmark: [`docs/benchmarks/libvlc-callbacks.md`](docs/benchmarks/libvlc-callbacks.md)
- Render-mode benchmark: [`docs/benchmarks/render-modes.md`](docs/benchmarks/render-modes.md)

## Install / download

One command in any terminal:

```bash
curl -fsSL https://raw.githubusercontent.com/fs8280071-gif/player-tui/main/install.sh | bash
```

This clones or updates the source in
`${XDG_DATA_HOME:-$HOME/.local/share}/player-tui`, builds `player-tui` into
`$HOME/.local/bin`, and adds that directory to your shell PATH when needed.
Building requires Go 1.24+, `libvlc-dev`, and `pkg-config`.

Prefer to inspect the script first:

```bash
curl -fsSL https://raw.githubusercontent.com/fs8280071-gif/player-tui/main/install.sh -o install.sh
less install.sh
bash install.sh
```

Manual source download:

```bash
git clone https://github.com/fs8280071-gif/player-tui.git
cd player-tui
```

Or without git:

```bash
curl -fsSL https://github.com/fs8280071-gif/player-tui/archive/refs/heads/main.tar.gz | tar -xz
cd player-tui-main
```

### Installer options

```bash
PLAYER_TUI_DIR="$HOME/src/player-tui" \
PLAYER_TUI_BIN_DIR="$HOME/.local/bin" \
bash install.sh
```

When `install.sh` is run from a source checkout, it builds that checkout. Set
`PLAYER_TUI_FROM_REMOTE=1` to force a fresh clone/update instead.

## Usage

```bash
player-tui play /path/to/video.mkv
player-tui play https://example.com/video.m3u8 --start 00:02:30
player-tui play /path/to/video.mkv --render quadrant
player-tui play /path/to/video.mkv --render sextant --dither
PLAYER_TUI_RENDER=halfblock player-tui play /path/to/video.mkv
```

Keys: `space` pause/resume, `left`/`right` seek -5 s/+5 s, `q` quit. Use
`PLAYER_TUI_DUMMY_AUDIO=1` for headless tests with no audio device.

### Render modes

| Mode | Samples per cell | Glyphs | Notes |
|---|---|---|---|
| `halfblock` | 1x2 | `▀` | Most portable; two colors per cell. |
| `quadrant` | 2x2 | Quadrant Block Elements | Good detail; widely supported. |
| `sextant` | 2x3 | Sextant Block Elements (Unicode 13) | Most detail; needs a modern font. |
| `auto` | - | - | Uses quadrant for normal terminals, halfblock for `TERM=dumb`/`linux` or `PLAYER_TUI_ASCII=1`. |

Every mode uses exact area averaging when downscaling and preserves the source
aspect ratio, assuming terminal cells are twice as tall as they are wide. The
status row is reserved at the bottom; the video is centered in the remaining
space. `--dither` adds 4x4 Bayer ordered dithering before cell quantization; it
improves gradients but roughly doubles or triples the ANSI output size.

Font coverage cannot be queried reliably. If quadrant or sextant glyphs render
as boxes, use `--render halfblock` or set `PLAYER_TUI_RENDER=halfblock`.
`PLAYER_TUI_RENDER` overrides `auto`.

### Benchmarking render modes

`bench` decodes a clip, renders frames into memory, and reports timing without
touching the terminal:

```bash
player-tui bench /path/to/clip.mp4 --render halfblock --frames 180 --cols 120 --rows 40
player-tui bench /path/to/clip.mp4 --render quadrant  --frames 180 --cols 120 --rows 40
player-tui bench /path/to/clip.mp4 --render sextant   --frames 180 --cols 120 --rows 40
```

Compare `render_p50` (median per-frame renderer time) and `render_fps` first;
those isolate the renderer. `display_fps` is end-to-end decode + render and can
be capped by the source frame rate or device throttling. Keep `--cols`,
`--rows`, and `--frames` identical between runs. The phone results and exact
commands are in the render-mode benchmark document.

## External player handoff

The planned command surface accepts a local file or a direct stream URL, so
another TUI, downloader, or file manager can launch it as its external player:

```bash
player-tui open /path/to/video.mkv
player-tui open /path/to/video.mkv --title "My Video" \
  --sub-file /path/to/video.srt --start 00:02:30
player-tui open https://example.com/video.m3u8
player-tui /path/to/video.mkv
```

Completed downloads can be opened directly by path, and media browsers can hand
off a stream URL without a separate wrapper. Generic handoff only; the player
does not include a scraper or downloader for third-party catalogues.

## Requirements

- Linux, macOS, or WSL (primary target: Linux)
- libVLC 3.x at runtime and `libvlc-dev` + `pkg-config` to build
- Go 1.24+ only if building from source

## License and naming

MIT. See [`LICENSE`](LICENSE). The project will not use VLC's name, cone logo, or
branding, and will dynamically link libVLC under the LGPL. Third-party notices
will be recorded in `LICENSES.md` before release.

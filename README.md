# player-tui

VLC-class terminal media player: local and streamed audio/video, subtitles,
library, terminal-native rendering, and optional offline AI.

## Status

The first runnable MVP slice works: `player-tui play <file>` opens local media
with libVLC, renders I420 video as truecolor half-blocks, and supports pause,
seek, and quit. Phase C design docs are complete (`TECH_STACK.md`, `DESIGN.md`,
`LICENSES.md`). The broader player (browser, queue, subtitles, Kitty/Sixel,
library, AI) is still in progress.

- Feature inventory: [`FEATURES.md`](FEATURES.md)
- Roadmap: [`ROADMAP.md`](ROADMAP.md)
- Architecture: [`ARCHITECTURE.md`](ARCHITECTURE.md)
- Tech stack: [`TECH_STACK.md`](TECH_STACK.md)
- UI design: [`DESIGN.md`](DESIGN.md)
- Third-party licenses: [`LICENSES.md`](LICENSES.md)
- Current status and decisions: [`PROGRESS.md`](PROGRESS.md)
- libVLC benchmark: [`docs/benchmarks/libvlc-callbacks.md`](docs/benchmarks/libvlc-callbacks.md)

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
```

Keys: `space` pause/resume, `left`/`right` seek -5 s/+5 s, `q` quit. The
half-block renderer needs truecolor for the best result. Use
`PLAYER_TUI_DUMMY_AUDIO=1` for headless tests with no audio device.

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

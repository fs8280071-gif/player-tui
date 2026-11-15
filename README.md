# player-tui

VLC-class terminal media player: local and streamed audio/video, subtitles,
library, terminal-native rendering, and optional offline AI.

## Status

Phase B (research and feature inventory) is complete. There is **no runnable
player binary yet**; the MVP build starts in Phase C. The installer currently
downloads the source and will build the binary automatically once a Go build
target exists.

- Feature inventory and build plan: [`FEATURES.md`](FEATURES.md)
- Current status and decisions: [`PROGRESS.md`](PROGRESS.md)
- libVLC benchmark: [`docs/benchmarks/libvlc-callbacks.md`](docs/benchmarks/libvlc-callbacks.md)

## Install / download

One command in any terminal:

```bash
curl -fsSL https://raw.githubusercontent.com/fs8280071-gif/player-tui/main/install.sh | bash
```

This clones or updates the source in
`${XDG_DATA_HOME:-$HOME/.local/share}/player-tui`. If Go and a build target are
present, it also builds `player-tui` into `$HOME/.local/bin`.

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

## Requirements

- Linux, macOS, or WSL (primary target: Linux)
- libVLC 3.x at runtime (the player will link dynamically)
- Go 1.24+ only if building from source

## License and naming

License is not finalized yet. The project will not use VLC's name, cone logo, or
branding, and will dynamically link libVLC under the LGPL. Third-party notices
will be recorded in `LICENSES.md` before release.

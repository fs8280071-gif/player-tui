# Third-party licenses

player-tui itself is MIT-licensed; see [`LICENSE`](LICENSE). This file records
the licenses of direct dependencies and runtime components. It must be complete
before any public binary release.

**No VLC branding.** player-tui uses libVLC as a library only. It does not ship
VLC's name, cone logo, or artwork.

## Runtime dependencies

| Component | Version | License | Linking | Notes |
|---|---|---|---|---|
| libVLC (`libvlc`) | 3.0.24 | LGPL-2.1-or-later | Dynamic | Core engine. Distributed by the OS/package manager, not bundled by default. |
| VLC plugins/modules | 3.0.24 | LGPL-2.1-or-later and GPL-2.0-or-later | Dynamic, loaded by libVLC | Codec/access/output modules vary by distribution; installed as a system package. |
| Go standard library | 1.24.x | BSD-3-Clause | Static | Runtime linked into the binary. |
| `golang.org/x/term` | pinned in `go.mod` | BSD-3-Clause | Static | Raw mode and terminal size. |
| `github.com/rockorager/vaxis` | pinned in `go.mod` | MIT | Static | Planned TUI layer (cell grid, input, resize). |

### libVLC licensing note

libVLC is dynamically linked, which keeps the MIT project compatible with the
LGPL. The user's distribution supplies libVLC and its plugins. If a future build
bundles libVLC or GPL-only plugins, the distribution and notices must be reviewed
again. Source for the exact libVLC build is available from the distribution or
VideoLAN.

## Planned V1 dependencies (not yet linked)

| Component | License | Use | Verify before shipping |
|---|---|---|---|
| `modernc.org/sqlite` | BSD-3-Clause | Library index, FTS5 | Yes |
| `github.com/dhowden/tag` | MIT | Audio metadata | Yes |
| `github.com/asticode/go-astisub` | MIT | SRT/VTT/ASS parsing | Yes |
| `github.com/mattn/go-sixel` | MIT | Sixel encoding (fallback/custom encoder planned) | Yes |
| `github.com/godbus/dbus/v5` | BSD-2-Clause | MPRIS/DBus (V1 parity) | Yes |
| `github.com/fsnotify/fsnotify` | BSD-3-Clause | Library watcher | Yes |
| `ffmpeg`/`ffprobe` | LGPL-2.1+ or GPL-2.0+ depending on build | External metadata/audio extraction | Yes; never bundle a GPL build by accident |

## Planned AI components (models and runtimes)

Models are optional, downloaded only with explicit consent, and never committed
to this repository. Check the exact model card and weights license at install
time.

| Component | License (verify) | Use |
|---|---|---|
| whisper.cpp | MIT | Offline speech recognition runtime |
| Whisper model weights | MIT (OpenAI release; verify model card) | ASR |
| Silero VAD | MIT | Voice activity detection |
| Opus-MT models | CC-BY-4.0 (verify per model) | Offline translation |
| llama.cpp | MIT | Local assistant runtime |
| Small local LLMs | Varies (Apache-2.0/MIT/Llama community) | Ask-the-video, commands |
| Demucs | MIT | Source separation |
| DeepFilterNet | Apache-2.0 or MIT (verify release) | Noise removal |
| ONNX Runtime | MIT | Optional inference backend |

## Attribution rules

- Do not copy code without its license and copyright notice.
- Keep `go.sum` and this file in sync when dependencies change.
- Ship license texts for static dependencies in release archives.
- Do not use VLC trademarks, the cone logo, or imply endorsement by VideoLAN.
- Record any bundled font, icon set, theme, or test media asset here.

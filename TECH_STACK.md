# player-tui tech stack

Status: Phase C decision record, 2026-11-16. See `ARCHITECTURE.md` for the
system design and `PROGRESS.md` for benchmark evidence.

## Summary

| Area | Choice | Why | Trade-off |
|---|---|---|---|
| Language | Go 1.24 | Fast enough for I420 sampling and ANSI output; simple cross-builds; cgo is contained. | Smaller TUI ecosystem than Rust; no compile-time memory safety across the cgo edge. |
| Playback | libVLC 3.0.24, dynamically linked | Codec/container/stream coverage and audio clock for free; matches installed Debian package. | LGPL/libVLC runtime dependency; callback API is less flexible than FFmpeg. |
| libVLC binding | Thin in-repo cgo layer | `adrg/libvlc-go` exposes no raw video callbacks; the layer is small and auditable. | Maintain against libVLC 3.x header changes. |
| Video path | I420 callback -> YUV sampling in Go -> terminal | Benchmarked best path: 33-41 fps at 1080p display-only; no full RGBA intermediate. | Callback copy and per-cell sampling cost CPU. |
| Terminal (MVP slice) | `golang.org/x/term` + direct ANSI compositor | Smallest runnable slice; raw mode, alternate screen, input, resize. | Not a full widget toolkit; panes/palette arrive with the Vaxis layer. |
| TUI target | Vaxis (`github.com/rockorager/vaxis`) | Cell grid, Unicode width, key parsing, resize; active project, MIT. | Raw graphics output needs a compositor spike; text panes land after the primitives. |
| Block modes | Custom truecolor renderer | Halfblock (1x2, `▀`), quadrant (2x2), and sextant (2x3) samples per cell; exact area averaging; portable and works over SSH. | High output volume, especially with dither; needs frame dropping and a low-bandwidth mode. |
| Kitty/Sixel | Custom encoders (later) | No runtime dependency; control over chunking and cleanup. | Larger test matrix; protocol edge cases. |
| Subtitles | `asticode/go-astisub` for SRT/VTT; custom ASS text | Mature parsers; styling approximated in cells. | Bitmap subtitles remain libVLC's job. |
| Library (V1) | `modernc.org/sqlite` + FTS5 | Pure Go, no CGO or system SQLite dependency. | Slightly slower than C SQLite; acceptable at 1k-100k items. |
| Metadata (V1) | `dhowden/tag` + `ffprobe` fallback | Go tags for common audio; ffprobe for containers/tracks. | External ffprobe dependency for uncommon formats. |
| AI (V1) | External worker processes; whisper.cpp, Silero VAD | Keeps large runtimes and crashes out of the player process; offline by default. | Process supervision and model management. |
| Config | TOML | Human-editable, comments, mature Go parser. | Another parser to validate. |
| IPC (V1) | Unix socket, JSON Lines | Scriptable, observable, simple framing. | Per-platform peer credential code. |
| Build | Go modules; `go build -trimpath` | Single binary except dynamic libVLC; install.sh uses it. | cgo cross-compilation needs a C toolchain and target libVLC. |
| Test | `go test`, race detector, PTY tests, golden frame/ANSI snapshots | Covers engine, renderer, terminal restore. | Graphics terminals still need a manual matrix. |

## Versions verified in this environment

- Go 1.24.4 (linux/arm64).
- libVLC 3.0.24 via `pkg-config --modversion libvlc`.
- `golang.org/x/term` v0.36.0 pinned (Go 1.24-compatible; later versions need Go 1.26).
- `github.com/rockorager/vaxis` v0.17.1 available.

Pin exact versions in `go.mod`/`go.sum`; do not rely on latest at build time.

## Dependency rules

- `engine`, `video`, `render`, `term`, and `ui` stay independent of each other
  except through the contracts in `ARCHITECTURE.md`.
- No cgo outside `internal/engine/libvlc`.
- No network calls during playback, indexing, or AI inference.
- Third-party licenses are recorded in `LICENSES.md` before release.

## Deliberately not chosen

- **mpv JSON IPC:** rejected after the engine benchmarks; it adds a process and
  still needs a renderer, while libVLC gives the callback path directly.
- **Rust/Ratatui:** viable, but the benchmark showed Go meets the budget and the
  libVLC cgo surface is small. Revisit only with new measurements.
- **FFmpeg-only engine:** more control, much more work (demux, decode, sync,
  network protocols). LibVLC is the MVP engine; an FFmpeg backend can implement
  the same `Engine` interface later.
- **Textual/Python:** fastest prototype, weakest single-binary and performance
  story.

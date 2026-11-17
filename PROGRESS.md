# PROGRESS

## Status

- Phase C design docs are complete: `ROADMAP.md`, `ARCHITECTURE.md`,
  `TECH_STACK.md`, `DESIGN.md`, `LICENSES.md`.
- MVP implementation started: a runnable playback slice exists (`player-tui
  play <file>`), with libVLC I420 callbacks, truecolor halfblock/quadrant/sextant
  rendering, exact area averaging, optional ordered dithering, aspect-correct
  layout, pause, seek, terminal restore, and a render benchmark command.
- Next: capability detection/`doctor`, frame scheduler, subtitle rendering,
  queue, Kitty/Sixel, and the Vaxis pane UI.
- Blocked: no local audio/Kitty/Sixel for end-to-end testing. End-to-end display
  FPS varies with phone thermal state; renderer-only p50 numbers are stable.

## Decisions

- License: MIT (`LICENSE`). Dynamically linked LGPL libVLC is compatible;
  third-party notices are recorded in `LICENSES.md`.
- Playback engine: libVLC 3.0.24 (LGPL-2.1+), dynamically linked. Former mpv IPC
  idea was rejected after benchmarks; do not switch engines without a new record.
- Language: Go 1.24 + Vaxis (`github.com/rockorager/vaxis`). libVLC access via a
  thin in-repo cgo layer; `adrg/libvlc-go` exposes no raw video callbacks.
- Frame path: request I420, average the full pixel area for each output sample,
  convert to RGB in Go, and emit truecolor block glyphs. Render modes are
  halfblock (1x2), quadrant (2x2), and sextant (2x3); `auto` picks quadrant for
  normal terminals and halfblock for dumb/linux terminals. RV32 fallback and the
  explicit drop scheduler are still TODO; libVLC audio remains the master clock.
- libVLC pads its callback height (1080p arrives as 1090 rows, 720p as 738).
  The player reads the visible size with `libvlc_video_get_size` and exposes the
  unpadded frame, so aspect fitting uses 1920x1088 or 1280x720.
- Do not enable `avcodec-fast` by default. Keep automatic decoder threads.
- Audio: no audio / dummy for benchmarks until real audio is configured.
- VLC reference: `/root/vlc-3.0.24` (outside the repo); `vlc-android/` is
  gitignored and kept for reference only.
- Terminal layer for the first slice: `golang.org/x/term` plus a small ANSI
  compositor. Vaxis remains the target for panes/palette; `term.Surface` keeps
  the swap contained.
- Frame delivery for the first slice: poll the C-owned I420 buffer at 60 Hz and
  copy only when the display callback increments the sequence. This avoids Go
  callbacks from libVLC threads; a mailbox/clock abstraction stays on the Phase
  C follow-up list.
- Remotes: GitLab `https://gitlab.com/fs8280071-gif/player-tui` (origin,
  private) and GitHub `https://github.com/fs8280071-gif/player-tui` (public).

## Done

- Inspected workspace and found `vlc-android`
  (remoteaccess-0.27.0-3-g14a742245), not VLC core.
- Confirmed VLC stable 3.0.24 and Debian libVLC 3.0.24.
- Installed `libvlc-dev` + `vlc-plugin-base` and downloaded/extracted VLC
  3.0.24 source to `/root/vlc-3.0.24`.
- Created private GitLab project and set `origin`.
- Checked `adrg/libvlc-go`: active (last code push 2026-07-16), MIT, but no
  video-callback API, so the project uses a thin cgo layer.
- Built and ran the libVLC video-callback benchmark in Go 1.24 + cgo. Full
  results are in `docs/benchmarks/libvlc-callbacks.md`.
- Audited the libVLC 3.0.24 subtitle/SPU API: no public text-subtitle callback;
  `libvlc_video_set_subtitle_file` is deprecated; bitmap subtitles are blended
  into callback frames.
- Researched competitors (`mpv`, `mpvc`, SparkPlayer, CLIAMP, termusic, kew,
  cmus, musikcube, ncmpcpp, ncplayer, chafa, Yazi, VLC, Kodi), terminal graphics
  support, VLC AI-subtitle status, and AI/model licenses.
- Verified name availability on crates.io, npm, PyPI, Homebrew, AUR, Debian, and
  GitLab; documented the unresolved GitHub exact-name conflict.
- Confirmed current Vaxis module path and versions; `charmbracelet/vaxis` does
  not exist. Confirmed availability of `dhowden/tag`, `modernc.org/sqlite`,
  `mattn/go-sixel`, `asticode/go-astisub`, `godbus/dbus/v5`, and `fsnotify`.
- Wrote `FEATURES.md`: 146 feature rows across 11 categories, cross-cutting
  design decisions, MVP criteria, top-20 build order, phases, risks, and sources.
- Created the public GitHub repo, added `README.md`, and added `install.sh` for
  one-command source download/update; it will build once a Go target exists.
- Added the MIT license and documented generic external-player handoff for local
  files and direct stream URLs.
- Added the MVP launcher (`C12`: `open` plus `--title`, `--sub-file`, `--start`)
  and the V1 Jellyfin/Plex/DLNA client (`L13`) to `FEATURES.md`.
- Wrote `ROADMAP.md` and `ARCHITECTURE.md`; Phase C design baseline.
- Wrote `TECH_STACK.md`, `DESIGN.md`, and `LICENSES.md` (Phase C complete).
- Implemented the first MVP slice in Go 1.24:
  - `cmd/player-tui` with `play`/`open`, `--title`, `--sub-file`, `--start`.
  - Thin cgo libVLC player with I420 format/lock/unlock/display callbacks, a
    mutex-protected frame buffer, and sequence-based frame polling.
  - Half-block truecolor renderer with aspect fit, letterboxing, centered
    layout, and a status row.
  - Raw mode, alternate screen, key parsing (space/arrows/q), SIGINT/SIGTERM
    handling, and idempotent terminal restore.
  - `install.sh` now checks libVLC, builds with `go build`, installs to
    `$HOME/.local/bin`, and adds that directory to the shell PATH.
- Generated a 5-second 320x240 MPEG-4 test clip with ffmpeg and drove the binary
  through a PTY: half-block cells, truecolor SGR, pause, seek clamp, quit,
  alt-screen enter/leave, and cursor restore all passed (exit 0).
- Replaced sparse nearest sampling with exact area averaging over Y/U/V for every
  output sample.
- Added halfblock, quadrant, and sextant renderers with `--render`, `auto`,
  `PLAYER_TUI_RENDER`, and `--dither` (4x4 Bayer). Quadrant/sextant use a
  two-color per-cell clustering pass; sextant uses Unicode 13 block glyphs with
  `▌`/`▐`/`█` fallbacks for missing masks.
- Fixed aspect handling for decoder-padded frames by querying the visible size
  from libVLC (1920x1088 source is no longer treated as 1920x1090).
- Added `player-tui bench` for in-memory rendering with `--frames`, `--cols`,
  `--rows`, `--render`, and `--dither`; reports display FPS, render p50/p95,
  render FPS, and bytes per frame.
- PTY-tested halfblock, quadrant, sextant, auto, and dither at 320x240 and
  1920x1088; all modes restored the terminal and exited cleanly.
- Added an escape sanitizer (`internal/sanitize`) with unit tests and a fuzz
  target; titles, notices, doctor values, and all CLI/app errors pass through it
  before reaching the terminal.
- Added Catppuccin, Gruvbox, Nord, and high-contrast themes (`--theme` or the
  config file), with `NO_COLOR` forcing reverse-video instead of palette colors.
- Added a TOML config (`~/.config/player-tui/config.toml`, override with
  `PLAYER_TUI_CONFIG`) for audio, render, dither, theme, volume, and resume, plus
  per-file resume positions stored in `resume.json` next to it.
- Added volume +/- and mute, speed [ / ], subtitle/audio track cycling, a
  help overlay (`?`), and speed/track notices in the status row.
- Added a multi-file queue (`player-tui play a.mp4 b.mp4`) with `n`/`p` for
  next/previous and a `[current/total]` status prefix.
- Added `--audio` (default off; libVLC uses a dummy output unless enabled) and
  `docs/AUDIO-SETUP.md` with Termux-side and Debian-side PulseAudio-over-TCP
  commands. Termux-side commands were documented only, never executed.
- Added a `doctor` command (terminal size, TERM/COLORTERM/TERM_PROGRAM,
  graphics/truecolor detection, chosen render mode, fallbacks), volume and mute
  fields in the status snapshot, percentage in the status row, and explicit
  errors for missing, empty, directory, and space-containing paths.
- Added Kitty, Sixel, and iTerm2 encoders as separate packages
  (`internal/render/{kitty,sixel,iterm2}`) with unit tests against fixed byte
  sequences, `--render kitty|sixel|iterm2`, and auto-detection from
  `KITTY_WINDOW_ID`, `TERM`, `TERM_PROGRAM`, and `COLORTERM`.
  **Encoders are unit-tested only; not yet tested on a real Kitty, Sixel, or
  iTerm2 terminal.**

## Benchmark summary

- Exact area-average rendering at 120x40, 90-frame medians (see
  `docs/benchmarks/render-modes.md`): 720p halfblock 10.2 ms/frame (98 fps),
  quadrant 12.9 ms (77 fps), sextant 14.5 ms (69 fps). 1080p halfblock 17.2 ms
  (58 fps), quadrant 18.2 ms (55 fps), sextant 19.9 ms (50 fps).
- `--dither` adds roughly 0-3 ms/frame and roughly doubles to triples ANSI
  output (about 108-141 KB per frame at 120x40 versus 43-46 KB without dither).
- End-to-end `display_fps` on 30 fps sources was 29-30.5 fps in the final runs;
  earlier throttled runs fell to 18-25 fps. Use `render_p50` for mode
  comparisons.
- The callback frame copy and decoder share the CPU with the renderer at 1080p;
  halfblock has the most end-to-end headroom. No frame scheduler yet, so a slow
  renderer delays display rather than dropping late frames.
- Older libVLC callback path benchmark: I420 chosen over RV32; decoder had
  headroom, callback-vout conversion/copy/output was the bottleneck. Keep that
  record in `docs/benchmarks/libvlc-callbacks.md`.

## Next

- Continue the MVP: `doctor`, frame scheduler/drop metrics, external subtitle
  cues, queue/playlists, and the Vaxis pane UI. Validate graphics encoders on
  real terminals.
- Renderer follow-ups: automatic font/tofu probing where possible, a
  low-bandwidth mode, and a summed-area-table path if 1080p60 becomes a target.
- AI v1 follows MVP: offline ASR, VAD, batch jobs, model manager, privacy/wipe,
  and scheduler. Live audio-tap features remain V2.

## Blocked / open

- Terminal here has no Kitty/Sixel and no audio output; graphics encoders and
  audio paths need unit tests and external/real-device tests.
- Real audio is not configured yet; PulseAudio over TCP from Termux is the next
  audio step, followed by audio-master-clock sync.
- End-to-end 1080p performance varies with phone thermal state; laptop
  benchmarks are still required.
- GitHub exact-name conflict is unresolved; pick a distinctive name before
  public release.
- Embedded text subtitles need an FFmpeg demux/decode path in V1.

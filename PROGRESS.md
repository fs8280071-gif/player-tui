# PROGRESS

## Status

- Phase: B complete; `FEATURES.md` written and reviewed
- Next: Phase C planning docs and MVP build, after user approval
- Blocked: `glab` token; no local audio/Kitty/Sixel for end-to-end testing

## Decisions

- Playback engine: libVLC 3.0.24 (LGPL-2.1+), dynamically linked. Former mpv IPC
  idea was rejected after benchmarks; do not switch engines without a new record.
- Language: Go 1.24 + Vaxis (`github.com/rockorager/vaxis`). libVLC access via a
  thin in-repo cgo layer; `adrg/libvlc-go` exposes no raw video callbacks.
- Frame path: request I420, sample Y/U/V at the terminal grid, convert to RGB in
  Go; RV32 fallback. Audio is the master clock and late video frames are dropped.
- Do not enable `avcodec-fast` by default. Keep automatic decoder threads.
- Audio: no audio / dummy for benchmarks until real audio is configured.
- VLC reference: `/root/vlc-3.0.24` (outside the repo); `vlc-android/` is
  gitignored and kept for reference only.
- Remote: `https://gitlab.com/fs8280071-gif/player-tui` (private).

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
- Wrote `FEATURES.md`: 144 feature rows across 11 categories, cross-cutting
  design decisions, MVP criteria, top-20 build order, phases, risks, and sources.

## Benchmark summary

- Go half-block render: ~1.1-1.5 ms/frame at 120x40; ASCII ~0.3 ms/frame.
  Pure render: 649-937 fps at 120x40, 402 fps at 200x50.
- RV32: 720p ~26 fps; 1080p ~17 fps. I420: 720p ~26.5 fps at lower CPU; 1080p
  display-only **33-38 fps** (best runs 38-41 fps). I420 is the chosen path.
- The callback frame copy costs about 1 fps at 1080p; the decoder has headroom
  (about 55-66 fps at 720p, 88-122 fps at 1080p). Callback-vout
  conversion/copy/output is the bottleneck.
- Fast-decode (`avcodec-fast`) and `skip4` cut decoder CPU ~20-25% but did not
  improve display fps consistently; `threads=1` hurts 1080p.
- 120x40 half-block output is ~176 KB/frame (5.3 MB/s at 30 fps); 200x50 is
  ~366 KB/frame (11 MB/s at 30 fps). A low-bandwidth mode and frame dropping
  are required.
- Conclusion: continue with Go; do not switch to Rust. Validate I420 + 1080p on
  a mid-range laptop later.

## Next

- Phase C (after approval): `ARCHITECTURE.md`, `TECH_STACK.md`, `DESIGN.md`,
  `ROADMAP.md`, `LICENSES.md`, then the MVP build starting with engine
  lifecycle, clock/sync, terminal safety, and the half-block renderer.
- AI v1 follows MVP: offline ASR, VAD, batch jobs, model manager, privacy/wipe,
  and scheduler. Live audio-tap features remain V2.

## Blocked / open

- `glab` reports an invalid token; GitLab MCP and `git push` work.
- Terminal here has no Kitty/Sixel and no audio output; graphics encoders and
  audio paths need unit tests and external/real-device tests.
- 1080p software decode is below 30 fps on this phone; the I420 path helps, but
  laptop benchmarks are still required.
- GitHub exact-name conflict is unresolved; pick a distinctive name before
  public release.
- Embedded text subtitles need an FFmpeg demux/decode path in V1.
- `LICENSES.md` is not created yet.

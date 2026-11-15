# PROGRESS

## Status

- Phase: A complete; libVLC callback benchmark complete
- Next: Phase B research and FEATURES.md (waiting for approval)
- Blocked: none for the benchmark

## Decisions

- Playback engine: libVLC 3.0.24 (LGPL-2.1+), dynamically linked.
- Language: Go 1.24 + Vaxis. libVLC access via a thin in-repo cgo layer;
  `adrg/libvlc-go` is maintained but exposes no raw video callbacks.
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
  video-callback API, so the project will use a thin cgo layer.
- Built and ran the libVLC video-callback benchmark in Go 1.24 + cgo. Full
  results are in `docs/benchmarks/libvlc-callbacks.md`.

## Benchmark summary

- Go half-block render: ~1.1-1.5 ms/frame at 120x40; ASCII ~0.3 ms/frame.
  Pure render: 649-937 fps at 120x40, 402 fps at 200x50.
- 720p software decode: ~26 fps; 720p end-to-end: 27.2 fps at 120x40.
- 1080p software decode: ~15-17 fps; 1080p end-to-end: 17.0 fps at 120x40.
- The frame copy costs about 1 fps at 1080p; decode is the bottleneck.
- 120x40 half-block output is ~176 KB/frame (5.3 MB/s at 30 fps); 200x50 is
  ~366 KB/frame (11 MB/s at 30 fps). A low-bandwidth mode and frame dropping
  are required.
- Conclusion: continue with Go; do not switch to Rust. Validate 1080p on a
  mid-range laptop later.

## Next

- Phase B: research mpv/terminal players, terminal graphics support, user
  complaints, on-device AI models and licenses; verify name availability; write
  FEATURES.md.
- Then Phase C: ARCHITECTURE.md, TECH_STACK.md, DESIGN.md, ROADMAP.md,
  LICENSES.md, and the MVP build.

## Blocked / open

- `glab` reported an invalid token; the GitLab MCP works. Git push still needs a
  test.
- Terminal here has no Kitty/Sixel and no audio output; graphics encoders need
  unit tests and external terminals.
- 1080p software decode is below 30 fps on this phone. Target hardware must be
  benchmarked separately.

# libVLC video-callback benchmark

Date: 2026-11-15

## Question

Can libVLC 3.0.24 deliver decoded video frames through
`libvlc_video_set_callbacks` fast enough for a Go terminal renderer, and is Go
the right language for the frame path?

## Environment

- Debian 13 aarch64 under PRoot on Android
- 8 CPU cores, 5.6 GiB RAM (~1.2 GiB available at the time of the run)
- No `/dev/dri` (no GPU), no audio server, `/dev/snd` not accessible
- libVLC 3.0.24 (Debian package), plugins from `vlc-plugin-base`
- Go 1.24.4, cgo, gcc
- Software decoding forced with `--avcodec-hw=none`
- Video callbacks negotiated as `RV32`; one decode buffer plus one latest-frame
  buffer; callbacks copy each decoded frame into the latest-frame buffer
- Renderer: Go nearest-neighbour downscale plus ANSI truecolor half-block
  (`▀`, foreground/background per cell) or ASCII, output kept in memory
- Test media (downloaded to `/tmp`, not committed): Big Buck Bunny 10 s H.264,
  1920x1080 and 1280x720, about 5 MB each

## Results

### Decode-only

| Input | Window | Frames | FPS | CPU (one core = 100%) |
|---|---:|---:|---:|---:|
| 1080p, no frame copy | 5.00 s | 77 | 15.39 | 316% |
| 1080p, frame copy | 5.00 s | 72 | 14.39 | 316% |
| 720p, no frame copy | 5.00 s | 128 | 25.59 | 157% |

The 8 MB-per-frame copy costs about 1 fps at 1080p, so it is not the limiting
factor.

### Decode plus live half-block rendering, 120x40 cells

| Input | Window | Decoded | Rendered | FPS | Dropped | Avg render | Output | Pure render |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| 720p | 8.00 s | 218 | 218 | 27.24 | 0 | 1.921 ms | 176,344 B | 1.541 ms/frame (649 fps) |
| 1080p | 8.00 s | 136 | 136 | 17.00 | 0 | 1.517 ms | 176,203 B | 1.067 ms/frame (937 fps) |

### Pure render cost (snapshot frame, 50 iterations)

| Mode | Input | Cells | ms/frame | fps | Output/frame |
|---|---|---:|---:|---:|---:|
| Half-block | 720p | 80x24 | 0.653 | 1,532 | 70,463 B |
| Half-block | 1080p | 120x40 | 1.067 | 937 | 175,995 B |
| Half-block | 1080p | 200x50 | 2.485 | 402 | 366,330 B |
| ASCII | 720p | 120x40 | 0.296 | 3,383 | 4,840 B |

The 2 s end-to-end runs for the additional sizes are startup-skewed and are not
used for the fps conclusion; the pure-render numbers isolate the Go path.

## Conclusion

- **Go passes the render benchmark.** Half-block rendering at 120x40 costs about
  1-2 ms per frame, and ASCII costs about 0.3 ms. Go and cgo are not the
  bottleneck.
- **The callback API works.** Format negotiation, frame delivery, and shutdown
  are stable with libVLC 3.0.24. The 1080p output height is 1088, not 1080, due
  to decoder alignment; the renderer must use the negotiated dimensions.
- **Decode is the bottleneck on this device.** 1080p software H.264 runs at
  about 15-17 fps and uses >3 cores; 720p runs at about 26-27 fps. There is no
  GPU to accelerate it. mpv would use the same FFmpeg software decoder and would
  face the same limit here, so this is not a reason to switch from Go to Rust or
  from libVLC to mpv.
- **Hardware decoding with callbacks is constrained by libVLC.** The API
  documentation warns that hardware decoding is either disabled or requires a
  copy from device memory, and that subpictures are blended on the CPU. Target
  laptops should be benchmarked separately; this phone cannot represent them.
- **Terminal bandwidth matters.** 120x40 half-block is about 176 KB per frame,
  or 5.3 MB/s at 30 fps. 200x50 is about 366 KB per frame, or 11 MB/s at 30 fps.
  A low-bandwidth/SSH mode and frame dropping are required. Audio must remain
  the master clock, with video frames dropped when the terminal or decoder
  cannot keep up.

## Decision

Continue with Go + Vaxis + a thin in-repo cgo layer over the libVLC C API. Do not
switch to Rust based on this benchmark. Design the renderer to drop frames and
adapt resolution; target 720p smooth playback on this machine and validate 1080p
on a mid-range laptop.

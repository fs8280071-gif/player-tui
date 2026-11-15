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
- Original runs negotiated `RV32`; one decode buffer plus one latest-frame buffer;
  callbacks copy each decoded frame into the latest-frame buffer. Later runs add
  I420, RV16, fast-decode, and decoder-stat comparisons below.
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
- **The original RV32 path made decode look like the bottleneck.** 1080p software
  H.264 displayed at about 15-17 fps and used >3 cores; 720p displayed at about
  26-27 fps. Later I420 and decoder-stat runs show the decoder has headroom: the
  callback-vout conversion/copy/output path is the limiter. I420 raises 1080p
  display-only throughput to 33-38 fps without materially changing CPU. mpv would
  use the same FFmpeg software decoder and would face the same device limit, so
  this is not a reason to switch from Go to Rust or from libVLC to mpv.
- **Hardware decoding with callbacks is constrained by libVLC.** The API
  documentation warns that hardware decoding is either disabled or requires a
  copy from device memory, and that subpictures are blended on the CPU. Target
  laptops should be benchmarked separately; this phone cannot represent them.
- **Terminal bandwidth matters.** 120x40 half-block is about 176 KB per frame,
  or 5.3 MB/s at 30 fps. 200x50 is about 366 KB per frame, or 11 MB/s at 30 fps.
  A low-bandwidth/SSH mode and frame dropping are required. Audio must remain
  the master clock, with video frames dropped when the terminal or decoder
  cannot keep up.

## Fast-decode and thread matrix

`avcodec-fast`, `avcodec-skiploopfilter` (`skip4`), and `avcodec-threads=1` were
tested to see whether decoder headroom or callback load improved display fps.
Matrix: 720p30 and 1080p60, 6 s window, `-mode none`, RV32, default and explicit
libVLC options. `fast` is non-spec and error-prone.

| Input | Options | Display FPS | Decoder CPU | Player CPU | Copy CPU | Decode lag |
|---|---|---:|---:|---:|---:|---:|
| 720p30 | none | ~26 | 62-66% | 92-98% | 20-23% | 7.0-7.2 s |
| 720p30 | fast | ~24.3 | 49-50% | 79-83% | 17-18% | 10.8-10.9 s |
| 720p30 | skip4 | ~25.8 | 46-47% | 76-79% | 17% | 9.5-9.6 s |
| 720p30 | skip4+fast | ~24.8 | 43% | 73-76% | 16% | 9.5-9.6 s |
| 720p30 | threads=1 | ~26.5 | 47% | 88-91% | 21% | 6.9 s |
| 720p30 | threads=1+fast | ~24.6 | 36% | 64-68% | 15% | 9.8 s |
| 720p30 | threads=1+skip4 | ~26.2 | 35% | 63-67% | 14% | 8.9 s |
| 1080p60 | none | ~7.4-8.2 | 141-145% | 199-256% | 45-51% | ~48 s |
| 1080p60 | fast | ~7.9-8.4 | 108-110% | 191-241% | 38-40% | ~47 s |
| 1080p60 | skip4 | ~5.8-6.7 | 108-119% | 185-222% | 38-41% | ~53 s |
| 1080p60 | skip4+fast | ~8.2-8.5 | 99-106% | 178-207% | 35-37% | ~51 s |

One earlier 1080p default run reached 12.8 fps, but the repeatable range was
about 7-8 fps. The run-to-run spread is thermal/scheduler noise, not an option
effect.

Interpretation:

- No fast-decode combination produced a consistent display-fps gain.
- `fast` and `skip4` cut decoder CPU by roughly 20-25% and may help battery,
  thermal headroom, or weak CPUs, but they also increased decode lag in these
  runs. Do not enable them by default.
- `threads=1` helps 720p CPU but hurts 1080p. Keep libVLC's automatic thread
  selection.
- The bottleneck remains outside the decoder: callback vout conversion/copy and
  the terminal display path.

## Decoder and callback stats

`libvlc_media_get_stats` and callback timing confirm decoder headroom:

- 720p30 decoder: about 55-66 fps with `fast`/`skip4`; display about 20-28 fps.
- 1080p60 decoder: about 88-122 fps; display about 7-18 fps.
- Callback overhead: about 3-5 ms/frame on this device.
- 1080p display at 17 fps is about 59 ms/frame. The Go renderer accounts for only
  about 1-2 ms, so roughly 58 ms is callback copy, conversion, queueing, and
  terminal output, not Go drawing.

The decoder is not the reason 1080p is slow. The callback-vout path is.

## I420 vs RV32 vs RV16

RV32 allocates one packed 4-byte pixel per pixel and copies 8 MB/frame at 1080p.
I420 uses three planes and can be sampled at the terminal grid, avoiding a
full-resolution RGB conversion in the callback. This is the most important
result for the renderer design.

| Input | Chroma | Window | Frames | FPS | CPU | Avg display/frame |
|---|---|---:|---:|---:|---:|---:|
| 720p30 | RV32 | 6 s | 158 | 26.0 | 159% | 5.80 ms |
| 720p30 | I420 | 6 s | 161 | 26.5 | 124% | 4.81 ms |
| 1080p60 | RV32 | 6 s | 102 | 16.7 | 337% | 10.32 ms |
| 1080p60 | I420 | 6 s | 231 | 38.3 | 343% | 7.16 ms |
| 1080p60 | RV32 | 5 s | 95 | 18.8 | 339% | 9.55 ms |
| 1080p60 | I420 | 5 s | 208 | 41.2 | 344% | 6.62 ms |
| 1080p60 | RV32 | 8 s | 139 | 17.1 | 346% | 10.20 ms |
| 1080p60 | I420 | 8 s | 235 | 28.9 | 343% | 7.82 ms |

Summary: I420 raises 1080p60 display-only throughput from about 17 fps to
**33-38 fps** (best runs 38-41 fps) at similar CPU. At 720p30 it is roughly flat
(26.5 vs 26 fps) but cuts CPU from about 159% to 124%. RV16 (RGB565) was tested
as a middle ground: 720p ~22.8 fps at 172% CPU, 1080p ~23.6 fps at 292% CPU. It
did not beat I420 and loses color precision, so it is not the primary path.

Caveats:

- I420 measurements use `-mode none` because `bench_slice_latest` returns NULL
  for planar I420. There is no Go I420 sampler/renderer yet, so the display
  numbers are an upper bound, not an end-to-end terminal result.
- The 1080p I420 result still varies (28.9-41.2 fps). Treat ~30 fps as the
  planning target at 120x40 and re-measure on a laptop.
- The 1080p negotiated height is 1088 due to decoder alignment; use negotiated
  dimensions, not container dimensions.
- These are 5-10 s clips on a phone under PRoot with no audio. Long-run thermal
  behavior, audio clock interaction, and memory pressure are not covered.

## Updated conclusion

- Request **I420** from libVLC, sample Y/U/V at the terminal grid, and convert to
  RGB in Go. Fall back to RV32 if I420 negotiation fails.
- The audio-master clock, latest-frame mailbox, and drop policy remain required.
  Do not let video back-pressure stall audio.
- Use auto threads. Do not ship `avcodec-fast` as a default; expose it only as a
  troubleshooting/headroom option.
- The callback path is viable at 720p on this phone and should reach 1080p30 on a
  laptop. Real laptop benchmarks are still required before release.

## Decision

Continue with Go + Vaxis + a thin in-repo cgo layer over the libVLC C API. Do not
switch to Rust based on this benchmark. Design the renderer to drop frames and
adapt resolution; target 720p smooth playback on this machine and validate 1080p
on a mid-range laptop. Prefer I420 for the frame path; RV32 is the fallback.

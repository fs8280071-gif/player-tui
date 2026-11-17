# Render-mode benchmark

Date: 2026-11-16

## Setup

- Host: Debian proot on Android 13, 8 cores (2 x 2.0 GHz, 6 x 1.8 GHz)
- Go 1.24.4, libVLC 3.0.24
- Grid: 120x40 terminal cells (status row reserved; 120x34 video cells for
  both 16:9 and 1920x1088 sources), 90 rendered frames
- Clips: H.264, 8 s, `testsrc2`, generated with:

```bash
ffmpeg -f lavfi -i testsrc2=size=1280x720:rate=30:duration=8 \
  -c:v libx264 -preset ultrafast -crf 23 -pix_fmt yuv420p bench-720p.mp4
ffmpeg -f lavfi -i testsrc2=size=1920x1088:rate=30:duration=8 \
  -c:v libx264 -preset ultrafast -crf 23 -pix_fmt yuv420p bench-1080p.mp4
```

The decoder's callback buffer is padded (for example 1920x1090 for a 1920x1088
source). The player queries `libvlc_video_get_size` and exposes only the visible
frame dimensions, so the aspect ratio and sampling area use 1920x1088.

## Results

Medians of three runs. `render_p50` is the median time to render one frame into
the ANSI buffer; `render_fps` is `1 / render_p50`. `display_fps` is the
end-to-end rate while decoding and is kept for reference only.

### 1280x720

| Mode | render p50 | render fps | bytes/frame |
|---|---:|---:|---:|
| halfblock | 10.2 ms | 98 | 45.5 KB |
| quadrant | 12.9 ms | 77 | 45.0 KB |
| sextant | 14.5 ms | 69 | 45.5 KB |

### 1920x1088

| Mode | render p50 | render fps | bytes/frame |
|---|---:|---:|---:|
| halfblock | 17.2 ms | 58 | 42.9 KB |
| quadrant | 18.2 ms | 55 | 43.3 KB |
| sextant | 19.9 ms | 50 | 43.6 KB |

### With `--dither`

Medians of three runs.

| Clip | Mode | render p50 | render fps | bytes/frame |
|---|---|---:|---:|---:|
| 720p | halfblock | 11.8 ms | 85 | 140.8 KB |
| 720p | quadrant | 13.6 ms | 73 | 140.5 KB |
| 720p | sextant | 17.3 ms | 58 | 108.7 KB |
| 1080p | halfblock | 18.6 ms | 54 | 140.5 KB |
| 1080p | quadrant | 19.5 ms | 51 | 140.4 KB |
| 1080p | sextant | 21.2 ms | 47 | 107.8 KB |

## How to reproduce

```bash
player-tui bench bench-720p.mp4 --render halfblock --frames 90 --cols 120 --rows 40
player-tui bench bench-720p.mp4 --render quadrant  --frames 90 --cols 120 --rows 40
player-tui bench bench-720p.mp4 --render sextant   --frames 90 --cols 120 --rows 40

player-tui bench bench-1080p.mp4 --render halfblock --frames 90 --cols 120 --rows 40
player-tui bench bench-1080p.mp4 --render quadrant  --frames 90 --cols 120 --rows 40
player-tui bench bench-1080p.mp4 --render sextant   --frames 90 --cols 120 --rows 40
```

Run each command three times and compare the median `render_p50`. Keep the
clip, `--frames`, `--cols`, and `--rows` identical between modes. Use
`display_fps` for a sanity check only; this phone throttles and libVLC frame
delivery is noisy, so some end-to-end runs fall below the source frame rate even
when `render_fps` has headroom.

## Interpretation

- All three modes average every source pixel once per rendered frame. Changing
  the grid from 1x2 to 2x2 or 2x3 samples adds clustering and color decisions,
  not more source-area reads, so the cost increase is moderate.
- Quadrant measured 6-27% slower than halfblock; sextant measured 16-42%
  slower, depending on clip and run.
- Dithering roughly doubles to triples ANSI output because neighboring cells
  almost always change color. It costs about 0-3 ms per frame and about 60-100
  KB per frame at this grid.
- `auto` uses quadrant on normal terminals and halfblock for `TERM=dumb`,
  `TERM=linux`, or `PLAYER_TUI_ASCII=1`. Font coverage cannot be queried, so
  `PLAYER_TUI_RENDER` and `--render` remain the manual fallback.

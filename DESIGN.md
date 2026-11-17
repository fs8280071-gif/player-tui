# player-tui design

Status: Phase C design, 2026-11-16. This document fixes the screen layout, keys,
and visual language before the UI grows beyond the playback slice.

## Principles

- The video rectangle is sacred: no text chrome overlaps it in half-block mode.
- Every action is reachable from the keyboard; the mouse is optional.
- The player starts in under 200 ms and never waits for the network.
- Status is always visible: position, duration, state, title, volume, renderer.
- Restore the terminal on every exit path, including signals and panics.
- Respect `NO_COLOR`; degrade to 256/16 colors, then ASCII.

## Layout

### Wide terminal (200x50)

```text
┌ browser ──────────────┐┌ video / now playing ────────────────────────────────┐┌ queue ─────────────┐
│ ..                    ││                                                     ││ 3  next.mkv      │
│ Movies/               ││                                                     ││ 4  later.mp4     │
│  Arrival.mkv          ││              ▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀               ││                  │
│  Dune.mkv             ││              ▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀               ││                  │
│  Music/               ││              ▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀               ││                  │
│                       ││                                                     ││                  │
└───────────────────────┘└─────────────────────────────────────────────────────┘└──────────────────┘
 Dune.mkv                                    00:42:17 / 02:35:00   playing   42%   vol 80%   HB
```

### Compact terminal (80x24)

```text
┌ video ───────────────────────────────────────────────────────────────────────┐
│                                                                            │
│                        ▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀                          │
│                        ▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀                          │
│                        ▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀                          │
│                                                                            │
└────────────────────────────────────────────────────────────────────────────┘
 Dune.mkv                            00:42:17 / 02:35:00  playing  42%  vol 80%
```

The compact layout shows one pane at a time. `Tab` cycles player, browser, queue;
`1`/`2`/`3` jump directly. The status bar is always the last terminal row.

### Playback slice (what ships first)

```text
+--------------------------------------------------------------+
|                    half-block video area                     |
|                                                              |
|                                                              |
+--------------------------------------------------------------+
 sample.mkv            00:07 / 00:10   playing   [space] pause
```

No borders, no panes: the renderer owns rows `1..h-1`; row `h` is status. This
proves the engine, conversion, frame dropping, and terminal restore path.

## Panes

| Pane | Contents | Default width |
|---|---|---|
| Browser | Path breadcrumb, folders first, files with duration/size | 30% |
| Player | Video, title overlay only when paused, subtitle band below video | remaining |
| Queue | Up next, current item marked, shuffle/repeat icons | 22% |
| Status | Title, elapsed/total, state, percent, volume, renderer | 1 row |
| Command | `:` prompt or `Ctrl+P` palette, errors in red | 1-3 rows |

## Keys

| Key | Action |
|---|---|
| `space` | Play/pause |
| `left` / `right` | Seek -5 s / +5 s |
| `Shift+left` / `Shift+right` | Seek -30 s / +30 s |
| `up` / `down` | Volume +5 / -5 (browser: move selection) |
| `m` | Mute |
| `f` | Frame step when paused |
| `[` / `]` | Speed -0.1x / +0.1x |
| `s` | Cycle subtitle track |
| `a` | Cycle audio track |
| `,` / `.` | Subtitle delay -100 ms / +100 ms |
| `Tab` | Next pane |
| `Enter` | Play selection / enqueue with `Ctrl+Enter` |
| `n` / `p` | Next / previous in queue |
| `?` | Help overlay |
| `:` | Command line |
| `Ctrl+P` | Command palette |
| `q` | Quit (confirm if playing) |
| `Esc` | Close overlay, else quit |

Vim aliases: `j/k` selection, `h/l` seek, `g/G` top/bottom, `/` search.
All bindings live in one table so help and the palette stay in sync.

## Visual language

- Truecolor block rendering by default. `halfblock` uses `▀` with foreground =
  top pixel and background = bottom pixel; `quadrant` and `sextant` use 2x2 and
  2x3 block glyphs with two colors per cell.
- Exact area averaging is used when a sample covers more than one source pixel.
  The source aspect ratio drives the centered video rectangle; terminal cells
  are assumed to be twice as tall as they are wide.
- `--dither` applies 4x4 Bayer ordered dithering before cell quantization. It is
  off by default because it more than doubles ANSI output.
- `auto` picks quadrant for normal terminals and halfblock for `TERM=dumb`,
  `TERM=linux`, or `PLAYER_TUI_ASCII=1`; font coverage cannot be queried.
- Themes are TOML with a small fixed set of roles: `bg`, `fg`, `dim`, `accent`,
  `error`, `playing`, `paused`, `border`, `selection`.
- Ship Catppuccin Mocha, Gruvbox Dark, Nord, and High Contrast.
- Honor `NO_COLOR`, `COLORTERM`, and `TERM`; 256-color fallback quantizes the
  same palette, 16-color uses named ANSI colors.
- Paused shows a centered `PAUSED` label outside the last decoded frame; it is
  not burned into the video rectangle.
- Errors are one line in the status bar plus an actionable second line; never a
  modal that hides the player.

## Typography and safety

- All names, titles, and subtitle text pass through the terminal sanitizer:
  strip C0/C1, ESC, OSC, DCS, CSI sequences; collapse newlines in titles.
- Nerd Font icons are optional and always paired with ASCII fallback glyphs.
- Wide and combining characters are measured by cell width, not bytes.
- Subtitle text wraps by cell width, maximum three lines in the compact layout.

## Frame scheduling

- Render loop targets 30 fps by default, 60 fps maximum for small terminals.
- The mailbox keeps the newest frame; late frames are dropped.
- The status bar redraws at most 4 times per second.
- On resize, recompute the video rectangle, clear the screen, and continue from
  the last frame; never queue stale sizes.

## Open design work

- Pane borders vs. borderless layout in narrow terminals.
- Kitty/Sixel overlay strategy for subtitles and paused labels.
- Command palette grouping and fuzzy-match highlighting.
- Screen-reader plain mode: one event per line, no cursor movement.

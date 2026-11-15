# player-tui - Feature Inventory and Build Plan

Status: Phase B research deliverable
Date: 2026-11-15
Scope: VLC-class terminal media player for local and streamed audio/video, with a
library, subtitles, terminal-native video rendering, and optional offline AI.

This document is deliberately broader than the first release. It records what
users of comparable players already get, where terminal players are weak, and
which features are worth building first. It is not a promise that every V2 item
will ship.

## 1. Executive summary

- The terminal player market is split between mature audio players (`cmus`,
  `musikcube`, `termusic`, `kew`, `ncmpcpp`, `CLIAMP`) and video/image tools
  (`mpv`, `mpvc`, `ncplayer`, `SparkPlayer`, `chafa`, `Yazi`). No maintained
  terminal player combines VLC-class format breadth, a real media library,
  synchronized terminal video, first-class subtitles, and offline AI.
- The strongest immediate opportunity is not another music TUI. It is a
  **single terminal application that can open almost anything, render video in
  the terminal, manage a library and queue, handle subtitles properly, and work
  over SSH without a GUI or cloud service**.
- The strongest differentiator is **local AI that is useful before playback and
  during playback**: offline subtitles, translation, semantic search, chapters,
  and "ask the video", all with visible citations and timestamp jumps.
- MVP should be: engine lifecycle, audio-master clock and A/V sync, Kitty/Sixel/
  half-block/ASCII rendering, queue and playlists, external subtitles, themes,
  terminal restore and escape safety, a file browser, a CLI, and `doctor`.
- AI v1 should follow MVP: offline ASR subtitles for local files, VAD, batch
  generation, a model manager, and privacy controls. Live audio-tap features
  such as real-time translation and visualizers are V2 because libVLC 3 has no
  public visualization tap and audio callbacks replace normal output.
- Measured baseline on the phone: 720p30 at 120x40 reaches about 27 fps with
  I420; 1080p60 reaches 33-38 fps with I420 in display-only mode. The remaining
  bottleneck is callback-vout conversion/copy, not the decoder. Laptop targets
  must be re-measured before release.

## 2. Method and evidence

- Compared official feature pages, manuals, source code, headers, and release
  notes. No competitor claims are based only on memory.
- Demand is recorded as High/Medium/Low with evidence, not invented counts.
  Star counts change; use them as directional signals.
- `FEATURES.md` was checked against the project constraints: free/open-source,
  no VLC branding, dynamic libVLC linkage, no paid cloud AI, and terminal-only
  UI.
- The local environment has no audio server, no GPU, and no Kitty/Sixel
  terminal. Terminal matrix claims come from official documentation and must be
  smoke-tested on real devices before release.
- Research date: 2026-11-15.

Legend:

- **Priority**: MVP = first usable release; V1 = first public/parity release;
  V2 = later/experimental.
- **Demand**: High = baseline expectation or repeated requests/plugins;
  Medium = common in comparable players or clear niche; Low = rare/experimental.
- **Difficulty**: Easy = days; Medium = 1-2 weeks; Hard = multi-week or risky.
- **Who has it**: examples, not an exhaustive list. "None found" means no
  maintained terminal player found with the feature as a first-class feature.

## 3. Competitive landscape

| Player | Type | Video | Art/lyrics | Subtitles | Library | AI | Notes |
|---|---|---|---|---|---|---|---|
| [VLC][VLC] | GUI media player | Yes | Art | Yes | Basic | Demo only | Format breadth and subtitle baseline; no terminal UI. |
| [mpv][MPV] | CLI/GUI video player | Yes, incl. terminal `tct`/`sixel`/`kitty` | No | Yes | No | No | Best terminal video baseline; v0.42.0 reworked terminal outputs [MPV042]; scripting via JSON IPC; image output is not synchronized with other terminal output. |
| [mpvc][MPVC] | mpv + POSIX sh + fzf | Yes | No | Via mpv | No | No | Tiny, composable, fuzzy file picking. |
| [SparkPlayer][SPARK] | Rust terminal media player | Yes | Yes | Partial/verify | No | No | Album art, 11 visualizers, graphics and half-block. |
| [CLIAMP][CLIAMP] | Go terminal music player | No | Partial | No | Providers | No | Winamp-like, EQ, plugins, many providers; about 4.6k stars. |
| [termusic][TERMUSIC] | Rust terminal music player | No | Yes | No | Yes | No | YouTube/NetEase, lyrics/art, multiple backends. |
| [kew][KEW] | C terminal music player | No | Yes | No | Yes | No | Gapless, cover art, visualizer, synced lyrics, ReplayGain, crossfade. |
| [cmus][CMUS] | C terminal music player | No | Via plugins | No | Yes | No | Mature queue/filters/streaming; art and lyrics are external. |
| [musikcube][MUSIKCUBE] | C++ terminal music system | No | Yes | No | Yes | No | Library server, plugins, high-resolution audio. |
| [ncmpcpp][NCMPCPP] | MPD client | No | Lyrics/visualizer | No | MPD | No | Tags, playlists, search, visualizer, lyrics. |
| [ncplayer][NCPLAYER] | notcurses media player | Yes | No | No | No | No | Direct cell blitting, many blitter modes. |
| [chafa][CHAFA] | Image-to-terminal converter | Images/GIF | N/A | N/A | No | No | ANSI/Unicode/Sixel/Kitty output; passthrough for tmux/screen. |
| [Yazi][YAZI] | Rust file manager | Preview | Preview | No | Files | No | Multi-protocol image/video preview, plugins. |
| [Kodi][KODI] | GUI media center | Yes | Yes | Yes | Yes | No | Library and subtitle baseline; no terminal UI. |

Demand synthesis:

- **High**: playback, format breadth, subtitles, terminal video, queue/library,
  themes, keyboard-first CLI, and terminal safety.
- **Medium**: album art, lyrics, visualizers, resume, equalizer/ReplayGain,
  MPRIS/media keys, library search.
- **Low**: disc playback, recording/transcoding, network libraries, niche AI.
- **Strong signals**: CLIAMP at roughly 4.6k stars, SparkPlayer and
  [`cmus-cover-art`][CMUSCOVER] at roughly 55 and 53 stars respectively, and
  recurring requests for terminal video and cover art. GitHub issue searches for
  cmus lyrics/art are sparse (two loosely related issues, [CMUS1191] and
  [CMUS443]), so do not claim a large quantified demand from issues alone.
- **Largest open gap**: no terminal player ships local AI subtitles, semantic
  search, or video Q&A. VLC has demonstrated offline AI subtitles but has not
  shipped them in stable VLC 3.0.24; that release was primarily a security update
  with 50+ VLC/codec fixes ([VLCSEC][VLCSEC], [VLCAI][VLCAI]).

## 4. Feature inventory

### 4.1 Playback

| ID | Feature | Description | Who has it | Demand | Priority | Difficulty | Library / design notes | Source |
|---|---|---|---|---|---|---|---|---|
| P01 | Open local files, folders, URLs | Accept paths, directories, playlists, and network URLs; enqueue or play now. | VLC, mpv, mpvc | High | MVP | Easy | libVLC media creation; canonicalize paths; reject NUL/control characters. | [VLC][VLC] |
| P02 | Playback clock and A/V sync | Audio-master clock, monotonic interpolation, pause/seek re-anchor, drift correction. | VLC, mpv, Kodi | High | MVP | Hard | `libvlc_media_player_get_time`/position plus Go monotonic clock; drop late video; no-audio fallback. | [LIBVLC][LIBVLC] |
| P03 | Engine lifecycle and transport | Explicit state machine: idle, opening, buffering, playing, paused, seeking, ended, error; play/pause/stop/next/prev. | VLC, mpv | High | MVP | Medium | One owner goroutine; idempotent shutdown; event-driven state; no callback-to-UI blocking. | [LIBVLC][LIBVLC] |
| P04 | Seek, time display, chapters | Seek by seconds/percent, show elapsed/remaining, list and jump to chapters. | VLC, mpv, Kodi | High | MVP | Medium | `libvlc_media_player_set_time`; chapter API; clamp to duration; debounce scrub. | [LIBVLC][LIBVLC] |
| P05 | Queue, shuffle, repeat | Add/remove/reorder queue, play next/previous, shuffle, repeat one/all. | VLC, mpv, CLIAMP | High | MVP | Medium | In-memory queue plus persistence; preserve current item across restart. | [VLC][VLC] |
| P06 | Playlist import/export | Read/write M3U/M3U8, PLS, XSPF; relative paths and URLs. | VLC, mpv, Kodi | High | V1 | Easy | Prefer libVLC playlist parsing; validate and rewrite paths safely. | [VLC][VLC] |
| P07 | Resume and remember | Remember position per media, last volume, last track, and optional resume prompt. | VLC, Kodi | High | V1 | Easy | SQLite keyed by canonical path + size + mtime; ignore streams. | [VLC][VLC] |
| P08 | Gapless and crossfade | Gapless album playback; optional crossfade for music. | cmus, kew, VLC | Medium | V1 | Medium | Gapless via libVLC playlist; crossfade may need two players or VLC sout; hard to make seamless. | [CMUS][CMUS] |
| P09 | Network streams | HTTP(S), HLS, DASH, RTSP, RTP, RTMP, SRT; reconnect and buffer status. | VLC, mpv | High | V1 | Easy | libVLC handles protocols; expose cache/buffer and errors; no custom DRM. | [VLC][VLC] |
| P10 | Discs, devices, network shares | DVD/Blu-ray/CD-DA, SMB/NFS/SFTP, UPnP/DLNA where libVLC supports it. | VLC | Medium | V2 | Medium | libVLC modules; no CSS/DRM circumvention; platform-dependent. | [VLC][VLC] |
| P11 | Record and convert | Record stream to file; simple remux/transcode presets. | VLC | Low | V2 | Hard | libVLC sout or external FFmpeg; not core to MVP; document limits. | [VLC][VLC] |
| P12 | Screenshot | Save current video frame as PNG/JPEG. | VLC, mpv | Medium | V1 | Easy | `libvlc_video_take_snapshot`; path safety; no overwrite without flag. | [LIBVLC][LIBVLC] |
| P13 | Speed, frame-step, A-B loop | Playback rate, single-frame step, loop a selected range. | VLC, mpv | Medium | V1 | Medium | `set_rate`; next-frame API; custom A-B loop over position polling. Pitch correction is V2. | [LIBVLC][LIBVLC] |
| P14 | Track selection and sync delays | Audio/video/subtitle track menus; audio and subtitle delay adjustment. | VLC, mpv | High | MVP | Easy | `libvlc_video_get_spu*`, track descriptions, `set_spu_delay`, audio delay API. | [LIBVLC][LIBVLC] |
| P15 | Error recovery, headless, buffer | Recover from decode/network errors; audio-only/headless mode; bounded stream buffer. | VLC, mpv | High | V1 | Medium | Error events; restart policy; `--no-video`; cache options. | [LIBVLC][LIBVLC] |

### 4.2 Formats

| ID | Feature | Description | Who has it | Demand | Priority | Difficulty | Library / design notes | Source |
|---|---|---|---|---|---|---|---|---|
| F01 | Containers | MP4/MOV, MKV/WebM, AVI, MPEG-TS/PS, FLV, OGG, ASF/WMV, and more. | VLC, mpv, Kodi | High | MVP | Easy | libVLC/FFmpeg; show container and track details in `doctor`/status. | [VLC][VLC] |
| F02 | Video codecs | H.264, HEVC, VP8/VP9, AV1, MPEG-2/4, ProRes, DNxHD, Theora, APV where available. | VLC, mpv, Kodi | High | MVP | Easy | libVLC; VLC 3.0.24 bundles FFmpeg 8.1.2 and adds APV/ATRAC support. | [VLC3024][VLC3024] |
| F03 | Audio codecs | AAC, MP3, Opus, Vorbis, FLAC, ALAC, WAV, AC-3/E-AC-3, DTS, TrueHD, APE, WavPack, ATRAC3/9. | VLC, mpv, Kodi | High | MVP | Easy | libVLC; passthrough options depend on audio device; terminal app cannot guarantee HD passthrough. | [VLC3024][VLC3024] |
| F04 | HDR, 10-bit, color | Detect HDR/10-bit, tone-map to terminal SDR, preserve correct matrix/range. | VLC, mpv, Kodi | Medium | V1 | Hard | libVLC filters plus Go tone mapping; terminal is 8-bit sRGB-ish; no true HDR output. | [VLC][VLC] |
| F05 | Images and cover art | JPEG/PNG/WebP/GIF/BMP; embedded art and sidecar images. | VLC, SparkPlayer, Yazi | High | V1 | Easy | stdlib + `golang.org/x/image`; cache decoded images; render via active protocol. | [XIMAGE][XIMAGE] |
| F06 | Playlist formats | M3U/M3U8, PLS, XSPF, ASX; CUE for disc images later. | VLC, mpv | High | V1 | Easy | libVLC or custom parser; fuzz parser; canonicalize. | [VLC][VLC] |
| F07 | Streaming/disc/share protocols | HLS, DASH, RTSP, RTP, RTMP, SRT, SMB, NFS, SFTP, UPnP/DLNA, DVD/Blu-ray/CD-DA. | VLC, mpv | High | V1 | Medium | libVLC modules; network shares and discs are V2 in practice. | [VLC][VLC] |
| F08 | Capture and live sources | Webcam, microphone, desktop capture where platform permits. | VLC | Low | V2 | Hard | libVLC capture modules; permissions and platform APIs; not MVP. | [VLC][VLC] |
| F09 | DRM and CSS policy | Do not implement DRM circumvention; use system libraries only if present and legal. | VLC | Medium | MVP | Easy | Document policy; fail with a clear message; no bundled libdvdcss. | [VLC][VLC] |

### 4.3 Subtitles

| ID | Feature | Description | Who has it | Demand | Priority | Difficulty | Library / design notes | Source |
|---|---|---|---|---|---|---|---|---|
| S01 | External subtitle files | Load sidecar SRT/VTT/ASS/SSA, choose encoding, attach to current media. | VLC, mpv | High | MVP | Easy | `go-astisub` for SRT/VTT; custom ASS text extractor; `libvlc_media_player_add_slave` optional. | [ASTISUB][ASTISUB] |
| S02 | Embedded text subtitles | Extract text tracks from MKV/MP4 and render them as terminal text. | VLC, mpv | High | V1 | Hard | libVLC has no text callback; use FFmpeg demux/decode path or optional `ffmpeg` subprocess. | [LIBVLC][LIBVLC] |
| S03 | Bitmap subtitles | PGS, VobSub, CEA-608/708 where libVLC can blend them into video frames. | VLC, mpv | Medium | V1 | Medium | Let libVLC blend bitmap subs; terminal text overlay cannot position them separately. | [VLC3024][VLC3024] |
| S04 | Track selection and delay | Cycle/select subtitle tracks; positive/negative delay; remember per media. | VLC, mpv | High | MVP | Easy | SPU description + `set_spu_delay`; persist track choice. | [LIBVLC][LIBVLC] |
| S05 | Terminal text styling | Wrap, position, background box, bold/italic/color, max lines; safe ANSI only. | mpv, VLC | High | MVP | Medium | Custom renderer; strip untrusted escapes; map ASS colors/italics approximately. | [MPV][MPV] |
| S06 | Encoding and normalization | UTF-8/UTF-16/legacy detection; normalize line endings and BOM; reject invalid sequences. | VLC, mpv | High | MVP | Medium | `golang.org/x/text/encoding`; chardet fallback; fuzz. | [VLC][VLC] |
| S07 | Dual, SDH, forced | Two subtitle tracks at once; SDH/sound-event text; forced-only mode. | VLC, mpv, Kodi | Medium | V1 | Medium | Two text renderers; bitmap tracks cannot dual-render; forced flag from track metadata. | [VLC][VLC] |
| S08 | Sidecar search and convert | Find matching local subtitle files; convert SRT/VTT/ASS; export cleaned text. | VLC, mpv | Medium | V1 | Medium | Local only; no paid subtitle APIs; convert via `go-astisub`/custom. | [ASTISUB][ASTISUB] |
| S09 | ASS effects and karaoke | Approximate ASS positioning/animations/karaoke; do not promise libass pixel parity. | VLC, mpv | Low | V2 | Hard | `libass` optional for bitmap rendering; terminal text loses positioning/animation. | [LIBASS][LIBASS] |
| S10 | Safe subtitle escaping | Never pass raw subtitle bytes to the terminal; allow only generated SGR. | Required | High | MVP | Medium | Escape C0/C1, OSC, DCS, CSI; cap line length; fuzz parser. | [CWE150][CWE150] |

Subtitle design note: use a hybrid path. For **text** subtitles, parse and render
in the terminal so they can be styled, wrapped, translated, and searched. For
**bitmap** subtitles, let libVLC blend them into the callback frame. Do not
enable both paths for the same track. Embedded text extraction is the hard part:
libVLC 3 exposes track descriptions and delays but no text callback, so V1 needs
an FFmpeg demux/decode path. MVP can ship external text subtitles and bitmap
subtitles, with embedded text extraction clearly marked V1.

### 4.4 Audio

| ID | Feature | Description | Who has it | Demand | Priority | Difficulty | Library / design notes | Source |
|---|---|---|---|---|---|---|---|---|
| A01 | Volume and mute | Per-player volume, mute toggle, persisted level, no clipping surprises. | VLC, mpv, CLIAMP | High | MVP | Easy | `libvlc_audio_set_volume`; soft ramp on pause/resume; show percent. | [LIBVLC][LIBVLC] |
| A02 | Device selection and hotplug | Choose output device, remember by name, handle device disappearance. | VLC, mpv | High | V1 | Medium | `libvlc_audio_output_device_enum`; poll hotplug; fall back to default with a toast. | [LIBVLC][LIBVLC] |
| A03 | Equalizer, filters, downmix | 10-band EQ, compressor, stereo widening, normalization, channel downmix. | VLC, CLIAMP, kew | Medium | V1 | Medium | libVLC equalizer API and `--audio-filter`; presets; persist per output. | [LIBVLC][LIBVLC] |
| A04 | ReplayGain and normalization | Track/album ReplayGain, peak limiting, optional loudness normalization. | cmus, kew, VLC | Medium | V1 | Medium | Read tags; apply gain via volume chain or filter; avoid double normalization. | [CMUS][CMUS] |
| A05 | Audio tap, visualizer, spectrum | Tap PCM for FFT/oscilloscope and optional AI live audio. | SparkPlayer, CLIAMP, kew | Medium | V2 | Hard | libVLC 3 has no public visualization tap; audio callbacks replace output. Needs custom sink or offline decode. | [LIBVLC][LIBVLC] |
| A06 | Spatial audio and low latency | Ambisonics/3D audio where supported; low-latency output mode. | VLC, mpv | Low | V2 | Hard | libVLC/device dependent; terminal UI only exposes controls. | [VLC][VLC] |
| A07 | Audio description and loudness AI | Generate/duck descriptions; level dialogue vs effects. | None stable | Low | V2 | Hard | AI V2; local TTS and DSP; licensing review. | [PIPER][PIPER] |
| A08 | Source separation and karaoke | Remove vocals, isolate dialogue, karaoke mode. | None stable | Low | V2 | Hard | Demucs worker; CPU-heavy; model license/weights review. | [DEMUCS][DEMUCS] |

Audio design note: do **not** promise a real-time visualizer in MVP. libVLC 3
has no visualization callback in the public header, and `libvlc_audio_set_callbacks`
replaces normal audio output. A visualizer requires either a custom audio sink
(and cross-platform output code) or offline/parallel decode. Treat it as V2 and
gate it behind an explicit "audio tap" mode. The same tap is the foundation for
live ASR/translation, so design the ring buffer and clock once.

### 4.5 Video rendering

| ID | Feature | Description | Who has it | Demand | Priority | Difficulty | Library / design notes | Source |
|---|---|---|---|---|---|---|---|---|
| V01 | Capability detection | Detect Kitty, Sixel, iTerm2, truecolor, cell size, multiplexer, and fallbacks. | Yazi, chafa | High | MVP | Medium | DA1/XTSMGRAPHICS/env/query; `doctor` prints results; never assume. | [YAZI][YAZI] |
| V02 | Renderer ladder | Automatic choice with manual override: Kitty, iTerm2, Sixel, half-block, ASCII/ANSI. | mpv, Yazi, chafa | High | MVP | Medium | `--render=auto/kitty/sixel/iterm2/halfblock/ascii`; `--no-graphics`. | [MPVT][MPVT] |
| V03 | Kitty graphics | Raw RGB image, chunked base64, image IDs, placement, delete, resize, animation later. | Kitty, Ghostty, WezTerm, iTerm2, mpv | High | MVP | Medium | Custom encoder; zlib optional; handle tmux/Zellij passthrough; cleanup on exit. | [KITTY][KITTY] |
| V04 | Sixel | Palette quantization, band encoding, scroll region, cleanup, tmux/screen passthrough. | xterm, WezTerm, Windows Terminal, chafa | High | MVP | Hard | Custom encoder or [`go-sixel`][GOSIXEL]; truecolor to 256 palette; test disposal and scroll. | [AWSY][AWSY] |
| V05 | iTerm2 inline images | OSC 1337 inline image with width/height and aspect preservation. | iTerm2, WezTerm | Medium | V1 | Medium | Custom encoder; size limits; cleanup; fall back to Kitty/Sixel/half-block. | [ITERM2][ITERM2] |
| V06 | Half-block truecolor | Two vertical pixels per cell using `▀`; 24-bit color with 256/16 fallback. | mpv, SparkPlayer, chafa | High | MVP | Easy | Custom renderer; batch rows; no per-cell writes; test odd heights. | [MPVT][MPVT] |
| V07 | ASCII/ANSI fallback | Luminance ramp with optional 256/16 colors; readable at low bandwidth. | mpv, chafa, ncplayer | High | MVP | Easy | Custom; edge enhancement optional; `--ascii-ramp`; no-color mode. | [CHAFA][CHAFA] |
| V08 | Scaling, crop, letterbox, aspect | Fit video to cells, preserve aspect ratio, crop/zoom, pixel aspect ratio. | mpv, VLC | High | MVP | Medium | Go scaler (`x/image/draw`) at target grid; avoid full-res copies. | [XIMAGE][XIMAGE] |
| V09 | Frame scheduler and dropping | Render at terminal cadence; drop late frames; never block audio; bounded queues. | mpv, SparkPlayer | High | MVP | Hard | Latest-frame mailbox; audio clock; drop thresholds; metrics in status. | [BENCH](docs/benchmarks/libvlc-callbacks.md) |
| V10 | Adaptive quality and low bandwidth | Reduce resolution/cells/fps/colors under load or over SSH. | mpv, chafa | High | V1 | Medium | Auto-quality ladder; manual low-bandwidth mode; show current mode. | [MPVT][MPVT] |
| V11 | Cell size and synchronized output | Query cell pixels; use DEC 2026 synchronized updates where supported. | Terminal-dependent | High | MVP | Medium | CSI 14/16 t; fallback 8x16; guard with capability cache. | [XTERMCTL][XTERMCTL] |
| V12 | Cleanup, scroll, resize | Delete images, clear stale cells, redraw on resize, avoid scrollback corruption. | Kitty, WezTerm, chafa | High | MVP | Medium | Cursor save/restore; scroll region; debounce resize; idempotent cleanup. | [KITTY][KITTY] |
| V13 | HDR, 10-bit, gamma | Tone-map HDR/10-bit to terminal color; correct gamma/matrix/range. | VLC, mpv | Medium | V2 | Hard | libVLC tone mapping + Go color conversion; no true HDR terminal support. | [VLC][VLC] |
| V14 | Overlays and snapshot | Draw subtitles/progress/status over video; save current frame. | VLC, mpv | High | V1 | Medium | Text overlay in cell grid; snapshot via libVLC; no GUI. | [LIBVLC][LIBVLC] |
| V15 | Video filters and hardware decode | Deinterlace, rotate, crop, denoise; optional hardware decode. | VLC, mpv | Medium | V2 | Hard | libVLC filters/options; callback path may force software; target laptops. | [VLC][VLC] |
| V16 | Multiplexer and SSH mode | Work under tmux/Zellij/screen and over SSH; choose passthrough or fallback. | Yazi, chafa, mpv | High | V1 | Medium | Detect `TMUX`, `ZELLIJ`, `STY`, `SSH_*`; low-bandwidth fallback. | [CHAFA14][CHAFA14] |

#### Terminal graphics matrix

This matrix is for planning and `doctor`; every row needs a real-device smoke
test before release. Capability can vary by version, build options, and
multiplexer configuration.

| Host | Kitty graphics | Sixel | iTerm2 inline | Notes |
|---|---|---|---|---|
| [Kitty][KITTY] | Yes | No | No | Reference implementation; image IDs, placement, zlib, animation. |
| [Ghostty][GHOSTTY] | Yes | No | No | 1.4.0 (2026-10-27) claims 100% Kitty graphics support, including animation and relative placement. |
| [WezTerm][WEZTERM] | Yes | Experimental | Yes | Sixel has been experimental since 2020; verify current build. |
| [iTerm2][ITERM2] | Yes | Not documented | Yes | Supports its own inline images and Kitty graphics. |
| [Windows Terminal][WT] | No | Yes | No | Sixel from 1.22.10352.0; classic conhost/Windows Console is unsupported. |
| [xterm][XTERM] | No | Yes | No | Sixel support depends on build; historically opt-in via `-ti vt340`. |
| tmux | Passthrough | Official support | Passthrough | Kitty passthrough needs `allow-passthrough`; Sixel is documented by Are We Sixel Yet. |
| [Zellij][ZELLIJ] | Native per pane | If host supports | No | `support_kitty_graphics_protocol`; re-advertises Sixel only when host supports it. |
| GNU Screen | Passthrough | Passthrough | Passthrough | Best-effort; chafa 1.14 supports Kitty/Sixel passthrough for screen. |

Renderer selection order: **Kitty -> iTerm2 -> Sixel -> half-block -> ASCII**.
Prefer Kitty on iTerm2 because the same encoder covers Kitty, Ghostty, WezTerm,
and iTerm2. Never emit graphics before raw mode/alternate screen is active, and
always send cleanup on normal exit, signal, panic, and resize.

### 4.6 Library

| ID | Feature | Description | Who has it | Demand | Priority | Difficulty | Library / design notes | Source |
|---|---|---|---|---|---|---|---|---|
| L01 | File and folder browser | Navigate files/directories, preview metadata, play or enqueue with the keyboard. | Yazi, mpvc, VLC | High | MVP | Medium | Vaxis UI; canonical paths; show hidden files toggle; Unicode-safe. | [YAZI][YAZI] |
| L02 | Scan, metadata, cover art | Recursive scan, tags, duration, codecs, embedded/sidecar art. | cmus, musikcube, termusic | High | V1 | Medium | `dhowden/tag` first; TagLib optional; libVLC parse for duration/codecs. | [DHOWDEN][DHOWDEN] |
| L03 | SQLite index, FTS, fuzzy search | Persistent library, full-text and fuzzy search, fast filters. | cmus, musikcube, ncmpcpp | High | V1 | Medium | `modernc.org/sqlite` + FTS5; `sahilm/fuzzy` for short queries. | [SQLITE][SQLITE] |
| L04 | Filters and sort | Artist/album/genre/year/format/folder filters; sort by title/date/duration. | cmus, ncmpcpp, musikcube | High | V1 | Easy | SQL predicates; saved views; no query strings from untrusted input. | [CMUS][CMUS] |
| L05 | Incremental watcher | Detect added/changed/removed files without full rescans. | musikcube, cmus | Medium | V1 | Medium | `fsnotify` with debounce; ignore symlink loops; rescan fallback. | [FSNOTIFY][FSNOTIFY] |
| L06 | Playlist and queue persistence | Save/load playlists and queue state; crash-safe. | VLC, cmus, CLIAMP | High | V1 | Easy | SQLite transactions; stable IDs; relative paths where possible. | [VLC][VLC] |
| L07 | Play counts and ratings | Track play count, last played, rating, skip count. | cmus, musikcube, kew | Medium | V2 | Easy | SQLite; opt-in; never upload. | [CMUS][CMUS] |
| L08 | Smart playlists | Rule-based playlists (genre, rating, recently added, not played). | musikcube, cmus | Medium | V2 | Medium | Query DSL compiled to SQL; preview count before save. | [MUSIKCUBE][MUSIKCUBE] |
| L09 | Duplicates and missing files | Detect duplicate tracks; mark missing files without deleting history. | musikcube | Medium | V1 | Medium | Hash only on demand; missing state; relink by size/tag. | [MUSIKCUBE][MUSIKCUBE] |
| L10 | Network library | Browse SMB/NFS/SFTP/UPnP libraries. | VLC, Kodi | Low | V2 | Hard | libVLC modules; credentials handling; no custom auth store. | [VLC][VLC] |
| L11 | Tag editing | Edit common tags and cover art. | musikcube, EasyTAG | Low | V2 | Hard | TagLib write path; backup; atomic replace; format-specific limits. | [TAGLIB][TAGLIB] |
| L12 | Local-only, privacy, stats | No telemetry; library stays local; index wipe and stats. | cmus, musikcube | High | MVP | Easy | Local DB; `--privacy`; wipe command; no network calls by default. | [CMUS][CMUS] |

Library schema sketch: `tracks`, `albums`, `artists`, `covers`, `playlists`,
`playlist_items`, `queue`, `bookmarks`, `settings`, `ai_jobs`, `transcripts`,
`segments`, `embeddings`. Use migrations, WAL, and stable IDs. Keep media files
read-only unless the user explicitly edits tags.

### 4.7 UI

| ID | Feature | Description | Who has it | Demand | Priority | Difficulty | Library / design notes | Source |
|---|---|---|---|---|---|---|---|---|
| U01 | Adaptive layout | Video/now-playing center, browser left, queue right, status bottom; collapse for narrow terminals. | All | High | MVP | Medium | Vaxis; minimum 80x24; tabs on narrow; remembered pane sizes. | [VAXIS][VAXIS] |
| U02 | Browser, queue, now-playing panes | Keyboard-focused panes with clear focus indicators. | cmus, CLIAMP, Yazi | High | MVP | Medium | Single focus model; no mouse required; fast redraw. | [CLIAMP][CLIAMP] |
| U03 | Status, seek, chapters | Title, time, duration, volume, codec, fps, dropped frames, seek bar, chapters. | mpv, VLC | High | MVP | Medium | Compact and expanded modes; no layout shift while playing. | [MPV][MPV] |
| U04 | Command line and palette | `:` command line plus fuzzy command palette. | mpvc, Yazi, Vim | High | MVP | Medium | Parser with quotes/escapes; completion; command registry. | [MPVC][MPVC] |
| U05 | Help and keymap | Context help, key cheat sheet, vim/arrow presets, remapping. | CLIAMP, ncmpcpp | High | MVP | Easy | Keymap TOML; conflict detection; `?` overlay. | [CLIAMP][CLIAMP] |
| U06 | Themes and NO_COLOR | TOML themes, truecolor/256/16 fallback, `NO_COLOR`, high contrast. | CLIAMP, kew, ncmpcpp | High | MVP | Easy | Semantic color roles; never hard-code ANSI; validate contrast. | [NOCOLOR][NOCOLOR] |
| U07 | Icons, ASCII, Unicode | Nerd Font icons with ASCII fallback; wide-char/grapheme-safe layout. | Yazi, CLIAMP | High | MVP | Medium | `--icons=auto/nerd/ascii/none`; Vaxis width handling; test CJK/emoji. | [VAXIS][VAXIS] |
| U08 | Toasts and notifications | Non-blocking errors, track changes, AI job status. | CLIAMP, Yazi | Medium | V1 | Easy | Queue with TTL; never overwrite critical status; log full error. | [CLIAMP][CLIAMP] |
| U09 | Album art | Embedded/sidecar art in now-playing pane; protocol-aware rendering. | SparkPlayer, cmus plugin, termusic | High | V1 | Medium | Decode once; cache; ASCII fallback; never block audio. | [SPARK][SPARK] |
| U10 | Lyrics | Embedded/sidecar `.lrc`, synced highlighting, offset. | termusic, kew, ncmpcpp | High | V1 | Medium | LRC parser; UTF-8; offset; no paid lyrics API. | [TERMUSIC][TERMUSIC] |
| U11 | Visualizer pane | Spectrum/oscilloscope display when audio tap is available. | SparkPlayer, CLIAMP, kew | Medium | V2 | Medium | Depends on A05; frame budget; disable in low-bandwidth mode. | [SPARK][SPARK] |
| U12 | Menus, search, settings | Track/chapter menus, incremental search, settings editor. | VLC, Yazi | High | V1 | Medium | Keyboard-first; no modal traps; validate on save. | [YAZI][YAZI] |
| U13 | Title and resize | Terminal title with track/progress; clean redraw on resize. | mpv, CLIAMP | High | MVP | Easy | Escape title text; debounce; restore title on exit. | [MPV][MPV] |

### 4.8 Scripting and CLI

| ID | Feature | Description | Who has it | Demand | Priority | Difficulty | Library / design notes | Source |
|---|---|---|---|---|---|---|---|---|
| C01 | Non-interactive play/open | `player-tui play/open <path-or-url>` without entering the TUI. | mpv, mpvc | High | V1 | Easy | Same engine; `--no-tui`; JSON errors. | [MPV][MPV] |
| C02 | Remote commands | Play/pause/seek/volume/queue via CLI to a running instance. | mpvc, mpv IPC | High | V1 | Easy | Thin client over IPC; no shell injection. | [MPVC][MPVC] |
| C03 | JSON events | JSON Lines state/track/queue/error events for scripts. | mpv JSON IPC | High | V1 | Easy | Stable schema; versioned; stdout only in `--json`. | [MPV][MPV] |
| C04 | Unix socket IPC | Authenticated local control socket; Windows named pipe later. | mpv, mpvc | High | V1 | Medium | `$XDG_RUNTIME_DIR/player-tui.sock`; file mode 0600; peer checks. | [XDG][XDG] |
| C05 | Completions and man page | Bash/Zsh/Fish/PowerShell completions; man page. | Standard | Medium | V1 | Easy | Cobra or stdlib; generated and tested. | [COBRA][COBRA] |
| C06 | Config file | TOML config, XDG paths, portable mode, validation. | CLIAMP, kew, mpv | High | MVP | Easy | `go-toml`; defaults; `config path/edit`; no secrets. | [TOML][TOML] |
| C07 | Plugins | Lua/WASM hooks for metadata, commands, providers, UI. | CLIAMP, Yazi | Medium | V2 | Hard | Sandbox; no filesystem/network by default; versioned API. | [CLIAMP][CLIAMP] |
| C08 | MCP server | Local stdio MCP tools for playback, library, subtitles, AI. | None found | Medium | V2 | Medium | JSON-RPC; read-only tools by default; no network listener. | [MCP][MCP] |
| C09 | Doctor | Show libVLC, codecs, terminal capabilities, audio devices, models, config. | chafa, Yazi | High | MVP | Medium | `--json`; test pattern; actionable fixes. | [YAZI][YAZI] |
| C10 | Library/AI CLI wrappers | `library scan/search`, `subs gen`, `models list/download`, `ai index`. | None combined | High | V1 | Easy | Shared service layer; progress/cancel; JSON output. | [MCP][MCP] |
| C11 | Exit codes and logs | Predictable exit codes; structured logs; redaction. | Standard | High | MVP | Easy | `slog`; `--log-file`; no secrets/PII; `--verbose`. | [GOFUZZ][GOFUZZ] |

Command surface target:

```text
player-tui [paths/URLs]
player-tui play|open <target>
player-tui queue add|list|clear
player-tui library scan|search
player-tui subs gen|translate|convert
player-tui models list|download|remove
player-tui doctor [--json]
player-tui remote play|pause|seek|volume|status
player-tui mcp
player-tui config path|edit|validate
player-tui completions <shell>
```

Exit codes: 0 success, 2 usage, 3 media/open error, 4 missing dependency, 5
permission, 6 cancelled, 1 other. Keep them stable once documented.

### 4.9 Accessibility

| ID | Feature | Description | Who has it | Demand | Priority | Difficulty | Library / design notes | Source |
|---|---|---|---|---|---|---|---|---|
| AC01 | Full keyboard operation | Every action reachable without a mouse. | cmus, CLIAMP, Yazi | High | MVP | Easy | Focus model; shortcut discoverability; no mouse-only controls. | [CMUS][CMUS] |
| AC02 | Plain mode and screen reader | Disable graphics/animations; emit stable text and status for screen readers. | None found | High | MVP | Medium | `--plain`; line-oriented status; no alternate screen; no image protocols. | [NOCOLOR][NOCOLOR] |
| AC03 | High contrast and colorblind palettes | Accessible themes with contrast checks and non-color cues. | Some | High | V1 | Easy | Theme validator; symbols/labels in addition to color. | [NOCOLOR][NOCOLOR] |
| AC04 | NO_COLOR | Respect `NO_COLOR`; never require color. | CLI tools | High | MVP | Easy | Disable SGR except reset; monochrome theme; test. | [NOCOLOR][NOCOLOR] |
| AC05 | Reduced motion and photosensitivity | Disable flashing, rapid animation, and high-motion visualizers. | None found | Medium | V1 | Easy | `--reduced-motion`; cap redraw; no strobing effects. | [NOCOLOR][NOCOLOR] |
| AC06 | Large text and compact modes | Larger row spacing/simpler layout; compact single-line mode. | Some | High | V1 | Medium | Layout presets; minimum terminal size; preserve focus. | [VAXIS][VAXIS] |
| AC07 | Focus indicators | Clear, non-color-only focus and selection states. | Yazi, cmus | Medium | V1 | Easy | Border/glyph marker; high-contrast theme. | [YAZI][YAZI] |
| AC08 | Configurable timing | Seek step, toast TTL, double-press window, resume threshold. | Some | Medium | V2 | Medium | Config schema; sane defaults; validation. | [TOML][TOML] |
| AC09 | i18n, RTL, motor access | Translatable UI; RTL-aware layout where terminal permits; no chords required. | VLC, Kodi | Medium | V1 | Medium | `golang.org/x/text`; message catalogs; key remap; single-key alternatives. | [VLC][VLC] |
| AC10 | Audio description and sound events | AI-generated descriptions and captions for non-speech audio. | None stable | Low | V2 | Hard | AI V2; local TTS/AST; mark generated content; confidence. | [PIPER][PIPER] |

### 4.10 Platform

| ID | Feature | Description | Who has it | Demand | Priority | Difficulty | Library / design notes | Source |
|---|---|---|---|---|---|---|---|---|
| PL01 | Linux | Primary target: amd64/arm64, libVLC, ALSA/Pulse/PipeWire. | All | High | MVP | Medium | Dynamic libVLC; package docs; no bundled codecs unless legal. | [VLC][VLC] |
| PL02 | macOS | Homebrew, libVLC, MPRIS equivalent/Now Playing later. | VLC, mpv | High | V1 | Medium | cgo/dylib; codesigning/notarization; audio permissions. | [LIBVLC][LIBVLC] |
| PL03 | Windows and WSL | Native Windows later; WSL first for terminal users. | VLC, mpv | High | V1/V2 | Hard | Windows Terminal Sixel; libVLC DLL packaging; WSL audio caveats. | [WT][WT] |
| PL04 | Terminal compatibility matrix | Document/test Kitty, Ghostty, WezTerm, iTerm2, Windows Terminal, xterm, tmux, Zellij, screen. | chafa, Yazi | High | MVP | Medium | `doctor` + smoke tests; version-specific notes. | [AWSY][AWSY] |
| PL05 | Packaging and install | Tarball, deb/rpm, Homebrew, Scoop/winget, AppImage where licensing permits. | VLC, mpv | High | V1 | Medium | GoReleaser; dynamic libVLC dependency; no telemetry. | [GORELEASER][GORELEASER] |
| PL06 | Config paths and portable mode | XDG on Unix, known folders on Windows/macOS, `--config-dir`. | Standard | High | MVP | Easy | `os.UserConfigDir`; portable `./config`; migration. | [XDG][XDG] |
| PL07 | SSH, headless, containers | Work over SSH; audio-only/headless; document container device limits. | mpv, chafa | High | V1 | Medium | Low-bandwidth mode; no graphics fallback; no audio device error path. | [MPVT][MPVT] |
| PL08 | Security and escape safety | Escape untrusted metadata/subtitles/filenames; no shell; path canonicalization. | Required | High | MVP | Medium | Central `SafeText`; allowlist SGR; fuzz; CWE-150 tests. | [CWE150][CWE150] |
| PL09 | Terminal restore and crash recovery | Restore cursor/screen/raw mode/mouse/title on exit, signal, panic. | All TUIs | High | MVP | Medium | Vaxis close in `defer`; signal handler; idempotent cleanup; image delete. | [VAXIS][VAXIS] |
| PL10 | Fuzzing and malformed inputs | Fuzz subtitle, playlist, metadata, URL, and config parsers. | Security baseline | Medium | V1 | Medium | Go native fuzzing; corpus; timeouts; no network. | [GOFUZZ][GOFUZZ] |
| PL11 | Signed releases and SBOM | Checksums, signatures, SBOM, reproducible-ish builds. | Standard | Medium | V1 | Medium | GoReleaser + cosign; CycloneDX; publish verification steps. | [GORELEASER][GORELEASER] |
| PL12 | Diagnostics bundle | Redacted config/log/doctor output for bug reports. | Some | High | V1 | Easy | Zip with manifest; no media paths by default; preview before write. | [YAZI][YAZI] |
| PL13 | Proxy, TLS, locale | Respect proxy env, system CA, locale/encoding without breaking paths. | VLC, mpv | Medium | V1 | Medium | libVLC options; `x/text`; UTF-8 normalization. | [VLC][VLC] |
| PL14 | Licenses, name, no telemetry | Third-party notices, no VLC branding, no telemetry/update pings by default. | Required | High | MVP | Easy | `LICENSES.md`; dynamic libVLC notice; name review; opt-in update check only. | [VLC][VLC] |
| PL15 | Update check | Optional version check with explicit consent; no auto-download. | VLC, mpv | Low | V2 | Easy | Signed metadata; disable by default; offline mode. | [GORELEASER][GORELEASER] |

Platform notes: Windows native is V2 because cgo + libVLC DLL packaging + Windows
Terminal graphics need dedicated CI and real-device testing. WSL is the lower-risk
path for Windows users. Linux is the reference platform; macOS must be tested on
real hardware, not cross-compiled assumptions.

### 4.11 AI

AI is optional, local-only by default, and layered on top of normal playback. A
missing model must never break playback. All generated content is labeled,
editable, exportable, and removable. The app must not upload audio/video or
require an account.

| ID | Feature | Description | Who has it | Demand | Priority | Difficulty | Library / design notes | Source |
|---|---|---|---|---|---|---|---|---|
| AI01 | Offline ASR subtitles | Generate timed subtitles for a local file with a local Whisper model. | VLC demo only; no stable terminal player | High | V1 | Hard | `whisper.cpp` worker; FFmpeg decode to 16 kHz mono WAV for batch; progress/cancel. | [WHISPERCPP][WHISPERCPP] |
| AI02 | Translation | Translate generated/external subtitles offline; target language selectable. | VLC demo; no terminal player | High | V2 | Hard | Opus-MT/CTranslate2 [CT2] or local LLM; segment-level; preserve timestamps; glossary. | [OPUSMT][OPUSMT] |
| AI03 | Dual-language subtitles | Show original and translation together; per-line alignment. | VLC demo; no terminal player | Medium | V2 | Medium | Two text tracks; wrap carefully; toggle; same timing model. | [VLCAI][VLCAI] |
| AI04 | Batch subtitle generation | Queue a folder/library; resume; skip existing; export `.srt`/`.vtt`. | None combined | High | V1 | Medium | Job queue; one worker on low-end; atomic output; disk quota. | [WHISPERCPP][WHISPERCPP] |
| AI05 | Auto subtitle sync | Align existing subtitles to audio; estimate global offset or per-segment drift. | Some GUI tools | Medium | V2 | Hard | VAD/ASR anchors; prefer global offset first; never silently rewrite source. | [SILERO][SILERO] |
| AI06 | Voice activity detection | Avoid hallucinated text in silence; segment audio for ASR. | Whisper tooling | High | V1 | Easy | Silero VAD; padding; merge short segments; configurable threshold. | [SILERO][SILERO] |
| AI07 | Speaker diarization | Label speakers; optional speaker-colored subtitles. | GUI tools | Medium | V2 | Hard | pyannote/other worker; model license/gated access review; keep labels local. | [WHISPERCPP][WHISPERCPP] |
| AI08 | Sound-event captions | Caption music, laughter, impacts, applause. | Some GUI tools | Medium | V2 | Hard | Audio tagging model; confidence threshold; merge with ASR; avoid spam. | [SILERO][SILERO] |
| AI09 | Dialogue enhancement | Reduce noise, dereverb, boost dialogue. | Some GUI tools | Medium | V2 | Hard | DeepFilterNet worker; offline or audio-tap mode; A/B toggle; CPU limits. | [DEEPFILTER][DEEPFILTER] |
| AI10 | Loudness leveling | Normalize dialogue/effects consistently. | VLC, mpv filters | Medium | V2 | Medium | DSP/EBU R128; avoid pumping; per-user gain. | [VLC][VLC] |
| AI11 | Source separation and karaoke | Isolate vocals/dialogue; karaoke mode. | None stable | Low | V2 | Hard | Demucs worker; CPU-heavy; model/weights license review. | [DEMUCS][DEMUCS] |
| AI12 | Semantic video index | Transcribe and embed segments with timestamps. | None found | Medium | V2 | Hard | Transcript segments + local embeddings; SQLite vector table; cosine in Go. | [LLAMACPP][LLAMACPP] |
| AI13 | Ask the video | Natural-language Q&A over transcript; answers cite timestamps and jump there. | None found | Medium | V2 | Hard | Top-k retrieval + local LLM; show sources; refuse when evidence is weak. | [LLAMACPP][LLAMACPP] |
| AI14 | Auto chapters | Segment video into chapters with titles. | Some GUI tools | Medium | V2 | Medium | Transcript segmentation + optional LLM titles; editable; no scene detection required. | [LLAMACPP][LLAMACPP] |
| AI15 | Intro, recap, credits | Detect/skip or label common segments. | Some GUI tools | Low | V2 | Hard | Audio/text heuristics + LLM; always preview; never auto-skip without consent. | [LLAMACPP][LLAMACPP] |
| AI16 | Semantic search | Search library/video by meaning, not just filename/tag. | None found | Medium | V2 | Medium | Embeddings for transcripts/descriptions; cite source; rebuild/wipe. | [LLAMACPP][LLAMACPP] |
| AI17 | Natural-language command bar and voice | "Play the next unwatched episode", "make subtitles smaller"; optional push-to-talk. | None found | Low | V2 | Hard | Intent parser + tool calls; confirm destructive actions; local ASR; no always-on mic. | [WHISPERCPP][WHISPERCPP] |
| AI18 | Smart playlists and music tagging | Mood/BPM/key/genre tags; natural-language playlist rules. | Some GUI tools | Low | V2 | Hard | Audio models + tags; license review; user-editable; no cloud. | [LLAMACPP][LLAMACPP] |
| AI19 | Lyrics alignment | Align plain lyrics to audio; produce synced LRC. | Some GUI tools | Medium | V2 | Medium | Forced alignment or ASR anchors; preserve original; export. | [WHISPERCPP][WHISPERCPP] |
| AI20 | AI upscaling/interpolation | Improve low-res video or frame rate. | GUI players | Low | V2 | Hard | Terminal resolution limits value; model/GPU cost; experimental only. | [ONNX][ONNX] |
| AI21 | Audio description | Generate spoken descriptions for visual scenes. | None stable | Low | V2 | Hard | Vision model + local TTS; timing ducking; explicit opt-in; labeling. | [PIPER][PIPER] |
| AI22 | Model manager | Catalog, download, verify SHA256, size/disk, remove, offline mode. | None combined | High | V1 | Medium | Models in `$XDG_DATA_HOME/player-tui/models`; no weights in repo; resumable downloads. | [WHISPERCPP][WHISPERCPP] |
| AI23 | Privacy and index wipe | No telemetry/upload; delete transcripts, embeddings, jobs, models. | Required | High | V1 | Easy | Per-library opt-in; `ai wipe`; clear temp audio; redact logs. | [MCP][MCP] |
| AI24 | Bring-your-own endpoint | Optional local/OpenAI-compatible endpoint for LLM features; disabled by default. | GUI tools | Low | V2 | Medium | No default cloud; API key env/keyring only; warn on remote URL. | [LLAMACPP][LLAMACPP] |
| AI25 | Job scheduler and confidence | Queue priorities, pause on battery, CPU/RAM limits, confidence badges. | None combined | High | V1 | Medium | Worker pool; nice/ionice; resumable; show model/confidence; cancel. | [WHISPERCPP][WHISPERCPP] |

AI architecture:

- Main app owns playback and UI. AI work runs in **separate worker processes**:
  `whisper.cpp`/`faster-whisper` [FASTERWHISPER] for ASR, `llama.cpp` for
  LLM/embeddings, optional ONNX Runtime for audio/vision models. No model code is
  linked into the TUI process.
- Batch ASR for local files: decode audio with FFmpeg/libav to 16 kHz mono WAV,
  run VAD, run Whisper, write timestamped segments, then export SRT/VTT. This is
  the V1 path; it does not require an audio tap.
- Live ASR/translation/visualizer requires an audio tap. libVLC 3 audio callbacks
  replace normal output, so V2 needs a custom sink (or a parallel decode path)
  with a ring buffer. Do not claim real-time subtitles until this is proven on
  Linux/macOS/Windows.
- Whisper code is [WHISPER] and the large-v3 model is [WHISPERMODEL]. Model
  licenses are separate from code licenses. Record each model, source,
  hash, license, size, and required attribution in `LICENSES.md`/model catalog.
  Do not bundle weights without explicit redistribution permission.

## 5. Cross-cutting design decisions

### 5.1 Engine

- Use libVLC 3.0.24 through a thin in-repo cgo layer. Dynamic link
  `libvlc.so.5`/`libvlccore.so.9`; do not copy VLC source into the repo.
- Keep one engine owner goroutine. libVLC callbacks only enqueue events; they
  never block on UI or rendering.
- Explicit state machine and idempotent teardown. A panic, SIGINT, SIGTERM, or
  SIGHUP must still restore the terminal and stop VLC.
- Record the engine trade-off in this document: the original plan considered
  mpv JSON IPC, but libVLC was approved after benchmarks. Do not switch engines
  without a new decision record.

### 5.2 Video pipeline

- Request **I420** from libVLC; sample Y/U/V at the target cell grid and convert
  to RGB in Go. Fall back to RV32 if I420 negotiation fails.
- Measured: 1080p60 I420 display-only 33-38 fps vs RV32 about 8 fps; 720p30 I420
  about 26.5 fps vs RV32 about 26 fps, with CPU dropping from about 159% to 124%.
  The callback copy is about 1 fps at 1080p; the remaining cost is conversion.
- Use a latest-frame mailbox. If the renderer is busy, replace the pending frame,
  never grow an unbounded queue. Audio never waits for video.
- Frame dropping: audio master clock; drop video frames that are late by a
  threshold; keep audio continuous. No audio: monotonic clock with the same
  drop policy.
- Auto-quality order: reduce render cells/resolution, then drop thresholds, then
  only if the decoder is behind consider `avcodec-skiploopfilter`/`avcodec-fast`.
- Do not enable `avcodec-fast` by default. It is non-spec and error-prone; it
  reduced decoder CPU by about 20-25% in tests but did not improve display fps.
- `threads=1` hurts 1080p; keep automatic thread selection.
- Output budget: 120x40 half-block is about 176 KB/frame (5.3 MB/s at 30 fps);
  200x50 is about 366 KB/frame (11 MB/s). Low-bandwidth mode and dropping are
  required, not optional.

### 5.3 Rendering protocols

- Capability detection must be conservative: environment hints, DA1/XTSMGRAPHICS,
  cursor/cell-size queries, and a user override. If uncertain, use half-block.
- Kitty: raw RGB, chunked base64, image IDs, placement, delete, zlib optional.
  Test under tmux with `allow-passthrough` and under Zellij.
- Sixel: quantize to 256 colors, band encode, use scroll regions, and clean up.
  Sixel is slower and more variable; test xterm, WezTerm, Windows Terminal, tmux.
- iTerm2: OSC 1337 inline images; prefer Kitty when both are available.
- Half-block: `▀` with foreground top pixel and background bottom pixel; batch
  rows; support truecolor and 256/16-color fallback.
- ASCII: luminance ramp, optional edge enhancement, no color; the lowest common
  denominator for SSH, screen, and accessibility.
- Synchronized output (DEC 2026) where available; otherwise accept some tearing.
- Always clean up images and restore cursor/title on exit. Never leave the
  terminal in raw mode.

### 5.4 Subtitles

- Text path: parse external SRT/VTT/ASS, embedded text (V1 via FFmpeg), and
  AI-generated SRT; render in the terminal with safe SGR.
- Bitmap path: let libVLC blend PGS/VobSub/CEA into video frames; do not try to
  extract and position bitmaps in the cell grid.
- Disable the libVLC subtitle track when using the text path to avoid double
  rendering.
- Map ASS styles approximately: bold/italic/color/outline; ignore positioning,
  rotation, animations, and karaoke timing unless S09 is implemented.
- Escape all untrusted text; cap line length and line count; never allow OSC 52,
  OSC 8, DCS, or raw CSI from media metadata/subtitles.

### 5.5 Library and storage

- SQLite with WAL and migrations; prefer pure-Go `modernc.org/sqlite` to avoid a
  second cgo dependency. FTS5 for search, fuzzy matching for short queries.
- Store paths as canonical absolute paths plus a stable ID. Never delete media
  because it is missing; mark it unavailable.
- Cover art cache keyed by content hash; bounded disk usage; user can clear it.
- AI data is separate: transcripts, segments, embeddings, jobs. `ai wipe` clears
  it without touching the library.
- No network calls by default except user-requested media URLs. No telemetry.

### 5.6 UI and input

- Keyboard-first, Vim and arrow presets, no required chords. Mouse optional.
- Command line `:` and fuzzy palette share one command registry.
- Themes are semantic roles, not raw colors. Validate contrast; support
  `NO_COLOR`, 256-color, 16-color, and ASCII/icon fallbacks.
- Plain mode disables alternate screen, graphics, animation, and color; emits a
  stable line-oriented status suitable for screen readers and logs.
- Every error has a short toast and a full log entry with a stable error code.

### 5.7 Security

- Treat filenames, tags, lyrics, subtitles, playlist entries, URLs, and error
  strings as hostile. Strip/escape C0/C1, OSC, DCS, CSI, and bidi controls.
- Allow only SGR sequences generated by the app. Never echo raw terminal input.
- No shell interpolation. Use `exec.Command` with argument arrays; validate paths.
- Canonicalize paths and resolve symlinks according to policy; prevent traversal
  in playlists and subtitle sidecar search.
- IPC socket is user-only (0600), same-user peer check where available. MCP is
  stdio-only by default; destructive tools require explicit confirmation.
- Fuzz parsers and add regression tests for escape injection.
- Do not implement DRM circumvention. If a system library is present, document
  that it is optional and user-supplied.

### 5.8 Testing and verification

- Unit tests for state machine, clock, subtitle parsing/styling, escaping, path
  handling, config, and library queries.
- Golden-image tests for half-block/ASCII and protocol encoders; decode the
  emitted sequence in a test terminal parser where possible.
- [PTY][PTY] tests for startup/shutdown, resize, signals, raw-mode restore, and prompts.
- Real-device smoke tests: Kitty, Ghostty, WezTerm, iTerm2, Windows Terminal,
  xterm, tmux, Zellij, GNU Screen, SSH, 80x24 and 200x60.
- Media corpus: 720p30, 1080p60, odd dimensions, 10-bit/HDR, ASS/PGS subtitles,
  multichannel audio, Unicode filenames, malformed playlists/subtitles.
- Performance targets (laptop, to be confirmed): 720p30 >= 24 fps at 120x40;
  1080p60 >= 30 fps with I420 at 120x40; audio drift < 50 ms over 10 minutes;
  clean terminal restore on every exit path.

## 6. MVP scope and release criteria

MVP is the first usable terminal player, not a feature-complete competitor:

- Open local files/directories/URLs; play/pause/stop/seek; queue; external
  subtitles; audio track selection; volume.
- Kitty, Sixel, half-block, and ASCII rendering with capability detection,
  frame dropping, cleanup, and terminal restore.
- Basic now-playing/status/seek UI, command palette/help, themes/NO_COLOR,
  file browser, and `doctor`.
- Escape safety, signal/panic restore, and no raw terminal leaks.
- No AI required. No library index required for MVP, but a simple file browser
  and queue are required.
- Target 80x24 and 120x40; handle resize and odd widths/heights.
- If a terminal lacks graphics, playback must still work via half-block/ASCII.

MVP acceptance:

1. A 10-minute 720p30 file plays with audio/video sync and no terminal corruption.
2. Seeking, pause/resume, track switching, and subtitle delay work repeatedly.
3. SIGINT/SIGTERM/panic restore the terminal and stop libVLC cleanly.
4. Kitty and Sixel images are deleted on exit; half-block leaves no stale cells.
5. Malformed subtitle/metadata/playlist input cannot inject terminal escapes.
6. `doctor` correctly identifies the test terminal and explains fallbacks.

## 7. Top 20 build-first list

This is the recommended order. It reaches a usable MVP first, then adds library
and automation, then the first AI differentiator. It is a build order, not the
complete MVP scope (see section 6); some items can proceed in parallel once the
engine contract is stable.

| # | ID | Feature | Phase | Why now / exit criterion |
|---|---|---|---|---|
| 1 | P03 | Engine lifecycle and transport | MVP | No playback without a reliable state machine; exit when open/play/pause/stop/seek/end/error are repeatable and teardown is idempotent. |
| 2 | P02 | Playback clock and A/V sync | MVP | Audio must be the master clock; exit when 10-minute playback drift is < 50 ms and pause/seek re-anchor cleanly. |
| 3 | PL08 + PL09 | Escape safety and terminal restore | MVP | A terminal player that corrupts the terminal is unusable; exit when every exit path restores state and injection tests pass. |
| 4 | V01 | Capability detection and `doctor` | MVP | Pick the right renderer and explain failures; exit when the matrix detects Kitty/Sixel/iTerm2/half-block/ASCII correctly. |
| 5 | V06 | Half-block truecolor renderer | MVP | Universal fallback and fastest path to end-to-end video; exit when 720p30 at 120x40 is smooth and colors are correct. |
| 6 | V03 | Kitty renderer | MVP | Best quality on Kitty/Ghostty/WezTerm/iTerm2; exit when image placement, resize, and cleanup work. |
| 7 | V04 | Sixel renderer | MVP | Broad compatibility with xterm/WezTerm/Windows Terminal; exit when images display and are disposed correctly. |
| 8 | V09 | Frame scheduler and dropping | MVP | Prevent unbounded latency; exit when audio never stalls, late frames drop, and status shows the drop count. |
| 9 | F01-F03 | Container/video/audio baseline | MVP | The app must open real media; exit when the corpus plays, seeks, and reports tracks without crashing. |
| 10 | L01 | File and folder browser | MVP | Users need to find media; exit when browsing/opening/enqueueing works with spaces, Unicode, and symlinks. |
| 11 | P05 | Queue, shuffle, repeat | MVP | Core playback workflow; exit when next/prev/reorder/shuffle/repeat survive a restart. |
| 12 | S01 | External subtitles | MVP | Subtitle support is a baseline expectation; exit when SRT/VTT/ASS load, select, delay, and persist. |
| 13 | S05 | Terminal text subtitle styling | MVP | The reason to parse text instead of burning it in; exit when wrapping, background, colors, and safe escaping work. |
| 14 | U01 | Adaptive layout, status, seek | MVP | The player must be usable at 80x24 and 200x60; exit when panes/focus/status/seek/chapters do not shift or overlap. |
| 15 | U04 + U05 | Command palette, help, keymap | MVP | Discoverability without a manual; exit when every action is reachable and remappable. |
| 16 | U06 | Themes and NO_COLOR | MVP | Terminal diversity and accessibility; exit when three themes, 256/16-color fallback, and `NO_COLOR` all work. |
| 17 | L03 | SQLite library index and fuzzy search | V1 | Turns a file browser into a player; exit when 1k tracks scan and search returns in < 50 ms. |
| 18 | C01 + C03 | Non-interactive play and JSON events | V1 | Automation and testing; exit when scripts can play/seek/query and consume stable JSON Lines. |
| 19 | AC02 | Plain mode and screen-reader output | MVP | Accessibility is not optional; exit when every core action works without graphics/color and status is line-oriented. |
| 20 | AI01 | Offline ASR subtitles | V1 (AI v1) | First major differentiator; exit when a local file generates an SRT with progress/cancel, no network, and editable output. |

Deliberately **not** in the top 20: iTerm2-specific output (Kitty covers it),
network/disc playback, recording, visualizers, real-time translation, semantic
Q&A, plugins, and native Windows. They are V1/V2 and should not delay the core.

## 8. Suggested phases

- **Phase C (MVP)**: every MVP row, starting with top 1-16 and 19. Engine, clock,
  security/restore, renderers, formats, browser, queue, external subtitles,
  volume/track selection, layout, commands, themes, plain mode. No AI, no library
  DB required.
- **Phase D (AI v1 + library)**: top 17-18 and 20; AI04, AI06, AI22, AI23, AI25.
  Batch ASR, VAD, model manager, privacy/wipe, job scheduler. Library index and
  CLI/JSON land here.
- **Phase E (parity)**: iTerm2, resume, playlists, screenshots,
  [MPRIS][MPRIS] ([godbus][GODBUS]), audio device/EQ/ReplayGain, embedded text
  subtitles, network streams, macOS, packaging. This is the first "public"
  release.
- **Phase F (AI v2 + platform)**: translation, dual subtitles, semantic index,
  ask-the-video, chapters, semantic search, native Windows, discs/shares,
  visualizers/audio tap, plugins/MCP.

## 9. Risks and open questions

Risks:

- **libVLC 3 callback limits**: no text subtitle callback; audio tap replaces
  output; callback vout conversion is the performance bottleneck. Mitigations:
  hybrid subtitle path, I420, own renderer, V2 audio sink.
- **Embedded text subtitles**: requires FFmpeg demux/decode or an external
  `ffmpeg` fallback. Decide before promising full subtitle parity.
- **Terminal fragmentation**: Sixel/tmux/screen behavior varies; capability
  detection can be wrong. Mitigation: conservative fallback, `doctor`, real
  device tests, `--render` override.
- **cgo/packaging**: libVLC dynamic linking, macOS notarization, Windows DLLs,
  cross-compilation. Mitigation: Linux first, WSL for Windows, CI per platform.
- **AI performance**: CPU-only Whisper/LLM can be slow; model sizes and RAM vary.
  Mitigation: tiny/base defaults, job queue, progress/cancel, no playback impact.
- **Model and weight licenses**: code license != model license; some weights are
  gated or non-commercial. Mitigation: catalog, verify, no bundling without
  permission, user download with consent.
- **Name conflict**: `player-tui` is free on package registries but not unique on
  GitHub. Pick a distinctive name before public release.
- **No local audio/GPU/Kitty/Sixel**: development is partially blind; real-device
  testing is mandatory.

Open questions:

1. Final project name. The license is **MIT** (see `LICENSE`); dynamically
   linked LGPL libVLC is compatible, but third-party notices still need review.
2. Windows native in V1 or WSL-only until V2?
3. Link FFmpeg for embedded text subtitles, or require an external `ffmpeg`
   fallback in V1?
4. When to build the audio tap: V2 with visualizers/live ASR, or earlier if live
   translation becomes a priority?
5. Default AI model sizes for low-end laptops and the maximum acceptable disk/RAM.
6. Whether to ship a library DB in MVP or keep MVP file-browser-only (current
   recommendation: file-browser-only; library in V1).

## 10. Out of scope

- GUI, mobile, web, and cloud sync.
- Paid cloud AI, accounts, telemetry, ads, or model upload.
- DRM circumvention, including bundled CSS/Blu-ray decryption.
- Full video editing, encoding suite, DVD authoring, or broadcast tooling.
- Plugin marketplace, themes marketplace, or social features.
- 3D/VR, Chromecast/AirPlay output, and multi-room audio.
- Pixel-perfect libass rendering in the terminal (approximation only).
- Native Windows in MVP; WSL is the supported path initially.

## 11. Sources

Reference links used above. Re-check versions and URLs before publishing.

[VLC]: https://www.videolan.org/vlc/features.html
[VLC3024]: https://www.videolan.org/vlc/releases/3.0.24.html
[VLCSEC]: https://www.videolan.org/security/sb-vlc3024.html
[VLCAI]: https://www.ghacks.net/2025/01/13/vlc-to-add-support-for-offline-ai-subtitles-and-translations/
[KODI]: https://kodi.tv/about/features
[MPV]: https://mpv.io/manual/stable/
[MPVT]: https://mpv.io/manual/stable/#terminal
[MPV042]: https://github.com/mpv-player/mpv/releases/tag/v0.42.0
[MPVC]: https://github.com/gmt4/mpvc
[SPARK]: https://github.com/dividebysandwich/sparkplayer
[CLIAMP]: https://github.com/bjarneo/cliamp
[TERMUSIC]: https://github.com/tramhao/termusic
[KEW]: https://github.com/ravachol/kew
[CMUS]: https://cmus.github.io/
[CMUSCOVER]: https://github.com/nogizhopaboroda/cmus-cover-art
[CMUS1191]: https://github.com/cmus/cmus/issues/1191
[CMUS443]: https://github.com/cmus/cmus/issues/443
[MUSIKCUBE]: https://musikcube.com/
[NCMPCPP]: https://github.com/ncmpcpp/ncmpcpp
[NCPLAYER]: https://github.com/dankamongmen/ncplayer
[CHAFA]: https://github.com/hpjansson/chafa
[CHAFA14]: https://hpjansson.org/blag/2024/01/18/chafa-1-14-all-singing-all-dancing/
[YAZI]: https://github.com/sxyazi/yazi
[KITTY]: https://sw.kovidgoyal.net/kitty/graphics-protocol/
[GHOSTTY]: https://ghostty.org/docs/install/release-notes/1-4-0
[WEZTERM]: https://wezterm.org/features.html
[ITERM2]: https://iterm2.com/documentation-images.html
[WT]: https://github.com/microsoft/terminal/releases/tag/v1.22.10352.0
[XTERM]: https://xterm.dev/manpage-xterm
[XTERMCTL]: https://invisible-island.net/xterm/ctlseqs/ctlseqs.html
[ZELLIJ]: https://zellij.dev/documentation/compatibility.html
[AWSY]: https://www.arewesixelyet.com/
[LIBVLC]: https://videolan.videolan.me/vlc/group__libvlc.html
[VAXIS]: https://github.com/rockorager/vaxis
[ASTISUB]: https://github.com/asticode/go-astisub
[LIBASS]: https://github.com/libass/libass
[TAGLIB]: https://taglib.org/
[DHOWDEN]: https://github.com/dhowden/tag
[SQLITE]: https://pkg.go.dev/modernc.org/sqlite
[FSNOTIFY]: https://github.com/fsnotify/fsnotify
[GOSIXEL]: https://github.com/mattn/go-sixel
[GODBUS]: https://github.com/godbus/dbus
[COBRA]: https://github.com/spf13/cobra
[TOML]: https://github.com/pelletier/go-toml
[XIMAGE]: https://pkg.go.dev/golang.org/x/image
[WHISPER]: https://github.com/openai/whisper
[WHISPERCPP]: https://github.com/ggml-org/whisper.cpp
[WHISPERMODEL]: https://huggingface.co/openai/whisper-large-v3
[FASTERWHISPER]: https://github.com/SYSTRAN/faster-whisper
[SILERO]: https://github.com/snakers4/silero-vad
[OPUSMT]: https://huggingface.co/Helsinki-NLP/opus-mt-en-es
[CT2]: https://github.com/OpenNMT/CTranslate2
[DEMUCS]: https://github.com/facebookresearch/demucs
[DEEPFILTER]: https://github.com/Rikorose/DeepFilterNet
[LLAMACPP]: https://github.com/ggml-org/llama.cpp
[PIPER]: https://github.com/rhasspy/piper
[ONNX]: https://github.com/microsoft/onnxruntime
[MCP]: https://modelcontextprotocol.io/
[CWE150]: https://cwe.mitre.org/data/definitions/150.html
[GOFUZZ]: https://go.dev/doc/security/fuzz/
[GORELEASER]: https://goreleaser.com/
[NOCOLOR]: https://no-color.org/
[MPRIS]: https://specifications.freedesktop.org/mpris-spec/latest/
[XDG]: https://specifications.freedesktop.org/basedir-spec/latest/
[PTY]: https://github.com/creack/pty

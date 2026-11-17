# player-tui roadmap

Status: Phase C design complete; MVP playback slice implemented. This roadmap is the delivery view; [`FEATURES.md`](FEATURES.md)
is the feature inventory and [`ARCHITECTURE.md`](ARCHITECTURE.md) is the MVP
technical design.

## Release definitions

| Milestone | Meaning | Scope |
|---|---|---|
| **MVP** | First usable terminal player | Local files and direct URLs, launcher command, libVLC playback, terminal renderers, subtitles, queue, basic library browser, `doctor`. No AI, no server clients. |
| **V1** | First public release | MVP plus persistent library, IPC/JSON, personal media servers, subtitles workflow, AI v1 (offline ASR), and the Phase E parity work. |
| **V2** | Depth and platform expansion | AI v2, plugins/MCP, discs/shares, visualizers, native Windows. |

Priority labels in `FEATURES.md` are requests, not phase promises. A V1 item may
land during Phase D or Phase E.

## Phase C: MVP

Target: a single local file or direct URL plays smoothly, can be launched from
the command line, and every exit path restores the terminal.

Current slice: local files and direct URLs play through libVLC with I420
half-block, quadrant, or sextant output, exact area averaging, optional dither,
pause/seek/quit, and terminal restore. `play` is an alias for `open`. The
remaining Phase C items are listed below.

### C1. Core playback and launcher

- [ ] Engine lifecycle and transport: open, play, pause, stop, seek, rate, end,
      error, idempotent teardown (`P03`).
- [ ] Playback clock and A/V sync; audio is the master clock (`P02`).
- [ ] I420 video callbacks with RV32 fallback; bounded frame mailbox (`V01`,
      `V06`).
- [x] **Generic launcher command**: `player-tui open <local-file-or-url>` with
      `--title`, `--sub-file`, and `--start` (`C12`, MVP).
- [x] `--start` parsing: `SS`, `MM:SS`, `HH:MM:SS`; bounds-checked and applied
      after media opens.
- [x] `--sub-file` validates and attaches an external subtitle before playback.
- [x] `--title` overrides the display title only; never changes the source name
      or metadata.
- [x] Opening a target enters the TUI by default; a non-TUI mode remains a V1
      concern (`C01`).
- [ ] Queue, shuffle, repeat, and next/previous (`P05`).
- [ ] Volume, mute, audio/subtitle track selection (`P04`, `P06`).
- [ ] External SRT/VTT/basic ASS subtitles with delay (`S01`, `S05`).

### C2. Terminal output

- [ ] Capability detection and `doctor` (`V01`, `C09`).
- [x] Half-block truecolor renderer (`V06`).
- [x] Quadrant (2x2) and sextant (2x3) block modes, exact area averaging,
      `--dither`, and `--render auto` with halfblock fallback.
- [x] In-memory render benchmark with render p50/p95 and display FPS.
- [ ] Kitty renderer with image IDs, placement, resize, and cleanup (`V03`).
- [ ] Sixel renderer with palette/repeat handling and cleanup (`V04`).
- [ ] ASCII fallback for narrow, monochrome, and non-graphics terminals.
- [ ] Frame scheduler with late-frame dropping and status metrics (`V09`).
- [x] Escape safety, alternate-screen/raw-mode restore, signal/panic cleanup
      (`PL08`, `PL09`).
- [ ] Reduced-bandwidth mode and configurable output budget.

### C3. UI and workflow

- [ ] Adaptive layout at 80x24 and 120x40 (`U01`).
- [ ] File/folder browser with Unicode, spaces, and symlinks (`L01`).
- [ ] Status/seek/chapters, queue and now-playing panes (`U02`, `U03`).
- [ ] Command line, palette, help, and remappable keymap (`U04`, `U05`).
- [ ] Themes, `NO_COLOR`, and 256/16-color fallback (`U06`).
- [ ] Plain mode and line-oriented screen-reader output (`AC02`).
- [ ] Config file, exit codes, and redacted structured logs (`C06`, `C11`).

### Phase C exit criteria

1. A 10-minute 720p30 file plays with A/V sync and no terminal corruption.
2. `player-tui open` works for a local file and a direct URL, including
   `--title`, `--sub-file`, and `--start`.
3. A malformed `--start`, missing subtitle, or unreadable target fails with a
   stable exit code and an actionable error.
4. Seeking, pause/resume, track switching, and subtitle delay work repeatedly.
5. SIGINT/SIGTERM/panic restore the terminal and stop libVLC cleanly.
6. Kitty/Sixel images are removed on exit; half-block leaves no stale cells.
7. `doctor` identifies the test terminal and explains fallbacks.

## Phase D: V1 library, automation, servers, AI v1

Target: turn the player into a local library and make it usable from scripts and
the user's own media servers.

### D1. Generic launcher handoff (MVP carries forward)

The launcher contract is deliberately source-agnostic:

```text
player-tui open <local-file-or-url>
  [--title <text>]
  [--sub-file <path>]
  [--start <SS|MM:SS|HH:MM:SS>]
```

- Any downloader, media browser, or file manager can invoke this command.
- A completed download is opened by local path.
- A direct stream URL is opened without a scraper or catalogue integration.
- It does not include a third-party catalogue scraper or downloader.

### D2. Personal media servers (V1)

**Jellyfin, Plex, and DLNA/UPnP client for the user's own servers (`L13`).**

- Unified `Catalog` interface: authenticate, list libraries, browse items,
  search, resolve playable URL, report progress, list tracks/subtitles.
- Jellyfin adapter: REST API plus WebSocket session/state updates.
- Plex adapter: server discovery, user token, library metadata, direct-play URL,
  progress reporting. Treat community-documented endpoints as compatibility
  surface and test against current server versions.
- DLNA/UPnP adapter: SSDP discovery, device description, ContentDirectory
  browse, metadata extraction, and direct HTTP playback.
- Credentials: OS keyring or environment/file fallback; never in logs, crash
  reports, URLs echoed to the terminal, or SQLite plaintext by default.
- Network behavior: user-initiated requests only, bounded timeouts, explicit
  proxy/TLS behavior, no cloud relay, no telemetry.
- Playback: resolve a direct-play URL and hand it to libVLC with headers/cookies
  only as required; do not proxy or transcode unless a later ADR adds it.
- Failure handling: server offline, expired token, unsupported codec, and
  unreachable direct-play URL must degrade to an actionable error, not a hang.

### D3. Library and automation

- [ ] SQLite index, FTS5, fuzzy search (`L03`).
- [ ] Metadata, tags, duration, and cover art (`L02`).
- [ ] Filters, sort, playlists, queue persistence, missing-file handling
      (`L04`-`L06`, `L09`).
- [ ] Incremental watcher (`L05`).
- [ ] Non-interactive control, Unix socket IPC, JSON Lines events (`C01`-`C04`).
- [ ] Completions and man page (`C05`).
- [ ] Library/AI CLI wrappers (`C10`).

### D4. AI v1

- [ ] Offline ASR subtitle generation with progress/cancel (`AI01`).
- [ ] VAD, model manager, job scheduler, privacy/wipe, and editable output
      (`AI04`, `AI06`, `AI22`, `AI23`, `AI25`).

### Phase D exit criteria

1. A 1k-item library scans and searches in under 50 ms on the reference laptop.
2. Jellyfin, Plex, and a DLNA server can each be configured, browsed, and played
   without exposing credentials in logs.
3. A direct-play failure returns a clear error and leaves the player usable.
4. Scripts can open, control, and observe playback over JSON Lines.
5. ASR produces a valid, editable SRT without network access.

## Phase E: V1 parity and public release

- iTerm2 renderer, resume, playlists, screenshots.
- MPRIS/DBus, audio device selection, EQ/ReplayGain, embedded text subtitles.
- Network stream polish, macOS packaging, release installers, `LICENSES.md`.
- Accessibility audit, performance pass, real-device terminal matrix.
- Public release once Phase C/D/E acceptance and license notices are complete.

## Phase F: V2

- Translation, dual subtitles, semantic index, ask-the-video, chapter generation.
- Native Windows, discs and network shares, visualizers/audio tap.
- Plugins, MCP server, advanced automation.

## Cross-phase work

- Keep `README.md`, `PROGRESS.md`, and `FEATURES.md` current.
- Maintain `ARCHITECTURE.md` decisions with ADRs when behavior changes.
- Test on real Kitty, Sixel, half-block, and plain terminals before calling a
  renderer done.
- Keep `LICENSES.md` complete before any public release.
- No scraper/downloader for third-party catalogues; server support is for media
  the user already has access to.

## Risks

- libVLC callback limits and terminal-specific graphics behavior.
- cgo packaging and dynamic libVLC availability.
- Plex API compatibility and server version drift.
- Credential storage portability across Linux, macOS, and WSL.
- AI model size, CPU-only performance, and model-license constraints.
- Local testing gaps: no audio device, Kitty, or Sixel terminal on the current
  development phone.

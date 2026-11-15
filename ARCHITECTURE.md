# player-tui architecture

Status: Phase C design baseline, 2026-11-15. Scope is the MVP, with explicit
extension points for the V1 library, personal media servers, and AI features
described in [`FEATURES.md`](FEATURES.md) and [`ROADMAP.md`](ROADMAP.md).

## 1. Goals and non-goals

Goals:

- Play local files and direct URLs with reliable A/V sync in a terminal.
- Launch from another program with `player-tui open <local-file-or-url>` and
  honor `--title`, `--sub-file`, and `--start`.
- Render through Kitty, Sixel, half-block, or ASCII, with automatic fallback.
- Keep the UI responsive under slow terminals, large frames, and network stalls.
- Keep libVLC, cgo, terminal protocols, and UI state isolated behind testable
  interfaces.
- Leave clean extension points for a SQLite library, Jellyfin/Plex/DLNA clients,
  IPC, and offline AI.

Non-goals for MVP:

- Downloading or scraping third-party catalogues.
- Server-side transcoding, cloud relay, or account discovery.
- Embedded text-subtitle extraction from every container.
- Live translation, semantic search, and audio-tap features.
- Native Windows support. Linux first, macOS/WSL later.

## 2. System shape

```text
                         +----------------------+
  CLI / external tool -->| cmd/player-tui       |
                         | flags, exit codes    |
                         +----------+-----------+
                                    |
                         +----------v-----------+
                         | internal/app         |
                         | lifecycle, wiring    |
                         +--+--------+--------+--+
                            |        |        |
              +-------------+  +-----+----+  +----------------+
              |                |          |                   |
       +------v------+  +------v----+  +--v-----------+  +----v---------+
       | engine      |  | ui        |  | catalog      |  | ai (V1)      |
       | libVLC      |  | Vaxis     |  | Jellyfin/    |  | worker       |
       | adapter     |  | state     |  | Plex/DLNA    |  | supervisor   |
       +------+------+  +------+----+  +------+-------+  +--------------+
              |                |              |
       +------v------+  +------v----+  +------v-------+
       | video frame |  | terminal  |  | storage      |
       | mailbox     |  | compositor|  | SQLite (V1)  |
       | + clock     |  | renderers |  | secrets      |
       +-------------+  +-----------+  +--------------+
```

The process is a single Go binary. libVLC, the terminal, and the UI share the
process for MVP. AI work runs in separate worker processes when added. A future
ADR may isolate the media engine for crash containment; the interfaces below
already make that possible.

## 3. Package layout

```text
cmd/player-tui/                 process entry, exit-code mapping
internal/app/                   startup, dependency wiring, shutdown
internal/cli/                   open/play commands, flag parsing, validation
internal/config/                TOML, XDG paths, precedence, validation
internal/engine/                Engine, MediaSpec, events, state, clock
internal/engine/libvlc/         cgo bindings, callbacks, event bridge
internal/video/                 Frame, I420 sampling, scaling, frame mailbox
internal/render/                Renderer contract, capability registry
internal/render/halfblock/      truecolor cell renderer
internal/render/kitty/          Kitty graphics encoder
internal/render/sixel/          Sixel encoder
internal/render/ascii/          luminance/ASCII fallback
internal/term/                  Vaxis integration, capability probe, compositor
internal/ui/                    state store, panes, keymap, palette, commands
internal/subtitle/              SRT/VTT/basic ASS parsing, cue timeline
internal/playlist/              queue and repeat/shuffle rules
internal/library/               index, scanner, metadata (V1)
internal/catalog/               Catalog interface and registry (V1)
internal/catalog/jellyfin/      Jellyfin adapter (V1)
internal/catalog/plex/          Plex adapter (V1)
internal/catalog/dlna/          SSDP + ContentDirectory adapter (V1)
internal/ipc/                   Unix socket client/server, JSON Lines (V1)
internal/ai/                    job model, worker supervision (V1)
internal/platform/              signals, paths, keyring, terminal restore
```

Dependency direction is one-way: `app` -> feature packages -> shared domain
packages. `engine`, `video`, `subtitle`, and `catalog` must not import `ui` or
`term`. `render` may import `video` and `term`, never `engine`.

## 4. Runtime model

| Goroutine | Owns | Notes |
|---|---|---|
| `main` / app | lifecycle, signal context, exit code | Starts and joins the rest. |
| Vaxis loop | terminal input, text screen, resize | Single reader of terminal input. |
| engine owner | libVLC instance, player, media | Serializes engine calls with a mutex or command channel. |
| libVLC event bridge | callbacks into a buffered Go channel | Never blocks a libVLC thread. |
| video callback path | C-plane copy into pooled frames | Minimal work; no allocation-heavy conversion. |
| render scheduler | frame choice and drop policy | Latest-frame-wins; no unbounded queue. |
| subtitle parser | cue loading and lookup | Separate from playback clock. |
| library scanner (V1) | scan jobs, SQLite writes | Bounded worker pool. |
| AI supervisor (V1) | external worker processes | No AI work in the UI process. |

Rules:

- All terminal writes go through one serialized `term.Output`.
- UI state is owned by the Vaxis/UI goroutine. Other goroutines send typed
  messages; they do not mutate widgets or keymaps directly.
- Engine methods enqueue work for the engine owner and return promptly; blocking
  libVLC calls never run on the Vaxis/UI goroutine.
- Every long-lived goroutine receives a context and has a bounded shutdown path.
- Panic recovery is installed at the app boundary so terminal restore runs.
  Native libVLC crashes cannot be recovered in-process; the cgo surface and
  teardown order are kept small and tested.

## 5. Domain contracts

### 5.1 MediaSpec

`MediaSpec` is the only input the engine needs:

```go
type MediaSpec struct {
    Target       string
    Title        string
    SubtitleFile string
    Start        time.Duration
    Headers      http.Header // trusted adapters only; never CLI-exposed
    Cookies      []*http.Cookie
    Source       SourceRef   // file, direct URL, or catalog item
}
```

`Target` is a canonical local path or an allowlisted URL scheme. `Title` is a
display override only. `SubtitleFile` must be a regular file. `Start` is clamped
to duration after the media opens. `Headers` and `Cookies` are used only by
trusted V1 catalog adapters and are redacted everywhere.

### 5.2 Engine

```go
type Engine interface {
    Open(context.Context, MediaSpec) error
    Play() error
    Pause() error
    Stop() error
    Seek(time.Duration) error
    SetRate(float64) error
    SetVolume(int) error
    SetMute(bool) error
    Tracks() TrackSet
    SelectTrack(kind TrackKind, id string) error
    SetSubtitleFile(string) error
    SetSubtitleDelay(time.Duration) error
    Snapshot() PlayerState
    Events() <-chan Event
    Frames() <-chan *video.Frame
    Close() error
}
```

The engine exposes media time as the authoritative clock. The UI never talks to
libVLC directly.

### 5.3 Renderer

```go
type Renderer interface {
    Name() string
    Init(term.Surface, RenderConfig) error
    Draw(*video.Frame, Rect) error
    Clear(Rect) error
    Resize(Rect) error
    Close() error
}
```

Renderers receive a viewport and a frame, never a libVLC handle. Capability
probing selects a renderer, and `--render` can override it.

### 5.4 Catalog (V1)

```go
type Catalog interface {
    ID() string
    Name() string
    Authenticate(context.Context, Credentials) error
    Libraries(context.Context) ([]Library, error)
    Browse(context.Context, ItemID) ([]Item, error)
    Search(context.Context, string) ([]Item, error)
    Resolve(context.Context, ItemID) (MediaSpec, error)
    ReportProgress(context.Context, ItemID, time.Duration) error
}
```

Adapters are independent packages registered at startup. The library and UI see
only the interface.

## 6. libVLC adapter

### 6.1 Ownership

`internal/engine/libvlc.Player` owns exactly one `libvlc_instance_t`, one
`libvlc_media_player_t`, and the current `libvlc_media_t`. One owner goroutine
performs open/play/stop/teardown. Release order is: detach callbacks, stop player,
release media, release player, release instance. Teardown is idempotent.

An engine generation number increments on every `Open` and `Stop`. Events from
an older generation are ignored. This prevents a late `EndReached` or
`EncounteredError` from a previous file mutating the new playback state.

### 6.2 Video callbacks

Use `libvlc_video_set_callbacks` plus
`libvlc_video_set_format_callbacks` in the libVLC 3 API:

- Request **I420**; fall back to RV32 only if the format callback refuses or the
  setup fails.
- The format callback records width, height, pitches, and plane offsets.
- A single C-allocated buffer holds all planes. No Go pointer is retained by C.
- A C-allocated context stores a numeric engine handle. C callbacks call a small
  exported Go shim, which looks up the engine in a `sync.Map`. The C layer never
  stores a Go pointer.
- `lock` returns the C plane pointer. `display` copies the active planes into a
  pooled Go `Frame` and publishes it. Copying happens while the C buffer is
  valid; callbacks never retain C memory.
- The display callback carries no PTS in libVLC 3. Stamp each copied frame with
  a monotonic arrival time and a sampled media time. The scheduler treats media
  time as authoritative and uses arrival time only for interpolation.

Frame shape:

```go
type Frame struct {
    Width, Height int
    Y, U, V       []byte
    Strides       [3]int
    MediaTime     time.Duration
    Arrival       time.Time
    Sequence      uint64
}
```

Frames come from a pool. The mailbox holds at most one pending frame. A newer
frame replaces an older one and increments a drop counter. The renderer returns
frames to the pool after encoding.

### 6.3 Event bridge

Attach to the media-player event manager for lifecycle, buffering, time,
position, length, seekability, tracks, and errors. Callbacks push compact events
into a buffered channel using non-blocking sends. If the channel is full, set an
overflow flag; the pump reconciles state with `libvlc_media_player_get_state`,
time, and track getters. High-frequency `TimeChanged` events are coalesced.

The engine state machine is:

```text
Idle -> Opening -> Buffering? -> Playing <-> Paused
                                  |  ^
                                  v  |
                                Seeking
                                  |
                                  v
                                Ended/Error -> Idle/Closed
```

`Stopped` and `EndReached` are distinct. `Seeking` is cleared only after a time
or seek-completion event for the current generation. Errors carry a stable code
and a user-facing message; raw libVLC strings are logged, not rendered directly.

### 6.4 Clock and scheduling

`Clock` interpolates between libVLC media-time samples using Go's monotonic
clock. It freezes while paused, re-anchors after seek, and never uses wall-clock
time as the master. Target drift is < 50 ms over ten minutes.

The render scheduler ticks faster than the output frame rate, asks the clock for
`now`, and selects the newest frame:

- If `frame.MediaTime < now - 50ms`, drop it.
- If the frame is early, keep the previous frame until the next tick.
- For unknown timestamps, use arrival time and a conservative 30 fps cap.
- No queue grows without bound. Late frames are always preferable to latency.

### 6.5 Audio and tracks

MVP uses libVLC's audio output; no audio tap. Volume, mute, audio-track and
subtitle-track selection go through the engine. Audio is the master clock by
construction. V1 adds device selection and ReplayGain; V2 may add an audio tap
for visualization, which replaces the normal output and needs a separate ADR.

## 7. Terminal and rendering

### 7.1 Terminal surface

`internal/term` owns:

- Vaxis initialization, raw mode, alternate screen, focus, resize, and input.
- Capability detection: Kitty, Sixel, iTerm2, truecolor, 256/16 colors, width.
- A single serialized output path with synchronized-update begin/end where
  supported.
- Idempotent restore on normal exit, signal, panic, and failed startup.

Vaxis remains the source of cell-grid text and key events. A `Compositor`
coordinates text cells with raw graphics commands. Renderers never write to
stdout directly. A Phase C spike must confirm the exact Vaxis hook for raw
graphics output; if Vaxis cannot safely expose it, the compositor owns the raw
writer and Vaxis is limited to input and screen state. This is an implementation
detail behind `term.Surface`, not a UI-visible behavior.

### 7.2 Renderer selection

1. Explicit `--render`/config override, if valid.
2. Kitty if the terminal and multiplexer support it.
3. Sixel if supported and not known-broken in the current multiplexer.
4. Half-block truecolor.
5. Half-block 256/16 color.
6. ASCII.

`doctor` reports the detected chain, the selected renderer, and why each higher
choice was rejected. Fallback must work even if capability detection is wrong.

### 7.3 Frame conversion

`internal/video` converts I420 to the renderer's needs without a full RGBA
intermediate unless the renderer requires pixels:

- Half-block: box-filter each cell's top and bottom region to two RGB colors;
  emit foreground/background SGR pairs. This is the default video path.
- ASCII: luminance ramp with optional color.
- Kitty/Sixel: scale to the pixel rectangle and encode once per frame; cache
  row tables and reuse buffers.
- Never scale on the libVLC callback thread.

Video occupies a dedicated rectangle. MVP text chrome lives outside that
rectangle. Overlaying text on Kitty/Sixel video is deferred unless the
compositor can do it without breaking image cleanup; half-block can composite
text by skipping cells. Subtitles in graphics modes use a reserved subtitle band
or the half-block fallback when an overlay is required.

### 7.4 Backpressure

- The frame mailbox drops stale frames.
- The terminal writer has a bounded queue; if the terminal is slower than the
  target, drop frames rather than queueing milliseconds of latency.
- Low-bandwidth mode reduces cell resolution, color depth, and frame rate.
- Rendering never blocks audio or libVLC callbacks.

## 8. UI

The UI is a single-goroutine state machine:

```text
terminal event / engine event / timer / command result
                         |
                         v
                    ui.Update
                         |
                         v
                    ui.View -> Vaxis screen
```

- `ui.State` holds panes, focus, queue, playback snapshot, subtitles, status, and
  transient toasts.
- `ui.Message` is a closed set of typed messages; no ad-hoc shared maps.
- `ui.Command` describes an effect (open, seek, save config). Commands are
  executed by the app layer, not by widgets.
- The keymap is data, validated at load time, and shared by help and palette.
- The command registry is shared with the CLI where semantics overlap.
- Rendering is capped independently of input; resize triggers renderer re-init.

MVP panes: browser, now-playing/video, queue, status/seek, help/palette. Narrow
terminals collapse to tabs or a single pane. Every action is keyboard reachable.

## 9. Launcher command

The launcher is the external integration contract and an MVP feature (`C12`):

```text
player-tui open <local-file-or-url>
  [--title <text>]
  [--sub-file <path>]
  [--start <SS|MM:SS|HH:MM:SS>]
```

Flow:

1. Parse and validate flags before entering the TUI.
2. Canonicalize a local path (`filepath.Abs` + symlink resolution). Reject
   directories and non-regular files with exit 3.
3. For URLs, allow only `http`, `https`, `rtsp`, `rtmp`, `srt`, and `file`;
   reject credentials embedded in URLs unless explicitly supported later.
4. Validate `--sub-file` as a readable regular file with a size cap.
5. Sanitize `--title`: valid UTF-8, control characters removed, bounded length.
6. Parse `--start`; negative or malformed values exit 2.
7. Open the engine, attach the subtitle, seek, start playback, and enter the UI.
8. Report failures with stable exit codes; never print secrets or raw headers.

`player-tui <target>` is shorthand for `player-tui open <target>`. A non-TUI
mode (`C01`) is V1. The command does not inspect catalogues, resolve third-party
links, or download media.

## 10. Subtitles

- External SRT, VTT, and basic ASS are parsed in-process into a cue timeline.
- Cue text is sanitized for terminal safety before layout.
- The subtitle lookup uses the master clock, not wall time.
- Text subtitles are rendered as terminal cells with wrapping, style, and delay.
- Bitmap subtitles (PGS/VobSub) and embedded text tracks are delegated to libVLC
  where possible. libVLC 3 has no text-subtitle callback, so in-app styling of
  embedded text is not promised in MVP.
- Do not load both libVLC's text renderer and the in-app renderer for the same
  external file; choose one path to avoid duplicates.

## 11. Configuration, storage, and secrets

Paths follow XDG:

- config: `$XDG_CONFIG_HOME/player-tui/config.toml`
- data: `$XDG_DATA_HOME/player-tui/`
- cache: `$XDG_CACHE_HOME/player-tui/`
- state/socket: `$XDG_RUNTIME_DIR/player-tui/` when available

Precedence: command-line flags > environment > config file > defaults. Config
validation reports all errors before startup. Writes are atomic (temp file +
rename). No telemetry and no network calls by default.

V1 library uses a pure-Go SQLite driver (`modernc.org/sqlite`), WAL mode,
embedded migrations, and a single writer. Media files are read-only unless the
user explicitly edits tags. FTS5 backs search; stable IDs survive rescans.

Secrets use the OS keyring where available, with a 0600 file fallback for
headless systems. Tokens, cookies, Authorization headers, and full URLs with
query tokens must never appear in logs, terminal output, crash reports, or the
SQLite database in plaintext.

## 12. V1 personal media servers

**Jellyfin, Plex, and DLNA/UPnP are V1 (`L13`).** They are clients for servers
the user already runs; they are not catalogue scrapers.

- `internal/catalog` defines the interface and a registry; adapters are isolated.
- **Jellyfin**: server URL plus API key/token; REST browse/search; WebSocket for
  session state; resolve a direct-play URL; report progress. Use the documented
  API and version-test against current stable servers.
- **Plex**: manually configured server URL plus user token; library browse,
  metadata, direct-play URL, progress. Plex's HTTP API is community-documented,
  so treat endpoint behavior as a compatibility surface and test it.
- **DLNA/UPnP**: SSDP discovery, device description, ContentDirectory SOAP
  browse, DIDL-Lite parsing, and direct HTTP playback. Bound XML sizes and HTTP
  timeouts; select the network interface explicitly.
- Direct play only in V1. If the server cannot direct-play a file, return an
  actionable error; do not silently transcode or proxy.
- Headers/cookies required by a server are supplied as trusted `MediaSpec` data
  and mapped through an allowlist of libVLC options.
- Credentials are stored through `platform.SecretStore`; adapter tests use fake
  servers with no real tokens.

## 13. IPC and automation (V1)

- A local Unix socket at `$XDG_RUNTIME_DIR/player-tui/control.sock`, mode 0600.
- JSON Lines request/response and event stream; schema is versioned.
- Linux peer credentials (`SO_PEERCRED`) and macOS `getpeereid`; reject remote
  peers. Windows named pipes later.
- Commands: play/pause/seek/volume/status/queue, plus read-only library queries.
- No network listener. No shell execution. Logs redact command arguments that
  may contain paths or tokens.
- CLI exit codes and JSON errors remain stable once documented.

## 14. AI (V1)

The AI subsystem is out of the MVP binary path. When added:

- `internal/ai` supervises external worker processes (for example whisper.cpp)
  rather than linking large runtimes into the UI process.
- Jobs are persisted, cancellable, resumable, and run at low priority.
- Audio extraction uses an external `ffmpeg` process with explicit arguments,
  never a shell string.
- Model downloads are explicit, checksummed, license-reviewed, and stored under
  the cache/data directory.
- No network access during inference. No playback stall: AI work is throttled
  and can be suspended while video is active.

## 15. Security and robustness

- Treat filenames, titles, metadata, subtitle text, playlist entries, and server
  responses as untrusted. Strip or visibly escape C0/C1 control bytes before
  terminal output.
- Allowlist URL schemes and libVLC media options. Never pass user input as a
  libVLC option string.
- Bound subtitle files, playlists, JSON lines, XML documents, and HTTP bodies.
- Use HTTP timeouts, response-size limits, and TLS verification. Do not disable
  certificate checks.
- Do not log credentials, cookies, Authorization headers, or full query strings.
- Validate IPC peer identity and socket permissions.
- Keep the cgo boundary minimal; no C callback may block or call back into UI
  code. All C memory is freed on every teardown path.
- Restore the terminal on every Go exit path. A native crash is documented as a
  process-level failure, not silently recovered.

## 16. Performance budgets

| Path | Budget |
|---|---|
| 720p30 half-block at 120x40 | >= 24 fps, no audio stall |
| 1080p60 I420 on reference laptop | >= 30 fps display path |
| A/V drift over 10 minutes | < 50 ms |
| UI input latency | < 50 ms |
| Frame copy | bounded by 1 buffer per active frame; pool reuse |
| Memory | no unbounded queues; cap subtitle/playlist/XML sizes |
| Startup | terminal probe + engine init before first frame; no network by default |

Measure with the existing benchmark harness and add renderer-specific counters
for dropped frames, encode time, and output bytes. Benchmarks must not run in
the normal test path.

## 17. Testing

- Unit: flag/`--start` parsing, path/URL validation, title sanitization, clock
  interpolation, scheduler drop rules, subtitle parsing, queue rules.
- Integration: libVLC with a small media corpus; open/play/seek/end/error;
  repeated teardown; race detector.
- Renderer: golden cell output for half-block/ASCII; protocol byte snapshots and
  cleanup sequences for Kitty/Sixel; PTY tests for resize and restore.
- Terminal: vttest-style manual matrix plus automated PTY tests; test `tmux`,
  `screen`, and `zellij` where practical.
- Catalog (V1): `httptest` Jellyfin/Plex fakes, DLNA XML fixtures, timeout and
  malformed-response tests.
- AI (V1): worker cancellation, missing model, corrupt model, no-network run.
- Fuzz: subtitle, playlist, metadata, URL, and IPC parsers.

## 18. Failure handling

| Failure | Behavior |
|---|---|
| libVLC missing/incompatible | Exit 4 with install guidance; no partial TUI. |
| No usable renderer | Fall back to ASCII; `doctor` explains why. |
| Invalid `--start` / missing subtitle | Exit 2 or 3 before playback; actionable message. |
| Media open error | Exit 3; preserve terminal; no stale state. |
| Server offline / expired token | Show error, keep library usable, offer re-auth. |
| Direct-play URL unsupported | Explain codec/container issue; no silent transcode. |
| Terminal resize during playback | Re-init renderer, clear old images, resume. |
| Writer backpressure | Drop frames, keep audio and input responsive. |
| Go panic | Restore terminal, log redacted stack, exit non-zero. |

## 19. Decisions and open issues

Accepted decisions:

- ADR-001: libVLC 3 engine, dynamically linked, with a thin in-repo cgo layer.
- ADR-002: I420 callback path; RV32 fallback; audio-master clock.
- ADR-003: Vaxis for terminal input and text grid; a serialized compositor for
  raw graphics.
- ADR-004: single process for MVP; engine interfaces allow future isolation.
- ADR-005: launcher is source-agnostic; no catalogue scraping.
- ADR-006: personal media servers are direct-play clients for user-owned servers.
- ADR-007: MIT project license; LGPL libVLC linked dynamically; `LICENSES.md`
  remains required before public release.

Open issues to resolve during Phase C:

1. Exact Vaxis hook for safe raw graphics output and synchronized updates.
2. Whether Kitty/Sixel can share a screen with Vaxis text cells without clearing
   the video rect; fallback is a dedicated video region.
3. Frame timestamp quality from libVLC 3 display callbacks; tune interpolation.
4. Overlay strategy for subtitles on Kitty/Sixel; half-block supports compositing
   now.
5. Keyring backend availability on headless Linux and WSL.
6. Plex API version support and token re-auth flow.
7. Native crash containment and whether V2 moves the engine to a supervisor.

The next Phase C deliverable is `TECH_STACK.md`, followed by `DESIGN.md`. No
implementation starts before the architecture and stack decisions are reviewed.

# player-tui — Master Build Prompt

Put this file in the root of your repo and tell OpenCode: "Read player-tui-master-prompt.md and follow it." It tells the AI what to research, what to build, which tools to use, how the terminal interface should look, and which AI features to add.

## 1. Your role and goal

You are a senior systems engineer and terminal-UI designer. Build **player-tui**: a free, open-source media player that lives entirely inside the terminal, with VLC-class power, keyboard-first control, and on-device AI. It must play audio and video, show real video in the terminal when the terminal allows it, and degrade gracefully everywhere else, including over SSH.

This is VLC-class, not a code copy. Do not use VLC’s name, cone logo or branding. mpv, FFmpeg and any other library you reuse have their own licenses, so record each one in LICENSES.md. Also check that the name player-tui is free on crates.io, GitHub, GitLab and the main package managers before publishing.

## 2. Step one: research the whole web first

Before writing code, search broadly and verify current facts instead of relying on memory:

- Study: mpv (including its terminal video outputs and JSON IPC), mpvc, SparkPlayer, cmus, musikcube, ncmpcpp, termusic, kew, CLIAMP, notcurses/ncplayer, chafa, Yazi, plus VLC and Kodi for the feature baseline.
- Terminal graphics: Kitty graphics protocol, Sixel, iTerm2 inline images, truecolor half-blocks. Check which terminals support what today (Kitty, Ghostty, WezTerm, foot, Konsole, Windows Terminal, iTerm2, xterm) and how tmux, Zellij and screen affect it.
- Where users complain: Reddit, GitHub and GitLab issues, forums. Search things like “terminal video player”, “what I wish cmus had”, “mpv terminal tips”.
- AI: on-device speech recognition, translation, audio enhancement and video search that can run without a GPU or cloud account.
- For every feature record: name, what it does, who has it, how often users ask for it, source link. Remove duplicates and anything that needs a paid cloud service.

## 3. Step two: write FEATURES.md

Create FEATURES.md grouped by category: Playback, Formats, Subtitles, Audio, Video rendering, Library, UI, Scripting and CLI, Accessibility, Platform, AI. For each feature give a description, priority (MVP / V1 / V2), difficulty (easy / medium / hard), the library needed, and the source link. End with a top-20 “build first” list.

## 4. Seed knowledge (verify and expand, do not treat as complete)

**Core features to match VLC**

- Plays almost any file, folder, disc image, URL and stream (HTTP, HLS, DASH, RTSP, network shares)
- Hardware decoding where available; HDR tone mapping to terminal colors
- Subtitles (SRT, ASS/SSA, PGS), multiple audio tracks, chapters, A–B loop, speed control, frame step, audio/subtitle delay
- Equalizer, filters, convert/record stream, screenshots
- Playlists, queue, resume position, remember per-file settings

**What a terminal player can do better**

- Starts instantly, tiny memory use, works over SSH and in tmux/Zellij
- Fuzzy-search library and command palette (fzf-style), vim-style keys plus arrows, optional mouse
- Scriptable: clean CLI, JSON output, IPC socket, shell completions, a Lua or WASM plugin system
- MPRIS and media-key support, lyrics (LRC) in sync, audio visualizers, album art
- Headless mode for servers and low-bandwidth mode for slow connections (smaller frames, lower fps, or audio-only)
- A `player-tui doctor` command that reports terminal capabilities and tells users the best settings

**Video rendering ladder (auto-detect, then fall back)**

1. Kitty graphics protocol
2. Sixel
3. iTerm2 inline images
4. Truecolor half-block characters
5. ASCII/ANSI 256-color

Audio is always the master clock: if the terminal cannot keep up, drop video frames, never stutter the sound.

**AI features people are waiting for**

- **Offline AI subtitles and live translation.** VideoLAN demoed this at CES 2025 (local models, no cloud, 100+ languages) but it had still not shipped in stable VLC by August 2026, so a terminal player that ships it first stands out. Whisper only translates into English, so pair it with a separate local translation model. Show subtitles as real terminal text, and add `player-tui subs gen file.mkv` for batch use.
- Ask-the-video: type a question or describe a scene and jump to it (`:find the car chase`)
- AI chapters, recap and summary; skip intro, recap and credits
- Auto subtitle sync: align any subtitle file to the audio
- Dialogue boost, noise removal, voice isolation, auto volume leveling
- Speaker-colored subtitles and sound-event captions for deaf and hard-of-hearing users
- Natural-language commands (“go back 30 seconds”, “subtitles in Spanish”) and optional voice control
- Smart playlists from the local library only; auto-tagging of messy music metadata
- Optional AI upscaling and frame interpolation for graphics-protocol modes, with a low-power fallback

## 4b. Advanced AI features (the main differentiator)

player-tui treats AI as a first-class layer, not an add-on. Everything runs on the user’s machine by default, every feature can be switched off, and models download only with permission. Build these in addition to the list above.

**Video understanding (a background indexer)**

- An idle-time indexer transcribes speech, detects shots and scenes, and embeds frames with a small vision model. Results go into the local SQLite library. It is pausable and resumable.
- Semantic search across the whole library: `:find the scene where they argue in the kitchen` returns timestamps with thumbnails drawn through the terminal graphics protocol.
- Ask-the-video: `:ask who is the man in the red coat?` or `:ask summarize the last 10 minutes` is answered by a small local LLM over the transcript and scene notes, with timestamps you can jump to. A no-spoilers mode only uses what comes before the current position.
- Auto chapters, scene titles and a “previously on…” recap. Intro and credits detection by audio fingerprint across the episodes of a season.
- Speaker tracking: label speakers once, then jump to the next line by that person.

**Language and subtitles**

- Live offline subtitles and translation, with two languages stacked on screen, ASR error cleanup, reading-speed-aware line splitting, and translation of existing SRT/ASS files that keeps their styling.
- Language-learning mode: select a word in the subtitle for its meaning, loop the current sentence, export it to Anki.
- Local dubbing with text-to-speech that keeps the original speaker’s style (later phase).

**Audio intelligence**

- Dialogue boost, noise removal, loudness leveling, voice and instrument separation for karaoke or practice, and sound-event captions such as [door slams] or [music].
- Music: auto-tagging, mood and genre, BPM and key detection, lyrics fetched and aligned to the audio, smart radio and crossfade.

**Picture intelligence (graphics-protocol terminals)**

- Upscaling, frame interpolation, denoise, auto-crop of black bars, and smart HDR to SDR tone mapping, on GPU when present.
- AI audio description for blind users: a local vision-language model describes scenes and a voice reads them (heavy, optional).

**Assistant and automation**

- Natural-language command bar: `:play something calm for coding`, `:skip to the credits`, `:Spanish subtitles and 0.9x speed`. A small local model turns the text into a structured command, shows it, and asks before anything destructive. The model may only choose from a whitelist of player commands and never runs shell code.
- Push-to-talk voice control with local speech recognition.
- MCP server and JSON IPC mode, so AI agents such as OpenCode can control playback, query the index and ask for summaries.
- Recommendations and smart playlists from the local library only, with a short “why this?” explanation.

**AI engine rules**

- A model manager: `player-tui models list | install | remove` showing size, license and speed on this machine. On first run, benchmark the hardware and pick the best model size (CPU-only, GPU through CUDA, Metal or Vulkan, or NPU).
- AI jobs run in a separate low-priority process. They pause automatically if playback needs the CPU, so audio and video never stutter.
- Optional bring-your-own endpoint (Ollama, a llama.cpp server, or the user’s own API key), off by default and clearly labeled whenever data would leave the machine.
- Privacy: nothing leaves the device by default, and one command wipes the index.
- Safety and quality: AI never overwrites user files (it writes sidecar .srt or .json files), shows confidence, and uses voice-activity detection to avoid invented subtitles in silence. Each feature ships with test media and expected outputs.
- Roadmap placement: indexer and subtitles in AI v1; semantic search, ask-the-video, assistant and MCP in AI v2; picture AI and audio description in a new AI v3 phase.

## 5. Step three: architecture and tools

Explain each choice and its trade-offs, then pick one. Evaluate these options:

| Area | Options to evaluate |
|---|---|
| Playback engine | Start with mpv driven over JSON IPC (fastest MVP; recent mpv releases reworked their Sixel, Kitty and TCT terminal outputs), then move to libmpv or FFmpeg with your own frame renderer for full control |
| Language | Rust (Ratatui + crossterm + ratatui-image), Go (Bubble Tea or Vaxis), or Python (Textual) for a quick prototype |
| Audio | libmpv, or symphonia + cpal/rodio, or miniaudio |
| Metadata and library | lofty or TagLib, ffprobe, SQLite for the library index |
| AI runtime | whisper.cpp or faster-whisper, ONNX Runtime, llama.cpp for small local models; models download only with permission |
| AI models | Whisper (speech), Silero VAD, Opus-MT (translation), Demucs or DeepFilterNet (audio). Check every model’s license. |
| Build and release | Cargo or Go modules, cargo-dist or GoReleaser, Homebrew, AUR, Nix, winget/Scoop, deb/rpm, Docker image |
| CI | GitHub Actions, mirrored to GitLab CI, on Linux, macOS and Windows |
| Testing | Unit tests, TUI snapshot tests, PTY integration tests, a terminal matrix (Kitty, Ghostty, WezTerm, foot, Windows Terminal, iTerm2, xterm, tmux, over SSH), a large test-media library, fuzzing |
| Tools for building | OpenCode with the GitHub and GitLab MCP servers, Context7 for library docs, asciinema or VHS for demo recordings |

Rules: AI runs on-device and offline by default, no telemetry without consent, a single static binary where possible.

## 6. Design brief

- Layout: library/file browser on the left, video or now-playing in the center, queue on the right (toggle each), status bar with progress and volume at the bottom, command line on `:`
- Adaptive: a compact layout at 80×24 and a wide layout at 200×50; redraw cleanly on resize
- Themes in a TOML file (ship Catppuccin, Gruvbox, Nord, plus a high-contrast theme); respect NO_COLOR
- Keys: vim-style plus arrows, a `?` help overlay, a command palette on Ctrl+P, everything remappable
- Seek bar with chapter marks and AI scene markers; spectrum visualizer for music
- Nerd Font icons optional, always with a plain ASCII fallback; handle wide Unicode characters correctly
- Accessibility: screen-reader-friendly plain mode, reduced motion, color-blind-safe palettes, full keyboard control
- Always restore the terminal properly on exit or crash
- Wireframe every screen as ASCII mockups in DESIGN.md before coding

## 7. Technical requirements and support

**Product requirements**

- Platforms: Linux, macOS, Windows Terminal and WSL
- Performance: startup under 200 ms, smooth 1080p playback in graphics-protocol terminals on a mid-range laptop, low CPU in audio-only mode, stable memory over long sessions
- Security: strip or escape terminal control sequences from file names, metadata, lyrics and subtitles so media files cannot inject escape codes into the terminal; fuzz every parser; sign releases

**Support setup after launch**

- README with demo GIFs, a man page, FAQ, and a terminal compatibility table
- Issue templates that ask for the `player-tui doctor` output
- Opt-in diagnostics bundle (logs and codec info), community-managed translations
- Stable and beta release channels

## 8. Deliverables and build order

Produce in this order: FEATURES.md, ARCHITECTURE.md (with diagrams), TECH_STACK.md, DESIGN.md, ROADMAP.md, LICENSES.md, then the code. Use the GitHub MCP to create the repo and one issue per MVP feature.

1. **MVP:** audio and video playback through mpv, file browser, queue, subtitles, Kitty/Sixel/half-block rendering with auto-detect, themes
2. **Parity:** streaming, filters, convert and record, library search, MPRIS, scripting, `doctor`
3. **AI v1:** offline subtitles and translation, skip intro, auto subtitle sync, audio enhancement
4. **AI v2:** ask-the-video, natural-language commands, smart playlists, optional upscaling
5. **Polish:** packaging for every platform, docs site, plugin API, first public release

At the end of each phase, run the tests, report what works and what is missing, and ask me before starting the next phase.

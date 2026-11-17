# Evening handoff

Date: 2026-11-17

All ten requested tasks are complete. The binary is installed, both remotes are
up to date, and the work stopped before any AI feature. Remaining items are
manual real-terminal/audio validation, not unfinished code.

## Completed (10/10)

1. Kitty, Sixel, and iTerm2 encoders with tests, `--render` selection, and
   environment auto-detection (Kitty, iTerm2, Sixel, then quadrant).
2. Status row with position, duration, volume/mute/rate, render mode, and help
   hints; resize redraw; `doctor`; clear errors for missing, empty, directory,
   and space-containing paths.
3. `--audio` flag (dummy output by default) and `docs/AUDIO-SETUP.md` for
   PulseAudio over loopback TCP from Termux. Termux commands were not run.
4. Multi-file queue (`player-tui play a b c`) with `n` / `p`.
5. Volume `+`/`-`, mute `m`, speed `[`/`]`, subtitle `s`, audio `a`, help `?`.
6. TOML config plus per-file resume; near-end positions are forgotten.
7. Four themes (`catppuccin`, `gruvbox`, `nord`, `high-contrast`) and
   `NO_COLOR` reverse-video fallback.
8. Text sanitizer with unit tests and a fuzz target (`FuzzText`, 3 s run passed).
9. File browser for no arguments or `player-tui browse [dir]`.
10. README usage/keys/render/compatibility docs and verified `install.sh`.

## Commits

`739e5ec..79b6ba4` on `main`, pushed to `origin` (GitLab) and `github` (GitHub).

Key hashes: `0a4b7b1` encoders, `d87c78e` status/doctor/errors, `6ac18a7` audio,
`3daee1e` queue, `70a6121` controls/help, `db7fe8f` themes, `0a44f53`
config/resume, `a26e4a6` sanitizer, `57ae758` browser, `5381365` README,
`86ba01a`/`79b6ba4` status and unreadable-media fixes.

## Verified

- `go test ./...` and `go vet ./...` pass after every change.
- PTY playback of `/tmp/bbb-720p-10s.mp4` with a two-file queue: pause, seek,
  volume, mute, speed, subtitle/audio cycling, help, next/previous, quit; exit 0.
- PTY browser: help, navigation, directory entry/parent, file selection; exit 0.
- Every renderer emitted its marker; Kitty emitted image delete and the terminal
  restored the alt screen after quit.
- Resume saved 5.164 s, and the next launch started at `00:05`.
- A text file renamed `.mp4` now exits 3 with
  `cannot play <file>: unreadable or unsupported media`.
- Path with spaces played (exit 0); missing path exits 3 with a clear message;
  `--help` exits 0.
- `install.sh` built and installed successfully from both the repository and a
  fresh local clone into `$HOME/.local/bin` (already on PATH).
- Fuzz target passed a 3 s run (earlier task).

## Blocked / not verified

- No Kitty/Sixel/iTerm2-capable display and no audio output in this environment.
  Graphics encoders are byte/PTY-tested only; images were not seen on screen, and
  only the dummy audio path ran.
- Termux-side PulseAudio commands and ssh tunneling were documented, not
  executed. Audio/clock sync remains untested.
- tmux/Zellij/screen passthrough and real font/tofu behavior were not tested.
- AI features (ASR, VAD, models) were intentionally not started.

## How to run it

```bash
export PATH="$HOME/.local/bin:$PATH"
player-tui doctor
player-tui play /tmp/bbb-720p-10s.mp4
player-tui play a.mp4 b.mp4 c.mkv
player-tui                                  # browse current directory
player-tui browse /path/to/media
```

Playback keys: `space` pause, `left`/`right` seek, `+`/`-` volume, `m` mute,
`[`/`]` speed, `s`/`a` subtitle/audio track, `n`/`p` queue next/previous,
`?` help, `q` quit.

## Manual checks (about 10 minutes)

1. In the terminal you will actually use: `player-tui doctor`. Confirm the
   reported capabilities and auto render mode.
2. Play the 10 s clip above; try pause, seek, speed, help, and quit.
3. Visual graphics check, one at a time:

   ```bash
   player-tui play /tmp/bbb-720p-10s.mp4 --render kitty
   player-tui play /tmp/bbb-720p-10s.mp4 --render sixel
   player-tui play /tmp/bbb-720p-10s.mp4 --render iterm2
   ```

   Expect an image, not escape-code garbage. On failure use `--render halfblock`
   and record the terminal name/version.
4. Audio path: run the Termux commands below, then on Debian:

   ```bash
   # Termux
   pkg install pulseaudio
   pulseaudio --start --exit-idle-time=-1
   pactl load-module module-native-protocol-tcp listen=127.0.0.1 port=4713 \
     auth-ip-acl=127.0.0.1 auth-anonymous=1

   # Debian
   export PULSE_SERVER=tcp:127.0.0.1:4713
   pactl info
   player-tui play /tmp/bbb-720p-10s.mp4 --audio
   ```

   Full topology/troubleshooting: `docs/AUDIO-SETUP.md`.
5. Resume check: play any clip for 6+ seconds, press `q`, then relaunch it and
   confirm it starts near the saved position. Config lives in
   `~/.config/player-tui/`.
6. Error check:

   ```bash
   printf 'not media\n' > /tmp/bad.mp4
   player-tui play /tmp/bad.mp4; echo "exit=$?"
   ```

   Expect the unreadable-media message and exit code 3.

## Notes

- Audio is off by default; volume keys show `audio is off (use --audio)`.
- `q` exits 0 after normal playback; a natural end stays on the last frame until
  `q` (no auto-advance yet).
- The installed binary was rebuilt from the final commit. Re-run
  `bash install.sh` after future changes.
- Do not expose PulseAudio beyond `127.0.0.1`; never use anonymous auth on a
  non-loopback listener.

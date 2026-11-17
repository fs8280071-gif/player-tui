# Audio setup: PulseAudio over TCP on loopback

Audio is **off by default**. Without `--audio`, player-tui starts libVLC with a
dummy audio output, which is what the headless and benchmark runs use.

This guide covers the supported development setup: player-tui runs inside Debian
(proot or a chroot) on the same Android device where PulseAudio runs in Termux.
Both sides share the loopback interface. Keep the TCP listener on `127.0.0.1`;
never bind PulseAudio to `0.0.0.0`.

> The Termux commands below are documentation only. They were not executed while
> building this repo. Adapt package names to your Termux setup.

## Topology

```text
Termux (host)                         Debian (proot/chroot)
PulseAudio :4713  <--- loopback --->  player-tui --audio
PULSE_SERVER=tcp:127.0.0.1:4713
```

If player-tui runs on a different machine, see [Remote Debian](#remote-debian).

## Termux side

Run these in Termux (not inside Debian):

```bash
pkg update
pkg install pulseaudio
pulseaudio --start --exit-idle-time=-1
```

Allow only loopback TCP clients:

```bash
pactl load-module module-native-protocol-tcp \
  listen=127.0.0.1 port=4713 \
  auth-ip-acl=127.0.0.1 auth-anonymous=1
```

Verify the listener exists:

```bash
pactl list short modules | grep native-protocol-tcp
ss -ltn 2>/dev/null | grep 4713 || netstat -ltn 2>/dev/null | grep 4713
```

If Termux restarts, run the `pulseaudio --start` and `pactl load-module` lines
again, or put them in `~/.termux/boot/` (requires Termux:Boot).

## Debian side

Install the client tools if they are missing:

```bash
sudo apt-get update
sudo apt-get install -y pulseaudio-utils
```

Point libVLC/PulseAudio at the Termux daemon:

```bash
export PULSE_SERVER=tcp:127.0.0.1:4713
pactl info
```

`pactl info` should print the Termux server name and a default sink. Then play
with audio enabled:

```bash
player-tui play /path/to/video.mp4 --audio
```

To make the setting persistent for the current shell only:

```bash
echo 'export PULSE_SERVER=tcp:127.0.0.1:4713' >> ~/.bashrc
```

Use `PLAYER_TUI_DUMMY_AUDIO=1` to force the dummy output even when `--audio` is
passed (useful for repeatable tests).

## Remote Debian

If player-tui runs on a laptop and PulseAudio runs on the phone, forward a local
port over SSH instead of exposing PulseAudio on the network:

```bash
# On the Debian machine:
ssh -N -L 4713:127.0.0.1:4713 user@phone
export PULSE_SERVER=tcp:127.0.0.1:4713
player-tui play /path/to/video.mp4 --audio
```

The SSH connection terminates the loopback listener on the phone side and
tunnels it to the Debian machine. Termux needs an SSH server (`pkg install
openssh` and `sshd`) for this path.

## Troubleshooting

- `Connection refused`: the Termux PulseAudio daemon is not running or the TCP
  module is not loaded. Re-run the two Termux commands above.
- `pactl: command not found`: install `pulseaudio-utils` on Debian.
- Audio works in `pactl` but not player-tui: confirm `--audio` is present and
  `PLAYER_TUI_DUMMY_AUDIO` is not set to `1`.
- Video plays but audio crackles: check `pactl list short sinks` and the Android
  battery optimizations for Termux; this setup is not low-latency.
- Never use `auth-anonymous=1` with a non-loopback `listen` address.

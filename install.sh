#!/usr/bin/env bash
set -euo pipefail

REPO_URL="${PLAYER_TUI_REPO_URL:-https://github.com/fs8280071-gif/player-tui.git}"
INSTALL_DIR="${PLAYER_TUI_DIR:-${XDG_DATA_HOME:-$HOME/.local/share}/player-tui}"
BIN_DIR="${PLAYER_TUI_BIN_DIR:-$HOME/.local/bin}"

SCRIPT_DIR=""
if [ -n "${BASH_SOURCE[0]:-}" ] && [ -f "${BASH_SOURCE[0]}" ]; then
  SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
fi

LOCAL_SOURCE=0
if [ "${PLAYER_TUI_FROM_REMOTE:-0}" != "1" ] &&
   [ -n "$SCRIPT_DIR" ] && [ -f "$SCRIPT_DIR/go.mod" ]; then
  INSTALL_DIR="$SCRIPT_DIR"
  LOCAL_SOURCE=1
fi

say() {
  printf '%s\n' "$*"
}

die() {
  printf 'error: %s\n' "$*" >&2
  exit 1
}

command -v git >/dev/null 2>&1 || die "git is required to download player-tui"

mkdir -p "$(dirname "$INSTALL_DIR")"

if [ "$LOCAL_SOURCE" -eq 1 ]; then
  say "Using local source in $INSTALL_DIR"
elif [ -d "$INSTALL_DIR/.git" ]; then
  say "Updating source in $INSTALL_DIR"
  git -C "$INSTALL_DIR" pull --ff-only
elif [ -e "$INSTALL_DIR" ]; then
  die "$INSTALL_DIR exists but is not a git checkout"
else
  say "Downloading player-tui to $INSTALL_DIR"
  git clone --depth 1 "$REPO_URL" "$INSTALL_DIR"
fi

[ -f "$INSTALL_DIR/go.mod" ] || die "no Go build target found in $INSTALL_DIR"
command -v go >/dev/null 2>&1 || die "Go 1.24+ is required to build player-tui"
command -v pkg-config >/dev/null 2>&1 || die "pkg-config is required to find libVLC"
pkg-config --exists libvlc 2>/dev/null || die \
  "libVLC development files not found. Install libvlc-dev (Debian/Ubuntu), vlc-devel (Fedora), or vlc (Arch), then rerun."

go_version="$(go env GOVERSION 2>/dev/null || true)"
say "Building player-tui with $go_version"
mkdir -p "$BIN_DIR"
(
  cd "$INSTALL_DIR"
  go build -trimpath -ldflags "-s -w" -o "$BIN_DIR/player-tui" ./cmd/player-tui
)
say "Installed: $BIN_DIR/player-tui"

ensure_path() {
  case ":$PATH:" in
    *":$BIN_DIR:"*) return 0 ;;
  esac

  local marker="# player-tui PATH"
  local line
  line="export PATH=\"$BIN_DIR:\$PATH\""
  local updated=0
  for rc in "$HOME/.profile" "$HOME/.bashrc"; do
    [ -f "$rc" ] || continue
    if ! grep -Fq "$marker" "$rc"; then
      printf '\n%s\n%s\n' "$marker" "$line" >>"$rc"
      updated=1
    fi
  done
  if [ "$updated" -eq 1 ]; then
    say "Added $BIN_DIR to PATH in your shell profile."
    say "Open a new terminal, or run: export PATH=\"$BIN_DIR:\$PATH\""
  else
    say "Add $BIN_DIR to PATH to run player-tui from any terminal."
  fi
}

ensure_path

say ""
say "Play something:"
say "  player-tui play /path/to/video.mkv"

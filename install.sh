#!/usr/bin/env bash
set -euo pipefail

REPO_URL="${PLAYER_TUI_REPO_URL:-https://github.com/fs8280071-gif/player-tui.git}"
INSTALL_DIR="${PLAYER_TUI_DIR:-${XDG_DATA_HOME:-$HOME/.local/share}/player-tui}"
BIN_DIR="${PLAYER_TUI_BIN_DIR:-$HOME/.local/bin}"

say() {
  printf '%s\n' "$*"
}

die() {
  printf 'error: %s\n' "$*" >&2
  exit 1
}

command -v git >/dev/null 2>&1 || die "git is required to download player-tui"

mkdir -p "$(dirname "$INSTALL_DIR")"

if [ -d "$INSTALL_DIR/.git" ]; then
  say "Updating source in $INSTALL_DIR"
  git -C "$INSTALL_DIR" pull --ff-only
elif [ -e "$INSTALL_DIR" ]; then
  die "$INSTALL_DIR exists but is not a git checkout"
else
  say "Downloading player-tui to $INSTALL_DIR"
  git clone --depth 1 "$REPO_URL" "$INSTALL_DIR"
fi

if [ ! -f "$INSTALL_DIR/go.mod" ]; then
  say "Source downloaded."
  say "No runnable player yet: Phase B planning is complete, MVP build starts in Phase C."
  exit 0
fi

command -v go >/dev/null 2>&1 || die "Go 1.24+ is required to build player-tui"

mkdir -p "$BIN_DIR"

if [ -d "$INSTALL_DIR/cmd/player-tui" ]; then
  say "Building player-tui"
  (cd "$INSTALL_DIR" && go build -trimpath -o "$BIN_DIR/player-tui" ./cmd/player-tui)
elif [ -f "$INSTALL_DIR/main.go" ]; then
  say "Building player-tui"
  (cd "$INSTALL_DIR" && go build -trimpath -o "$BIN_DIR/player-tui" .)
else
  die "No Go build target found (expected cmd/player-tui or main.go)"
fi

say "Installed: $BIN_DIR/player-tui"
case ":$PATH:" in
  *":$BIN_DIR:"*) ;;
  *) say "Add $BIN_DIR to PATH to run player-tui from any terminal." ;;
esac

#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 2 ]]; then
    echo "usage: open.sh KIND PORT [TITLE]" >&2
    exit 2
fi

KIND=$1
PORT=$2
TITLE=${3:-"Aiki PDP — $KIND"}
HERE=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
OBSERVER="$HERE/observer.ai"

pick_terminal() {
    if [[ -n "${AIKI_PDP_TERMINAL:-}" ]]; then
        command -v "$AIKI_PDP_TERMINAL" || {
            echo "aiki-pdp: terminal not found: $AIKI_PDP_TERMINAL" >&2
            exit 1
        }
        return
    fi
    for candidate in x-terminal-emulator gnome-terminal mate-terminal xfce4-terminal konsole xterm; do
        if command -v "$candidate" >/dev/null 2>&1; then
            command -v "$candidate"
            return
        fi
    done
    echo "aiki-pdp: no terminal emulator found; set AIKI_PDP_TERMINAL" >&2
    exit 1
}

TERMINAL=$(pick_terminal)
REAL_TERMINAL=$(readlink -f "$TERMINAL" 2>/dev/null || printf '%s\n' "$TERMINAL")
TERM_NAME=$(basename "$REAL_TERMINAL")

if [[ "$(basename "$TERMINAL")" == "x-terminal-emulator" ]]; then
    "$TERMINAL" --title="$TITLE" -e aiki "$OBSERVER" "$KIND" "$PORT" >/dev/null 2>&1 &
    exit 0
fi

case "$TERM_NAME" in
    gnome-terminal*|mate-terminal*)
        "$TERMINAL" --title="$TITLE" -e aiki "$OBSERVER" "$KIND" "$PORT" >/dev/null 2>&1 &
        ;;
    xfce4-terminal*)
        "$TERMINAL" --title="$TITLE" --command="aiki '$OBSERVER' '$KIND' '$PORT'" >/dev/null 2>&1 &
        ;;
    konsole*)
        "$TERMINAL" -p "tabtitle=$TITLE" -e aiki "$OBSERVER" "$KIND" "$PORT" >/dev/null 2>&1 &
        ;;
    xterm*)
        "$TERMINAL" -T "$TITLE" -e aiki "$OBSERVER" "$KIND" "$PORT" >/dev/null 2>&1 &
        ;;
    *)
        "$TERMINAL" -e aiki "$OBSERVER" "$KIND" "$PORT" >/dev/null 2>&1 &
        ;;
esac

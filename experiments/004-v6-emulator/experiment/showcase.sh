#!/usr/bin/env bash
set -euo pipefail

HERE=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
cd "$HERE"

pick_port() {
    if [[ -n "${AIKI_PDP_PORT:-}" ]]; then
        printf '%s\n' "$AIKI_PDP_PORT"
        return
    fi
    python3 - <<'PYPORT'
import socket
s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
s.bind(("127.0.0.1", 0))
print(s.getsockname()[1])
s.close()
PYPORT
}

export AIKI_PDP_PORT=$(pick_port)

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
OBSERVER="$HERE/observe/observer.ai"

open_view() {
    local kind=$1
    local title=$2
    if [[ "$(basename "$TERMINAL")" == "x-terminal-emulator" ]]; then
        "$TERMINAL" --title="$title" -e aiki "$OBSERVER" "$kind" "$AIKI_PDP_PORT" >/dev/null 2>&1 &
        return
    fi
    case "$TERM_NAME" in
        gnome-terminal*|mate-terminal*)
            "$TERMINAL" --title="$title" -e aiki "$OBSERVER" "$kind" "$AIKI_PDP_PORT" >/dev/null 2>&1 &
            ;;
        xfce4-terminal*)
            "$TERMINAL" --title="$title" --command="aiki '$OBSERVER' '$kind' '$AIKI_PDP_PORT'" >/dev/null 2>&1 &
            ;;
        konsole*)
            "$TERMINAL" -p "tabtitle=$title" -e aiki "$OBSERVER" "$kind" "$AIKI_PDP_PORT" >/dev/null 2>&1 &
            ;;
        xterm*)
            "$TERMINAL" -T "$title" -e aiki "$OBSERVER" "$kind" "$AIKI_PDP_PORT" >/dev/null 2>&1 &
            ;;
        *)
            "$TERMINAL" -e aiki "$OBSERVER" "$kind" "$AIKI_PDP_PORT" >/dev/null 2>&1 &
            ;;
    esac
}

echo "Aiki PDP-11/40 monitor"
echo "observer port: $AIKI_PDP_PORT"
echo "opening CPU, UNIBUS, tape, and KL11 observers"
echo

open_view cpu "Aiki PDP — CPU"
open_view unibus "Aiki PDP — UNIBUS"
open_view tape "Aiki PDP — Tape"
open_view kl11 "Aiki PDP — KL11"

exec aiki "$HERE/aiki-pdp.ai"

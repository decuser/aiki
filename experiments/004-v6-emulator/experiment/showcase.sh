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

OBSERVER_LAUNCHER="$HERE/observe/open.sh"
export AIKI_PDP_OBSERVER_LAUNCHER="$OBSERVER_LAUNCHER"

open_view() {
    local kind=$1
    local title=$2
    "$OBSERVER_LAUNCHER" "$kind" "$AIKI_PDP_PORT" "$title"
}

echo "Aiki PDP-11/40 monitor"
echo "observer port: $AIKI_PDP_PORT"
echo "opening CPU, UNIBUS, tape, RK11, and KL11 observers"
echo

open_view cpu "Aiki PDP — CPU"
open_view unibus "Aiki PDP — UNIBUS"
open_view tape "Aiki PDP — Tape"
open_view rk "Aiki PDP — RK11"
open_view kl11 "Aiki PDP — KL11"

exec aiki "$HERE/aiki-pdp.ai"

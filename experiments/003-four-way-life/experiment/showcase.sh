#!/usr/bin/env bash
set -euo pipefail

HERE=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
cd "$HERE"

GENERATIONS=${FWL_GENERATIONS:-400}
WIDTH=${FWL_WIDTH:-60}
HEIGHT=${FWL_HEIGHT:-40}
SEED=${FWL_SEED:-42}

STAMP=$(date +%Y%m%d-%H%M%S)
SHOW_DIR="$HERE/../results/showcase-$STAMP"
mkdir -p "$SHOW_DIR"

for n in 1 2 3 4; do
    : >"$SHOW_DIR/worker-$n.log"
done

export FWL_SHOWCASE_DIR="$SHOW_DIR"

pick_terminal() {
    if [[ -n "${FWL_TERMINAL:-}" ]]; then
        command -v "$FWL_TERMINAL" || {
            echo "showcase: terminal not found: $FWL_TERMINAL" >&2
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

    echo "showcase: no terminal emulator found; set FWL_TERMINAL" >&2
    exit 1
}

TERMINAL=$(pick_terminal)
REAL_TERMINAL=$(readlink -f "$TERMINAL" 2>/dev/null || printf '%s\n' "$TERMINAL")
TERM_NAME=$(basename "$REAL_TERMINAL")

open_view() {
    local title=$1
    local logfile=$2
    local viewer="$HERE/worker-view.sh"

    # Debian's x-terminal-emulator wrapper is reliably compatible with -e,
    # even when it resolves to gnome-terminal.wrapper. Prefer that proven
    # contract whenever the selected executable is x-terminal-emulator.
    if [[ "$(basename "$TERMINAL")" == "x-terminal-emulator" ]]; then
        "$TERMINAL" --title="$title" -e "$viewer" "$title" "$logfile" "$COORD_PID" >/dev/null 2>&1 &
        return
    fi

    case "$TERM_NAME" in
        gnome-terminal*)
            "$TERMINAL" --title="$title" -e "$viewer" "$title" "$logfile" "$COORD_PID" >/dev/null 2>&1 &
            ;;
        mate-terminal*)
            "$TERMINAL" --title="$title" -e "$viewer" "$title" "$logfile" "$COORD_PID" >/dev/null 2>&1 &
            ;;
        xfce4-terminal*)
            "$TERMINAL" --title="$title" --command="$viewer $(printf '%q' "$title") $(printf '%q' "$logfile") $COORD_PID" >/dev/null 2>&1 &
            ;;
        konsole*)
            "$TERMINAL" -p "tabtitle=$title" -e "$viewer" "$title" "$logfile" "$COORD_PID" >/dev/null 2>&1 &
            ;;
        xterm*)
            "$TERMINAL" -T "$title" -e "$viewer" "$title" "$logfile" "$COORD_PID" >/dev/null 2>&1 &
            ;;
        *)
            "$TERMINAL" -e "$viewer" "$title" "$logfile" "$COORD_PID" >/dev/null 2>&1 &
            ;;
    esac
}

cleanup() {
    if [[ -n "${COORD_PID:-}" ]] && kill -0 "$COORD_PID" 2>/dev/null; then
        kill -INT "$COORD_PID" 2>/dev/null || true
        wait "$COORD_PID" 2>/dev/null || true
    fi
}
trap cleanup EXIT INT TERM

echo "Four-Way Life showcase"
echo "results: $SHOW_DIR"
echo "terminal emulator: $TERMINAL"
echo "grid: ${WIDTH}x${HEIGHT}, generations: $GENERATIONS, seed: $SEED"
echo
echo "Coordinator/log terminal: this window"
echo "Canvas: separate window"
echo "Workers: four separate terminal windows"
echo

# Keep the real coordinator in this terminal while allowing us to open the
# four status views. Worker protocol stdout remains private to coordinator.
aiki coordinator.ai canvas "$GENERATIONS" "$WIDTH" "$HEIGHT" "$SEED" \
    > >(tee "$SHOW_DIR/coordinator.log") 2>&1 &
COORD_PID=$!

open_view "Four-Way Life — Engine A / FILE" "$SHOW_DIR/worker-1.log"
open_view "Four-Way Life — Engine B / ENV+TIME" "$SHOW_DIR/worker-2.log"
open_view "Four-Way Life — Engine C / PROCESS" "$SHOW_DIR/worker-3.log"
open_view "Four-Way Life — Engine D / COMPUTE" "$SHOW_DIR/worker-4.log"

set +e
wait "$COORD_PID"
STATUS=$?
set -e
COORD_PID=

echo
echo "showcase coordinator exit: $STATUS"
echo "logs retained in: $SHOW_DIR"
exit "$STATUS"

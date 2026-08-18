#!/bin/sh
set -eu

HERE=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
cd "$HERE"

OUT=$(mktemp)
cleanup() {
    rm -f "$OUT"
}
trap cleanup EXIT HUP TERM

echo "Four-Way Life Gate 3 systems acceptance"
echo
echo "--- signal + lock ---"

aiki coordinator.ai systems 1000 60 40 42 >"$OUT" 2>&1 &
PID=$!

# Waiting for the observer address also proves systems initialization has
# acquired the lock, installed signal handling, and opened the listener.
i=0
while ! grep -q 'four-way-life: observer ' "$OUT"; do
    if ! kill -0 "$PID" 2>/dev/null; then
        cat "$OUT"
        echo "FAIL: coordinator exited during startup" >&2
        exit 1
    fi
    i=$((i + 1))
    if [ "$i" -gt 100 ]; then
        kill "$PID" 2>/dev/null || true
        wait "$PID" 2>/dev/null || true
        cat "$OUT"
        echo "FAIL: coordinator did not initialize systems spine" >&2
        exit 1
    fi
    sleep 0.05
done

LOCK_STATE=$(aiki lock_probe.ai)
[ "$LOCK_STATE" = "LOCKED" ] || {
    cat "$OUT"
    echo "FAIL: expected lock to be held, got $LOCK_STATE" >&2
    kill "$PID" 2>/dev/null || true
    wait "$PID" 2>/dev/null || true
    exit 1
}
echo "lock held: PASS"

kill -INT "$PID"
if ! wait "$PID"; then
    cat "$OUT"
    echo "FAIL: signal shutdown returned nonzero" >&2
    exit 1
fi

if ! grep -q 'four-way-life: signal :interrupt' "$OUT"; then
    cat "$OUT"
    echo "FAIL: interrupt was not observed" >&2
    exit 1
fi
if ! grep -qx 'CLOSED' "$OUT"; then
    cat "$OUT"
    echo "FAIL: coordinator did not complete graceful shutdown" >&2
    exit 1
fi
echo "signal shutdown: PASS"

LOCK_STATE=$(aiki lock_probe.ai)
[ "$LOCK_STATE" = "FREE" ] || {
    cat "$OUT"
    echo "FAIL: expected lock released after shutdown, got $LOCK_STATE" >&2
    exit 1
}
echo "lock released: PASS"

if [ ! -s ../results/four-way-life.log ]; then
    echo "FAIL: generation log is missing or empty" >&2
    exit 1
fi
if ! grep -q '^GEN ' ../results/four-way-life.log; then
    echo "FAIL: generation log contains no generation records" >&2
    exit 1
fi
echo "generation log: PASS"

echo
echo "--- terminal raw/restore ---"
aiki terminal_probe.ai

echo
echo "GATE 3 PASS"

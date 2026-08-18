#!/bin/sh
set -eu

HERE=$(CDPATH= cd -- "$(dirname "$0")" && pwd)
RESULTS="$HERE/../results"
mkdir -p "$RESULTS"
cd "$HERE"

if ! command -v aiki >/dev/null 2>&1; then
    echo "experiment: aiki not found on PATH" >&2
    exit 1
fi

STAMP=$(date '+%Y-%m-%d-%H%M%S.%3N')
RESULT="$RESULTS/run-$STAMP.txt"
ONE=$(mktemp)
TWO=$(mktemp)
trap 'rm -f "$ONE" "$TWO"' EXIT

run_experiment() {
    echo "Experiment 003 — Four-Way Life"
    echo "Gate 1 — deterministic five-process Life core"
    echo
    printf 'Aiki executable: '
    command -v aiki
    printf 'Aiki version:    '
    aiki -v
    echo
    echo "--- Pure Life rules ---"
    aiki test life_test.ai
    echo
    echo "--- Line protocol ---"
    aiki test protocol_test.ai
    echo
    echo "--- Determinism run 1 ---"
    if ! aiki coordinator.ai headless 12 24 18 42 >"$ONE" 2>&1; then
        cat "$ONE"
        echo "FAIL: coordinator run 1 failed" >&2
        exit 1
    fi
    cat "$ONE"
    if ! grep -qx 'PASS' "$ONE"; then
        echo "FAIL: coordinator run 1 did not reach PASS" >&2
        exit 1
    fi
    echo
    echo "--- Determinism run 2 ---"
    if ! aiki coordinator.ai headless 12 24 18 42 >"$TWO" 2>&1; then
        cat "$TWO"
        echo "FAIL: coordinator run 2 failed" >&2
        exit 1
    fi
    cat "$TWO"
    if ! grep -qx 'PASS' "$TWO"; then
        echo "FAIL: coordinator run 2 did not reach PASS" >&2
        exit 1
    fi
    echo
    if ! cmp -s "$ONE" "$TWO"; then
        echo "FAIL: identical seed/input produced different committed generations" >&2
        diff -u "$ONE" "$TWO" >&2 || true
        exit 1
    fi
    echo "DETERMINISM PASS"
    echo
    echo "Interactive canvas: aiki coordinator.ai canvas 100 48 32 42"
}

run_experiment 2>&1 | tee "$RESULT"
printf '\nresult: %s\n' "$RESULT"

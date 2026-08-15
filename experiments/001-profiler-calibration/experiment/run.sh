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

run_profile() {
    name=$1
    echo "================================================================"
    echo "$name"
    echo "================================================================"
    aiki profile --counts "$name"
    echo
}

run_experiment() {
    echo "Experiment 001 — Profiler Calibration"
    echo
    printf 'Aiki executable: '
    command -v aiki
    printf 'Aiki version:    '
    aiki -v
    echo

    echo "---- Sanity check: visible decade scaling -----------------------"
    echo "Expected leaf calculations: 10, 100, 1000, 10000"
    echo "Expected total loop iterations: 10, 110, 1110, 11110"
    echo
    run_profile sanity-10.ai
    run_profile sanity-100.ai
    run_profile sanity-1000.ai
    run_profile sanity-10000.ai

    echo "---- Native exact count -----------------------------------------"
    run_profile native-exact.ai

    echo "---- Native loop scaling ----------------------------------------"
    run_profile native-loop-10.ai
    run_profile native-loop-100.ai
    run_profile native-loop-1000.ai

    echo "---- One-level self-host differential scaling -------------------"
    run_profile selfhost-1x-0.ai
    run_profile selfhost-1x-1.ai
    run_profile selfhost-1x-2.ai
    run_profile selfhost-1x-4.ai

    echo "---- Two-level self-host differential scaling -------------------"
    run_profile selfhost-2x-0.ai
    run_profile selfhost-2x-1.ai
    run_profile selfhost-2x-2.ai

    echo "Experiment complete."
}

run_experiment 2>&1 | tee "$RESULT"
printf '\nresult: %s\n' "$RESULT"

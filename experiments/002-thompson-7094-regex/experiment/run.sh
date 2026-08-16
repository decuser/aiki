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

run_experiment() {
    echo "Experiment 002 — Thompson 7094 Regex Reconstruction"
    echo "Phases I–IV — machine, Thompson reconstruction, compiler, and operator monitor"
    echo
    printf 'Aiki executable: '
    command -v aiki
    printf 'Aiki version:    '
    aiki -v
    echo
    echo "--- Phase I machine corpus ---"
    aiki test machine_test.ai
    echo
    echo "--- Phase II Thompson corpus ---"
    aiki test phase2_test.ai
    echo
    echo "--- Phase III compiler corpus ---"
    aiki test phase3_test.ai
    echo
    echo "--- Phase IV monitor corpus ---"
    aiki test monitor_test.ai
    echo
    echo "--- End-to-end demonstration ---"
    aiki demo.ai
    echo
    echo "--- Monitor demonstration ---"
    aiki console.ai monitor-demo.cmd
}

run_experiment 2>&1 | tee "$RESULT"
printf '\nresult: %s\n' "$RESULT"

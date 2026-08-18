#!/usr/bin/env bash
set -euo pipefail

AIKI=${AIKI:-aiki}
AIKI_BIN=$(command -v "$AIKI")
OUT=${1:-../analyses/hotpath}
mkdir -p "$OUT"

WIDTH=${FWL_WIDTH:-60}
HEIGHT=${FWL_HEIGHT:-40}
SEED=${FWL_SEED:-42}

TMP_WORKER=.profile-worker.ai
TMP_COORD=.profile-coordinator.ai

cleanup() {
    rm -f "$TMP_WORKER" "$TMP_COORD"
}
trap cleanup EXIT HUP INT TERM

make_worker() {
    local owner=$1
    awk -v owner="$owner" '
        /^let args = system\.args\(\)/ {
            print "let args = [\"" owner "\"]"
            next
        }
        { print }
    ' worker.ai >"$TMP_WORKER"
}

make_coordinator() {
    local gens=$1
    awk -v gens="$gens" -v width="$WIDTH" -v height="$HEIGHT" -v seed="$SEED" '
        /^let mode = /        { print "let mode = \"headless\""; next }
        /^let generations = / { print "let generations = " gens; next }
        /^let width = /       { print "let width = " width; next }
        /^let height = /      { print "let height = " height; next }
        /^let seed = /        { print "let seed = " seed; next }
        { print }
    ' coordinator.ai >"$TMP_COORD"
}

echo "Four-Way Life hot-path profiling"
echo "baseline: $(git rev-parse --short HEAD 2>/dev/null || echo unknown)"
echo "grid: ${WIDTH}x${HEIGHT} seed=${SEED}"
echo "output: $OUT"
echo

FWL_WIDTH="$WIDTH" FWL_HEIGHT="$HEIGHT" FWL_SEED="$SEED" \
    "$AIKI" profile_frame.ai >"$OUT/worker-frame.txt"

for OWNER in 1 4; do
    PREFIX="$OUT/worker-${OWNER}-${WIDTH}x${HEIGHT}"
    make_worker "$OWNER"

    echo "--- Worker $OWNER: attributed Aiki profile ---"
    FWL_SEED="$SEED" FWL_MUTATION_RATE=2 \
        "$AIKI" profile "$TMP_WORKER" \
        <"$OUT/worker-frame.txt" >"$PREFIX.attributed.txt"
    sed -n '/Aiki source attribution/,$p' "$PREFIX.attributed.txt" | head -120
    echo

    echo "--- Worker $OWNER: correlated CPU profile ---"
    FWL_SEED="$SEED" FWL_MUTATION_RATE=2 \
        "$AIKI" profile --cpu "$PREFIX.cpu.pprof" "$TMP_WORKER" \
        <"$OUT/worker-frame.txt" >"$PREFIX.cpu-run.txt"

    go tool pprof -top "$AIKI_BIN" "$PREFIX.cpu.pprof" \
        >"$PREFIX.cpu-top.txt" 2>&1 || true
    go tool pprof -tags "$AIKI_BIN" "$PREFIX.cpu.pprof" \
        >"$PREFIX.cpu-tags.txt" 2>&1 || true
    head -45 "$PREFIX.cpu-top.txt"
    echo

    echo "--- Worker $OWNER: allocation profile ---"
    FWL_SEED="$SEED" FWL_MUTATION_RATE=2 \
        "$AIKI" profile --allocs "$PREFIX.allocs.pprof" "$TMP_WORKER" \
        <"$OUT/worker-frame.txt" >"$PREFIX.allocs-run.txt"

    go tool pprof -top -alloc_space "$AIKI_BIN" "$PREFIX.allocs.pprof" \
        >"$PREFIX.allocs-top.txt" 2>&1 || true
    head -45 "$PREFIX.allocs-top.txt"
    echo
done

echo "--- Coordinator: 5-generation attributed profile ---"
make_coordinator 5
"$AIKI" profile "$TMP_COORD" >"$OUT/coordinator-5gen.attributed.txt"
sed -n '/Aiki source attribution/,$p' "$OUT/coordinator-5gen.attributed.txt" | head -100

echo
echo "HOTPATH PROFILE PASS"

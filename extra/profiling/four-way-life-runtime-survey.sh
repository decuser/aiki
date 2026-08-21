#!/bin/sh
set -eu

AIKI=${AIKI:-./aiki}
OUT=${1:-/tmp/aiki-four-way-life-runtime-survey}
WIDTH=${FWL_WIDTH:-60}
HEIGHT=${FWL_HEIGHT:-40}
SEED=${FWL_SEED:-42}
GENERATIONS=${FWL_GENERATIONS:-5}

ROOT=$(pwd)
case "$AIKI" in
    /*) AIKI_BIN=$AIKI ;;
    *) AIKI_BIN="$ROOT/${AIKI#./}" ;;
esac

if [ ! -x "$AIKI_BIN" ]; then
    echo "four-way-life survey: executable $AIKI_BIN is missing" >&2
    exit 1
fi

EXP="$ROOT/experiments/003-four-way-life/experiment"
if [ ! -d "$EXP" ]; then
    echo "four-way-life survey: experiment tree not found: $EXP" >&2
    exit 1
fi

rm -rf "$OUT"
mkdir -p "$OUT"

TMP_COORD="$EXP/.runtime-survey-coordinator.ai"
TMP_WORKER="$EXP/.runtime-survey-worker.ai"

cleanup() {
    rm -f "$TMP_COORD" "$TMP_WORKER"
}
trap cleanup EXIT HUP INT TERM

{
    echo "generated_utc=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
    echo "git=$(git describe --tags --always --dirty 2>/dev/null || echo unavailable)"
    echo "head=$(git rev-parse HEAD 2>/dev/null || echo unavailable)"
    echo "aiki=$($AIKI_BIN version 2>/dev/null || echo unknown)"
    echo "go=$(go version 2>/dev/null || echo unavailable)"
    echo "grid=${WIDTH}x${HEIGHT}"
    echo "seed=$SEED"
    echo "generations=$GENERATIONS"
} > "$OUT/manifest.txt"

pprof_views() {
    name=$1
    profile="$OUT/$name.allocs.pprof"

    go tool pprof -top -nodecount=40 -sample_index=alloc_space \
        "$AIKI_BIN" "$profile" > "$OUT/$name.alloc-space.flat.txt" 2>&1 || true
    go tool pprof -top -cum -nodecount=40 -sample_index=alloc_space \
        "$AIKI_BIN" "$profile" > "$OUT/$name.alloc-space.cum.txt" 2>&1 || true
    go tool pprof -top -nodecount=40 -sample_index=alloc_objects \
        "$AIKI_BIN" "$profile" > "$OUT/$name.alloc-objects.flat.txt" 2>&1 || true
    go tool pprof -top -cum -nodecount=40 -sample_index=alloc_objects \
        "$AIKI_BIN" "$profile" > "$OUT/$name.alloc-objects.cum.txt" 2>&1 || true
}

make_coordinator() {
    awk -v gens="$GENERATIONS" -v width="$WIDTH" -v height="$HEIGHT" -v seed="$SEED" '
        /^let args = system\.args\(\)/ { print "let args = []"; next }
        /^let mode = /        { print "let mode = \"headless\""; next }
        /^let generations = / { print "let generations = " gens; next }
        /^let width = /       { print "let width = " width; next }
        /^let height = /      { print "let height = " height; next }
        /^let seed = /        { print "let seed = " seed; next }
        { print }
    ' "$EXP/coordinator.ai" > "$TMP_COORD"
}

make_worker() {
    owner=$1
    awk -v owner="$owner" '
        /^let args = system\.args\(\)/ {
            print "let args = [\"" owner "\"]"
            next
        }
        { print }
    ' "$EXP/worker.ai" > "$TMP_WORKER"
}

cd "$EXP"

# Ensure coordinator-spawned workers resolve to the exact executable being
# surveyed rather than to an unrelated installed `aiki`.
PATH="$ROOT:$PATH"
export PATH

echo "Four-Way Life runtime survey"
echo "grid: ${WIDTH}x${HEIGHT}, generations: ${GENERATIONS}, seed: ${SEED}"
echo "output: $OUT"

make_coordinator

echo "--- coordinator ---"
"$AIKI_BIN" profile --counts -allocs "$OUT/life-coordinator.allocs.pprof" \
    "$TMP_COORD" > "$OUT/life-coordinator.counts.txt"
pprof_views life-coordinator

# Generate one deterministic generation frame used identically by all four
# workers. Each worker remains an independent Aiki interpreter profile.
FWL_WIDTH="$WIDTH" \
FWL_HEIGHT="$HEIGHT" \
FWL_SEED="$SEED" \
    "$AIKI_BIN" profile_frame.ai > "$OUT/worker-frame.txt"

for OWNER in 1 2 3 4; do
    NAME="life-worker-${OWNER}"
    make_worker "$OWNER"
    echo "--- worker $OWNER ---"
    FWL_SEED="$SEED" \
    FWL_MUTATION_RATE=2 \
        "$AIKI_BIN" profile --counts -allocs "$OUT/$NAME.allocs.pprof" \
        "$TMP_WORKER" < "$OUT/worker-frame.txt" > "$OUT/$NAME.counts.txt"
    pprof_views "$NAME"
done

# Full application timing is intentionally separate from per-process pprof.
# This is the real five-process system: coordinator + four worker interpreters.
echo "--- full five-process run ---"
(
    time "$AIKI_BIN" coordinator.ai headless "$GENERATIONS" \
        "$WIDTH" "$HEIGHT" "$SEED" > "$OUT/five-process.transcript.txt"
) 2> "$OUT/five-process.time.txt"

if ! grep -qx 'PASS' "$OUT/five-process.transcript.txt"; then
    echo "four-way-life survey: full five-process run did not reach PASS" >&2
    exit 1
fi

cat > "$OUT/README.txt" <<EOF2
Four-Way Life runtime evidence.

Canonical frame:
  coordinator: headless, ${GENERATIONS} generations, ${WIDTH}x${HEIGHT}, seed ${SEED}
  workers: one deterministic generation frame each, owners 1..4

Per-process allocation evidence:
  life-coordinator.{counts,alloc-space.*,alloc-objects.*}
  life-worker-1.{counts,alloc-space.*,alloc-objects.*}
  life-worker-2.{counts,alloc-space.*,alloc-objects.*}
  life-worker-3.{counts,alloc-space.*,alloc-objects.*}
  life-worker-4.{counts,alloc-space.*,alloc-objects.*}

Whole-system evidence:
  five-process.time.txt
  five-process.transcript.txt

The five pprof profiles are deliberately not merged. Worker A/B/C/D exercise
different load-bearing domains, and preserving their identities is part of the
benchmark's value.
EOF2

echo "FOUR-WAY LIFE SURVEY PASS"
echo "results: $OUT"

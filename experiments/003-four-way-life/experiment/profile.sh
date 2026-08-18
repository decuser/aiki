#!/bin/sh
set -eu

AIKI=${AIKI:-aiki}
OUT=${1:-../analyses/profile}
mkdir -p "$OUT"

WIDTH=${FWL_WIDTH:-60}
HEIGHT=${FWL_HEIGHT:-40}
SEED=${FWL_SEED:-42}

echo "Four-Way Life profiling"
echo "grid: ${WIDTH}x${HEIGHT} seed=${SEED}"
echo "output: $OUT"

for GENS in 1 5 20; do
    NAME="coordinator-${WIDTH}x${HEIGHT}-${GENS}gen.counts.txt"
    echo "--- coordinator: ${GENS} generation(s) ---"
    FWL_MODE=headless \
    FWL_GENERATIONS="$GENS" \
    FWL_WIDTH="$WIDTH" \
    FWL_HEIGHT="$HEIGHT" \
    FWL_SEED="$SEED" \
        "$AIKI" profile --counts coordinator.ai >"$OUT/$NAME"
    tail -20 "$OUT/$NAME"
    echo
 done

# Profile each worker independently on the same deterministic generation frame.
# This is essential: `aiki profile coordinator.ai` profiles only the coordinator
# process; each worker is a separate Aiki interpreter.
"$AIKI" profile_frame.ai >"$OUT/worker-frame.txt"
for OWNER in 1 2 3 4; do
    NAME="worker-${OWNER}-${WIDTH}x${HEIGHT}.counts.txt"
    echo "--- worker ${OWNER}: one generation ---"
    FWL_OWNER="$OWNER" \
    FWL_SEED="$SEED" \
    FWL_MUTATION_RATE=2 \
        "$AIKI" profile --counts worker.ai <"$OUT/worker-frame.txt" >"$OUT/$NAME"
    tail -20 "$OUT/$NAME"
    echo
 done

# One correlated CPU profile at a representative coordinator size. Worker CPU
# profiles can be requested similarly once the count data identifies a target.
echo "--- coordinator CPU profile: 5 generations ---"
FWL_MODE=headless \
FWL_GENERATIONS=5 \
FWL_WIDTH="$WIDTH" \
FWL_HEIGHT="$HEIGHT" \
FWL_SEED="$SEED" \
    "$AIKI" profile --cpu "$OUT/coordinator-5gen.cpu.pprof" coordinator.ai \
    >"$OUT/coordinator-5gen.attributed.txt"

echo "PROFILE PASS"

#!/bin/sh
set -eu

AIKI=${AIKI:-./aiki}
OUT=${1:-/tmp/aiki-runtime-realization-survey}
mkdir -p "$OUT"

{
    echo "generated_utc=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
    echo "aiki=$($AIKI version 2>/dev/null || echo unknown)"
    echo "go=$(go version 2>/dev/null || echo unavailable)"
    echo "host=$(uname -a 2>/dev/null || echo unavailable)"
    echo "git=$(git describe --tags --always --dirty 2>/dev/null || echo unavailable)"
} > "$OUT/manifest.txt"

profile_one() {
    name=$1
    src=$2
    shift 2

    echo "profiling $name"
    "$@" "$AIKI" profile --counts -allocs "$OUT/$name.allocs.pprof" "$src" \
        > "$OUT/$name.counts.txt"

    go tool pprof -top -nodecount=40 -sample_index=alloc_space \
        "$AIKI" "$OUT/$name.allocs.pprof" \
        > "$OUT/$name.alloc-space.flat.txt" 2>&1 || true
    go tool pprof -top -cum -nodecount=40 -sample_index=alloc_space \
        "$AIKI" "$OUT/$name.allocs.pprof" \
        > "$OUT/$name.alloc-space.cum.txt" 2>&1 || true
    go tool pprof -top -nodecount=40 -sample_index=alloc_objects \
        "$AIKI" "$OUT/$name.allocs.pprof" \
        > "$OUT/$name.alloc-objects.flat.txt" 2>&1 || true
    go tool pprof -top -cum -nodecount=40 -sample_index=alloc_objects \
        "$AIKI" "$OUT/$name.allocs.pprof" \
        > "$OUT/$name.alloc-objects.cum.txt" 2>&1 || true
}

profile_one selfhost \
    extra/profiling/selfhost-three-level.ai \
    env

profile_one pdp10x \
    experiments/004-v6-emulator/experiment/diagnostics/cut7_perf_loop.ai \
    env AIKI_PDP_PERF_SCALE=10

cat > "$OUT/README.txt" <<EOF2
Runtime realization survey results.

Primary files to inspect:
  selfhost.counts.txt
  selfhost.alloc-space.flat.txt
  selfhost.alloc-space.cum.txt
  selfhost.alloc-objects.flat.txt
  selfhost.alloc-objects.cum.txt
  pdp10x.counts.txt
  pdp10x.alloc-space.flat.txt
  pdp10x.alloc-space.cum.txt
  pdp10x.alloc-objects.flat.txt
  pdp10x.alloc-objects.cum.txt

Allocation profiles are ordinary Go allocation-site profiles and are not Aiki
source-label correlated.
EOF2

printf 'results: %s\n' "$OUT"

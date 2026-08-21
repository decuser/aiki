#!/bin/sh
set -eu

AIKI=${AIKI:-./aiki}
OUT=${1:-/tmp/aiki-runtime-survey-alpha38}
ARCHIVE=${OUT}.tar.gz

if [ ! -x "$AIKI" ]; then
    echo "alpha38 survey: executable $AIKI is missing" >&2
    echo "build the current tree first (make), then rerun this command" >&2
    exit 1
fi

rm -rf "$OUT" "$ARCHIVE"
mkdir -p "$OUT"

{
    echo "generated_utc=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
    echo "head=$(git rev-parse HEAD 2>/dev/null || echo unavailable)"
    echo "tag=$(git tag --points-at HEAD 2>/dev/null | paste -sd, - || echo unavailable)"
    echo "branch=$(git branch --show-current 2>/dev/null || echo unavailable)"
    echo "describe=$(git describe --tags --always --dirty 2>/dev/null || echo unavailable)"
    echo "aiki=$($AIKI version 2>/dev/null || echo unknown)"
    echo "go=$(go version 2>/dev/null || echo unavailable)"
    echo "host=$(uname -a 2>/dev/null || echo unavailable)"
} > "$OUT/alpha38-manifest.txt"

# Preserve the exact current runtime totals independently of pprof output.
"$AIKI" profile --counts extra/profiling/selfhost-three-level.ai \
    > "$OUT/selfhost-baseline.counts.txt"

env AIKI_PDP_PERF_SCALE=10 "$AIKI" profile --counts \
    experiments/004-v6-emulator/experiment/diagnostics/cut7_perf_loop.ai \
    > "$OUT/pdp10x-baseline.counts.txt"

# Reuse the established allocation-space/object survey unchanged.
extra/profiling/runtime-realization-survey.sh "$OUT/pprof" \
    > "$OUT/survey-run.txt" 2>&1

# Third evidence leg: a heterogeneous, deterministic five-process Aiki
# application. Keep coordinator and workers separate rather than merging their
# pprof profiles.
extra/profiling/four-way-life-runtime-survey.sh "$OUT/four-way-life" \
    > "$OUT/four-way-life-run.txt" 2>&1

cat > "$OUT/README.txt" <<'EOF2'
Post-alpha-38 runtime realization survey.

Authoritative selection basis:
  alpha38-manifest.txt
  selfhost-baseline.counts.txt
  pdp10x-baseline.counts.txt
  pprof/selfhost.alloc-space.flat.txt
  pprof/selfhost.alloc-space.cum.txt
  pprof/selfhost.alloc-objects.flat.txt
  pprof/selfhost.alloc-objects.cum.txt
  pprof/pdp10x.alloc-space.flat.txt
  pprof/pdp10x.alloc-space.cum.txt
  pprof/pdp10x.alloc-objects.flat.txt
  pprof/pdp10x.alloc-objects.cum.txt
  four-way-life/life-coordinator.*
  four-way-life/life-worker-1.*
  four-way-life/life-worker-2.*
  four-way-life/life-worker-3.*
  four-way-life/life-worker-4.*
  four-way-life/five-process.time.txt

The three evidence families are intentionally different:
  selfhost       language implementation / evaluator pressure
  PDP-11         systems-emulator / interpreter-loop pressure
  Four-Way Life  multiprocess / IPC / heterogeneous library pressure

Four-Way Life pprof files remain per process. Do not sum or merge them before
inspection; coordinator and workers A-D exercise deliberately different domains.

This survey is attribution-only. It does not presume that the ranking from
alpha-37 survives the environment optimization wave.
EOF2

tar -czf "$ARCHIVE" -C "$(dirname "$OUT")" "$(basename "$OUT")"

printf 'results: %s\n' "$OUT"
printf 'archive: %s\n' "$ARCHIVE"

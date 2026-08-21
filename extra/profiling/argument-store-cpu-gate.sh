#!/bin/sh
set -eu

AIKI=${AIKI:-./aiki}
OUT=${1:-/tmp/aiki-argument-store-cpu-gate}

if [ ! -x "$AIKI" ]; then
    echo "argument-store cpu gate: executable not found: $AIKI" >&2
    exit 1
fi

rm -rf "$OUT"
mkdir -p "$OUT"

{
    echo "generated_utc=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
    echo "git=$(git describe --tags --always --dirty 2>/dev/null || echo unavailable)"
    echo "head=$(git rev-parse HEAD 2>/dev/null || echo unavailable)"
    echo "go=$(go version 2>/dev/null || echo unavailable)"
} > "$OUT/manifest.txt"

run_cpu() {
    name=$1
    shift

    echo "== $name CPU profile =="
    "$@" 2>&1 | tee "$OUT/$name.counts.txt"

    go tool pprof \
        -top \
        -nodecount=60 \
        "$AIKI" "$OUT/$name.cpu.pprof" \
        > "$OUT/$name.cpu.flat.txt" 2>&1

    go tool pprof \
        -top \
        -cum \
        -nodecount=60 \
        "$AIKI" "$OUT/$name.cpu.pprof" \
        > "$OUT/$name.cpu.cum.txt" 2>&1

    go tool pprof \
        -top \
        -nodecount=80 \
        -focus='argFrame|ArgFrame|acquireArgValues|releaseArgFrame|evalCallArgsFor|applyUserFunctionOwned|sync\.\(\*Pool\)|sync\.Pool' \
        "$AIKI" "$OUT/$name.cpu.pprof" \
        > "$OUT/$name.cpu.argument-focus.txt" 2>&1 || true
}

run_cpu selfhost \
    "$AIKI" profile --counts \
    --cpu "$OUT/selfhost.cpu.pprof" \
    extra/profiling/selfhost-three-level.ai

echo
echo "== pdp10x CPU profile =="
AIKI_PDP_PERF_SCALE=10 \
    "$AIKI" profile --counts \
    --cpu "$OUT/pdp10x.cpu.pprof" \
    experiments/004-v6-emulator/experiment/diagnostics/cut7_perf_loop.ai \
    2>&1 | tee "$OUT/pdp10x.counts.txt"

go tool pprof -top -nodecount=60 \
    "$AIKI" "$OUT/pdp10x.cpu.pprof" \
    > "$OUT/pdp10x.cpu.flat.txt" 2>&1

go tool pprof -top -cum -nodecount=60 \
    "$AIKI" "$OUT/pdp10x.cpu.pprof" \
    > "$OUT/pdp10x.cpu.cum.txt" 2>&1

go tool pprof -top -nodecount=80 \
    -focus='argFrame|ArgFrame|acquireArgValues|releaseArgFrame|evalCallArgsFor|applyUserFunctionOwned|sync\.\(\*Pool\)|sync\.Pool' \
    "$AIKI" "$OUT/pdp10x.cpu.pprof" \
    > "$OUT/pdp10x.cpu.argument-focus.txt" 2>&1 || true

tar -czf "${OUT}.tar.gz" -C "$(dirname "$OUT")" "$(basename "$OUT")"

echo
echo "argument-store CPU evidence: $OUT"
echo "archive: ${OUT}.tar.gz"

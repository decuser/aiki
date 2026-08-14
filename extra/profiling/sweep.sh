#!/bin/sh
set -eu

AIKI=${AIKI:-./aiki}
OUT=${1:-profile-results}
mkdir -p "$OUT"

{
    echo "generated_utc=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
    echo "aiki=$($AIKI version 2>/dev/null || echo unknown)"
    echo "go=$(go version 2>/dev/null || echo unavailable)"
    echo "host=$(uname -a 2>/dev/null || echo unavailable)"
    echo "git=$(git describe --tags --always --dirty 2>/dev/null || echo unavailable)"
} > "$OUT/manifest.txt"

for src in extra/profiling/[0-9][0-9]-*.ai; do
    name=$(basename "$src" .ai)
    echo "profiling $name"
    "$AIKI" profile \
        -cpu "$OUT/$name.cpu.pprof" \
        -allocs "$OUT/$name.allocs.pprof" \
        "$src" > "$OUT/$name.txt"

    go tool pprof -top -nodecount=20 "$AIKI" "$OUT/$name.cpu.pprof" \
        > "$OUT/$name.cpu.top.txt" 2>&1 || true
    go tool pprof -tags "$OUT/$name.cpu.pprof" \
        > "$OUT/$name.cpu.tags.txt" 2>&1 || true
    go tool pprof -top -sample_index=alloc_space -nodecount=20 "$AIKI" "$OUT/$name.allocs.pprof" \
        > "$OUT/$name.allocs.top.txt" 2>&1 || true
done

printf 'results: %s\n' "$OUT"
